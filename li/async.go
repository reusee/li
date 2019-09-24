package li

type (
	RunInMainLoop func(fn any)
	RunWhenIdle   func(fn any)
)
