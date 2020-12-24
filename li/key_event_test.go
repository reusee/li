package li

import (
	"strings"
	"testing"
)

func TestStrokeSpecHints(t *testing.T) {
	withHelloEditor(t, func(
		scope Scope,
		emitRune EmitRune,
		ctrl func(string),
		moment *Moment,
		getScreenString GetScreenString,
		view *View,
		calLineHeights CalculateLineHeights,
	) {

		emitRune('f')
		var height int
		scope.Sub(
			&moment, &[2]int{0, 1},
		).Call(CalculateSumLineHeight, &height)
		eq(t,
			height, 2,
		)
		info := calLineHeights(moment, [2]int{0, 1})
		eq(t,
			info[0], 2,
		)
		lines := getScreenString(view.ContentBox)
		eq(t,
			strings.HasPrefix(lines[1], "press any key"), true,
		)

		emitRune('e')
		scope.Sub(
			&moment, &[2]int{0, 1},
		).Call(CalculateSumLineHeight, &height)
		eq(t,
			height, 1,
		)
		info = calLineHeights(moment, [2]int{0, 1})
		eq(t,
			info[0], 1,
		)

		emitRune('j') // next line
		emitRune('f')
		lines = getScreenString(view.ContentBox)
		eq(t,
			strings.HasPrefix(lines[2], "press any key"), true,
		)
		emitRune('f')

		emitRune('d') // dd
		lines = getScreenString(view.ContentBox)
		eq(t,
			strings.HasPrefix(lines[2], "press Rune[d] to delete cu"), true,
		)
		emitRune('a')

		emitRune('d') // dd
		lines = getScreenString(view.ContentBox)
		eq(t,
			strings.HasPrefix(lines[2], "press Rune[d] to delete cu"), true,
		)
		emitRune('a')

		emitRune('z') // zt zb zz
		lines = getScreenString(view.ContentBox)
		eq(t,
			strings.HasPrefix(lines[2], "press Rune[b] to scroll to"), true,
			strings.HasPrefix(lines[3], "press Rune[t] to scroll to"), true,
			strings.HasPrefix(lines[4], "press Rune[z] to scroll to"), true,
		)
		emitRune('a')

	})
}
