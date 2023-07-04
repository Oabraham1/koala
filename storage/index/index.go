package index

import (
	"encoding/binary"
	"io"
	"os"

	"github.com/tysonmote/gommap"
)

var (
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth        = offWidth + posWidth
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

func (index *Index) Write(offset uint32, position uint64) error {
	if uint64(len(index.mmap)) < index.size+entWidth {
		return io.EOF
	}
	binary.BigEndian.PutUint32(index.mmap[index.size:index.size+offWidth], offset)            // 4 represents the size of the offset
	binary.BigEndian.PutUint64(index.mmap[index.size+offWidth:index.size+entWidth], position) // 12 represents the size of the offset + position

	index.size += uint64(entWidth)
	return nil
}

func (index *Index) Read(input int64) (output uint32, position uint64, err error) {
	if index.size == 0 {
		return 0, 0, io.EOF
	}

	if input == -1 {
		output = uint32((index.size / entWidth) - 1)
	} else {
		output = uint32(input)
	}

	position = uint64(output) * entWidth
	if index.size < position+entWidth {
		return 0, 0, io.EOF
	}

	output = binary.BigEndian.Uint32(index.mmap[position : position+offWidth])            // 4 represents the size of the offset
	position = binary.BigEndian.Uint64(index.mmap[position+offWidth : position+entWidth]) // 12 represents the size of the offset + position
	return output, position, nil
}

func (index *Index) GetIndexName() string {
	return index.file.Name()
}

func (index *Index) GetSize() uint64 {
	return index.size
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
