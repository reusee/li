package treesitter

/*
#include <tree_sitter/api.h>
#include <tree-sitter-go/src/parser.c>

#cgo CFLAGS: -I${SRCDIR}/tree-sitter/lib/include
#cgo LDFLAGS: ${SRCDIR}/tree-sitter/libtree-sitter.a
*/
import "C"
import "unsafe"

type (
	TSParser = *C.TSParser
	TSTree   = *C.TSTree
	TSNode   = C.TSNode
	TSPoint  = C.TSPoint
)

type Parser struct {
	Parser TSParser
	Tree   TSTree
	Root   TSNode
}

func ParseGo(src unsafe.Pointer, l int) *Parser {
	parser := C.ts_parser_new()
	C.ts_parser_set_language(parser, C.tree_sitter_go())
	tree := C.ts_parser_parse_string(
		parser,
		nil,
		(*C.char)(src),
		C.uint(l),
	)
	root := C.ts_tree_root_node(tree)
	return &Parser{
		Parser: parser,
		Tree:   tree,
		Root:   root,
	}
}

func (p *Parser) Close() {
	C.ts_tree_delete(p.Tree)
	C.ts_parser_delete(p.Parser)
}

func Walk(node TSNode, fn func(TSNode)) {
	fn(node)
	count := C.ts_node_child_count(node)
	for i := C.uint(0); i < count; i++ {
		child := C.ts_node_child(node, i)
		Walk(child, fn)
	}
}

func NodeType(node TSNode) string {
	return C.GoString(C.ts_node_type(node))
}

func NodePosition(node TSNode) (
	startRow, startCol int,
	endRow, endCol int,
) {
	p := C.ts_node_start_point(node)
	startRow = int(p.row)
	startCol = int(p.column)
	p = C.ts_node_end_point(node)
	endRow = int(p.row)
	endCol = int(p.column)
	return
}

func Point(row int, col int) TSPoint {
	return TSPoint{C.uint(row), C.uint(col)}
}

func (p *Parser) NodeAt(point TSPoint) TSNode {
	return C.ts_node_descendant_for_point_range(p.Root, point, point)
}
