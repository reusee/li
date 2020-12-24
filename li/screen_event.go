package li

import (
	"github.com/gdamore/tcell"
)

type ScreenEvent = tcell.Event

type (
	EmitRune  func(r rune)
	EmitKey   func(k tcell.Key)
	EmitEvent func(ScreenEvent)
)

func HandleScreenEvent(
	ev ScreenEvent,
	derive Derive,
	handleMouse HandleMouseEvent,
	handleKey HandleKeyEvent,
) {

	switch ev := ev.(type) {

	case *tcell.EventKey:
		handleKey(ev)

	case *tcell.EventMouse:
		handleMouse(ev)

	case *tcell.EventResize:
		width, height := ev.Size()
		derive(
			func() (Width, Height) {
				return Width(width), Height(height)
			},
		)

	}

}
