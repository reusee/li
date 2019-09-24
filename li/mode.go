package li

import (
	"reflect"
	"strings"
)

type Mode any

type (
	CurrentModes func(args ...[]Mode) []Mode
)

func (_ Provide) ModesAccessor(
	derive Derive,
) (
	fn CurrentModes,
) {

	modes := []Mode{
		// default modes
		new(ReadMode),
		new(ContextMode),
		new(SystemMode),
	}

	fn = func(args ...[]Mode) []Mode {
		if len(args) > 0 {
			for _, arg := range args {
				modes = arg
			}
			// must be derive to trigger dependencies recalculate
			derive(
				func() CurrentModes {
					return fn
				},
			)
		}
		return modes
	}

	return
}

func (_ Provide) ModeStatus(
	on On,
) Init2 {

	on(EvRenderStatus, func(
		getModes CurrentModes,
		add AddStatusLine,
		styles []Style,
	) {
		modes := getModes()
		add("")
		add("modes", Bold(true), AlignRight, Padding(0, 2, 0, 0))
		for _, mode := range modes {
			name := reflect.TypeOf(mode).Elem().Name()
			name = strings.TrimSuffix(name, "Mode")
			s := styles[0]
			if name == "Edit" {
				s = styles[1]
			}
			add(s, name, AlignRight, Padding(0, 2, 0, 0))
		}
	})

	return nil
}
