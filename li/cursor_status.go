package li

import "strconv"

func (_ Provide) CursorStatus(
	on On,
) Init2 {

	on(EvRenderStatus, func(
		cur CurrentView,
		add AddStatusLine,
		scope Scope,
	) {
		focusing := cur()
		if focusing == nil {
			return
		}
		line := focusing.CursorLine + 1
		col := focusing.CursorCol + 1
		add("")
		add("cursor", Bold(true), AlignRight, Padding(0, 2, 0, 0))
		add(strconv.Itoa(line), AlignRight, Padding(0, 2, 0, 0))
		add(strconv.Itoa(col), AlignRight, Padding(0, 2, 0, 0))
		moment := focusing.GetMoment()
		if parser := moment.GetParser(scope); parser != nil {
			pos := focusing.cursorPosition()
			add(
				moment.GetSyntaxAttr(scope, pos.Line, pos.Cell),
				AlignRight, Padding(0, 2, 0, 0),
			)
		}
	})

	return nil
}
