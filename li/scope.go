package li

import "github.com/reusee/dscope"

//TODO delete Global

type Def struct{}

func NewScope() Scope {
	return dscope.New(
		dscope.Methods(Def{})...,
	)
}

type Update func(decls ...any) Scope

func (_ Def) ToBeImplement() (
	_ Update,
	_ *Scope,
) {
	panic("these should be implemented")
}
