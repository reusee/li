package li

import (
	"strings"

	"github.com/gdamore/tcell"
)

type KeyEvent = *tcell.EventKey

type StrokeSpec struct {
	Sequence    []string
	Predict     any
	Func        any
	Command     CommandSpec
	CommandName string
}

type KeyStrokeHandler interface {
	StrokeSpecs() any
}

type (
	GetStrokeSpecs   func() ([]StrokeSpec, bool)
	SetStrokeSpecs   func([]StrokeSpec, bool)
	ResetStrokeSpecs func(
		noResetN NoResetN,
	) []StrokeSpec
)

func (_ Provide) StrokeSpecsAccessor() (
	get GetStrokeSpecs,
	set SetStrokeSpecs,
) {
	var cur []StrokeSpec
	var isInitial bool
	get = func() ([]StrokeSpec, bool) {
		return cur, isInitial
	}
	set = func(s []StrokeSpec, b bool) {
		cur = s
		isInitial = b
	}
	return
}

func (_ Provide) StrokeSpecs(
	curModes CurrentModes,
	setN SetN,
	overlays []Overlay,
	scope Scope,
	set SetStrokeSpecs,
	get GetStrokeSpecs,
) (
	reset ResetStrokeSpecs,
) {

	var initial []StrokeSpec
	for _, overlay := range overlays {
		if overlay.KeyStrokeHandler != nil {
			var specs []StrokeSpec
			scope.Call(overlay.KeyStrokeHandler.StrokeSpecs(), &specs)
			initial = append(initial, specs...)
		}
	}
	for _, mode := range curModes() {
		if mode, ok := mode.(KeyStrokeHandler); ok {
			var specs []StrokeSpec
			scope.Call(mode.StrokeSpecs(), &specs)
			initial = append(initial, specs...)
		}
	}

	reset = func(
		noResetN NoResetN,
	) []StrokeSpec {
		if !noResetN {
			setN(0)
		}
		set(initial, true)
		return initial
	}

	if _, isInitial := get(); isInitial {
		// after command palette is closed, if current specs is the initial then reset, otherwise keep current
		reset(false)
	}

	return
}

func strokeSpecsFromSequenceCommand(m map[string]string) []StrokeSpec {
	var specs []StrokeSpec
	for seq, name := range m {
		if _, ok := NamedCommands[name]; !ok {
			panic(me(nil, "no such command: %s", name))
		}
		sequence := strings.Split(seq, " ")
		var nice []string
		for _, s := range sequence {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			nice = append(nice, s)
		}
		sequence = nice
		specs = append(specs, StrokeSpec{
			Sequence:    sequence,
			CommandName: strings.TrimSpace(name),
		})
	}
	return specs
}

type (
	SetLastKeyEvent func(KeyEvent)
	GetLastKeyEvent func() KeyEvent
)

func (_ Provide) KeyLogging() (
	set SetLastKeyEvent,
	get GetLastKeyEvent,
) {
	var ev KeyEvent
	set = func(e KeyEvent) {
		ev = e
	}
	get = func() KeyEvent {
		return ev
	}
	return
}

type evKeyEventHandled struct{}

var EvKeyEventHandled = new(evKeyEventHandled)

func HandleKeyEvent(
	ev KeyEvent,
	reset ResetStrokeSpecs,
	set SetStrokeSpecs,
	get GetStrokeSpecs,
	scope Scope,
	setEv SetLastKeyEvent,
	commands Commands,
	recording MacroRecording,
	record RecordMacroKey,
	trigger Trigger,
) {

	defer func() {
		trigger(scope, EvKeyEventHandled)
	}()

	if recording {
		record(ev)
	}

	specs, _ := get()
	if len(specs) == 0 {
		specs = reset(false)
	}

	r := ev.Name()
	setEv(ev)
	keyScope := scope.Sub(func() KeyEvent {
		return ev
	})

	var nextSpecs []StrokeSpec
	for _, spec := range specs {
		match := false
		if len(spec.Sequence) == 1 && spec.Sequence[0] == r {
			// match sequence
			match = true

		} else if len(spec.Sequence) > 1 && spec.Sequence[0] == r {
			// match sequence prefix
			newSpec := spec
			newSpec.Sequence = spec.Sequence[1:]
			nextSpecs = append(nextSpecs, newSpec)

		} else if spec.Predict != nil { // assuming len(spec.Sequence) == 0
			// call predict function
			keyScope.Call(spec.Predict, &match)
		}

		if match {

			// call handling function
			var fn any
			if spec.Func != nil {
				fn = spec.Func
			} else if spec.Command.Func != nil {
				fn = spec.Command.Func
			} else if spec.CommandName != "" && commands[spec.CommandName].Func != nil {
				fn = commands[spec.CommandName].Func
			}
			var abort Abort
			keyScope.Sub(func() Func { return fn }).Call(
				ExecuteCommandFunc,
				&abort,
			)

			if !abort {
				return
			}
		}

	}

	// no match
	set(nextSpecs, false)
}

func (_ Provide) KeyStatus(
	on On,
) Init2 {

	on(EvCollectStatusSections, func(
		add AddStatusSection,
		getKeyEv GetLastKeyEvent,
	) {
		if ev := getKeyEv(); ev != nil {
			add("key", [][]any{
				{ev.Name(), AlignRight, Padding(0, 2, 0, 0)},
			})
		}
	})

	return nil
}
