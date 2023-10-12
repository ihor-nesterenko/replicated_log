// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.24.3
// source: api.proto

package api

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ReplicatorClient is the client API for Replicator service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ReplicatorClient interface {
	Replicate(ctx context.Context, in *ReplicateRequest, opts ...grpc.CallOption) (*ReplicateResponse, error)
}

type replicatorClient struct {
	cc grpc.ClientConnInterface
}

func NewReplicatorClient(cc grpc.ClientConnInterface) ReplicatorClient {
	return &replicatorClient{cc}
}

func (c *replicatorClient) Replicate(ctx context.Context, in *ReplicateRequest, opts ...grpc.CallOption) (*ReplicateResponse, error) {
	out := new(ReplicateResponse)
	err := c.cc.Invoke(ctx, "/api.Replicator/Replicate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ReplicatorServer is the server API for Replicator service.
// All implementations must embed UnimplementedReplicatorServer
// for forward compatibility
type ReplicatorServer interface {
	Replicate(context.Context, *ReplicateRequest) (*ReplicateResponse, error)
	mustEmbedUnimplementedReplicatorServer()
}

// UnimplementedReplicatorServer must be embedded to have forward compatible implementations.
type UnimplementedReplicatorServer struct {
}

func (UnimplementedReplicatorServer) Replicate(context.Context, *ReplicateRequest) (*ReplicateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Replicate not implemented")
}
func (UnimplementedReplicatorServer) mustEmbedUnimplementedReplicatorServer() {}

// UnsafeReplicatorServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ReplicatorServer will
// result in compilation errors.
type UnsafeReplicatorServer interface {
	mustEmbedUnimplementedReplicatorServer()
}

func RegisterReplicatorServer(s grpc.ServiceRegistrar, srv ReplicatorServer) {
	s.RegisterService(&Replicator_ServiceDesc, srv)
}

func _Replicator_Replicate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReplicateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplicatorServer).Replicate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.Replicator/Replicate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplicatorServer).Replicate(ctx, req.(*ReplicateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Replicator_ServiceDesc is the grpc.ServiceDesc for Replicator service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Replicator_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.Replicator",
	HandlerType: (*ReplicatorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Replicate",
			Handler:    _Replicator_Replicate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api.proto",
}
