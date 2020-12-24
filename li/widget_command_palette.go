package li

import (
	"sort"

	"github.com/gdamore/tcell"
	"github.com/junegunn/fzf/src/util"
)

func ShowCommandPalette(
	scope Scope,
	screen Screen,
	commands Commands,
	run RunInMainLoop,
	pushOverlay PushOverlay,
) {

	// initial candidates
	type Candidate struct {
		Left          string
		LeftMatchLen  int
		Right         string
		RightMatchLen int
		Func          Func
		Score         int

		LeftChars  *util.Chars
		RightChars *util.Chars
	}
	var initialCandidates []Candidate
	leftLen := 0
	rightLen := 0
	for _, spec := range commands {
		leftText := spec.Name
		rightText := spec.Desc
		if l := displayWidth(leftText); l > leftLen {
			leftLen = l
		}
		if l := displayWidth(rightText); l > rightLen {
			rightLen = l
		}
		leftChars := util.RunesToChars([]rune(leftText))
		rightChars := util.RunesToChars([]rune(rightText))
		candidate := Candidate{
			Left:  leftText,
			Right: rightText,
			Func:  spec.Func,

			LeftChars:  &leftChars,
			RightChars: &rightChars,
		}
		initialCandidates = append(initialCandidates, candidate)
	}
	sort.SliceStable(initialCandidates, func(i, j int) bool {
		l1 := len(initialCandidates[i].Left)
		l2 := len(initialCandidates[j].Left)
		if l1 != l2 {
			return l1 < l2
		}
		return initialCandidates[i].Left > initialCandidates[j].Left
	})

	// states
	var runes []rune
	runesLen := 0
	var candidates []Candidate
	index := 0
	updateCandidates := func() {
	do:
		// reset states
		index = 0
		runesLen = runesDisplayWidth(runes)
		// set candidates
		if len(runes) == 0 {
			candidates = initialCandidates
			return
		}
		var newCandidates []Candidate
		for _, candidate := range candidates {
			leftMatched, leftMatchLen, score1 := fuzzyMatched(runes, candidate.LeftChars)
			rightMatched, rightMatchLen, score2 := fuzzyMatched(runes, candidate.RightChars)
			if !leftMatched && !rightMatched {
				continue
			}
			candidate.LeftMatchLen = leftMatchLen
			candidate.RightMatchLen = rightMatchLen
			candidate.Score = score1
			if score2 > score1 {
				candidate.Score = score2
			}
			newCandidates = append(newCandidates, candidate)
		}
		if len(runes) > 0 && len(newCandidates) == 0 {
			runes = runes[:len(runes)-1]
			goto do
		}
		candidates = newCandidates
		sort.SliceStable(candidates, func(i, j int) bool {
			return candidates[i].Score > candidates[j].Score
		})
	}
	updateCandidates()

	paddingTop := 1
	paddingLeft := 2
	paddingRight := 2
	paddingBottom := 1
	marginTop := 2
	marginLeft := 4
	marginRight := 4
	marginBottom := 2

	var id ID
	palette := WidgetDialog{

		OnKey: func(
			ev KeyEvent,
			scope Scope,
			closeOverlay CloseOverlay,
		) {
			switch ev.Key() {

			case tcell.KeyEscape:
				// close
				closeOverlay(id)

			case tcell.KeyBackspace2, tcell.KeyBackspace:
				if len(runes) > 0 {
					runes = runes[:len(runes)-1]
				}
				updateCandidates()

			case tcell.KeyRune:
				runes = append(runes, ev.Rune())
				updateCandidates()

			case tcell.KeyEnter:
				run(func(scope Scope) {
					fn := candidates[index].Func
					scope.Sub(&fn).Call(ExecuteCommandFunc)
				})
				closeOverlay(id)

			case tcell.KeyUp, tcell.KeyCtrlP:
				index--
				if index < 0 {
					index = len(candidates) - 1
				}

			case tcell.KeyDown, tcell.KeyCtrlN:
				index++
				if index >= len(candidates) {
					index = 0
				}

			}
		},

		Element: ElementFrom(func(
			set SetContent,
			box Box,
			getStyle GetStyle,
			defaultStyle Style,
		) Element {

			//TODO scrollable

			top := box.Top + marginTop
			length := paddingLeft + leftLen + 1 + rightLen + paddingRight
			if length > box.Width()-marginLeft-marginRight {
				length = box.Width() - marginLeft - marginRight
			}
			left := box.Left
			right := left + length
			maxBottom := box.Bottom - marginBottom
			bottom := top + paddingTop + 1 + len(candidates) + paddingBottom
			if bottom > maxBottom {
				bottom = maxBottom
			}
			style := darkerOrLighterStyle(defaultStyle, -10)
			hlStyle := getStyle("Highlight")

			box = Box{
				Top:    top,
				Left:   left,
				Right:  right,
				Bottom: bottom,
			}

			return Rect(
				box,
				Fill(true),
				Padding(paddingTop, paddingRight, paddingBottom, paddingLeft),
				style,

				// input area
				Text(
					func(parent Box) Box {
						return Box{
							Left:   parent.Left,
							Right:  parent.Right,
							Top:    parent.Top,
							Bottom: parent.Top + 1,
						}
					},
					string(runes),
					style.Underline(true),
					func(box Box) {
						screen.ShowCursor(box.Left+runesLen, box.Top)
					},
				),

				// candidates
				func() (ret []Element) {

					for i := 0; i < len(candidates) && box.Top+1+i < box.Bottom; i++ {
						i := i

						s := style
						if i == index {
							s = hlStyle(s)
						}

						candidate := candidates[i]

						ret = append(ret, Text(
							func(parent Box) Box {
								return Box{
									Top:    parent.Top + 1 + i,
									Left:   parent.Left,
									Right:  parent.Right,
									Bottom: parent.Top + 1 + i,
								}
							},
							s,
							rightPad(candidate.Left, ' ', leftLen)+
								" "+
								rightPad(candidate.Right, ' ', rightLen),
							OffsetStyleFunc(func(i int) StyleFunc {
								fn := SameStyle
								if i < leftLen {
									fn = fn.SetBold(true)
								} else {
									fn = fn.SetBold(false)
								}
								if i < leftLen && i < candidate.LeftMatchLen {
									fn = fn.SetUnderline(true)
								} else if i > leftLen && i-leftLen-1 < candidate.RightMatchLen {
									fn = fn.SetUnderline(true)
								} else {
									fn = fn.SetUnderline(false)
								}
								return fn
							}),
						))

					}

					return
				}(),
			)

		}),
	}

	overlay := OverlayObject(palette)
	id = pushOverlay(overlay)
}

func (_ Command) ShowCommandPalette() (spec CommandSpec) {
	spec.Func = func(scope Scope) {
		scope.Call(ShowCommandPalette)
	}
	return
}
