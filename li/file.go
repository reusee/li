package li

import (
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func (_ Command) ChoosePathAndLoad() (spec CommandSpec) {
	spec.Desc = "load file or dir"
	spec.Func = func(scope Scope) (NoResetN, NoLogImitation) {
		scope.Sub(func() func(string) {
			return func(path string) {
				var buffers []*Buffer
				var err error
				scope.Sub(func() string {
					return path
				}).Call(NewBuffersFromPath, &buffers, &err)
				if err != nil {
					return
				}
				for _, buffer := range buffers {
					scope.Sub(func() *Buffer {
						return buffer
					}).Call(NewViewFromBuffer)
				}
			}
		}).Call(ShowFileChooser)
		return true, true
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

func SyncBufferMomentToFile(
	buffer *Buffer,
	moment *Moment,
	linkedAll LinkedAll,
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
		err = me(nil, "buffer moment is not loaded from current disk file")
		return
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

func SyncViewToFile(
	cur CurrentView,
	scope Scope,
) (err error) {
	view := cur()
	if view == nil {
		return
	}
	scope.Sub(func() (*Buffer, *Moment) {
		return view.Buffer, view.Moment
	}).Call(SyncBufferMomentToFile, &err)
	if err != nil {
		scope.Sub(func() []string {
			return strings.Split(err.Error(), "\n")
		}).Call(ShowMessage)
	}
	return
}

func (_ Command) SyncViewToFile() (spec CommandSpec) {
	spec.Desc = "sync current view buffer moment to disk file"
	spec.Func = SyncViewToFile
	return
}
