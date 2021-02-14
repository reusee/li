package li

//TODO

type KeyHandler interface {
	IsKeyHandler()
}

type ExpectingKey struct {
	Key  rune
	Cont KeyHandler
}

func (_ ExpectingKey) IsKeyHandler() {}

type PredictKey struct {
	Func func(rune) bool
	Cont KeyHandler
}

func (_ PredictKey) IsKeyHandler() {}

type ExecuteFunc struct {
	Func func()
	Cont KeyHandler
}

func (_ ExecuteFunc) IsKeyHandler() {}

type ExecuteCommand struct {
	Name string
	Spec CommandSpec
	Cont KeyHandler
}

func (_ ExecuteCommand) IsKeyHandler() {}

type KeyHandlerHint struct {
	Hint []string
	Cont KeyHandler
}

func (_ KeyHandlerHint) IsKeyHandler() {}
