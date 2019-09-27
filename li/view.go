package li

import (
	"sync"
	"sync/atomic"
)

type ViewID int64

type View struct {
	sync.RWMutex

	ID      ViewID
	Buffer  *Buffer
	moment  *Moment
	Stainer Stainer

	Box        Box
	ContentBox Box

	ViewMomentState

	FrameBuffer     *FrameBuffer
	FrameBufferArgs ViewUIArgs

	//TODO eviction
	MomentStates map[*Moment]ViewMomentState

	//TODO merge moment segments
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
		moment: moment,
		Stainer: func() Stainer {
			if fn, ok := languageStainers[buffer.language]; ok {
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

func (v View) cursorPosition(scope Scope) Position {
	if v.CursorLine >= v.GetMoment().NumLines() {
		return Position{
			Line: -1,
			Cell: -1,
		}
	}
	line := v.GetMoment().GetLine(scope, v.CursorLine)
	if line == nil {
		return Position{
			Line: -1,
			Cell: -1,
		}
	}
	col := 0
	for i := 0; i <= len(line.Cells); i++ {
		if col >= v.CursorCol {
			return Position{
				Line: v.CursorLine,
				Cell: i,
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
		Cell: -1,
	}
}

type evMomentSwitched struct{}

var EvMomentSwitched = new(evMomentSwitched)

func (v *View) switchMoment(scope Scope, m *Moment) {
	v.Lock()
	// save
	v.MomentStates[v.moment] = v.ViewMomentState
	// restore
	old := v.moment
	v.moment = m
	if state, ok := v.MomentStates[m]; ok {
		v.ViewMomentState = state
	}
	v.Unlock()
	// trigger event
	scope.Call(func(
		trigger Trigger,
	) {
		trigger(scope.Sub(
			func() (*View, *Buffer, [2]*Moment) {
				return v, v.Buffer, [2]*Moment{old, m}
			},
		), EvMomentSwitched)
	})
}

func (v *View) GetMoment() (m *Moment) {
	v.RLock()
	defer v.RUnlock()
	m = v.moment
	return
}

func (_ Provide) ViewEvents(
	on On,
	j AppendJournal,
	config DebugConfig,
) Init2 {

	if config.Verbose {
		on(EvMomentSwitched, func(view *View, ms [2]*Moment) {
			j("view %d switch moment from %d to %d", view.ID, ms[0].ID, ms[1].ID)
		})
	}

	// buffer saving state
	on(EvCollectStatusSections, func(
		add AddStatusSection,
		v CurrentView,
		styles []Style,
	) {
		view := v()
		if view == nil {
			return
		}
		if view.Buffer.LastSyncFileInfo == view.GetMoment().FileInfo {
			return
		}
		add("file", [][]any{
			{"unsaved", styles[1], AlignRight, Padding(0, 2, 0, 0)},
		})
	})

	return nil
}

func WithCurrentViewMoment(
	v CurrentView,
	m CurrentMoment,
) (*View, *Moment) {
	return v(), m()
}
