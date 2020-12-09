package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/gddo/httputil/header"
	"github.com/podded/bouncer"
	"github.com/podded/podded"
	"github.com/podded/podded/queues"
	"github.com/podded/podded/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

func (as *ApiServer) handleInsertRequest(w http.ResponseWriter, r *http.Request) {
	// The initial part of this code is sourced from here:
	//https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
	// It was released under the MIT License found here: https://opensource.org/licenses/MIT

	// If the Content-Type header is present, check that it has the value
	// application/json. Note that we are using the gddo/httputil/header
	// package to parse and extract the value here, so the check works
	// even if the client includes additional charset or boundary
	// information in the header.
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
	}

	// Use http.MaxBytesReader to enforce a maximum read of 1KB from the
	// response body. A request body larger than that will now result in
	// Decode() returning a "http: request body too large" error.
	r.Body = http.MaxBytesReader(w, r.Body, 1024)

	// Setup the decoder and call the DisallowUnknownFields() method on it.
	// This will cause Decode() to return a "json: unknown field ..." error
	// if it encounters any extra unexpected fields in the JSON. Strictly
	// speaking, it returns an error for "keys which do not match any
	// non-ignored, exported fields in the destination".
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var idhp types.IDHashPair
	err := dec.Decode(&idhp)

	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		// Catch any syntax errors in the JSON and send an error message
		// which interpolates the location of the problem to make it
		// easier for the client to fix.
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// In some circumstances Decode() may also return an
		// io.ErrUnexpectedEOF error for syntax errors in the JSON. There
		// is an open issue regarding this at
		// https://github.com/golang/go/issues/25956.
		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			http.Error(w, msg, http.StatusBadRequest)

		// Catch any type errors, like trying to assign a string in the
		// JSON request body to a int field in our Person struct. We can
		// interpolate the relevant field name and position into the error
		// message to make it easier for the client to fix.
		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)",
				unmarshalTypeError.Field, unmarshalTypeError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// Catch the error caused by extra unexpected fields in the request
		// body. We extract the field name from the error message and
		// interpolate it in our custom error message. There is an open
		// issue at https://github.com/golang/go/issues/29035 regarding
		// turning this into a sentinel error.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			http.Error(w, msg, http.StatusBadRequest)

		// An io.EOF error is returned by Decode() if the request body is
		// empty.
		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			http.Error(w, msg, http.StatusBadRequest)

		// Catch the error caused by the request body being too large. Again
		// there is an open issue regarding turning this into a sentinel
		// error at https://github.com/golang/go/issues/30715.
		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			http.Error(w, msg, http.StatusRequestEntityTooLarge)

		// Otherwise default to logging the error and sending a 500 Internal
		// Server Error response.
		default:
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	if dec.More() {
		msg := "Request body must only contain a single JSON object"
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// Now that we have a semi valid response lets start the clock. 30s, thats the longest I want a req to take.
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	// At this point we have a single ID Hash pair (idhp) that we can now operate on. Lets check if it is valid.
	// If it is we return a successful response to user so they can get on with their day
	// 201 if new killmail, 202 if killmail already exists but will be recalculated
	// 200 if killmail exists but wont be updated
	// 400 will be returned if the hash is bad

	kmdb := as.Goop.MongoClient.Database("podded").Collection("killmails")

	stored := types.IDHashPair{}
	var found bool = true

	err = kmdb.FindOne(ctx, bson.M{"_id": idhp.ID}).Decode(&stored)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			found = false
		} else {
			// TODO return a nice error to the user
			log.Fatal(err)
		}
	}

	if !found {
		// We do not have this hash pair yet, awesome. I love new data

		// TODO check and make sure that the hash is actually valid before inserting it

		_, err = kmdb.InsertOne(ctx, idhp)
		if err != nil {
			// There was a problem saving
			log.Printf("ERROR: %s\n", err)
			msg := "Failed to create record in db, not your fault though"
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		queues.AddToIngest(idhp.ID, as.Goop.RedisClient)
		w.WriteHeader(http.StatusCreated)
		return
	}

	// This ID already exists.... Check to see if the hashes the same
	if stored.Hash == idhp.Hash {
		// Sweet, we already have it but lets check for the magic update header presence
		if r.Header.Get(podded.MagicHeader) != "" {
			// We want to update this entry so queue it to be updated
			queues.AddToIngest(idhp.ID, as.Goop.RedisClient)
			w.WriteHeader(http.StatusAccepted)
			return
		}
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}
	// Need to check if what we have been given is valid.
	// There are a couple of status codes we could get back from esi that we should handle specifically
	// 200 - Valid killmail with the data for said kill
	// 422 - Invalid killmail_id and/or killmail_hash

	req := bouncer.Request{
		URL:    fmt.Sprintf("https://esi.evetech.net/v1/killmails/%d/%s/", idhp.ID, idhp.Hash),
		Method: "GET",
	}

	res, status, err := as.Goop.BouncerClient.MakeRequest(req)
	if err != nil {
		log.Printf("ERROR: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 422
	if status == http.StatusUnprocessableEntity {
		msg := "Invalid killmail_id and/or killmail_hash"
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// 200
	if status == http.StatusOK {
		// This new ID Hash pair is valid... Update it and remark it for ingest
		filter := bson.M{"_id": idhp.ID}
		update := bson.M{"$set": bson.M{"hash": idhp.Hash}}
		_, err = kmdb.UpdateOne(ctx, filter, update) // TODO Check the error handling
		if err != nil {
			// There was a problem saving
			log.Printf("ERROR: %s\n", err)
			msg := "Failed to create record in db, not your fault though"
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		var mail types.RawMail
		err := json.Unmarshal(res.Body, &mail)
		if err != nil {
			// Put this back on the queue for ingest but dont process here.. Though this shouldn happen...
			queues.AddToIngest(idhp.ID, as.Goop.RedisClient)
			w.WriteHeader(http.StatusCreated)
			return
		}

		f := bson.M{"_id": mail.KillmailID}
		u := bson.M{"$set": bson.M{"esi_v1": mail}}
		_, _ = kmdb.UpdateOne(ctx, f, u) // Not fussed about the error here // TODO Check the error
		return
	}

	// Weird response from ESI, pass it through
	msg := fmt.Sprintf("We have a weird response from esi. Status Code: %d", status)
	http.Error(w, msg, http.StatusInternalServerError)
}
