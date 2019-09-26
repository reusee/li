package li

import "time"

type evCollectCompletionCandidate struct{}

var EvCollectCompletionCandidate = new(evCollectCompletionCandidate)

type CompletionCandidate struct {
	Text string
}

type AddCompletionCandidate func(CompletionCandidate)

func (_ Provide) Completion(
	on On,
	run RunInMainLoop,
) Init2 {

	on(EvKeyEventHandled, func(
		curView CurrentView,
		procs CompletionProcs,
		config CompletionConfig,
		trigger Trigger,
		scope Scope,
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
			return
		}

		// delay
		time.AfterFunc(time.Millisecond*time.Duration(config.DelayMilliseconds), func() {

			if skip(scope) {
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
					return
				}

				// show
				run(func(
					j AppendJournal,
				) {
					j("completion candidates %+v", candidates)
				})

			}
		})

	})

	return nil
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
