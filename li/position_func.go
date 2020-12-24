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
		Cell: 0,
	}
}

func PosNextLineBegin(
	cur CurrentView,
) Position {
	v := cur()
	line := v.CursorLine
	if line < v.GetMoment().NumLines()-1 {
		line++
	}
	return Position{
		Line: line,
		Cell: 0,
	}
}

func PosLineEnd(
	cur CurrentView,
) (pos Position) {
	v := cur()
	line := v.GetMoment().GetLine(v.CursorLine)
	pos.Line = v.CursorLine
	pos.Cell = len(line.Cells) - 1
	return
}

func PosPrevRune(
	cur CurrentView,
) Position {
	v := cur()
	pos := v.cursorPosition()
	moment := v.GetMoment()
	if pos.Cell == 0 {
		// at line begin
		if pos.Line > 0 {
			// prev line
			line := moment.GetLine(pos.Line - 1)
			cell := line.Cells[len(line.Cells)-1]
			return Position{
				Line: pos.Line - 1,
				Cell: cell.RuneOffset,
			}
		} else {
			return pos
		}
	} else {
		line := moment.GetLine(pos.Line)
		cell := line.Cells[pos.Cell-1]
		return Position{
			Line: pos.Line,
			Cell: cell.RuneOffset,
		}
	}
}

func PosNextRune(
	cur CurrentView,
) Position {
	v := cur()
	pos := v.cursorPosition()
	moment := v.GetMoment()
	line := moment.GetLine(pos.Line)
	if pos.Cell == len(line.Cells)-1 {
		// at line end
		if pos.Line < moment.NumLines()-1 {
			// next line
			return Position{
				Line: pos.Line + 1,
				Cell: 0,
			}
		} else {
			return pos
		}
	} else {
		cell := line.Cells[pos.Cell+1]
		return Position{
			Line: pos.Line,
			Cell: cell.RuneOffset,
		}
	}
}

func PosWordEnd(
	cur CurrentView,
) (
	pos Position,
) {
	v := cur()
	pos = v.cursorPosition()
	line := v.GetMoment().GetLine(v.CursorLine)
	var lastCategory RuneCategory
	for i := pos.Cell; i < len(line.Cells); i++ {
		pos.Cell = i
		category := runeCategory(line.Cells[i].Rune)
		if lastCategory > 0 && lastCategory != category {
			break
		}
		lastCategory = category
	}
	return
}

func PosWordBegin(
	cur CurrentView,
) (
	pos Position,
) {
	//TODO not tested
	v := cur()
	pos = v.cursorPosition()
	line := v.GetMoment().GetLine(pos.Line)
	for pos.Cell > 0 {
		category := runeCategory(line.Cells[pos.Cell].Rune)
		idx := pos.Cell - 1
		if idx < 0 {
			break
		}
		prevCategory := runeCategory(line.Cells[idx].Rune)
		if category != prevCategory {
			break
		}
		pos.Cell--
	}
	return
}

func PosPrevLineEnd(
	cur CurrentView,
) (
	pos Position,
) {
	v := cur()
	pos = v.cursorPosition()
	if pos.Line == 0 {
		pos.Cell = 0
		return
	}
	pos.Line--
	pos.Cell = len(v.GetMoment().GetLine(pos.Line).Cells) - 1
	return
}
