package li

import "reflect"

type Command struct{}

type CommandSpec struct {
	Name string
	Desc string
	Func Func
}

type Commands = map[string]CommandSpec

var NamedCommands = func() Commands {
	m := make(map[string]CommandSpec)
	o := reflect.ValueOf(new(Command))
	t := o.Elem().Type()
	for i := 0; i < o.NumMethod(); i++ {
		spec := o.Method(i).Interface().(func() CommandSpec)()
		name := t.Method(i).Name
		spec.Name = name
		if spec.Desc == "" {
			spec.Desc = spec.Name
		}
		m[spec.Name] = spec
	}
	return m
}()

func (_ Provide) Commands() Commands {
	return NamedCommands
}

func ExecuteCommandFunc(
	fn Func,
	scope Scope,
	logImitation LogImitation,
	set SetStrokeSpecs,
	reset ResetStrokeSpecs,
) (
	abort Abort,
) {

call:
	// if true, do not reset context number
	var noResetN NoResetN
	// if non-empty, set as new stroke specs
	var specs []StrokeSpec
	// if true, do not log imitation
	var noLogImitation NoLogImitation
	// if non-nil, call again
	var moreFunc Func
	// abort execution
	scope.Call(
		fn,
		&noResetN,
		&specs,
		&noLogImitation,
		&moreFunc,
		&abort,
	)
	if !noLogImitation {
		logImitation(fn)
	}
	if moreFunc != nil {
		fn = moreFunc
		goto call
	}

	if len(specs) > 0 {
		// set new specs
		set(specs, false)
	} else if !abort {
		// reset to inital specs
		reset(noResetN)
	}

	return
}
