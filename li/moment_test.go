package li

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestMomentFromBytes(t *testing.T) {
	withEditorBytes(t, []byte("abc"), func(
		view *View,
		scope Scope,
	) {
		eq(t,
			view.Moment.NumLines(), 1,
		)
		scope.Sub(func() Move {
			return Move{AbsLine: intP(999)}
		}).Call(MoveCursor)
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
		str := view.Moment.GetLowerContent()
		eq(t,
			str, "hello, world!\n你好，世界！\nこんにちは、世界！\n",
		)
	})
}

func TestSubContentHashing(t *testing.T) {
	withHelloEditor(t, func(
		moment *Moment,
	) {
		eq(t,
			fmt.Sprintf("%x", moment.GetSubContentHash(0)),
			fmt.Sprintf("%x", sha256.Sum256([]byte("Hello, world!\n"))),
			fmt.Sprintf("%x", moment.GetSubContentHash(2)),
			fmt.Sprintf("%x", sha256.Sum256([]byte("Hello, world!\n你好，世界！\nこんにちは、世界！\n"))),
			fmt.Sprintf("%x", moment.GetSubContentHash(1)),
			fmt.Sprintf("%x", sha256.Sum256([]byte("Hello, world!\n你好，世界！\n"))),
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
	) {
		moment.Language = LanguageGo
		parser := moment.GetParser()
		eq(t,
			parser != nil, true,
		)
		scope.Call(ChangeToWordEnd)
		parser = view.Moment.GetParser()
		eq(t,
			parser != nil, true,
		)
	})
}
