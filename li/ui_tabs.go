package li

import (
	"path"
)

func Tabs(
	cur CurrentView,
	getStyle GetStyle,
	views Views,
) (ret Element) {

	style := darkerOrLighterStyle(getStyle("Default"), 15)
	hlStyle := getStyle("Highlight")
	fg, _, _ := hlStyle.Decompose()
	hlStyle = style.Foreground(fg)

	current := cur()
	var viewNames []string
	curIndex := -1
	for _, view := range views {
		if view == current {
			curIndex = len(viewNames)
		}
		viewNames = append(viewNames, path.Base(view.Buffer.Path))
	}

	ret = Rect(
		Fill(true),
		style,

		func(
			box Box,
		) (ret []Element) {

			if len(viewNames) == 0 {
				return
			}

			sizes := split(box.Width(), len(viewNames))

			left := box.Left
			for i, name := range viewNames {
				var s Style
				if i == curIndex {
					s = hlStyle
				} else {
					s = style
				}
				size := sizes[i]

				ret = append(ret, Text(
					s,
					name,
					Box{
						Left:   left,
						Right:  left + size,
						Top:    box.Top,
						Bottom: box.Bottom,
					},
				))

				left += size
			}

			return
		},
	)

	return
}
