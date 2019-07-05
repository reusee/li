package li

import (
	"math/rand"

	"github.com/gdamore/tcell"
)

type Stainer interface {
	Line() any
}

type RandomStainer struct{}

func (_ RandomStainer) Line() any {
	return func(
		line *Line,
	) []*Color {
		var ret []*Color
		for i := 0; i < len(line.Runes); i++ {
			color := tcell.NewRGBColor(
				int32(rand.Intn(256)),
				int32(rand.Intn(256)),
				int32(rand.Intn(256)),
			)
			ret = append(ret, &color)
		}
		return ret
	}
}
