package li

type ShowMessage func(
	lines []string,
)

func (_ Provide) ShowMessage(
	scope Scope,
	pushOverlay PushOverlay,
) ShowMessage {
	return func(
		lines []string,
	) {

		length := 0
		for _, line := range lines {
			w := 0
			for _, r := range line {
				w += runeDisplayWidth(r)
			}
			if w > length {
				length = w
			}
		}
		paddingHorizontal := 5
		paddingVertical := 2

		var id ID
		msgBox := WidgetDialog{

			OnKey: func(
				ev KeyEvent,
				scope Scope,
				closeOverlay CloseOverlay,
			) {
				switch ev.Name() {
				case "Enter", "Esc":
					// close
					closeOverlay(id)
				}
			},

			Element: ElementFrom(func(
				box Box,
				defaultStyle Style,
			) Element {

				left := (box.Left+box.Right)/2 - paddingHorizontal - length/2
				top := (box.Top+box.Bottom)/2 - paddingVertical - len(lines)/2
				return Rect(
					darkerOrLighterStyle(defaultStyle, -10),
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
		id = pushOverlay(overlay)
	}
}
