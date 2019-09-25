package li

import (
	"go/token"
	"strings"
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

func (s *GoLexicalStainer) Line() any {
	return func(
		moment *Moment,
		lineNum LineNumber,
		appendJournal AppendJournal,
		scope Scope,
	) (
		fns []StyleFunc,
	) {

		key := GoLexicalStainerCacheKey{moment.ID, lineNum}
		if v, ok := s.cache.Load(key); ok {
			return v.([]StyleFunc)
		}

		line := moment.GetLine(scope, int(lineNum))
		for _, cell := range line.Cells {
			attr := moment.GetSyntaxAttr(scope, int(lineNum), cell.RuneOffset)
			fns = append(fns, s.AttrStyleFunc(attr))
		}

		s.cache.Store(key, fns)

		return
	}
}

func (s *GoLexicalStainer) AttrStyleFunc(attr string) StyleFunc {
	if token.IsKeyword(attr) {
		return KeywordStyleFunc
	} else if strings.HasSuffix(attr, "_literal") {
		return LiteralStyleFunc
	}
	switch attr {
	case "type_identifier":
		return TypeStyleFunc
	case "comment":
		return CommentStyleFunc
	case "bool", "byte", "complex64", "complex128", "error", "float32", "float64",
		"int", "int8", "int16", "int32", "int64", "rune", "string",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"true", "false", "iota",
		"nil",
		"append", "cap", "close", "complex", "copy", "delete", "imag", "len",
		"make", "new", "panic", "print", "println", "real", "recover":
		return BuiltInStyleFunc
	case "escape_sequence":
		return LiteralStyleFunc
	}
	return nil
}
