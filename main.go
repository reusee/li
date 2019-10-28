package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell"
	"github.com/reusee/e/v2"
	"github.com/reusee/li/li"
)

type (
	Scope = li.Scope
	any   = interface{}
)

var (
	me     = e.Default.WithStack().WithName("li-editor")
	ce, he = e.New(me)
	pt     = fmt.Printf
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
	scope = scope.Sub(func() li.Views { return views })
	for _, path := range os.Args[1:] {
		var buffers []*li.Buffer
		var err error
		scope.Sub(func() string {
			return path
		}).Call(li.NewBuffersFromPath, &buffers, &err)
		if err != nil {
			return
		}
		for _, buffer := range buffers {
			scope.Sub(func() *li.Buffer {
				return buffer
			}).Call(li.NewViewFromBuffer, &err)
			ce(err)
		}
	}

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

	scope = scope.Sub(
		func() li.SetContent {
			return screen.SetContent
		},
	)

	for {

		trigger(scope, li.EvLoopBegin)

		select {

		case ev := <-screenEvents:
			scope.Sub(func() li.ScreenEvent { return ev }).Call(li.HandleScreenEvent)
			resetRenderTimer()

		case <-sigExit:
			return

		case fn := <-funcCalls:
			scope.Call(fn)
			resetRenderTimer()

		case <-renderTimer.C:
			// render
			var root li.Element
			scope.Call(li.Root, &root)
			scope.Call(root.RenderFunc())
			screen.Show()

		}

		if len(derives) > 0 {
			scope = scope.Sub(derives...)
			derives = derives[:0]
		}

		trigger(scope, li.EvLoopEnd)

	}

}
