package li

import "path/filepath"

type (
	CurrentView   func(...*View) *View
	CurrentMoment func() *Moment
)

type evCurrentViewChanged struct{}

var EvCurrentViewChanged = new(evCurrentViewChanged)

func (_ Provide) CurrentView(
	link Link,
	linkedOne LinkedOne,
	j AppendJournal,
	scope Scope,
	trigger Trigger,
) (
	fn CurrentView,
) {
	type Flag struct{}
	var flag Flag
	fn = func(views ...*View) (ret *View) {
		for _, view := range views {
			link(flag, view)
			path, err := filepath.Abs(view.Buffer.Path)
			ce(err)
			j("switch to %s", path)
		}
		linkedOne(flag, &ret)
		if len(views) > 0 {
			trigger(scope.Sub(
				&ret,
			), EvCurrentViewChanged)
		}
		return
	}
	return
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
