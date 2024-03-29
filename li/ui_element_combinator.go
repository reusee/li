package li

import "fmt"

// ElementFrom

type _ElementFrom struct {
	UIDesc
}

var _ Element = _ElementFrom{}

func (e _ElementFrom) RenderFunc() any {
	return func(
		scope Scope,
		setContent SetContent,
	) {

		var children []Element

		e.IterSpecs(scope, func(v any) {
			switch v := v.(type) {

			case Element:
				if v != nil {
					children = append(children, v)
				}

			case []Element:
				for _, elem := range v {
					if elem != nil {
						children = append(children, elem)
					}
				}

			default:
				panic(we(fmt.Errorf("unknown spec %#v\n", v)))
			}
		})

		renderAll(scope, children...)

	}
}

func ElementFrom(specs ...any) _ElementFrom {
	return _ElementFrom{
		UIDesc: NewUIDesc(specs),
	}
}

// ElementFunc

type _ElementFunc struct {
	fn       any
	provides []any
}

var _ Element = ElementFunc(nil)

func (f _ElementFunc) RenderFunc() any {
	return func(
		scope Scope,
	) {
		scope.Fork(f.provides...).Call(f.fn)
	}
}

func ElementFunc(fn any, provides ...any) _ElementFunc {
	return _ElementFunc{
		fn:       fn,
		provides: provides,
	}
}

// ElementWith

type _ElementWith struct {
	elem     Element
	provides []any
}

var _ Element = _ElementWith{}

func (s _ElementWith) RenderFunc() any {
	return func(
		scope Scope,
	) {
		renderAll(scope.Fork(s.provides...), s.elem)
	}
}

func ElementWith(elem Element, provides ...any) _ElementWith {
	return _ElementWith{
		elem:     elem,
		provides: provides,
	}
}
