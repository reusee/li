package li

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell"
)

type SimScreen struct {
	tcell.Screen
}

func (_ SimScreen) SetCursorShape(shape CursorShape) {
}

type (
	GetSimScreenContents func() ([]tcell.SimCell, int, int)
	GetScreenString      func(Box) []string
)

func withEditor(fn any) {
	// scope
	var scope Scope
	funcCalls := make(chan any, 128)
	var derives []any
	scope = NewGlobal(
		func() Derive {
			return func(inits ...any) {
				derives = append(derives, inits...)
			}
		},
		func() RunInMainLoop {
			return func(fn any) {
				funcCalls <- fn
			}
		},
	)

	screen := tcell.NewSimulationScreen("")
	ce(screen.Init())
	var on On
	scope.Assign(&on)
	on(EvExit, func() {
		screen.Fini()
	})
	screen.EnableMouse()
	screen.SetSize(80, 25)
	scope = scope.Sub(
		func() Screen {
			return SimScreen{
				Screen: screen,
			}
		},
		func() SetContent {
			return screen.SetContent
		},
		func() AppendJournal {
			return func(format string, args ...any) {
				fmt.Printf(format, args...)
				fmt.Printf("\n")
			}
		},
	)

	events := make(chan ScreenEvent, 512)

	// panic handling
	var exit Exit
	var renderTimer RenderTimer
	var trigger Trigger
	scope.Assign(
		&exit,
		&renderTimer,
		&trigger,
	)
	defer exit()

	applyDerives := func() {
		if len(derives) > 0 {
			scope = scope.Sub(derives...)
			derives = derives[:0]
		}
	}

	loop := func() {
	loop:
		for {

			trigger(scope, EvLoopBegin)
			applyDerives()

			var root Element
			scope.Call(Root, &root)
			renderAll(scope, root)
			screen.Show()
			applyDerives()

			select {

			case ev := <-events:
				scope.Sub(func() ScreenEvent { return ev }).Call(HandleScreenEvent)

			case fn := <-funcCalls:
				scope.Call(fn)

			case <-renderTimer.C:

			default:
				break loop

			}
			applyDerives()

			trigger(scope, EvLoopEnd)
			applyDerives()

		}
		applyDerives()
	}

	scope.Sub(
		func() *Scope {
			return &scope
		},

		func() func(x string) {
			return func(x string) {
				switch x {

				case "loop":
					loop()

				default:
					panic("no " + x)
				}
			}

		},

		func() EmitRune {
			return func(r rune) {
				events <- tcell.NewEventKey(tcell.KeyRune, r, 0)
				loop()
			}
		},

		func() EmitKey {
			return func(key tcell.Key) {
				events <- tcell.NewEventKey(key, 0, 0)
				loop()
			}
		},

		func() EmitEvent {
			return func(ev ScreenEvent) {
				events <- ev
				loop()
			}
		},

		func() GetSimScreenContents {
			return func() ([]tcell.SimCell, int, int) {
				return screen.GetContents()
			}
		},

		func(
			contents GetSimScreenContents,
		) GetScreenString {
			return func(box Box) (ret []string) {
				cells, width, height := contents()
				for y := box.Top; y < box.Bottom && y < height; y++ {
					buf := new(strings.Builder)
					x := box.Left
					for x < box.Right && x < width {
						cell := cells[y*width+x]
						buf.Write(cell.Bytes)
						x += runesDisplayWidth(cell.Runes)
					}
					ret = append(ret, buf.String())
				}
				return
			}
		},
	).Call(fn)

}

func withHelloEditor(t *testing.T, fn any) {
	withEditorBytes(t, []byte("Hello, world!\n你好，世界！\nこんにちは、世界！\n"), fn)
}

func withEditorBytes(t *testing.T, bs []byte, fn any) {
	withEditor(func(
		s Scope,
	) {

		var buf *Buffer
		var err error
		s.Sub(func() []byte {
			return bs
		}).Call(NewBufferFromBytes, &buf, &err)
		if err != nil {
			t.Fatal(err)
		}
		if buf == nil {
			t.Fatal()
		}

		var view *View
		s.Sub(func() *Buffer {
			return buf
		}).Call(NewViewFromBuffer, &view, &err)
		if err != nil {
			t.Fatal(err)
		}
		if view == nil {
			t.Fatal()
		}

		s.Sub(
			func() (*View, *Buffer, *Moment) {
				return view, buf, view.GetMoment()
			},
		).Call(fn)

	})
}

func eq(t *testing.T, args ...any) {
	t.Helper()
	if len(args)%2 != 0 {
		t.Fatal("must be even number of args")
	}
	type Result struct {
		J1    []byte
		J2    []byte
		Equal bool
	}
	var results []Result
	for i := 0; i < len(args); i += 2 {
		j1, err := json.Marshal(args[i])
		if err != nil {
			t.Fatal(err)
		}
		j2, err := json.Marshal(args[i+1])
		if err != nil {
			t.Fatal(err)
		}
		results = append(results, Result{
			J1:    j1,
			J2:    j2,
			Equal: bytes.Equal(j1, j2),
		})
	}
	fatal := false
	for i, res := range results {
		if !res.Equal {
			fatal = true
			fmt.Printf(
				"pair %d not equal:\ngot %s\n------\nexpected %s\n",
				i+1,
				res.J1,
				res.J2,
			)
		}
	}
	if fatal {
		t.Fatal()
	}
}

func TestGetScreenString(t *testing.T) {
	withHelloEditor(t, func(
		view *View,
		get GetScreenString,
		scope Scope,
		emitRune EmitRune,
	) {
		scope.Call(EnableEditMode)
		emitRune('H')
		lines := get(view.ContentBox)
		eq(t,
			strings.HasPrefix(lines[0], "HHello, world!"), true,
			strings.HasPrefix(lines[1], "你好，世界！"), true,
			strings.HasPrefix(lines[2], "こんにちは、世界！"), true,
		)
	})
}
