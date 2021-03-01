package li

import (
	"fmt"
	"strconv"
)

type ContextMode struct {
	Number int
}

var _ KeyStrokeHandler = new(ContextMode)

func (c *ContextMode) StrokeSpecs() any {
	return func() []StrokeSpec {
		return []StrokeSpec{

			{
				Predict: func(ev KeyEvent) bool {
					r := ev.Rune()
					return r >= '0' && r <= '9'
				},
				Func: func(
					ev KeyEvent,
				) {
					n := int(ev.Rune() - '0')
					if n == 0 && c.Number > 0 {
						c.Number = c.Number * 10
					} else if n > 0 {
						c.Number = c.Number*10 + n
					}
				},
			},

			{
				Sequence: []string{"Esc"},
				Func: func() {
					c.Number = 0
				},
			},
		}
	}
}

func (_ Provide) ContextStatus(
	on On,
) OnStartup {
	return func() {

		on(EvCollectStatusSections, func(
			getModes CurrentModes,
			add AddStatusSection,
		) {
			for _, mode := range getModes() {
				m, ok := mode.(*ContextMode)
				if ok {
					add("context", [][]any{
						{"num: " + strconv.Itoa(m.Number), AlignRight, Padding(0, 2, 0, 0)},
					})
					break
				}
			}
		})

		on(EvCollectLineHints, func(
			getModes CurrentModes,
			curView CurrentView,
			add AddLineHint,
		) {
			view := curView()
			if view == nil {
				return
			}
			for _, mode := range getModes() {
				m, ok := mode.(*ContextMode)
				if ok && m.Number > 0 {
					add(view.GetMoment(), view.CursorLine, []string{
						fmt.Sprintf("context number: %d", m.Number),
					})
					break
				}
			}
		})

	}
}

type WithContextNumber func(fn func(int))

type SetContextNumber func(int)

func (_ Provide) ContextNumber(
	getModes CurrentModes,
) (
	with WithContextNumber,
	set SetContextNumber,
) {
	with = func(fn func(int)) {
		for _, mode := range getModes() {
			m, ok := mode.(*ContextMode)
			if ok {
				fn(m.Number)
				m.Number = 0
				break
			}
		}
	}
	set = func(i int) {
		for _, mode := range getModes() {
			m, ok := mode.(*ContextMode)
			if ok {
				m.Number = i
				break
			}
		}
	}
	return
}
