package main

import (
	"go/ast"
	"go/doc"
	"go/token"
)

func functionDefinition(f *ast.FuncDecl, fset *token.FileSet) string {
	def := ast.FuncDecl{
		Name: f.Name,
		Type: f.Type,
		Recv: f.Recv,
		Doc:  nil,
		Body: nil,
	}
	return nodeString(fset, &def)
}

func objectFromFunc(f *doc.Func, fset *token.FileSet) *GoObject {
	file := fset.File(f.Decl.Pos())
	obj := GoObject{
		Name:       f.Name,
		Type:       FUNC,
		Doc:        f.Doc,
		Definition: functionDefinition(f.Decl, fset),
		Body:       nodeString(fset, f.Decl),
		Items:      make([]*GoObject, 0),
		Location:   file.Name(),
		StartLine:  file.Line(f.Decl.Pos()),
		EndLine:    file.Line(f.Decl.End()),
	}
	// loop over the attributes of the function
	for _, param := range f.Decl.Type.Params.List {
		obj.Items = append(obj.Items,
			&GoObject{
				Name:      param.Names[0].Name,
				Type:      nodeString(fset, param.Type),
				Doc:       param.Doc.Text(),
				Items:     make([]*GoObject, 0),
				Location:  file.Name(),
				StartLine: file.Line(param.Pos()),
				EndLine:   file.Line(param.End()),
			},
		)
	}

	return &obj
}
