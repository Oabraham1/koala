package grpc

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"testing"

	"github.com/oabraham1/kola/proto/v1"
	"github.com/oabraham1/kola/storage/index"
	log "github.com/oabraham1/kola/storage/log"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func setup(t *testing.T, fn func(*Config)) (client proto.LogClient, config *Config, teardown func()) {
	t.Helper()

	l, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	clientOptions := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial(l.Addr().String(), clientOptions...)
	require.NoError(t, err)

	directory, err := ioutil.TempDir("", "grpc_server_test")
	require.NoError(t, err)

	clog, err := log.NewLog(directory, index.Config{})
	require.NoError(t, err)

	cfg := &Config{
		CommitLog: clog,
	}
	if fn != nil {
		fn(cfg)
	}
	server, err := NewGRPCServer(cfg)
	require.NoError(t, err)

	go func() {
		server.Serve(l)
	}()

	client = proto.NewLogClient(conn)
	return client, cfg, func() {
		server.Stop()
		conn.Close()
		l.Close()
		clog.Remove()
	}
}

func TestGrpcServer(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T, client proto.LogClient, config *Config){
		"produce/consume a message to/from the log succeeds": testProduceConsume,
		"produce/consume stream succeeds":                    testProduceConsumeStream,
		"consume past the end of the log fails":              testConsumePastEndOfLog,
	} {
		t.Run(scenario, func(t *testing.T) {
			client, config, teardown := setup(t, nil)
			defer teardown()
			fn(t, client, config)
		})
	}
}

func testProduceConsume(t *testing.T, client proto.LogClient, config *Config) {
	ctx := context.Background()
	want := &proto.Data{Properties: json.RawMessage(`{"test": true}`), Timestamp: json.RawMessage(`"2020-01-01T00:00:00Z"`)}

	produce, err := client.Produce(ctx, &proto.ProduceRequest{Data: want})
	require.NoError(t, err)
	consume, err := client.Consume(ctx, &proto.ConsumeRequest{Offset: produce.Offset})
	require.NoError(t, err)
	require.Equal(t, want.Properties, consume.Data.Properties)
	require.Equal(t, produce.Offset, consume.Data.Offset)
}

func testProduceConsumeStream(t *testing.T, client proto.LogClient, config *Config) {
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

func testConsumePastEndOfLog(t *testing.T, client proto.LogClient, config *Config) {
	want := &proto.Data{Properties: json.RawMessage(`{"test": true}`), Timestamp: json.RawMessage(`"2020-01-01T00:00:00Z"`)}
	produce, err := client.Produce(context.Background(), &proto.ProduceRequest{Data: want})
	require.NoError(t, err)

	consume, err := client.Consume(context.Background(), &proto.ConsumeRequest{Offset: produce.Offset})
	require.NoError(t, err)
	require.Equal(t, want.Properties, consume.Data.Properties)
	require.Equal(t, want.Timestamp, consume.Data.Timestamp)
	require.Equal(t, produce.Offset, consume.Data.Offset)
}
