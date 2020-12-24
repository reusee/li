package li

func (_ Command) InsertNewline() (spec CommandSpec) {
	spec.Desc = "insert newline at cursor"
	spec.Func = func(
		scope Scope,
		cur CurrentView,
	) {
		view := cur()
		if view == nil {
			return
		}
		indent := getAdjacentIndent(view, view.CursorLine, view.CursorLine+1)
		fn := PositionFunc(PosCursor)
		str := "\n" + indent
		scope.Sub(
			&fn, &str,
		).Call(InsertAtPositionFunc)
	}
	return
}

func (_ Command) InsertTab() (spec CommandSpec) {
	spec.Desc = "insert tab at cursor"
	spec.Func = func(
		scope Scope,
	) {
		fn := PositionFunc(PosCursor)
		str := "\t"
		scope.Sub(
			&fn, &str,
		).Call(InsertAtPositionFunc)
	}
	return
}

func (_ Command) Append() (spec CommandSpec) {
	spec.Desc = "start append at current cursor"
	spec.Func = func(scope Scope, move MoveCursor) {
		move(Move{RelRune: 1})
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
		lineEnd LineEnd,
	) {
		view := cur()
		if view == nil {
			return
		}
		indent := getAdjacentIndent(view, view.CursorLine, view.CursorLine+1)
		fn := PositionFunc(PosLineEnd)
		str := "\n" + indent
		scope.Sub(
			&fn, &str, &view,
		).Call(InsertAtPositionFunc)
		lineEnd()
		scope.Call(EnableEditMode)
	}
	return
}

func (_ Command) EditNewLineAbove() (spec CommandSpec) {
	spec.Desc = "insert new line above the current line and enable edit mode"
	spec.Func = func(
		scope Scope,
		cur CurrentView,
		moveCursor MoveCursor,
		lineEnd LineEnd,
	) {
		view := cur()
		if view == nil {
			return
		}
		indent := getAdjacentIndent(view, view.CursorLine-1, view.CursorLine)
		fn := PositionFunc(PosLineBegin)
		str := indent + "\n"
		scope.Sub(
			&fn, &str, &view,
		).Call(InsertAtPositionFunc)
		moveCursor(Move{RelLine: -1})
		lineEnd()
		scope.Call(EnableEditMode)
	}
	return
}

func getIndent(view *View, lineNum int) string {
	line := view.GetMoment().GetLine(lineNum)
	if line == nil {
		return ""
	}
	if line.NonSpaceDisplayOffset == nil {
		return ""
	}
	var runes []rune
	for _, cell := range line.Cells {
		if cell.DisplayOffset >= *line.NonSpaceDisplayOffset {
			break
		}
		runes = append(runes, cell.Rune)
	}
	return string(runes)
}

func getAdjacentIndent(view *View, upwardLine int, downwardLine int) string {
	upwardIndent := 0
	var upwardRunes []rune
	for {
		line := view.GetMoment().GetLine(upwardLine)
		if line == nil {
			break
		}
		if line.NonSpaceDisplayOffset == nil {
			upwardLine--
			continue
		}
		if *line.NonSpaceDisplayOffset > upwardIndent {
			upwardIndent = *line.NonSpaceDisplayOffset
			for _, cell := range line.Cells {
				if cell.DisplayOffset >= upwardIndent {
					break
				}
				upwardRunes = append(upwardRunes, cell.Rune)
			}
		}
		break
	}

	downwardIndent := 0
	var downwardRunes []rune
	for {
		line := view.GetMoment().GetLine(downwardLine)
		if line == nil {
			break
		}
		if line.NonSpaceDisplayOffset == nil {
			downwardLine++
			continue
		}
		if *line.NonSpaceDisplayOffset > downwardIndent {
			downwardIndent = *line.NonSpaceDisplayOffset
			for _, cell := range line.Cells {
				if cell.DisplayOffset >= downwardIndent {
					break
				}
				downwardRunes = append(downwardRunes, cell.Rune)
			}
		}
		break
	}

	if upwardIndent > downwardIndent {
		return string(upwardRunes)
	}
	return string(downwardRunes)
}

func (_ Command) ChangeToWordEnd() (spec CommandSpec) {
	spec.Desc = "change text from current cursor position to end of word"
	spec.Func = ChangeToWordEnd
	return
}

func (_ Command) DeleteLine() (spec CommandSpec) {
	spec.Desc = "delete current line"
	spec.Func = DeleteLine
	return
}

func (_ Command) AppendAtLineEnd() (spec CommandSpec) {
	spec.Desc = "append at line end"
	spec.Func = func(scope Scope, lineEnd LineEnd) {
		lineEnd()
		scope.Call(EnableEditMode)
	}
	return
}

func (_ Command) ChangeLine() (spec CommandSpec) {
	spec.Desc = "change current line"
	spec.Func = func(
		scope Scope,
		v CurrentView,
	) {
		view := v()
		if view == nil {
			return
		}
		indent := getIndent(view, view.CursorLine)
		var begin, end Position
		scope.Call(PosLineBegin, &begin)
		scope.Call(PosLineEnd, &end)
		str := ""
		scope.Sub(
			&Range{begin, end}, &str,
		).Call(ReplaceWithinRange)
		fn := PositionFunc(PosCursor)
		scope.Sub(
			&fn, &indent,
		).Call(InsertAtPositionFunc)
		scope.Call(EnableEditMode)
	}
	return
}
