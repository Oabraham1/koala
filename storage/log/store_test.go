package log

import (
	"encoding/binary"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func openFile(name string) (*os.File, int64, error) {
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return nil, 0, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, 0, err
	}

	return file, fileInfo.Size(), nil
}

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

func TestClose(t *testing.T) {
	file, err := ioutil.TempFile("", "store_close_test")
	require.NoError(t, err)
	defer os.Remove(file.Name())

	store, err := NewStore(file)
	require.NoError(t, err)

	_, _, err = store.Write([]byte("hello world"))
	require.NoError(t, err)

	file, size_one, err := openFile(file.Name())
	require.NoError(t, err)

	// Test the Close method.
	err = store.Close()
	require.NoError(t, err)

	_, size_two, err := openFile(file.Name())
	require.NoError(t, err)
	require.True(t, size_two > size_one)

	// Test that the file was closed.
	_, err = store.File.Stat()
	require.Error(t, err)
}
