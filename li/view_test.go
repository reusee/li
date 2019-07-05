package li

import "testing"

func TestView(t *testing.T) {
	withHelloEditor(t, func(
		view *View,
	) {

		eq(t,
			view.ID > 0, true,
			view.Buffer != nil, true,
			view.Moment != nil, true,
			view.Box.Width(), 80,
			view.Box.Height(), 25,
			view.ViewportLine, 0,
			view.ViewportCol, 0,
			view.CursorLine, 0,
			view.CursorCol, 0,
		)

	})
}
