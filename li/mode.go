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

	on(EvCollectStatusSections, func(
		getModes CurrentModes,
		add AddStatusSection,
		styles []Style,
	) {
		modes := getModes()
		var lines [][]any
		for _, mode := range modes {
			name := reflect.TypeOf(mode).Elem().Name()
			name = strings.TrimSuffix(name, "Mode")
			s := styles[0]
			if name == "Edit" {
				s = styles[1]
			}
			lines = append(lines, []any{
				s, name, AlignRight, Padding(0, 2, 0, 0),
			})
		}
		add("modes", lines)
	})

	return nil
}
