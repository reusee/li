package li

import "testing"

func TestSelection(t *testing.T) {
	withHelloEditor(t, func(
		view *View,
		move MoveCursor,
		toggle ToggleSelection,
	) {

		toggle()
		move(Move{RelLine: 1})
		r := view.selectedRange()
		eq(t,
			r.Begin.Line, 0,
			r.Begin.Cell, 0,
			r.End.Line, 1,
			r.End.Cell, 1,
		)
		toggle()
		r = view.selectedRange()
		eq(t,
			r == nil, true,
		)

		toggle()
		move(Move{RelLine: -1})
		r = view.selectedRange()
		eq(t,
			r.Begin.Line, 0,
			r.Begin.Cell, 0,
			r.End.Line, 1,
			r.End.Cell, 0,
		)

	})
}
