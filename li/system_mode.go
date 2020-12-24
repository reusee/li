package li

import "time"

type SystemMode struct{}

var _ KeyStrokeHandler = new(SystemMode)

func (_ SystemMode) StrokeSpecs() any {
	return func() []StrokeSpec {
		return []StrokeSpec{
			{Sequence: []string{"F11"}, CommandName: "About"},
			{Sequence: []string{"F12"}, CommandName: "Exit"},
		}
	}
}

func (_ Command) Exit() (spec CommandSpec) {
	spec.Func = func(exit Exit) {
		exit()
	}
	spec.Desc = "exit editor"
	return
}

func (_ Command) About() (spec CommandSpec) {
	spec.Func = func(show ShowMessage) {
		show(
			[]string{
				"li editor",
				time.Now().Format("2006-01-02 15:04:05"),
			},
		)
	}
	spec.Desc = "about editor"
	return
}
