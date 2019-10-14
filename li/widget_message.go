package li

func ShowMessage(
	scope Scope,
	lines []string,
) {

	length := 0
	for _, line := range lines {
		runes := []rune(line)
		w := 0
		for _, r := range runes {
			w += runeWidth(r)
		}
		if w > length {
			length = w
		}
	}
	paddingHorizontal := 5
	paddingVertical := 2

	var id ID
	msgBox := WidgetDialog{

		OnKey: func(ev KeyEvent, scope Scope) {
			switch ev.Name() {
			case "Enter", "Esc":
				// close
				scope.Sub(&id).Call(CloseOverlay)
			}
		},

		Element: ElementFrom(func(
			box Box,
			getStyle GetStyle,
		) Element {

			left := (box.Left+box.Right)/2 - paddingHorizontal - length/2
			top := (box.Top+box.Bottom)/2 - paddingVertical - len(lines)/2
			return Rect(
				darkerOrLighterStyle(getStyle("Default"), -10),
				Fill(true),
				Box{
					Left:   left,
					Right:  left + paddingHorizontal*2 + length,
					Top:    top,
					Bottom: top + paddingVertical*2 + len(lines),
				},
				Padding(paddingVertical, paddingHorizontal),
				Text(
					lines,
				),
			)

		}),
	}

	overlay := OverlayObject(msgBox)
	scope.Sub(&overlay).Call(PushOverlay, &id)
}
