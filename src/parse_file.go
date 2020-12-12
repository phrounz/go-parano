package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"strings"
)

//------------------------------------------------------------------------------

var fileBytes []byte

var debugInfo = false

//------------------------------------------------------------------------------

type visitor struct {
	father     *visitor
	node       *node
	pkgDecl    map[*ast.GenDecl]bool
	depthLevel int
}

type node struct {
	bytes           string
	name            string
	bytesIndexBegin int
	bytesIndexEnd   int
	typeStr         string
	nodeObj         *ast.Node
	depthLevel      int
	father          *node
	children        []*node
}

func (n *node) nextNode() *node {
	var j = -1
	for i, n2 := range n.father.children {
		if n2 == n {
			j = i
		} else if j != -1 && i == j+1 {
			return n2
		}
	}
	return nil
}

func (n *node) display() {
	fmt.Printf("%s%s [", strings.Repeat("\t", n.depthLevel), n.typeStr)
	if n.name != "" {
		fmt.Printf("[name=%s]", n.name)
	}
	for i, b := range n.bytes {
		if b == '\n' {
			if i < len(n.bytes) {
				fmt.Printf("[...]")
			}
			break
		}
		fmt.Printf("%c", b)
	}
	fmt.Printf("]\n")
}

//------------------------------------------------------------------------------

func (n *node) visit(fnCall func(*node)) {
	if debugInfo {
		fmt.Printf("==> ")
		n.display()
	}
	fnCall(n)
	for _, child := range n.children {
		child.visit(fnCall)
	}
}

//------------------------------------------------------------------------------

func readFile(filepath string) *node {
	var fs = token.NewFileSet()
	var err error
	fileBytes, err = ioutil.ReadFile(filepath)
	if err != nil {
		panic("could not parse file:" + err.Error())
	}
	f, err := parser.ParseFile(fs, filepath, nil, parser.ParseComments|parser.AllErrors)
	if err != nil {
		panic("could not parse file:" + err.Error())
	}
	v := newVisitor(f)
	ast.Walk(&v, f)
	return v.node
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
		node:       &node{depthLevel: 0},
	}
}

//------------------------------------------------------------------------------

func (v visitor) Visit(nodeObj ast.Node) ast.Visitor {

	if nodeObj == nil {
		return nil
	}

	var n = &node{
		depthLevel:      v.depthLevel + 1,
		bytes:           string(fileBytes[nodeObj.Pos()-1 : nodeObj.End()-1]),
		bytesIndexBegin: int(nodeObj.Pos() - 1),
		bytesIndexEnd:   int(nodeObj.End() - 1),
		nodeObj:         &nodeObj,
		father:          v.node,
		typeStr:         "(unknown)",
	}

	v.node.children = append(v.node.children, n)

	if debugInfo {
		fmt.Printf("%s%T [", strings.Repeat("\t", n.depthLevel), nodeObj)
		for i, b := range n.bytes {
			if b == '\n' {
				if i < len(n.bytes) {
					fmt.Printf("[...]")
				}
				break
			}
			fmt.Printf("%c", b)
		}
		fmt.Printf("]\n")
	}

	switch d := nodeObj.(type) {
	case *ast.CallExpr:
		var i = strings.Index(n.bytes, "(")
		n.name = n.bytes[0:i]
		n.typeStr = "CallExpr"
	case *ast.CommentGroup:
		n.typeStr = "CommentGroup"
	case *ast.StructType:
		n.typeStr = "StructType"
	case *ast.FieldList:
		n.typeStr = "FieldList"
	case *ast.Field:
		n.typeStr = "Field"
	case *ast.ValueSpec:
		n.typeStr = "ValueSpec"
	case *ast.AssignStmt:
		n.typeStr = "AssignStmt"
	case *ast.IfStmt:
		n.typeStr = "IfStmt"
	case *ast.BinaryExpr:
		n.typeStr = "BinaryExpr"
	case *ast.KeyValueExpr:
		n.typeStr = "KeyValueExpr"
	case *ast.CompositeLit:
		n.typeStr = "CompositeLit"
	case *ast.RangeStmt:
		n.typeStr = "RangeStmt"
	case *ast.FuncDecl:
		n.name = d.Name.Name
		n.typeStr = "FuncDecl"
	case *ast.TypeSpec:
		n.name = d.Name.Name
		n.typeStr = "TypeSpec"
	case *ast.GenDecl:
		//fmt.Printf("##### %s\n", n.bytes)
		n.typeStr = "GenDecl"
		if d.Tok != token.VAR {
			return v
		}
		for _, spec := range d.Specs {
			if value, ok := spec.(*ast.ValueSpec); ok {
				for _, name := range value.Names {
					if name.Name == "_" {
						continue
					}
				}
			}
		}
	case *ast.Ident:
		n.name = n.bytes
		n.typeStr = "Ident"
	}

	return &visitor{
		father:     &v,
		depthLevel: n.depthLevel,
		node:       n,
	}
}

//------------------------------------------------------------------------------
