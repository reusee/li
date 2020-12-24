package li

import (
	"math/rand"

	"github.com/gdamore/tcell"
)

type Stainer interface {
	Line() dyn
}

type NoopStainer struct{}

var _ Stainer = NoopStainer{}

func (_ NoopStainer) Line() dyn {
	return func(
		moment *Moment,
		line *Line,
		lineNumber LineNumber,
	) (ret []*StyleFunc) {
		ret = make([]*StyleFunc, len(line.Runes()))
		return
	}
}

type RandomStainer struct{}

var _ Stainer = RandomStainer{}

func (_ RandomStainer) Line() dyn {
	return func(
		line *Line,
	) []*StyleFunc {
		var ret []*StyleFunc
		runes := line.Runes()
		for i := 0; i < len(runes); i++ {
			fn := SetFG(tcell.NewRGBColor(
				int32(rand.Intn(256)),
				int32(rand.Intn(256)),
				int32(rand.Intn(256)),
			))
			ret = append(ret, &fn)
		}
		return ret
	}
}
