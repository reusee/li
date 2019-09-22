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
	on On,
) Screen { // NOCOVER, testing codes uses tcell.SimulationScreen

	screen, err := tcell.NewScreen()
	ce(err)
	on(EvExit, func() {
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
