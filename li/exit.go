package li

import "sync"

type (
	OnExit  func(cb func())
	Exit    func()
	SigExit chan struct{}
)

func (_ Provide) Exit() (
	onExit OnExit,
	exit Exit,
	sigExit SigExit,
) {

	var cbs []func()

	onExit = func(cb func()) {
		cbs = append(cbs, cb)
	}

	var once sync.Once
	sigExit = make(chan struct{})
	exit = func() {
		once.Do(func() {
			for _, cb := range cbs {
				cb()
			}
			close(sigExit)
		})
	}

	return
}
