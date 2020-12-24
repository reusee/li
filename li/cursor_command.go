package li

func (_ Command) MoveLeft() (spec CommandSpec) {
	spec.Func = func(move MoveCursor) {
		move(Move{RelRune: -1})
	}
	return
}

func (_ Command) MoveDown() (spec CommandSpec) {
	spec.Func = func(move MoveCursor) {
		move(Move{RelLine: 1})
	}
	return
}

func (_ Command) MoveUp() (spec CommandSpec) {
	spec.Func = func(move MoveCursor) {
		move(Move{RelLine: -1})
	}
	return
}

func (_ Command) MoveRight() (spec CommandSpec) {
	spec.Func = func(move MoveCursor) {
		move(Move{RelRune: 1})
	}
	return
}

func (_ Command) PageDown() (spec CommandSpec) {
	spec.Func = func(scope Scope) {
		scope.Call(PageDown)
	}
	return
}

func (_ Command) PageUp() (spec CommandSpec) {
	spec.Func = func(scope Scope) {
		scope.Call(PageUp)
	}
	return
}

func (_ Command) NextEmptyLine() (spec CommandSpec) {
	spec.Func = func(scope Scope) {
		scope.Call(NextEmptyLine)
	}
	return
}

func (_ Command) PrevEmptyLine() (spec CommandSpec) {
	spec.Func = func(scope Scope) {
		scope.Call(PrevEmptyLine)
	}
	return
}

func (_ Command) LineBegin() (spec CommandSpec) {
	spec.Func = func(scope Scope) {
		scope.Call(LineBegin)
	}
	return
}

func (_ Command) LineEnd() (spec CommandSpec) {
	spec.Func = func(scope Scope) {
		scope.Call(LineEnd)
	}
	return
}

func (_ Command) NextRune() (spec CommandSpec) {
	spec.Func = NextRune
	spec.Desc = "focus next specified rune in the same line"
	return
}

func (_ Command) PrevRune() (spec CommandSpec) {
	spec.Func = PrevRune
	spec.Desc = "focus previous specified rune in the same line"
	return
}

func (_ Command) NextLineWithRune() (spec CommandSpec) {
	spec.Desc = "jump to next line with specified rune"
	spec.Func = NextLineWithRune
	return
}

func (_ Command) PrevLineWithRune() (spec CommandSpec) {
	spec.Desc = "jump to previous line with specified rune"
	spec.Func = PrevLineWithRune
	return
}

func (_ Command) PrevDedentLine() (spec CommandSpec) {
	spec.Desc = "jump to previous dedent line"
	spec.Func = PrevDedentLine
	return
}

func (_ Command) NextDedentLine() (spec CommandSpec) {
	spec.Desc = "jump to next dedent line"
	spec.Func = NextDedentLine
	return
}
