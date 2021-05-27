package li

import "testing"

func TestHooks(t *testing.T) {
	withEditor(func(
		on On,
		onNext OnNext,
		trigger Trigger,
		scope Scope,
	) {

		trigger(42)

		n := 0
		on(func(
			ev int,
		) {
			n++
		})
		trigger(42)
		if n != 1 {
			t.Fatal()
		}
		trigger(42)
		if n != 2 {
			t.Fatal()
		}

		onNext(func(
			ev int,
		) {
			n++
		})
		trigger(42)
		if n != 4 {
			t.Fatal()
		}
		trigger(42)
		if n != 5 {
			t.Fatal()
		}

	})
}
