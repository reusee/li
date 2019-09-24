package li

func (_ Provide) Completion(
	on On,
	run RunInMainLoop,
) Init2 {

	on(EvViewRendered, func(
		view *View,
		moment *Moment,
		curModes CurrentModes,
		procs CompletionProcs,
	) {

		editing := false
		for _, mode := range curModes() {
			if _, ok := mode.(*EditMode); ok {
				editing = true
				break
			}
		}
		if !editing {
			return
		}

		state := view.ViewMomentState

		procs <- func() {

			//TODO calculate candidates

			//TODO show
			run(func(
				j AppendJournal,
			) {
				j("%d rendered, %+v", view.ID, state)
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
