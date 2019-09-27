package li

import (
	"sort"
	"time"
)

type evCollectCompletionCandidate struct{}

var EvCollectCompletionCandidate = new(evCollectCompletionCandidate)

type CompletionCandidate struct {
	Text             string
	Rank             float64
	MatchRuneOffsets []int
}

type AddCompletionCandidate func(CompletionCandidate)

func (_ Provide) Completion(
	on On,
	run RunInMainLoop,
) Init2 {

	var completionOverlayID ID
	closeOverlay := func() {
		if completionOverlayID > 0 {
			id := completionOverlayID
			run(func(scope Scope) {
				scope.Sub(func() ID { return id }).Call(CloseOverlay)
			})
		}
	}

	on(EvKeyEventHandled, func(
		curView CurrentView,
		procs CompletionProcs,
		config CompletionConfig,
		trigger Trigger,
		scope Scope,
		maxWidth Width,
		maxHeight Height,
	) {

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
				trigger(scope.Sub(
					func() AddCompletionCandidate {
						return func(c CompletionCandidate) {
							candidates = append(candidates, c)
						}
					},
					func() (*View, *Moment, ViewMomentState) {
						return view, moment, state
					},
				), EvCollectCompletionCandidate)

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
				cursorY := view.ContentBox.Top + (view.CursorLine - view.ViewportLine)
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

				// update
				run(func(
					scope Scope,
					j AppendJournal,
				) {
					// close
					closeOverlay()
					if len(candidates) == 0 {
						return
					}

					// push overlay
					scope.Sub(func() OverlayObject {
						return &CompletionList{
							Box:        box,
							Candidates: candidates,
							Below:      below,
						}
					}).Call(PushOverlay, &completionOverlayID)

				})

			}
		})

	})

	return nil
}

type CompletionList struct {
	Box        Box
	Candidates []CompletionCandidate
	Below      bool
}

var _ Element = new(CompletionList)

var _ KeyStrokeHandler = new(CompletionList)

func (c *CompletionList) RenderFunc() any {
	return func(
		scope Scope,
		cur CurrentView,
		maxWidth Width,
		maxHeight Height,
		style Style,
	) Element {

		style = style.Background(HexColor(0x123456))

		box := c.Box
		box.Left++
		var texts []Element
		for _, candidate := range c.Candidates {
			candidate := candidate
			texts = append(texts, Text(
				box,
				candidate.Text,
				OffsetStyleFunc(func(i int) Style {
					for _, offset := range candidate.MatchRuneOffsets {
						if offset == i {
							return style.Underline(true)
						}
					}
					return style
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
	return func() []StrokeSpec {
		return []StrokeSpec{
			{
				Sequence: []string{"Tab"},
				Func: func() {
					//TODO
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
