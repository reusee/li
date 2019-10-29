package li

import (
	"fmt"
	"math"
	"sort"
	"sync"
)

type ViewUIArgs struct {
	MomentID     MomentID
	Width        int
	Height       int
	IsFocus      bool
	HintsVersion int
	ViewMomentState
}

var _ Element = new(View)

func (view *View) RenderFunc() any {
	return func(
		cur CurrentView,
		defaultStyle Style,
		scope Scope,
		getStyle GetStyle,
		box Box,
		screen Screen,
		procs ViewRenderProcs,
		j AppendJournal,
		scrollConfig ScrollConfig,
		uiConfig UIConfig,
		trigger Trigger,
		getLineHints GetLineHints,
	) Element {

		moment := view.GetMoment()

		defer func() {
			trigger(scope.Sub(
				&moment, &view,
			), EvViewRendered)
		}()

		// content box
		lineNumLength := displayWidth(fmt.Sprintf("%d", view.ViewportLine))
		if l := displayWidth(fmt.Sprintf("%d", view.ViewportLine+view.Box.Height())); l > lineNumLength {
			lineNumLength = l
		}
		contentBox := view.Box
		contentBox.Left += lineNumLength + 2
		view.ContentBox = contentBox

		currentView := cur()

		// cursor position
		defer func() {
			if view == currentView {
				y := contentBox.Top
				moment := view.GetMoment()
				var lineHeight int
				scope.Sub(&moment, &[2]int{
					view.ViewportLine, view.CursorLine,
				}).Call(CalculateSumLineHeight, &lineHeight)
				y += lineHeight
				screen.ShowCursor(
					contentBox.Left+(view.CursorCol-view.ViewportCol),
					y,
				)
			}
		}()

		// line hints
		hints, version := getLineHints()

		// frame buffer cache
		args := ViewUIArgs{
			MomentID:        moment.ID,
			Width:           view.Box.Width(),
			Height:          view.Box.Height(),
			IsFocus:         view == currentView,
			HintsVersion:    version,
			ViewMomentState: view.ViewMomentState,
		}
		if view.FrameBuffer != nil && args == view.FrameBufferArgs {
			return view.FrameBuffer
		}

		// strip hints
		n := sort.Search(len(hints), func(i int) bool {
			return hints[i].Moment.ID >= moment.ID
		})
		hints = hints[n:]
		n = sort.Search(len(hints), func(i int) bool {
			return hints[i].Moment.ID > moment.ID
		})
		hints = hints[:n]

		hintStyle := getStyle("Hint")

		// line number box
		lineNumBox := view.Box
		lineNumBox.Top = contentBox.Top
		lineNumBox.Right = lineNumBox.Left + lineNumLength + 2

		frameBuffer := NewFrameBuffer(box)
		set := SetContent(frameBuffer.SetContent)
		defer func() {
			view.FrameBufferArgs = args
			view.FrameBuffer = frameBuffer
		}()

		// style
		hlStyle := getStyle("Highlight")
		lineNumStyle := defaultStyle

		// indent-based background
		indentStyle := func(style Style, lineNum int, offset int) Style {
			for lineNum >= 0 {
				line := moment.GetLine(scope, lineNum)
				if line == nil {
					break
				}
				nonSpaceOffset := line.NonSpaceDisplayOffset
				if nonSpaceOffset != nil {
					if offset >= *nonSpaceOffset {
						return darkerOrLighterStyle(
							style,
							int32(math.Min(
								float64(0+(*nonSpaceOffset)%24),
								float64(24-(*nonSpaceOffset)%24),
							)),
						)
					}
				}
				lineNum--
			}
			return style
		}

		// lines
		selectedRange := view.selectedRange(scope)
		wg := new(sync.WaitGroup)
		loopLineNum := view.ViewportLine
		loopY := contentBox.Top
		var lineHeights map[int]int
		scope.Sub(&moment, &[2]int{
			view.ViewportLine, view.ViewportLine + view.Box.Height(),
		}).Call(CalculateLineHeights, &lineHeights)
		for loopY < contentBox.Bottom {
			y := loopY
			lineNum := loopLineNum
			lineHeight := 1
			if h, ok := lineHeights[loopLineNum]; ok {
				lineHeight = h
			}
			loopY += lineHeight
			loopLineNum += 1
			wg.Add(1)
			procs <- func() {
				defer wg.Done()

				isCurrentLine := lineNum == view.CursorLine
				var lines []Line
				if lineNum < moment.NumLines() {
					lines = append(lines, *moment.GetLine(scope, lineNum))
				}

				// line number
				if isCurrentLine {
					// show absolute
					box := lineNumBox
					box.Top += (y - contentBox.Top)
					scope.Sub(
						&box, &defaultStyle, &set,
					).Call(Text(
						fmt.Sprintf("%d", lineNum+1),
						hlStyle,
						Fill(true),
						AlignRight,
						Padding(0, 1, 0, 0),
					).RenderFunc())
				} else if lineNum < moment.NumLines() {
					// show relative
					rel := lineNum - view.CursorLine
					if rel < 0 {
						rel = -rel
					}
					box := lineNumBox
					box.Top += (y - contentBox.Top)
					scope.Sub(
						&box, &defaultStyle, &set,
					).Call(Text(
						fmt.Sprintf("%d", rel),
						lineNumStyle,
						Fill(true),
						AlignLeft,
						Padding(0, 1, 0, 0),
					).RenderFunc())
				} else {
					box := lineNumBox
					box.Top += (y - contentBox.Top)
					scope.Sub(
						&box, &defaultStyle, &set,
					).Call(Text(
						"",
						lineNumStyle,
						Fill(true),
					).RenderFunc())
				}

				baseStyle := defaultStyle
				if y == contentBox.Bottom-1 {
					if currentView == view {
						baseStyle = hlStyle
					}
					baseStyle = baseStyle.Underline(true)
				}
				if isCurrentLine {
					baseStyle = darkerOrLighterStyle(baseStyle, 20)
				}
				blockStyle := baseStyle

				var hintLines []string
				n := sort.Search(len(hints), func(i int) bool {
					return hints[i].Line >= lineNum
				})
				if n < len(hints) && hints[n].Line == lineNum {
					hintLines = hints[n].Hints
				}

				for i := 0; i < lineHeight; i++ {
					var line *Line
					if i < len(lines) {
						// moment line
						line = &lines[i]
					}

					x := contentBox.Left
					if line != nil {
						// moment content

						cells := line.Cells
						skip := view.ViewportCol
						leftSkip := false
						for skip > 0 && len(cells) > 0 {
							skip -= cells[0].DisplayWidth
							cells = cells[1:]
							leftSkip = true
						}

						var cellColors []*Color
						var cellStyleFuncs []StyleFunc
						if view.Stainer != nil {
							l := LineNumber(lineNum)
							scope.Sub(
								&moment, &line, &l,
							).Call(
								view.Stainer.Line(),
								&cellColors,
								&cellStyleFuncs,
							)
						}

						// cells
						for cellNum, cell := range cells {

							// right truncated
							if x >= contentBox.Right {
								set(
									contentBox.Right-1, y,
									'>', nil,
									hlStyle,
								)
								break
							}

							// indent style
							if line.NonSpaceDisplayOffset == nil ||
								cell.DisplayOffset <= *line.NonSpaceDisplayOffset {
								blockStyle = indentStyle(baseStyle, lineNum, cell.DisplayOffset)
							}

							// style
							style := blockStyle

							// selected range style
							if selectedRange != nil && selectedRange.Contains(Position{
								Line: lineNum,
								Cell: cellNum,
							}) {
								// selected range
								style = style.Underline(true)
								style = darkerOrLighterStyle(style, 20)
							}

							if leftSkip && x == contentBox.Left {
								// left truncated
								set(
									x, y,
									'<', nil,
									hlStyle,
								)
							} else {
								// cell style
								if cellNum < len(cellColors) {
									if color := cellColors[cellNum]; color != nil {
										style = style.Foreground(*color)
									}
								} else if cellNum < len(cellStyleFuncs) {
									if fn := cellStyleFuncs[cellNum]; fn != nil {
										style = fn(style)
									}
								}
								// set content
								set(
									x, y,
									cell.Rune, nil,
									style,
								)
							}

							if cell.DisplayWidth > cell.Width {
								// expanded tabs
								for i := 0; i < cell.DisplayWidth-cell.Width; i++ {
									if x+1+i >= contentBox.Right {
										break
									}
									set(
										x+1+i, y,
										' ', nil,
										style,
									)
								}
							}

							x += cell.DisplayWidth
						}

					} else if i >= len(lines) && i < len(lines)+len(hintLines) {
						// hint
						content := hintLines[i-len(lines)]
						runes := []rune(content)
						for _, r := range runes {
							set(
								x, y,
								r, nil,
								hintStyle,
							)
							x += runeWidth(r)
						}
					}

					// fill blank
					for ; x < contentBox.Right; x++ {
						offset := x - contentBox.Left
						if line == nil ||
							line.NonSpaceDisplayOffset == nil ||
							offset <= *line.NonSpaceDisplayOffset {
							blockStyle = indentStyle(baseStyle, lineNum, offset)
						}
						set(
							x, y,
							' ', nil,
							blockStyle,
						)
					}

					y++
				}

			}
		}
		wg.Wait()

		return frameBuffer
	}
}

type evViewRendered struct{}

var EvViewRendered = new(evViewRendered)

type ViewRenderProcs chan func()

func (_ Provide) ViewRenderProcs() (
	ch ViewRenderProcs,
) {

	ch = make(chan func(), numCPU*8)
	for i := 0; i < numCPU; i++ {
		go func() {
			for {
				(<-ch)()
			}
		}()
	}

	return
}
