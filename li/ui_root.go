package li

func Root(
	scope Scope,
	width Width,
	height Height,
	getStyle GetStyle,
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

	statusWidth := 15
	if statusWidth > box.Width() {
		statusWidth = box.Width() / 10
	}
	viewBox := Box{0, statusWidth, box.Height(), box.Width()}

	return ElementFrom(

		// status
		ElementWith(
			ElementFrom(Status),
			func() Box {
				return Box{0, 0, box.Height(), statusWidth}
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

		//
	)

}
