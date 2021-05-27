package li

import (
	"github.com/gdamore/tcell"
)

type (
	MacroRecording   bool
	GetMacroName     func() string
	StartMacroRecord func(name string)
	StopMacroRecord  func() (name string, events []KeyEvent)
	RecordMacroKey   func(KeyEvent)
)

func (_ Provide) DefaultMacroRecordingState() MacroRecording {
	return false
}

func (_ Provide) KeyMacro(
	derive Derive,
) (
	getName GetMacroName,
	start StartMacroRecord,
	stop StopMacroRecord,
	record RecordMacroKey,
) {
	var name string
	var events []KeyEvent
	start = func(n string) {
		name = n
		events = events[:0]
		derive(
			func() MacroRecording {
				return true
			},
		)
	}
	stop = func() (string, []KeyEvent) {
		derive(
			func() MacroRecording {
				return false
			},
		)
		return name, events
	}
	record = func(ev KeyEvent) {
		events = append(events, ev)
	}
	getName = func() string {
		return name
	}
	return
}

func ToggleMacroRecording(
	recording MacroRecording,
) Func {

	if recording {
		// stop
		return func(
			stop StopMacroRecord,
		) {
			name, events := stop()
			log("%s %d\n", name, len(events))
		}
	}

	// start
	var waitMacroName Func
	waitMacroName = func() []StrokeSpec {
		return []StrokeSpec{
			{
				Predict: func() bool {
					return true
				},
				Func: func(
					ev KeyEvent,
				) Func {
					if ev.Key() != tcell.KeyRune {
						// if not rune, retry
						return waitMacroName
					}
					name := string(ev.Rune())
					return func(start StartMacroRecord) {
						start(name)
					}
				},
			},
		}
	}

	return waitMacroName
}

func (_ Command) ToggleMacroRecording() (spec CommandSpec) {
	spec.Func = ToggleMacroRecording
	spec.Desc = "toggle key macro recording"
	return
}

func (_ Provide) KeyMacroStatus(
	on On,
) OnStartup {
	return func() {

		on(func(
			ev EvCollectStatusSections,
			recording MacroRecording,
			getMacroName GetMacroName,
		) {
			if recording {
				ev.Add("macro", [][]any{
					{getMacroName(), AlignRight, Padding(0, 2, 0, 0)},
				})
			}
		})

	}
}
