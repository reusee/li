package li

import (
	"github.com/reusee/dscope"
)

type Provide struct{}

type Derive func(inits ...any)

type OnInit struct{}

func (_ Provide) OnInit(
	InitCompletion,
	InitCompletionWords,
	InitContextStatus,
	InitCursorStatus,
	InitFormatterGo,
	InitKeyEventHooks,
	InitKeyMacroStatus,
	InitLSP,
	InitLayoutStatus,
	InitLineHints,
	InitModeStatus,
	InitShowViewListFlag,
	InitCurrentViewFilePathStatus,
	InitViewEvents,
) (_ OnInit) {
	return
}

func NewGlobal(fns ...any) Scope {
	inits := dscope.Methods(Provide{})
	inits = append(inits, fns...)
	scope := dscope.New(inits...)

	dscope.Get[OnInit](scope)

	return scope
}
