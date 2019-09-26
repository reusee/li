package li

type (
	RunInMainLoop func(fn any)
	RunWhenIdle   func(fn any)
)

type LoopStep func()

func (_ Provide) Loop(
	run RunInMainLoop,
) (
	step LoopStep,
) {

	step = func() {
		run(func() {})
	}

	return
}
