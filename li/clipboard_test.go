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
