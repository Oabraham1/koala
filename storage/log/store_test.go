package log

import (
	"encoding/binary"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadAndWrite(t *testing.T) {
	file, err := ioutil.TempFile("", "store_read_and_write_test")
	require.NoError(t, err)
	defer os.Remove(file.Name())

	store, err := NewStore(file)
	require.NoError(t, err)

	// Test the Write method.
	for i := uint64(1); i < 4; i++ {
		n, position, err := store.Write([]byte("hello world"))
		width := uint64(len([]byte("hello world"))) + 8
		require.NoError(t, err)
		require.Equal(t, position+n, width*i)
	}

	// Test the Read method.
	var position uint64
	for i := uint64(1); i < 4; i++ {
		read, err := store.Read(position)
		require.NoError(t, err)
		require.Equal(t, []byte("hello world"), read)
		position += uint64(len(read)) + 8
	}

	// Test the ReadAt method.
	for i, offset := uint64(1), int64(0); i < 4; i++ {
		bytes := make([]byte, 8)
		n, err := store.ReadAt(bytes, offset)
		require.NoError(t, err)
		offset += int64(n)

		size := binary.BigEndian.Uint64(bytes)
		bytes = make([]byte, size)
		n, err = store.ReadAt(bytes, offset)
		require.NoError(t, err)
		require.Equal(t, []byte("hello world"), bytes)
		require.Equal(t, int(size), n)
		offset += int64(n)
	}
}
