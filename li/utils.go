package li

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"github.com/reusee/dscope"
)

var (
	is     = errors.Is
	numCPU = runtime.NumCPU()
	never  = time.Date(9102, 1, 1, 1, 1, 1, 1, time.Local)
)

type (
	Scope = dscope.Scope
	any   = interface{}
	dyn   = interface{}
	M     = map[string]any
)

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
	if path == "" {
		return
	}
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

func toJSON(o any) string {
	buf := new(strings.Builder)
	encoder := json.NewEncoder(buf)
	encoder.SetIndent("", "    ")
	ce(encoder.Encode(o))
	return buf.String()
}
