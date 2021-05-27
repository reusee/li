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

type EvBufferCreated struct {
	Buffer *Buffer
	Moment *Moment
}

type NewBufferFromFile func(
	path string,
) (
	buffer *Buffer,
	err error,
)

func (_ Provide) NewBufferFromFile(
	scope Scope,
	link Link,
	trigger Trigger,
	newMoment NewMomentFromFile,
) NewBufferFromFile {
	return func(path string) (buffer *Buffer, err error) {
		defer he(&err)

		id := BufferID(atomic.AddInt64(&nextBufferID, 1))
		moment, linebreak, err := newMoment(path)
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

		trigger(EvBufferCreated{
			Buffer: buffer,
			Moment: moment,
		})

		return
	}
}

type NewBufferFromBytes func(
	bs []byte,
) (
	buffer *Buffer,
	err error,
)

func (_ Provide) NewBufferFromBytes(
	link Link,
	trigger Trigger,
	newMoment NewMomentFromBytes,
) NewBufferFromBytes {
	return func(bs []byte) (buffer *Buffer, err error) {
		defer he(&err)

		id := BufferID(atomic.AddInt64(&nextBufferID, 1))
		var moment *Moment
		var linebreak Linebreak
		moment, linebreak, err = newMoment(bs)
		ce(err)

		buffer = &Buffer{
			ID:        id,
			Linebreak: linebreak,
		}
		link(buffer, moment)

		trigger(EvBufferCreated{
			Buffer: buffer,
			Moment: moment,
		})

		return
	}
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

type NewBuffersFromPath func(
	path string,
) (
	buffers []*Buffer,
	err error,
)

func (_ Provide) NewBuffersFromPath(
	scope Scope,
	link Link,
	newMoment NewMomentsFromPath,
) NewBuffersFromPath {
	return func(path string) (buffers []*Buffer, err error) {
		defer he(&err)

		moments, linebreaks, paths, err := newMoment(path)
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
}

type EvBufferLanguageChanged struct {
	Buffer  *Buffer
	OldLang Language
	NewLang Language
}

func (b *Buffer) SetLanguage(scope Scope, lang Language) {
	oldLang := b.language
	b.language = lang
	if oldLang != lang {
		scope.Call(func(
			trigger Trigger,
		) {
			trigger(EvBufferLanguageChanged{
				Buffer:  b,
				OldLang: oldLang,
				NewLang: lang,
			})
		})
	}
}
