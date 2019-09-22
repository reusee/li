package li

import "sync"

type (
	Exit    func()
	SigExit chan struct{}

	evExit struct{}
)

var EvExit = new(evExit)

func (_ Provide) Exit(
	trigger Trigger,
	scope Scope,
) (
	exit Exit,
	sigExit SigExit,
) {

	var once sync.Once
	sigExit = make(chan struct{})
	exit = func() {
		once.Do(func() {
			trigger(scope, EvExit)
			close(sigExit)
		})
	}

	return
}
