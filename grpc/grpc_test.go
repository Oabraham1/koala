package grpc

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"testing"

	"github.com/oabraham1/koala/authentication"
	configuration "github.com/oabraham1/koala/config"
	"github.com/oabraham1/koala/proto/v1"
	"github.com/oabraham1/koala/storage/index"
	log "github.com/oabraham1/koala/storage/log"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func setup(t *testing.T, fn func(*Config)) (rootClient proto.LogClient, nobodyClient proto.LogClient, config *Config, teardown func()) {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	newClient := func(certificatePath, keyPath string) (*grpc.ClientConn, proto.LogClient, []grpc.DialOption) {
		clientTLSConfig, err := configuration.SetupTLSConfiguration(configuration.TLSConfig{CAFile: configuration.CAFile, KeyFile: keyPath, CertificateFile: certificatePath, Server: false})
		require.NoError(t, err)

		clientTLSCredentials := credentials.NewTLS(clientTLSConfig)
		options := []grpc.DialOption{grpc.WithTransportCredentials(clientTLSCredentials)}
		connection, err := grpc.Dial(l.Addr().String(), options...)
		require.NoError(t, err)
		client := proto.NewLogClient(connection)
		return connection, client, options
	}

	var rootConnection *grpc.ClientConn
	rootConnection, rootClient, _ = newClient(configuration.RootClientCertificateFile, configuration.RootClientKeyFile)

	var nobodyConnection *grpc.ClientConn
	nobodyConnection, nobodyClient, _ = newClient(configuration.NobodyClientCertificateFile, configuration.NobodyClientKeyFile)

	serverTLSConfig, err := configuration.SetupTLSConfiguration(configuration.TLSConfig{
		CertificateFile: configuration.ServerCertFile,
		KeyFile:         configuration.ServerKeyFile,
		CAFile:          configuration.CAFile,
		ServerAddress:   l.Addr().String(),
		Server:          true,
	})
	require.NoError(t, err)

	serverCredentials := credentials.NewTLS(serverTLSConfig)

	directory, err := os.MkdirTemp("", "koala_grpc_test")
	require.NoError(t, err)

	clog, err := log.NewLog(directory, index.Config{})
	require.NoError(t, err)

	authorizer := authentication.NewAuthorizer(configuration.AccessControlModelFile, configuration.AccessControlPolicyFile)

	cfg := &Config{
		CommitLog:  clog,
		Authorizer: authorizer,
	}

	if fn != nil {
		fn(cfg)
	}
	server, err := NewGRPCServer(cfg, grpc.Creds(serverCredentials))
	require.NoError(t, err)

	go func() {
		server.Serve(l)
	}()

	return rootClient, nobodyClient, cfg, func() {
		server.Stop()
		rootConnection.Close()
		nobodyConnection.Close()
		l.Close()
	}
}

func TestGrpcServer(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T, rootClient proto.LogClient, nobodyClient proto.LogClient, config *Config){
		"produce/consume a message to/from the log succeeds": testProduceConsume,
		"produce/consume stream succeeds":                    testProduceConsumeStream,
		"consume past the end of the log fails":              testConsumePastEndOfLog,
		"unauthorized client fails to produce":               testUnauthorizedAccess,
	} {
		t.Run(scenario, func(t *testing.T) {
			rootClient,
				nobodyClient,
				config,
				teardown := setup(t, nil)
			defer teardown()
			fn(t, rootClient, nobodyClient, config)
		})
	}
}

func testProduceConsume(t *testing.T, client, _ proto.LogClient, config *Config) {
	ctx := context.Background()
	want := &proto.Data{Properties: json.RawMessage(`{"test": true}`), Timestamp: json.RawMessage(`"2020-01-01T00:00:00Z"`)}

	produce, err := client.Produce(ctx, &proto.ProduceRequest{Data: want})
	require.NoError(t, err)
	consume, err := client.Consume(ctx, &proto.ConsumeRequest{Offset: produce.Offset})
	require.NoError(t, err)
	require.Equal(t, want.Properties, consume.Data.Properties)
	require.Equal(t, produce.Offset, consume.Data.Offset)
}

func testProduceConsumeStream(t *testing.T, client, _ proto.LogClient, config *Config) {
	ctx := context.Background()
	want := []*proto.Data{
		{Properties: json.RawMessage(`{"test": true}`), Timestamp: json.RawMessage(`"2020-01-01T00:00:00Z"`), Offset: 0},
		{Properties: json.RawMessage(`{"test": true}`), Timestamp: json.RawMessage(`"2021-01-01T00:00:01Z"`), Offset: 1},
		{Properties: json.RawMessage(`{"test": true}`), Timestamp: json.RawMessage(`"2022-01-01T00:00:02Z"`), Offset: 2},
	}
	{
		stream, err := client.ProduceStream(ctx)
		require.NoError(t, err)

		for offset, data := range want {
			err = stream.Send(&proto.ProduceRequest{Data: data})
			require.NoError(t, err)
			response, err := stream.Recv()
			require.NoError(t, err)
			if response.Offset != uint64(offset) {
				t.Fatalf("got offset %d, want %d", response.Offset, offset)
			}
		}
	}
	{
		stream, err := client.ConsumeStream(ctx, &proto.ConsumeRequest{Offset: 0})
		require.NoError(t, err)

		for i, data := range want {
			response, err := stream.Recv()
			require.NoError(t, err)
			require.Equal(t, response.Data, &proto.Data{Properties: data.Properties, Timestamp: data.Timestamp, Offset: uint64(i)})
		}
	}
}

func testUnauthorizedAccess(t *testing.T, _, client proto.LogClient, config *Config) {
	ctx := context.Background()
	produce, err := client.Produce(ctx, &proto.ProduceRequest{Data: &proto.Data{Properties: json.RawMessage(`{"test": true}`)}})
	if produce != nil {
		t.Fatalf("got %v, want nil", produce)
	}
	gotCode, wantCode := status.Code(err), codes.PermissionDenied
	if gotCode != wantCode {
		t.Fatalf("got error code %s, want %s", gotCode, wantCode)
	}
	consume, err := client.Consume(ctx, &proto.ConsumeRequest{Offset: 0})
	if consume != nil {
		t.Fatalf("got %v, want nil", consume)
	}
	gotCode, wantCode = status.Code(err), codes.PermissionDenied
	if gotCode != wantCode {
		t.Fatalf("got error code %s, want %s", gotCode, wantCode)
	}

}

func testConsumePastEndOfLog(t *testing.T, client, _ proto.LogClient, config *Config) {
	want := &proto.Data{Properties: json.RawMessage(`{"test": true}`), Timestamp: json.RawMessage(`"2020-01-01T00:00:00Z"`)}
	produce, err := client.Produce(context.Background(), &proto.ProduceRequest{Data: want})
	require.NoError(t, err)

	consume, err := client.Consume(context.Background(), &proto.ConsumeRequest{Offset: produce.Offset})
	require.NoError(t, err)
	require.Equal(t, want.Properties, consume.Data.Properties)
	require.Equal(t, want.Timestamp, consume.Data.Timestamp)
	require.Equal(t, produce.Offset, consume.Data.Offset)
}
