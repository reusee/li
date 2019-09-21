package li

import (
	"sort"
	"sync/atomic"
)

type ViewGroupID int64

type ViewGroup struct {
	ID ViewGroupID
	ViewGroupSpec
	LayoutIndex int
}

type ViewGroupSpec struct {
	SortKeys []string
	Layouts  []string
}

type ViewGroups []*ViewGroup

func (_ Provide) DefaultViewGroups(
	config ViewGroupConfig,
	current CurrentViewGroup,
) ViewGroups {
	var groups []*ViewGroup
	for _, groupSpec := range config.Groups {
		groups = append(groups, &ViewGroup{
			ID:            ViewGroupID(atomic.AddInt64(&nextViewGroupID, 1)),
			ViewGroupSpec: groupSpec,
		})
	}
	current(groups...)
	return groups
}

var nextViewGroupID int64

type ViewGroupConfig struct {
	Groups      []ViewGroupSpec
	Layouts     []string
	LayoutIndex int
}

func (_ Provide) ViewGroupConfig(
	getConfig GetConfig,
	idx ViewGroupLayoutIndex,
) (
	config ViewGroupConfig,
) {

	var data struct {
		ViewGroup ViewGroupConfig
	}
	ce(getConfig(&data))
	config = data.ViewGroup

	if len(config.Layouts) == 0 {
		config.Layouts = []string{
			"Stacked",
			"BinarySplit",
			"HorizontalSplit",
			"VerticalSplit",
		}
	}

	if len(config.Groups) == 0 {
		config.Groups = append(config.Groups, ViewGroupSpec{
			SortKeys: []string{"ID"},
			Layouts: []string{
				"Stacked",
				"BinarySplit",
				"HorizontalSplit",
				"VerticalSplit",
			},
		})
	}

	idx(config.LayoutIndex)

	return
}

type (
	CurrentViewGroup func(...*ViewGroup) *ViewGroup
)

func (_ Provide) ViewGroupAccessor(
	link Link,
	linkedOne LinkedOne,
) (
	fn CurrentViewGroup,
) {

	type Flag struct{}
	var flag Flag

	fn = func(groups ...*ViewGroup) (ret *ViewGroup) {
		for _, group := range groups {
			link(flag, group)
		}
		linkedOne(flag, &ret)
		return
	}

	return
}

func (g *ViewGroup) GetViews(scope Scope) []*View {
	var linkedAll LinkedAll
	scope.Assign(&linkedAll)
	var views []*View
	linkedAll(g, &views)

	if len(views) > 0 && len(g.SortKeys) > 0 {
		// sort
		var keyFuncs []ViewSortFunc
		for _, name := range g.SortKeys {
			if fn, ok := NamedViewSortKeys[name]; ok {
				keyFuncs = append(keyFuncs, fn)
			}
		}
		sortValues := make(map[ViewID][]int)
		for _, view := range views {
			for _, fn := range keyFuncs {
				sortValues[view.ID] = append(
					sortValues[view.ID],
					fn(view),
				)
			}
		}
		sort.SliceStable(views, func(i, j int) bool {
			for k := range keyFuncs {
				v1 := sortValues[views[i].ID][k]
				v2 := sortValues[views[j].ID][k]
				if v1 != v2 {
					return v1 < v2
				}
			}
			return true
		})

	}

	return views
}
