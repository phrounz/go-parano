package node

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

//------------------------------------------------------------------------------

func ReadFile(filepath string) (*Node, []byte) {
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
	return v.node, fileBytes
}

//------------------------------------------------------------------------------

type Node struct {
	Bytes           string
	Name            string
	BytesIndexBegin int
	BytesIndexEnd   int
	TypeStr         string
	nodeObj         *ast.Node
	DepthLevel      int
	Father          *Node
	Children        []*Node
}

//------------------------------------------------------------------------------

func (n *Node) NextNode() *Node {
	var j = -1
	for i, n2 := range n.Father.Children {
		if n2 == n {
			j = i
		} else if j != -1 && i == j+1 {
			return n2
		}
	}
	return nil
}

//------------------------------------------------------------------------------

func (n *Node) Display() {
	fmt.Printf("%s%s [", strings.Repeat("\t", n.DepthLevel), n.TypeStr)
	if n.Name != "" {
		fmt.Printf("[name=%s]", n.Name)
	}
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

//------------------------------------------------------------------------------

func (n *Node) Visit(fnCall func(*Node)) {
	fnCall(n)
	for _, child := range n.Children {
		child.Visit(fnCall)
	}
}

//------------------------------------------------------------------------------
