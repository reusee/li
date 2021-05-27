package li

import (
	"reflect"
	"strings"
	"sync"
)

type Mode any

type (
	CurrentModes func(args ...[]Mode) []Mode
)

type modesLock struct {
	*sync.RWMutex
}

func (_ Provide) ModesVars() (
	l modesLock,
	p *[]Mode,
) {
	l = modesLock{
		RWMutex: new(sync.RWMutex),
	}
	modes := []Mode{
		// default modes
		new(ReadMode),
		new(ContextMode),
		new(SystemMode),
	}
	p = &modes
	return
}

func (_ Provide) ModesAccessor(
	derive Derive,
	trigger Trigger,
	l modesLock,
	ptr *[]Mode,
) (
	fn CurrentModes,
) {

	fn = func(args ...[]Mode) []Mode {
		if len(args) > 0 {
			l.Lock()
			defer l.Unlock()
			for _, arg := range args {
				*ptr = arg
			}
			// must be derive to trigger dependencies recalculate
			derive(
				func() CurrentModes {
					return fn
				},
			)
			trigger(EvModesChanged{
				Modes: *ptr,
			})
		} else {
			l.RLock()
			defer l.RUnlock()
		}
		ms := make([]Mode, len(*ptr))
		copy(ms, *ptr)
		return ms
	}

	return
}

type EvModesChanged struct {
	Modes []Mode
}

func (_ Provide) ModeStatus(
	on On,
) OnStartup {
	return func() {

		on(func(
			ev EvCollectStatusSections,
			getModes CurrentModes,
		) {
			modes := getModes()
			var lines [][]any
			for _, mode := range modes {
				name := reflect.TypeOf(mode).Elem().Name()
				name = strings.TrimSuffix(name, "Mode")
				s := ev.Styles[0]
				if name == "Edit" {
					s = ev.Styles[1]
				}
				lines = append(lines, []any{
					s, name, AlignRight, Padding(0, 2, 0, 0),
				})
			}
			ev.Add("modes", lines)
		})

	}
}
