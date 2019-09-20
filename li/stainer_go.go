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

var goSyntaxColors = map[string]Color{
	"type_identifier":  HexColor(0x0099CC),
	"identifier":       HexColor(0x00CC99),
	"argument_list":    HexColor(0x9900CC),
	"parameter_list":   HexColor(0x9900CC),
	"block":            HexColor(0x99CC00),
	"return_statement": HexColor(0xCC0099),
	"field_identifier": HexColor(0xCC9900),
}

func (s *GoLexicalStainer) Line() any {
	return func(
		moment *Moment,
		lineNum LineNumber,
		appendJournal AppendJournal,
	) (
		colors []*Color,
	) {

		key := GoLexicalStainerCacheKey{moment.ID, lineNum}
		if v, ok := s.cache.Load(key); ok {
			return v.([]*Color)
		}

		parser := moment.GetParser()
		line := moment.GetLine(int(lineNum))
		notHandled := make(map[string]bool)
		for _, cell := range line.Cells {
			node := parser.NodeAt(treesitter.Point(int(lineNum), cell.RuneOffset))
			nodeType := treesitter.NodeType(node)
			if color, ok := goSyntaxColors[nodeType]; ok && color != black {
				colors = append(colors, &color)
			} else {
				notHandled[nodeType] = true
				colors = append(colors, nil)
			}
		}

		if len(notHandled) > 0 {
			appendJournal("%+v", notHandled)
		}

		s.cache.Store(key, colors)

		return
	}
}
