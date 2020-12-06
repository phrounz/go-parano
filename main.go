// parser parses the go programs in the given paths and prints
// the top five most common names of local variables and variables
// defined at package level.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strings"
	"unicode"
)

var bytes []byte
var items []item
var passedPrivateToFileMarker bool

const constMarkerName = "privateToFileMarker"

func main() {
	fs := token.NewFileSet()

	var err error
	bytes, err = ioutil.ReadFile("main.go")
	if err != nil {
		panic("could not parse file:" + err.Error())
	}
	f, err := parser.ParseFile(fs, "main.go", nil, parser.AllErrors)
	if err != nil {
		panic("could not parse file:" + err.Error())
	}
	items = make([]item, 0)
	v := newVisitor(f)
	ast.Walk(&v, f)
	fmt.Printf("\n\n\n%+v\n", items)
}

type item struct {
	name     string
	isGlobal bool
}

type visitor struct {
	pkgDecl    map[*ast.GenDecl]bool
	depthLevel int
}

func newVisitor(f *ast.File) visitor {
	decls := make(map[*ast.GenDecl]bool)
	for _, decl := range f.Decls {
		if v, ok := decl.(*ast.GenDecl); ok {
			decls[v] = true
		}
	}

	return visitor{
		pkgDecl: decls,
	}
}

//my comment

var privateToFileMarker bool

func (v visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}

	fmt.Printf("%s%T - [", strings.Repeat("\t", v.depthLevel), n)
	for i, b := range bytes[n.Pos()-1 : n.End()-1] {
		if b == '\n' {
			if i < len(bytes[n.Pos()-1:n.End()-1]) {
				fmt.Printf("[...]")
			}
			break
		}
		fmt.Printf("%c", b)
	}
	fmt.Printf("]\n")

	switch d := n.(type) {
	case *ast.AssignStmt:
		if d.Tok != token.DEFINE {
			return v
		}
		for _, name := range d.Lhs {
			v.local(name)
		}
	case *ast.RangeStmt:
		v.local(d.Key)
		v.local(d.Value)
	case *ast.FuncDecl:
		fmt.Printf("=====> %s\n", d.Name)
		if passedPrivateToFileMarker && v.depthLevel == 1 && isPublic(d.Name) {
			fmt.Println("FATAL: declaration " + d.Name.String() + " after marker " + constMarkerName)
			os.Exit(1)
		}
		items = append(items, item{
			name:     d.Name.String(),
			isGlobal: v.depthLevel == 1,
		})
		if d.Recv != nil {
			v.localList(d.Recv.List)
		}
		v.localList(d.Type.Params.List)
		if d.Type.Results != nil {
			v.localList(d.Type.Results.List)
		}
	case *ast.GenDecl:
		if d.Tok != token.VAR {
			return v
		}
		for _, spec := range d.Specs {
			if value, ok := spec.(*ast.ValueSpec); ok {
				for _, name := range value.Names {
					if name.Name == "_" {
						continue
					}
					if name.Name == constMarkerName {
/*
	https://medium.com/justforfunc/understanding-go-programs-with-go-parser-c4e88a6edb87
*/

// parser parses the go programs in the given paths and prints
// the top five most common names of local variables and variables
// defined at package level.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strings"
	"unicode"
)

var bytes []byte
var items []item
var passedPrivateToFileMarker bool

const constMarkerName = "privateToFileMarker"

func main() {
	fs := token.NewFileSet()

	var err error
	bytes, err = ioutil.ReadFile("main.go")
	if err != nil {
		panic("could not parse file:" + err.Error())
	}
	f, err := parser.ParseFile(fs, "main.go", nil, parser.AllErrors)
	if err != nil {
		panic("could not parse file:" + err.Error())
	}
	items = make([]item, 0)
	v := newVisitor(f)
	ast.Walk(&v, f)
	fmt.Printf("\n\n\n%+v\n", items)
}

type item struct {
	name     string
	isGlobal bool
}

type visitor struct {
	pkgDecl    map[*ast.GenDecl]bool
	depthLevel int
}

func newVisitor(f *ast.File) visitor {
	decls := make(map[*ast.GenDecl]bool)
	for _, decl := range f.Decls {
		if v, ok := decl.(*ast.GenDecl); ok {
			decls[v] = true
		}
	}

	return visitor{
		pkgDecl: decls,
	}
}

//my comment

var privateToFileMarker bool

func (v visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}

	fmt.Printf("%s%T - [", strings.Repeat("\t", v.depthLevel), n)
	for i, b := range bytes[n.Pos()-1 : n.End()-1] {
		if b == '\n' {
			if i < len(bytes[n.Pos()-1:n.End()-1]) {
				fmt.Printf("[...]")
			}
			break
		}
		fmt.Printf("%c", b)
	}
	fmt.Printf("]\n")

	switch d := n.(type) {
	case *ast.AssignStmt:
		if d.Tok != token.DEFINE {
			return v
		}
		for _, name := range d.Lhs {
			v.local(name)
		}
	case *ast.RangeStmt:
		v.local(d.Key)
		v.local(d.Value)
	case *ast.FuncDecl:
		fmt.Printf("=====> %s\n", d.Name)
		if passedPrivateToFileMarker && v.depthLevel == 1 && isPublic(d.Name) {
			fmt.Println("FATAL: declaration " + d.Name.String() + " after marker " + constMarkerName)
			os.Exit(1)
		}
		items = append(items, item{
			name:     d.Name.String(),
			isGlobal: v.depthLevel == 1,
		})
		if d.Recv != nil {
			v.localList(d.Recv.List)
		}
		v.localList(d.Type.Params.List)
		if d.Type.Results != nil {
			v.localList(d.Type.Results.List)
		}
	case *ast.GenDecl:
		if d.Tok != token.VAR {
			return v
		}
		for _, spec := range d.Specs {
			if value, ok := spec.(*ast.ValueSpec); ok {
				for _, name := range value.Names {
					if name.Name == "_" {
						continue
					}
					if name.Name == constMarkerName {
						passedPrivateToFileMarker = true
					} else if passedPrivateToFileMarker && isPublic(name) {
						fmt.Println("FATAL: declaration " + name.Name + " after marker " + constMarkerName)
						os.Exit(1)
					}
					items = append(items, item{
						name:     name.Name,
						isGlobal: v.pkgDecl[d],
					})
				}
			}
		}
	}

	return visitor{depthLevel: v.depthLevel + 1}
}

func isPublic(name *ast.Ident) bool {
	return unicode.IsUpper([]rune(name.Name)[0])
}

var VariableTest bool

func (v *visitor) local(n ast.Node) {
	ident, ok := n.(*ast.Ident)
	if !ok {
		return
	}
	if ident.Name == "_" || ident.Name == "" {
		return
	}
	if ident.Obj != nil && ident.Obj.Pos() == ident.Pos() {
		//fmt.Printf("test;%+v\n", v.items)
		items = append(items, item{
			name:     ident.Name,
			isGlobal: false,
		})
	}
}

func (v *visitor) localList(fs []*ast.Field) {
	for _, f := range fs {
		for _, name := range f.Names {
			v.local(name)
		}
	}
}
