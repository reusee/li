package li

func (_ Provide) LayoutStatus(
	on On,
) Init2 {

	on(EvRenderStatus, func(
		add AddStatusLine,
		viewGroupConfig ViewGroupConfig,
		groupLayoutIndex ViewGroupLayoutIndex,
		curGroup CurrentViewGroup,
	) {
		group := curGroup()
		add("")
		add("layout", Bold(true), AlignRight, Padding(0, 2, 0, 0))
		add(viewGroupConfig.Layouts[groupLayoutIndex()], AlignRight, Padding(0, 2, 0, 0))
		add(group.Layouts[group.LayoutIndex], AlignRight, Padding(0, 2, 0, 0))
	})

	return nil
}
