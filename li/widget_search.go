package li

import (
	"fmt"
	"regexp"
	"strconv"
)

func ShowSearchDialog(scope Scope) {

	type Result struct {
		LineNumber      int
		Content         string
		BeginRuneOffset int
		EndRuneOffset   int
	}
	var results []Result

	var id ID
	dialog := &SelectionDialog{

		Title: "Search",

		OnClose: func(_ Scope) {
			scope.Sub(func() ID { return id }).Call(CloseOverlay)
		},

		OnSelect: func(_ Scope, id ID) {
			scope.Sub(func() ID { return id }).Call(CloseOverlay)
			scope.Sub(func() Move {
				return Move{AbsLine: intP(results[id].LineNumber - 1)}
			}).Call(MoveCursor)
		},

		OnUpdate: func(scope Scope, runes []rune) (ids []ID, maxLen int, initIndex int) {
			if len(runes) == 0 {
				return
			}
			results = results[:0]

			var cur CurrentView
			scope.Assign(&cur)
			view := cur()
			pattern, err := regexp.Compile("(?i)" + string(runes))
			if err != nil {
				errorStr := err.Error()
				results = []Result{
					{
						Content:         errorStr,
						BeginRuneOffset: 0,
						EndRuneOffset:   len([]rune(errorStr)) - 1,
					},
				}
				ids = []ID{0}
				maxLen = displayWidth(errorStr)
				return
			}

			// find
			maxLine := 0
			initIndexSet := false
			for i := 0; i < view.Moment.NumLines(); i++ {
				line := view.Moment.GetLine(i)
				loc := pattern.FindStringIndex(line.content)
				if len(loc) == 0 {
					continue
				}
				ids = append(ids, ID(len(results)))
				result := Result{
					LineNumber: i + 1,
					Content:    line.content,
					BeginRuneOffset: func() int {
						n := 0
						for i, cell := range line.Cells {
							if n == loc[0] {
								return i
							}
							n += cell.RuneLen
						}
						panic("impossible")
					}(),
					EndRuneOffset: func() int {
						n := 0
						for i, cell := range line.Cells {
							if n == loc[1] {
								return i
							}
							n += cell.RuneLen
						}
						panic("impossible")
					}(),
				}
				results = append(results, result)
				if line.DisplayWidth > maxLen {
					maxLen = line.DisplayWidth
				}
				if i+1 > maxLine {
					maxLine = i + 1
				}
				if !initIndexSet && i >= view.CursorLine {
					initIndex = len(results) - 1
					initIndexSet = true
				}
			}

			lineNumDisplayLen := len(fmt.Sprintf("%d", maxLine))
			maxL := 0
			for i, result := range results {
				result.Content = fmt.Sprintf(
					"%"+strconv.Itoa(lineNumDisplayLen)+"d",
					result.LineNumber,
				) + ":" + fmt.Sprintf(
					"%-"+strconv.Itoa(maxLen)+"s",
					result.Content,
				)
				result.BeginRuneOffset += lineNumDisplayLen + 1
				result.EndRuneOffset += lineNumDisplayLen + 1
				results[i] = result
				if w := displayWidth(result.Content); w > maxL {
					maxL = w
				}
			}
			maxLen = maxL

			return
		},

		CandidateElement: func(scope Scope, id ID) Element {
			var box Box
			var style Style
			var getStyle GetStyle
			var focus ID
			scope.Assign(&box, &style, &getStyle, &focus)
			if id == focus {
				style = style.Underline(true)
			}
			result := results[id]
			hlStyle := getStyle("Highlight")
			fg, _, _ := hlStyle.Decompose()
			hlStyle = style.Foreground(fg).Bold(true)
			return Text(
				box,
				result.Content,
				OffsetStyleFunc(func(i int) Style {
					if i >= result.BeginRuneOffset && i < result.EndRuneOffset {
						return hlStyle
					}
					return style
				}),
			)
		},
	}

	scope.Sub(func() OverlayObject { return dialog }).Call(PushOverlay, &id)
}

func (_ Command) ShowSearchDialog() (spec CommandSpec) {
	spec.Desc = "show search dialog"
	spec.Func = ShowSearchDialog
	return
}
