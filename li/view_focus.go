package li

import "path/filepath"

type CurrentMoment func() *Moment

type evCurrentViewChanged struct{}

var EvCurrentViewChanged = new(evCurrentViewChanged)

type CurrentView func(...*View) *View

func (_ Provide) CurrentView(
	link Link,
	linkedOne LinkedOne,
	j AppendJournal,
	trigger Trigger,
	scope Scope,
) CurrentView {
	type Anchor struct{}
	var anchor Anchor
	return func(vs ...*View) (ret *View) {
		for _, v := range vs {
			link(anchor, v)
			path, err := filepath.Abs(v.Buffer.Path)
			ce(err)
			j("switch to %s", path)
		}
		linkedOne(anchor, &ret)
		if len(vs) > 0 {
			trigger(scope.Sub(
				&ret,
			), EvCurrentViewChanged)
		}
		return
	}
}

func AsCurrentView(view *View) (
	_ func() CurrentView,
) {
	return func() CurrentView {
		return func(vs ...*View) *View {
			if len(vs) > 0 {
				panic("not updatable")
			}
			return view
		}
	}
}

func (_ Provide) CurrentMoment(
	v CurrentView,
) CurrentMoment {
	return func() *Moment {
		view := v()
		if view == nil {
			return nil
		}
		return view.GetMoment()
	}
}

func AsCurrentMoment(moment *Moment) (
	_ func() CurrentMoment,
) {
	return func() CurrentMoment {
		return func() *Moment {
			return moment
		}
	}
}

func FocusNextViewInGroup(
	curGroup CurrentViewGroup,
	cur CurrentView,
	linkedAll LinkedAll,
) {
	group := curGroup()
	if group == nil {
		return
	}
	var views []*View
	linkedAll(group, &views)
	if len(views) == 0 {
		return
	}
	index := -1
	v := cur()
	for i, view := range views {
		if v == view {
			index = i
			break
		}
	}
	if index < 0 {
		index = 0
	} else {
		index++
		if index == len(views) {
			index = 0
		}
	}
	cur(views[index])
}

func (_ Command) FocusNextViewInGroup() (spec CommandSpec) {
	spec.Desc = "focus next view in the same group"
	spec.Func = FocusNextViewInGroup
	return
}

func FocusPrevViewInGroup(
	curGroup CurrentViewGroup,
	cur CurrentView,
	linkedAll LinkedAll,
) {
	group := curGroup()
	if group == nil {
		return
	}
	var views []*View
	linkedAll(group, &views)
	if len(views) == 0 {
		return
	}
	index := -1
	v := cur()
	for i, view := range views {
		if v == view {
			index = i
			break
		}
	}
	index--
	if index < 0 {
		index = len(views) - 1
	}
	cur(views[index])
}

func (_ Command) FocusPrevViewInGroup() (spec CommandSpec) {
	spec.Desc = "focus previous view in the same group"
	spec.Func = FocusPrevViewInGroup
	return
}

func (_ Provide) CurrentViewFilePathStatus(
	on On,
) OnStartup {
	return func() {
		on(EvCollectStatusSections, func(
			add AddStatusSection,
			cur CurrentView,
		) {
			view := cur()
			if view == nil {
				return
			}
			parts := splitDir(view.Buffer.AbsPath)
			if len(parts) == 0 {
				return
			}
			var lines [][]any
			for _, part := range parts {
				lines = append(lines, []any{
					part,
					AlignRight,
					Padding(0, 2, 0, 0),
				})
			}
			add("path", lines)
		})
	}
}
