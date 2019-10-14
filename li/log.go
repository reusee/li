package li

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var logFile = func() *os.File {
	configDir := getConfigDir()
	f, err := os.OpenFile(
		filepath.Join(configDir, "log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_SYNC,
		0644,
	)
	ce(err)
	return f
}()

var outputFileLock sync.Mutex

func log(format string, args ...any) {
	outputFileLock.Lock()
	defer outputFileLock.Unlock()
	fmt.Fprintf(logFile, "%s ", time.Now().Format("15:04:05.999"))
	fmt.Fprintf(logFile, format, args...)
}

var Log = log
