package li

import (
	"strconv"
	"strings"
	"testing"
)

func TestWidgetSelectionDialog(t *testing.T) {
	withEditor(func(scope Scope) {

		var id ID
		dialog := &SelectionDialog{

			Title: "foo",

			OnClose: func(_ Scope) {
				scope.Sub(func() ID { return id }).Call(CloseOverlay)
			},

			OnSelect: func(_ Scope, id ID) {
				log("%d\n", id)
			},

			OnUpdate: func(_ Scope, runes []rune) (
				ids []ID,
				maxLen int,
				initIndex int,
			) {
				n, err := strconv.Atoi(string(runes))
				if err != nil {
					return
				}
				if n > 200 {
					n = 200
				}
				for i := 0; i < n; i++ {
					ids = append(ids, ID(i))
				}
				maxLen = (n - 1) * 3
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
				return Text(
					box,
					strings.Repeat("foo", int(id+1)),
					s,
				)
			},
		}
		scope.Sub(func() OverlayObject {
			return dialog
		}).Call(PushOverlay, &id)

	})
}
