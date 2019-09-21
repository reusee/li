package li

import (
	"fmt"
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
	) Element {

		currentView := cur()

		// content box
		lineNumLength := displayWidth(fmt.Sprintf("%d", view.ViewportLine))
		if l := displayWidth(fmt.Sprintf("%d", view.ViewportLine+view.Box.Height())); l > lineNumLength {
			lineNumLength = l
		}
		contentBox := view.Box
		contentBox.Left += lineNumLength + 2

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
			MomentID:        view.Moment.ID,
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
		//TODO optimize
		indentStyle := func(style Style, lineNum int, offset int) Style {
			for lineNum >= 0 {
				line := view.Moment.GetLine(lineNum)
				if line == nil {
					break
				}
				nonSpaceOffset := line.NonSpaceOffset
				if nonSpaceOffset != nil {
					if offset >= *nonSpaceOffset {
						return darkerOrLighterStyle(
							style,
							int32(1*(*nonSpaceOffset)),
						)
					}
				}
				lineNum--
			}
			return style
		}

		// lines
		moment := view.Moment
		selectedRange := view.selectedRange()
		for i := 0; i < contentBox.Height(); i++ {
			lineNum := view.ViewportLine + i
			isCurrentLine := lineNum == view.CursorLine
			var line *Line
			if lineNum < moment.NumLines() {
				line = moment.GetLine(lineNum)
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

			if line != nil {

				cells := line.Cells
				skip := view.ViewportCol
				leftSkip := false
				for skip > 0 && len(cells) > 0 {
					skip -= cells[0].DisplayWidth
					cells = cells[1:]
					leftSkip = true
				}

				lineStyle := defaultStyle
				if y == contentBox.Bottom-1 {
					if currentView == view {
						lineStyle = hlStyle
					}
					lineStyle = lineStyle.Underline(true)
				}
				if isCurrentLine {
					lineStyle = darkerOrLighterStyle(lineStyle, 20)
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

					// style
					style := lineStyle

					// selected range style
					if selectedRange != nil && selectedRange.Contains(Position{
						Line: lineNum,
						Rune: cellNum,
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
						style := indentStyle(style, lineNum, x-contentBox.Left)
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

					if cell.DisplayWidth > cell.RuneWidth {
						// expanded tabs
						for i := 0; i < cell.DisplayWidth-cell.RuneWidth; i++ {
							if x+1+i >= contentBox.Right {
								break
							}
							set(
								x+1+i, y,
								' ', nil,
								indentStyle(style, lineNum, x-contentBox.Left),
							)
						}
					}

					x += cell.DisplayWidth
				}
			}

			style := defaultStyle
			if y == contentBox.Bottom-1 {
				// current view
				if currentView == view {
					style = hlStyle
				}
				style = style.Underline(true)
			} else {
				// current cursor line
				if isCurrentLine {
					style = darkerOrLighterStyle(style, 20)
				}
			}

			// fill blank
			for ; x < contentBox.Right; x++ {
				set(
					x, y,
					' ', nil,
					indentStyle(style, lineNum, x-contentBox.Left),
				)
			}

		}

		return frameBuffer
	}
}
