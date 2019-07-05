package li

import "testing"

func TestNoImitate(t *testing.T) {
	withEditor(func(scope Scope) {
		scope.Sub(func() Func {
			return NamedCommands["Imitate"].Func
		}).Call(ExecuteCommandFunc)
	})
}
