package li

type (
	ViewGroupLayoutIndex func(i ...int) int
)

func (_ Provide) ViewGroupLayoutIndex() (
	fn ViewGroupLayoutIndex,
) {
	var index int
	fn = func(is ...int) int {
		for _, i := range is {
			index = i
		}
		return index
	}
	return
}

func NextViewGroupLayout(
	idx ViewGroupLayoutIndex,
	config ViewGroupConfig,
) {
	cur := idx()
	cur++
	if cur >= len(config.Layouts) {
		cur = 0
	}
	idx(cur)
}

func (_ Command) NextViewGroupLayout() (spec CommandSpec) {
	spec.Desc = "switch to next view group layout"
	spec.Func = NextViewGroupLayout
	return
}

func PrevViewGroupLayout(
	idx ViewGroupLayoutIndex,
	config ViewGroupConfig,
) {
	cur := idx()
	cur--
	if cur < 0 {
		cur = len(config.Layouts) - 1
	}
	idx(cur)
}

func (_ Command) PrevViewGroupLayout() (spec CommandSpec) {
	spec.Desc = "switch to previous view group layout"
	spec.Func = PrevViewGroupLayout
	return
}
