package main

import "C"
import (
	"bytes"
	"encoding/json"
	"go/ast"
	"go/doc"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path"
	"strings"
	"unsafe"
)

const (
	TYPE   = "type"
	FUNC   = "function"
	METH   = "method"
	CONST  = "const"
	VAR    = "var"
	STRUCT = "struct"
	PKG    = "package"
)

// GoObject is a struct representing a Go object (package, type, function, variable, constant)
// with its documentation and content.
type GoObject struct {
	Path       string            `json:"path"` // object identifier
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Doc        string            `json:"doc"`
	Tag        map[string]string `json:"tag"`
	Definition string            `json:"definition"`
	Body       string            `json:"body"`
	Items      []*GoObject       `json:"items"`
	Location   string            `json:"location"`
	StartLine  int               `json:"start_line"`
	EndLine    int               `json:"end_line"`
}

func (o *GoObject) SetLocationPrefix(prefix string) {
	o.Location = path.Join(prefix, o.Location)
	for _, item := range o.Items {
		item.SetLocationPrefix(prefix)
	}
}

func collectFiles(directory string) (*token.FileSet, []*ast.File, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, nil, err
	}

	fset := token.NewFileSet()
	files := make([]*ast.File, 0)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		f := path.Join(directory, info.Name())

		if strings.HasSuffix(f, ".go") {
			src, err := os.ReadFile(f) //#nosec G304 -- If the file is not valid, that will be triggered by parser.ParseFile
			if err != nil {
				continue
			}
			// parse
			af, err := parser.ParseFile(fset, info.Name(), string(src), parser.ParseComments)
			if err != nil {
				continue
			}
			files = append(files, af)
		}
	}
	return fset, files, nil
}

func rootObject(pkg *doc.Package) (*GoObject, map[string]any) {
	index := make(map[string]any)
	for _, t := range pkg.Types {
		index[t.Name] = t
		// when defining a new type we may use it as an enum;
		// its instances are attached to t.Consts by go/doc
		for _, m := range t.Consts {
			for _, name := range m.Names {
				index[name] = &valueEntry{doc: m, name: name, valueType: CONST}
			}
		}
	}
	for _, f := range pkg.Funcs {
		index[f.Name] = f
	}
	for _, c := range pkg.Consts {
		// case of multi declaration, we index each name separately
		for _, name := range c.Names {
			index[name] = &valueEntry{doc: c, name: name, valueType: CONST}
		}
	}
	for _, v := range pkg.Vars {
		// case of multi declaration, we index each name separately
		for _, name := range v.Names {
			index[name] = &valueEntry{doc: v, name: name, valueType: VAR}
		}
	}

	return &GoObject{
		Name:       pkg.Name,
		Type:       "package",
		Doc:        pkg.Doc,
		Definition: "",
		Items:      make([]*GoObject, 0),
		Location:   path.Base(pkg.Filenames[0]),
	}, index
}

func nodeString(fset *token.FileSet, node ast.Node) string {
	var buf bytes.Buffer
	format.Node(&buf, fset, node)
	return buf.String()
}

func parseTag(tag *ast.BasicLit) map[string]string {
	if tag == nil || len(tag.Value) < 2 {
		return nil
	}
	s := strings.Trim(tag.Value, " `") // strip spaces & backticks
	result := map[string]string{}
	for s != "" {
		// skip spaces
		for len(s) > 0 && s[0] == ' ' {
			s = s[1:]
		}
		if s == "" {
			break
		}
		// read key (up to ':')
		i := 0
		for i < len(s) && s[i] != ':' && s[i] != '"' && s[i] != ' ' {
			i++
		}
		if i == 0 || i >= len(s)-1 || s[i] != ':' || s[i+1] != '"' {
			break
		}
		key := s[:i]
		s = s[i+2:] // skip :"

		// read quoted value
		i = 0
		for i < len(s) && s[i] != '"' {
			if s[i] == '\\' {
				i++ // skip escaped char
			}
			i++
		}
		if i >= len(s) {
			break
		}
		result[key] = s[:i]
		s = s[i+1:] // skip closing "
	}
	return result
}

//export collect
func collect(cImportPath *C.char, cDir *C.char, cObject *C.char, outBuf *C.char, outSize C.int) C.int {
	importPath := C.GoString(cImportPath) // like github.com/user/project0
	dir := C.GoString(cDir)               // like /home/user/projects/project0
	object := C.GoString(cObject)         // like "github.com/user/project0/pkg/subpkg.MyType" or "github.com/user/project0/pkg/subpkg.MyFunc"

	// remove the prepended import path
	object = strings.TrimPrefix(object, importPath)
	object = strings.Trim(object, "/")
	// now object is like "pkg/subpkg.MyType" or "pkg/subpkg.MyFunc"
	// split module and object
	// if !strings.Contains(object, ".") {
	// 	// prepend a dot
	// 	object = fmt.Sprintf(".%s", object) // like ".MyType" or ".MyFunc"
	// }
	parts := strings.Split(object, ".")
	mod := []string{}
	obj := parts[0]
	if len(parts) == 2 {
		mod = strings.Split(parts[0], "/") // like ["pkg", "subpkg"]
		obj = parts[1]                     // like "MyType" or "MyFunc"
	}
	// like "MyType" or "MyFunc"

	// rebuild the full path (os agnostic) to the module containing the object
	location := path.Join(append([]string{dir}, mod...)...)

	// read the module files
	fset, files, err := collectFiles(location)
	if err != nil {
		return -2
	}
	// Build package doc (PreserveAST to access function body notably)
	pkg, err := doc.NewFromFiles(fset, files, dir, doc.PreserveAST|doc.AllDecls|doc.AllMethods)
	if err != nil {
		return -3
	}

	// index the package
	_, index := rootObject(pkg)

	// look for the object
	var result *GoObject
	switch o := index[obj].(type) {
	case *doc.Type:
		result = objectFromType(o, fset)
	case *doc.Func:
		result = objectFromFunc(o, fset)
	case *valueEntry:
		result = objectFromValue(o, fset)
	default:
		return -5
	}

	result.Path = object
	result.SetLocationPrefix(path.Join(mod...))

	data, err := json.Marshal(result)
	if err != nil || C.int(len(data)+1) > outSize {
		return -6
	}
	buf := unsafe.Slice((*byte)(unsafe.Pointer(outBuf)), int(outSize))
	copy(buf, data)
	buf[len(data)] = 0
	return C.int(len(data))
}

func main() {}
