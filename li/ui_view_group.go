package li

func (g *ViewGroup) RenderFunc() any {
	return func(
		box Box,
		scope Scope,
		getCur CurrentView,
		screen Screen,
		linkedAll LinkedAll,
	) Element {

		var views []*View
		cur := getCur()
		max, split := NamedLayouts[g.Layouts[g.LayoutIndex]](box)
		if max == 1 {
			views = append(views, cur)
		} else {
			views = g.GetViews(scope)
		}

		// views
		if len(views) == 0 {
			return nil
		}

		// layout
		viewBoxes := split(len(views))
		for i, view := range views {
			newBox := viewBoxes[i]
			if view.Box != newBox.Box {
				view.Box = newBox.Box
			}
		}

		// elements
		var elements []Element
		for _, view := range views {
			elements = append(elements, view)
		}

		return Rect(
			elements,
		)
	}
}
