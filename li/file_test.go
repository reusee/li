package li

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFileSync(t *testing.T) {
	withEditor(func(
		scope Scope,
		linkedOne LinkedOne,
	) {

		// temp file
		f, err := ioutil.TempFile("", "*")
		ce(err)
		defer os.Remove(f.Name())
		path := f.Name()
		_, err = f.Write([]byte("foobar"))
		ce(err)
		ce(f.Close())

		// buffer
		var buffer *Buffer
		scope.Sub(func() string {
			return path
		}).Call(NewBufferFromFile, &buffer, &err)
		ce(err)

		// moment
		var moment *Moment
		linkedOne(buffer, &moment)
		eq(t,
			moment != nil, true,
		)

		// view
		var view *View
		scope.Sub(func() *Buffer {
			return buffer
		}).Call(NewViewFromBuffer, &view)
		eq(t,
			view != nil, true,
			view.Buffer == buffer, true,
			view.Moment == moment, true,
		)

		// sync
		scope.Call(DeleteRune)
		scope.Call(SyncViewToFile, &err)
		ce(err)

		// undo then sync
		scope.Call(Undo)
		scope.Call(SyncViewToFile, &err)
		ce(err)

	})
}
