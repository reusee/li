package li

type (
	MomentLine struct {
		Moment *Moment
		Line   int
	}
	LineHints      map[MomentLine][]string
	AddLineHint    func(*Moment, int, []string)
	ClearLineHints func()
)

func (_ Provide) LineHints() (
	m LineHints,
	add AddLineHint,
	clear ClearLineHints,
) {
	m = make(map[MomentLine][]string)
	add = func(moment *Moment, line int, hints []string) {
		key := MomentLine{moment, line}
		m[key] = append(m[key], hints...)
	}
	clear = func() {
		for k := range m {
			delete(m, k)
		}
	}
	return
}

type evCollectLineHints struct{}

var EvCollectLineHints = new(evCollectLineHints)

func (_ Provide) CollectLineHints(
	on On,
) Init2 {
	on(EvLoopBegin, func(
		clear ClearLineHints,
		trigger Trigger,
		scope Scope,
	) {
		clear()
		trigger(scope, EvCollectLineHints)
	})
	return nil
}
