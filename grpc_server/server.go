package grpc_server

import (
	"context"
	"github.com/replicated_log/api"
	"github.com/replicated_log/storage"
)

type Server struct {
	api.UnimplementedReplicatorServer
}

func (s *Server) Replicate(_ context.Context, in *api.ReplicateRequest) (*api.ReplicateResponse, error) {
	storage.GetList().Add(in.GetMsg())
	return &api.ReplicateResponse{
		Msg: storage.GetList().Get(),
	}, nil
}
