/*
	https://medium.com/justforfunc/understanding-go-programs-with-go-parser-c4e88a6edb87
*/
package node

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

//------------------------------------------------------------------------------

var DebugInfo = false

//------------------------------------------------------------------------------

type visitor struct {
	father     *visitor
	node       *Node
	pkgDecl    map[*ast.GenDecl]bool
	depthLevel int
}

//------------------------------------------------------------------------------

func newVisitor(f *ast.File) visitor {
	decls := make(map[*ast.GenDecl]bool)
	for _, decl := range f.Decls {
		if v, ok := decl.(*ast.GenDecl); ok {
			decls[v] = true
		}
	}
	return visitor{
		pkgDecl:    decls,
		depthLevel: 0,
		node:       &Node{DepthLevel: 0},
	}
}

//------------------------------------------------------------------------------

func (v visitor) Visit(nodeObj ast.Node) ast.Visitor {

	if nodeObj == nil {
		return nil
	}

	var n = &Node{
		DepthLevel:      v.depthLevel + 1,
		Bytes:           string(fileBytes[nodeObj.Pos()-1 : nodeObj.End()-1]),
		BytesIndexBegin: int(nodeObj.Pos() - 1),
		BytesIndexEnd:   int(nodeObj.End() - 1),
		nodeObj:         &nodeObj,
		Father:          v.node,
		TypeStr:         "(unknown)",
	}

	v.node.Children = append(v.node.Children, n)

	if DebugInfo {
		fmt.Printf("%s%T [", strings.Repeat("\t", n.DepthLevel), nodeObj)
		for i, b := range n.Bytes {
			if b == '\n' {
				if i < len(n.Bytes) {
					fmt.Printf("[...]")
				}
				break
			}
			fmt.Printf("%c", b)
		}
		fmt.Printf("]\n")
	}

	switch d := nodeObj.(type) {
	case *ast.File:
		n.TypeStr = "File"
	case *ast.CallExpr:
		var i = strings.Index(n.Bytes, "(")
		n.Name = n.Bytes[0:i]
		n.TypeStr = "CallExpr"
	case *ast.CommentGroup:
		n.TypeStr = "CommentGroup"
	case *ast.StructType:
		n.TypeStr = "StructType"
	case *ast.FieldList:
		n.TypeStr = "FieldList"
	case *ast.Field:
		n.TypeStr = "Field"
	case *ast.ValueSpec:
		n.TypeStr = "ValueSpec"
	case *ast.AssignStmt:
		n.TypeStr = "AssignStmt"
	case *ast.IfStmt:
		n.TypeStr = "IfStmt"
	case *ast.BinaryExpr:
		n.TypeStr = "BinaryExpr"
	case *ast.KeyValueExpr:
		n.TypeStr = "KeyValueExpr"
	case *ast.CompositeLit:
		n.TypeStr = "CompositeLit"
	case *ast.RangeStmt:
		n.TypeStr = "RangeStmt"
	case *ast.SelectorExpr:
		n.TypeStr = "SelectorExpr"
	case *ast.FuncDecl:
		n.Name = d.Name.Name
		n.TypeStr = "FuncDecl"
	case *ast.TypeSpec:
		n.Name = d.Name.Name
		n.TypeStr = "TypeSpec"
	case *ast.GenDecl:
		n.TypeStr = "GenDecl"
		if d.Tok != token.VAR {
			return v
		}
		// for _, spec := range d.Specs {
		// 	if value, ok := spec.(*ast.ValueSpec); ok {
		// 		for _, name := range value.Names {
		// 			if name.Name == "_" {
		// 				continue
		// 			}
		// 		}
		// 	}
		// }
	case *ast.Ident:
		n.Name = n.Bytes
		if n.Name == "" {
			panic("empty bytes for *ast.Ident ?")
		}
		n.TypeStr = "Ident"
	}

	return &visitor{
		father:     &v,
		depthLevel: n.DepthLevel,
		node:       n,
	}
}

//------------------------------------------------------------------------------
