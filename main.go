/*
	https://medium.com/justforfunc/understanding-go-programs-with-go-parser-c4e88a6edb87
*/

// zparser parses the go programs in the given paths and prints
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

//!PARANOIDBROCCOLI_PRIVATE_TO_FILE

//------------------------------------------------------------------------------

type infosFile struct {
	stmts                   []stmt
	globalPrivateToFileDecl map[string]bool
}

//------------------------------------------------------------------------------

func main() {
	if len(os.Args) == 1 {
		usage()
	}

	var pkgDir string

	if os.Args[1] == "-dir" {
		if len(os.Args) != 3 {
			usage()
		}
		pkgDir = os.Args[2]
	} else if os.Args[1] == "-pkg" {
		if len(os.Args) != 3 {
			usage()
		}
		pkgDir = os.Getenv("GOPATH") + "/src/" + os.Args[2]
	} else {
		usage()
	}

	_, err := os.Stat(pkgDir)
	if os.IsNotExist(err) {
		panic("Folder " + pkgDir + " does not exist.")
	}

	var files []string
	files, err = filepath.Glob(pkgDir + "/*.go")
	if err != nil {
		panic(err)
	}
	if len(files) == 0 {
		panic("No sources files in " + pkgDir)
	}

	var infosByFile = make(map[string]infosFile)
	for _, filepath := range files {
		fmt.Println("Scanning file " + filepath + " ...")
		var fs = token.NewFileSet()
		var err error
		bytes, err = ioutil.ReadFile(filepath)
		if err != nil {
			panic("could not parse file:" + err.Error())
		}
		f, err := parser.ParseFile(fs, filepath, nil, parser.ParseComments|parser.AllErrors)
		if err != nil {
			panic("could not parse file:" + err.Error())
		}
		stmts = make([]stmt, 0)
		var globalPrivateToFileDecl = make(map[string]bool)
		v := newVisitor(f)
		ast.Walk(&v, f)
		//fmt.Printf("\n\n\n%+v\n", stmts)

		for _, stmt := range stmts {
			if stmt.privateToFile && stmt.isGlobal {
				globalPrivateToFileDecl[stmt.name] = true
			}
		}
		//fmt.Printf("\n\n\nglobalPrivateToFileDecl: %+v\n", globalPrivateToFileDecl)

		infosByFile[filepath] = infosFile{
			stmts:                   stmts,
			globalPrivateToFileDecl: globalPrivateToFileDecl,
		}
	}

	fmt.Println("Checking ...")

	for filepath1, infosFile := range infosByFile {
		for _, stmt := range infosFile.stmts {
			for filepath2, infosFile2 := range infosByFile {
				if filepath1 != filepath2 {
					//fmt.Printf("%s %+v\n", stmt.name, infosFile2.globalPrivateToFileDecl)
					if _, ok := infosFile2.globalPrivateToFileDecl[stmt.name]; ok {
						fmt.Println("FATAL: cannot use " + stmt.name + " in " + filepath1 + ", declared with " + constPrivateToFileComment + " in " + filepath2)
						os.Exit(1)
					}
				}
			}
		}
	}

	//------------
	// test below

	testVar = true
	var tata testType1
	tata.machin = 1
	testFunction()
}

//------------------------------------------------------------------------------

//!PB_PRIVATE_TO_FILE
func usage() {
	fmt.Println("usage: " + os.Args[0] + " [ -dir <dir> | -pkg <pkg> ]")
	os.Exit(1)
}

//------------------------------------------------------------------------------
