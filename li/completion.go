package li

import (
	"sort"
	"sync/atomic"
	"time"
)

type EvCollectCompletionCandidate struct {
	Add    AddCompletionCandidate
	View   *View
	Moment *Moment
	State  ViewMomentState
}

type CompletionCandidate struct {
	Text             string
	Rank             float64
	MatchRuneOffsets []int
	Begin            Position
	End              Position
}

type AddCompletionCandidate func(CompletionCandidate)

type ContinueCompletion *int64

func (_ Provide) ContinueCompletion() ContinueCompletion {
	var cont int64
	return &cont
}

func (_ Provide) Completion(
	on On,
	run RunInMainLoop,
) OnStartup {

	return func() {
		var completionOverlayID ID
		closeOverlay := func() {
			if completionOverlayID > 0 {
				id := completionOverlayID
				run(func(
					scope Scope,
					closeOverlay CloseOverlay,
				) {
					closeOverlay(id)
				})
			}
		}

		on(func(
			ev EvKeyEventHandled,
			curView CurrentView,
			procs CompletionProcs,
			config CompletionConfig,
			trigger Trigger,
			scope Scope,
			maxWidth Width,
			maxHeight Height,
			cont ContinueCompletion,
			calLineHeight CalculateSumLineHeight,
		) {

			// continue, dont update
			if atomic.CompareAndSwapInt64(cont, 1, 0) {
				return
			}

			view := curView()
			if view == nil {
				return
			}
			moment := view.GetMoment()
			state := view.ViewMomentState

			skip := func(scope Scope) (b bool) {
				scope.Call(func(
					curModes CurrentModes,
					curView CurrentView,
				) {
					// skip if not editing
					if !IsEditing(curModes()) {
						b = true
						return
					}
					cur := curView()
					// skip if view switched
					if cur != view {
						b = true
						return
					}
					// skip if state changed
					if cur.ViewMomentState != state {
						b = true
						return
					}
				})
				return
			}
			if skip(scope) {
				closeOverlay()
				return
			}

			// delay
			time.AfterFunc(time.Millisecond*time.Duration(config.DelayMilliseconds), func() {

				if skip(scope) {
					closeOverlay()
					return
				}

				// async
				procs <- func() {

					// collect candidates
					var candidates []CompletionCandidate
					trigger(EvCollectCompletionCandidate{
						Add: func(c CompletionCandidate) {
							candidates = append(candidates, c)
						},
						View:   view,
						Moment: moment,
						State:  state,
					})

					if skip(scope) {
						closeOverlay()
						return
					}

					// sort
					sort.SliceStable(candidates, func(i, j int) bool {
						c1 := candidates[i]
						c2 := candidates[j]
						if c1.MatchRuneOffsets[0] != c2.MatchRuneOffsets[0] {
							return c1.MatchRuneOffsets[0] < c2.MatchRuneOffsets[0]
						}
						if c1.Rank != c2.Rank {
							return c1.Rank > c2.Rank
						}
						return c1.Text < c2.Text
					})

					// position
					width := 0
					for _, candidate := range candidates {
						if w := displayWidth(candidate.Text); w > width {
							width = w
						}
					}
					width += 2 // padding
					if width > int(maxWidth)-10 {
						width = int(maxWidth) - 10
					}
					view.RLock()
					defer view.RUnlock()
					cursorY := view.ContentBox.Top
					lineHeight := calLineHeight(moment, [2]int{view.ViewportLine, view.CursorLine})
					cursorY += lineHeight
					height := len(candidates)
					below := true
					var maxH int
					if cursorY < int(maxHeight)/2 {
						maxH = int(maxHeight) - cursorY - 1
					} else {
						below = false
						maxH = cursorY
					}
					if height > maxH {
						height = maxH
					}
					cursorX := view.ContentBox.Left + (view.CursorCol - view.ViewportCol)
					left := cursorX - 1
					if left+width > int(maxWidth) {
						left = int(maxWidth) - width
					}
					right := left + width
					top := cursorY + 1 // below
					bottom := top + height
					if !below {
						bottom = cursorY
						top = bottom - height
					}
					box := Box{top, left, bottom, right}

					// truncate
					candidates = candidates[:height]

					// reverse
					if !below {
						for i := 0; i < len(candidates)/2; i++ {
							candidates[i], candidates[len(candidates)-1-i] = candidates[len(candidates)-1-i], candidates[i]
						}
					}

					// update
					run(func(
						scope Scope,
						j AppendJournal,
						pushOverlay PushOverlay,
					) {
						// close
						closeOverlay()
						if len(candidates) == 0 {
							return
						}

						// push overlay
						overlay := OverlayObject(&CompletionList{
							Box:        box,
							Candidates: candidates,
							Below:      below,
							Moment:     moment,
							View:       view,
						})
						completionOverlayID = pushOverlay(overlay)

					})

				}
			})

		})

	}
}

type CompletionList struct {
	Box        Box
	Candidates []CompletionCandidate
	Below      bool
	Moment     *Moment
	View       *View

	index *int
}

var _ Element = new(CompletionList)

var _ KeyStrokeHandler = new(CompletionList)

func (c *CompletionList) RenderFunc() any {
	return func(
		scope Scope,
		cur CurrentView,
		maxWidth Width,
		maxHeight Height,
		getStyle GetStyle,
	) Element {

		style := getStyle("Completion")
		selectedStyle := getStyle("CompletionSelected")

		box := c.Box
		box.Left++
		var texts []Element
		for idx, candidate := range c.Candidates {
			candidate := candidate
			lineStyle := style
			if c.index != nil && idx == *c.index {
				lineStyle = selectedStyle
			}
			texts = append(texts, Text(
				box,
				candidate.Text,
				OffsetStyleFunc(func(i int) StyleFunc {
					for _, offset := range candidate.MatchRuneOffsets {
						if offset == i {
							return lineStyle.SetUnderline(true)
						}
					}
					return lineStyle
				}),
			))
			box.Top++
		}

		return Rect(
			c.Box,
			Fill(true),
			Padding(0, 1, 0, 1),
			style,
			texts,
		)

	}
}

func (c *CompletionList) StrokeSpecs() any {
	return func(
		cont ContinueCompletion,
	) []StrokeSpec {
		return []StrokeSpec{
			{
				Sequence: []string{"Tab"},

				Func: func(
					scope Scope,
				) {
					apply := func(index int) {
						c.index = &index
						candidate := c.Candidates[index]
						scope.Sub(
							AsCurrentView(c.View),
							AsCurrentMoment(c.Moment),
						).Call(func(
							replace ReplaceWithinRange,
						) {
							replace(
								Range{
									candidate.Begin,
									candidate.End,
								},
								candidate.Text,
							)
						})

					}
					if c.index == nil {
						index := 0
						if !c.Below {
							index = len(c.Candidates) - 1
						}
						apply(index)
					} else if c.Below && *c.index < len(c.Candidates)-1 {
						index := *c.index + 1
						apply(index)
					} else if !c.Below && *c.index >= 1 {
						index := *c.index - 1
						apply(index)
					} else {
						c.index = nil
						c.View.switchMoment(scope, c.Moment)
					}

					atomic.AddInt64(cont, 1)
				},
			},
		}
	}
}

type CompletionProcs chan func()

func (_ Provide) CompletionProcs() (
	p CompletionProcs,
) {

	p = make(chan func(), 512)
	for i := 0; i < numCPU; i++ {
		go func() {
			for {
				(<-p)()
			}
		}()
	}

	return
}

type CompletionConfig struct {
	DelayMilliseconds int
}

func (_ Provide) CompletionConfig(
	get GetConfig,
) CompletionConfig {
	var config struct {
		Completion CompletionConfig
	}
	ce(get(&config))
	return config.Completion
}
