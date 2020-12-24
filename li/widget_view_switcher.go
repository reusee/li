package li

import (
	"sort"

	"github.com/junegunn/fzf/src/util"
)

func ShowViewSwitcher(
	scope Scope,
	pushOverlay PushOverlay,
	closeOverlay CloseOverlay,
) {

	// states
	type Candidate struct {
		Name     string
		MatchLen int
		Score    int
		View     *View
	}
	var candidates []Candidate
	var maxLength int
	updateCandidates := func(scope Scope, runes []rune) {

		// reset state
		candidates = candidates[:0]
		maxLength = 0

		// collect
		var views Views
		scope.Assign(&views)
		for _, view := range views {
			name := view.Buffer.Path
			if w := displayWidth(name); w > maxLength {
				maxLength = w
			}
			chars := util.RunesToChars([]rune(name))
			matched, matchLen, score := fuzzyMatched(runes, &chars)
			if !matched {
				continue
			}
			candidates = append(candidates, Candidate{
				Name:     name,
				MatchLen: matchLen,
				Score:    score,
				View:     view,
			})
		}

		sort.SliceStable(candidates, func(i, j int) bool {
			score1 := candidates[i].Score
			score2 := candidates[j].Score
			if score1 != score2 {
				return score1 > score2
			}
			return candidates[i].Name < candidates[j].Name
		})

	}
	updateCandidates(scope, nil)

	var id ID
	dialog := &SelectionDialog{

		Title: "Switch View",

		OnClose: func(_ Scope) {
			closeOverlay(id)
		},

		OnSelect: func(scope Scope, id ID) {
			var cur CurrentView
			scope.Assign(&cur)
			closeOverlay(id)
			if int(id) < len(candidates) {
				view := candidates[id].View
				cur(view)
			}
		},

		OnUpdate: func(scope Scope, runes []rune) (ids []ID, maxLen int, initIndex int) {
			updateCandidates(scope, runes)
			maxLen = maxLength
			for i := range candidates {
				ids = append(ids, ID(i))
			}
			return
		},

		CandidateElement: func(scope Scope, id ID) Element {
			var box Box
			var focus ID
			var style Style
			var getStyle GetStyle
			scope.Assign(&box, &focus, &style, &getStyle)
			s := style
			if id == focus {
				hlStyle := getStyle("Highlight")(s)
				fg, _, _ := hlStyle.Decompose()
				s = s.Foreground(fg)
			}
			candidate := candidates[id]
			return Text(
				box,
				candidate.Name,
				s,
				OffsetStyleFunc(func(i int) StyleFunc {
					fn := SameStyle
					if i < candidate.MatchLen {
						fn = fn.SetUnderline(true)
					} else {
						fn = fn.SetUnderline(false)
					}
					return fn
				}),
			)
		},
	}

	overlay := OverlayObject(dialog)
	id = pushOverlay(overlay)
}

func (_ Command) ShowViewSwitcher() (spec CommandSpec) {
	spec.Desc = "show view switcher"
	spec.Func = func(scope Scope) {
		scope.Call(ShowViewSwitcher)
	}
	return
}
