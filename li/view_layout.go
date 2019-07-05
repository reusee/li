package li

func NextViewLayout(
	curGroup CurrentViewGroup,
) {
	group := curGroup()
	group.LayoutIndex++
	if group.LayoutIndex >= len(group.Layouts) {
		group.LayoutIndex = 0
	}
}

func (_ Command) NextViewLayout() (spec CommandSpec) {
	spec.Desc = "switch to next view layout of current view group"
	spec.Func = NextViewLayout
	return
}

func PrevViewLayout(
	curGroup CurrentViewGroup,
) {
	group := curGroup()
	group.LayoutIndex--
	if group.LayoutIndex < 0 {
		group.LayoutIndex = len(group.Layouts) - 1
	}
}

func (_ Command) PrevViewLayout() (spec CommandSpec) {
	spec.Desc = "switch to previous view layout of current view group"
	spec.Func = PrevViewLayout
	return
}
