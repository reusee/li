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
		move MoveCursor,
	) {

		// empty move
		move(Move{})
		eq(t,
			view.CursorLine, 0,
			view.CursorCol, 0,
		)

		// move cursor, col
		move(Move{AbsCol: intP(1)})
		eq(t,
			view.CursorCol, 1,
		)
		move(Move{RelRune: 2})
		eq(t,
			view.CursorCol, 3,
		)
		move(Move{RelRune: -1})
		eq(t,
			view.CursorCol, 2,
		)
		move(Move{RelRune: 200})
		eq(t,
			view.CursorCol, 18,
			view.CursorLine, 2,
		)

		// move cursor, line
		move(Move{AbsLine: intP(1)})
		eq(t,
			view.CursorLine, 1,
		)
		move(Move{RelLine: -1})
		eq(t,
			view.CursorLine, 0,
		)

		// move down to middle of rune
		move(Move{AbsLine: intP(0), AbsCol: intP(1)})
		move(Move{RelLine: 1})
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 1,
		)

		// move left to wide rune
		move(Move{AbsLine: intP(1), AbsCol: intP(2)})
		move(Move{RelRune: -1})
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 1,
		)

		// move right from wide rune
		move(Move{AbsLine: intP(1), AbsCol: intP(0)})
		move(Move{RelRune: 1})
		eq(t,
			view.CursorCol, 2,
			view.CursorLine, 1,
		)

		// context number
		move(Move{AbsLine: intP(0), AbsCol: intP(0)})
		setN(3)
		move(Move{RelRune: 1})
		eq(t,
			view.CursorCol, 3,
			view.CursorLine, 0,
		)

		// crossing line break
		move(Move{AbsLine: intP(1), AbsCol: intP(0)})
		move(Move{RelRune: -10})
		eq(t,
			view.CursorCol, 4,
			view.CursorLine, 0,
		)
		move(Move{RelRune: 12})
		eq(t,
			view.CursorCol, 4,
			view.CursorLine, 1,
		)

		// move negative col
		move(Move{AbsLine: intP(0), AbsCol: intP(0)})
		move(Move{RelRune: -1})
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 0,
		)

		// move negative line
		move(Move{AbsLine: intP(0), AbsCol: intP(0)})
		move(Move{RelLine: -1})
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 0,
		)

		// move out of range line
		move(Move{AbsLine: intP(3), AbsCol: intP(0)})
		move(Move{RelLine: 1})
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 2,
		)

		// prefer col
		move(Move{AbsLine: intP(2), AbsCol: intP(18)})
		eq(t,
			view.CursorCol, 18,
			view.CursorLine, 2,
			view.PreferCursorCol, 18,
		)
		move(Move{RelLine: -1})
		eq(t,
			view.CursorCol, 12,
			view.CursorLine, 1,
		)
		move(Move{RelLine: 1})
		eq(t,
			view.CursorCol, 18,
			view.CursorLine, 2,
		)

		// at line begin and line end
		move(Move{AbsLine: intP(2), AbsCol: intP(0)})
		move(Move{RelRune: -1})
		eq(t,
			view.CursorCol, 12,
			view.CursorLine, 1,
		)
		move(Move{RelRune: 1})
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 2,
		)

		// move commands
		move(Move{AbsLine: intP(0), AbsCol: intP(1)})
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
		move(Move{AbsLine: intP(0), AbsCol: intP(1)})
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
		move(Move{AbsLine: intP(0), AbsCol: intP(0)})
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
		move(Move{AbsLine: intP(0), AbsCol: intP(0)})
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
		move(Move{AbsLine: intP(0), AbsCol: intP(0)})
		s.Fork(func() Func {
			return commands["NextRune"].Func
		}).Call(ExecuteCommandFunc)
		emitRune('e')
		eq(t,
			view.CursorCol, 1,
			view.CursorLine, 0,
		)
		s.Fork(func() Func {
			return commands["NextRune"].Func
		}).Call(ExecuteCommandFunc)
		emitRune(',')
		eq(t,
			view.CursorCol, 5,
			view.CursorLine, 0,
		)
		s.Fork(func() Func {
			return commands["NextRune"].Func
		}).Call(ExecuteCommandFunc)
		emitRune('z') // not exists
		eq(t,
			view.CursorCol, 5,
			view.CursorLine, 0,
		)
		s.Fork(func() Func {
			return commands["NextRune"].Func
		}).Call(ExecuteCommandFunc)
		emitKey(tcell.KeyEsc) // not rune
		eq(t,
			view.CursorCol, 5,
			view.CursorLine, 0,
		)
		s.Fork(func() Func {
			return commands["PrevRune"].Func
		}).Call(ExecuteCommandFunc)
		emitRune('H') // not rune
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 0,
		)
		s.Fork(func() Func {
			return commands["PrevRune"].Func
		}).Call(ExecuteCommandFunc)
		emitKey(tcell.KeyTab) // not rune
		eq(t,
			view.CursorCol, 0,
			view.CursorLine, 0,
		)
		s.Fork(func() Func {
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
		move MoveCursor,
	) {
		move(Move{AbsLine: intP(0), AbsCol: intP(0)})
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
		s = s.Fork(
			func() ScrollConfig {
				return ScrollConfig{
					PaddingTop:    1,
					PaddingBottom: 1,
				}
			},
		)
		var move MoveCursor
		s.Assign(&move)
		move(Move{AbsLine: intP(0), AbsCol: intP(0)})
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
		move MoveCursor,
	) {
		move(Move{AbsLine: intP(0), AbsCol: intP(0)})
		scope.Call(commands["PageDown"].Func)
		scope.Call(commands["PageUp"].Func)
		scope.Call(commands["NextEmptyLine"].Func)
		scope.Call(commands["PrevEmptyLine"].Func)
		scope.Call(commands["LineBegin"].Func)
		scope.Call(commands["LineEnd"].Func)
		scope.Fork(func() Func {
			return commands["NextRune"].Func
		}).Call(ExecuteCommandFunc)
		emitKey(tcell.KeyEsc) // not rune
	})
}
