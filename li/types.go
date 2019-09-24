package li

type (
	Func       any
	ID         int64
	AfterFunc  any
	LineNumber int
	Abort      bool
)

type (
	Width  int
	Height int
)

type Position struct {
	Line int
	Cell int
}

func (p Position) Before(p2 Position) bool {
	if p.Line != p2.Line {
		return p.Line < p2.Line
	}
	return p.Cell < p2.Cell
}

type Range struct {
	Begin Position
	End   Position
}

func (r Range) Contains(p Position) bool {
	if p.Line < r.Begin.Line {
		return false
	}
	if p.Line == r.Begin.Line && p.Cell < r.Begin.Cell {
		return false
	}
	if p.Line == r.End.Line && p.Cell >= r.End.Cell {
		return false
	}
	if p.Line > r.End.Line {
		return false
	}
	return true
}
