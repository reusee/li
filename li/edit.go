package li

import (
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Op uint8

const (
	OpInsert Op = iota
	OpDelete
)

type Change struct {
	Op     Op
	String string   // for Insert
	Begin  Position // for Insert or Delete
	// for Delete operation, one of End and Number must be set
	End    Position // for Delete
	Number int      // for Delete
}

func ApplyChange(
	moment *Moment,
	change Change,
	config BufferConfig,
	link Link,
	linkedOne LinkedOne,
) (
	newMoment *Moment,
	numRunesInserted int,
) {

	id := MomentID(atomic.AddInt64(&nextMomentID, 1))
	newMoment = NewMoment()
	newMoment.T0 = time.Now()
	newMoment.ID = id
	newMoment.Previous = moment
	newMoment.Change = change

	// validate begin position
	if change.Begin.Line >= moment.NumLines() {
		// bad line
		return
	}
	if line := moment.GetLine(change.Begin.Line); change.Begin.Rune > len(line.Cells) {
		// bad rune offset
		return
	}

	newMoment.subContentHashes = make([]*HashSum, 0, len(moment.subContentHashes))
	newMoment.subContentHashStates = make([]*[]byte, 0, len(moment.subContentHashStates))

	var newLines []*Line

	switch change.Op {

	case OpInsert:
		newLines = append(newLines, moment.lines[:change.Begin.Line]...)
		newMoment.subContentHashStates = append(
			newMoment.subContentHashStates,
			moment.subContentHashStates[:change.Begin.Line]...,
		)
		newMoment.subContentHashes = append(
			newMoment.subContentHashes,
			moment.subContentHashes[:change.Begin.Line]...,
		)
		line := moment.GetLine(change.Begin.Line)
		offset := 0
		for _, cell := range line.Cells[:change.Begin.Rune] {
			offset += cell.RuneLen
		}
		content := line.content[:offset] + change.String + line.content[offset:]
		numRunesInserted += len([]rune(change.String))
		changingLastLine := change.Begin.Line == moment.NumLines()-1
		lines := splitLines(content)
		for i, content := range lines {
			if changingLastLine && i == len(lines)-1 {
				// add newline to the last line
				if !strings.HasSuffix(content, "\n") {
					content += "\n"
					numRunesInserted++
				}
			}
			newLines = append(newLines, &Line{
				content:  content,
				initOnce: new(sync.Once),
				config:   &config,
			})
		}
		newLines = append(newLines, moment.lines[change.Begin.Line+1:]...)

	case OpDelete:
		// resolve change.Number
		if change.Number > 0 {
			change.End = change.Begin
			// iterate
			for change.Number > 0 {
				line := moment.GetLine(change.End.Line)
				if line == nil {
					change.Number = 0
					change.End.Line--
					change.End.Rune = len(moment.GetLine(change.End.Line).Cells) - 1
				} else {
					if change.End.Rune+change.Number >= len(line.Cells) {
						// next line
						change.Number -= len(line.Cells) - change.End.Rune
						change.End.Rune = 0
						change.End.Line++
					} else {
						change.End.Rune += change.Number
						change.Number = 0
					}
				}
			}
		}

		if change.Begin == change.End {
			newMoment = moment
			return
		}

		// assemble new lines
		newLines = append(newLines, moment.lines[:change.Begin.Line]...)
		newMoment.subContentHashStates = append(
			newMoment.subContentHashStates,
			moment.subContentHashStates[:change.Begin.Line]...,
		)
		newMoment.subContentHashes = append(
			newMoment.subContentHashes,
			moment.subContentHashes[:change.Begin.Line]...,
		)
		var b strings.Builder
		for lineNum := change.Begin.Line; lineNum <= change.End.Line; lineNum++ {
			if lineNum >= moment.NumLines() {
				break
			}
			if lineNum == change.Begin.Line {
				for _, cell := range moment.GetLine(lineNum).Cells {
					if cell.RuneOffset >= change.Begin.Rune {
						break
					}
					b.WriteRune(cell.Rune)
				}
			}
			if lineNum == change.End.Line {
				for _, cell := range moment.GetLine(lineNum).Cells {
					if cell.RuneOffset < change.End.Rune {
						continue
					}
					b.WriteRune(cell.Rune)
				}
			}
		}
		changingLastLine := change.End.Line >= moment.NumLines()-1
		lines := splitLines(b.String())
		for i, content := range lines {
			if changingLastLine && i == len(lines)-1 {
				// add newline to the last line
				if !strings.HasSuffix(content, "\n") {
					content += "\n"
				}
			}
			newLines = append(newLines, &Line{
				content:  content,
				initOnce: new(sync.Once),
				config:   &config,
			})
		}
		res := change.End.Line + 1
		if res < moment.NumLines() {
			newLines = append(newLines, moment.lines[res:]...)
		}

	}

	newMoment.lines = newLines
	newMoment.subContentHashStates = append(
		newMoment.subContentHashStates,
		make(
			[]*[]byte,
			len(newMoment.lines)-len(newMoment.subContentHashStates),
		)...,
	)
	newMoment.subContentHashes = append(
		newMoment.subContentHashes,
		make(
			[]*HashSum,
			len(newMoment.lines)-len(newMoment.subContentHashes),
		)...,
	)
	var buffer *Buffer
	linkedOne(moment, &buffer)
	if buffer != nil {
		link(buffer, newMoment)
	}

	return
}

func InsertAtPositionFunc(
	fn PositionFunc,
	str string,
	getCur CurrentView,
	scope Scope,
) {

	view := getCur()
	if view == nil {
		return
	}

	var position Position
	scope.Call(fn, &position)
	change := Change{
		Op:     OpInsert,
		Begin:  position,
		String: str,
	}
	var newMoment *Moment
	var nRunesInserted int
	scope.Sub(func() (*Moment, Change) {
		return view.Moment, change
	}).Call(ApplyChange, &newMoment, &nRunesInserted)

	view.switchMoment(newMoment)

	scope.Sub(func() Move {
		return Move{AbsLine: intP(position.Line), AbsCol: intP(position.Col)}
	}).Call(MoveCursor)
	scope.Sub(func() Move {
		return Move{RelRune: nRunesInserted}
	}).Call(MoveCursor)

}

func DeleteWithinRange(
	r Range,
	cur CurrentView,
	scope Scope,
) {
	view := cur()
	if view == nil {
		return
	}
	change := Change{
		Op:    OpDelete,
		Begin: r.Begin,
		End:   r.End,
	}
	var newMoment *Moment
	scope.Sub(func() (*Moment, Change) {
		return view.Moment, change
	}).Call(ApplyChange, &newMoment)
	view.switchMoment(newMoment)
	scope.Sub(func() Move {
		return Move{AbsLine: intP(r.Begin.Line), AbsCol: intP(r.Begin.Col)}
	}).Call(MoveCursor)
}

func DeleteWithinPositionFuncs(
	fns [2]PositionFunc,
	scope Scope,
	cur CurrentView,
) {
	view := cur()
	if view == nil {
		return
	}
	var begin Position
	scope.Call(fns[0], &begin)
	var end Position
	scope.Call(fns[1], &end)
	scope.Sub(func() Range {
		return Range{
			Begin: begin,
			End:   end,
		}
	}).Call(DeleteWithinRange)
}

func DeletePrevRune(
	scope Scope,
) {
	scope.Sub(func() [2]PositionFunc {
		return [2]PositionFunc{
			PosPrevRune,
			PosCursor,
		}
	}).Call(DeleteWithinPositionFuncs)
}

func DeleteRune(
	scope Scope,
) {
	scope.Sub(func() [2]PositionFunc {
		return [2]PositionFunc{
			PosCursor,
			PosNextRune,
		}
	}).Call(DeleteWithinPositionFuncs)
}

func _Delete(
	cur CurrentView,
	scope Scope,
	afterFunc AfterFunc,
) {
	view := cur()
	if view == nil {
		return
	}

	// delete selected
	if r := view.selectedRange(); r != nil {
		scope.Sub(func() Range {
			return *r
		}).Call(DeleteWithinRange)
		view.SelectionAnchor = nil
	}

	if afterFunc != nil {
		scope.Call(afterFunc)
	}

}

func Delete(
	scope Scope,
) {
	scope.Sub(func() AfterFunc {
		return func() {}
	}).Call(_Delete)
}

func ChangeText(
	scope Scope,
) {
	scope.Sub(func() AfterFunc {
		return func(scope Scope) {
			scope.Call(EnableEditMode)
		}
	}).Call(_Delete)
}
