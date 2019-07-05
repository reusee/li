package li

import "sync/atomic"

type BufferID int64

type (
	Linebreak string
)

type Buffer struct {
	ID               BufferID
	Path             string
	LastSyncFileInfo FileInfo
	Linebreak        Linebreak
}

var nextBufferID int64

func NewBufferFromFile(
	path string,
	scope Scope,
	link Link,
) (
	buffer *Buffer,
	err error,
) {
	defer he(&err)

	id := BufferID(atomic.AddInt64(&nextBufferID, 1))
	var moment *Moment
	var linebreak Linebreak
	scope.Sub(func() string {
		return path
	}).Call(NewMomentFromFile, &moment, &linebreak, &err)
	ce(err)

	buffer = &Buffer{
		ID:               id,
		Path:             path,
		LastSyncFileInfo: moment.FileInfo,
		Linebreak:        linebreak,
	}
	link(buffer, moment)

	return
}

func NewBufferFromBytes(
	bs []byte,
	scope Scope,
	link Link,
) (
	buffer *Buffer,
	err error,
) {
	defer he(&err)

	id := BufferID(atomic.AddInt64(&nextBufferID, 1))
	var moment *Moment
	var linebreak Linebreak
	scope.Sub(func() []byte {
		return bs
	}).Call(NewMomentFromBytes, &moment, &linebreak, &err)
	ce(err)

	buffer = &Buffer{
		ID:        id,
		Linebreak: linebreak,
	}
	link(buffer, moment)

	return
}

type BufferConfig struct {
	ExpandTabs bool
	TabWidth   int
}

func (_ Provide) BufferConfig(
	getConfig GetConfig,
) BufferConfig {
	var config struct {
		Buffer BufferConfig
	}

	config.Buffer.ExpandTabs = true
	config.Buffer.TabWidth = 4

	ce(getConfig(&config))

	return config.Buffer
}

func NewBuffersFromPath(
	path string,
	scope Scope,
	link Link,
) (
	buffers []*Buffer,
	err error,
) {
	defer he(&err)

	var moments []*Moment
	var linebreaks []Linebreak
	var paths []string
	scope.Sub(func() string {
		return path
	}).Call(NewMomentsFromPath, &moments, &linebreaks, &paths, &err)
	ce(err)

	for i, moment := range moments {
		linebreak := linebreaks[i]
		id := BufferID(atomic.AddInt64(&nextBufferID, 1))
		buffer := &Buffer{
			ID:               id,
			Path:             paths[i],
			LastSyncFileInfo: moment.FileInfo,
			Linebreak:        linebreak,
		}
		buffers = append(buffers, buffer)
		link(buffer, moment)
	}

	return
}
