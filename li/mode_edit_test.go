package li

import (
	"testing"

	"github.com/gdamore/tcell"
)

func TestEditModeRedoAfterSwitching(t *testing.T) {
	withEditorBytes(t, []byte("foo"), func(
		scope Scope,
		view *View,
	) {
		var config EditModeConfig
		config.DisableSequence = "jj"
		scope = scope.Sub(func() EditModeConfig {
			return config
		})
		scope.Call(EnableEditMode)

		var getModes CurrentModes
		scope.Assign(&getModes)
		modes := getModes()
		_, ok := modes[0].(*EditMode)
		if !ok {
			t.Fatal()
		}

		scope.Sub(func() KeyEvent {
			return tcell.NewEventKey(tcell.KeyRune, 'j', 0)
		}).Call(HandleKeyEvent)
		scope.Sub(func() KeyEvent {
			return tcell.NewEventKey(tcell.KeyRune, 'j', 0)
		}).Call(HandleKeyEvent)

		modes = getModes()
		_, ok = modes[0].(*EditMode)
		if ok {
			t.Fatal()
		}

		scope.Call(RedoLatest)
		eq(t,
			// no extra 'j'
			view.Moment.GetLine(0).content, "foo",
		)
	})
}