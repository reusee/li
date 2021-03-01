package li

import (
	"fmt"
	"path"
	"sync/atomic"
	"time"
)

func ViewArea(
	viewGroupConfig ViewGroupConfig,
	groupLayoutIndex ViewGroupLayoutIndex,
	box Box,
	getStyle GetStyle,
	defaultStyle Style,
	views Views,
	viewGroups ViewGroups,
	linkedAll LinkedAll,
	screen Screen,
	curGroup CurrentViewGroup,
	shouldShowList ShouldShowViewList,
	j AppendJournal,
	cur CurrentView,
	scope Scope,
	config UIConfig,
) Element {

	// cursor
	screen.ShowCursor(box.Left, box.Top)

	var subs []Element

	// view groups
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
	style := defaultStyle
	for i, group := range groups {
		s := darkerOrLighterStyle(style, int32(i)*2)
		subs = append(subs, ElementWith(
			group,
			func() (Box, Style) {
				return groupBoxes[i].Box, s
			},
		))
	}

	// view list
	if shouldShowList() {
		group := curGroup()
		groupIndex := func() int {
			for i, g := range groups {
				if g == group {
					return i
				}
			}
			return 0
		}()
		focusing := cur()
		views := group.GetViews(scope)
		if len(views) > 0 {

			left := box.Right - config.ViewList.Width
			if left > config.ViewList.MarginLeft {
				left = config.ViewList.MarginLeft
			}
			frameBox := Box{
				Top:    box.Top,
				Bottom: box.Bottom,
				Left:   left,
				Right:  left + config.ViewList.Width,
			}
			contentBox := Box{
				Top:    frameBox.Top,
				Bottom: frameBox.Bottom,
				Left:   frameBox.Left + 1,
				Right:  frameBox.Right - 1,
			}
			style = darkerOrLighterStyle(style, 15)
			hlStyle := getStyle("Highlight")(style)
			fg, _, _ := hlStyle.Decompose()
			hlStyle = style.Foreground(fg)

			focusLine := func() int {
				for i, view := range views {
					if view == focusing {
						return i
					}
				}
				return 0
			}()

			listElem := ElementWith(
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
							s = s.Underline(false)
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
								Fill(true),
							))
						}
						return
					}),
					focusLine,
				),
				func() Box {
					return Box{
						contentBox.Top + 1,
						contentBox.Left,
						contentBox.Bottom,
						contentBox.Right,
					}
				},
			)

			subs = append(subs, Rect(
				frameBox,
				style,
				Fill(true),
				Text(
					Box{contentBox.Top, contentBox.Left, contentBox.Top + 1, contentBox.Right},
					fmt.Sprintf("group %d / %d", groupIndex+1, len(groups)),
					Bold(true), AlignLeft, Fill(true), style,
				),
				listElem,
			))
		}
	}

	return Rect(
		Fill(true),
		subs,
	)
}

type ShouldShowViewList func() bool

type ShouldShowViewListState *int64

func (_ Provide) ShouldShowViewListState() (
	s ShouldShowViewListState,
	get ShouldShowViewList,
) {
	var i int64
	s = &i
	get = func() bool {
		return atomic.LoadInt64(&i) > 0
	}
	return
}

func (_ Provide) ShowViewListFlag(
	on On,
	config UIConfig,
	run RunInMainLoop,
	p ShouldShowViewListState,
) OnStartup {
	return func() {

		on(EvCurrentViewChanged, func() {
			atomic.AddInt64(p, 1)
			time.AfterFunc(time.Second*time.Duration(config.ViewList.HideTimeoutSeconds), func() {
				run(func() {
					atomic.AddInt64(p, -1)
				})
			})
		})

	}
}
