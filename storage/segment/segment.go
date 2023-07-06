package segment

import (
	"fmt"
	"os"
	"path"

	proto "github.com/oabraham1/koala/proto/v1"
	"github.com/oabraham1/koala/storage/index"
	"github.com/oabraham1/koala/storage/store"
	protoc "google.golang.org/protobuf/proto"
)

// Segment represents a segment of the log
type Segment struct {
	store      *store.Store
	index      *index.Index
	baseOffset uint64
	nextOffset uint64
	config     index.Config
}

// NewSegment creates a new segment
func NewSegment(directory string, baseOffset uint64, config index.Config) (*Segment, error) {
	segment := &Segment{
		baseOffset: baseOffset,
		config:     config,
	}
	var err error
	storeFile, err := os.OpenFile(
		path.Join(directory, fmt.Sprintf("%d%s", baseOffset, ".store")),
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, err
	}
	if segment.store, err = store.NewStore(storeFile); err != nil {
		return nil, err
	}
	indexFile, err := os.OpenFile(
		path.Join(directory, fmt.Sprintf("%d%s", baseOffset, ".index")),
		os.O_RDWR|os.O_CREATE,
		0644,
	)
	if err != nil {
		return nil, err
	}
	if segment.index, err = index.NewIndex(indexFile, config); err != nil {
		return nil, err
	}
	if offset, _, err := segment.index.Read(-1); err != nil {
		segment.nextOffset = baseOffset
	} else {
		segment.nextOffset = baseOffset + uint64(offset) + 1
	}
	return segment, nil
}

// Write writes a record to the segment
func (segment *Segment) Write(record *proto.Data) (offset uint64, err error) {
	cursor := segment.nextOffset
	record.Offset = cursor
	proto, err := protoc.Marshal(record)
	if err != nil {
		return 0, err
	}
	_, position, err := segment.store.Write(proto)
	if err != nil {
		return 0, err
	}
	if err = segment.index.Write(
		uint32(segment.nextOffset-uint64(segment.baseOffset)),
		position,
	); err != nil {
		return 0, err
	}
	segment.nextOffset++
	return cursor, nil
}

// Read reads a record from the segment
func (segment *Segment) Read(offset uint64) (*proto.Data, error) {
	_, position, err := segment.index.Read(int64(offset - segment.baseOffset))
	if err != nil {
		return nil, err
	}
	pos, err := segment.store.Read(position)
	if err != nil {
		return nil, err
	}
	record := &proto.Data{}
	err = protoc.Unmarshal(pos, record)
	return record, err
}

// GetStore returns the store of the segment
func (segment *Segment) GetStore() *store.Store {
	return segment.store
}

// IsMaxed returns true if the segment is maxed out
func (segment *Segment) IsMaxed() bool {
	return segment.store.GetSize() >= segment.config.Segment.MaxStoreBytes ||
		segment.index.GetSize() >= segment.config.Segment.MaxIndexBytes
}

// GetNextOffset returns the next offset of the segment
func (segment *Segment) GetNextOffset() uint64 {
	return segment.nextOffset
}

// GetBaseOffset returns the base offset of the segment
func (segment *Segment) GetBaseOffset() uint64 {
	return segment.baseOffset
}

// Close closes the segment
func (segment *Segment) Close() error {
	if err := segment.index.Close(); err != nil {
		return err
	}
	return segment.store.Close()
}

// Remove removes the segment
func (segment *Segment) Remove() error {
	if err := segment.Close(); err != nil {
		return err
	}
	if err := os.Remove(segment.index.GetIndexName()); err != nil {
		return err
	}
	return os.Remove(segment.store.Name())
}
