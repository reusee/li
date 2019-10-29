package li

import (
	"strings"
	"testing"
	"time"
)

func TestCompletionList(t *testing.T) {
	withHelloEditor(t, func(
		scope Scope,
		emitRune EmitRune,
		view *View,
		getScreenString GetScreenString,
		ctrl func(string),
		config CompletionConfig,
	) {
		scope.Call(EnableEditMode)
		emitRune('w')
		time.Sleep(time.Millisecond * time.Duration(config.DelayMilliseconds))
		ctrl("loop")
		lines := getScreenString(view.ContentBox)
		// ensure popup showed
		eq(t,
			strings.HasPrefix(lines[1], " wHello "), true,
			strings.HasPrefix(lines[2], " world  "), true,
		)
	})
}
