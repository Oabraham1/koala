package segment

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"testing"

	protoutil "github.com/oabraham1/kola/proto/v1"
	"github.com/oabraham1/kola/storage/index"
	"github.com/stretchr/testify/require"
)

var (
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth        = offWidth + posWidth
)

func TestSegment(t *testing.T) {
	directory, _ := ioutil.TempDir("", "Testing Segment file")
	defer os.Remove(directory)

	want := &protoutil.Data{Properties: json.RawMessage(`{"test": "test"}`)}

	config := index.Config{}
	config.Segment.MaxStoreBytes = 1024
	config.Segment.MaxIndexBytes = entWidth * 3

	segment, err := NewSegment(directory, 16, config)
	require.NoError(t, err)
	require.Equal(t, uint64(16), segment.nextOffset, segment.nextOffset)
	require.False(t, segment.IsMaxed())

	for i := uint64(0); i < 3; i++ {
		offset, err := segment.Write(want)
		require.NoError(t, err)
		require.Equal(t, 16+i, offset)

		got, err := segment.Read(offset)
		require.NoError(t, err)
		require.Equal(t, want.Properties, got.Properties)
	}

	_, err = segment.Write(want)
	require.Equal(t, io.EOF, err)

	require.True(t, segment.IsMaxed())

	config.Segment.MaxStoreBytes = uint64(len(want.Properties) * 3)
	config.Segment.MaxIndexBytes = 1024

	segment, err = NewSegment(directory, 16, config)
	require.NoError(t, err)
	require.True(t, segment.IsMaxed())

	err = segment.Remove()
	require.NoError(t, err)
	segment, err = NewSegment(directory, 16, config)
	require.NoError(t, err)
	require.False(t, segment.IsMaxed())
}
