package queues

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/podded/bouncer"
	"github.com/podded/podded/ectoplasma"
	"github.com/podded/podded/types"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
	"strconv"
)

type ESIKillmailIngest struct {
	goop *ectoplasma.PodGoo
}

func NewESIKillmailIngest(goop *ectoplasma.PodGoo) *ESIKillmailIngest {
	return &ESIKillmailIngest{goop: goop}
}

func (sc *ESIKillmailIngest) StartScraper() (err error) {
	if sc == nil || sc.goop == nil {
		return errors.New("Need to init ESIKillmailIngest / goop struct")
	}

	ctx := context.Background()

	kmdb := sc.goop.MongoClient.Database("podded").Collection("killmails")

	for {
		// Check if any available IDS in the ingest queue
		res, err := GetFromIngestQueueBlocking(sc.goop.RedisClient)
		if err != nil {
			log.Println(err)
			continue
		}
		// res[0] will always contain the queue name, res[1] is the value we are getting out of redis

		killid, err := strconv.Atoi(res[1])
		if err != nil {
			// TODO Load this onto the error queue to be looked at later
			log.Printf("Invalid kill id popped from queue %s", res[1])
		}

		// First get the hash we have stored for this id.

		idhp := types.IDHashPair{}

		err = kmdb.FindOne(ctx, bson.M{"_id": killid}).Decode(&idhp)

		if err != nil {
			// Welp!
			log.Printf("Failed to get hash for ID %d, err: %s", killid, err)
			continue
		}

		r := bouncer.Request{
			URL:    fmt.Sprintf("https://esi.evetech.net/v1/killmails/%d/%s/", idhp.ID, idhp.Hash),
			Method: "GET",
		}

		// TODO Add etag support back in...
		//if idhp.Killmail.ETag != "" {
		//	r.ETag = idhp.Killmail.ETag
		//}

		resp, _, err := sc.goop.BouncerClient.MakeRequest(r)
		if err != nil {
			log.Printf("ERROR making esi request for id %d: %s", killid, err)
			continue
		}

		status := resp.StatusCode

		// 422
		if status == http.StatusUnprocessableEntity {
			// Invalid killmail_id and/or killmail_hash remove this from the database
			filter := bson.M{"_id": killid}
			_, err = kmdb.DeleteOne(ctx, filter, nil)
			if err != nil {
				log.Printf("Failed to delete id %d, err: %s", killid, err)
			}
			continue
		}

		// 200
		if status == http.StatusOK {
			// This new ID Hash pair is valid... INSERT ALL THE THINGS!
			var mail types.RawMail
			err = json.Unmarshal(resp.Body, &mail)
			if err != nil {
				// Put this on the error queue as something is up
				// TODO Implement the Error Queue
				AddToErrorQueue(idhp.ID, err.Error(), ingest, sc.goop.RedisClient)
				log.Printf("ERROR: Failed to decode esi response for %d, err: %s", killid, err)
				continue
			}
			mail.ETag = resp.ETag

			// Need to save this
			id := mail.KillmailID

			filter := bson.M{"_id": mail.KillmailID}
			update := bson.M{"$set": bson.M{"esi_v1": mail}}
			_, _ = kmdb.UpdateOne(ctx, filter, update) // Not fussed about the error here // TODO Check the error

			// We have successfully scraped the killmail and updated the database
			IngestCompleted(id, sc.goop.RedisClient)

			continue
		}

		if status == http.StatusNotModified {
			// Nothing changed so nothing to do
			continue
		}

		// We got some other response code...
		// Put this on the error queue as something is up
		// TODO Implement the Error Queue
		AddToErrorQueue(idhp.ID, err.Error(), "PODDED:Q:INGEST", sc.goop.RedisClient)
		log.Printf("ERROR: Bad esi response for %d, code: %d, body: %s", killid, status, string(resp.Body))
	}
}
