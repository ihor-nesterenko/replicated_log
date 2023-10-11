package main

import (
	"context"
	"github.com/replicated_log/api"
	"github.com/replicated_log/storage"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

type replicator interface {
	GetList() []string
	AddMsg(msg string)
	Replicate(msg string)
}

type replicatedLog struct {
	list     storage.List
	replicas []string
}

func newReplicatedLog(replicas []string) replicator {
	return &replicatedLog{
		list:     storage.NewList(),
		replicas: replicas,
	}
}

func (r *replicatedLog) GetList() []string {
	return r.list.Get()
}

func (r *replicatedLog) AddMsg(msg string) {
	r.list.Add(msg)
}

func (r *replicatedLog) Replicate(msg string) {
	for _, replica := range r.replicas {
		r.replicate(replica, msg)
	}
}

func (r *replicatedLog) replicate(replica, msg string) {
	conn, err := grpc.Dial(replica, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error().Err(err).Msg("failed to dial to replica node")
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
		log.Error().Err(err).Msg("failed to replicate message")
	}
}
