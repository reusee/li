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

type MoveCursor func(
	move Move,
)

func (_ Provide) MoveCursor(
	cur CurrentView,
	scope Scope,
	withN WithContextNumber,
	run RunInMainLoop,
	trigger Trigger,
	scrollToCursor ScrollToCursor,
) MoveCursor {

	return func(move Move) {
		// apply context number to relative moves
		withN(func(n int) {
			if n > 0 {
				move.RelLine *= n
				move.RelRune *= n
			}
		})

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

		moment := view.GetMoment()
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
				if position.Line >= 0 && position.Cell >= 0 { // cursorPos may return -1, -1
					lineInfo := moment.GetLine(position.Line)
					for position.Line <= maxLine && n > 0 {
						// forward one rune
						n--
						if position.Cell >= len(lineInfo.Cells)-1 {
							// at line end, proceed next line
							col += 1
							position.Line += 1
							position.Cell = 0
							lineInfo = moment.GetLine(position.Line)
							if lineInfo == nil {
								break
							}
						} else {
							col += lineInfo.Cells[position.Cell].DisplayWidth
							position.Cell += 1
						}
					}
				}

			} else if n < 0 {
				// iter backward
				n = -n
				position := currentPosition
				if position.Line >= 0 && position.Cell >= 0 { // cursorPos may return -1, -1
					lineInfo := moment.GetLine(position.Line)
					for position.Line >= 0 && n > 0 {
						n--
						if position.Cell == 0 {
							// at line begin, proceed last line
							col -= 1
							position.Line -= 1
							lineInfo = moment.GetLine(position.Line)
							if lineInfo == nil {
								break
							}
							position.Cell = len(lineInfo.Cells) - 1
						} else {
							position.Cell -= 1
							col -= lineInfo.Cells[position.Cell].DisplayWidth
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
		scrollToCursor()

		trigger(scope.Sub(
			&view, &moment, &[2]Position{currentPosition, view.cursorPosition()},
		), EvCursorMoved)
	}

}

type evCursorMoved struct{}

var EvCursorMoved = new(evCursorMoved)

type PageDown func()

func (_ Provide) PageDown(
	cur CurrentView,
	scope Scope,
	config ScrollConfig,
	moveCursor MoveCursor,
) PageDown {
	return func() {
		view := cur()
		if view == nil {
			return
		}
		moment := view.GetMoment()

		scrollHeight := view.Box.Height() - config.PaddingBottom
		line := view.ViewportLine
		var lineHeights map[int]int
		scope.Sub(&moment, &[2]int{line, line + scrollHeight}).
			Call(CalculateLineHeights, &lineHeights)
		scrollLines := 0
		for {
			if h, ok := lineHeights[line]; ok {
				scrollHeight -= h
			} else {
				scrollHeight--
			}
			if scrollHeight < 0 {
				break
			}
			if line > moment.NumLines()-1 {
				break
			}
			line++
			scrollLines++
		}

		if view.ViewportLine != line {
			view.ViewportLine = line
		}
		moveCursor(Move{RelLine: scrollLines})
	}
}

type PageUp func()

func (_ Provide) PageUp(
	cur CurrentView,
	scope Scope,
	config ScrollConfig,
	moveCursor MoveCursor,
) PageUp {
	return func() {
		view := cur()
		if view == nil {
			return
		}
		moment := view.GetMoment()

		scrollHeight := view.Box.Height() - config.PaddingTop
		line := view.ViewportLine
		var lineHeights map[int]int
		scope.Sub(&moment, &[2]int{line - scrollHeight - 1, line}).
			Call(CalculateLineHeights, &lineHeights)
		for {
			l := line - 1
			if l < 0 {
				break
			}
			if h, ok := lineHeights[l]; ok {
				scrollHeight -= h
			} else {
				scrollHeight--
			}
			if scrollHeight < 0 {
				break
			}
			line--
		}
		lines := view.ViewportLine - line
		if line == 0 && scrollHeight > 0 {
			// viewport not moving, set cursor line to 0
			lines = view.CursorLine
		}

		if view.ViewportLine != line {
			view.ViewportLine = line
		}
		moveCursor(Move{RelLine: -lines})
	}
}

type NextEmptyLine func()

func (_ Provide) NextEmptyLine(
	cur CurrentView,
	moveCursor MoveCursor,
	scrollToCursor ScrollToCursor,
) NextEmptyLine {
	return func() {
		view := cur()
		if view == nil {
			return
		}
		n := view.CursorLine + 1
		moment := view.GetMoment()
		maxLine := moment.NumLines()
		for n < maxLine {
			line := moment.GetLine(n)
			if line.AllSpace {
				break
			}
			n++
		}
		moveCursor(Move{AbsLine: &n, AbsCol: intP(0)})
		scrollToCursor()
	}
}

type PrevEmptyLine func()

func (_ Provide) PrevEmptyLine(
	cur CurrentView,
	moveCursor MoveCursor,
	scrollToCursor ScrollToCursor,
) PrevEmptyLine {
	return func() {
		view := cur()
		if view == nil {
			return
		}
		n := view.CursorLine - 1
		for n > 0 {
			line := view.GetMoment().GetLine(n)
			if line.AllSpace {
				break
			}
			n--
		}
		moveCursor(Move{AbsLine: &n, AbsCol: intP(0)})
		scrollToCursor()
	}
}

type LineBegin func()

func (_ Provide) LineBegin(
	cur CurrentView,
	moveCursor MoveCursor,
	scrollToCursor ScrollToCursor,
) LineBegin {
	return func() {
		view := cur()
		if view == nil {
			return
		}
		zero := 0
		moveCursor(Move{
			AbsCol: &zero,
		})
		scrollToCursor()
	}
}

type LineEnd func()

func (_ Provide) LineEnd(
	cur CurrentView,
	moveCursor MoveCursor,
	scrollToCursor ScrollToCursor,
) LineEnd {
	return func() {
		view := cur()
		if view == nil {
			return
		}
		largeCol := math.MaxInt32
		moveCursor(Move{
			AbsCol: &largeCol,
		})
		scrollToCursor()
	}
}

func NextRune() []StrokeSpec {
	return []StrokeSpec{
		{
			Predict: func() bool {
				return true
			},
			Hints: []string{
				"press any key to jump",
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
				fn = func(
					getCur CurrentView,
					moveCursor MoveCursor,
				) {
					cur := getCur()
					if cur == nil { // NOCOVER
						return
					}
					moment := cur.GetMoment()
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
					moveCursor(Move{
						AbsCol: &col,
					})
				}
				return
			},
		},
	}
}

func PrevRune() []StrokeSpec {
	return []StrokeSpec{
		{
			Predict: func() bool {
				return true
			},
			Hints: []string{
				"press any key to jump",
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
				fn = func(
					getCur CurrentView,
					moveCursor MoveCursor,
				) {
					cur := getCur()
					if cur == nil { // NOCOVER
						return
					}
					moment := cur.GetMoment()
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
					moveCursor(Move{
						AbsCol: &foundCol,
					})
				}
				return
			},
		},
	}
}

func NextLineWithRune() []StrokeSpec {
	return []StrokeSpec{
		{
			Predict: func() bool {
				return true
			},
			Hints: []string{
				"press any key to jump",
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
				fn = func(
					cur CurrentView,
					moveCursor MoveCursor,
				) {

					view := cur()
					if view == nil {
						return
					}
					moment := view.GetMoment()
					for line := view.CursorLine + 1; line < moment.NumLines(); line++ {
						col := 0
						for _, cell := range moment.GetLine(line).Cells {
							if cell.Rune == toFind {
								moveCursor(Move{
									AbsLine: intP(line),
									AbsCol:  intP(col),
								})
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

func PrevLineWithRune() []StrokeSpec {
	return []StrokeSpec{
		{
			Predict: func() bool {
				return true
			},
			Hints: []string{
				"press any key to jump",
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
				fn = func(
					cur CurrentView,
					moveCursor MoveCursor,
				) {

					view := cur()
					if view == nil {
						return
					}
					moment := view.GetMoment()
					for line := view.CursorLine - 1; line >= 0; line-- {
						col := 0
						for _, cell := range moment.GetLine(line).Cells {
							if cell.Rune == toFind {
								moveCursor(Move{
									AbsLine: intP(line),
									AbsCol:  intP(col),
								})
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

type PrevDedentLine func()

func (_ Provide) PrevDedentLine(
	cur CurrentView,
	moveCursor MoveCursor,
	scrollToCursor ScrollToCursor,
) PrevDedentLine {
	return func() {
		view := cur()
		if view == nil {
			return
		}
		n := view.CursorLine - 1
		moment := view.GetMoment()
		for n > 0 {
			line := moment.GetLine(n)
			nextLine := moment.GetLine(n + 1)
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
		if offset := moment.GetLine(n).NonSpaceDisplayOffset; offset != nil {
			col = *offset
		}
		moveCursor(Move{
			AbsLine: &n,
			AbsCol:  intP(col),
		})
		scrollToCursor()
	}
}

type NextDedentLine func()

func (_ Provide) NextDedentLine(
	cur CurrentView,
	scrollToCursor ScrollToCursor,
	moveCursor MoveCursor,
) NextDedentLine {
	return func() {
		view := cur()
		if view == nil {
			return
		}
		n := view.CursorLine + 1
		moment := view.GetMoment()
		for n < moment.NumLines() {
			line := moment.GetLine(n)
			nextLine := moment.GetLine(n + 1)
			if nextLine == nil {
				break
			}
			if line.NonSpaceDisplayOffset == nil &&
				line.NonSpaceDisplayOffset != nil &&
				nextLine.NonSpaceDisplayOffset != nil &&
				*line.NonSpaceDisplayOffset < *nextLine.NonSpaceDisplayOffset {
				break
			}
			n++
		}
		if n >= moment.NumLines() {
			n = moment.NumLines() - 1
		}
		col := 0
		if offset := moment.GetLine(n).NonSpaceDisplayOffset; offset != nil {
			col = *offset
		}
		moveCursor(Move{
			AbsLine: &n,
			AbsCol:  intP(col),
		})
		scrollToCursor()
	}
}
