package li

import (
	"math"

	"github.com/gdamore/tcell"
)

type CursorShape uint8

const (
	CursorBlock CursorShape = iota
	CursorBeam
)

type Move struct {
	RelLine int
	RelRune int
	AbsLine *int
	AbsCol  *int
}

func MoveCursor(
	move Move,
	cur CurrentView,
	scope Scope,
	useN UseN,
	run RunInMainLoop,
	trigger Trigger,
) {

	// apply context number to relative moves
	if n := useN(); n > 0 {
		move.RelLine *= n
		move.RelRune *= n
	}

	// get current view
	view := cur()
	if view == nil {
		return
	}

	// get line
	var line int
	if move.AbsLine != nil {
		line = *move.AbsLine
	} else {
		line = view.CursorLine
		line += move.RelLine
	}

	moment := view.Moment
	maxLine := moment.NumLines() - 1
	currentPosition := view.cursorPosition()

	// get col
	var col int
	forward := true // to determine align direction
	if move.AbsCol != nil || move.AbsLine != nil {
		forward = false
	}
	if move.AbsCol != nil {
		// absolute
		col = *move.AbsCol

	} else {
		// relative
		col = view.CursorCol
		n := move.RelRune
		// convert relative runes to relative columns by iterating cells
		if n > 0 {
			// iter forward
			position := currentPosition
			if position.Line >= 0 && position.Rune >= 0 { // cursorPos may return -1, -1
				lineInfo := moment.GetLine(position.Line)
				for position.Line <= maxLine && n > 0 {
					// forward one rune
					n--
					if position.Rune >= len(lineInfo.Cells)-1 {
						// at line end, proceed next line
						col += 1
						position.Line += 1
						position.Rune = 0
						lineInfo = moment.GetLine(position.Line)
						if lineInfo == nil {
							break
						}
					} else {
						col += lineInfo.Cells[position.Rune].DisplayWidth
						position.Rune += 1
					}
				}
			}

		} else if n < 0 {
			// iter backward
			n = -n
			position := currentPosition
			if position.Line >= 0 && position.Rune >= 0 { // cursorPos may return -1, -1
				lineInfo := moment.GetLine(position.Line)
				for position.Line >= 0 && n > 0 {
					n--
					if position.Rune == 0 {
						// at line begin, proceed last line
						col -= 1
						position.Line -= 1
						lineInfo = moment.GetLine(position.Line)
						if lineInfo == nil {
							break
						}
						position.Rune = len(lineInfo.Cells) - 1
					} else {
						position.Rune -= 1
						col -= lineInfo.Cells[position.Rune].DisplayWidth
					}
				}
			}

		}
	}
	// moving up / down
	if move.RelRune == 0 && move.RelLine != 0 {
		forward = false
	}

	// wrap line and col to valid position
calculate:
	var maxCol int
	if line < 0 {
		line = 0
		goto calculate
	} else if line > maxLine {
		line = maxLine
		goto calculate
	} else {
		maxCol = moment.GetLine(line).DisplayWidth - 1
		if maxCol < 0 {
			maxCol = 0
		}
	}
	if move.RelLine != 0 && view.PreferCursorCol > col {
		// moving up or down
		col = view.PreferCursorCol
	}
	if col < 0 {
		if line == 0 {
			col = 0
		} else {
			line--
			col = moment.GetLine(line).DisplayWidth + col
			goto calculate
		}
	} else if col > maxCol {
		if forward {
			if line < maxLine {
				col = col - moment.GetLine(line).DisplayWidth
				line++
				goto calculate
			} else {
				col = maxCol
			}
		} else {
			col = maxCol
		}
	}

	// align to rune boundary
	cells := moment.GetLine(line).Cells
	for {
		n := col
		for _, cell := range cells {
			if n <= 0 {
				break
			}
			n -= cell.DisplayWidth
		}
		if n != 0 {
			if forward { // NOCOVER, above codes already done this
				col += 1
			} else {
				col -= 1
			}
		} else {
			break
		}
	}

	// no change
	if view.CursorLine == line && view.CursorCol == col {
		return
	}

	// set prefer col
	if col > view.PreferCursorCol || // prefer larger col
		move.RelRune != 0 || // moving left / right
		move.AbsCol != nil || // setting absolute col
		move.AbsLine != nil || // setting absolute line
		false {
		view.PreferCursorCol = col
	}

	// set cursor
	view.CursorLine = line
	view.CursorCol = col

	// update
	scope.Call(ScrollToCursor)

	trigger(scope.Sub(
		func() (*View, *Moment, [2]Position) {
			return view, moment, [2]Position{currentPosition, view.cursorPosition()}
		},
	), EvCursorMoved)

}

type evCursorMoved struct{}

var EvCursorMoved = new(evCursorMoved)

func (_ Command) MoveLeft() (spec CommandSpec) {
	spec.Func = func(scope Scope) {
		scope.Sub(func() Move { return Move{RelRune: -1} }).Call(MoveCursor)
	}
	return
}

func (_ Command) MoveDown() (spec CommandSpec) {
	spec.Func = func(scope Scope) {
		scope.Sub(func() Move { return Move{RelLine: 1} }).Call(MoveCursor)
	}
	return
}

func (_ Command) MoveUp() (spec CommandSpec) {
	spec.Func = func(scope Scope) {
		scope.Sub(func() Move { return Move{RelLine: -1} }).Call(MoveCursor)
	}
	return
}

func (_ Command) MoveRight() (spec CommandSpec) {
	spec.Func = func(scope Scope) {
		scope.Sub(func() Move { return Move{RelRune: 1} }).Call(MoveCursor)
	}
	return
}

func PageDown(
	cur CurrentView,
	scope Scope,
	config ScrollConfig,
) {
	view := cur()
	if view == nil {
		return
	}
	lines := view.Box.Height() - config.PaddingBottom
	line := view.ViewportLine
	line += lines
	if max := view.Moment.NumLines() - 1; line > max {
		line = max
	}
	if view.ViewportLine != line {
		view.ViewportLine = line
	}
	scope.Sub(func() Move { return Move{RelLine: lines} }).Call(MoveCursor)
}

func (_ Command) PageDown() (spec CommandSpec) {
	spec.Func = func(scope Scope) {
		scope.Call(PageDown)
	}
	return
}

func PageUp(
	cur CurrentView,
	scope Scope,
	config ScrollConfig,
) {
	view := cur()
	if view == nil {
		return
	}
	lines := view.Box.Height() - config.PaddingTop
	line := view.ViewportLine
	line -= lines
	if line < 0 {
		line = 0
	}
	if view.ViewportLine != line {
		view.ViewportLine = line
	}
	scope.Sub(func() Move { return Move{RelLine: -lines} }).Call(MoveCursor)
}

func (_ Command) PageUp() (spec CommandSpec) {
	spec.Func = func(scope Scope) {
		scope.Call(PageUp)
	}
	return
}

func NextEmptyLine(
	cur CurrentView,
	scope Scope,
) {
	view := cur()
	if view == nil {
		return
	}
	n := view.CursorLine + 1
	maxLine := view.Moment.NumLines()
	for n < maxLine {
		line := view.Moment.GetLine(n)
		if line.AllSpace {
			break
		}
		n++
	}
	scope.Sub(func() Move { return Move{AbsLine: &n, AbsCol: intP(0)} }).Call(MoveCursor)
	scope.Call(ScrollToCursor)
}

func (_ Command) NextEmptyLine() (spec CommandSpec) {
	spec.Func = func(scope Scope) {
		scope.Call(NextEmptyLine)
	}
	return
}

func PrevEmptyLine(
	cur CurrentView,
	scope Scope,
) {
	view := cur()
	if view == nil {
		return
	}
	n := view.CursorLine - 1
	for n > 0 {
		line := view.Moment.GetLine(n)
		if line.AllSpace {
			break
		}
		n--
	}
	scope.Sub(func() Move { return Move{AbsLine: &n, AbsCol: intP(0)} }).Call(MoveCursor)
	scope.Call(ScrollToCursor)
}

func (_ Command) PrevEmptyLine() (spec CommandSpec) {
	spec.Func = func(scope Scope) {
		scope.Call(PrevEmptyLine)
	}
	return
}

func LineBegin(
	cur CurrentView,
	scope Scope,
) {
	view := cur()
	if view == nil {
		return
	}
	zero := 0
	scope.Sub(func() Move {
		return Move{
			AbsCol: &zero,
		}
	}).Call(MoveCursor)
	scope.Call(ScrollToCursor)
}

func (_ Command) LineBegin() (spec CommandSpec) {
	spec.Func = func(scope Scope) {
		scope.Call(LineBegin)
	}
	return
}

func LineEnd(
	cur CurrentView,
	scope Scope,
) {
	view := cur()
	if view == nil {
		return
	}
	largeCol := math.MaxInt32
	scope.Sub(func() Move {
		return Move{
			AbsCol: &largeCol,
		}
	}).Call(MoveCursor)
	scope.Call(ScrollToCursor)
}

func (_ Command) LineEnd() (spec CommandSpec) {
	spec.Func = func(scope Scope) {
		scope.Call(LineEnd)
	}
	return
}

func NextRune() []StrokeSpec {
	return []StrokeSpec{
		{
			Predict: func() bool {
				return true
			},
			Func: func(
				ev KeyEvent,
			) (
				fn Func,
			) {
				if ev.Key() != tcell.KeyRune {
					return
				}
				toFind := ev.Rune()
				// to make imitation work
				fn = func(
					getCur CurrentView,
					scope Scope,
				) {
					cur := getCur()
					if cur == nil { // NOCOVER
						return
					}
					moment := cur.Moment
					line := moment.GetLine(cur.CursorLine)
					if line == nil { // NOCOVER
						return
					}
					// locate current cell
					col := -1
					cellIndex := -1
					for i, cell := range line.Cells {
						if col < cur.CursorCol {
							col += cell.DisplayWidth
						} else {
							cellIndex = i
							break
						}
					}
					if col == -1 || cellIndex == -1 { // NOCOVER
						return
					}
					found := false
					for _, cell := range line.Cells[cellIndex:] {
						col += cell.DisplayWidth
						if cell.Rune == toFind {
							found = true
							break
						}
					}
					if !found {
						return
					}
					scope.Sub(func() Move {
						return Move{
							AbsCol: &col,
						}
					}).Call(MoveCursor)
				}
				return
			},
		},
	}
}

func (_ Command) NextRune() (spec CommandSpec) {
	spec.Func = NextRune
	spec.Desc = "focus next specified rune in the same line"
	return
}

func PrevRune() []StrokeSpec {
	return []StrokeSpec{
		{
			Predict: func() bool {
				return true
			},
			Func: func(
				ev KeyEvent,
			) (
				fn Func,
			) {
				if ev.Key() != tcell.KeyRune {
					return
				}
				toFind := ev.Rune()
				// to make imitation work
				fn = func(
					getCur CurrentView,
					scope Scope,
				) {
					cur := getCur()
					if cur == nil { // NOCOVER
						return
					}
					moment := cur.Moment
					line := moment.GetLine(cur.CursorLine)
					if line == nil { // NOCOVER
						return
					}
					// search from line begin
					col := 0
					foundCol := -1
					for _, cell := range line.Cells {
						if col >= cur.CursorCol {
							break
						}
						if cell.Rune == toFind {
							foundCol = col
						}
						col += cell.DisplayWidth
					}
					if foundCol < 0 {
						return
					}
					scope.Sub(func() Move {
						return Move{
							AbsCol: &foundCol,
						}
					}).Call(MoveCursor)
				}
				return
			},
		},
	}
}

func (_ Command) PrevRune() (spec CommandSpec) {
	spec.Func = PrevRune
	spec.Desc = "focus previous specified rune in the same line"
	return
}

func NextLineWithRune() []StrokeSpec {
	return []StrokeSpec{
		{
			Predict: func() bool {
				return true
			},

			Func: func(
				ev KeyEvent,
			) (
				fn Func,
			) {
				if ev.Key() != tcell.KeyRune {
					return
				}
				toFind := ev.Rune()
				// return a new func to make imitation work
				fn = func(
					cur CurrentView,
					scope Scope,
				) {

					view := cur()
					if view == nil {
						return
					}
					moment := view.Moment
					for line := view.CursorLine + 1; line < moment.NumLines(); line++ {
						col := 0
						for _, cell := range moment.GetLine(line).Cells {
							if cell.Rune == toFind {
								scope.Sub(func() Move {
									return Move{
										AbsLine: intP(line),
										AbsCol:  intP(col),
									}
								}).Call(MoveCursor)
								return
							}
							col += cell.DisplayWidth
						}
					}

				}

				return fn
			},
		},
	}
}

func (_ Command) NextLineWithRune() (spec CommandSpec) {
	spec.Desc = "jump to next line with specified rune"
	spec.Func = NextLineWithRune
	return
}

func PrevLineWithRune() []StrokeSpec {
	return []StrokeSpec{
		{
			Predict: func() bool {
				return true
			},

			Func: func(
				ev KeyEvent,
			) (
				fn Func,
			) {
				if ev.Key() != tcell.KeyRune {
					return
				}
				toFind := ev.Rune()
				// return a new func to make imitation work
				fn = func(
					cur CurrentView,
					scope Scope,
				) {

					view := cur()
					if view == nil {
						return
					}
					moment := view.Moment
					for line := view.CursorLine - 1; line >= 0; line-- {
						col := 0
						for _, cell := range moment.GetLine(line).Cells {
							if cell.Rune == toFind {
								scope.Sub(func() Move {
									return Move{
										AbsLine: intP(line),
										AbsCol:  intP(col),
									}
								}).Call(MoveCursor)
								return
							}
							col += cell.DisplayWidth
						}
					}

				}

				return fn
			},
		},
	}
}

func (_ Command) PrevLineWithRune() (spec CommandSpec) {
	spec.Desc = "jump to previous line with specified rune"
	spec.Func = PrevLineWithRune
	return
}

func PrevDedentLine(
	cur CurrentView,
	scope Scope,
) {
	view := cur()
	if view == nil {
		return
	}
	n := view.CursorLine - 1
	for n > 0 {
		line := view.Moment.GetLine(n)
		nextLine := view.Moment.GetLine(n + 1)
		if line.NonSpaceDisplayOffset == nil &&
			nextLine != nil ||
			line.NonSpaceDisplayOffset != nil &&
				nextLine.NonSpaceDisplayOffset != nil &&
				*line.NonSpaceDisplayOffset < *nextLine.NonSpaceDisplayOffset {
			break
		}
		n--
	}
	if n < 0 {
		n = 0
	}
	col := 0
	if offset := view.Moment.GetLine(n).NonSpaceDisplayOffset; offset != nil {
		col = *offset
	}
	scope.Sub(func() Move {
		return Move{
			AbsLine: &n,
			AbsCol:  intP(col),
		}
	}).Call(MoveCursor)
	scope.Call(ScrollToCursor)
}

func (_ Command) PrevDedentLine() (spec CommandSpec) {
	spec.Desc = "jump to previous dedent line"
	spec.Func = PrevDedentLine
	return
}

func NextDedentLine(
	cur CurrentView,
	scope Scope,
) {
	view := cur()
	if view == nil {
		return
	}
	n := view.CursorLine + 1
	for n < view.Moment.NumLines() {
		line := view.Moment.GetLine(n)
		nextLine := view.Moment.GetLine(n + 1)
		if nextLine == nil {
			break
		}
		if line.NonSpaceDisplayOffset == nil &&
			nextLine != nil ||
			line.NonSpaceDisplayOffset != nil &&
				nextLine.NonSpaceDisplayOffset != nil &&
				*line.NonSpaceDisplayOffset < *nextLine.NonSpaceDisplayOffset {
			break
		}
		n++
	}
	if n >= view.Moment.NumLines() {
		n = view.Moment.NumLines() - 1
	}
	col := 0
	if offset := view.Moment.GetLine(n).NonSpaceDisplayOffset; offset != nil {
		col = *offset
	}
	scope.Sub(func() Move {
		return Move{
			AbsLine: &n,
			AbsCol:  intP(col),
		}
	}).Call(MoveCursor)
	scope.Call(ScrollToCursor)
}

func (_ Command) NextDedentLine() (spec CommandSpec) {
	spec.Desc = "jump to next dedent line"
	spec.Func = NextDedentLine
	return
}
