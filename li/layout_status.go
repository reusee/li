package li

func (_ Provide) LayoutStatus(
	on On,
) OnStartup {
	return func() {

		on(EvCollectStatusSections, func(
			add AddStatusSection,
			viewGroupConfig ViewGroupConfig,
			groupLayoutIndex ViewGroupLayoutIndex,
			curGroup CurrentViewGroup,
		) {
			group := curGroup()
			add("layout", [][]any{
				{viewGroupConfig.Layouts[groupLayoutIndex()], AlignRight, Padding(0, 2, 0, 0)},
				{group.Layouts[group.LayoutIndex], AlignRight, Padding(0, 2, 0, 0)},
			})
		})

	}
}
