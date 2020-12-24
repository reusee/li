package li

func (_ Command) InsertNewline() (spec CommandSpec) {
	spec.Desc = "insert newline at cursor"
	spec.Func = func(
		insert InsertAtPositionFunc,
		cur CurrentView,
	) {
		view := cur()
		if view == nil {
			return
		}
		indent := getAdjacentIndent(view, view.CursorLine, view.CursorLine+1)
		fn := PositionFunc(PosCursor)
		str := "\n" + indent
		insert(str, fn)
	}
	return
}

func (_ Command) InsertTab() (spec CommandSpec) {
	spec.Desc = "insert tab at cursor"
	spec.Func = func(
		insert InsertAtPositionFunc,
	) {
		fn := PositionFunc(PosCursor)
		str := "\t"
		insert(str, fn)
	}
	return
}

func (_ Command) Append() (spec CommandSpec) {
	spec.Desc = "start append at current cursor"
	spec.Func = func(enable EnableEditMode, move MoveCursor) {
		move(Move{RelRune: 1})
		enable()
	}
	return
}

func (_ Command) DeletePrevRune() (spec CommandSpec) {
	spec.Desc = "delete previous rune at cursor"
	spec.Func = func(del DeletePrevRune) {
		del()
	}
	return
}

func (_ Command) DeleteRune() (spec CommandSpec) {
	spec.Desc = "delete one rune at cursor"
	spec.Func = func(del DeleteRune) {
		del()
	}
	return
}

func (_ Command) Delete() (spec CommandSpec) {
	spec.Desc = "delete selected or text object"
	spec.Func = func(del Delete) Abort {
		return del()
	}
	return
}

func (_ Command) Change() (spec CommandSpec) {
	spec.Desc = "change selected or text object"
	spec.Func = func(change ChangeText) Abort {
		return change()
	}
	return
}

func (_ Command) EditNewLineBelow() (spec CommandSpec) {
	spec.Desc = "insert new line below the current line and enable edit mode"
	spec.Func = func(
		cur CurrentView,
		lineEnd LineEnd,
		insert InsertAtPositionFunc,
		enable EnableEditMode,
	) {
		view := cur()
		if view == nil {
			return
		}
		indent := getAdjacentIndent(view, view.CursorLine, view.CursorLine+1)
		fn := PositionFunc(PosLineEnd)
		str := "\n" + indent
		insert(str, fn)
		lineEnd()
		enable()
	}
	return
}

func (_ Command) EditNewLineAbove() (spec CommandSpec) {
	spec.Desc = "insert new line above the current line and enable edit mode"
	spec.Func = func(
		enable EnableEditMode,
		cur CurrentView,
		moveCursor MoveCursor,
		lineEnd LineEnd,
		insert InsertAtPositionFunc,
	) {
		view := cur()
		if view == nil {
			return
		}
		indent := getAdjacentIndent(view, view.CursorLine-1, view.CursorLine)
		fn := PositionFunc(PosLineBegin)
		str := indent + "\n"
		insert(str, fn)
		moveCursor(Move{RelLine: -1})
		lineEnd()
		enable()
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
	spec.Func = func(change ChangeToWordEnd) {
		change()
	}
	return
}

func (_ Command) DeleteLine() (spec CommandSpec) {
	spec.Desc = "delete current line"
	spec.Func = func(del DeleteLine) {
		del()
	}
	return
}

func (_ Command) AppendAtLineEnd() (spec CommandSpec) {
	spec.Desc = "append at line end"
	spec.Func = func(enable EnableEditMode, lineEnd LineEnd) {
		lineEnd()
		enable()
	}
	return
}

func (_ Command) ChangeLine() (spec CommandSpec) {
	spec.Desc = "change current line"
	spec.Func = func(
		scope Scope,
		v CurrentView,
		insert InsertAtPositionFunc,
		enable EnableEditMode,
		replace ReplaceWithinRange,
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
		replace(Range{begin, end}, str)
		fn := PositionFunc(PosCursor)
		insert(indent, fn)
		enable()
	}
	return
}
