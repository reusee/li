package li

import (
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func (_ Command) ChoosePathAndLoad() (spec CommandSpec) {
	spec.Desc = "load file or dir"
	spec.Func = func(
		scope Scope,
		newView NewViewFromBuffer,
		choose ShowFileChooser,
		newBuffers NewBuffersFromPath,
	) {
		choose(func(path string) {
			buffers, err := newBuffers(path)
			if err != nil {
				return
			}
			for _, buffer := range buffers {
				newView(buffer)
			}
		})
	}
	return
}

type FileInfo struct {
	Name    string
	Size    int64
	ModTime time.Time
	IsDir   bool
}

func getFileInfo(path string) (info FileInfo, err error) {
	osInfo, err := os.Stat(path)
	if err != nil {
		return
	}
	info.Name = osInfo.Name()
	info.Size = osInfo.Size()
	info.ModTime = osInfo.ModTime()
	info.IsDir = osInfo.IsDir()
	return
}

type SyncBufferMomentToFile func(
	buffer *Buffer,
	moment *Moment,
) (
	err error,
)

func (_ Provide) SyncBufferMomentToFile(
	linkedAll LinkedAll,
) SyncBufferMomentToFile {
	return func(
		buffer *Buffer,
		moment *Moment,
	) (
		err error,
	) {

		// get disk file info
		diskFileInfo, err := getFileInfo(buffer.Path)
		if err != nil {
			return err
		}

		// check whether moment is loaded from current disk file
		ok := false
		var moments []*Moment
		linkedAll(buffer, &moments)
		for _, m := range moments {
			if m.FileInfo == diskFileInfo {
				ok = true
				break
			}
		}
		if !ok {
			return we(fe("buffer moment is not loaded from current disk file"))
		}

		// save
		err = ioutil.WriteFile(buffer.Path, []byte(moment.GetContent()), 0644)
		if err != nil {
			return
		}

		// update file info
		diskFileInfo, err = getFileInfo(buffer.Path)
		if err != nil {
			return err
		}
		moment.FileInfo = diskFileInfo
		buffer.LastSyncFileInfo = diskFileInfo

		return
	}
}

func SyncViewToFile(
	cur CurrentView,
	sync SyncBufferMomentToFile,
	show ShowMessage,
) (err error) {
	view := cur()
	if view == nil {
		return
	}
	moment := view.GetMoment()
	err = sync(view.Buffer, moment)
	if err != nil {
		msg := strings.Split(err.Error(), "\n")
		show(msg)
	}
	return
}

func (_ Command) SyncViewToFile() (spec CommandSpec) {
	spec.Desc = "sync current view buffer moment to disk file"
	spec.Func = SyncViewToFile
	return
}
