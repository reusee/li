package li

import (
	"time"

	"github.com/gdamore/tcell"
)

type EditMode struct {
	matchStates     []editModeMatchState
	disableSeqRunes []rune
}

type editModeMatchState dyn

func EnableEditMode(
	cur CurrentModes,
	screen Screen,
) {

	modes := cur()
	enabled := false
	for _, mode := range modes {
		if _, ok := mode.(*EditMode); ok {
			enabled = true
			break
		}
	}
	if enabled {
		return
	}

	newModes := make([]Mode, len(modes)+1)
	copy(newModes[1:], modes)
	newModes[0] = new(EditMode)
	cur(newModes)

	screen.SetCursorShape(CursorBeam)

}

func DisableEditMode(
	cur CurrentModes,
	screen Screen,
) {

	modes := cur()
	filtered := make([]Mode, 0, len(modes))
	for _, mode := range modes {
		if _, ok := mode.(*EditMode); ok {
			continue
		}
		filtered = append(filtered, mode)
	}
	cur(filtered)

	screen.SetCursorShape(CursorBlock)

}

type EditModeConfig struct {
	DisableSequence string
	SequenceCommand map[string]string
}

func (_ Provide) EditModeConfig(
	getConfig GetConfig,
) EditModeConfig {
	var config struct {
		EditMode EditModeConfig
	}
	ce(getConfig(&config))
	return config.EditMode
}

var _ KeyStrokeHandler = new(EditMode)

func (e *EditMode) matchDisableSeq(
	index int,
	late time.Time,
	rollback *Moment,
) editModeMatchState {
	return func(
		ev KeyEvent,
		cur CurrentView,
		scope Scope,
		dropLink DropLink,
	) (bool, editModeMatchState) {
		// not match sequence
		if ev.Rune() != e.disableSeqRunes[index] {
			return false, nil
		}
		// late
		if time.Now().After(late) {
			return false, nil
		}
		// match next
		if index+1 < len(e.disableSeqRunes) {
			return false, e.matchDisableSeq(
				index+1,
				time.Now().Add(time.Millisecond*100),
				rollback,
			)
		}
		// trigger
		if rollback != nil {
			view := cur()
			moment := view.GetMoment()
			for moment != rollback {
				dropLink(view.Buffer, moment)
				moment = moment.Previous
			}
			view.switchMoment(scope, rollback)
		}
		scope.Call(DisableEditMode)
		return true, nil
	}
}

func (e *EditMode) StrokeSpecs() any {
	return func(
		config EditModeConfig,
	) []StrokeSpec {

		e.disableSeqRunes = []rune(config.DisableSequence)

		specs := []StrokeSpec{

			{
				Predict: func(ev KeyEvent) bool {
					return ev.Key() == tcell.KeyRune
				},
				Func: func(
					scope Scope,
					cur CurrentView,
					ev KeyEvent,
				) {

					// match disable sequence
					e.matchStates = append(
						e.matchStates,
						e.matchDisableSeq(0, never, cur().GetMoment()),
					)
					ts := e.matchStates[:0]
					for _, thread := range e.matchStates {
						var next editModeMatchState
						var handled bool
						scope.Call(thread, &next, &handled)
						if handled {
							e.matchStates = e.matchStates[:0]
							return
						} else if next != nil {
							ts = append(ts, next)
						}
					}
					e.matchStates = ts

					// insert
					fn := PositionFunc(PosCursor)
					str := string(ev.Rune())
					scope.Sub(
						&fn, &str,
					).Call(InsertAtPositionFunc)

				},
			},

			// special keys
			{
				Sequence:    []string{"Backspace2"},
				CommandName: "DeletePrevRune",
			},
			{
				Sequence:    []string{"Backspace"},
				CommandName: "DeletePrevRune",
			},
			{
				Sequence:    []string{"Enter"},
				CommandName: "InsertNewline",
			},
			{
				Sequence:    []string{"Delete"},
				CommandName: "DeleteRune",
			},
			{
				Sequence:    []string{"Tab"},
				CommandName: "InsertTab",
			},

			//
		}

		specs = append(specs, strokeSpecsFromSequenceCommand(config.SequenceCommand)...)

		return specs
	}
}

func (_ Command) EnableEditMode() (spec CommandSpec) {
	spec.Desc = "enable edit mode"
	spec.Func = EnableEditMode
	return
}

func (_ Command) DisableEditMode() (spec CommandSpec) {
	spec.Desc = "disable edit mode"
	spec.Func = DisableEditMode
	return
}

func IsEditing(modes []Mode) bool {
	editing := false
	for _, mode := range modes {
		if _, ok := mode.(*EditMode); ok {
			editing = true
			break
		}
	}
	return editing
}
