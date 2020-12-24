package li

import (
	"testing"
)

func TestMomentFromBytes(t *testing.T) {
	withEditorBytes(t, []byte("abc"), func(
		view *View,
		scope Scope,
		move MoveCursor,
	) {
		eq(t,
			view.GetMoment().NumLines(), 1,
		)
		move(Move{AbsLine: intP(999)})
		eq(t,
			view.CursorLine, 0,
		)
	})
}

func TestSplitLines(t *testing.T) {
	lines := splitLines("")
	eq(t,
		len(lines), 1,
		lines[0] == "", true,
	)
	lines = splitLines("\n")
	eq(t,
		len(lines), 1,
		lines[0] == "\n", true,
	)
	lines = splitLines("\n\n")
	eq(t,
		len(lines), 2,
		lines[0] == "\n", true,
		lines[1] == "\n", true,
	)
	lines = splitLines("a\nb")
	eq(t,
		len(lines), 2,
		lines[0] == "a\n", true,
		lines[1] == "b", true,
	)
	lines = splitLines("a\nb\n")
	eq(t,
		len(lines), 2,
		lines[0] == "a\n", true,
		lines[1] == "b\n", true,
	)
	lines = splitLines("foo")
	eq(t,
		len(lines), 1,
		lines[0] == "foo", true,
	)
}

func TestLowerContent(t *testing.T) {
	withHelloEditor(t, func(
		view *View,
	) {
		str := view.GetMoment().GetLowerContent()
		eq(t,
			str, "hello, world!\n你好，世界！\nこんにちは、世界！\n",
		)
	})
}

func TestDerivedMomentLanguage(t *testing.T) {
	withEditorBytes(t, []byte(`package main
	  func main() {}
	`), func(
		moment *Moment,
		view *View,
		scope Scope,
		buffer *Buffer,
	) {
		buffer.SetLanguage(scope, LanguageGo)
		parser := moment.GetParser(scope)
		eq(t,
			parser != nil, true,
		)
		scope.Call(ChangeToWordEnd)
		parser = view.GetMoment().GetParser(scope)
		eq(t,
			parser != nil, true,
		)
	})
}

func TestCellUTF16Offset(t *testing.T) {
	withHelloEditor(t, func(
		m *Moment,
		scope Scope,
	) {
		line := m.GetLine(scope, 0)
		eq(t,
			line.Cells[0].UTF16Offset, 0,
			line.Cells[1].UTF16Offset, 2,
			line.Cells[2].UTF16Offset, 4,
		)
	})
}

func TestMomentByteOffsetToPosition(t *testing.T) {
	withHelloEditor(t, func(
		m *Moment,
		scope Scope,
	) {
		pos := m.ByteOffsetToPosition(scope, 0)
		eq(t,
			pos.Line, 0,
			pos.Cell, 0,
		)
		pos = m.ByteOffsetToPosition(scope, 1)
		eq(t,
			pos.Line, 0,
			pos.Cell, 1,
		)
		pos = m.ByteOffsetToPosition(scope, 13)
		eq(t,
			pos.Line, 0,
			pos.Cell, 13,
		)
		pos = m.ByteOffsetToPosition(scope, 14)
		eq(t,
			pos.Line, 1,
			pos.Cell, 0,
		)
		pos = m.ByteOffsetToPosition(scope, 15)
		eq(t,
			pos.Line, 1,
			pos.Cell, 0,
		)
		pos = m.ByteOffsetToPosition(scope, 17)
		eq(t,
			pos.Line, 1,
			pos.Cell, 1,
		)
		pos = m.ByteOffsetToPosition(scope, 33)
		eq(t,
			pos.Line, 2,
			pos.Cell, 0,
		)
	})
}
