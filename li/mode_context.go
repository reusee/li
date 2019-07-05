package li

import "fmt"

type ContextMode struct{}

var _ KeyStrokeHandler = new(ContextMode)

type NoResetN bool

type (
	SetN func(n int)
	GetN func() int
	UseN func() int
)

func (_ Provide) NumberContext() (
	set SetN,
	get GetN,
	use UseN,
) {

	var n int
	set = func(i int) {
		n = i
	}
	get = func() int {
		return n
	}
	use = func() (ret int) {
		ret = n
		n = 0
		return
	}

	return
}

func makeNumHandler(i int) StrokeSpec {
	return StrokeSpec{
		Sequence: []string{fmt.Sprintf("Rune[%d]", i)},
		Func: func(
			getN GetN,
			setN SetN,
			scope Scope,
		) NoResetN {
			n := getN()
			if n == 0 {
				setN(i)
			} else {
				setN(n*10 + i)
			}
			return true
		},
	}
}

func (_ ContextMode) StrokeSpecs() any {
	return func() []StrokeSpec {
		return []StrokeSpec{
			// n
			makeNumHandler(0),
			makeNumHandler(1),
			makeNumHandler(2),
			makeNumHandler(3),
			makeNumHandler(4),
			makeNumHandler(5),
			makeNumHandler(6),
			makeNumHandler(7),
			makeNumHandler(8),
			makeNumHandler(9),
		}
	}
}
