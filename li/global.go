package li

import (
	"github.com/reusee/dscope"
)

type Provide struct{}

type Derive func(inits ...any)

type OnStartup func()

var _ dscope.Reducer = OnStartup(nil)

func (_ OnStartup) IsReducer() {}

func (_ Provide) OnStartup() OnStartup {
	return func() {}
}

func NewGlobal(fns ...any) Scope {
	inits := dscope.Methods(Provide{})
	inits = append(inits, fns...)
	scope := dscope.New(inits...)

	scope.Call(func(
		onStartup OnStartup,
	) {
		onStartup()
	})

	return scope
}
