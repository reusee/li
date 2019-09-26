package li

import (
	"fmt"
	"math"
	"sync"
)

type ViewUIArgs struct {
	MomentID MomentID
	Width    int
	Height   int
	IsFocus  bool
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
	) Element {

		moment := view.GetMoment()

		defer func() {
			trigger(scope.Sub(
				func() (*Moment, *View) {
					return moment, view
				},
			), EvViewRendered)
		}()

		// outline
		displayOffset := 0
		for i := 0; i < box.Height(); i++ {
			lineNum := view.ViewportLine + i
			line := moment.GetLine(scope, lineNum)
			if line == nil {
				continue
			}
			if line.NonSpaceDisplayOffset == nil {
				continue
			}
			displayOffset = *line.NonSpaceDisplayOffset
			break
		}
		var outlineLines []int
		for i := view.ViewportLine; i >= 0 && i >= view.ViewportLine-uiConfig.MaxOutlineDistance; i-- {
			line := moment.GetLine(scope, i)
			if line == nil {
				continue
			}
			if line.NonSpaceDisplayOffset == nil {
				continue
			}
			if *line.NonSpaceDisplayOffset < displayOffset {
				outlineLines = append(outlineLines, i)
				displayOffset = *line.NonSpaceDisplayOffset
			}
		}
		var paddingTop, paddingBottom int
		if view.Box.Height() > scrollConfig.PaddingTop+scrollConfig.PaddingBottom {
			paddingTop = scrollConfig.PaddingTop
			paddingBottom = scrollConfig.PaddingBottom
		}
		if len(outlineLines) > paddingTop {
			outlineLines = outlineLines[len(outlineLines)-paddingTop:]
		}
		if len(outlineLines) > paddingBottom {
			outlineLines = outlineLines[len(outlineLines)-paddingBottom:]
		}

		// content box
		lineNumLength := displayWidth(fmt.Sprintf("%d", view.ViewportLine))
		if l := displayWidth(fmt.Sprintf("%d", view.ViewportLine+view.Box.Height())); l > lineNumLength {
			lineNumLength = l
		}
		contentBox := view.Box
		contentBox.Left += lineNumLength + 2
		contentBox.Top += len(outlineLines)

		currentView := cur()

		// cursor position
		defer func() {
			if view == currentView {
				screen.ShowCursor(
					contentBox.Left+(view.CursorCol-view.ViewportCol),
					contentBox.Top+(view.CursorLine-view.ViewportLine),
				)
			}
		}()

		// frame buffer cache
		args := ViewUIArgs{
			MomentID:        moment.ID,
			Width:           view.Box.Width(),
			Height:          view.Box.Height(),
			IsFocus:         view == currentView,
			ViewMomentState: view.ViewMomentState,
		}
		if view.FrameBuffer != nil && args == view.FrameBufferArgs {
			return view.FrameBuffer
		}

		// line number box
		lineNumBox := view.Box
		lineNumBox.Top = contentBox.Top
		lineNumBox.Right = lineNumBox.Left + lineNumLength + 2

		frameBuffer := NewFrameBuffer(box)
		set := frameBuffer.SetContent
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
		for i := 0; i < contentBox.Height(); i++ {
			i := i
			wg.Add(1)
			procs <- func() {
				defer wg.Done()

				lineNum := view.ViewportLine + i
				isCurrentLine := lineNum == view.CursorLine
				var line *Line
				if lineNum < moment.NumLines() {
					line = moment.GetLine(scope, lineNum)
				}
				y := contentBox.Top + i
				x := contentBox.Left

				// line number
				if isCurrentLine {
					// show absolute
					scope.Sub(func() (Box, Style, SetContent) {
						box := lineNumBox
						box.Top += i
						return box, defaultStyle, set
					}).Call(Text(
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
					scope.Sub(func() (Box, Style, SetContent) {
						box := lineNumBox
						box.Top += i
						return box, defaultStyle, set
					}).Call(Text(
						fmt.Sprintf("%d", rel),
						lineNumStyle,
						Fill(true),
						AlignLeft,
						Padding(0, 1, 0, 0),
					).RenderFunc())
				} else {
					scope.Sub(func() (Box, Style, SetContent) {
						box := lineNumBox
						box.Top += i
						return box, defaultStyle, set
					}).Call(Text(
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
				lineStyle := baseStyle

				if line != nil {

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
						scope.Sub(func() (*Moment, *Line, LineNumber) {
							return moment, line, LineNumber(lineNum)
						}).Call(
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
							lineStyle = indentStyle(baseStyle, lineNum, cell.DisplayOffset)
						}

						// style
						style := lineStyle

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
									lineStyle,
								)
							}
						}

						x += cell.DisplayWidth
					}
				}

				// fill blank
				for ; x < contentBox.Right; x++ {
					offset := x - contentBox.Left
					if line == nil ||
						line.NonSpaceDisplayOffset == nil ||
						offset <= *line.NonSpaceDisplayOffset {
						lineStyle = indentStyle(baseStyle, lineNum, offset)
					}
					set(
						x, y,
						' ', nil,
						lineStyle,
					)
				}
			}
		}
		wg.Wait()

		// outlines
		outlineBox := contentBox
		outlineBox.Top = view.Box.Top
		outlineBox.Bottom = contentBox.Top - 1
		lineNumBox.Bottom = lineNumBox.Top - 1
		lineNumBox.Top = outlineBox.Top
		for i, lineNum := range outlineLines {
			x := outlineBox.Left
			y := outlineBox.Top + (len(outlineLines) - i - 1)

			baseStyle := defaultStyle
			if i == 0 {
				baseStyle = baseStyle.Underline(true)
			}
			lineStyle := baseStyle
			line := moment.GetLine(scope, lineNum)

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
				scope.Sub(func() (*Moment, *Line, LineNumber) {
					return moment, line, LineNumber(lineNum)
				}).Call(
					view.Stainer.Line(),
					&cellColors,
					&cellStyleFuncs,
				)
			}

			// show absolute line number
			scope.Sub(func() (Box, Style, SetContent) {
				box := lineNumBox
				box.Top = y
				return box, defaultStyle, set
			}).Call(Text(
				fmt.Sprintf("%d", lineNum+1),
				hlStyle,
				Fill(true),
				AlignRight,
				Padding(0, 1, 0, 0),
			).RenderFunc())

			// cells
			for cellNum, cell := range cells {

				// right truncated
				if x >= outlineBox.Right {
					set(
						outlineBox.Right-1, y,
						'>', nil,
						hlStyle,
					)
					break
				}

				// indent style
				if line.NonSpaceDisplayOffset == nil ||
					cell.DisplayOffset <= *line.NonSpaceDisplayOffset {
					lineStyle = indentStyle(baseStyle, lineNum, cell.DisplayOffset)
				}

				style := lineStyle

				if leftSkip && x == outlineBox.Left {
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
						if x+1+i >= outlineBox.Right {
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

			// fill blank
			for ; x < outlineBox.Right; x++ {
				offset := x - contentBox.Left
				if line == nil ||
					line.NonSpaceDisplayOffset == nil ||
					offset <= *line.NonSpaceDisplayOffset {
					lineStyle = indentStyle(baseStyle, lineNum, offset)
				}
				set(
					x, y,
					' ', nil,
					lineStyle,
				)
			}

		}

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
