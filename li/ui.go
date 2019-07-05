package li

import (
	"reflect"
)

type Element interface {
	RenderFunc() any
}

type (
	BGColor   Color
	FGColor   Color
	Bold      bool
	Underline bool
	Fill      bool
	Point     [2]int
	Align     uint8
)

const (
	AlignLeft Align = iota
	AlignRight
	AlignCenter
)

// UIDesc

type UIDesc struct {
	InitFuncs []any
}

func NewUIDesc(specs []any) UIDesc {
	var initFuncs []any
	for _, spec := range specs {
		spec := spec
		if t := reflect.TypeOf(spec); t.Kind() == reflect.Func && t.Name() == "" {
			initFuncs = append(initFuncs, spec)
		} else {
			// wrap
			initFuncs = append(initFuncs, func() any {
				return spec
			})
		}
	}
	return UIDesc{
		InitFuncs: initFuncs,
	}
}

func (u UIDesc) IterSpecs(scope Scope, cb func(any)) {
	for _, fn := range u.InitFuncs {
		rets := scope.Call(fn)
		for _, ret := range rets {
			cb(ret.Interface())
		}
	}
}

// Margin

type _Margin []int

func Margin(spec ...int) _Margin {
	return _Margin(spec)
}

// Padding

type _Padding []int

func Padding(spec ...int) _Padding {
	return _Padding(spec)
}

// util functions

func renderAll(scope Scope, elements ...Element) {
	for len(elements) > 0 {
		var next []Element
		for _, elem := range elements {
			var e Element
			var es []Element
			scope.Call(elem.RenderFunc(), &e, &es)
			if e != nil {
				next = append(next, e)
			}
			for _, e := range es {
				if e != nil {
					next = append(next, e)
				}
			}
		}
		elements = next
	}
}
