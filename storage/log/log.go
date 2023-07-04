package log

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"

	proto "github.com/oabraham1/kola/proto/v1"
	"github.com/oabraham1/kola/storage/index"
	"github.com/oabraham1/kola/storage/segment"
	"github.com/oabraham1/kola/storage/store"
)

type Log struct {
	mutex         sync.RWMutex
	Directory     string
	Config        index.Config
	activeSegment *segment.Segment
	segments      []*segment.Segment
}

type OriginReader struct {
	*store.Store
	offset int64
}

func (log *Log) Setup() error {
	files, err := ioutil.ReadDir(log.Directory)
	if err != nil {
		return err
	}
	var baseOffsets []uint64
	for _, file := range files {
		offsetString := strings.TrimSuffix(file.Name(), path.Ext(file.Name()))
		offset, _ := strconv.ParseUint(offsetString, 10, 0)
		baseOffsets = append(baseOffsets, offset)
	}
	sort.Slice(baseOffsets, func(i, j int) bool {
		return baseOffsets[i] < baseOffsets[j]
	})
	for i := 0; i < len(baseOffsets); i++ {
		if err = log.NewSegment(baseOffsets[i]); err != nil {
			return err
		}
		i++
	}
	if log.segments == nil {
		if err = log.NewSegment(log.Config.Segment.InitialOffset); err != nil {
			return err
		}
	}
	return nil
}

func NewLog(directory string, config index.Config) (*Log, error) {
	if config.Segment.MaxStoreBytes == 0 {
		config.Segment.MaxStoreBytes = 1024
	}
	if config.Segment.MaxIndexBytes == 0 {
		config.Segment.MaxIndexBytes = 1024
	}
	log := &Log{
		Directory: directory,
		Config:    config,
	}
	return log, log.Setup()
}

func (log *Log) NewSegment(offset uint64) error {
	segment, err := segment.NewSegment(log.Directory, offset, log.Config)
	if err != nil {
		return err
	}
	log.segments = append(log.segments, segment)
	log.activeSegment = segment
	return nil
}

func (log *Log) Write(data *proto.Data) (uint64, error) {
	log.mutex.Lock()
	defer log.mutex.Unlock()
	offset, err := log.activeSegment.Write(data)
	if err != nil {
		return 0, err
	}
	if log.activeSegment.IsMaxed() {
		err = log.NewSegment(offset + 1)
	}
	return offset, err
}

func (log *Log) Read(offset uint64) (*proto.Data, error) {
	log.mutex.RLock()
	defer log.mutex.RUnlock()
	var segment *segment.Segment
	for _, s := range log.segments {
		if s.GetBaseOffset() <= offset && offset < s.GetNextOffset() {
			segment = s
			break
		}
	}
	if segment == nil || segment.GetNextOffset() <= offset {
		return nil, &proto.ErrorOffsetOutOfRange{Offset: offset}
	}
	return segment.Read(offset)
}

func (origin *OriginReader) Read(p []byte) (int, error) {
	n, err := origin.ReadAt(p, origin.offset)
	origin.offset += int64(n)
	return n, err
}

func (log *Log) ReadLowestOffset() (uint64, error) {
	log.mutex.RLock()
	defer log.mutex.RUnlock()
	return log.segments[0].GetBaseOffset(), nil
}

func (log *Log) ReadHighestOffset() (uint64, error) {
	log.mutex.RLock()
	defer log.mutex.RUnlock()
	offset := log.segments[len(log.segments)-1].GetNextOffset()
	if offset == 0 {
		return 0, nil
	}
	return offset - 1, nil
}

func (log *Log) TruncateLowest(offset uint64) error {
	log.mutex.Lock()
	defer log.mutex.Unlock()
	var segments []*segment.Segment
	for _, s := range log.segments {
		if s.GetNextOffset() <= offset+1 {
			if err := s.Remove(); err != nil {
				return err
			}
			continue
		}
		segments = append(segments, s)
	}
	log.segments = segments
	return nil
}

func (log *Log) Reader() io.Reader {
	log.mutex.RLock()
	defer log.mutex.RUnlock()
	readers := make([]io.Reader, len(log.segments))
	for i, segment := range log.segments {
		readers[i] = &OriginReader{segment.GetStore(), 0}
	}
	return io.MultiReader(readers...)
}

func (log *Log) Close() error {
	log.mutex.Lock()
	defer log.mutex.Unlock()
	for _, segment := range log.segments {
		if err := segment.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (log *Log) Remove() error {
	if err := log.Close(); err != nil {
		return err
	}
	return os.RemoveAll(log.Directory)
}

func (log *Log) Reset() error {
	if err := log.Remove(); err != nil {
		return err
	}
	return log.Setup()
}
