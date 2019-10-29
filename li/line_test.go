package li

import "testing"

func TestLineHeight(t *testing.T) {
	withHelloEditor(t, func(
		moment *Moment,
		on On,
		scope Scope,
		ctrl func(string),
	) {
		on(EvCollectLineHints, func(
			add AddLineHint,
		) {
			add(moment, 0, []string{"42"})
		})
		ctrl("loop")

		var info map[int]int
		scope.Sub(
			&moment, &[2]int{0, 1},
		).Call(CalculateLineHeights, &info)
		eq(t,
			info[0], 2,
		)

		var height int
		scope.Sub(
			&moment, &[2]int{0, 1},
		).Call(CalculateSumLineHeight, &height)
		eq(t,
			height, 2,
		)

	})
}
