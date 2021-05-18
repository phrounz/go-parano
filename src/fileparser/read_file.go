package fileparser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
)

//------------------------------------------------------------------------------

var fileBytes []byte

//------------------------------------------------------------------------------

// FileInfo is the output of ReadFile().
type FileInfo struct {
	PackageName   string
	FileBuffer    []byte
	RootNode      *Node
	FileConstants []ConstValue
}

//------------------------------------------------------------------------------

// ReadFile parses a Go source file and returns informations about it.
func ReadFile(filepath string) FileInfo {
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
	var fi = FileInfo{
		FileBuffer:    fileBytes,
		RootNode:      v.node,
		FileConstants: retrieveAllConstants(v.node),
	}
	if len(fi.RootNode.Children) > 0 && fi.RootNode.Children[0].TypeStr == "File" &&
		len(fi.RootNode.Children[0].Children) > 0 && fi.RootNode.Children[0].Children[0].TypeStr == "Ident" {
		fi.PackageName = fi.RootNode.Children[0].Children[0].Name
	}
	return fi
}

//------------------------------------------------------------------------------
