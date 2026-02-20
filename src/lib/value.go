package main

import (
	"go/ast"
	"go/doc"
	"go/token"
)

// valueEntry pairs a *doc.Value with the specific name from the block being indexed.
// This is needed for multi-declaration blocks (var/const (...)) where a single
// *doc.Value covers several names but we need to resolve each one individually.
type valueEntry struct {
	doc       *doc.Value
	name      string
	valueType string
}

func valueDefinition(v *ast.GenDecl, fset *token.FileSet) string {
	def := ast.GenDecl{
		Tok:    v.Tok,
		Specs:  v.Specs,
		TokPos: v.TokPos,
		Lparen: v.Lparen,
		Rparen: v.Rparen,
		Doc:    nil,
	}
	return nodeString(fset, &def)
}

func valueSpecDefinition(v *ast.ValueSpec, fset *token.FileSet) string {
	def := ast.ValueSpec{
		Names:  v.Names,
		Type:   v.Type,
		Values: v.Values,
		Doc:    nil,
	}
	return nodeString(fset, &def)
}

func objectFromValue(v *valueEntry, fset *token.FileSet) *GoObject {
	file := fset.File(v.doc.Decl.Pos())

	obj := GoObject{
		Name:       v.name,
		Type:       v.valueType,
		Doc:        v.doc.Doc,
		Definition: valueDefinition(v.doc.Decl, fset),
		Items:      make([]*GoObject, 0),
		Location:   file.Name(),
		StartLine:  file.Line(v.doc.Decl.Pos()),
		EndLine:    file.Line(v.doc.Decl.End()),
	}

	// find the specific spec for this name (case of multi declaration)
	// use the spec's own positions so StartLine/EndLine reflect the single entry
outer:
	for _, spec := range v.doc.Decl.Specs {
		vs := spec.(*ast.ValueSpec)
		for _, ident := range vs.Names {
			// we find a match with the initial name
			if ident.Name == v.name {
				obj.StartLine = file.Line(vs.Pos())
				obj.EndLine = file.Line(vs.End())
				obj.Definition = valueSpecDefinition(vs, fset)
				if vs.Doc != nil {
					obj.Doc = vs.Doc.Text()
				}
				break outer
			}
		}
	}
	return &obj
}
