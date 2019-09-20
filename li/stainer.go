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
		for i := 0; i < len(line.Runes); i++ {
			ret = append(ret, nil)
		}
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
		for i := 0; i < len(line.Runes); i++ {
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
