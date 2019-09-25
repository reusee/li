package li

type Selections []Range

func ToggleSelection(
	cur CurrentView,
	scope Scope,
) {
	view := cur()
	if view == nil {
		return
	}
	if view.SelectionAnchor != nil {
		view.SelectionAnchor = nil
		return
	}
	position := view.cursorPosition(scope)
	view.SelectionAnchor = &position
}

func (_ Command) ToggleSelection() (spec CommandSpec) {
	spec.Desc = "toggle selection"
	spec.Func = ToggleSelection
	return
}

func (v *View) selectedRange(scope Scope) *Range {
	if v.SelectionAnchor == nil {
		return nil
	}
	anchor := *v.SelectionAnchor
	cursor := v.cursorPosition(scope)
	if cursor.Before(anchor) {
		return &Range{
			Begin: cursor,
			End:   anchor,
		}
	}
	moment := v.GetMoment()
	line := moment.GetLine(scope, cursor.Line)
	var end Position
	if cursor.Cell == len(line.Cells)-1 {
		// at line end
		if cursor.Line < moment.NumLines()-1 {
			// next line
			end = Position{
				Line: cursor.Line + 1,
				Cell: 0,
			}
		} else {
			end = cursor
		}
	} else {
		cell := line.Cells[cursor.Cell+1]
		end = Position{
			Line: cursor.Line,
			Cell: cell.RuneOffset,
		}
	}
	return &Range{
		Begin: anchor,
		End:   end,
	}
}
