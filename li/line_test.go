package li

import "testing"

func TestLineHeight(t *testing.T) {
	withHelloEditor(t, func(
		moment *Moment,
		on On,
		calLineHeights CalculateLineHeights,
		ctrl func(string),
		calLineHeight CalculateSumLineHeight,
	) {
		on(EvCollectLineHints, func(
			add AddLineHint,
		) {
			add(moment, 0, []string{"42"})
		})
		ctrl("loop")

		info := calLineHeights(moment, [2]int{0, 1})
		eq(t,
			info[0], 2,
		)

		height := calLineHeight(moment, [2]int{0, 1})
		eq(t,
			height, 2,
		)

	})
}
