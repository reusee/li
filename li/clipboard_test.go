package li

import (
	"fmt"
	"testing"
)

func TestClipString(t *testing.T) {
	withHelloEditor(t, func(
		moment *Moment,
	) {

		clip := Clip{
			Moment: moment,
			Range: Range{
				Begin: Position{0, 0},
				End:   Position{9999, 9999},
			},
		}
		str := clip.String()
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
		str = clip.String()
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
		str = clip.String()
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
		str = clip.String()
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
		str = clip.String()
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
		str = clip.String()
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
		newClip NewClipFromSelection,
		moveCursor MoveCursor,
		toggle ToggleSelection,
	) {
		toggle()
		newClip()
		var clip Clip
		linkedOne(buffer, &clip)
		eq(t,
			clip.String(), "H",
		)

		moveCursor(Move{RelLine: 1})
		newClip()
		linkedOne(buffer, &clip)
		eq(t,
			clip.String(), "Hello, world!\n你",
		)

		// no selection
		toggle()
		newClip()
		linkedOne(buffer, &clip)
		eq(t,
			clip.String(), "Hello, world!\n你",
		)
	})
}

func TestInsertLastClip(t *testing.T) {
	withHelloEditor(t, func(
		cur CurrentView,
		insert InsertLastClip,
		newClip NewClipFromSelection,
		toggle ToggleSelection,
	) {
		view := cur()

		insert()
		eq(t,
			view.GetMoment().GetLine(0).content, "Hello, world!\n",
		)

		toggle()
		newClip()
		insert()
		eq(t,
			view.GetMoment().GetLine(0).content, "HHello, world!\n",
		)

	})
}
