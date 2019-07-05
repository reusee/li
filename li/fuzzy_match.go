package li

import (
	"github.com/junegunn/fzf/src/algo"
	"github.com/junegunn/fzf/src/util"
)

var fzfSlab = util.MakeSlab(2048, 2048)

func fuzzyMatched(
	pattern []rune,
	chars *util.Chars,
) (
	ok bool,
	maxMatchLen int,
	score int,
) {
	result, pos := algo.FuzzyMatchV2(
		false, false, true,
		chars, pattern,
		true, fzfSlab,
	)
	if pos == nil {
		return
	}
	ok = true
	maxMatchLen = result.End
	score = result.Score
	return
}
