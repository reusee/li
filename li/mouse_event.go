package li

import (
	"github.com/gdamore/tcell"
)

type MouseEvent = *tcell.EventMouse

type MouseConfig struct {
	ScrollLines int
}

func (_ Provide) MouseConfig(
	getConfig GetConfig,
) MouseConfig {
	var config struct {
		Mouse MouseConfig
	}
	ce(getConfig(&config))
	ret := config.Mouse
	if ret.ScrollLines == 0 {
		ret.ScrollLines = 3
	}
	return ret
}

type HandleMouseEvent func(
	ev MouseEvent,
)

func (_ Provide) HandleMouseEvent(
	mouseConfig MouseConfig,
	moveCursor MoveCursor,
) HandleMouseEvent {

	return func(
		ev MouseEvent,
	) {
		mask := ev.Buttons()
		if mask&tcell.WheelDown > 0 {
			// scroll down
			moveCursor(Move{RelLine: mouseConfig.ScrollLines})
		} else if mask&tcell.WheelUp > 0 {
			// scroll up
			moveCursor(Move{RelLine: -mouseConfig.ScrollLines})
		}
	}

}
