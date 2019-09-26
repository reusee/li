package li

import (
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"unicode"
)

func (_ Provide) CollectWords(
	on On,
	cur CurrentView,
	scope Scope,
	j AppendJournal,
) Init2 {

	type Word struct {
		Text       string
		LowerRunes []rune
	}

	type Req struct {
		Pattern []rune
		Fn      func([]Word)
	}

	wordPattern := regexp.MustCompile(`[a-zA-Z0-9]+`)
	shard := numCPU
	var newLines []chan *Line
	var reqs []chan Req
	for i := 0; i < shard; i++ {
		newLines = append(newLines, make(chan *Line, 10000))
		reqs = append(reqs, make(chan Req))
	}

	for i := 0; i < shard; i++ {
		i := i
		go func() {

			wordSet := make(map[string]Word)

			for {
				select {

				case line := <-newLines[i]:
					if len(line.content) == 0 {
						continue
					}

					//TODO do not add incomplete words
					indexPairs := wordPattern.FindAllStringIndex(line.content, -1)
					for _, pair := range indexPairs {
						word := line.content[pair[0]:pair[1]]
						if _, ok := wordSet[word]; ok {
							continue
						}
						wordSet[word] = Word{
							Text:       word,
							LowerRunes: []rune(strings.ToLower(word)),
						}
					}

				case req := <-reqs[i]:
					var words []Word
					for _, word := range wordSet {
						pi := 0
						wi := 0
						for pi < len(req.Pattern) && wi < len(word.LowerRunes) {
							if req.Pattern[pi] == word.LowerRunes[wi] {
								pi++
								wi++
							} else {
								wi++
							}
						}
						if pi < len(req.Pattern) {
							continue
						}
						words = append(words, word)
					}
					req.Fn(words)

				}
			}

		}()
	}

	var n int64
	on(EvLineInitialized, func(
		line *Line,
	) {
		newLines[int(atomic.AddInt64(&n, 1))%shard] <- line
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
		for i, r := range patternRunes {
			patternRunes[i] = unicode.ToLower(r)
		}

		candidateWords := make(map[string]Word)
		var l sync.Mutex
		wg := new(sync.WaitGroup)
		wg.Add(shard)
		for i := 0; i < shard; i++ {
			reqs[i] <- Req{
				Pattern: patternRunes,
				Fn: func(words []Word) {
					l.Lock()
					for _, word := range words {
						if _, ok := candidateWords[word.Text]; ok {
							continue
						}
						candidateWords[word.Text] = word
					}
					l.Unlock()
					wg.Done()
				},
			}
		}
		wg.Wait()

		for _, word := range candidateWords {
			add(CompletionCandidate{
				Text: word.Text,
			})
		}

	})

	return nil
}
