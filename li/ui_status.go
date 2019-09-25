package li

import (
	"path"
	"sort"
)

type StatusSection struct {
	Title string
	Lines [][]any
}

type evCollectStatusSections struct{}

var EvCollectStatusSections = new(evCollectStatusSections)

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
	hlStyle := getStyle("Highlight")
	fg, _, _ := hlStyle.Decompose()
	hlStyle = style.Foreground(fg)

	// collect sections
	var sections []StatusSection
	trigger(scope.Sub(
		func() AddStatusSection {
			return func(title string, lines [][]any) {
				sections = append(sections, StatusSection{title, lines})
			}
		},
		func() []Style {
			return []Style{style, hlStyle}
		},
	), EvCollectStatusSections)
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
	for i, section := range sections {
		if i > 0 {
			subs = append(subs, Text(lineBox, ""))
			lineBox.Top++
			lineBox.Bottom++
		}
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
	}

	// views
	group := curGroup()
	//groupIndex := func() int {
	//	for i, g := range groups {
	//		if g == group {
	//			return i
	//		}
	//	}
	//	return 0
	//}()
	focusing := cur()
	views := group.GetViews(scope)
	if len(views) > 0 {
		//TODO
		//addTextLine("")
		//addTextLine(
		//	fmt.Sprintf("group %d / %d", groupIndex+1, len(groups)),
		//	Bold(true), AlignRight, Padding(0, 2, 0, 0),
		//)
		box := Box{
			Top:    lineBox.Top,
			Left:   box.Left,
			Right:  box.Right,
			Bottom: box.Bottom,
		}
		focusLine := func() int {
			for i, view := range views {
				if view == focusing {
					return i
				}
			}
			return 0
		}()
		subs = append(subs, ElementWith(
			VerticalScroll(
				ElementFrom(func(
					box Box,
				) (ret []Element) {
					for i, view := range views {
						name := path.Base(view.Buffer.Path)
						s := style
						if view == focusing {
							s = hlStyle
						}
						if view.Buffer.LastSyncFileInfo == view.GetMoment().FileInfo {
							s = s.Underline(false)
						} else {
							s = s.Underline(true)
						}
						ret = append(ret, Text(
							Box{
								Top:    box.Top + i,
								Left:   box.Left,
								Right:  box.Right,
								Bottom: box.Top + i + 1,
							},
							name,
							s,
							AlignRight,
							Padding(0, 2, 0, 0),
						))
					}
					return
				}),
				focusLine,
			),
			func() Box {
				return box
			},
		))
	}

	return Rect(
		style,
		box,
		Fill(true),
		subs,
	)

}
