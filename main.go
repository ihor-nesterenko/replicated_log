package main

import (
	"encoding/json"
	"errors"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
)

const (
	ctHeader = "Content-Type"
)

type resp struct {
	Data []string `json:"data"`
	Err  error    `json:"err"`
}

var methods = map[string]func(w http.ResponseWriter, r *http.Request){
	http.MethodGet: listMsgs,
}

func main() {
	//init human readable logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	var cfg config

	if err := envconfig.Process("", &cfg); err != nil {
		log.Log().Err(err).Msg("failed to read config from env vars")
		return
	}

	initList()
	//if main node -> add POST handler
	if cfg.IsMainNode {
		methods[http.MethodPost] = appendMsgs
	}

	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		handler, ok := methods[r.Method]
		if !ok {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler(w, r)
	})

	http.ListenAndServe(":8080", nil)
}

func listMsgs(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(ctHeader, "application/json")

	list := inMemoryList.Get()
	resp := resp{
		Data: list,
	}
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func appendMsgs(w http.ResponseWriter, r *http.Request) {
	msgRaw, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := resp{
			Err: errors.New("failed to read request body"),
		}
		if err = json.NewEncoder(w).Encode(&resp); err != nil {
			log.Error().Err(err).Msg("failed to write http response body")
		}
		return
	}

	inMemoryList.Add(string(msgRaw))

	listMsgs(w, r)
}
