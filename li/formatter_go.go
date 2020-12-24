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
				src = bytes.TrimRight(src, "\n")
				formatted, err := formatGoSource(job.buffer.AbsPath, src)
				if err != nil {
					// do nothing if format error
					continue
				}
				formatted = bytes.TrimRight(formatted, "\n")
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
						moveCursor MoveCursor,
						apply ApplyChange,
					) {
						if job.view.GetMoment() != job.moment {
							return
						}

						// apply changes
						t0 := time.Now()
						offset := 0
					loop_diffs:
						for _, diff := range diffs {

							position := moment.ByteOffsetToPosition(offset)
							var numRunesInserted int

							switch diff.Type {

							case diffmatchpatch.DiffDelete:
								moment, numRunesInserted = apply(
									moment,
									Change{
										Op:    OpDelete,
										Begin: position,
										End:   moment.ByteOffsetToPosition(offset + len(diff.Text)),
									},
								)

							case diffmatchpatch.DiffInsert:
								moment, numRunesInserted = apply(
									moment,
									Change{
										Op:     OpInsert,
										String: diff.Text,
										Begin:  position,
									},
								)
								offset += len(diff.Text)

							case diffmatchpatch.DiffEqual:
								offset += len(diff.Text)
								continue loop_diffs
							}

							job.view.switchMoment(scope, moment)
							col := moment.GetLine(position.Line).Cells[position.Cell].DisplayOffset
							moveCursor(Move{AbsLine: intP(position.Line), AbsCol: &col})
							moveCursor(Move{RelRune: numRunesInserted})

						}

						content := moment.GetBytes()
						content = bytes.TrimRight(content, "\n")
						if !bytes.Equal(content, formatted) {
							j("%x", content[len(content)-10:])
							j("%x", formatted[len(formatted)-10:])
							j("formatter bug, moment content not match %d %d", len(content), len(formatted))
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
		curModes CurrentModes,
	) {
		if buffer.language != LanguageGo {
			return
		}
		if IsEditing(curModes()) {
			return
		}
		c <- Job{
			view:   view,
			buffer: buffer,
			moment: moments[1],
		}
	})

	on(EvModesChanged, func(
		v CurrentView,
		modes []Mode,
	) {
		view := v()
		if view == nil {
			return
		}
		buffer := view.Buffer
		if buffer.language != LanguageGo {
			return
		}
		if IsEditing(modes) {
			return
		}
		c <- Job{
			view:   view,
			buffer: buffer,
			moment: view.GetMoment(),
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
