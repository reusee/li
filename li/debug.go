package li

type DebugConfig struct {
	Verbose bool
}

func (_ Provide) DebugConfig(
	get GetConfig,
) (
	debugConfig DebugConfig,
) {

	var config struct {
		Debug DebugConfig
	}
	ce(get(&config))

	debugConfig = config.Debug

	return
}

func (_ Command) Foo() (spec CommandSpec) {
	spec.Desc = "foo"
	spec.Func = func(scope Scope) {
	}
	return
}
