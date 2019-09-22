package li

import "sync"

type (
	On      func(ev any, fn any)
	OnNext  func(ev any, fn any)
	Trigger func(scope Scope, ev any)
)

func (_ Provide) Hook() (
	on On,
	onNext OnNext,
	trigger Trigger,
) {

	type callback struct {
		fn      any
		oneshot bool
	}
	callbacks := make(map[any][]callback)
	var l sync.Mutex

	on = func(ev any, fn any) {
		l.Lock()
		defer l.Unlock()
		callbacks[ev] = append(callbacks[ev], callback{
			fn: fn,
		})
	}

	onNext = func(ev any, fn any) {
		l.Lock()
		defer l.Unlock()
		callbacks[ev] = append(callbacks[ev], callback{
			fn:      fn,
			oneshot: true,
		})
	}

	trigger = func(scope Scope, ev any) {
		l.Lock()
		defer l.Unlock()
		i := 0
		cs := callbacks[ev]
		for i < len(cs) {
			callback := cs[i]
			scope.Call(callback.fn)
			if callback.oneshot {
				cs = append(cs[:i], cs[i+1:]...)
				continue
			}
			i++
		}
		callbacks[ev] = cs
	}

	return
}
