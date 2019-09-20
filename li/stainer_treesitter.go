package li

import "sync"

type (
	TSParserCache struct {
		*sync.Map
	}
)

func (_ Provide) Treesitter() (
	parserCache TSParserCache,
) {

	parserCache = TSParserCache{
		Map: new(sync.Map),
	}

	return
}
