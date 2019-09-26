package li

import "testing"

func TestSegmentsSub(t *testing.T) {
	segments := Segments{
		&Segment{
			lines: make([]*Line, 3),
		},
		&Segment{
			lines: make([]*Line, 4),
		},
		&Segment{
			lines: make([]*Line, 5),
		},
	}

	sub := segments.Sub(-1, 0)
	eq(t,
		len(sub), 0,
	)

	sub = segments.Sub(-1, 1)
	eq(t,
		len(sub), 1,
		len(sub[0].lines), 1,
		sub.Len(), 1,
	)

	sub = segments.Sub(-1, 2)
	eq(t,
		len(sub), 1,
		len(sub[0].lines), 2,
		sub.Len(), 2,
	)

	sub = segments.Sub(-1, 3)
	eq(t,
		len(sub), 1,
		len(sub[0].lines), 3,
		sub.Len(), 3,
	)

	sub = segments.Sub(-1, 4)
	eq(t,
		len(sub), 2,
		len(sub[0].lines), 3,
		len(sub[1].lines), 1,
		sub.Len(), 4,
	)

	sub = segments.Sub(-1, 7)
	eq(t,
		len(sub), 2,
		len(sub[0].lines), 3,
		len(sub[1].lines), 4,
		sub.Len(), 7,
	)

	sub = segments.Sub(-1, 8)
	eq(t,
		len(sub), 3,
		len(sub[0].lines), 3,
		len(sub[1].lines), 4,
		len(sub[2].lines), 1,
		sub.Len(), 8,
	)

	sub = segments.Sub(-1, 12)
	eq(t,
		len(sub), 3,
		len(sub[0].lines), 3,
		len(sub[1].lines), 4,
		len(sub[2].lines), 5,
		sub.Len(), 12,
	)

	sub = segments.Sub(-1, 13)
	eq(t,
		len(sub), 3,
		len(sub[0].lines), 3,
		len(sub[1].lines), 4,
		len(sub[2].lines), 5,
		sub.Len(), 12,
	)

	sub = segments.Sub(0, -1)
	eq(t,
		len(sub), 3,
		len(sub[0].lines), 3,
		len(sub[1].lines), 4,
		len(sub[2].lines), 5,
		sub.Len(), 12,
	)

	sub = segments.Sub(1, -1)
	eq(t,
		len(sub), 3,
		len(sub[0].lines), 2,
		len(sub[1].lines), 4,
		len(sub[2].lines), 5,
		sub.Len(), 11,
	)

	sub = segments.Sub(2, -1)
	eq(t,
		len(sub), 3,
		len(sub[0].lines), 1,
		len(sub[1].lines), 4,
		len(sub[2].lines), 5,
		sub.Len(), 10,
	)

	sub = segments.Sub(3, -1)
	eq(t,
		len(sub), 2,
		len(sub[0].lines), 4,
		len(sub[1].lines), 5,
		sub.Len(), 9,
	)

	sub = segments.Sub(4, -1)
	eq(t,
		len(sub), 2,
		len(sub[0].lines), 3,
		len(sub[1].lines), 5,
		sub.Len(), 8,
	)

	sub = segments.Sub(5, -1)
	eq(t,
		len(sub), 2,
		len(sub[0].lines), 2,
		len(sub[1].lines), 5,
		sub.Len(), 7,
	)

	sub = segments.Sub(6, -1)
	eq(t,
		len(sub), 2,
		len(sub[0].lines), 1,
		len(sub[1].lines), 5,
		sub.Len(), 6,
	)

	sub = segments.Sub(7, -1)
	eq(t,
		len(sub), 1,
		len(sub[0].lines), 5,
		sub.Len(), 5,
	)

	sub = segments.Sub(11, -1)
	eq(t,
		len(sub), 1,
		len(sub[0].lines), 1,
		sub.Len(), 1,
	)

	sub = segments.Sub(12, -1)
	eq(t,
		len(sub), 0,
	)

	sub = segments.Sub(13, -1)
	eq(t,
		len(sub), 0,
	)

}
