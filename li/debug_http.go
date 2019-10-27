package li

import (
	"net/http"
	_ "net/http/pprof"
)

func init() {
	go func() {
		return //TODO
		ce(http.ListenAndServe(":58764", nil))
	}()
}
