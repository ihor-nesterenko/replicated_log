package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/replicated_log/api"
	"github.com/replicated_log/grpc_server"
	"github.com/replicated_log/storage"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"io"
	"net"
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

	//FIXME: add proper lifecycle for concurrent processes here
	//if main node -> add POST handler
	if cfg.IsMainNode {
		log.Info().Msg("Service configured as main node")
		methods[http.MethodPost] = appendMsgs
	} else { //start grpc server to enable replication
		go func() {
			//FIXME: move this to init function
			log.Info().Msg("Service started as child node; starting replication gRPC server")
			lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 50051))
			if err != nil {
				log.Error().Err(err).Msg("failed to listen")
				return
			}
			s := grpc.NewServer()
			defer s.Stop()
			api.RegisterReplicatorServer(s, &grpc_server.Server{})
			if err := s.Serve(lis); err != nil {
				log.Error().Err(err).Msg("failed to listen")
				return
			}
		}()
	}

	//FIXME: replace DefaultServeMux with custom one
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
