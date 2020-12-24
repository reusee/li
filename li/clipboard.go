package li

import "strings"

type Clip struct {
	Moment *Moment
	Range  Range
	str    *string
}

func (c *Clip) String() string {
	if c.str == nil {
		if c.Range.End.Before(c.Range.Begin) {
			c.Range.End, c.Range.Begin = c.Range.Begin, c.Range.End
		}

		buf := new(strings.Builder)
		lineNum := c.Range.Begin.Line

		line := c.Moment.GetLine(lineNum)
		if line == nil {
			return ""
		}
		begin := line.Cells[c.Range.Begin.Cell].ByteOffset
		end := len(line.content)
		if c.Range.End.Line == c.Range.Begin.Line {
			end = line.Cells[c.Range.End.Cell].ByteOffset
		}
		buf.WriteString(line.content[begin:end])
		lineNum++

		for {
			if lineNum >= c.Moment.NumLines() ||
				lineNum > c.Range.End.Line {
				break
			}
			line = c.Moment.GetLine(lineNum)
			if lineNum == c.Range.End.Line {
				buf.WriteString(
					line.content[:line.Cells[c.Range.End.Cell].ByteOffset],
				)
			} else {
				buf.WriteString(line.content)
			}
			lineNum++
		}

		str := buf.String()
		c.str = &str
	}
	return *c.str
}

type NewClipFromSelection func()

func (_ Provide) NewClipFromSelection(
	cur CurrentView,
	link Link,
) NewClipFromSelection {
	return func() {
		view := cur()
		if view == nil {
			return
		}
		r := view.selectedRange()
		if r == nil {
			return
		}
		clip := Clip{
			Moment: view.GetMoment(),
			Range:  *r,
		}
		link(view.Buffer, clip)
	}
}

func (_ Command) NewClipFromSelection() (spec CommandSpec) {
	spec.Desc = "create new clip from current selection (copy)"
	spec.Func = func(
		newClip NewClipFromSelection,
		toggle ToggleSelection,
	) {
		newClip()
		toggle()
	}
	return
}

type InsertLastClip func()

func (_ Provide) InsertLastClip(
	cur CurrentView,
	linkedOne LinkedOne,
	scope Scope,
) InsertLastClip {
	return func() {
		view := cur()
		if view == nil {
			return
		}
		var clip Clip
		if linkedOne(view.Buffer, &clip) == 0 {
			return
		}
		str := clip.String()
		fn := PositionFunc(PosCursor)
		scope.Sub(&str, &fn).Call(InsertAtPositionFunc)
	}
}

func (_ Command) InsertLastClip() (spec CommandSpec) {
	spec.Desc = "insert contents of last created clip (paste)"
	spec.Func = func(
		insert InsertLastClip,
	) {
		insert()
	}
	return
}
