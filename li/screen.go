package li

import (
	"github.com/gdamore/tcell"
)

type Screen interface {
	tcell.Screen
	SetCursorShape(CursorShape)
}

type TcellScreen struct {
	tcell.Screen
}

func (t TcellScreen) SetCursorShape(shape CursorShape) {
	switch shape {
	case CursorBeam:
		pt("\033[6 q")
	case CursorBlock:
		pt("\033[2 q")
	}
}

func (_ Provide) Screen(
	on On,
) Screen { // NOCOVER, testing codes uses tcell.SimulationScreen

	screen, err := tcell.NewScreen()
	ce(err)
	on(EvExit, func() {
		screen.Fini()
	})
	ce(screen.Init()) // NOCOVER
	screen.EnableMouse()

	return TcellScreen{
		Screen: screen,
	}
}

func (_ Provide) ScreenSize(
	screen Screen,
) (
	width Width,
	height Height,
) {

	w, h := screen.Size()
	width = Width(w)
	height = Height(h)

	return
}
