package http_server

import (
	"context"
	"github.com/replicated_log/lifecycle"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

type Server struct {
	lifecycle.Lifecycle

	server   http.Server
	replicas []string
	handlers map[string]func(w http.ResponseWriter, r *http.Request)
}

func NewServer(handlers map[string]func(w http.ResponseWriter, r *http.Request)) lifecycle.Lifecycle {
	return &Server{
		handlers: handlers,
	}
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
