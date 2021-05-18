package node

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
)

//------------------------------------------------------------------------------

var fileBytes []byte

//------------------------------------------------------------------------------

// ReadFile parses a Go source file and returns informations about it.
func ReadFile(filepath string) (rootNode *Node, fileConstants []ConstValue, fileBuffer []byte, packageName string) {
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
	rootNode = v.node
	fileConstants = retrieveAllConstants(v.node)
	fileBuffer = fileBytes
	if len(rootNode.Children) > 0 && rootNode.Children[0].TypeStr == "File" &&
		len(rootNode.Children[0].Children) > 0 && rootNode.Children[0].Children[0].TypeStr == "Ident" {
		packageName = rootNode.Children[0].Children[0].Name
	}
	return
}

//------------------------------------------------------------------------------
