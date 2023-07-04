package log

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	proto "github.com/oabraham1/kola/proto/v1"
	"github.com/oabraham1/kola/storage/index"
	"github.com/stretchr/testify/require"
	protoc "google.golang.org/protobuf/proto"
)

const (
	lenWidth = 8
)

func TestLog(t *testing.T) {
	for scenerio, fn := range map[string]func(t *testing.T, log *Log){
		"write record succeeds": testLogWrite,
		"read record succeeds":  testLogRead,
		"offset out of range":   testLogOutOfRange,
		"init with existing":    testLogInitExisting,
		"read":                  testLogReader,
		"truncate_lowest":       testLogTruncate,
	} {
		t.Run(scenerio, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "Testing Log")
			require.NoError(t, err)
			defer os.RemoveAll(dir)
			config := index.Config{}
			config.Segment.MaxStoreBytes = 32
			log, err := NewLog(dir, config)
			require.NoError(t, err)

			fn(t, log)
		})
	}
}

func testLogWrite(t *testing.T, log *Log) {
	append := &proto.Data{Properties: json.RawMessage(`{"test": "test"}`)}
	offset, err := log.Write(append)
	require.NoError(t, err)
	require.Equal(t, uint64(0), offset)
}

func testLogRead(t *testing.T, log *Log) {
	append := &proto.Data{Properties: json.RawMessage(`{"test": "test"}`)}
	offset, err := log.Write(append)
	require.NoError(t, err)
	require.Equal(t, uint64(0), offset)

	read, err := log.Read(offset)
	require.NoError(t, err)
	require.Equal(t, append.Properties, read.Properties)
}

func testLogOutOfRange(t *testing.T, log *Log) {
	read, err := log.Read(100)
	require.Nil(t, read)
	require.Error(t, err)
}

func testLogInitExisting(t *testing.T, log *Log) {
	append := &proto.Data{Properties: json.RawMessage(`{"test": "test"}`)}
	for i := 0; i < 3; i++ {
		_, err := log.Write(append)
		require.NoError(t, err)
	}
	require.NoError(t, log.Close())

	offset, err := log.ReadLowestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(0), offset)
	offset, err = log.ReadHighestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(2), offset)

	n, err := NewLog(log.Directory, log.Config)
	require.NoError(t, err)

	offset, err = n.ReadLowestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(0), offset)
	offset, err = n.ReadHighestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(2), offset)
}

func testLogReader(t *testing.T, log *Log) {
	append := &proto.Data{Properties: json.RawMessage(`{"test": "test"}`)}
	offset, err := log.Write(append)
	require.NoError(t, err)
	require.Equal(t, uint64(0), offset)

	reader := log.Reader()
	buffer, error := ioutil.ReadAll(reader)
	require.NoError(t, error)

	record := &proto.Data{}
	err = protoc.Unmarshal(buffer[lenWidth:], record)
	require.NoError(t, err)
	require.Equal(t, append.Properties, record.Properties)
}

func testLogTruncate(t *testing.T, log *Log) {
	append := &proto.Data{Properties: json.RawMessage(`{"test": "test"}`)}
	for i := 0; i < 3; i++ {
		_, err := log.Write(append)
		require.NoError(t, err)
	}

	err := log.TruncateLowest(1)
	require.NoError(t, err)

	_, err = log.Read(0)
	require.Error(t, err)
}
