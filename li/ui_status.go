package li

import (
	"sort"
)

type StatusSection struct {
	Title string
	Lines [][]any
}

type EvCollectStatusSections struct {
	Add    AddStatusSection
	Styles []Style
}

type AddStatusSection func(string, [][]any)

func Status(
	scope Scope,
	box Box,
	cur CurrentView,
	getStyle GetStyle,
	style Style,
	curGroup CurrentViewGroup,
	groups ViewGroups,
	trigger Trigger,
) (
	ret Element,
) {

	style = darkerOrLighterStyle(style, 15)
	hlStyle := getStyle("Highlight")(style)
	fg, _, _ := hlStyle.Decompose()
	hlStyle = style.Foreground(fg)

	// collect sections
	var sections []StatusSection
	add := AddStatusSection(func(title string, lines [][]any) {
		sections = append(sections, StatusSection{title, lines})
	})
	trigger(EvCollectStatusSections{
		Add:    add,
		Styles: []Style{style, hlStyle},
	})
	sort.Slice(sections, func(i, j int) bool {
		return sections[i].Title < sections[j].Title
	})

	// render
	lineBox := Box{
		Top:    box.Top,
		Left:   box.Left,
		Right:  box.Right,
		Bottom: box.Top + 1,
	}
	var subs []Element

	// sections
	for _, section := range sections {
		// title
		subs = append(subs, Text(section.Title, Bold(true), AlignRight, Padding(0, 2, 0, 0), lineBox))
		lineBox.Top++
		lineBox.Bottom++
		// lines
		for _, line := range section.Lines {
			subs = append(subs, Text(append(line, lineBox)...))
			lineBox.Top++
			lineBox.Bottom++
		}
		subs = append(subs, Text(lineBox, ""))
		lineBox.Top++
		lineBox.Bottom++
	}

	return Rect(
		style,
		box,
		Fill(true),
		subs,
	)

}
