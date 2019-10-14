package li

import (
	"path/filepath"
	"sync/atomic"
)

type BufferID int64

type (
	Linebreak string
)

type Buffer struct {
	ID               BufferID
	Path             string
	AbsPath          string
	AbsDir           string
	LastSyncFileInfo FileInfo
	Linebreak        Linebreak
	language         Language
}

var nextBufferID int64

type evBufferCreated struct{}

var EvBufferCreated = new(evBufferCreated)

func NewBufferFromFile(
	path string,
	scope Scope,
	link Link,
	trigger Trigger,
) (
	buffer *Buffer,
	err error,
) {
	defer he(&err)

	id := BufferID(atomic.AddInt64(&nextBufferID, 1))
	var moment *Moment
	var linebreak Linebreak
	scope.Sub(
		&path,
	).Call(NewMomentFromFile, &moment, &linebreak, &err)
	ce(err)

	absPath, err := filepath.Abs(path)
	ce(err)
	buffer = &Buffer{
		ID:               id,
		Path:             path,
		AbsPath:          absPath,
		AbsDir:           filepath.Dir(absPath),
		LastSyncFileInfo: moment.FileInfo,
		Linebreak:        linebreak,
	}
	link(buffer, moment)
	buffer.SetLanguage(scope, LanguageFromPath(path))

	trigger(scope.Sub(
		&buffer, &moment,
	), EvBufferCreated)

	return
}

func NewBufferFromBytes(
	bs []byte,
	scope Scope,
	link Link,
	trigger Trigger,
) (
	buffer *Buffer,
	err error,
) {
	defer he(&err)

	id := BufferID(atomic.AddInt64(&nextBufferID, 1))
	var moment *Moment
	var linebreak Linebreak
	scope.Sub(
		&bs,
	).Call(NewMomentFromBytes, &moment, &linebreak, &err)
	ce(err)

	buffer = &Buffer{
		ID:        id,
		Linebreak: linebreak,
	}
	link(buffer, moment)

	trigger(scope.Sub(
		&buffer, &moment,
	), EvBufferCreated)

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
	scope.Sub(
		&path,
	).Call(NewMomentsFromPath, &moments, &linebreaks, &paths, &err)
	ce(err)

	for i, moment := range moments {
		linebreak := linebreaks[i]
		id := BufferID(atomic.AddInt64(&nextBufferID, 1))
		absPath, err := filepath.Abs(paths[i])
		ce(err)
		buffer := &Buffer{
			ID:               id,
			Path:             paths[i],
			AbsPath:          absPath,
			AbsDir:           filepath.Dir(absPath),
			LastSyncFileInfo: moment.FileInfo,
			Linebreak:        linebreak,
		}
		link(buffer, moment)
		buffer.SetLanguage(scope, LanguageFromPath(paths[i]))
		buffers = append(buffers, buffer)
	}

	return
}

type evBufferLanguageChanged struct{}

var EvBufferLanguageChanged = new(evBufferLanguageChanged)

func (b *Buffer) SetLanguage(scope Scope, lang Language) {
	oldLang := b.language
	b.language = lang
	if oldLang != lang {
		scope.Call(func(
			trigger Trigger,
		) {
			trigger(scope.Sub(
				&b, &[2]Language{oldLang, lang},
			), EvBufferLanguageChanged)
		})
	}
}
