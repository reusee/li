package li

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
