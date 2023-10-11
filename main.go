package main

import (
	"encoding/json"
	"errors"
	"github.com/kelseyhightower/envconfig"
	"github.com/replicated_log/grpc_server"
	"github.com/replicated_log/http_server"
	"github.com/replicated_log/lifecycle"
	"github.com/replicated_log/storage"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	ctHeader = "Content-Type"
)

type resp struct {
	Data []string `json:"data"`
	Err  error    `json:"err"`
}

//go:generate protoc --go_out=./api --go_opt=paths=source_relative --go-grpc_out=./api --go-grpc_opt=paths=source_relative api.proto
func main() {
	//init human readable logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	var cfg config

	log.Info().Msg("Starting replicated log iteration 1")
	log.Info().Msg("Reading config from envs")
	if err := envconfig.Process("", &cfg); err != nil {
		log.Log().Err(err).Msg("failed to read config from env vars")
		return
	}

	log.Info().Msg("Initializing in memory storage")
	storage.InitList()

	var methods = map[string]func(w http.ResponseWriter, r *http.Request){
		http.MethodGet: listMsgs,
	}
	var apis []lifecycle.Lifecycle

	//Init http server
	if cfg.IsMainNode {
		log.Info().Msg("Service configured as main node")
		methods[http.MethodPost] = appendMsgs
	} else {
		log.Info().Msg("Service configured as child node")
		apis = append(apis, &grpc_server.Server{})
	}

	httpServer := http_server.NewServer(methods)
	apis = append(apis, httpServer)

	//listen for syscalls to gracefully shutdown all goroutines
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	for _, api := range apis {
		api := api
		go func() {
			if err := api.Start(); err != nil {
				log.Warn().Err(err).Msg("Error on Start")
			}
		}()
	}

	log.Info().Msg("Replicator Started")
	//wait for os signal
	<-done

	//gracefully stop all child processes
	for _, api := range apis {
		if err := api.Stop(); err != nil {
			log.Warn().Err(err).Msg("Error on Stop")
		}
	}

	log.Info().Msg("Replicator gracefully stopped")
}

func listMsgs(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(ctHeader, "application/json")

	list := storage.GetList().Get()
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

	storage.GetList().Add(string(msgRaw))

	listMsgs(w, r)
}
