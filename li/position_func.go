package li

import "unicode"

// PositionFunc returns a Position value.
// Assuming current view is not null and current cursor is valid
type PositionFunc any

func PosCursor(
	cur CurrentView,
	scope Scope,
) Position {
	view := cur()
	return view.cursorPosition(scope)
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
	scope Scope,
) (pos Position) {
	v := cur()
	line := v.GetMoment().GetLine(scope, v.CursorLine)
	pos.Line = v.CursorLine
	pos.Cell = len(line.Cells) - 1
	return
}

func PosPrevRune(
	cur CurrentView,
	scope Scope,
) Position {
	v := cur()
	pos := v.cursorPosition(scope)
	moment := v.GetMoment()
	if pos.Cell == 0 {
		// at line begin
		if pos.Line > 0 {
			// prev line
			line := moment.GetLine(scope, pos.Line-1)
			cell := line.Cells[len(line.Cells)-1]
			return Position{
				Line: pos.Line - 1,
				Cell: cell.RuneOffset,
			}
		} else {
			return pos
		}
	} else {
		line := moment.GetLine(scope, pos.Line)
		cell := line.Cells[pos.Cell-1]
		return Position{
			Line: pos.Line,
			Cell: cell.RuneOffset,
		}
	}
}

func PosNextRune(
	cur CurrentView,
	scope Scope,
) Position {
	v := cur()
	pos := v.cursorPosition(scope)
	moment := v.GetMoment()
	line := moment.GetLine(scope, pos.Line)
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

func runeCategory(r rune) int {
	if unicode.IsDigit(r) || unicode.IsLetter(r) || r == '_' {
		return 0
	} else if unicode.IsSpace(r) {
		return 1
	}
	return 99
}

func PosWordEnd(
	cur CurrentView,
	scope Scope,
) (
	pos Position,
) {
	v := cur()
	pos = v.cursorPosition(scope)
	line := v.GetMoment().GetLine(scope, v.CursorLine)
	lastCategory := -1
	for i := pos.Cell; i < len(line.Cells); i++ {
		pos.Cell = i
		category := runeCategory(line.Cells[i].Rune)
		if lastCategory != -1 && lastCategory != category {
			break
		}
		lastCategory = category
	}
	return
}

func PosWordBegin(
	cur CurrentView,
	scope Scope,
) (
	pos Position,
) {
	//TODO not tested
	v := cur()
	pos = v.cursorPosition(scope)
	line := v.GetMoment().GetLine(scope, pos.Line)
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
	scope Scope,
) (
	pos Position,
) {
	v := cur()
	pos = v.cursorPosition(scope)
	if pos.Line == 0 {
		pos.Cell = 0
		return
	}
	pos.Line--
	pos.Cell = len(v.GetMoment().GetLine(scope, pos.Line).Cells) - 1
	return
}
