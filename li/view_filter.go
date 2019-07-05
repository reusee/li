package li

import "reflect"

type ViewFilter struct{}

var NamedViewFilters = func() map[string]any {
	m := make(map[string]any)
	v := reflect.ValueOf(new(ViewFilter))
	t := v.Type()
	for i := 0; i < v.NumMethod(); i++ {
		name := t.Method(i).Name
		fn := v.Method(i).Interface()
		m[name] = fn
	}
	return m
}()

func (_ ViewFilter) Any(view *View) bool {
	return true
}
