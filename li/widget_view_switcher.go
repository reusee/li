package li

import (
	"sort"

	"github.com/junegunn/fzf/src/util"
)

func ShowViewSwitcher(scope Scope) {

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

		sort.Slice(candidates, func(i, j int) bool {
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
			scope.Sub(func() ID { return id }).Call(CloseOverlay)
		},

		OnSelect: func(scope Scope, id ID) {
			var cur CurrentView
			scope.Assign(&cur)
			scope.Sub(func() ID { return id }).Call(CloseOverlay)
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
				hlStyle := getStyle("Highlight")
				fg, _, _ := hlStyle.Decompose()
				s = s.Foreground(fg)
			}
			candidate := candidates[id]
			return Text(
				box,
				candidate.Name,
				s,
				OffsetStyleFunc(func(i int) Style {
					runeStyle := s
					if i < candidate.MatchLen {
						runeStyle = runeStyle.Underline(true)
					} else {
						runeStyle = runeStyle.Underline(false)
					}
					return runeStyle
				}),
			)
		},
	}

	scope.Sub(func() OverlayObject { return dialog }).Call(PushOverlay, &id)
}

func (_ Command) ShowViewSwitcher() (spec CommandSpec) {
	spec.Desc = "show view switcher"
	spec.Func = func(scope Scope) (NoResetN, NoLogImitation) {
		scope.Call(ShowViewSwitcher)
		return true, true
	}
	return
}
