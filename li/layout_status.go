package li

func (_ Provide) LayoutStatus(
	on On,
) OnStartup {
	return func() {

		on(func(
			ev EvCollectStatusSections,
			viewGroupConfig ViewGroupConfig,
			groupLayoutIndex ViewGroupLayoutIndex,
			curGroup CurrentViewGroup,
		) {
			group := curGroup()
			ev.Add("layout", [][]any{
				{viewGroupConfig.Layouts[groupLayoutIndex()], AlignRight, Padding(0, 2, 0, 0)},
				{group.Layouts[group.LayoutIndex], AlignRight, Padding(0, 2, 0, 0)},
			})
		})

	}
}
