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
	"path/filepath"
)

var bytes []byte
var stmts []stmt
var passedPrivateToFileMarker bool

const constPrivateToFileMarkerName = "privateToFileMarker"

type infosFile struct {
	stmts                   []stmt
	globalPrivateToFileDecl map[string]bool
}

//------------------------------------------------------------------------------

func main() {
	var files, err = filepath.Glob("*.go")
	if err != nil {
		panic(err)
	}
	var infosByFile = make(map[string]infosFile)
	for _, filepath := range files {
		var fs = token.NewFileSet()
		var err error
		bytes, err = ioutil.ReadFile(filepath)
		if err != nil {
			panic("could not parse file:" + err.Error())
		}
		f, err := parser.ParseFile(fs, filepath, nil, parser.AllErrors)
		if err != nil {
			panic("could not parse file:" + err.Error())
		}
		stmts = make([]stmt, 0)
		var globalPrivateToFileDecl = make(map[string]bool)
		passedPrivateToFileMarker = false
		v := newVisitor(f)
		ast.Walk(&v, f)
		//fmt.Printf("\n\n\n%+v\n", stmts)

		for _, stmt := range stmts {
			if stmt.afterPrivateToFileMarker && stmt.isGlobal {
				globalPrivateToFileDecl[stmt.name] = true
			}
		}

		infosByFile[filepath] = infosFile{
			stmts:                   stmts,
			globalPrivateToFileDecl: globalPrivateToFileDecl,
		}
	}

	for filepath1, infosFile := range infosByFile {
		for _, stmt := range infosFile.stmts {
			for filepath2, infosFile2 := range infosByFile {
				if filepath1 != filepath2 {
					//fmt.Printf("%s %+v\n", stmt.name, infosFile2.globalPrivateToFileDecl)
					if _, ok := infosFile2.globalPrivateToFileDecl[stmt.name]; ok {
						fmt.Println("FATAL: declaration in " + filepath1 + " cannot call " + stmt.name + " in " + filepath2 + " (after " + constPrivateToFileMarkerName + ")")
						os.Exit(1)
					}
				}
			}
		}
	}
	testFunction()
}
