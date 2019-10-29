package li

import "testing"

func TestStrokeSpecHints(t *testing.T) {
	withHelloEditor(t, func(
		scope Scope,
		emitRune EmitRune,
		ctrl func(string),
		moment *Moment,
	) {
		emitRune('f')
		var height int
		scope.Sub(
			&moment, &[2]int{0, 1},
		).Call(CalculateSumLineHeight, &height)
		eq(t,
			height, 2,
		)
		emitRune('e')
		scope.Sub(
			&moment, &[2]int{0, 1},
		).Call(CalculateSumLineHeight, &height)
		eq(t,
			height, 1,
		)
	})
}
