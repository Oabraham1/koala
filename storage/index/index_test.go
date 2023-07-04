package index

import (
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"testing"
)

func TestIndex(t *testing.T) {
	file, err := os.CreateTemp(os.TempDir(), "index_test")
	require.NoError(t, err)
	defer os.Remove(file.Name())

	config := Config{}
	config.Segment.MaxIndexBytes = 1024
	index, err := NewIndex(file, config)
	require.NoError(t, err)
	_, _, err = index.Read(-1)
	require.Error(t, err)
	require.Equal(t, file.Name(), index.GetIndexName())

	entries := []struct {
		Offset   uint32
		Position uint64
	}{
		{Offset: 0, Position: 0},
		{Offset: 1, Position: 10},
	}

	for _, want := range entries {
		err = index.Write(want.Offset, want.Position)
		require.NoError(t, err)

		_, position, err := index.Read(int64(want.Offset))
		require.NoError(t, err)
		require.Equal(t, want.Position, position)
	}

	_, _, err = index.Read(int64(len(entries)))
	require.Equal(t, io.EOF, err)
	_ = index.Close()

	file, _ = os.OpenFile(file.Name(), os.O_RDWR, 0600)
	index, err = NewIndex(file, config)
	require.NoError(t, err)
	offset, position, err := index.Read(-1)
	require.NoError(t, err)
	require.Equal(t, uint32(1), offset)
	require.Equal(t, entries[1].Position, position)
}
