package li

import (
	"time"

	"github.com/gdamore/tcell"
)

type EditMode struct{}

func EnableEditMode(
	cur CurrentModes,
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

	// block cursor
	pt("\033[6 q")

}

func DisableEditMode(
	cur CurrentModes,
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

	// bar cursor
	pt("\033[2 q")

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

var _ KeyStrokeHandler = new(ReadMode)

func (_ EditMode) StrokeSpecs() any {
	return func(
		config EditModeConfig,
	) []StrokeSpec {

		disableSeqRunes := []rune(config.DisableSequence)
		inserted := make([]*tcell.EventKey, len(disableSeqRunes))

		specs := []StrokeSpec{

			// input
			{
				Predict: func(ev KeyEvent) bool {
					return ev.Key() == tcell.KeyRune
				},
				Func: func(
					ev KeyEvent,
					scope Scope,
					cur CurrentView,
					dropLink DropLink,
				) {
					r := ev.Rune()

					if len(disableSeqRunes) > 0 {
						if r == disableSeqRunes[len(disableSeqRunes)-1] {
							trigger := true
							for i := len(disableSeqRunes) - 2; i >= 0; i-- {
								if inserted[i+1] == nil ||
									inserted[i+1].Rune() != disableSeqRunes[i] ||
									time.Since(inserted[i+1].When()) > (time.Millisecond*100*time.Duration(len(disableSeqRunes))) {
									trigger = false
									break
								}
							}
							if trigger {
								view := cur()
								moment := view.Moment
								for i := len(disableSeqRunes) - 2; i >= 0; i-- {
									dropLink(view.Buffer, moment)
									moment = moment.Previous
								}
								view.switchMoment(moment)
								scope.Call(DisableEditMode)
								return
							}
						}
					}

					scope.Sub(func() (PositionFunc, string) {
						return PosCursor, string(r)
					}).Call(InsertAtPositionFunc)
					if len(inserted) > 0 {
						copy(inserted[0:len(inserted)-1], inserted[1:len(inserted)])
						inserted[len(inserted)-1] = ev
					}
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
