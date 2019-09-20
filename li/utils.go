package li

/*
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/reusee/dscope"
	"github.com/reusee/e/v2"
	textwidth "golang.org/x/text/width"
)

var (
	me       = e.Default.WithStack().WithName("li")
	ce, he   = e.New(me)
	pt       = fmt.Printf
	t0       = time.Now()
	NewScope = dscope.New
	is       = errors.Is
)

type (
	Scope = dscope.Scope
	any   = interface{}
	dyn   = interface{}
)

var runeWidths sync.Map

func runeWidth(r rune) int {
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

func displayWidth(s string) (l int) {
	for _, r := range []rune(s) {
		l += runeWidth(r)
	}
	return
}

func runesDisplayWidth(runes []rune) (l int) {
	for _, r := range runes {
		l += runeWidth(r)
	}
	return
}

func leftPad(s string, pad rune, l int) string {
	padLen := l - displayWidth(s)
	return strings.Repeat(string(pad), padLen) + s
}

func rightPad(s string, pad rune, l int) string {
	padLen := l - displayWidth(s)
	return s + strings.Repeat(string(pad), padLen)
}

func split(i, n int) []int {
	base := i / n
	res := i - base*n
	var ret []int
	for i := 0; i < res; i++ {
		ret = append(ret, base+1)
	}
	for len(ret) < n {
		ret = append(ret, base)
	}
	return ret
}

func intP(i int) *int {
	return &i
}

func splitDir(path string) (ret []string) {
	dir, name := filepath.Split(path)
	if dir == "/" {
		return []string{name}
	}
	ret = append(splitDir(filepath.Clean(dir)), name)
	return ret
}

func cfree(p unsafe.Pointer) {
	C.free(p)
}
