package li

type ScrollConfig struct {
	PaddingTop    int
	PaddingBottom int
}

func (_ Provide) ScrollConfig(
	getConfig GetConfig,
) ScrollConfig {
	var config struct {
		Scroll ScrollConfig
	}
	config.Scroll.PaddingTop = 5
	config.Scroll.PaddingBottom = 5
	ce(getConfig(&config))
	return config.Scroll
}

func ScrollToCursor(
	cur CurrentView,
	scope Scope,
	config ScrollConfig,
) {

	view := cur()
	if view == nil {
		return
	}

	// move viewport column
	col := view.CursorCol
	viewportCol := view.ViewportCol
	if col < viewportCol {
		viewportCol = viewportCol - (viewportCol - col)
	} else if col >= viewportCol+view.Box.Width() {
		viewportCol -= viewportCol + view.Box.Width() - col - 1
	}

	// move viewport line
	line := view.CursorLine
	viewportLine := view.ViewportLine
	min, max := view.calculateViewportLineRange(
		scope,
		view.GetMoment(), line,
		config.PaddingTop,
		config.PaddingBottom,
	)
	log("%d %d\n", min, max)
	if viewportLine < min {
		viewportLine = min
	} else if viewportLine > max {
		viewportLine = max
	}

	// no change
	if view.ViewportLine == viewportLine && view.ViewportCol == viewportCol {
		return
	}

	view.ViewportLine = viewportLine
	view.ViewportCol = viewportCol

}

func (v *View) calculateViewportLineRange(
	scope Scope,
	moment *Moment,
	line int,
	paddingTop int,
	paddingBottom int,
) (
	min int,
	max int,
) {

	if paddingTop+paddingBottom > v.Box.Height() {
		paddingTop = 0
		paddingBottom = 0
	}

	var lineHeights map[int]int
	scope.Sub(&moment, &[2]int{
		line - v.Box.Height() - 1, line,
	}).Call(CalculateLineHeights, &lineHeights)

	min = line
	height := v.Box.Height() - paddingBottom
	if h, ok := lineHeights[line]; ok {
		height -= h
	} else {
		height--
	}
	for {
		if min < 0 {
			min = 0
			break
		}
		l := min - 1
		if l < 0 {
			break
		}
		if h, ok := lineHeights[l]; ok {
			height -= h
		} else {
			height--
		}
		if height < 0 {
			break
		}
		min--
	}

	max = line
	height = paddingTop
	for {
		if max < 0 {
			max = 0
			break
		}
		l := max - 1
		if l < 0 {
			break
		}
		if h, ok := lineHeights[l]; ok {
			height -= h
		} else {
			height--
		}
		if height < 0 {
			break
		}
		max--
	}

	return
}

func ScrollEnd(
	cur CurrentView,
	scope Scope,
) {
	view := cur()
	if view == nil {
		return
	}
	line := view.GetMoment().NumLines() - 1
	scope.Sub(&Move{AbsLine: &line}).Call(MoveCursor)
	scope.Call(ScrollToCursor)
}

func ScrollHome(
	scope Scope,
) {
	zero := 0
	scope.Sub(&Move{AbsLine: &zero, AbsCol: &zero}).Call(MoveCursor)
	scope.Call(ScrollToCursor)
}

func (_ Command) ScrollHome() (spec CommandSpec) {
	spec.Func = func(scope Scope) {
		scope.Call(ScrollHome)
	}
	return
}

func ScrollAbsOrEnd(
	withN WithContextNumber,
	scope Scope,
) {
	withN(func(n int) {
		if n > 0 {
			n--
			scope.Sub(&Move{AbsLine: &n}).Call(MoveCursor)
		} else {
			scope.Call(ScrollEnd)
		}
	})
}

func ScrollAbsOrHome(
	withN WithContextNumber,
	scope Scope,
) {
	withN(func(n int) {
		if n > 0 {
			n--
			scope.Sub(&Move{AbsLine: &n}).Call(MoveCursor)
		} else {
			scope.Call(ScrollHome)
		}
	})
}

func (_ Command) ScrollAbsOrEnd() (spec CommandSpec) {
	spec.Func = func(scope Scope) {
		scope.Call(ScrollAbsOrEnd)
	}
	return
}

func (_ Command) ScrollAbsOrHome() (spec CommandSpec) {
	spec.Func = func(scope Scope) {
		scope.Call(ScrollAbsOrHome)
	}
	return
}

func ScrollCursorToUpper(
	cur CurrentView,
	config ScrollConfig,
	scope Scope,
) {
	view := cur()
	if view == nil {
		return
	}
	_, max := view.calculateViewportLineRange(
		scope,
		view.GetMoment(),
		view.CursorLine,
		config.PaddingTop,
		config.PaddingBottom,
	)
	view.ViewportLine = max
}

func (_ Command) ScrollCursorToUpper() (spec CommandSpec) {
	spec.Desc = "scroll to make cursor position at upper half of current view"
	spec.Func = ScrollCursorToUpper
	return
}

func ScrollCursorToMiddle(
	cur CurrentView,
	config ScrollConfig,
	scope Scope,
) {
	view := cur()
	if view == nil {
		return
	}
	min, max := view.calculateViewportLineRange(
		scope,
		view.GetMoment(),
		view.CursorLine,
		config.PaddingTop,
		config.PaddingBottom,
	)
	view.ViewportLine = (max + min) / 2
}

func (_ Command) ScrollCursorToMiddle() (spec CommandSpec) {
	spec.Desc = "scroll to make cursor position at the middle of current view"
	spec.Func = ScrollCursorToMiddle
	return
}

func ScrollCursorToLower(
	cur CurrentView,
	config ScrollConfig,
	scope Scope,
) {
	view := cur()
	if view == nil {
		return
	}
	min, _ := view.calculateViewportLineRange(
		scope,
		view.GetMoment(),
		view.CursorLine,
		config.PaddingTop,
		config.PaddingBottom,
	)
	view.ViewportLine = min
}

func (_ Command) ScrollCursorToLower() (spec CommandSpec) {
	spec.Desc = "scroll to make cursor position at lower half of current view"
	spec.Func = ScrollCursorToLower
	return
}
