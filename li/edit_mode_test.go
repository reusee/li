package li

import (
	"testing"

	"github.com/gdamore/tcell"
)

func TestEditModeRedoAfterSwitching(t *testing.T) {
	withEditorBytes(t, []byte("foo"), func(
		scope Scope,
		enable EnableEditMode,
		view *View,
	) {
		var config EditModeConfig
		config.DisableSequence = "jj"
		scope = scope.Fork(func() EditModeConfig {
			return config
		})
		enable()

		var getModes CurrentModes
		var handleKey HandleKeyEvent
		scope.Assign(&getModes, &handleKey)

		modes := getModes()
		_, ok := modes[0].(*EditMode)
		if !ok {
			pt("%#v\n", modes)
			t.Fatalf("got %T", modes[0])
		}

		// trigger
		handleKey(tcell.NewEventKey(tcell.KeyRune, 'j', 0))
		handleKey(tcell.NewEventKey(tcell.KeyRune, 'j', 0))
		modes = getModes()
		_, ok = modes[0].(*EditMode)
		if ok {
			t.Fatal()
		}

		scope.Call(RedoLatest)
		eq(t,
			// no extra 'j'
			view.GetMoment().GetLine(0).content, "foo",
		)

		// not trigger
		enable()
		handleKey(tcell.NewEventKey(tcell.KeyRune, 'j', 0))
		handleKey(tcell.NewEventKey(tcell.KeyRune, 'k', 0))
		modes = getModes()
		_, ok = modes[0].(*EditMode)
		if !ok {
			t.Fatal()
		}

	})
}
