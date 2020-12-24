package li

import (
	"strings"
	"sync"
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

type ApplyChange func(
	moment *Moment,
	change Change,
) (
	newMoment *Moment,
	numRunesInserted int,
)

func (_ Provide) ApplyChange(
	config BufferConfig,
	link Link,
	linkedOne LinkedOne,
) ApplyChange {

	return func(
		moment *Moment,
		change Change,
	) (
		newMoment *Moment,
		numRunesInserted int,
	) {
		newMoment = NewMoment(moment)
		newMoment.Change = change

		// validate begin position
		if change.Begin.Line >= moment.NumLines() {
			// bad line
			return
		}
		if line := moment.GetLine(change.Begin.Line); change.Begin.Cell > len(line.Cells) {
			// bad rune offset
			return
		}

		var newSegments Segments

		switch change.Op {

		case OpInsert:
			newSegments = moment.segments.Sub(-1, change.Begin.Line)
			line := moment.GetLine(change.Begin.Line)
			offset := 0
			for _, cell := range line.Cells[:change.Begin.Cell] {
				offset += cell.Len
			}
			content := line.content[:offset] + change.String + line.content[offset:]
			numRunesInserted += len([]rune(change.String))
			changingLastLine := change.Begin.Line == moment.NumLines()-1
			lines := splitLines(content)
			newSegment := new(Segment)
			for i, content := range lines {
				if changingLastLine && i == len(lines)-1 {
					// add newline to the last line
					if !strings.HasSuffix(content, "\n") {
						content += "\n"
						numRunesInserted++
					}
				}
				newSegment.lines = append(newSegment.lines, &Line{
					content:  content,
					initOnce: new(sync.Once),
					config:   &config,
				})
			}
			newSegments = append(newSegments, newSegment)
			newSegments = append(newSegments, moment.segments.Sub(change.Begin.Line+1, -1)...)

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
						change.End.Cell = len(moment.GetLine(change.End.Line).Cells) - 1
					} else {
						if change.End.Cell+change.Number >= len(line.Cells) {
							// next line
							change.Number -= len(line.Cells) - change.End.Cell
							change.End.Cell = 0
							change.End.Line++
						} else {
							change.End.Cell += change.Number
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
			newSegments = moment.segments.Sub(-1, change.Begin.Line)
			var b strings.Builder
			for lineNum := change.Begin.Line; lineNum <= change.End.Line; lineNum++ {
				if lineNum >= moment.NumLines() {
					break
				}
				if lineNum == change.Begin.Line {
					for _, cell := range moment.GetLine(lineNum).Cells {
						if cell.RuneOffset >= change.Begin.Cell {
							break
						}
						b.WriteRune(cell.Rune)
					}
				}
				if lineNum == change.End.Line {
					for _, cell := range moment.GetLine(lineNum).Cells {
						if cell.RuneOffset < change.End.Cell {
							continue
						}
						b.WriteRune(cell.Rune)
					}
				}
			}
			changingLastLine := change.End.Line >= moment.NumLines()-1
			lines := splitLines(b.String())
			newSegment := new(Segment)
			for i, content := range lines {
				if changingLastLine && i == len(lines)-1 {
					// add newline to the last line
					if !strings.HasSuffix(content, "\n") {
						content += "\n"
					}
				}
				newSegment.lines = append(newSegment.lines, &Line{
					content:  content,
					initOnce: new(sync.Once),
					config:   &config,
				})
			}
			newSegments = append(newSegments, newSegment)
			res := change.End.Line + 1
			if res < moment.NumLines() {
				newSegments = append(newSegments, moment.segments.Sub(res, -1)...)
			}

		}

		newMoment.segments = newSegments
		var buffer *Buffer
		linkedOne(moment, &buffer)
		if buffer != nil {
			link(buffer, newMoment)
		}

		return
	}

}

type InsertAtPositionFunc func(
	str string,
	fn PositionFunc,
)

func (_ Provide) InsertAtPositionFunc(
	v CurrentView,
	m CurrentMoment,
	scope Scope,
	moveCursor MoveCursor,
	apply ApplyChange,
) InsertAtPositionFunc {
	return func(
		str string,
		fn PositionFunc,
	) {

		view := v()
		if view == nil {
			return
		}

		position := fn()
		change := Change{
			Op:     OpInsert,
			Begin:  position,
			String: str,
		}
		var newMoment *Moment
		var nRunesInserted int
		moment := m()
		newMoment, nRunesInserted = apply(moment, change)

		view.switchMoment(scope, newMoment)

		col := newMoment.GetLine(position.Line).Cells[position.Cell].DisplayOffset
		moveCursor(Move{AbsLine: intP(position.Line), AbsCol: &col})
		moveCursor(Move{RelRune: nRunesInserted})

	}
}

type DeleteWithinRange func(
	r Range,
)

func (_ Provide) DeleteWithinRange(
	v CurrentView,
	m CurrentMoment,
	scope Scope,
	moveCursor MoveCursor,
	apply ApplyChange,
) DeleteWithinRange {
	return func(
		r Range,
	) {
		view := v()
		if view == nil {
			return
		}
		change := Change{
			Op:    OpDelete,
			Begin: r.Begin,
			End:   r.End,
		}
		var newMoment *Moment
		moment := m()
		newMoment, _ = apply(moment, change)
		view.switchMoment(scope, newMoment)
		col := newMoment.GetLine(r.Begin.Line).Cells[r.Begin.Cell].DisplayOffset
		moveCursor(Move{AbsLine: intP(r.Begin.Line), AbsCol: &col})
	}
}

type DeleteWithinPositionFuncs func(
	begin PositionFunc,
	end PositionFunc,
)

func (_ Provide) DeleteWithinPositionFuncs(
	scope Scope,
	cur CurrentView,
	deleteRange DeleteWithinRange,
) DeleteWithinPositionFuncs {
	return func(
		beginFn PositionFunc,
		endFn PositionFunc,
	) {
		view := cur()
		if view == nil {
			return
		}
		begin := beginFn()
		end := endFn()
		deleteRange(Range{
			Begin: begin,
			End:   end,
		})
	}
}

type ReplaceWithinRange func(
	r Range,
	text string,
) (
	newMoment *Moment,
)

func (_ Provide) ReplaceWithinRange(
	v CurrentView,
	m CurrentMoment,
	scope Scope,
	moveCursor MoveCursor,
	apply ApplyChange,
) ReplaceWithinRange {
	return func(
		r Range,
		text string,
	) (
		newMoment *Moment,
	) {

		view := v()
		if view == nil {
			return
		}
		moment := m()

		if r.Begin != r.End {
			// delete
			change := Change{
				Op:    OpDelete,
				Begin: r.Begin,
				End:   r.End,
			}
			moment, _ = apply(moment, change)
		}

		// insert
		change := Change{
			Op:     OpInsert,
			Begin:  r.Begin,
			String: text,
		}
		var nRunesInserted int
		moment, nRunesInserted = apply(moment, change)

		view.switchMoment(scope, moment)

		col := moment.GetLine(r.Begin.Line).Cells[r.Begin.Cell].DisplayOffset
		moveCursor(Move{AbsLine: intP(r.Begin.Line), AbsCol: &col})
		moveCursor(Move{RelRune: nRunesInserted})

		newMoment = moment
		return
	}

}

type DeletePrevRune func()

func (_ Provide) DeletePrevRune(
	del DeleteWithinPositionFuncs,
	begin PosPrevRune,
	end PosCursor,
) DeletePrevRune {
	return func() {
		del(
			PositionFunc(begin),
			PositionFunc(end),
		)
	}
}

type DeleteRune func()

func (_ Provide) DeleteRune(
	del DeleteWithinPositionFuncs,
	begin PosCursor,
	end PosNextRune,
) DeleteRune {
	return func() {
		del(
			PositionFunc(begin),
			PositionFunc(end),
		)
	}
}

type DeleteSelected func(
	afterFunc AfterFunc,
)

func (_ Provide) DeleteSelected(
	cur CurrentView,
	scope Scope,
	deleteRagne DeleteWithinRange,
) DeleteSelected {
	return func(
		afterFunc AfterFunc,
	) {
		view := cur()
		if view == nil {
			return
		}

		// delete selected
		if r := view.selectedRange(); r != nil {
			deleteRagne(*r)
			view.SelectionAnchor = nil
		}

		if afterFunc != nil {
			scope.Call(afterFunc)
		}

	}
}

type Delete func() (
	abort Abort,
)

func (_ Provide) Delete(
	cur CurrentView,
	deleteSelected DeleteSelected,
) Delete {
	return func() (
		abort Abort,
	) {
		view := cur()
		if view != nil && view.selectedRange() != nil {
			after := AfterFunc(func() {})
			deleteSelected(after)
		} else {
			abort = true
		}
		return
	}
}

type ChangeText func() (
	abort Abort,
)

func (_ Provide) ChangeText(
	cur CurrentView,
	deleteSelected DeleteSelected,
) ChangeText {
	return func() (
		abort Abort,
	) {

		if view := cur(); view != nil && view.selectedRange() != nil {
			// if selected
			after := AfterFunc(func(enable EnableEditMode) {
				enable()
			})
			deleteSelected(after)

		} else {
			abort = true
		}

		return
	}
}

type ChangeToWordEnd func()

func (_ Provide) ChangeToWordEnd(
	cur CurrentView,
	enable EnableEditMode,
	del DeleteWithinPositionFuncs,
	begin PosCursor,
	end PosWordEnd,
) ChangeToWordEnd {
	return func() {
		if cur() == nil {
			return
		}
		del(
			PositionFunc(begin),
			PositionFunc(end),
		)
		enable()
	}
}

type DeleteLine func()

func (_ Provide) DeleteLine(
	v CurrentView,
	m CurrentMoment,
	del DeleteWithinPositionFuncs,
	lineBegin LineBegin,
	posPrevLineEnd PosPrevLineEnd,
	posLineEnd PosLineEnd,
	posLineBegin PosLineBegin,
	posNextLineBegin PosNextLineBegin,
) DeleteLine {
	return func() {
		view := v()
		if view == nil {
			return
		}
		if view.CursorLine == m().NumLines()-1 {
			del(
				PositionFunc(posPrevLineEnd),
				PositionFunc(posLineEnd),
			)
			lineBegin()
		} else {
			del(
				PositionFunc(posLineBegin),
				PositionFunc(posNextLineBegin),
			)
		}
	}
}
