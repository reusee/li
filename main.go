package main

import (
	"fmt"
	"os"
	"time"

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
	scope = li.NewGlobal(func() li.Derive {
		return func(inits ...any) {
			scope = scope.Sub(inits...)
		}
	})

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

	// async func calls
	funcCalls := make(chan any, 128)
	idleFuncCalls := make(chan any, 128)
	scope = scope.Sub(
		func() li.RunInMainLoop {
			return func(fn any) {
				funcCalls <- fn
			}
		},
		func() li.RunWhenIdle {
			return func(fn any) {
				idleFuncCalls <- fn
			}
		},
	)

	// main loop
	var sigExit li.SigExit
	scope.Assign(&sigExit)

	scope = scope.Sub(
		func() li.SetContent {
			return screen.SetContent
		},
	)

	idleDuration := time.Second * 15
	idleTimer := time.NewTimer(idleDuration)
	renderDuration := time.Millisecond * 10
	renderTimer := time.NewTimer(renderDuration)

	for {

		idleTimer.Reset(idleDuration)

		select {

		case ev := <-screenEvents:
			scope.Sub(func() li.ScreenEvent { return ev }).Call(li.HandleScreenEvent)
			renderTimer.Reset(renderDuration)

		case <-sigExit:
			return

		case fn := <-funcCalls:
			scope.Call(fn)
			renderTimer.Reset(renderDuration)

		case <-renderTimer.C:
			// render
			var root li.Element
			scope.Call(li.Root, &root)
			scope.Call(root.RenderFunc())
			screen.Show()

		case <-idleTimer.C:
			t0 := time.Now()
		loop:
			for {
				select {
				case fn := <-idleFuncCalls:
					scope.Call(fn)
					if time.Since(t0) > time.Second {
						break loop
					}
				default:
					break loop
				}
			}
			renderTimer.Reset(renderDuration)

		}
	}

}
