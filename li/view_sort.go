package li

import "reflect"

type ViewSortKey struct{}

type ViewSortFunc = func(*View) int

var NamedViewSortKeys = func() map[string]ViewSortFunc {
	m := make(map[string]ViewSortFunc)
	v := reflect.ValueOf(new(ViewSortKey))
	t := v.Type()
	for i := 0; i < v.NumMethod(); i++ {
		name := t.Method(i).Name
		fn := v.Method(i).Interface().(ViewSortFunc)
		m[name] = fn
	}
	return m
}()

func (_ ViewSortKey) ID(view *View) int {
	return int(view.ID)
}
