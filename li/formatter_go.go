package li

import (
	"bytes"
	"go/format"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func (_ Provide) FormatterGo(
	on On,
	j AppendJournal,
	scope Scope,
	run RunInMainLoop,
	config FormatterConfig,
) Init2 {

	type Job struct {
		view   *View
		buffer *Buffer
		moment *Moment
	}

	c := make(chan Job, 512)

	for i := 0; i < numCPU; i++ {
		go func() {
			for {
				job := <-c

				src := job.moment.GetBytes()
				formatted, err := formatGoSource(job.buffer.AbsPath, src)
				if err != nil {
					// do nothing if format error
					continue
				}
				if bytes.Equal(src, formatted) {
					continue
				}

				moment := job.view.GetMoment()
				if moment != job.moment {
					// moment switched, skip
					continue
				}

				// calculate diffs
				h := diffmatchpatch.New()
				diffs := h.DiffMain(string(src), string(formatted), false)
				if len(diffs) == 0 {
					continue
				}

				// delay update
				time.AfterFunc(time.Second*time.Duration(config.DelaySeconds), func() {
					run(func(
						scope Scope,
					) {
						if job.view.GetMoment() != job.moment {
							return
						}

						// apply changes
						t0 := time.Now()
						offset := 0
					loop_diffs:
						for _, diff := range diffs {

							position := moment.ByteOffsetToPosition(scope, offset)
							var numRunesInserted int

							switch diff.Type {

							case diffmatchpatch.DiffDelete:
								scope.Sub(func() (*Moment, Change) {
									change := Change{
										Op:    OpDelete,
										Begin: position,
										End:   moment.ByteOffsetToPosition(scope, offset+len(diff.Text)),
									}
									return moment, change
								}).Call(ApplyChange, &moment, &numRunesInserted)

							case diffmatchpatch.DiffInsert:
								scope.Sub(func() (*Moment, Change) {
									change := Change{
										Op:     OpInsert,
										String: diff.Text,
										Begin:  position,
									}
									return moment, change
								}).Call(ApplyChange, &moment, &numRunesInserted)
								offset += len(diff.Text)

							case diffmatchpatch.DiffEqual:
								offset += len(diff.Text)
								continue loop_diffs
							}

							job.view.switchMoment(scope, moment)
							scope.Sub(func() Move {
								col := moment.GetLine(scope, position.Line).Cells[position.Cell].DisplayOffset
								return Move{AbsLine: intP(position.Line), AbsCol: &col}
							}).Call(MoveCursor)
							scope.Sub(func() Move {
								return Move{RelRune: numRunesInserted}
							}).Call(MoveCursor)

						}

						if !bytes.Equal(moment.GetBytes(), formatted) {
							j("formatter bug, moment content not match")
							return
						}
						j("auto format in %v", time.Since(t0))

					})
				})

			}
		}()
	}

	// formatter for go
	on(EvMomentSwitched, func(
		buffer *Buffer,
		moments [2]*Moment,
		view *View,
	) {
		if buffer.language != LanguageGo {
			return
		}
		c <- Job{
			view:   view,
			buffer: buffer,
			moment: moments[1],
		}
	})

	return nil
}

var formatGoSource = func() (
	fn func(
		path string,
		bs []byte,
	) ([]byte, error),
) {

	goimportsPath, err := exec.LookPath("goimports")
	if err != nil {
		return func(_ string, bs []byte) ([]byte, error) {
			return format.Source(bs)
		}
	}

	return func(path string, bs []byte) ([]byte, error) {
		cmd := exec.Command(
			goimportsPath,
			"-srcdir", filepath.Dir(path),
		)
		cmd.Stdin = bytes.NewReader(bs)
		return cmd.Output()
	}

}()
