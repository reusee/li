package li

import "testing"

func TestHooks(t *testing.T) {
	withEditor(func(
		on On,
		onNext OnNext,
		trigger Trigger,
		scope Scope,
	) {

		trigger(scope, 42)

		n := 0
		on(42, func() {
			n++
		})
		trigger(scope, 42)
		if n != 1 {
			t.Fatal()
		}
		trigger(scope, 42)
		if n != 2 {
			t.Fatal()
		}

		onNext(42, func() {
			n++
		})
		trigger(scope, 42)
		if n != 4 {
			t.Fatal()
		}
		trigger(scope, 42)
		if n != 5 {
			t.Fatal()
		}

	})
}
