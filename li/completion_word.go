package li

import (
	"regexp"
)

func (_ Provide) CollectWords(
	on On,
) Init2 {

	wordPattern := regexp.MustCompile(`[a-zA-Z0-9]+`)
	shard := numCPU

	newLines := make(chan *Line, 10000)
	for i := 0; i < shard; i++ {
		go func() {

			wordSet := make(map[string]struct{})

			for {
				select {

				case line := <-newLines:
					if len(line.content) == 0 {
						continue
					}
					words := wordPattern.FindAllString(line.content, -1)
					for _, word := range words {
						if _, ok := wordSet[word]; ok {
							continue
						}
						wordSet[word] = struct{}{}
					}

				}
			}

		}()
	}

	on(EvLineInitialized, func(
		line *Line,
	) {
		newLines <- line
	})

	on(EvCollectCompletionCandidate, func(
		moment *Moment,
		state ViewMomentState,
		scope Scope,
		add AddCompletionCandidate,
	) {

		// get pattern
		line := moment.GetLine(scope, state.CursorLine)
		var cell int
		col := 0
		for i := 0; i < len(line.Cells); i++ {
			if col >= state.CursorCol {
				break
			}
			col += line.Cells[i].DisplayWidth
			cell = i
		}
		endCell := cell + 1
		for cell > 0 {
			category := runeCategory(line.Cells[cell].Rune)
			idx := cell - 1
			if idx < 0 {
				break
			}
			prevCategory := runeCategory(line.Cells[idx].Rune)
			if category != prevCategory {
				break
			}
			cell--
		}
		if endCell == cell {
			// no pattern
			return
		}
		patternRunes := line.Runes[cell:endCell]

		//TODO
		add(CompletionCandidate{
			Text: string(patternRunes),
		})

	})

	return nil
}
