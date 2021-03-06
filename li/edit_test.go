package li

import (
	"bytes"
	"testing"
)

func TestApplyChange(t *testing.T) {
	withEditor(func(
		scope Scope,
		newMoment NewMomentFromBytes,
		apply ApplyChange,
	) {
		moment, _, err := newMoment([]byte(``))
		if err != nil {
			t.Fatal(err)
		}
		eq(t,
			moment != nil, true,
		)

		var m2 *Moment
		change := Change{
			Op: OpInsert,
			Begin: Position{
				Line: 0,
				Cell: 0,
			},
			String: "foo",
		}
		m2, _ = apply(moment, change)
		eq(t,
			m2 != nil, true,
			m2.ID > 0, true,
			m2.Previous == moment, true,
			m2.Change == change, true,
			m2.NumLines(), 1,
			m2.GetLine(0).content, "foo\n",
		)

		var m3 *Moment
		change = Change{
			Op: OpInsert,
			Begin: Position{
				Line: 0,
				Cell: 0,
			},
			String: "foo\nfoo\nfoo",
		}
		m3, _ = apply(m2, change)
		eq(t,
			m3 != nil, true,
			m3.ID > 0, true,
			m3.Previous == m2, true,
			m3.Change == change, true,
			m3.NumLines(), 3,
			m3.GetLine(0).content, "foo\n",
			m3.GetLine(1).content, "foo\n",
			m3.GetLine(2).content, "foofoo\n",
		)

		var m4 *Moment
		change = Change{
			Op: OpInsert,
			Begin: Position{
				Line: 2,
				Cell: 6,
			},
			String: "\nbar",
		}
		m4, _ = apply(m3, change)
		eq(t,
			m4 != nil, true,
			m4.ID > 0, true,
			m4.Previous == m3, true,
			m4.Change == change, true,
			m4.NumLines(), 4,

			m4.GetLine(0).content, "foo\n",
			m4.GetLine(1).content, "foo\n",
			m4.GetLine(2).content, "foofoo\n",
			m4.GetLine(3).content, "bar\n",
		)

		var m5 *Moment
		change = Change{
			Op: OpInsert,
			Begin: Position{
				Line: 1,
				Cell: 1,
			},
			String: "quux",
		}
		m5, _ = apply(m4, change)
		eq(t,
			m5 != nil, true,
			m5.ID > 0, true,
			m5.Previous == m4, true,
			m5.Change == change, true,
			m5.NumLines(), 4,

			m5.GetLine(0).content, "foo\n",
			m5.GetLine(1).content, "fquuxoo\n",
			m5.GetLine(2).content, "foofoo\n",
			m5.GetLine(3).content, "bar\n",
		)

		var m6 *Moment
		change = Change{
			Op: OpInsert,
			Begin: Position{
				Line: 0,
				Cell: 3,
			},
			String: "你好\n世界\n",
		}
		m6, _ = apply(m5, change)
		eq(t,
			m6 != nil, true,
			m6.ID > 0, true,
			m6.Previous == m5, true,
			m6.Change == change, true,
			m6.NumLines(), 6,

			m6.GetLine(0).content, "foo你好\n",
			m6.GetLine(1).content, "世界\n",
			m6.GetLine(2).content, "\n",
			m6.GetLine(3).content, "fquuxoo\n",
			m6.GetLine(4).content, "foofoo\n",
			m6.GetLine(5).content, "bar\n",
		)

		var m7 *Moment
		change = Change{
			Op: OpDelete,
			Begin: Position{
				Line: 0,
				Cell: 0,
			},
			Number: 3,
		}
		m7, _ = apply(m6, change)
		eq(t,
			m7 != nil, true,
			m7.ID > 0, true,
			m7.Previous == m6, true,
			m7.Change == change, true,
			m7.NumLines(), 6,

			m7.GetLine(0).content, "你好\n",
			m7.GetLine(1).content, "世界\n",
			m7.GetLine(2).content, "\n",
			m7.GetLine(3).content, "fquuxoo\n",
			m7.GetLine(4).content, "foofoo\n",

			m7.GetLine(5).content, "bar\n",
		)

		var m8 *Moment
		change = Change{
			Op: OpDelete,
			Begin: Position{
				Line: 0,
				Cell: 0,
			},
			Number: 3,
		}
		m8, _ = apply(m7, change)
		eq(t,
			m8 != nil, true,
			m8.ID > 0, true,
			m8.Previous == m7, true,
			m8.Change == change, true,
			m8.NumLines(), 5,

			m8.GetLine(0).content, "世界\n",
			m8.GetLine(1).content, "\n",
			m8.GetLine(2).content, "fquuxoo\n",
			m8.GetLine(3).content, "foofoo\n",
			m8.GetLine(4).content, "bar\n",
		)

		var m9 *Moment
		change = Change{
			Op: OpDelete,
			Begin: Position{
				Line: 0,
				Cell: 0,
			},
			Number: 9999,
		}
		m9, _ = apply(m8, change)
		eq(t,
			m9 != nil, true,
			m9.ID > 0, true,
			m9.Previous == m8, true,
			m9.Change == change, true,
			m9.NumLines(), 1,
		)

	})
}

func BenchmarkLargeBuf(b *testing.B) {
	withEditor(func(
		scope Scope,
		newMoment NewMomentFromBytes,
		apply ApplyChange,
	) {

		buf := new(bytes.Buffer)
		for i := 0; i < 1000000; i++ {
			buf.Write(bytes.Repeat([]byte("a"), 4096))
			buf.Write([]byte("\n"))
		}
		bs := buf.Bytes()
		moment, _, err := newMoment(bs)
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			change := Change{
				Op: OpInsert,
				Begin: Position{
					Line: 50000,
					Cell: 1024,
				},
				String: "foo",
			}
			apply(moment, change)
		}

	})
}

func TestDelete(t *testing.T) {
	withEditorBytes(t, []byte("a\nb"), func(
		scope Scope,
		view *View,
		apply ApplyChange,
	) {

		moment := view.GetMoment()
		change := Change{
			Op: OpDelete,
			Begin: Position{
				Line: 0,
				Cell: 0,
			},
			Number: 1,
		}
		var m2 *Moment
		m2, _ = apply(moment, change)
		eq(t,
			m2 != nil, true,
			m2.ID > 0, true,
			m2.Previous == moment, true,
			m2.Change == change, true,
			m2.NumLines(), 2,
			m2.GetLine(0).content, "\n",
			m2.GetLine(1).content, "b",
		)

		change = Change{
			Op: OpDelete,
			Begin: Position{
				Line: 0,
				Cell: 0,
			},
			Number: 2,
		}
		var m3 *Moment
		m3, _ = apply(m2, change)
		eq(t,
			m3 != nil, true,
			m3.ID > 0, true,
			m3.Previous == m2, true,
			m3.Change == change, true,
			m3.NumLines(), 1,
		)

	})
}

func TestDeleteSelection(t *testing.T) {
	withEditorBytes(t, []byte("foo"), func(
		scope Scope,
		toggle ToggleSelection,
		lineEnd LineEnd,
		del Delete,
	) {
		toggle()
		lineEnd()
		del()

		// select and delete empty line
		toggle()
		lineEnd()
		del()
	})
}

func TestNoViewEditCommands(t *testing.T) {
	withEditor(func(
		scope Scope,
		insert InsertAtPositionFunc,
		del DeletePrevRune,
		posCursor PosCursor,
	) {
		insert("foo", PositionFunc(posCursor))
		del()
	})
}

func TestInsertToEmptyBuffer(t *testing.T) {
	withEditorBytes(t, []byte(``), func(
		view *View,
		moment *Moment,
		insert InsertAtPositionFunc,
		posCursor PosCursor,
	) {
		insert("a", PositionFunc(posCursor))
		eq(t,
			view.CursorCol, 1,
		)
	})
}

func TestEditLineOverflow(t *testing.T) {
	withHelloEditor(t, func(
		view *View,
		scope Scope,
		apply ApplyChange,
	) {
		moment := view.GetMoment()
		apply(moment, Change{
			Begin: Position{
				Line: 9999,
			},
			Op:     OpDelete,
			Number: 9,
		})
		eq(t,
			view.GetMoment() == moment, true,
		)
	})
}

func TestEditRuneOffsetOverflow(t *testing.T) {
	withHelloEditor(t, func(
		view *View,
		scope Scope,
		apply ApplyChange,
	) {
		moment := view.GetMoment()
		apply(moment, Change{
			Begin: Position{
				Line: 0,
				Cell: 99999,
			},
			Op:     OpInsert,
			String: "foo",
		})
		eq(t,
			view.GetMoment() == moment, true,
		)
	})
}

func TestDeleteRuneOffsetOverflow(t *testing.T) {
	withHelloEditor(t, func(
		view *View,
		scope Scope,
		apply ApplyChange,
	) {
		moment := view.GetMoment()
		apply(moment, Change{
			Begin: Position{
				Line: 0,
				Cell: 99999,
			},
			Op:     OpDelete,
			Number: 9,
		})
		eq(t,
			view.GetMoment() == moment, true,
		)
	})
}

func TestDeletePrevRune(t *testing.T) {
	withEditorBytes(t, []byte("foo"), func(
		scope Scope,
		view *View,
		move MoveCursor,
		del DeletePrevRune,
	) {
		del()
		eq(t,
			view.GetMoment().NumLines(), 1,
		)

		move(Move{RelRune: 3})
		del()
		eq(t,
			view.GetMoment().NumLines(), 1,
			view.GetMoment().GetLine(0).content, "fo\n",
		)

		move(Move{RelLine: 0, RelRune: 2})
		scope.Call(NamedCommands["InsertNewline"].Func)
		eq(t,
			view.GetMoment().NumLines(), 2,
			view.GetMoment().GetLine(0).content, "fo\n",
			view.GetMoment().GetLine(1).content, "\n",
		)
	})
}

func TestInsertTab(t *testing.T) {
	withHelloEditor(t, func(
		view *View,
		scope Scope,
	) {
		scope.Call(NamedCommands["InsertTab"].Func)
		eq(t,
			view.GetMoment().GetLine(0).content, "\tHello, world!\n",
		)
	})
}

func TestCallDeleteRuneNoView(t *testing.T) {
	withEditor(func(del DeleteRune) {
		del()
	})
}

func TestCallDeleteNoView(t *testing.T) {
	withEditor(func(del Delete) {
		del()
	})
}

func TestDeleteMultiline(t *testing.T) {
	withHelloEditor(t, func(
		view *View,
		scope Scope,
		move MoveCursor,
		toggle ToggleSelection,
		del Delete,
	) {
		toggle()
		move(Move{RelLine: 1})
		move(Move{RelRune: 2})
		del()
		eq(t,
			view.GetMoment().NumLines(), 2,
			view.GetMoment().GetLine(0).content, "世界！\n",
		)
	})
}

func TestDeleteMultiline2(t *testing.T) {
	withHelloEditor(t, func(
		view *View,
		scope Scope,
		move MoveCursor,
		toggle ToggleSelection,
		del Delete,
	) {
		toggle()
		move(Move{RelLine: 2})
		move(Move{RelRune: 2})
		del()
		eq(t,
			view.GetMoment().NumLines(), 1,
			view.GetMoment().GetLine(0).content, "ちは、世界！\n",
		)
	})
}

func TestChangeText(t *testing.T) {
	withHelloEditor(t, func(
		view *View,
		scope Scope,
		curModes CurrentModes,
		move MoveCursor,
		toggle ToggleSelection,
		chagne ChangeText,
	) {
		toggle()
		move(Move{RelLine: 2})
		move(Move{RelRune: 2})
		chagne()
		eq(t,
			view.GetMoment().NumLines(), 1,
			view.GetMoment().GetLine(0).content, "ちは、世界！\n",
		)
		modes := curModes()
		_, ok := modes[0].(*EditMode)
		eq(t,
			ok, true,
		)
	})
}

func TestAppend(t *testing.T) {
	withHelloEditor(t, func(
		view *View,
		scope Scope,
		curModes CurrentModes,
	) {
		scope.Call(NamedCommands["Append"].Func)
		modes := curModes()
		_, ok := modes[0].(*EditMode)
		eq(t,
			ok, true,
			view.CursorCol, 1,
		)
	})
}

func TestEditNewLineBelowAndAbove(t *testing.T) {
	withHelloEditor(t, func(
		scope Scope,
		view *View,
	) {
		scope.Call(NamedCommands["EditNewLineBelow"].Func)
		eq(t,
			view.GetMoment().NumLines(), 4,
			view.GetMoment().GetLine(1).content, "\n",
		)
		scope.Call(NamedCommands["EditNewLineAbove"].Func)
		eq(t,
			view.GetMoment().NumLines(), 5,
			view.GetMoment().GetLine(1).content, "\n",
			view.GetMoment().GetLine(2).content, "\n",
		)
	})
}

func TestChangeToWordEnd(t *testing.T) {
	withHelloEditor(t, func(
		scope Scope,
		view *View,
		move MoveCursor,
		change ChangeToWordEnd,
	) {
		change()
		eq(t,
			view.GetMoment().GetLine(0).content, ", world!\n",
		)
		change()
		eq(t,
			view.GetMoment().GetLine(0).content, " world!\n",
		)
		change()
		eq(t,
			view.GetMoment().GetLine(0).content, "world!\n",
		)
		change()
		eq(t,
			view.GetMoment().GetLine(0).content, "!\n",
		)
		change()
		eq(t,
			view.GetMoment().GetLine(0).content, "\n",
		)
		change()
		eq(t,
			view.GetMoment().GetLine(0).content, "\n",
		)

		move(Move{RelLine: 1})
		change()
		eq(t,
			view.GetMoment().GetLine(1).content, "，世界！\n",
		)
		move(Move{RelRune: 2})
		change()
		eq(t,
			view.GetMoment().GetLine(1).content, "，世！\n",
		)
	})
}

func TestReplace(t *testing.T) {
	withEditorBytes(t, []byte("a\nb"), func(
		scope Scope,
		view *View,
		moment *Moment,
		replace ReplaceWithinRange,
	) {

		moment = replace(Range{
			Position{0, 0},
			Position{0, 0},
		}, "foo")
		eq(t,
			moment.NumLines(), 2,
			moment.GetLine(0).content, "fooa\n",
		)

		moment = replace(Range{
			Position{0, 0},
			Position{0, 1},
		}, "foo")
		eq(t,
			moment.NumLines(), 2,
			moment.GetLine(0).content, "fooooa\n",
		)

		moment = replace(Range{
			Position{0, 0},
			Position{1, 0},
		}, "foo")
		eq(t,
			moment.NumLines(), 1,
			moment.GetLine(0).content, "foob\n",
		)

	})
}

func TestDeleteLine(t *testing.T) {
	withHelloEditor(t, func(
		scope Scope,
		view *View,
		del DeleteLine,
	) {
		del()
		eq(t,
			view.GetMoment().NumLines(), 2,
		)
		del()
		eq(t,
			view.GetMoment().NumLines(), 1,
		)
		del()
		eq(t,
			view.GetMoment().NumLines(), 1,
		)
		del()
		eq(t,
			view.GetMoment().NumLines(), 1,
		)
	})
}
