package li

import "testing"

func TestLanguageServerProtocol(t *testing.T) {
	withEditorBytes(t, []byte(`
	  package main
	  func main() {
	  }
	`), func(
		buffer *Buffer,
		scope Scope,
	) {
		buffer.SetLanguage(scope, LanguageGo)
	})
}
