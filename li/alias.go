package li

import (
	"fmt"

	"github.com/reusee/e4"
)

var (
	ce = e4.Check.With(e4.WrapStacktrace)
	he = e4.Handle
	we = e4.Wrap.With(e4.WrapStacktrace)
	pt = fmt.Printf
)
