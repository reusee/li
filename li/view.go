package li

import (
	"sync/atomic"
)

type ViewID int64

type View struct {
	ID      ViewID
	Buffer  *Buffer
	Moment  *Moment
	Stainer Stainer

	Box Box

	ViewMomentState

	FrameBuffer     *FrameBuffer
	FrameBufferArgs ViewUIArgs

	MomentStates map[*Moment]ViewMomentState
}

type ViewMomentState struct {
	ViewportLine    int
	ViewportCol     int
	CursorLine      int
	CursorCol       int
	SelectionAnchor *Position
	PreferCursorCol int
}

type Views map[ViewID]*View

func (_ Provide) DefaultViews() Views {
	return make(Views)
}

var nextViewID int64

func NewViewFromBuffer(
	buffer *Buffer,
	width Width,
	height Height,
	views Views,
	scope Scope,
	_ ViewGroups, // dep
	curGroup CurrentViewGroup,
	link Link,
	cur CurrentView,
	linkedOne LinkedOne,
) (
	view *View,
	err error,
) {

	id := ViewID(atomic.AddInt64(&nextViewID, 1))
	var moment *Moment
	linkedOne(buffer, &moment)

	view = &View{
		ID:     id,
		Buffer: buffer,
		Moment: moment,
		Stainer: func() Stainer {
			if fn, ok := languageStainers[moment.language]; ok {
				return fn()
			}
			return new(NoopStainer)
		}(),
		ViewMomentState: ViewMomentState{
			ViewportLine: 0,
			ViewportCol:  0,
			CursorLine:   0,
			CursorCol:    0,
		},
		MomentStates: make(map[*Moment]ViewMomentState),
		Box: Box{
			Top:    0,
			Left:   0,
			Right:  int(width),
			Bottom: int(height),
		},
	}

	link(view, curGroup())
	cur(view)

	views[view.ID] = view

	return
}

func CloseView(
	cur CurrentView,
	views Views,
	scope Scope,
	derive Derive,
	dropLinked DropLinked,
) {
	c := cur()
	if c == nil {
		return
	}
	delete(views, c.ID)
	dropLinked(c)
	dropLinked(c.Buffer)
	derive(
		func() Views {
			return views
		},
	)
}

func (_ Command) CloseView() (spec CommandSpec) {
	spec.Func = func(scope Scope) {
		scope.Call(CloseView)
	}
	return
}

func (v View) cursorPosition() Position {
	if v.CursorLine >= v.Moment.NumLines() {
		return Position{
			Line: -1,
			Rune: -1,
			Col:  -1,
		}
	}
	line := v.Moment.GetLine(v.CursorLine)
	if line == nil {
		return Position{
			Line: -1,
			Rune: -1,
			Col:  -1,
		}
	}
	col := 0
	for i := 0; i <= len(line.Cells); i++ {
		if col >= v.CursorCol {
			return Position{
				Line: v.CursorLine,
				Rune: i,
				Col:  col,
			}
		}
		if i < len(line.Cells) {
			col += line.Cells[i].DisplayWidth
		} else {
			col += 1
		}
	}
	return Position{
		Line: -1,
		Rune: -1,
		Col:  -1,
	}
}

func (v *View) switchMoment(m *Moment) {
	// save
	v.MomentStates[v.Moment] = v.ViewMomentState
	// restore
	v.Moment = m
	if state, ok := v.MomentStates[m]; ok {
		v.ViewMomentState = state
	}
}
