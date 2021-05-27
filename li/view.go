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

type NewViewFromBuffer func(
	buffer *Buffer,
) (
	view *View,
	err error,
)

func (_ Provide) NewViewFromBuffer(
	width Width,
	height Height,
	views Views,
	_ ViewGroups, // dep
	curGroup CurrentViewGroup,
	link Link,
	cur CurrentView,
	linkedOne LinkedOne,
	trigger Trigger,
	languageStainers LanguageStainers,
) NewViewFromBuffer {
	return func(buffer *Buffer) (view *View, err error) {

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

		trigger(EvMomentSwitched{
			View:   view,
			Buffer: view.Buffer,
			Old:    nil,
			New:    moment,
		})

		return
	}
}

type CloseView func()

func (_ Provide) CloseView(
	cur CurrentView,
	views Views,
	derive Derive,
	dropLinked DropLinked,
) CloseView {
	return func() {
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
}

func (_ Command) CloseView() (spec CommandSpec) {
	spec.Func = func(c CloseView) {
		c()
	}
	return
}

func (v *View) cursorPosition() Position {
	if v.CursorLine >= v.GetMoment().NumLines() {
		return Position{
			Line: -1,
			Cell: -1,
		}
	}
	line := v.GetMoment().GetLine(v.CursorLine)
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

type EvMomentSwitched struct {
	View   *View
	Buffer *Buffer
	Old    *Moment
	New    *Moment
}

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
		trigger(EvMomentSwitched{
			View:   v,
			Buffer: v.Buffer,
			Old:    old,
			New:    m,
		})
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
) OnStartup {
	return func() {

		if config.Verbose {
			on(func(
				ev EvMomentSwitched,
			) {
				j("view %d switch moment from %d to %d",
					ev.View.ID, ev.Old.ID, ev.New.ID)
			})
		}

		// buffer saving state
		on(func(
			ev EvCollectStatusSections,
			v CurrentView,
		) {
			view := v()
			if view == nil {
				return
			}
			if view.Buffer.LastSyncFileInfo == view.GetMoment().FileInfo {
				return
			}
			ev.Add("file", [][]any{
				{"unsaved", ev.Styles[1], AlignRight, Padding(0, 2, 0, 0)},
			})
		})

	}
}
