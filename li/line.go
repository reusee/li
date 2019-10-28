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
		trigger(scope.Sub(
			&l,
		), EvLineInitialized)
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

var emptyLineHeightsInfo = make(map[int]int)

type (
	LineHint struct {
		Line  int
		Hints []string
	}
	AddLineHint func(LineHint)
)

type evCollectLineHint struct{}

var EvCollectLineHint = new(evCollectLineHint)

func CalculateLineHeights(
	moment *Moment,
	lineRange [2]int,
	scope Scope,
	trigger Trigger,
) (
	info map[int]int,
) {

	// set info[line] to line height other than 1
	// assuming line height is 1 if key not set

	var hints []LineHint
	add := AddLineHint(func(hint LineHint) {
		hints = append(hints, hint)
	})
	trigger(
		scope.Sub(&moment, &lineRange, &add),
		EvCollectLineHint,
	)
	if len(hints) > 0 {
		info = make(map[int]int)
	}
	for _, hint := range hints {
		info[hint.Line] += len(hint.Hints)
	}

	if info == nil {
		info = emptyLineHeightsInfo
	}
	return
}

func CalculateSumLineHeight(
	scope Scope,
	moment *Moment,
	lineRange [2]int,
) int {
	var info map[int]int
	scope.Sub(&moment, &lineRange).Call(CalculateLineHeights)
	if info == nil {
		return int(lineRange[1] - lineRange[0])
	}
	sum := 0
	for i := lineRange[0]; i < lineRange[1]; i++ {
		h, ok := info[i]
		if !ok {
			sum += 1
		} else {
			sum += h
		}
	}
	return sum
}
