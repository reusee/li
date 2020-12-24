package li

import (
	"sort"
	"sync"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

type Line struct {
	Cells                 []Cell
	DisplayWidth          int
	AllSpace              bool
	NonSpaceDisplayOffset *int

	runes    []rune
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

func (l *Line) Runes() []rune {
	l.init()
	return l.runes
}

func (l *Line) init() {
	l.initOnce.Do(func() {
		var cells []Cell
		allSpace := true
		displayOffset := 0
		utf16ByteOffset := 0
		byteOffset := 0
		l.runes = []rune(l.content)
		var nonSpaceOffset *int
		for i, r := range l.runes {
			width := runeDisplayWidth(r)
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
	})
}

type LineInitProcs chan []*Line

func (_ Provide) LineInitProcs() LineInitProcs {
	c := make(chan []*Line, 512)
	for i := 0; i < numCPU; i++ {
		go func() {
			for {
				lines := <-c
				for i := len(lines) - 1; i >= 0; i-- {
					lines[i].init()
				}
			}
		}()
	}
	return c
}

type CalculateLineHeights func(
	moment *Moment,
	lineRange [2]int,
) (
	info map[int]int,
)

func (_ Provide) CalculateLineHeights(
	scope Scope,
	trigger Trigger,
	getHints GetLineHints,
) CalculateLineHeights {

	return func(
		moment *Moment,
		lineRange [2]int,
	) (
		info map[int]int,
	) {
		hints, _ := getHints()
		info = make(map[int]int)
		for line := lineRange[0]; line < lineRange[1]; line++ {
			info[line] = 1
			n := sort.Search(len(hints), func(i int) bool {
				return hints[i].Moment.ID >= moment.ID &&
					hints[i].Line >= line
			})
			for ; n < len(hints); n++ {
				hint := hints[n]
				if hint.Moment.ID == moment.ID && hint.Line == line {
					// found
					info[line] += len(hint.Hints)
				}
			}
		}

		return
	}

}

func CalculateSumLineHeight(
	moment *Moment,
	lineRange [2]int,

	scope Scope,
	calculate CalculateLineHeights,
) int {
	info := calculate(moment, lineRange)
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
