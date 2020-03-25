package li

import (
	_ "net/http/pprof"
)

func init() {
	go func() {
		//ce(http.ListenAndServe(":58764", nil))
	}()
}
