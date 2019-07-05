package li

func ViewArea(
	viewGroupConfig ViewGroupConfig,
	groupLayoutIndex ViewGroupLayoutIndex,
	box Box,
	getStyle GetStyle,
	views Views,
	viewGroups ViewGroups,
	linkedAll LinkedAll,
	screen Screen,
	curGroup CurrentViewGroup,
) Element {

	screen.ShowCursor(box.Left, box.Top)

	var groups []*ViewGroup

	max, split := NamedLayouts[viewGroupConfig.Layouts[groupLayoutIndex()]](box)
	if max == 1 {
		cur := curGroup()
		if cur != nil {
			groups = []*ViewGroup{cur}
		}
	} else {
		for _, group := range viewGroups {
			var views []*View
			linkedAll(group, &views)
			if len(views) == 0 {
				continue
			}
			groups = append(groups, group)
		}
	}

	if len(groups) == 0 {
		return Rect(Fill(true))
	}

	groupBoxes := split(len(groups))
	var groupElements []Element
	style := getStyle("Default")
	for i, group := range groups {
		s := darkerOrLighterStyle(style, int32(i)*2)
		groupElements = append(groupElements, ElementWith(
			group,
			func() (Box, Style) {
				return groupBoxes[i].Box, s
			},
		))
	}

	return Rect(
		Fill(true),
		groupElements,
	)
}
