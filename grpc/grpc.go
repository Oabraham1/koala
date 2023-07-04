package grpc

import (
	"context"

	"github.com/oabraham1/kola/proto/v1"
	"google.golang.org/grpc"
)

type Config struct {
	CommitLog CommitLog
}

var _ proto.LogServer = (*GRPCServer)(nil)

type CommitLog interface {
	Write(*proto.Data) (uint64, error)
	Read(uint64) (*proto.Data, error)
}

type GRPCServer struct {
	proto.UnimplementedLogServer
	*Config
}

func newGRPCServer(config *Config) (grpcServer *GRPCServer, err error) {
	grpcServer = &GRPCServer{
		Config: config,
	}
	return grpcServer, nil
}

func NewGRPCServer(config *Config) (*grpc.Server, error) {
	grpcServer := grpc.NewServer()
	server, err := newGRPCServer(config)
	if err != nil {
		return nil, err
	}
	proto.RegisterLogServer(grpcServer, server)
	return grpcServer, nil
}

func (server *GRPCServer) Produce(ctx context.Context, request *proto.ProduceRequest) (*proto.ProduceResponse, error) {
	offset, err := server.CommitLog.Write(request.Data)
	if err != nil {
		return nil, err
	}
	return &proto.ProduceResponse{Offset: offset}, nil
}

func (server *GRPCServer) Consume(ctx context.Context, request *proto.ConsumeRequest) (*proto.ConsumeResponse, error) {
	data, err := server.CommitLog.Read(request.Offset)
	if err != nil {
		return nil, err
	}
	return &proto.ConsumeResponse{Data: data}, nil
}

func (server *GRPCServer) ProduceStream(stream proto.Log_ProduceStreamServer) error {
	for {
		request, err := stream.Recv()
		if err != nil {
			return err
		}
		response, err := server.Produce(stream.Context(), request)
		if err != nil {
			return err
		}
		if err := stream.Send(response); err != nil {
			return err
		}
	}
}

func (server *GRPCServer) ConsumeStream(request *proto.ConsumeRequest, stream proto.Log_ConsumeStreamServer) error {
	for {
		select {
		case <-stream.Context().Done():
			return nil
		default:
			response, err := server.Consume(stream.Context(), request)
			switch err.(type) {
			case nil:
			case *proto.ErrorOffsetOutOfRange:
				continue
			default:
				return err
			}
			if err := stream.Send(response); err != nil {
				return err
			}
			request.Offset++
		}
	}
}
