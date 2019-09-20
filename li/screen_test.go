package li

import (
	"testing"

	"github.com/gdamore/tcell"
)

func TestScreenResize(t *testing.T) {
	withHelloEditor(t, func(
		emitEvent EmitEvent,
		view *View,
		call func(string),
		getConfig GetConfig,
	) {
		var config struct {
			UI struct {
				JournalHeight int
			}
		}
		ce(getConfig(&config))

		emitEvent(tcell.NewEventResize(5, 3))
		eq(t,
			view.Box.Width(), 5,
			view.Box.Height(), 3-config.UI.JournalHeight,
		)

	})
}
