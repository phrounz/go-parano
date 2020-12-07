package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

//------------------------------------------------------------------------------

const constDebugInfo = false

//------------------------------------------------------------------------------

var bytes []byte
var stmts []stmt
var privateToFileLevel int

const constPrivateToFileComment = "//!PB_PRIVATE_TO_FILE"

//------------------------------------------------------------------------------

type stmt struct {
	name          string
	isGlobal      bool
	privateToFile bool
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

	var isPrivateToFile = false
	if privateToFileLevel != 0 {
		if v.depthLevel < privateToFileLevel {
			privateToFileLevel = 0
		} else if v.depthLevel == privateToFileLevel {
			isPrivateToFile = true
		}
	}

	if constDebugInfo {
		fmt.Printf("%s%T [", strings.Repeat("\t", v.depthLevel), n)
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
			name:          string(stmtStr[0:i]),
			isGlobal:      v.depthLevel == 1,
			privateToFile: false,
		})
	case *ast.CommentGroup:
		if string(stmtStr) == constPrivateToFileComment {
			privateToFileLevel = v.depthLevel

			var name string
			switch d2 := (*v.father).(type) {
			case *ast.TypeSpec:
				name = d2.Name.Name
			case *ast.FuncDecl:
				name = d2.Name.Name
			case *ast.GenDecl:
				for _, spec := range d2.Specs {
					if value, ok := spec.(*ast.ValueSpec); ok {
						for _, valueName := range value.Names {
							if valueName.Name == "_" {
								continue
							}
							name = valueName.Name
						}
					}
				}
			default:
				privateToFileLevel = v.depthLevel
			}
			if name != "" {
				var s = stmt{
					name:          name,
					isGlobal:      v.depthLevel == 2,
					privateToFile: true,
				}
				stmts = append(stmts, s)
				//fmt.Printf("=====>PRIVATE TO FILE: %+v\n", s)
			}
		}

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
		stmts = append(stmts, stmt{
			name:          d.Name.Name,
			isGlobal:      v.depthLevel == 1,
			privateToFile: false,
		})
		if d.Recv != nil {
			v.localList(d.Recv.List)
		}
		v.localList(d.Type.Params.List)
		if d.Type.Results != nil {
			v.localList(d.Type.Results.List)
		}
	case *ast.TypeSpec:
		if isPrivateToFile {
			//fmt.Printf("=====>PRIVATE TO FILE (2): %+v\n", d.Name.Name)
		}
		stmts = append(stmts, stmt{
			name:          d.Name.Name,
			isGlobal:      v.depthLevel == 1,
			privateToFile: isPrivateToFile,
		})
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
					stmts = append(stmts, stmt{
						name:          name.Name,
						isGlobal:      v.pkgDecl[d],
						privateToFile: false,
					})
				}
			}
		}
	case *ast.Ident:
		stmts = append(stmts, stmt{
			name:          d.Name,
			isGlobal:      v.depthLevel == 1,
			privateToFile: false,
		})
	}

	isPrivateToFile = false

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
			name:          ident.Name,
			isGlobal:      false,
			privateToFile: false,
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

// func checkPublicAfterMarker(v visitor, name string) {
// 	if passedPrivateToFileMarker && v.depthLevel == 1 && unicode.IsUpper([]rune(name)[0]) {
// 		fmt.Println("FATAL: declaration " + name + " after marker " + constPrivateToFileMarkerName)
// 		os.Exit(1)
// 	}
// }

//!PB_PRIVATE_TO_FILE
var testVar bool

//!PB_PRIVATE_TO_FILE
func testFunction() {
}

//!PB_PRIVATE_TO_FILE
type testType1 struct {
	machin int
}
type testType2 int

// type TestStruct struct {
// }

// var TestVariable bool
