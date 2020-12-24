package li

import "strconv"

func (_ Provide) CursorStatus(
	on On,
) Init2 {

	on(EvCollectStatusSections, func(
		cur CurrentView,
		add AddStatusSection,
		scope Scope,
	) {
		focusing := cur()
		if focusing == nil {
			return
		}
		line := focusing.CursorLine + 1
		col := focusing.CursorCol + 1
		lines := [][]any{
			{strconv.Itoa(line), AlignRight, Padding(0, 2, 0, 0)},
			{strconv.Itoa(col), AlignRight, Padding(0, 2, 0, 0)},
		}
		moment := focusing.GetMoment()
		if parser := moment.GetParser(scope); parser != nil {
			pos := focusing.cursorPosition()
			lines = append(lines, []any{
				moment.GetSyntaxAttr(scope, pos.Line, pos.Cell),
				AlignRight, Padding(0, 2, 0, 0),
			})
		}
		add("cursor", lines)
	})

	return nil
}
