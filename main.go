package main

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/replicated_log/config"
	"github.com/replicated_log/grpc_server"
	"github.com/replicated_log/http_server"
	"github.com/replicated_log/lifecycle"
	"github.com/replicated_log/storage"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

//go:generate protoc --go_out=./api --go_opt=paths=source_relative --go-grpc_out=./api --go-grpc_opt=paths=source_relative api.proto
func main() {
	//init human readable logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	var cfg config.Config

	log.Info().Msg("Starting replicated log iteration 1")
	log.Info().Msg("Reading config from envs")
	if err := envconfig.Process("", &cfg); err != nil {
		log.Log().Err(err).Msg("failed to read config from env vars")
		return
	}

	log.Debug().Msgf("%v", cfg.Replicas)
	log.Info().Msg("Initializing in memory storage")
	storage.InitList()

	var apis []lifecycle.Lifecycle
	if !cfg.IsMainNode {
		log.Info().Msg("Service configured as child node")
		apis = append(apis, &grpc_server.Server{Delay: cfg.Delay})
	}

	httpServer := http_server.NewServer(cfg)
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
