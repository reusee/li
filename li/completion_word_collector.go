package li

import (
	"regexp"
	"strings"
)

func (_ Provide) CollectWords(
	on On,
) Init2 {

	const partLen = 8

	type Node struct {
		Runes [partLen]rune
		Succs map[[partLen]rune]*Node
		Words map[string]*Line
	}

	wordPattern := regexp.MustCompile(`[a-zA-Z0-9]+`)

	newLines := make(chan *Line, 10000)
	for i := 0; i < numCPU; i++ {
		go func() {

			root := &Node{
				Succs: make(map[[partLen]rune]*Node),
				Words: make(map[string]*Line),
			}
			var set func(node *Node, parts [][partLen]rune, word string, line *Line)
			set = func(node *Node, parts [][partLen]rune, word string, line *Line) {
				if len(parts) == 0 {
					return
				}
				node.Words[word] = line
				for i, runes := range parts {
					succ, ok := node.Succs[runes]
					if !ok {
						succ = &Node{
							Runes: runes,
							Succs: make(map[[partLen]rune]*Node),
							Words: make(map[string]*Line),
						}
						node.Succs[runes] = succ
					}
					set(succ, parts[i+1:], word, line)
				}
			}

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
						word = strings.ToLower(word)
						var parts [][partLen]rune
						runes := []rune(word)
						for i := 0; i < len(runes); i += partLen {
							var array [partLen]rune
							end := i + partLen
							if end > len(runes) {
								end = len(runes)
							}
							copy(array[:], runes[i:end])
							parts = append(parts, array)
						}
						set(root, parts, word, line)
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
