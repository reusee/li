package treesitter

/*
#include <tree_sitter/api.h>
TSLanguage *tree_sitter_go();
*/
import "C"
import (
	"fmt"
	"testing"
)

func runTest(t *testing.T) {
	parser := C.ts_parser_new()
	defer C.ts_parser_delete(parser)
	C.ts_parser_set_language(parser, C.tree_sitter_go())
	tree := C.ts_parser_parse_string(
		parser,
		nil,
		C.CString(src),
		C.uint(len(src)),
	)
	defer C.ts_tree_delete(tree)
	rootNode := C.ts_tree_root_node(tree)
	fmt.Printf("%s\n", C.GoString(C.ts_node_string(rootNode)))
}

const src = `
package main

func main() {
}
`
