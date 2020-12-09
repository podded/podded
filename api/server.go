package api

import (
	"errors"
	"github.com/gorilla/mux"
	"github.com/podded/podded/ectoplasma"
	"net/http"
	"strconv"
	"time"
)

type (
	ApiServer struct {
		Goop *ectoplasma.PodGoo

		Host string
		Port int
	}
)

func (as *ApiServer) ListenAndServe() (err error) {

	if as == nil {
		return errors.New("Need initialised ApiServer struct")
	}

	if as.Goop == nil {
		return errors.New("Need initialised ApiServer struct with goop")
	}

	r := mux.NewRouter()

	// Hand submission of new IDHP
	r.HandleFunc("/api/submit", as.handleInsertRequest).Methods("POST")

	srv := http.Server{
		Addr:         as.Host + ":" + strconv.Itoa(as.Port),
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Second * 90,
		Handler:      r,
	}

	if err := srv.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
