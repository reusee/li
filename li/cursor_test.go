package li

import (
	"bytes"
	"testing"

	"github.com/gdamore/tcell"
)

func TestCursor(t *testing.T) {
	withHelloEditor(t, func(
		emitRune EmitRune,
		emitKey EmitKey,
		call func(string),
		s Scope,
		view *View,
		commands Commands,
		screen Screen,
		setN SetContextNumber,
	) {

		// empty move
		s.Sub(&Move{}).
			Call(MoveCursor)
		eq(t,
			view.CursorLine, 0,
			view.CursorCol, 0,
		)

		// move cursor, col
		s.Sub(&Move{AbsCol: intP(1)}).
			Call(MoveCursor)
		eq(t,
			view.CursorCol, 1,
		)
		s.Sub(&Move{RelRune: 2}).
			Call(MoveCursor)
		eq(t,
			view.CursorCol, 3,
		)
		s.Sub(&Move{RelRune: -1}).
			Call(MoveCursor)
		eq(t,
			view.CursorCol, 2,
		)
		s.Sub(&Move{RelRune: 200}).
			Call(MoveCursor)
		eq(t,
			view.CursorCol, 18,
			view.CursorLine, 2,
		)

		// move cursor, line
		s.Sub(&Move{AbsLine: intP(1)}).
			Call(MoveCursor)
		eq(t,
			view.CursorLine, 1,
		)
		s.Sub(&Move{RelLine: -1}).
			Call(MoveCursor)
		eq(t,
			view.CursorLine, 0,
		)

		// move down to middle of rune
		s.Sub(&Move{AbsLine: intP(0), AbsCol: intP(1)}).
			Call(MoveCursor)
		s.Sub(&Move{RelLine: 1}).
			Call(MoveCursor)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 1,
		)

		// move left to wide rune
		s.Sub(&Move{AbsLine: intP(1), AbsCol: intP(2)}).
			Call(MoveCursor)
		s.Sub(&Move{RelRune: -1}).
			Call(MoveCursor)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 1,
		)

		// move right from wide rune
		s.Sub(&Move{AbsLine: intP(1), AbsCol: intP(0)}).
			Call(MoveCursor)
		s.Sub(&Move{RelRune: 1}).
			Call(MoveCursor)
		eq(t,
			view.CursorCol, 2,
			view.CursorLine, 1,
		)

		// context number
		s.Sub(&Move{AbsLine: intP(0), AbsCol: intP(0)}).
			Call(MoveCursor)
		setN(3)
		s.Sub(&Move{RelRune: 1}).
			Call(MoveCursor)
		eq(t,
			view.CursorCol, 3,
			view.CursorLine, 0,
		)

		// crossing line break
		s.Sub(&Move{AbsLine: intP(1), AbsCol: intP(0)}).
			Call(MoveCursor)
		s.Sub(&Move{RelRune: -10}).
			Call(MoveCursor)
		eq(t,
			view.CursorCol, 4,
			view.CursorLine, 0,
		)
		s.Sub(&Move{RelRune: 12}).
			Call(MoveCursor)
		eq(t,
			view.CursorCol, 4,
			view.CursorLine, 1,
		)

		// move negative col
		s.Sub(&Move{AbsLine: intP(0), AbsCol: intP(0)}).
			Call(MoveCursor)
		s.Sub(&Move{RelRune: -1}).
			Call(MoveCursor)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 0,
		)

		// move negative line
		s.Sub(&Move{AbsLine: intP(0), AbsCol: intP(0)}).
			Call(MoveCursor)
		s.Sub(&Move{RelLine: -1}).
			Call(MoveCursor)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 0,
		)

		// move out of range line
		s.Sub(&Move{AbsLine: intP(3), AbsCol: intP(0)}).
			Call(MoveCursor)
		s.Sub(&Move{RelLine: 1}).
			Call(MoveCursor)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 2,
		)

		// prefer col
		s.Sub(&Move{AbsLine: intP(2), AbsCol: intP(18)}).
			Call(MoveCursor)
		eq(t,
			view.CursorCol, 18,
			view.CursorLine, 2,
			view.PreferCursorCol, 18,
		)
		s.Sub(&Move{RelLine: -1}).
			Call(MoveCursor)
		eq(t,
			view.CursorCol, 12,
			view.CursorLine, 1,
		)
		s.Sub(&Move{RelLine: 1}).
			Call(MoveCursor)
		eq(t,
			view.CursorCol, 18,
			view.CursorLine, 2,
		)

		// at line begin and line end
		s.Sub(&Move{AbsLine: intP(2), AbsCol: intP(0)}).
			Call(MoveCursor)
		s.Sub(&Move{RelRune: -1}).
			Call(MoveCursor)
		eq(t,
			view.CursorCol, 12,
			view.CursorLine, 1,
		)
		s.Sub(&Move{RelRune: 1}).
			Call(MoveCursor)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 2,
		)

		// move commands
		s.Sub(&Move{AbsLine: intP(0), AbsCol: intP(1)}).
			Call(MoveCursor)
		s.Call(commands["MoveLeft"].Func)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 0,
		)
		s.Call(commands["MoveDown"].Func)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 1,
		)
		s.Call(commands["MoveUp"].Func)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 0,
		)
		s.Call(commands["MoveRight"].Func)
		eq(t,
			view.CursorCol, 1,
			view.CursorLine, 0,
		)

		// page up / down
		s.Sub(&Move{AbsLine: intP(0), AbsCol: intP(1)}).
			Call(MoveCursor)
		s.Call(commands["PageDown"].Func)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 2,
		)
		s.Call(commands["PageUp"].Func)
		eq(t,
			view.CursorCol, 1, // prefer col
			view.CursorLine, 0,
		)

		// next empty line
		s.Sub(&Move{AbsLine: intP(0), AbsCol: intP(0)}).
			Call(MoveCursor)
		s.Call(commands["NextEmptyLine"].Func)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 2,
		)
		s.Call(commands["PrevEmptyLine"].Func)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 0,
		)

		// line begin / end
		s.Sub(&Move{AbsLine: intP(0), AbsCol: intP(0)}).
			Call(MoveCursor)
		s.Call(commands["LineEnd"].Func)
		eq(t,
			view.CursorCol, 13,
			view.CursorLine, 0,
		)
		s.Call(commands["LineBegin"].Func)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 0,
		)

		// next / prev rune
		s.Sub(&Move{AbsLine: intP(0), AbsCol: intP(0)}).
			Call(MoveCursor)
		s.Sub(func() Func {
			return commands["NextRune"].Func
		}).Call(ExecuteCommandFunc)
		emitRune('e')
		eq(t,
			view.CursorCol, 1,
			view.CursorLine, 0,
		)
		s.Sub(func() Func {
			return commands["NextRune"].Func
		}).Call(ExecuteCommandFunc)
		emitRune(',')
		eq(t,
			view.CursorCol, 5,
			view.CursorLine, 0,
		)
		s.Sub(func() Func {
			return commands["NextRune"].Func
		}).Call(ExecuteCommandFunc)
		emitRune('z') // not exists
		eq(t,
			view.CursorCol, 5,
			view.CursorLine, 0,
		)
		s.Sub(func() Func {
			return commands["NextRune"].Func
		}).Call(ExecuteCommandFunc)
		emitKey(tcell.KeyEsc) // not rune
		eq(t,
			view.CursorCol, 5,
			view.CursorLine, 0,
		)
		s.Sub(func() Func {
			return commands["PrevRune"].Func
		}).Call(ExecuteCommandFunc)
		emitRune('H') // not rune
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 0,
		)
		s.Sub(func() Func {
			return commands["PrevRune"].Func
		}).Call(ExecuteCommandFunc)
		emitKey(tcell.KeyTab) // not rune
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 0,
		)
		s.Sub(func() Func {
			return commands["PrevRune"].Func
		}).Call(ExecuteCommandFunc)
		emitRune('x') // not exists
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 0,
		)

	})
}

func TestPrevNextEmptyLine(t *testing.T) {
	withEditorBytes(t, []byte("a\n\nb\n\nc\n\nd\n"), func(
		s Scope,
		commands Commands,
		view *View,
	) {
		s.Sub(&Move{AbsLine: intP(0), AbsCol: intP(0)}).
			Call(MoveCursor)
		s.Call(commands["NextEmptyLine"].Func)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 1,
		)
		s.Call(commands["PrevEmptyLine"].Func)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 0,
		)
		s.Call(commands["NextEmptyLine"].Func)
		s.Call(commands["NextEmptyLine"].Func)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 3,
		)
		s.Call(commands["PrevEmptyLine"].Func)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 1,
		)

	})
}

func TestPageUpAndDown(t *testing.T) {
	withEditorBytes(t, bytes.Repeat([]byte("a\n"), 80), func(
		s Scope,
		commands Commands,
		view *View,
	) {
		s = s.Sub(
			func() ScrollConfig {
				return ScrollConfig{
					PaddingTop:    1,
					PaddingBottom: 1,
				}
			},
		)
		s.Sub(&Move{AbsLine: intP(0), AbsCol: intP(0)}).
			Call(MoveCursor)
		s.Call(commands["PageDown"].Func)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 24,
			view.ViewportLine, 23,
		)
		s.Call(commands["PageDown"].Func)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 48,
			view.ViewportLine, 47,
		)
		s.Call(commands["PageDown"].Func)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 72,
			view.ViewportLine, 71,
		)
		s.Call(commands["PageDown"].Func)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 79,
			view.ViewportLine, 78,
		)
		s.Call(commands["PageUp"].Func)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 55,
			view.ViewportLine, 54,
		)
		s.Call(commands["PageUp"].Func)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 31,
			view.ViewportLine, 30,
		)
		s.Call(commands["PageUp"].Func)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 7,
			view.ViewportLine, 6,
		)
		s.Call(commands["PageUp"].Func)
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 0,
			view.ViewportLine, 0,
		)
	})
}

func TestCursorNoView(t *testing.T) {
	withEditor(func(
		scope Scope,
		commands Commands,
		emitKey EmitKey,
	) {
		scope.Sub(&Move{AbsLine: intP(0), AbsCol: intP(0)}).
			Call(MoveCursor)
		scope.Call(commands["PageDown"].Func)
		scope.Call(commands["PageUp"].Func)
		scope.Call(commands["NextEmptyLine"].Func)
		scope.Call(commands["PrevEmptyLine"].Func)
		scope.Call(commands["LineBegin"].Func)
		scope.Call(commands["LineEnd"].Func)
		scope.Sub(func() Func {
			return commands["NextRune"].Func
		}).Call(ExecuteCommandFunc)
		emitKey(tcell.KeyEsc) // not rune
	})
}
