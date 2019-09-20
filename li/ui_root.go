package li

func Root(
	scope Scope,
	width Width,
	height Height,
	getStyle GetStyle,
	getJournalHeight JournalHeight,
) Element {

	box := Box{0, 0, int(height), int(width)}
	scope = scope.Sub(
		func() Box {
			return box
		},
		func() Style {
			return getStyle("Default")
		},
	)

	journalHeight := getJournalHeight()
	journalBox := Box{box.Bottom - 1, 0, box.Bottom, box.Width()}

	statusWidth := 15
	if statusWidth > box.Width() {
		statusWidth = box.Width() / 10
	}
	viewBox := Box{0, statusWidth, box.Height() - journalHeight, box.Width()}

	return ElementFrom(

		// status
		ElementWith(
			ElementFrom(Status),
			func() Box {
				return Box{0, 0, box.Height() - journalHeight, statusWidth}
			},
		),

		// tabs
		//ElementWith(
		//	ElementFrom(Tabs),
		//	func() Box {
		//		return Box{0, statusWidth, 1, box.Width()}
		//	},
		//),

		// view area
		ElementWith(
			ElementFrom(ViewArea),
			func() Box {
				return viewBox
			},
		),

		// overlay
		ElementWith(
			ElementFrom(OverlayUI),
			func() Box {
				return viewBox
			},
		),

		// journal
		ElementWith(
			ElementFrom(JournalUI),
			func() Box {
				return journalBox
			},
		),

		//
	)

}
