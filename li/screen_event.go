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
	scope Scope,
	derive Derive,
) {

	switch ev := ev.(type) {

	case *tcell.EventKey:
		scope.Sub(&ev).Call(HandleKeyEvent)

	case *tcell.EventMouse:
		scope.Sub(&ev).Call(HandleMouseEvent)

	case *tcell.EventResize:
		width, height := ev.Size()
		derive(
			func() (Width, Height) {
				return Width(width), Height(height)
			},
		)

	}

}
