package li

import "unicode"

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

func runeCategory(r rune) int {
	if unicode.IsDigit(r) || unicode.IsLetter(r) {
		return 0
	} else if unicode.IsSpace(r) {
		return 1
	}
	return 99
}

func PosWordEnd(
	cur CurrentView,
) (
	pos Position,
) {
	v := cur()
	pos = v.cursorPosition()
	line := v.GetMoment().GetLine(v.CursorLine)
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

func PosWordBegin() {
	//TODO
}
