package li

import (
	"strings"
	"sync"

	textwidth "golang.org/x/text/width"
)

var runeWidths sync.Map

func runeDisplayWidth(r rune) int {
	if v, ok := runeWidths.Load(r); ok {
		return v.(int)
	}
	prop := textwidth.LookupRune(r)
	kind := prop.Kind()
	width := 1
	if kind == textwidth.EastAsianAmbiguous ||
		kind == textwidth.EastAsianWide ||
		kind == textwidth.EastAsianFullwidth {
		width = 2
	}
	runeWidths.Store(r, width)
	return width
}

func runesDisplayWidth(runes []rune) (l int) {
	for _, r := range runes {
		l += runeDisplayWidth(r)
	}
	return
}

func displayWidth(s string) (l int) {
	return runesDisplayWidth([]rune(s))
}

func rightPad(s string, pad rune, l int) string {
	padLen := l - displayWidth(s)
	return s + strings.Repeat(string(pad), padLen)
}
