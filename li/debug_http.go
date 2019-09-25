package li

import (
	"net/http"
	_ "net/http/pprof"
)

func init() {
	go func() {
		ce(http.ListenAndServe(":58764", nil))
	}()
}
