package li

type (
	CurrentView func(...*View) *View
)

func (_ Provide) CurrentView(
	link Link,
	linkedOne LinkedOne,
) (
	fn CurrentView,
) {
	type Flag struct{}
	var flag Flag
	fn = func(views ...*View) (ret *View) {
		for _, view := range views {
			link(flag, view)
		}
		linkedOne(flag, &ret)
		return
	}
	return
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