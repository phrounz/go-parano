package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"strings"
	"unicode"
)

//------------------------------------------------------------------------------

const constDebugInfo = false

//------------------------------------------------------------------------------

type stmt struct {
	name                     string
	isGlobal                 bool
	afterPrivateToFileMarker bool
}

//------------------------------------------------------------------------------

type visitor struct {
	father     *ast.Node
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

func (v visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}

	var stmtStr = bytes[n.Pos()-1 : n.End()-1]

	if constDebugInfo {
		fmt.Printf("%s%T - [", strings.Repeat("\t", v.depthLevel), n)
		for i, b := range stmtStr {
			if b == '\n' {
				if i < len(stmtStr) {
					fmt.Printf("[...]")
				}
				break
			}
			fmt.Printf("%c", b)
		}
		fmt.Printf("]\n")
	}

	switch d := n.(type) {
	case *ast.CallExpr:
		var i = strings.Index(string(stmtStr), "(")
		stmts = append(stmts, stmt{
			name:                     string(stmtStr[0:i]),
			isGlobal:                 v.depthLevel == 1,
			afterPrivateToFileMarker: passedPrivateToFileMarker,
		})
	/*case *ast.Ident:
	switch (*v.father).(type) {
	case *ast.CallExpr:
		fmt.Printf("=====>FUNCTION CALL: %s\n", d.Name)
		stmts = append(stmts, stmt{
			name:                     d.Name,
			isGlobal:                 v.depthLevel == 1,
			afterPrivateToFileMarker: passedPrivateToFileMarker,
		})
	}*/
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
		//fmt.Printf("=====> %s\n", d.Name)
		checkPublicAfterMarker(v, d.Name.Name)
		stmts = append(stmts, stmt{
			name:                     d.Name.Name,
			isGlobal:                 v.depthLevel == 1,
			afterPrivateToFileMarker: passedPrivateToFileMarker,
		})
		if d.Recv != nil {
			v.localList(d.Recv.List)
		}
		v.localList(d.Type.Params.List)
		if d.Type.Results != nil {
			v.localList(d.Type.Results.List)
		}
	case *ast.TypeSpec:
		stmts = append(stmts, stmt{
			name:                     d.Name.Name,
			isGlobal:                 v.depthLevel == 1,
			afterPrivateToFileMarker: passedPrivateToFileMarker,
		})
		checkPublicAfterMarker(v, d.Name.Name)
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
					if name.Name == constPrivateToFileMarkerName {
						passedPrivateToFileMarker = true
					}
					checkPublicAfterMarker(v, name.Name)
					stmts = append(stmts, stmt{
						name:                     name.Name,
						isGlobal:                 v.pkgDecl[d],
						afterPrivateToFileMarker: passedPrivateToFileMarker,
					})
				}
			}
		}
	}

	return visitor{father: &n, depthLevel: v.depthLevel + 1}
}

func (v *visitor) local(n ast.Node) {
	ident, ok := n.(*ast.Ident)
	if !ok {
		return
	}
	if ident.Name == "_" || ident.Name == "" {
		return
	}
	if ident.Obj != nil && ident.Obj.Pos() == ident.Pos() {
		stmts = append(stmts, stmt{
			name:                     ident.Name,
			isGlobal:                 false,
			afterPrivateToFileMarker: passedPrivateToFileMarker,
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

func checkPublicAfterMarker(v visitor, name string) {
	if passedPrivateToFileMarker && v.depthLevel == 1 && unicode.IsUpper([]rune(name)[0]) {
		fmt.Println("FATAL: declaration " + name + " after marker " + constPrivateToFileMarkerName)
		os.Exit(1)
	}
}

var privateToFileMarker bool

func testFunction() {

}

// type TestStruct struct {
// }

// var TestVariable bool
