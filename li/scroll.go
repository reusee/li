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

type ScrollToCursor func()

func (_ Provide) ScrollToCursor(
	cur CurrentView,
	scope Scope,
	config ScrollConfig,
) ScrollToCursor {

	return func() {
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

	var calLineHeights CalculateLineHeights
	scope.Assign(&calLineHeights)
	lineHeights := calLineHeights(moment, [2]int{
		line - v.Box.Height() - 1, line,
	})

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

type ScrollEnd func()

func (_ Provide) ScrollEnd(
	cur CurrentView,
	scrollToCursor ScrollToCursor,
	moveCursor MoveCursor,
) ScrollEnd {
	return func() {
		view := cur()
		if view == nil {
			return
		}
		line := view.GetMoment().NumLines() - 1
		moveCursor(Move{AbsLine: &line})
		scrollToCursor()
	}
}

type ScrollHome func()

func (_ Provide) ScrollHome(
	scrollToCursor ScrollToCursor,
	moveCursor MoveCursor,
) ScrollHome {
	return func() {
		zero := 0
		moveCursor(Move{AbsLine: &zero, AbsCol: &zero})
		scrollToCursor()
	}
}

func (_ Command) ScrollHome() (spec CommandSpec) {
	spec.Func = func(home ScrollHome) {
		home()
	}
	return
}

type ScrollAbsOrEnd func()

func (_ Provide) ScrollAbsOrEnd(
	withN WithContextNumber,
	moveCursor MoveCursor,
	end ScrollEnd,
) ScrollAbsOrEnd {
	return func() {
		withN(func(n int) {
			if n > 0 {
				n--
				moveCursor(Move{AbsLine: &n})
			} else {
				end()
			}
		})
	}
}

type ScrollAbsOrHome func()

func (_ Provide) ScrollAbsOrHome(
	withN WithContextNumber,
	moveCursor MoveCursor,
	home ScrollHome,
) ScrollAbsOrHome {
	return func() {
		withN(func(n int) {
			if n > 0 {
				n--
				moveCursor(Move{AbsLine: &n})
			} else {
				home()
			}
		})
	}
}

func (_ Command) ScrollAbsOrEnd() (spec CommandSpec) {
	spec.Desc = "scroll to specified line or the end"
	spec.Func = func(end ScrollAbsOrEnd) {
		end()
	}
	return
}

func (_ Command) ScrollAbsOrHome() (spec CommandSpec) {
	spec.Desc = "scroll to specified line or the beginnig"
	spec.Func = func(home ScrollAbsOrHome) {
		home()
	}
	return
}

type ScrollCursorToUpper func()

func (_ Provide) ScrollCursorToUpper(
	cur CurrentView,
	config ScrollConfig,
	scope Scope,
) ScrollCursorToUpper {
	return func() {
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
}

func (_ Command) ScrollCursorToUpper() (spec CommandSpec) {
	spec.Desc = "scroll to make cursor position at upper half of current view"
	spec.Func = func(scroll ScrollCursorToUpper) {
		scroll()
	}
	return
}

type ScrollCursorToMiddle func()

func (_ Provide) ScrollCursorToMiddle(
	cur CurrentView,
	config ScrollConfig,
	scope Scope,
) ScrollCursorToMiddle {
	return func() {
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
}

func (_ Command) ScrollCursorToMiddle() (spec CommandSpec) {
	spec.Desc = "scroll to make cursor position at the middle of current view"
	spec.Func = func(scroll ScrollCursorToMiddle) {
		scroll()
	}
	return
}

type ScrollCursorToLower func()

func (_ Provide) ScrollCursorToLower(
	cur CurrentView,
	config ScrollConfig,
	scope Scope,
) ScrollCursorToLower {
	return func() {
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
}

func (_ Command) ScrollCursorToLower() (spec CommandSpec) {
	spec.Desc = "scroll to make cursor position at lower half of current view"
	spec.Func = func(scroll ScrollCursorToLower) {
		scroll()
	}
	return
}
