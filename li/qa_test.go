package li

import (
	"go/ast"
	"go/types"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestQA(
	t *testing.T,
) {
	pkgs, err := packages.Load(
		&packages.Config{
			Mode: packages.NeedTypesInfo |
				packages.NeedFiles |
				packages.NeedSyntax |
				packages.NeedTypes |
				packages.NeedName,
		},
	)
	ce(err)
	if packages.PrintErrors(pkgs) > 0 {
		return
	}

	// Provide type
	var defType types.Type
	for _, pkg := range pkgs {
		if pkg.Name != "li" {
			continue
		}
		defType = pkg.Types.Scope().Lookup("Provide").(*types.TypeName).Type()
	}
	if defType == nil {
		panic("def type not found")
	}

	// find unused injected params
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {

				fnDecl, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				obj := pkg.TypesInfo.Defs[fnDecl.Name]
				sig := obj.Type().(*types.Signature)
				recv := sig.Recv()
				if recv == nil {
					continue
				}
				if recv.Type() != defType {
					continue
				}

				args := fnDecl.Type.Params.List
				type Obj struct {
					Ident *ast.Ident
					Obj   types.Object
				}
				objs := make(map[types.Object]Obj)
				for _, arg := range args {
					for _, ident := range arg.Names {
						obj := pkg.TypesInfo.Defs[ident]
						objs[obj] = Obj{
							Ident: ident,
							Obj:   obj,
						}
					}
				}

				ast.Inspect(fnDecl.Body, func(node ast.Node) bool {
					ident, ok := node.(*ast.Ident)
					if !ok {
						return true
					}
					obj := pkg.TypesInfo.Uses[ident]
					if obj == nil {
						return true
					}
					delete(objs, obj)
					return true
				})

				for _, obj := range objs {
					if obj.Ident.Name == "_" {
						continue
					}
					pos := pkg.Fset.Position(obj.Ident.Pos())
					pt("unused param: %s at %s:%d\n",
						obj.Ident.Name,
						pos.Filename,
						pos.Line,
					)
				}

			}
		}
	}

}
