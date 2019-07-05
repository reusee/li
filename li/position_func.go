package li

// PositionFunc returns a Position value.
// Assuming current view is not null and current cursor is valid
type PositionFunc any

func PosCursor(
	cur CurrentView,
) Position {
	view := cur()
	return view.cursorPosition()
}

func PosLineBegin(
	cur CurrentView,
) Position {
	v := cur()
	return Position{
		Line: v.CursorLine,
		Rune: 0,
		Col:  0,
	}
}

func PosLineEnd(
	cur CurrentView,
) (pos Position) {
	v := cur()
	line := v.Moment.GetLine(v.CursorLine)
	pos.Line = v.CursorLine
	pos.Rune = len(line.Cells) - 1
	for i := 0; i < len(line.Cells)-1; i++ {
		pos.Col += line.Cells[i].DisplayWidth
	}
	return
}

func PosPrevRune(
	cur CurrentView,
) Position {
	v := cur()
	pos := v.cursorPosition()
	if pos.Col == 0 {
		// at line begin
		if pos.Line > 0 {
			// prev line
			line := v.Moment.GetLine(pos.Line - 1)
			cell := line.Cells[len(line.Cells)-1]
			return Position{
				Line: pos.Line - 1,
				Rune: cell.RuneOffset,
				Col:  cell.ColOffset,
			}
		} else {
			return pos
		}
	} else {
		line := v.Moment.GetLine(pos.Line)
		cell := line.Cells[pos.Rune-1]
		return Position{
			Line: pos.Line,
			Rune: cell.RuneOffset,
			Col:  cell.ColOffset,
		}
	}
}

func PosNextRune(
	cur CurrentView,
) Position {
	v := cur()
	pos := v.cursorPosition()
	line := v.Moment.GetLine(pos.Line)
	if pos.Rune == len(line.Cells)-1 {
		// at line end
		if pos.Line < v.Moment.NumLines()-1 {
			// next line
			return Position{
				Line: pos.Line + 1,
				Col:  0,
				Rune: 0,
			}
		} else {
			return pos
		}
	} else {
		cell := line.Cells[pos.Rune+1]
		return Position{
			Line: pos.Line,
			Col:  cell.ColOffset,
			Rune: cell.RuneOffset,
		}
	}
}
