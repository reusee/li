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

func HandleMouseEvent(
	ev MouseEvent,
	mouseConfig MouseConfig,
	scope Scope,
) {

	mask := ev.Buttons()

	if mask&tcell.WheelDown > 0 {
		// scroll down
		scope.Sub(&Move{RelLine: mouseConfig.ScrollLines}).
			Call(MoveCursor)

	} else if mask&tcell.WheelUp > 0 {
		// scroll up
		scope.Sub(&Move{RelLine: -mouseConfig.ScrollLines}).
			Call(MoveCursor)

	}

}
