package storage

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

type Store struct {
	*os.File
	mutex  sync.Mutex
	buffer *bufio.Writer
	size   uint64
}

func NewStore(file *os.File) (*Store, error) {
	dbFile, err := os.Stat(file.Name())
	if err != nil {
		return nil, err
	}

	size := uint64(dbFile.Size())
	return &Store{
		File:   file,
		buffer: bufio.NewWriter(file),
		size:   size,
	}, nil
}

// Write writes the given bytes to the store and returns the number of bytes written and the position of the write.
func (store *Store) Write(p []byte) (n uint64, position uint64, err error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	position = store.size
	if err := binary.Write(store.buffer, binary.BigEndian, uint64(len(p))); err != nil {
		return 0, 0, err
	}

	written, err := store.buffer.Write(p)
	if err != nil {
		return 0, 0, err
	}

	written += 8 // 8 bytes for the size of the data.

	store.size += uint64(written)
	return uint64(written), position, nil
}

// Read reads the bytes from the store at the given position.
func (store *Store) Read(position uint64) ([]byte, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if err := store.buffer.Flush(); err != nil {
		return nil, err
	}

	size := make([]byte, 8)
	if _, err := store.File.ReadAt(size, int64(position)); err != nil {
		return nil, err
	}

	data := make([]byte, binary.BigEndian.Uint64(size))
	if _, err := store.File.ReadAt(data, int64(position+8)); err != nil {
		return nil, err
	}
	return data, nil
}

// ReadAt reads the bytes from the store at the given offset.
func (store *Store) ReadAt(p []byte, offset int64) (int, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if err := store.buffer.Flush(); err != nil {
		return 0, err
	}
	return store.File.ReadAt(p, offset)
}

// Close flushes the buffer and closes the file. It persists any unwritten data to disk before closing the file.
func (store *Store) Close() error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	err := store.buffer.Flush()
	if err != nil {
		return err
	}
	return store.File.Close()
}
