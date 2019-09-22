package li

import (
	"sort"
	"time"
)

type UndoConfig struct {
	DurationMS1 time.Duration
}

func (_ Provide) UndoConfig(
	getConfig GetConfig,
) UndoConfig {
	var config struct {
		Undo UndoConfig
	}
	config.Undo.DurationMS1 = 3000
	ce(getConfig(&config))
	return config.Undo
}

func Undo(
	cur CurrentView,
	scope Scope,
) {
	view := cur()
	if view == nil {
		return
	}
	if view.Moment.Previous != nil {
		view.switchMoment(scope, view.Moment.Previous)
	}
}

func (_ Command) Undo() (spec CommandSpec) {
	spec.Desc = "undo"
	spec.Func = Undo
	return
}

func RedoLatest(
	cur CurrentView,
	linkedAll LinkedAll,
	scope Scope,
) {
	view := cur()
	if view == nil {
		return
	}
	var allMoments []*Moment
	linkedAll(view.Buffer, &allMoments)
	var moments []*Moment
	currentMoment := view.Moment
	for _, moment := range allMoments {
		if moment.Previous == currentMoment {
			moments = append(moments, moment)
		}
	}
	if len(moments) == 0 {
		return
	}
	sort.SliceStable(moments, func(i, j int) bool {
		return moments[i].ID > moments[j].ID
	})
	view.switchMoment(scope, moments[0])
}

func (_ Command) RedoLatest() (spec CommandSpec) {
	spec.Desc = "redo latest undo"
	spec.Func = RedoLatest
	return
}

func UndoDuration1(
	cur CurrentView,
	config UndoConfig,
	scope Scope,
) {
	view := cur()
	if view == nil {
		return
	}
	var next *Moment
	t0 := view.Moment.T0
	prev := view.Moment.Previous
	for {
		if prev == nil {
			break
		}
		next = prev
		if t0.Sub(prev.T0) > config.DurationMS1*time.Millisecond {
			break
		}
		prev = prev.Previous
	}
	if next != nil {
		view.switchMoment(scope, next)
	}
}

func (_ Command) UndoDuration1() (spec CommandSpec) {
	spec.Desc = "undo to previous moment at least Undo.DurationMS1 earlier"
	spec.Func = UndoDuration1
	return
}
