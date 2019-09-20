package li

import (
	"strings"
	"unsafe"

	"github.com/reusee/li/treesitter"
)

type Language int

const (
	LanguageUnknown Language = iota
	LanguageGo
)

func LanguageFromPath(path string) Language {
	if strings.HasSuffix(strings.ToLower(path), ".go") {
		return LanguageGo
	}
	return LanguageUnknown
}

var languageParsers = map[Language]func(*Moment) *treesitter.Parser{
	LanguageGo: func(m *Moment) *treesitter.Parser {
		return treesitter.ParseGo(
			unsafe.Pointer(m.GetCStringContent()),
			len(m.GetContent()),
		)
	},
}

var languageStainers = map[Language]func() Stainer{
	LanguageGo: func() Stainer {
		return new(GoLexicalStainer)
	},
}
