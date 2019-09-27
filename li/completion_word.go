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
) Init2 {

	type Word struct {
		Text       string
		LowerRunes []rune
	}

	type Req struct {
		Scope   Scope
		Pattern []rune
		Fn      func([]CompletionCandidate)
	}

	type CollectJob struct {
		Moment *Moment
	}

	wordPattern := regexp.MustCompile(`[a-zA-Z0-9]+`)
	shard := numCPU
	var jobs []chan CollectJob
	var reqs []chan Req
	for i := 0; i < shard; i++ {
		jobs = append(jobs, make(chan CollectJob, 10000))
		reqs = append(reqs, make(chan Req))
	}

	for i := 0; i < shard; i++ {
		i := i
		go func() {

			wordSets := make(map[HashSum]map[string]Word)

			for {
				select {

				case job := <-jobs[i]:

					for _, segment := range job.Moment.segments {
						sum := segment.Sum()
						if _, ok := wordSets[sum]; ok {
							continue
						}

						wordSet := make(map[string]Word)
						for _, line := range segment.lines {
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
						}

						wordSets[sum] = wordSet
					}

				case req := <-reqs[i]:
					var candidates []CompletionCandidate
					req.Scope.Call(func(
						views Views,
					) {
						for _, view := range views {
							moment := view.GetMoment()
							for _, segment := range moment.segments {
								set, ok := wordSets[segment.Sum()]
								if !ok {
									continue
								}
								for _, word := range set {
									pi := 0
									wi := 0
									var offsets []int
									for pi < len(req.Pattern) && wi < len(word.LowerRunes) {
										if req.Pattern[pi] == word.LowerRunes[wi] {
											offsets = append(offsets, wi)
											pi++
											wi++
										} else {
											wi++
										}
									}
									if pi < len(req.Pattern) {
										continue
									}
									candidates = append(candidates, CompletionCandidate{
										Text:             word.Text,
										Rank:             float64(wi) / float64(pi),
										MatchRuneOffsets: offsets,
									})
								}
							}
						}
					})
					req.Fn(candidates)

				}
			}

		}()
	}

	var n int64
	on(EvMomentSwitched, func(
		moments [2]*Moment,
	) {
		jobs[int(atomic.AddInt64(&n, 1))%shard] <- CollectJob{
			Moment: moments[1],
		}
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
		pattern := string(patternRunes)
		for i, r := range patternRunes {
			patternRunes[i] = unicode.ToLower(r)
		}

		allCandidates := make(map[string]CompletionCandidate)
		var l sync.Mutex
		wg := new(sync.WaitGroup)
		wg.Add(shard)
		for i := 0; i < shard; i++ {
			reqs[i] <- Req{
				Scope:   scope,
				Pattern: patternRunes,
				Fn: func(candidates []CompletionCandidate) {
					l.Lock()
					for _, candidate := range candidates {
						if _, ok := allCandidates[candidate.Text]; ok {
							continue
						}
						if candidate.Text == pattern {
							continue
						}
						allCandidates[candidate.Text] = candidate
					}
					l.Unlock()
					wg.Done()
				},
			}
		}
		wg.Wait()

		for _, candidate := range allCandidates {
			add(candidate)
		}

	})

	return nil
}
