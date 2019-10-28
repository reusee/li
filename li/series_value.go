package li

import "reflect"

type SeriesValue struct {
	Type      dyn
	Access    dyn // func(...Type) Type
	OnLink    dyn // func(Type)
	OnChanged dyn // func(Type)
}

func (s SeriesValue) Provider() dyn {
	// as link anchor
	type Anchor struct{}
	var anchor Anchor
	// returns func(Scope) Access
	return reflect.MakeFunc(
		// func(Scope) Access
		reflect.FuncOf(
			[]reflect.Type{
				reflect.TypeOf((*Scope)(nil)).Elem(),
			},
			[]reflect.Type{
				reflect.TypeOf(s.Access),
			},
			false,
		),
		func(args []reflect.Value) []reflect.Value {
			scope := args[0].Interface().(Scope)
			var link Link
			var linkedOne LinkedOne
			scope.Assign(&link, &linkedOne)
			// returns Access
			return []reflect.Value{
				reflect.MakeFunc(
					reflect.TypeOf(s.Access),
					func(args []reflect.Value) []reflect.Value {
						for _, arg := range args {
							for i := 0; i < arg.Len(); i++ {
								value := arg.Index(i)
								link(anchor, value.Interface())
								if s.OnLink != nil {
									reflect.ValueOf(s.OnLink).Call(
										[]reflect.Value{value},
									)
								}
							}
						}
						retPtr := reflect.New(reflect.TypeOf(s.Type))
						linkedOne(anchor, retPtr.Interface())
						if len(args) > 0 && s.OnChanged != nil {
							reflect.ValueOf(s.OnChanged).Call(
								[]reflect.Value{retPtr.Elem()},
							)
						}
						return []reflect.Value{retPtr.Elem()}
					},
				),
			}
		},
	).Interface()
}
