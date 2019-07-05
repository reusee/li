package li

import (
	"github.com/gdamore/tcell"
)

type Screen = tcell.Screen

type (
	Width  int
	Height int
)

func (_ Provide) Screen(
	onExit OnExit,
) Screen { // NOCOVER, testing codes uses tcell.SimulationScreen

	screen, err := tcell.NewScreen()
	ce(err)
	onExit(func() { // NOCOVER
		screen.Fini()
	})
	ce(screen.Init()) // NOCOVER
	screen.EnableMouse()

	return screen
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
