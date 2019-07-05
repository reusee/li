package li

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func logStack() {
	skip := 2
	pcs := make([]uintptr, 128)
	skip += runtime.Callers(skip, pcs)
	log("--- stack start ---\n")
	for len(pcs) > 0 {
		frames := runtime.CallersFrames(pcs)
		frame, more := frames.Next()
		for more {
			if !strings.Contains(frame.File, "reusee/li") {
				continue
			}
			log("-> file %s, line %d\n", frame.File, frame.Line)
			frame, more = frames.Next()
		}
		pcs = pcs[0:0]
		skip += runtime.Callers(skip, pcs)
	}
	log("... stack end .....\n")
}
