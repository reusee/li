package main

import (
	"os"

	"github.com/gdamore/tcell"
	"github.com/reusee/li/li"
)

type (
	Scope = li.Scope
	any   = interface{}
)

func main() {

	// scope
	var scope Scope
	funcCalls := make(chan any, 128)
	var derives []any
	scope = li.NewGlobal(
		func() li.Derive {
			return func(inits ...any) {
				derives = append(derives, inits...)
			}
		},
		func() li.RunInMainLoop {
			return func(fn any) {
				funcCalls <- fn
			}
		},
	)

	// open files
	views := make(li.Views)
	scope = scope.Fork(func() li.Views { return views })
	scope.Call(func(
		newBuffers li.NewBuffersFromPath,
		newView li.NewViewFromBuffer,
	) {
		for _, path := range os.Args[1:] {
			var buffers []*li.Buffer
			var err error
			buffers, err = newBuffers(path)
			if err != nil {
				return
			}
			for _, buffer := range buffers {
				_, err := newView(buffer)
				ce(err)
			}
		}
	})

	var screen li.Screen
	defer func() {
		screen.Fini()
	}()
	scope.Assign(&screen)
	screenEvents := make(chan li.ScreenEvent, 128)
	go func() {
		for {
			ev := screen.PollEvent()
			if mouse, ok := ev.(*tcell.EventMouse); ok && mouse.Buttons() == 0 {
				// no mouse mouve events
				continue
			}
			screenEvents <- ev
		}
	}()

	// panic handling
	var exit li.Exit
	scope.Assign(&exit)
	defer exit()

	// main loop
	var sigExit li.SigExit
	var renderTimer li.RenderTimer
	var resetRenderTimer li.ResetRenderTimer
	var trigger li.Trigger
	scope.Assign(
		&sigExit,
		&renderTimer,
		&resetRenderTimer,
		&trigger,
	)

	scope = scope.Fork(
		func() li.SetContent {
			return screen.SetContent
		},
	)

	applyDerives := func() {
		if len(derives) > 0 {
			scope = scope.Fork(derives...)
			derives = derives[:0]
		}

	}

	for {

		trigger(li.EvLoopBegin{})
		applyDerives()

		select {

		case ev := <-screenEvents:
			scope.Call(func(
				handle li.HandleScreenEvent,
			) {
				handle(ev)
			})
			resetRenderTimer()

		case <-sigExit:
			return

		case fn := <-funcCalls:
			scope.Call(fn)
			resetRenderTimer()

		case <-renderTimer.C:
			// render
			var root li.Element
			scope.Call(li.Root).Assign(&root)
			scope.Call(root.RenderFunc())
			screen.Show()

		}
		applyDerives()

		trigger(li.EvLoopEnd{})
		applyDerives()

	}

}
