package li

import "sync"

type (
	Exit    func()
	SigExit chan struct{}

	EvExit struct{}
)

func (_ Provide) Exit(
	trigger Trigger,
) (
	exit Exit,
	sigExit SigExit,
) {

	var once sync.Once
	sigExit = make(chan struct{})
	exit = func() {
		once.Do(func() {
			trigger(EvExit{})
			close(sigExit)
		})
	}

	return
}
