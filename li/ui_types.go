package li

type SetContent func(x int, y int, mainc rune, combc []rune, style Style)

type Box struct {
	Top    int
	Left   int
	Bottom int
	Right  int
}

type ZBox struct {
	Box
	Z int
}

func (b Box) Width() int {
	w := max(b.Right-b.Left, 0)
	return w
}

func (b Box) Height() int {
	h := max(b.Bottom-b.Top, 0)
	return h
}

func (b Box) Intersect(b2 Box) bool {
	return intersect(b.Left, b.Right, b2.Left, b2.Right) &&
		intersect(b.Top, b.Bottom, b2.Top, b2.Bottom)
}

func (b Box) Contains(b2 Box) bool {
	return b2.Left >= b.Left &&
		b2.Right <= b.Right &&
		b2.Top >= b.Top &&
		b2.Bottom <= b.Bottom
}

func intersect(a1, a2, b1, b2 int) bool {
	return (a1 < b1 && a2 >= b2) ||
		(a1 >= b1 && a1 < b2) ||
		(b1 < a1 && b2 >= a2) ||
		(b1 >= a1 && b1 < a2)
}
