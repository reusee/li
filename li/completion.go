package li

import "time"

func (_ Provide) Completion(
	on On,
	run RunInMainLoop,
) Init2 {

	on(EvViewRendered, func(
		view *View,
		moment *Moment,
		curModes CurrentModes,
		procs CompletionProcs,
		config CompletionConfig,
	) {

		if !IsEditing(curModes()) {
			return
		}

		state := view.ViewMomentState

		procs <- func() {

			// async calculate candidates
			//TODO

			// show
			time.AfterFunc(time.Millisecond*time.Duration(config.DelayMilliseconds), func() {
				run(func(
					j AppendJournal,
					curModes CurrentModes,
					curView CurrentView,
				) {

					if !IsEditing(curModes()) {
						// skip if not editing
						return
					}
					cur := curView()
					if cur != view {
						// skip if view switched
						return
					}
					if cur.ViewMomentState != state {
						// skip if state changed
						return
					}

					j("%d rendered, %+v", view.ID, state)
				})
			})

		}

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
