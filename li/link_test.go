package li

import "testing"

func TestLink(t *testing.T) {
	withEditor(func(
		scope Scope,
		link Link,
		linked LinkedOne,
		linkedAll LinkedAll,
		drop DropLinked,
	) {

		link(1, 2)
		var i int
		linked(1, &i)
		if i != 2 {
			t.Fatal()
		}

		link(2, 1)
		i = 0
		linked(1, &i)
		if i != 2 {
			t.Fatal()
		}

		var is []int
		linkedAll(1, &is)
		if len(is) != 1 {
			t.Fatal()
		}
		if is[0] != 2 {
			t.Fatal()
		}

		is = is[:0]
		linkedAll(2, &is)
		if len(is) != 1 {
			t.Fatal()
		}
		if is[0] != 1 {
			t.Fatal()
		}

		link(4, 5)
		link(5, 6)

		is = is[:0]
		linkedAll(2, &is)
		if len(is) != 1 {
			t.Fatal()
		}
		if is[0] != 1 {
			t.Fatal()
		}

		link(2, 1)
		i = 0
		linked(1, &i)
		if i != 2 {
			t.Fatal()
		}

		drop(1)

		is = is[:0]
		linkedAll(2, &is)
		if len(is) != 0 {
			t.Fatal()
		}

		i = 0
		linked(1, &i)
		if i != 0 {
			t.Fatal()
		}

	})
}

func TestLinkOrder(t *testing.T) {
	withEditor(func(
		link Link,
		linked LinkedOne,
		linkedAll LinkedAll,
		dropOne DropLinkedOne,
	) {

		link(1, 2)
		link(1, 3)
		link(1, 4)
		for i := 0; i < 100; i++ {
			var i int
			linked(1, &i)
			if i != 4 {
				t.Fatal()
			}
			var is []int
			linkedAll(1, &is)
			if len(is) != 3 {
				t.Fatal()
			}
			if is[0] != 4 || is[1] != 3 || is[2] != 2 {
				t.Fatal()
			}
		}

		dropOne(1)
		var is []int
		linkedAll(1, &is)
		if len(is) != 2 {
			t.Fatal()
		}
		if is[0] != 3 || is[1] != 2 {
			t.Fatal()
		}

	})
}

func TestDropLink(t *testing.T) {
	withEditor(func(
		link Link,
		dropLink DropLink,
		linkedAll LinkedAll,
	) {
		link(1, 2)
		link(1, 3)
		link(1, 4)

		var ints []int
		linkedAll(1, &ints)
		eq(t,
			len(ints), 3,
		)

		dropLink(1, 2)
		ints = ints[:0]
		linkedAll(1, &ints)
		eq(t,
			len(ints), 2,
		)
	})
}
