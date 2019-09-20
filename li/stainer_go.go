package li

import (
	"sync"
)

type GoLexicalStainer struct {
	//TODO eviction
	cache sync.Map
}

type GoLexicalStainerCacheKey struct {
	MomentID
	LineNumber
}

//TODO configurable
var goSyntaxStyle = map[string]StyleFunc{
	"type_identifier":  SetFG(HexColor(0x0099CC)).SetUnderline(true),
	"identifier":       SetFG(HexColor(0x00CC99)),
	"argument_list":    SetFG(HexColor(0x9900CC)),
	"parameter_list":   SetFG(HexColor(0x9900CC)),
	"block":            SetFG(HexColor(0x99CC00)),
	"return_statement": SetFG(HexColor(0xCC0099)).SetBold(true),
	"field_identifier": SetFG(HexColor(0xCC9900)),
}

func (s *GoLexicalStainer) Line() any {
	return func(
		moment *Moment,
		lineNum LineNumber,
		appendJournal AppendJournal,
	) (
		fns []StyleFunc,
	) {

		key := GoLexicalStainerCacheKey{moment.ID, lineNum}
		if v, ok := s.cache.Load(key); ok {
			return v.([]StyleFunc)
		}

		line := moment.GetLine(int(lineNum))
		for _, cell := range line.Cells {
			attr := moment.GetSyntaxAttr(int(lineNum), cell.RuneOffset)
			if fn, ok := goSyntaxStyle[attr]; ok && fn != nil {
				fns = append(fns, fn)
			} else {
				fns = append(fns, nil)
			}
		}

		s.cache.Store(key, fns)

		return
	}
}
