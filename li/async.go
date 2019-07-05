package li

type (
	RunInMainLoop func(fn any)
	RunWhenIdle   func(fn any)
)

func (_ Provide) RunInMainLoop() RunInMainLoop {
	panic("you need to implement this")
}

func (_ Provide) RunWhenIdle() RunWhenIdle {
	panic("you need to implement this")
}
