package li

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gdamore/tcell"
)

type SimScreen struct {
	tcell.Screen
}

func (_ SimScreen) SetCursorShape(shape CursorShape) {
}

func withEditor(fn any) {
	// scope
	var scope Scope
	funcCalls := make(chan any, 128)
	scope = NewGlobal(
		func() Derive {
			return func(inits ...any) {
				scope = scope.Sub(inits...)
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
	scope.Assign(&exit)
	defer exit()

	loop := func() {
	loop:
		for {

			var root Element
			scope.Call(Root, &root)
			renderAll(scope, root)
			screen.Show()

			select {

			case ev := <-events:
				scope.Sub(func() ScreenEvent { return ev }).Call(HandleScreenEvent)

			case fn := <-funcCalls:
				scope.Call(fn)

			default:
				break loop

			}
		}
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
				"pair %d not equal:\n%s\n------\n%s\n",
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
