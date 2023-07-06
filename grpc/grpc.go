package grpc

import (
	"context"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/oabraham1/koala/proto/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// Config contains the configuration for gRPC
type Config struct {
	CommitLog  CommitLog
	Authorizer Authorizer
}

// GRPCServer is a wrapper around proto.LogServer and Config
type Server struct {
	proto.UnimplementedLogServer
	*Config
}

// SubjectContextKey returns the subject from the context
type SubjectContextKey struct{}

// CommitLog is the interface that wraps the basic Write and Read methods
type CommitLog interface {
	Write(*proto.Data) (uint64, error)
	Read(uint64) (*proto.Data, error)
}

// Authorizer is the interface that wraps the basic Authorize method
type Authorizer interface {
	Authorize(subject, object, action string) error
}

var _ proto.LogServer = (*Server)(nil)

const (
	objectWildcard = "*"
	produceAction  = "produce"
	consumeAction  = "consume"
)

// newGRPCServer creates a new GRPCServer
func newGRPCServer(config *Config) (grpcServer *Server, err error) {
	grpcServer = &Server{
		Config: config,
	}
	return grpcServer, nil
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(config *Config, options ...grpc.ServerOption) (*grpc.Server, error) {
	options = append(options, grpc.StreamInterceptor(
		grpc_middleware.ChainStreamServer(grpc_auth.StreamServerInterceptor(Authenticate))), grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(grpc_auth.UnaryServerInterceptor(Authenticate))))

	grpcServer := grpc.NewServer(options...)
	server, err := newGRPCServer(config)
	if err != nil {
		return nil, err
	}
	proto.RegisterLogServer(grpcServer, server)
	return grpcServer, nil
}

// Produce is the implementation of the Produce RPC
func (server *Server) Produce(ctx context.Context, request *proto.ProduceRequest) (*proto.ProduceResponse, error) {
	if err := server.Authorizer.Authorize(Subject(ctx), objectWildcard, produceAction); err != nil {
		return nil, err
	}
	offset, err := server.CommitLog.Write(request.Data)
	if err != nil {
		return nil, err
	}
	return &proto.ProduceResponse{Offset: offset}, nil
}

// Consume is the implementation of the Consume RPC
func (server *Server) Consume(ctx context.Context, request *proto.ConsumeRequest) (*proto.ConsumeResponse, error) {
	if err := server.Authorizer.Authorize(Subject(ctx), objectWildcard, consumeAction); err != nil {
		return nil, err
	}
	data, err := server.CommitLog.Read(request.Offset)
	if err != nil {
		return nil, err
	}
	return &proto.ConsumeResponse{Data: data}, nil
}

// ProduceStream is the implementation of the ProduceStream RPC
func (server *Server) ProduceStream(stream proto.Log_ProduceStreamServer) error {
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

// ConsumeStream is the implementation of the ConsumeStream RPC
func (server *Server) ConsumeStream(request *proto.ConsumeRequest, stream proto.Log_ConsumeStreamServer) error {
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

// Authenticate authenticates the peer and returns the context
func Authenticate(ctx context.Context) (context.Context, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return ctx, status.New(codes.Unknown, "no peer found").Err()
	}
	if peer.AuthInfo == nil {
		return context.WithValue(ctx, SubjectContextKey{}, ""), nil
	}

	transportLayerSecurityInfo := peer.AuthInfo.(credentials.TLSInfo)
	subject := transportLayerSecurityInfo.State.VerifiedChains[0][0].Subject.CommonName
	return context.WithValue(ctx, SubjectContextKey{}, subject), nil
}

// Subject returns the subject from the context
func Subject(ctx context.Context) string {
	return ctx.Value(SubjectContextKey{}).(string)
}
