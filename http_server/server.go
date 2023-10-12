package http_server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/replicated_log/api"
	"github.com/replicated_log/config"
	"github.com/replicated_log/lifecycle"
	"github.com/replicated_log/storage"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"net/http"
	"time"
)

const (
	ctHeader = "Content-Type"
)

type resp struct {
	Data []string `json:"data"`
	Err  string   `json:"err"`
}

type Server struct {
	lifecycle.Lifecycle

	server   http.Server
	replicas []string
	handlers map[string]func(w http.ResponseWriter, r *http.Request)
}

func NewServer(cfg config.Config) lifecycle.Lifecycle {
	s := &Server{
		replicas: cfg.Replicas,
	}
	s.handlers = map[string]func(w http.ResponseWriter, r *http.Request){
		http.MethodGet: s.listMsgs,
	}
	if cfg.IsMainNode {
		log.Info().Msg("Service configured as main node")
		s.handlers[http.MethodPost] = s.appendMsgs
	}

	return s
}
func (s *Server) Start() error {
	log.Info().Msg("Starting http server")
	mux := http.NewServeMux()
	mux.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		handler, ok := s.handlers[r.Method]
		if !ok {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler(w, r)
	})

	s.server = http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	return s.server.ListenAndServe()
}

func (s *Server) Stop() error {
	//add shutdown on timeout just in case
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

func (s *Server) listMsgs(w http.ResponseWriter, _ *http.Request) {
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

func (s *Server) appendMsgs(w http.ResponseWriter, r *http.Request) {
	msgRaw, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := resp{
			Err: errors.New("failed to read request body").Error(),
		}
		if err = json.NewEncoder(w).Encode(&resp); err != nil {
			log.Error().Err(err).Msg("failed to write http response body")
		}
		return
	}
	msg := string(msgRaw)

	storage.GetList().Add(msg)

	for _, replica := range s.replicas {
		if err = s.replicate(replica, msg); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			resp := resp{
				Err: fmt.Errorf("failed to replicate log: %w", err).Error(),
			}
			if err = json.NewEncoder(w).Encode(&resp); err != nil {
				log.Error().Err(err).Msg("failed to write http response body")
			}
			return
		}
	}

	s.listMsgs(w, r)
}

func (s *Server) replicate(replica, msg string) error {
	conn, err := grpc.Dial(replica, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to dial to replica node: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close replica connection")
		}
	}()

	client := api.NewReplicatorClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = client.Replicate(ctx, &api.ReplicateRequest{Msg: msg})
	if err != nil {
		return fmt.Errorf("failed to replicate message: %w", err)
	}

	return nil
}
