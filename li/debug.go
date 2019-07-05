package li

func (_ Command) Foo() (spec CommandSpec) {
	spec.Desc = "foo"
	spec.Func = func(scope Scope) {
	}
	return
}
