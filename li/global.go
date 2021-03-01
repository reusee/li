package li

import (
	"reflect"

	"github.com/reusee/dscope"
)

type Provide struct{}

type Derive func(inits ...any)

type Init2 any

var init2Type = reflect.TypeOf((*Init2)(nil)).Elem()

type OnStartup func()

var _ dscope.Reducer = OnStartup(nil)

func (_ OnStartup) Reduce(_ dscope.Scope, vs []reflect.Value) reflect.Value {
	return dscope.Reduce(vs)
}

func (_ Provide) OnStartup() OnStartup {
	return func() {}
}

func NewGlobal(fns ...any) Scope {
	var inits []any
	var protoInit2s []any

	processFunc := func(fnValue reflect.Value) {
		fn := fnValue.Interface()
		if t := fnValue.Type(); t.NumOut() == 1 && t.Out(0) == init2Type {
			protoInit2s = append(protoInit2s, fn)
			return
		}
		inits = append(inits, fn)
	}

	provide := new(Provide)
	v := reflect.ValueOf(provide)
	for i := 0; i < v.NumMethod(); i++ {
		processFunc(v.Method(i))
	}

	for _, fn := range fns {
		processFunc(reflect.ValueOf(fn))
	}

	scope := dscope.New(inits...)

	var init2s []any
	for _, proto := range protoInit2s {
		res := scope.Call(proto)
		for _, ret := range res.Values {
			fn := ret.Interface()
			if fn == nil {
				continue
			}
			init2s = append(init2s, fn)
		}
	}
	scope = scope.Sub(init2s...)

	scope.Call(func(
		onStartup OnStartup,
	) {
		onStartup()
	})

	return scope
}
