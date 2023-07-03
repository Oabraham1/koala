package storage

import (
	"encoding/binary"
	"io"
	"os"

	"github.com/tysonmote/gommap"
)

type Index struct {
	file *os.File
	mmap gommap.MMap
	size uint64
}

type Config struct {
	Segment struct {
		MaxStoreBytes uint64
		MaxIndexBytes uint64
		InitialOffset uint64
	}
}

func NewIndex(file *os.File, config Config) (*Index, error) {
	index := &Index{
		file: file,
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	index.size = uint64(fileInfo.Size())
	if err = os.Truncate(file.Name(), int64(config.Segment.MaxIndexBytes)); err != nil {
		return nil, err
	}

	if index.mmap, err = gommap.Map(index.file.Fd(), gommap.PROT_READ|gommap.PROT_WRITE, gommap.MAP_SHARED); err != nil {
		return nil, err
	}
	return index, nil
}

func (index *Index) Write(offset uint64, position uint64) error {
	if uint64(len(index.mmap)) < index.size+16 {
		return io.EOF
	}
	binary.BigEndian.PutUint64(index.mmap[index.size:index.size+8], offset)      // 8 represents the size of the offset
	binary.BigEndian.PutUint64(index.mmap[index.size+8:index.size+16], position) // 16 represents the size of the offset + position

	index.size += 16
	return nil
}

func (index *Index) Read(offset int64) (output uint64, position uint64, err error) {
	if index.size == 0 {
		return 0, 0, io.EOF
	}

	if offset == -1 {
		output = uint64((index.size / 16) - 1)
	} else {
		output = uint64(offset)
	}

	position = uint64(output) * 16
	if index.size < position+16 {
		return 0, 0, io.EOF
	}

	output = binary.BigEndian.Uint64(index.mmap[position : position+8])      // 8 represents the size of the offset
	position = binary.BigEndian.Uint64(index.mmap[position+8 : position+16]) // 16 represents the size of the offset + position
	return output, position, nil
}

func (index *Index) GetIndexName() string {
	return index.file.Name()
}

func (index *Index) Close() error {
	if err := index.mmap.Sync(gommap.MS_SYNC); err != nil {
		return err
	}

	if err := index.file.Sync(); err != nil {
		return err
	}

	if err := index.file.Truncate(int64(index.size)); err != nil {
		return err
	}

	return index.file.Close()
}
