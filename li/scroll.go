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

	col := view.CursorCol
	line := view.CursorLine
	viewportCol := view.ViewportCol
	viewportLine := view.ViewportLine

	// move viewport
	if col < viewportCol {
		viewportCol = viewportCol - (viewportCol - col)
	} else if col >= viewportCol+view.Box.Width() {
		viewportCol -= viewportCol + view.Box.Width() - col - 1
	}
	var paddingTop, paddingBottom int
	if view.Box.Height() > config.PaddingTop+config.PaddingBottom {
		paddingTop = config.PaddingTop
		paddingBottom = config.PaddingBottom
	}
	if line < viewportLine+paddingTop {
		viewportLine = viewportLine - (viewportLine - line) - paddingTop
		if viewportLine < 0 {
			viewportLine = 0
		}
	} else if line >= viewportLine+view.Box.Height()-paddingBottom {
		viewportLine -= viewportLine + view.Box.Height() - line - 1 - paddingBottom
	}

	if view.ViewportLine == viewportLine && view.ViewportCol == viewportCol {
		return
	}

	view.ViewportLine = viewportLine
	view.ViewportCol = viewportCol

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
	useN UseN,
	scope Scope,
) {
	n := useN()
	if n > 0 {
		n--
		scope.Sub(&Move{AbsLine: &n}).Call(MoveCursor)
	} else {
		scope.Call(ScrollEnd)
	}
}

func ScrollAbsOrHome(
	useN UseN,
	scope Scope,
) {
	n := useN()
	if n > 0 {
		n--
		scope.Sub(&Move{AbsLine: &n}).Call(MoveCursor)
	} else {
		scope.Call(ScrollHome)
	}
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
) {
	view := cur()
	if view == nil {
		return
	}
	viewportLine := view.CursorLine - config.PaddingTop
	if viewportLine < 0 {
		viewportLine = 0
	}
	view.ViewportLine = viewportLine
}

func (_ Command) ScrollCursorToUpper() (spec CommandSpec) {
	spec.Desc = "scroll to make cursor position at upper half of current view"
	spec.Func = ScrollCursorToUpper
	return
}

func ScrollCursorToMiddle(
	cur CurrentView,
	config ScrollConfig,
) {
	view := cur()
	if view == nil {
		return
	}
	viewportLine := view.CursorLine - view.Box.Height()/2
	if viewportLine < 0 {
		viewportLine = 0
	}
	view.ViewportLine = viewportLine
}

func (_ Command) ScrollCursorToMiddle() (spec CommandSpec) {
	spec.Desc = "scroll to make cursor position at the middle of current view"
	spec.Func = ScrollCursorToMiddle
	return
}

func ScrollCursorToLower(
	cur CurrentView,
	config ScrollConfig,
) {
	view := cur()
	if view == nil {
		return
	}
	viewportLine := view.CursorLine - (view.Box.Height() - config.PaddingBottom) + 1
	if viewportLine < 0 {
		viewportLine = 0
	}
	view.ViewportLine = viewportLine
}

func (_ Command) ScrollCursorToLower() (spec CommandSpec) {
	spec.Desc = "scroll to make cursor position at lower half of current view"
	spec.Func = ScrollCursorToLower
	return
}
