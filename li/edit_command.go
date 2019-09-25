package li

func (_ Command) InsertNewline() (spec CommandSpec) {
	spec.Desc = "insert newline at cursor"
	spec.Func = func(
		scope Scope,
		cur CurrentView,
	) {
		view := cur()
		indent := getAdjacentIndent(scope, view, view.CursorLine, view.CursorLine+1)
		scope.Sub(func() (PositionFunc, string) {
			return PosCursor, "\n" + indent
		}).Call(InsertAtPositionFunc)
	}
	return
}

func (_ Command) InsertTab() (spec CommandSpec) {
	spec.Desc = "insert tab at cursor"
	spec.Func = func(
		scope Scope,
	) {
		scope.Sub(func() (PositionFunc, string) {
			return PosCursor, "\t"
		}).Call(InsertAtPositionFunc)
	}
	return
}

func (_ Command) Append() (spec CommandSpec) {
	spec.Desc = "start append at current cursor"
	spec.Func = func(scope Scope) {
		scope.Sub(func() Move { return Move{RelRune: 1} }).Call(MoveCursor)
		scope.Call(EnableEditMode)
	}
	return
}

func (_ Command) DeletePrevRune() (spec CommandSpec) {
	spec.Desc = "delete previous rune at cursor"
	spec.Func = DeletePrevRune
	return
}

func (_ Command) DeleteRune() (spec CommandSpec) {
	spec.Desc = "delete one rune at cursor"
	spec.Func = DeleteRune
	return
}

func (_ Command) Delete() (spec CommandSpec) {
	spec.Desc = "delete selected or text object"
	spec.Func = Delete
	return
}

func (_ Command) Change() (spec CommandSpec) {
	spec.Desc = "change selected or text object"
	spec.Func = ChangeText
	return
}

func (_ Command) EditNewLineBelow() (spec CommandSpec) {
	spec.Desc = "insert new line below the current line and enable edit mode"
	spec.Func = func(
		scope Scope,
		cur CurrentView,
	) {
		view := cur()
		indent := getAdjacentIndent(scope, view, view.CursorLine, view.CursorLine+1)
		scope.Sub(func() (PositionFunc, string, *View) {
			return PosLineEnd, "\n" + indent, view
		}).Call(InsertAtPositionFunc)
		scope.Call(LineEnd)
		scope.Call(EnableEditMode)
	}
	return
}

func (_ Command) EditNewLineAbove() (spec CommandSpec) {
	spec.Desc = "insert new line above the current line and enable edit mode"
	spec.Func = func(
		scope Scope,
		cur CurrentView,
	) {
		view := cur()
		indent := getAdjacentIndent(scope, view, view.CursorLine-1, view.CursorLine)
		scope.Sub(func() (PositionFunc, string, *View) {
			return PosLineBegin, indent + "\n", view
		}).Call(InsertAtPositionFunc)
		scope.Sub(func() Move { return Move{RelLine: -1} }).Call(MoveCursor)
		scope.Call(LineEnd)
		scope.Call(EnableEditMode)
	}
	return
}

func getAdjacentIndent(scope Scope, view *View, l1 int, l2 int) string {
	indent := 0
	var runes []rune
	lineNum := l1
	for {
		line := view.GetMoment().GetLine(scope, lineNum)
		if line == nil {
			break
		}
		if line.NonSpaceDisplayOffset == nil {
			lineNum--
			continue
		}
		if *line.NonSpaceDisplayOffset > indent {
			indent = *line.NonSpaceDisplayOffset
			for _, cell := range line.Cells {
				if cell.DisplayOffset >= indent {
					break
				}
				runes = append(runes, cell.Rune)
			}
		}
		break
	}
	lineNum = l2
	for {
		line := view.GetMoment().GetLine(scope, lineNum)
		if line == nil {
			break
		}
		if line.NonSpaceDisplayOffset == nil {
			lineNum++
			continue
		}
		if *line.NonSpaceDisplayOffset > indent {
			indent = *line.NonSpaceDisplayOffset
			for _, cell := range line.Cells {
				if cell.DisplayOffset >= indent {
					break
				}
				runes = append(runes, cell.Rune)
			}
		}
		break
	}
	return string(runes)
}

func (_ Command) ChangeToWordEnd() (spec CommandSpec) {
	spec.Desc = "change text from current cursor position to end of word"
	spec.Func = ChangeToWordEnd
	return
}
