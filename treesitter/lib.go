package treesitter

/*
#include <tree_sitter/api.h>
#include <tree-sitter-go/src/parser.c>

#cgo CFLAGS: -I${SRCDIR}/tree-sitter/lib/include
#cgo LDFLAGS: ${SRCDIR}/tree-sitter/libtree-sitter.a
*/
import "C"
import "fmt"

func Test() {
	parser := C.ts_parser_new()
	C.ts_parser_set_language(parser, C.tree_sitter_go())
	tree := C.ts_parser_parse_string(
		parser,
		nil,
		C.CString(src),
		C.uint(len(src)),
	)
	rootNode := C.ts_tree_root_node(tree)
	fmt.Printf("%s\n", C.GoString(C.ts_node_string(rootNode)))
}

const src = `
package main

func main() {
}
`
