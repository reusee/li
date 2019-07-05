package li

import (
	"fmt"
	"path"
	"reflect"
	"strconv"
	"strings"
)

func Status(
	scope Scope,
	box Box,
	cur CurrentView,
	getN GetN,
	getKeyEv GetLastKeyEvent,
	recording MacroRecording,
	getMacroName GetMacroName,
	getModes CurrentModes,
	getStyle GetStyle,
	style Style,
	curGroup CurrentViewGroup,
	groupLayoutIndex ViewGroupLayoutIndex,
	viewGroupConfig ViewGroupConfig,
	groups ViewGroups,
) (
	ret Element,
) {

	focusing := cur()
	style = darkerOrLighterStyle(style, 15)
	hlStyle := getStyle("Highlight")
	fg, _, _ := hlStyle.Decompose()
	hlStyle = style.Foreground(fg)

	lineBox := Box{
		Top:    box.Top,
		Left:   box.Left,
		Right:  box.Right,
		Bottom: box.Top + 1,
	}
	var subs []Element

	addTextLine := func(specs ...any) {
		specs = append(specs, lineBox)
		subs = append(subs, Text(specs...))
		lineBox.Top++
		lineBox.Bottom++
	}

	//addTextLine("Li Editor", AlignCenter, Bold(true))

	// modes
	modes := getModes()
	addTextLine("")
	addTextLine("modes", Bold(true), AlignRight, Padding(0, 2, 0, 0))
	for _, mode := range modes {
		name := reflect.TypeOf(mode).Elem().Name()
		name = strings.TrimSuffix(name, "Mode")
		s := style
		if name == "Edit" {
			s = hlStyle
		}
		addTextLine(s, name, AlignRight, Padding(0, 2, 0, 0))
	}

	// context number
	if n := getN(); n > 0 {
		addTextLine("")
		addTextLine("context", Bold(true), AlignRight, Padding(0, 2, 0, 0))
		addTextLine(strconv.Itoa(n), AlignRight, Padding(0, 2, 0, 0))
	}

	// macro
	if recording {
		addTextLine("")
		addTextLine("macro", Bold(true), AlignRight, Padding(0, 2, 0, 0))
		addTextLine(getMacroName(), AlignRight, Padding(0, 2, 0, 0))
	}

	// line and col
	if focusing != nil {
		line := focusing.CursorLine + 1
		col := focusing.CursorCol + 1
		addTextLine("")
		addTextLine("position", Bold(true), AlignRight, Padding(0, 2, 0, 0))
		addTextLine(strconv.Itoa(line), AlignRight, Padding(0, 2, 0, 0))
		addTextLine(strconv.Itoa(col), AlignRight, Padding(0, 2, 0, 0))
	}

	// last key
	if ev := getKeyEv(); ev != nil {
		addTextLine("")
		addTextLine("key", Bold(true), AlignRight, Padding(0, 2, 0, 0))
		addTextLine(ev.Name(), AlignRight, Padding(0, 2, 0, 0))
	}

	group := curGroup()
	groupIndex := func() int {
		for i, g := range groups {
			if g == group {
				return i
			}
		}
		return 0
	}()

	// layout
	addTextLine("")
	addTextLine("layout", Bold(true), AlignRight, Padding(0, 2, 0, 0))
	addTextLine(viewGroupConfig.Layouts[groupLayoutIndex()], AlignRight, Padding(0, 2, 0, 0))
	addTextLine(group.Layouts[group.LayoutIndex], AlignRight, Padding(0, 2, 0, 0))

	// views
	views := group.GetViews(scope)
	if len(views) > 0 {
		addTextLine("")
		addTextLine(
			fmt.Sprintf("group %d / %d", groupIndex+1, len(groups)),
			Bold(true), AlignRight, Padding(0, 2, 0, 0))
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
						if view.Buffer.LastSyncFileInfo == view.Moment.FileInfo {
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
