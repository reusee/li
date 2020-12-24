package li

import (
	"go/token"
	"strings"
	"sync"
)

type GoLexicalStainer struct {
	//TODO eviction
	cache        sync.Map
	syntaxStyles SyntaxStyles
}

type GoLexicalStainerCacheKey struct {
	MomentID
	LineNumber
}

type NewGoLexicalStainer func() *GoLexicalStainer

func (_ Provide) NewGoLexicalStainer(
	syntaxStyles SyntaxStyles,
) NewGoLexicalStainer {
	return func() *GoLexicalStainer {
		return &GoLexicalStainer{
			syntaxStyles: syntaxStyles,
		}
	}
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

		line := moment.GetLine(int(lineNum))
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
		return s.syntaxStyles.Keyword
	} else if strings.HasSuffix(attr, "_literal") {
		return s.syntaxStyles.Literal
	}
	switch attr {
	case "type_identifier":
		return s.syntaxStyles.Type
	case "comment":
		return s.syntaxStyles.Comment
	case "bool", "byte", "complex64", "complex128", "error", "float32", "float64",
		"int", "int8", "int16", "int32", "int64", "rune", "string",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"true", "false", "iota",
		"nil",
		"append", "cap", "close", "complex", "copy", "delete", "imag", "len",
		"make", "new", "panic", "print", "println", "real", "recover":
		return s.syntaxStyles.Builtin
	case "escape_sequence":
		return s.syntaxStyles.Literal
	}
	return nil
}
