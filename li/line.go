package li

import (
	"sync"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

type Line struct {
	Cells                 []Cell
	Runes                 []rune
	DisplayWidth          int
	AllSpace              bool
	NonSpaceDisplayOffset *int

	content  string
	initOnce *sync.Once
	config   *BufferConfig
}

type Cell struct {
	Rune          rune
	Len           int // number of bytes in utf8 encoding
	Width         int // visual width without padding
	DisplayWidth  int // visual width with padding
	DisplayOffset int // visual column offset with padding in line
	RuneOffset    int // rune offset in line
	ByteOffset    int // utf8 byte offset in line
	UTF16Offset   int // byte offset in utf16 encoding in line
}

func (l *Line) init(scope Scope) {
	l.initOnce.Do(func() {
		var cells []Cell
		allSpace := true
		displayOffset := 0
		utf16ByteOffset := 0
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
			runeLen := utf8.RuneLen(r)
			cell := Cell{
				Rune:          r,
				Len:           runeLen,
				Width:         width,
				DisplayWidth:  displayWidth,
				DisplayOffset: displayOffset,
				RuneOffset:    i,
				ByteOffset:    byteOffset,
				UTF16Offset:   utf16ByteOffset,
			}
			cells = append(cells, cell)
			l.DisplayWidth += cell.DisplayWidth
			if !unicode.IsSpace(r) {
				allSpace = false
				if nonSpaceOffset == nil {
					offset := displayOffset
					nonSpaceOffset = &offset
				}
			}
			displayOffset += displayWidth
			utf16ByteOffset += len(utf16.Encode([]rune{r})) * 2
			byteOffset += runeLen
		}
		l.NonSpaceDisplayOffset = nonSpaceOffset
		l.Cells = cells
		l.AllSpace = allSpace

		var trigger Trigger
		scope.Assign(&trigger)
		trigger(scope.Sub(func() *Line {
			return l
		}), EvLineInitialized)
	})
}

type evLineInitialized struct{}

var EvLineInitialized = new(evLineInitialized)

type LineInitProcs chan []*Line

func (_ Provide) LineInitProcs(
	scope Scope,
) LineInitProcs {
	c := make(chan []*Line, 512)
	for i := 0; i < numCPU; i++ {
		go func() {
			for {
				lines := <-c
				for i := len(lines) - 1; i >= 0; i-- {
					lines[i].init(scope)
				}
			}
		}()
	}
	return c
}
