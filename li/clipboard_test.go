package li

import (
	"fmt"
	"testing"
)

func TestClipString(t *testing.T) {
	withHelloEditor(t, func(
		moment *Moment,
		scope Scope,
	) {

		clip := Clip{
			Moment: moment,
			Range: Range{
				Begin: Position{0, 0},
				End:   Position{9999, 9999},
			},
		}
		str := clip.String(scope)
		if str != moment.GetContent() {
			t.Fatal()
		}

		clip = Clip{
			Moment: moment,
			Range: Range{
				Begin: Position{0, 0},
				End:   Position{0, 1},
			},
		}
		str = clip.String(scope)
		if str != "H" {
			t.Fatal()
		}

		clip = Clip{
			Moment: moment,
			Range: Range{
				Begin: Position{0, 0},
				End:   Position{1, 1},
			},
		}
		str = clip.String(scope)
		if str != "Hello, world!\n你" {
			fmt.Printf("%q\n", str)
			t.Fatal()
		}

		clip = Clip{
			Moment: moment,
			Range: Range{
				Begin: Position{1, 1},
				End:   Position{1, 1},
			},
		}
		str = clip.String(scope)
		if str != "" {
			fmt.Printf("%q\n", str)
			t.Fatal()
		}

		clip = Clip{
			Moment: moment,
			Range: Range{
				Begin: Position{999, 999},
				End:   Position{999, 999},
			},
		}
		str = clip.String(scope)
		if str != "" {
			fmt.Printf("%q\n", str)
			t.Fatal()
		}

		clip = Clip{
			Moment: moment,
			Range: Range{
				Begin: Position{0, 2},
				End:   Position{0, 0},
			},
		}
		str = clip.String(scope)
		if str != "He" {
			fmt.Printf("%q\n", str)
			t.Fatal()
		}

	})
}

func TestClipFromSelection(t *testing.T) {
	withHelloEditor(t, func(
		scope Scope,
		buffer *Buffer,
		linkedOne LinkedOne,
	) {
		scope.Call(ToggleSelection)
		scope.Call(NewClipFromSelection)
		var clip Clip
		linkedOne(buffer, &clip)
		eq(t,
			clip.String(scope), "H",
		)

		scope.Sub(&Move{RelLine: 1}).Call(MoveCursor)
		scope.Call(NewClipFromSelection)
		linkedOne(buffer, &clip)
		eq(t,
			clip.String(scope), "Hello, world!\n你",
		)

		// no selection
		scope.Call(ToggleSelection)
		scope.Call(NewClipFromSelection)
		linkedOne(buffer, &clip)
		eq(t,
			clip.String(scope), "Hello, world!\n你",
		)
	})
}

func TestInsertLastClip(t *testing.T) {
	withHelloEditor(t, func(
		scope Scope,
		cur CurrentView,
	) {
		view := cur()

		scope.Call(InsertLastClip)
		eq(t,
			view.GetMoment().GetLine(scope, 0).content, "Hello, world!\n",
		)

		scope.Call(ToggleSelection)
		scope.Call(NewClipFromSelection)
		scope.Call(InsertLastClip)
		eq(t,
			view.GetMoment().GetLine(scope, 0).content, "HHello, world!\n",
		)

	})
}