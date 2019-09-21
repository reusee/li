package li

import (
	"C"
	"encoding"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"
	"unsafe"

	"github.com/reusee/li/treesitter"
)

type MomentID int64

type Moment struct {
	T0 time.Time

	ID       MomentID
	Previous *Moment
	Change   Change
	lines    []*Line

	// hash state before line <key>
	subContentHashStates []*[]byte
	// hash sum of line [0, <key>]
	subContentHashes []*HashSum

	FileInfo FileInfo

	initContentOnce        sync.Once
	content                string
	initLowerContentOnce   sync.Once
	lowerContent           string
	initCStringContentOnce sync.Once
	cstringContent         *C.char

	Language       Language
	initParserOnce sync.Once
	parser         *treesitter.Parser
	syntaxAttrs    sync.Map

	finalizeFuncs sync.Map
}

func NewMoment(prev *Moment) *Moment {
	m := &Moment{
		T0:       time.Now(),
		ID:       MomentID(atomic.AddInt64(&nextMomentID, 1)),
		Previous: prev,
	}
	if prev != nil {
		m.Language = prev.Language
	}
	runtime.SetFinalizer(m, func(m *Moment) {
		m.finalizeFuncs.Range(func(_, v any) bool {
			v.(func())()
			return true
		})
	})
	return m
}

func (m *Moment) GetLine(i int) *Line {
	if i < 0 {
		return nil
	}
	if i >= m.NumLines() {
		return nil
	}
	line := m.lines[i]
	line.init()
	return line
}

func (m *Moment) GetContent() string {
	m.initContentOnce.Do(func() {
		var b strings.Builder
		for _, line := range m.lines {
			b.WriteString(line.content)
		}
		m.content = b.String()
	})
	return m.content
}

func (m *Moment) GetLowerContent() string {
	m.initLowerContentOnce.Do(func() {
		content := m.GetContent()
		m.lowerContent = strings.ToLower(content)
	})
	return m.lowerContent
}

func (m *Moment) GetCStringContent() *C.char {
	m.initCStringContentOnce.Do(func() {
		content := C.CString(m.GetContent())
		m.finalizeFuncs.Store(rand.Int63(), func() {
			cfree(unsafe.Pointer(m.cstringContent))
		})
		m.cstringContent = content
	})
	return m.cstringContent
}

func (m *Moment) GetParser() *treesitter.Parser {
	if m.Language == LanguageUnknown {
		return nil
	}
	m.initParserOnce.Do(func() {
		//TODO utilize tree-sitter incremental parsing
		if fn, ok := languageParsers[m.Language]; ok {
			m.parser = fn(m)
		}
		m.finalizeFuncs.Store(rand.Int63(), func() {
			m.parser.Close()
		})
	})
	return m.parser
}

func (m *Moment) GetSyntaxAttr(lineNum int, runeOffset int) string {
	key := Position{
		Line: lineNum,
		Rune: runeOffset,
	}
	if v, ok := m.syntaxAttrs.Load(key); ok {
		return v.(string)
	}
	parser := m.GetParser()
	if parser == nil {
		return ""
	}
	node := parser.NodeAt(treesitter.Point(lineNum, runeOffset))
	nodeType := treesitter.NodeType(node)
	attr := nodeType
	m.syntaxAttrs.Store(key, attr)
	return attr
}

func (m *Moment) NumLines() int {
	return len(m.lines)
}

func (m *Moment) GetSubContentHash(lineNum int) HashSum {
	p := m.subContentHashes[lineNum]
	if p == nil {
		m.hashSubContent(lineNum)
		p = m.subContentHashes[lineNum]
	}
	return *p
}

func (m *Moment) hashSubContent(lineNum int) {
	h := NewHash()
	if lineNum > 0 {
		p := m.subContentHashStates[lineNum-1]
		if p == nil {
			m.hashSubContent(lineNum - 1)
			p = m.subContentHashStates[lineNum-1]
		}
		packedState := *p
		ce(h.(encoding.BinaryUnmarshaler).UnmarshalBinary(packedState))
	}
	h.Write([]byte(m.lines[lineNum].content))
	sum := h.Sum(nil)
	var hSum HashSum
	copy(hSum[:], sum[:])
	m.subContentHashes[lineNum] = &hSum
	packedState, err := h.(encoding.BinaryMarshaler).MarshalBinary()
	ce(err)
	m.subContentHashStates[lineNum] = &packedState
}

type Line struct {
	Cells          []Cell
	Runes          []rune
	DisplayWidth   int
	AllSpace       bool
	NonSpaceOffset *int

	content  string
	initOnce *sync.Once
	config   *BufferConfig
}

func (l *Line) init() {
	l.initOnce.Do(func() {
		var cells []Cell
		allSpace := true
		byteOffset := 0
		l.Runes = []rune(l.content)
		var nonSpaceOffset *int
		for i, r := range l.Runes {
			width := runeWidth(r)
			var displayWidth int
			if r == '\t' && l.config.ExpandTabs {
				displayWidth = l.config.TabWidth
			} else {
				displayWidth = width
			}
			cell := Cell{
				Rune:         r,
				RuneLen:      len(string(r)),
				RuneWidth:    width,
				DisplayWidth: displayWidth,
				ByteOffset:   byteOffset,
				RuneOffset:   i,
			}
			cells = append(cells, cell)
			l.DisplayWidth += cell.DisplayWidth
			if !unicode.IsSpace(r) {
				allSpace = false
				if nonSpaceOffset == nil {
					offset := byteOffset
					nonSpaceOffset = &offset
				}
			}
			byteOffset += displayWidth
		}
		l.NonSpaceOffset = nonSpaceOffset
		l.Cells = cells
		l.AllSpace = allSpace
	})
}

type Cell struct {
	Rune         rune
	RuneLen      int // number of bytes in string
	RuneWidth    int // visual width
	DisplayWidth int // visual width with padding
	ByteOffset   int // byte offset
	RuneOffset   int // rune offset
}

var nextMomentID int64

func NewMomentFromFile(
	path string,
	scope Scope,
) (
	moment *Moment,
	linebreak Linebreak,
	err error,
) {
	defer he(&err)

	// read
	contentBytes, err := ioutil.ReadFile(path)
	ce(err, "read %s", path)

	scope.Sub(func() []byte {
		return contentBytes
	}).Call(NewMomentFromBytes, &moment, &linebreak)

	info, err := getFileInfo(path)
	ce(err)
	moment.FileInfo = info

	moment.Language = LanguageFromPath(path)

	return
}

func NewMomentFromBytes(
	bs []byte,
	scope Scope,
	config BufferConfig,
) (
	moment *Moment,
	linebreak Linebreak,
) {

	linebreak = "\n" // default

	content := string(bs)

	// split
	lineContents := splitLines(content)
	n := 0
	for i, lineContent := range lineContents {
		noCR := strings.TrimSuffix(lineContent, "\r")
		if len(noCR) != len(lineContent) {
			lineContents[i] = noCR
			n++
		}
	}
	if float64(n)/float64(len(lineContents)) > 0.4 {
		linebreak = "\r\n"
	}

	// lines
	var lines []*Line
	for _, content := range lineContents {
		line := &Line{
			content:  content,
			initOnce: new(sync.Once),
			config:   &config,
		}
		lines = append(lines, line)
	}

	moment = NewMoment(nil)
	moment.lines = lines
	moment.subContentHashStates = make([]*[]byte, len(lines))
	moment.subContentHashes = make([]*HashSum, len(lines))

	return
}

func NewMomentsFromPath(
	path string,
	scope Scope,
) (
	moments []*Moment,
	linebreaks []Linebreak,
	paths []string,
	err error,
) {

	stat, err := os.Stat(path)
	if err != nil {
		return
	}

	if stat.IsDir() {
		var f *os.File
		f, err = os.Open(path)
		if err != nil {
			return
		}
		defer f.Close()
		for {
			infos, err := f.Readdir(256)
			for _, info := range infos {
				if info.IsDir() {
					continue
				}
				name := info.Name()
				p := filepath.Join(path, name)
				var moment *Moment
				var linebreak Linebreak
				scope.Sub(func() string {
					return p
				}).Call(NewMomentFromFile, &moment, &linebreak, &err)
				if err != nil {
					continue
				}
				moments = append(moments, moment)
				linebreaks = append(linebreaks, linebreak)
				paths = append(paths, p)
			}
			if err == nil {
				break
			}
		}

	} else {
		var moment *Moment
		var linebreak Linebreak
		scope.Sub(func() string {
			return path
		}).Call(NewMomentFromFile, &moment, &linebreak, &err)
		if err != nil {
			return
		}
		moments = append(moments, moment)
		linebreaks = append(linebreaks, linebreak)
		paths = append(paths, path)
	}

	return
}

func splitLines(s string) (ret []string) {
	if s == "" {
		ret = append(ret, "")
		return
	}
	for len(s) > 0 {
		i := strings.Index(s, "\n")
		if i == -1 {
			ret = append(ret, s)
			return
		}
		ret = append(ret, s[:i+1])
		s = s[i+1:]
	}
	return
}
