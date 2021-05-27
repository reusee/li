package li

import (
	"reflect"
	"sync"
)

type (
	On      func(cb any)
	OnNext  func(cb any)
	Trigger func(ev any)
)

type HooksLock struct {
	*sync.Mutex
}

type HooksMap map[reflect.Type][]callback

type callback struct {
	fn      reflect.Value
	oneshot bool
}

func (_ Provide) HookVars() (
	l HooksLock,
	m HooksMap,
) {
	l = HooksLock{
		Mutex: new(sync.Mutex),
	}
	m = make(HooksMap)
	return
}

func (_ Provide) Hook(
	scope Scope,
	l HooksLock,
	m HooksMap,
) (
	on On,
	onNext OnNext,
	trigger Trigger,
) {

	on = func(cb any) {
		t := reflect.TypeOf(cb).In(0)
		l.Lock()
		defer l.Unlock()
		m[t] = append(m[t], callback{
			fn: reflect.ValueOf(cb),
		})
	}

	onNext = func(cb any) {
		t := reflect.TypeOf(cb).In(0)
		l.Lock()
		defer l.Unlock()
		m[t] = append(m[t], callback{
			fn:      reflect.ValueOf(cb),
			oneshot: true,
		})
	}

	trigger = func(ev any) {
		var fns []reflect.Value
		l.Lock()
		i := 0
		t := reflect.TypeOf(ev)
		cs := m[t]
		for i < len(cs) {
			callback := cs[i]
			fns = append(fns, callback.fn)
			if callback.oneshot {
				cs = append(cs[:i], cs[i+1:]...)
				continue
			}
			i++
		}
		m[t] = cs
		l.Unlock()
		evPtr := reflect.New(t)
		evPtr.Elem().Set(reflect.ValueOf(ev))
		subScope := scope.Sub(evPtr.Interface())
		for _, fn := range fns {
			subScope.CallValue(fn)
		}
	}

	return
}
