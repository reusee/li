package li

import (
	"bytes"
	"testing"
)

func TestApplyChange(t *testing.T) {
	withEditor(func(
		scope Scope,
	) {
		var moment *Moment
		scope.Sub(func() []byte {
			return []byte(``)
		}).Call(NewMomentFromBytes, &moment)
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
		scope.Sub(func() (*Moment, Change) {
			return moment, change
		}).Call(ApplyChange, &m2)
		eq(t,
			m2 != nil, true,
			m2.ID > 0, true,
			m2.Previous == moment, true,
			m2.Change == change, true,
			m2.NumLines(), 1,
			m2.GetLine(scope, 0).content, "foo\n",
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
		scope.Sub(func() (*Moment, Change) {
			return m2, change
		}).Call(ApplyChange, &m3)
		eq(t,
			m3 != nil, true,
			m3.ID > 0, true,
			m3.Previous == m2, true,
			m3.Change == change, true,
			m3.NumLines(), 3,
			m3.GetLine(scope, 0).content, "foo\n",
			m3.GetLine(scope, 1).content, "foo\n",
			m3.GetLine(scope, 2).content, "foofoo\n",
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
		scope.Sub(func() (*Moment, Change) {
			return m3, change
		}).Call(ApplyChange, &m4)
		eq(t,
			m4 != nil, true,
			m4.ID > 0, true,
			m4.Previous == m3, true,
			m4.Change == change, true,
			m4.NumLines(), 4,

			m4.GetLine(scope, 0).content, "foo\n",
			m4.GetLine(scope, 1).content, "foo\n",
			m4.GetLine(scope, 2).content, "foofoo\n",
			m4.GetLine(scope, 3).content, "bar\n",
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
		scope.Sub(func() (*Moment, Change) {
			return m4, change
		}).Call(ApplyChange, &m5)
		eq(t,
			m5 != nil, true,
			m5.ID > 0, true,
			m5.Previous == m4, true,
			m5.Change == change, true,
			m5.NumLines(), 4,

			m5.GetLine(scope, 0).content, "foo\n",
			m5.GetLine(scope, 1).content, "fquuxoo\n",
			m5.GetLine(scope, 2).content, "foofoo\n",
			m5.GetLine(scope, 3).content, "bar\n",
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
		scope.Sub(func() (*Moment, Change) {
			return m5, change
		}).Call(ApplyChange, &m6)
		eq(t,
			m6 != nil, true,
			m6.ID > 0, true,
			m6.Previous == m5, true,
			m6.Change == change, true,
			m6.NumLines(), 6,

			m6.GetLine(scope, 0).content, "foo你好\n",
			m6.GetLine(scope, 1).content, "世界\n",
			m6.GetLine(scope, 2).content, "\n",
			m6.GetLine(scope, 3).content, "fquuxoo\n",
			m6.GetLine(scope, 4).content, "foofoo\n",
			m6.GetLine(scope, 5).content, "bar\n",
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
		scope.Sub(func() (*Moment, Change) {
			return m6, change
		}).Call(ApplyChange, &m7)
		eq(t,
			m7 != nil, true,
			m7.ID > 0, true,
			m7.Previous == m6, true,
			m7.Change == change, true,
			m7.NumLines(), 6,

			m7.GetLine(scope, 0).content, "你好\n",
			m7.GetLine(scope, 1).content, "世界\n",
			m7.GetLine(scope, 2).content, "\n",
			m7.GetLine(scope, 3).content, "fquuxoo\n",
			m7.GetLine(scope, 4).content, "foofoo\n",

			m7.GetLine(scope, 5).content, "bar\n",
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
		scope.Sub(func() (*Moment, Change) {
			return m7, change
		}).Call(ApplyChange, &m8)
		eq(t,
			m8 != nil, true,
			m8.ID > 0, true,
			m8.Previous == m7, true,
			m8.Change == change, true,
			m8.NumLines(), 5,

			m8.GetLine(scope, 0).content, "世界\n",
			m8.GetLine(scope, 1).content, "\n",
			m8.GetLine(scope, 2).content, "fquuxoo\n",
			m8.GetLine(scope, 3).content, "foofoo\n",
			m8.GetLine(scope, 4).content, "bar\n",
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
		scope.Sub(func() (*Moment, Change) {
			return m8, change
		}).Call(ApplyChange, &m9)
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
	) {

		var moment *Moment
		scope.Sub(func() []byte {
			buf := new(bytes.Buffer)
			for i := 0; i < 1000000; i++ {
				buf.Write(bytes.Repeat([]byte("a"), 4096))
				buf.Write([]byte("\n"))
			}
			return buf.Bytes()
		}).Call(NewMomentFromBytes, &moment)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			var m9 *Moment
			change := Change{
				Op: OpInsert,
				Begin: Position{
					Line: 50000,
					Cell: 1024,
				},
				String: "foo",
			}
			scope.Sub(func() (*Moment, Change) {
				return moment, change
			}).Call(ApplyChange, &m9)
		}

	})
}

func TestDelete(t *testing.T) {
	withEditorBytes(t, []byte("a\nb"), func(
		scope Scope,
		view *View,
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
		scope.Sub(func() (*Moment, Change) {
			return moment, change
		}).Call(ApplyChange, &m2)
		eq(t,
			m2 != nil, true,
			m2.ID > 0, true,
			m2.Previous == moment, true,
			m2.Change == change, true,
			m2.NumLines(), 2,
			m2.GetLine(scope, 0).content, "\n",
			m2.GetLine(scope, 1).content, "b",
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
		scope.Sub(func() (*Moment, Change) {
			return m2, change
		}).Call(ApplyChange, &m3)
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
	) {
		scope.Call(ToggleSelection)
		scope.Call(LineEnd)
		scope.Call(Delete)

		// select and delete empty line
		scope.Call(ToggleSelection)
		scope.Call(LineEnd)
		scope.Call(Delete)
	})
}

func TestNoViewEditCommands(t *testing.T) {
	withEditor(func(
		scope Scope,
	) {
		scope.Sub(func() (PositionFunc, string) {
			return PosCursor, "foo"
		}).Call(InsertAtPositionFunc)
		scope.Call(DeletePrevRune)
	})
}

func TestInsertToEmptyBuffer(t *testing.T) {
	withEditorBytes(t, []byte(``), func(
		scope Scope,
		view *View,
		moment *Moment,
	) {
		scope.Sub(func() (PositionFunc, string) {
			return PosCursor, "a"
		}).Call(InsertAtPositionFunc)
		eq(t,
			view.CursorCol, 1,
		)
	})
}

func TestEditLineOverflow(t *testing.T) {
	withHelloEditor(t, func(
		view *View,
		scope Scope,
	) {
		moment := view.GetMoment()
		scope.Sub(func() Change {
			return Change{
				Begin: Position{
					Line: 9999,
				},
				Op:     OpDelete,
				Number: 9,
			}
		}).Call(ApplyChange)
		eq(t,
			view.GetMoment() == moment, true,
		)
	})
}

func TestEditRuneOffsetOverflow(t *testing.T) {
	withHelloEditor(t, func(
		view *View,
		scope Scope,
	) {
		moment := view.GetMoment()
		scope.Sub(func() Change {
			return Change{
				Begin: Position{
					Line: 0,
					Cell: 99999,
				},
				Op:     OpInsert,
				String: "foo",
			}
		}).Call(ApplyChange)
		eq(t,
			view.GetMoment() == moment, true,
		)
	})
}

func TestDeleteRuneOffsetOverflow(t *testing.T) {
	withHelloEditor(t, func(
		view *View,
		scope Scope,
	) {
		moment := view.GetMoment()
		scope.Sub(func() Change {
			return Change{
				Begin: Position{
					Line: 0,
					Cell: 99999,
				},
				Op:     OpDelete,
				Number: 9,
			}
		}).Call(ApplyChange)
		eq(t,
			view.GetMoment() == moment, true,
		)
	})
}

func TestDeletePrevRune(t *testing.T) {
	withEditorBytes(t, []byte("foo"), func(
		scope Scope,
		view *View,
	) {
		scope.Call(DeletePrevRune)
		eq(t,
			view.GetMoment().NumLines(), 1,
		)

		scope.Sub(&Move{RelRune: 3}).
			Call(MoveCursor)
		scope.Call(DeletePrevRune)
		eq(t,
			view.GetMoment().NumLines(), 1,
			view.GetMoment().GetLine(scope, 0).content, "fo\n",
		)

		scope.Sub(&Move{RelLine: 0, RelRune: 2}).
			Call(MoveCursor)
		scope.Call(NamedCommands["InsertNewline"].Func)
		eq(t,
			view.GetMoment().NumLines(), 2,
			view.GetMoment().GetLine(scope, 0).content, "fo\n",
			view.GetMoment().GetLine(scope, 1).content, "\n",
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
			view.GetMoment().GetLine(scope, 0).content, "\tHello, world!\n",
		)
	})
}

func TestCallDeleteRuneNoView(t *testing.T) {
	withEditor(func(scope Scope) {
		scope.Call(DeleteRune)
	})
}

func TestCallDeleteNoView(t *testing.T) {
	withEditor(func(scope Scope) {
		scope.Call(Delete)
	})
}

func TestDeleteMultiline(t *testing.T) {
	withHelloEditor(t, func(
		view *View,
		scope Scope,
	) {
		scope.Call(ToggleSelection)
		scope.Sub(&Move{RelLine: 1}).
			Call(MoveCursor)
		scope.Sub(&Move{RelRune: 2}).
			Call(MoveCursor)
		scope.Call(Delete)
		eq(t,
			view.GetMoment().NumLines(), 2,
			view.GetMoment().GetLine(scope, 0).content, "世界！\n",
		)
	})
}

func TestDeleteMultiline2(t *testing.T) {
	withHelloEditor(t, func(
		view *View,
		scope Scope,
	) {
		scope.Call(ToggleSelection)
		scope.Sub(&Move{RelLine: 2}).
			Call(MoveCursor)
		scope.Sub(&Move{RelRune: 2}).
			Call(MoveCursor)
		scope.Call(Delete)
		eq(t,
			view.GetMoment().NumLines(), 1,
			view.GetMoment().GetLine(scope, 0).content, "ちは、世界！\n",
		)
	})
}

func TestChangeText(t *testing.T) {
	withHelloEditor(t, func(
		view *View,
		scope Scope,
		curModes CurrentModes,
	) {
		scope.Call(ToggleSelection)
		scope.Sub(&Move{RelLine: 2}).
			Call(MoveCursor)
		scope.Sub(&Move{RelRune: 2}).
			Call(MoveCursor)
		scope.Call(ChangeText)
		eq(t,
			view.GetMoment().NumLines(), 1,
			view.GetMoment().GetLine(scope, 0).content, "ちは、世界！\n",
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
			view.GetMoment().GetLine(scope, 1).content, "\n",
		)
		scope.Call(NamedCommands["EditNewLineAbove"].Func)
		eq(t,
			view.GetMoment().NumLines(), 5,
			view.GetMoment().GetLine(scope, 1).content, "\n",
			view.GetMoment().GetLine(scope, 2).content, "\n",
		)
	})
}

func TestChangeToWordEnd(t *testing.T) {
	withHelloEditor(t, func(
		scope Scope,
		view *View,
	) {
		scope.Call(ChangeToWordEnd)
		eq(t,
			view.GetMoment().GetLine(scope, 0).content, ", world!\n",
		)
		scope.Call(ChangeToWordEnd)
		eq(t,
			view.GetMoment().GetLine(scope, 0).content, " world!\n",
		)
		scope.Call(ChangeToWordEnd)
		eq(t,
			view.GetMoment().GetLine(scope, 0).content, "world!\n",
		)
		scope.Call(ChangeToWordEnd)
		eq(t,
			view.GetMoment().GetLine(scope, 0).content, "!\n",
		)
		scope.Call(ChangeToWordEnd)
		eq(t,
			view.GetMoment().GetLine(scope, 0).content, "\n",
		)
		scope.Call(ChangeToWordEnd)
		eq(t,
			view.GetMoment().GetLine(scope, 0).content, "\n",
		)

		scope.Sub(&Move{RelLine: 1}).
			Call(MoveCursor)
		scope.Call(ChangeToWordEnd)
		eq(t,
			view.GetMoment().GetLine(scope, 1).content, "，世界！\n",
		)
		scope.Sub(&Move{RelRune: 2}).
			Call(MoveCursor)
		scope.Call(ChangeToWordEnd)
		eq(t,
			view.GetMoment().GetLine(scope, 1).content, "，世！\n",
		)
	})
}

func TestReplace(t *testing.T) {
	withEditorBytes(t, []byte("a\nb"), func(
		scope Scope,
		view *View,
		moment *Moment,
	) {

		scope.Sub(func() (Range, string) {
			return Range{
				Position{0, 0},
				Position{0, 0},
			}, "foo"
		}).Call(ReplaceWithinRange, &moment)
		eq(t,
			moment.NumLines(), 2,
			moment.GetLine(scope, 0).content, "fooa\n",
		)

		scope.Sub(func() (Range, string) {
			return Range{
				Position{0, 0},
				Position{0, 1},
			}, "foo"
		}).Call(ReplaceWithinRange, &moment)
		eq(t,
			moment.NumLines(), 2,
			moment.GetLine(scope, 0).content, "fooooa\n",
		)

		scope.Sub(func() (Range, string) {
			return Range{
				Position{0, 0},
				Position{1, 0},
			}, "foo"
		}).Call(ReplaceWithinRange, &moment)
		eq(t,
			moment.NumLines(), 1,
			moment.GetLine(scope, 0).content, "foob\n",
		)

	})
}

func TestDeleteLine(t *testing.T) {
	withHelloEditor(t, func(
		scope Scope,
		view *View,
	) {
		scope.Call(DeleteLine)
		eq(t,
			view.GetMoment().NumLines(), 2,
		)
		scope.Call(DeleteLine)
		eq(t,
			view.GetMoment().NumLines(), 1,
		)
		scope.Call(DeleteLine)
		eq(t,
			view.GetMoment().NumLines(), 1,
		)
		scope.Call(DeleteLine)
		eq(t,
			view.GetMoment().NumLines(), 1,
		)
	})
}
