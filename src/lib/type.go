package main

import (
	"go/ast"
	"go/doc"
	"go/token"
)

func objectFromType(t *doc.Type, fset *token.FileSet) *GoObject {
	file := fset.File(t.Decl.Pos())
	obj := GoObject{
		Name:       t.Name,
		Type:       TYPE,
		Doc:        t.Doc,
		Definition: valueDefinition(t.Decl, fset),
		Items:      make([]*GoObject, 0),
		Location:   file.Name(),
		StartLine:  file.Line(t.Decl.Pos()),
		EndLine:    file.Line(t.Decl.End()),
	}

	// append consts defined on the type (enums)
	for _, c := range t.Consts {
		for _, name := range c.Names {
			obj.Items = append(obj.Items,
				objectFromValue(
					&valueEntry{
						doc:       c,
						name:      name,
						valueType: CONST,
					},
					fset,
				),
			)
		}
	}

	// append methods defined on the type
	for _, m := range t.Methods {
		item := objectFromFunc(m, fset)
		item.Type = METH
		obj.Items = append(obj.Items, item)
	}

	for _, spec := range t.Decl.Specs {
		if ts, ok := spec.(*ast.TypeSpec); ok {
			if st, ok := ts.Type.(*ast.StructType); ok {
				obj.Type = STRUCT
				for _, field := range st.Fields.List {
					obj.Items = append(obj.Items, &GoObject{
						Name:      field.Names[0].Name,
						Type:      nodeString(fset, field.Type),
						Doc:       field.Doc.Text(),
						Tag:       parseTag(field.Tag),
						Items:     make([]*GoObject, 0),
						Location:  file.Name(),
						StartLine: file.Line(field.Pos()),
						EndLine:   file.Line(field.End()),
					})
				}
			}
		}
	}
	return &obj
}
