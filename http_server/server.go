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
	"sync"
	"time"
)

const (
	ctHeader = "Content-Type"
)

type resp struct {
	Data []string `json:"data"`
	Errs []string `json:"err"`
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
			Errs: []string{errors.New("failed to read request body").Error()},
		}
		if err = json.NewEncoder(w).Encode(&resp); err != nil {
			log.Error().Err(err).Msg("failed to write http response body")
		}
		return
	}
	msg := string(msgRaw)

	storage.GetList().Add(msg)

	replicasChs := make([]<-chan error, 0, len(s.replicas))
	for _, replica := range s.replicas {
		replicasChs = append(replicasChs, s.replicate(replica, msg))

	}

	res := merge(replicasChs...)
	var errorsFromReplicas []string
	for err := range res {
		if err != nil {
			errorsFromReplicas = append(errorsFromReplicas, fmt.Errorf("failed to replicate log: %w", err).Error())
		}
	}

	if len(errorsFromReplicas) != 0 {
		w.WriteHeader(http.StatusInternalServerError)
		resp := resp{
			Errs: errorsFromReplicas,
		}
		if err = json.NewEncoder(w).Encode(&resp); err != nil {
			log.Error().Err(err).Msg("failed to write http response body")
		}
		return
	}

	s.listMsgs(w, r)
}

func (s *Server) replicate(replica, msg string) <-chan error {
	resCh := make(chan error)
	go func() {
		defer close(resCh)

		conn, err := grpc.Dial(replica, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			resCh <- fmt.Errorf("failed to dial to replica node: %w", err)
			return
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
			resCh <- fmt.Errorf("failed to replicate message: %w", err)
			return
		}

		return
	}()
	return resCh
}

func merge[T any](chs ...<-chan T) <-chan T {
	var wg sync.WaitGroup
	out := make(chan T)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan T) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(chs))
	for _, c := range chs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
