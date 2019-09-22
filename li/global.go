package li

import "reflect"

type Provide struct{}

type Derive func(inits ...interface{})

type Init func() any

func NewGlobal(fns ...any) Scope {
	provide := new(Provide)
	var inits []interface{}
	v := reflect.ValueOf(provide)
	for i := 0; i < v.NumMethod(); i++ {
		fn := v.Method(i).Interface()
		if initFunc, ok := fn.(func() Init); ok {
			fn = initFunc()
		}
		inits = append(inits, fn)
	}
	inits = append(inits, fns...)
	return NewScope(inits...)
}
