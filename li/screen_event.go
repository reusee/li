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
		scope.Sub(func() KeyEvent { return ev }).Call(HandleKeyEvent)

	case *tcell.EventMouse:
		scope.Sub(func() MouseEvent { return ev }).Call(HandleMouseEvent)

	case *tcell.EventResize:
		width, height := ev.Size()
		derive(
			func() (Width, Height) {
				return Width(width), Height(height)
			},
		)

	}

}
