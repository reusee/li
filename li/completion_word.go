package li

import (
	"regexp"
)

func (_ Provide) CollectWords(
	on On,
) Init2 {

	wordPattern := regexp.MustCompile(`[a-zA-Z0-9]+`)

	newLines := make(chan *Line, 10000)
	for i := 0; i < numCPU; i++ {
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

	return nil
}
