package li

import (
	"testing"
	"time"
)

func TestUndoRedo(t *testing.T) {
	withHelloEditor(t, func(
		view *View,
		scope Scope,
		delRune DeleteRune,
	) {
		delRune()
		eq(t,
			view.GetMoment().GetLine(0).content, "ello, world!\n",
		)
		scope.Call(Undo)
		eq(t,
			view.GetMoment().GetLine(0).content, "Hello, world!\n",
		)
		scope.Call(RedoLatest)
		eq(t,
			view.GetMoment().GetLine(0).content, "ello, world!\n",
		)
	})
}

func TestUndoRedo2(t *testing.T) {
	withEditor(func(
		scope Scope,
	) {
		scope.Call(Undo)
		scope.Call(RedoLatest)
		scope.Call(UndoDuration1)
	})
}

func TestUndoRedo3(t *testing.T) {
	withHelloEditor(t, func(
		view *View,
		scope Scope,
	) {
		scope.Call(Undo)
		eq(t,
			view.GetMoment().GetLine(0).content, "Hello, world!\n",
		)
		scope.Call(RedoLatest)
		eq(t,
			view.GetMoment().GetLine(0).content, "Hello, world!\n",
		)
	})
}

func TestUndoRedo4(t *testing.T) {
	withHelloEditor(t, func(
		view *View,
		scope Scope,
		insert InsertAtPositionFunc,
		delRune DeleteRune,
	) {
		delRune()
		eq(t,
			view.GetMoment().GetLine(0).content, "ello, world!\n",
		)
		scope.Call(Undo)
		eq(t,
			view.GetMoment().GetLine(0).content, "Hello, world!\n",
		)
		insert("foo", PosCursor)
		eq(t,
			view.GetMoment().GetLine(0).content, "fooHello, world!\n",
		)
		scope.Call(Undo)
		eq(t,
			view.GetMoment().GetLine(0).content, "Hello, world!\n",
		)
		scope.Call(RedoLatest)
		eq(t,
			view.GetMoment().GetLine(0).content, "fooHello, world!\n",
		)
	})
}

func TestTimedUndo(t *testing.T) {
	withHelloEditor(t, func(
		scope Scope,
		view *View,
		insert InsertAtPositionFunc,
	) {
		var config UndoConfig
		scope.Assign(&config)
		config.DurationMS1 = 10
		scope = scope.Sub(func() UndoConfig { return config })

		m := view.GetMoment()
		t0 := time.Now()
		for {
			if time.Since(t0) > time.Millisecond*50 {
				break
			}
			insert("foo", PosCursor)
		}

		scope.Call(UndoDuration1)
		eq(t,
			view.GetMoment().T0.Sub(m.T0) > time.Millisecond*10, true,
		)
	})
}

func TestTimedUndo2(t *testing.T) {
	withHelloEditor(t, func(
		scope Scope,
		view *View,
	) {
		scope.Call(UndoDuration1)
	})
}
