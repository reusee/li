package li

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type (
	AppendJournal func(
		format string,
		args ...any,
	)
	JournalLines         func() []string
	JournalHeight        func(...int) int
	InitialJournalHeight int
)

func (_ Provide) Journal(
	derive Derive,
	uiConfig UIConfig,
	cont ContinueMainLoop,
) (
	appendJournal AppendJournal,
	get JournalLines,
	accessHeight JournalHeight,
	initialjournalheight InitialJournalHeight,
) {

	var lines []string
	var l sync.RWMutex

	appendJournal = func(format string, args ...any) {
		l.Lock()
		defer l.Unlock()
		if len(lines) > 2000 {
			lines = append([]string{}, lines[len(lines)-1000:]...)
		}
		str := fmt.Sprintf(format, args...)
		split := strings.Split(str, "\n")
		t := time.Now().Format("15:04:05.000000 ")
		for _, line := range split {
			lines = append(lines, t+line)
		}
		cont()
	}

	get = func() []string {
		l.RLock()
		defer l.RUnlock()
		return lines
	}

	initHeight := uiConfig.JournalHeight
	if initHeight == 0 {
		initHeight = 1
	}
	initialjournalheight = InitialJournalHeight(initHeight)

	height := initHeight
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

	//TODO wrap long lines
	//TODO scrolling
	lines := getLines()
	height := int(getHeight())
	if len(lines) > height {
		lines = lines[len(lines)-height:]
	}
	ret = Rect(
		Text(lines),
		Fill(true),
	)

	return
}

func (_ Command) ToggleJournalHeight() (spec CommandSpec) {
	spec.Desc = "toggle journal height"
	spec.Func = func(
		initHeight InitialJournalHeight,
		access JournalHeight,
		screenHeight Height,
	) {
		height := int(access())
		if height == int(initHeight) {
			height = int(screenHeight) - 10
		} else {
			height = int(initHeight)
		}
		access(height)
	}
	return
}
