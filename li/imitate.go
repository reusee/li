package li

type (
	LogImitation   func(any) interface{}
	GetImitation   func() any
	NoLogImitation bool
)

func (_ Provide) Imitation() (
	log LogImitation,
	get GetImitation,
) {
	var imitation any
	log = func(v any) interface{} {
		imitation = v
		return v
	}
	get = func() any {
		return imitation
	}
	return
}

func (_ Command) Imitate() (spec CommandSpec) {
	spec.Func = func(
		get GetImitation,
		scope Scope,
	) NoLogImitation {
		fn := get()
		if fn == nil {
			return true
		}
		scope.Call(fn)
		return true
	}
	spec.Desc = "imitate last action"
	return
}
