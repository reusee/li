package li

import (
	"fmt"
	"sync"

	"github.com/gdamore/tcell"
)

type SelectionDialog struct {
	Title string

	OnClose  func(scope Scope)
	OnSelect func(scope Scope, id ID)
	OnUpdate func(scope Scope, runes []rune) (
		ids []ID,
		maxLen int,
		initialIDIndex int,
	)
	CandidateElement func(scope Scope, id ID) Element

	runes              []rune
	runesLen           int
	index              int
	candidates         []ID
	maxCandidateLength int
	initOnce           sync.Once
}

var _ Element = new(SelectionDialog)

var _ KeyStrokeHandler = new(SelectionDialog)

func (d *SelectionDialog) RenderFunc() any {
	return func(
		box Box,
		scope Scope,
		getStyle GetStyle,
		defaultStyle Style,
	) {

		d.initOnce.Do(func() {
			d.updateCandidates(scope)
		})

		marginTop := 1
		marginRight := 2
		marginBottom := 1
		marginLeft := 2
		paddingTop := 1
		paddingRight := 2
		paddingBottom := 1
		paddingLeft := 2

		contentLength := 0

		if w := displayWidth(d.Title); w > contentLength {
			contentLength = w
		}
		if d.runesLen > contentLength {
			contentLength = d.runesLen
		}
		if d.maxCandidateLength > contentLength {
			contentLength = d.maxCandidateLength
		}

		maxContentLength := box.Width() - marginLeft - paddingLeft - marginRight - paddingRight
		if contentLength > maxContentLength {
			contentLength = maxContentLength
		}

		contentHeight := 1 + // title
			1 + // input
			len(d.candidates) // candidates

		maxContentHeight := box.Height() - marginTop - paddingTop - marginBottom - paddingBottom
		if contentHeight > maxContentHeight {
			contentHeight = maxContentHeight
		}

		marginBox := Box{
			Left:   box.Left,
			Right:  box.Left + marginLeft + paddingLeft + contentLength + paddingRight + marginRight,
			Top:    box.Top,
			Bottom: box.Top + marginTop + paddingTop + contentHeight + paddingBottom + marginBottom,
		}
		paddingBox := Box{
			Left:   box.Left + marginLeft,
			Right:  box.Left + marginLeft + paddingLeft + contentLength + paddingRight,
			Top:    box.Top + marginTop,
			Bottom: box.Top + marginTop + paddingTop + contentHeight + paddingBottom,
		}
		contentBox := Box{
			Left:   box.Left + marginLeft + paddingLeft,
			Right:  box.Left + marginLeft + paddingLeft + contentLength,
			Top:    box.Top + marginTop + paddingTop,
			Bottom: box.Top + marginTop + paddingTop + contentHeight,
		}

		style := darkerOrLighterStyle(defaultStyle, 40)

		titleElement := Text(
			Box{
				contentBox.Top, contentBox.Left, contentBox.Top + 1, contentBox.Right,
			},
			AlignCenter,
			style.Bold(true),
			d.Title,
		)

		inputLength := d.runesLen
		inputRunes := d.runes
		for inputLength > maxContentLength {
			// truncate
			inputRunes = inputRunes[1:]
			inputLength = runesDisplayWidth(inputRunes)
		}
		inputElement := Rect(
			Box{
				contentBox.Top + 1, contentBox.Left, contentBox.Top + 2, contentBox.Right,
			},
			style,
			Fill(true),
			Text(
				Box{
					contentBox.Top + 1, contentBox.Left, contentBox.Top + 2, contentBox.Right,
				},
				darkerOrLighterStyle(style, -10),
				Fill(true),
				string(inputRunes),
				func(box Box, screen Screen) {
					screen.ShowCursor(box.Left+inputLength, box.Top)
				},
			),
		)

		var candidateElements []Element
		if d.CandidateElement != nil {
			maxLines := contentBox.Height() - 2
			var viewportBegin int
			if len(d.candidates) > box.Height()-5 {
				viewportBegin = d.index - maxLines/2
			} else {
				viewportBegin = 0
			}
			if viewportBegin < 0 {
				viewportBegin = 0
			}
			viewportEnd := viewportBegin + maxLines
			numTopTruncated := 0
			numBottomTruncated := 0

			for i, id := range d.candidates {
				if i < viewportBegin {
					numTopTruncated++
					continue
				} else if i >= viewportEnd {
					numBottomTruncated++
					continue
				}
				candidateBox := Box{
					contentBox.Top + 2 + len(candidateElements), contentBox.Left,
					contentBox.Top + 3 + len(candidateElements), contentBox.Right,
				}
				candidateElements = append(
					candidateElements,
					d.CandidateElement(
						scope.Sub(
							&candidateBox,
							&style,
							&d.candidates[d.index],
						),
						id,
					),
				)
			}

			if numTopTruncated > 0 {
				str := fmt.Sprintf("%d..", numTopTruncated)
				hlStyle := getStyle("Highlight")(style)
				fg, _, _ := hlStyle.Decompose()
				candidateElements = append(candidateElements, Text(
					Box{
						contentBox.Top + 2, contentBox.Left,
						contentBox.Top + 3, contentBox.Left + displayWidth(str),
					},
					str,
					style.Foreground(fg),
				))
			}

			if numBottomTruncated > 0 {
				str := fmt.Sprintf("%d..", numBottomTruncated)
				hlStyle := getStyle("Highlight")(style)
				fg, _, _ := hlStyle.Decompose()
				candidateElements = append(candidateElements, Text(
					Box{
						contentBox.Bottom - 1, contentBox.Left,
						contentBox.Bottom, contentBox.Left + displayWidth(str),
					},
					str,
					style.Foreground(fg),
				))
			}
		}

		element := Rect(
			marginBox,

			Rect(
				paddingBox,
				style,
				Fill(true),

				Rect(
					contentBox,
					titleElement,
					inputElement,
					candidateElements,
				),
			),
		)

		renderAll(scope, element)
	}
}

func (d *SelectionDialog) StrokeSpecs() any {
	return func() []StrokeSpec {

		return []StrokeSpec{
			{
				Predict: func() bool {
					return true
				},
				Func: func(ev KeyEvent, scope Scope) {

					switch ev.Key() {

					case tcell.KeyEscape:
						if d.OnClose != nil {
							d.OnClose(scope)
						}

					case tcell.KeyBackspace2, tcell.KeyBackspace:
						if len(d.runes) > 0 {
							d.runes = d.runes[:len(d.runes)-1]
						}
						d.updateCandidates(scope)

					case tcell.KeyRune:
						d.runes = append(d.runes, ev.Rune())
						d.updateCandidates(scope)

					case tcell.KeyEnter:
						if d.OnSelect != nil && d.index < len(d.candidates) {
							d.OnSelect(scope, d.candidates[d.index])
						}
						if d.OnClose != nil {
							d.OnClose(scope)
						}

					case tcell.KeyUp, tcell.KeyCtrlP:
						d.index--
						if d.index < 0 {
							d.index = len(d.candidates) - 1
						}

					case tcell.KeyDown, tcell.KeyCtrlN:
						d.index++
						if d.index >= len(d.candidates) {
							d.index = 0
						}

					}

				},
			},
		}
	}
}

func (d *SelectionDialog) updateCandidates(scope Scope) {
	d.runesLen = runesDisplayWidth(d.runes)
	d.index = 0
	if d.OnUpdate != nil {
		d.candidates, d.maxCandidateLength, d.index = d.OnUpdate(scope, d.runes)
	}
}
