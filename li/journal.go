package li

import (
	"fmt"
	"strings"
)

type (
	AppendJournal func(
		format string,
		args ...any,
	)
	JournalLines  func() []string
	JournalHeight func(...int) int
)

func (_ Provide) Journal(
	derive Derive,
) (
	appendJournal AppendJournal,
	get JournalLines,
	accessHeight JournalHeight,
) {

	var lines []string

	appendJournal = func(format string, args ...any) {
		str := fmt.Sprintf(format, args...)
		lines = append(lines, strings.Split(str, "\n")...)
	}

	get = func() []string {
		return lines
	}

	height := int(1)
	accessHeight = func(updates ...int) int {
		if len(updates) > 0 {
			for _, update := range updates {
				height = update
			}
			derive(
				func() JournalHeight {
					return accessHeight
				},
			)
		}
		return height
	}

	return
}

func JournalUI(
	getLines JournalLines,
	getHeight JournalHeight,
) (
	ret Element,
) {

	lines := getLines()
	height := int(getHeight())
	if len(lines) > height {
		lines = lines[len(lines)-height:]
	}
	ret = Rect(
		Text(lines),
	)

	return
}
