package li

import (
	"sync"

	"github.com/reusee/li/treesitter"
)

type GoLexicalStainer struct {
	//TODO eviction
	cache sync.Map
}

type GoLexicalStainerCacheKey struct {
	MomentID
	LineNumber
}

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

		parser := moment.GetParser()
		line := moment.GetLine(int(lineNum))
		notHandled := make(map[string]bool)
		for _, cell := range line.Cells {
			node := parser.NodeAt(treesitter.Point(int(lineNum), cell.RuneOffset))
			nodeType := treesitter.NodeType(node)
			if fn, ok := goSyntaxStyle[nodeType]; ok && fn != nil {
				fns = append(fns, fn)
			} else {
				notHandled[nodeType] = true
				fns = append(fns, nil)
			}
		}

		if len(notHandled) > 0 {
			appendJournal("%+v", notHandled)
		}

		s.cache.Store(key, fns)

		return
	}
}
