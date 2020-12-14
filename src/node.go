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

//------------------------------------------------------------------------------

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

//------------------------------------------------------------------------------

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
