package grpc_server

import (
	"context"
	"errors"
	"fmt"
	"github.com/replicated_log/api"
	"github.com/replicated_log/storage"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"net"
	"time"
)

type Server struct {
	api.UnimplementedReplicatorServer

	server *grpc.Server
	Delay  int
}

func (s *Server) Start() error {
	log.Info().Msg("Starting replication gRPC server")
	const port = 50051

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return errors.Join(err, errors.New(fmt.Sprintf("failed to listen to %d", port)))
	}

	server := grpc.NewServer()
	api.RegisterReplicatorServer(server, s)
	s.server = server

	return server.Serve(lis)
}

func (s *Server) Stop() error {
	s.server.Stop()
	return nil
}

func (s *Server) Replicate(_ context.Context, in *api.ReplicateRequest) (*api.ReplicateResponse, error) {
	timer := time.NewTimer(time.Duration(s.Delay) * time.Second)
	defer timer.Stop()

	<-timer.C
	storage.GetList().Add(in.GetMsg())
	return &api.ReplicateResponse{
		Msg: storage.GetList().Get(),
	}, nil
}
