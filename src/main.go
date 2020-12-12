/*
	https://medium.com/justforfunc/understanding-go-programs-with-go-parser-c4e88a6edb87
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//------------------------------------------------------------------------------

type infosFile struct {
	rootNode                 *node
	globalPrivateToFileDecl  map[string]bool
	exhaustiveFillingStructs map[string]map[string]bool
}

const constPrivateToFileComment = "//!PB_PRIVATE_TO_FILE"

const constExaustiveFilling = "//!PB_EXHAUSTIVE_FILLING"

var colorRed = "\033[31m"
var colorDefault = "\033[39m"

//------------------------------------------------------------------------------

func main() {

	if len(os.Args) == 1 {
		usage()
	}

	var noColor = flag.Bool("nocolor", false, "disable color")
	var pkgDir = flag.String("dir", "", "source directory")
	var pkg = flag.String("pkg", "", "source package")
	var verbose = flag.Bool("v", false, "verbose info")
	var debug = flag.Bool("debug", false, "debug info")
	flag.Parse()

	debugInfo = *debug

	if *pkg != "" {
		*pkgDir = os.Getenv("GOPATH") + "/src/" + *pkg
	}

	_, err := os.Stat(*pkgDir)
	if os.IsNotExist(err) {
		panic("Folder " + *pkgDir + " does not exist.")
	}

	var files []string
	files, err = filepath.Glob(*pkgDir + "/*.go")
	if err != nil {
		panic(err)
	}
	if len(files) == 0 {
		panic("No sources files in " + *pkgDir)
	}

	var infosByFile = make(map[string]infosFile)
	for _, filepath := range files {
		if debugInfo {
			fmt.Println("===============================")
		}
		if *verbose {
			fmt.Println("Scanning file " + filepath + " ...")
		}

		//----
		// first pass => load file and get nodes

		var rootNode = readFile(filepath)

		//----
		// second pass => get globalPrivateToFileDecl/exhaustiveFillingStructs

		var globalPrivateToFileDecl = make(map[string]bool)
		var exhaustiveFillingStructs = make(map[string]map[string]bool)

		rootNode.visit(func(n *node) {
			if n.typeStr == "CommentGroup" && n.bytes == constPrivateToFileComment && n.father != nil {
				if n.father.typeStr == "GenDecl" {
					for _, n2 := range n.father.children {
						if n2.typeStr == "ValueSpec" {
							if len(n2.children) >= 2 {
								var name = n2.children[0].bytes
								if debugInfo {
									fmt.Printf("CCCC >=%s <=\n", name)
								}
								globalPrivateToFileDecl[name] = true
								break
							}
						}
					}
				} else if n.father.typeStr == "FuncDecl" {
					if debugInfo {
						fmt.Printf("AAAA >=%s %s<=\n", n.father.name, n.father.typeStr)
					}
					globalPrivateToFileDecl[n.father.name] = true
				} else {
					var nextNode = n.nextNode()
					if nextNode != nil && nextNode.typeStr == "TypeSpec" {
						if debugInfo {
							fmt.Printf("BBBB >=%s %s<=\n", nextNode.name, nextNode.typeStr)
						}
						globalPrivateToFileDecl[nextNode.name] = true
					}
				}
			} else if n.typeStr == "CommentGroup" && n.bytes == constExaustiveFilling && n.father != nil {
				var nextNode = n.nextNode()
				if nextNode != nil && nextNode.typeStr == "TypeSpec" {
					if debugInfo {
						fmt.Printf("ZZZZ >=%s %s<=\n", nextNode.name, nextNode.typeStr)
					}
					var keys = make(map[string]bool)
					for _, child1 := range nextNode.children {
						if child1.typeStr == "StructType" {
							for _, child2 := range child1.children {
								if child2.typeStr == "FieldList" {
									for _, field := range child2.children {
										if field.children[0].typeStr == "Ident" {
											keys[field.children[0].name] = true
										}
									}
								}
							}
						}
					}
					exhaustiveFillingStructs[nextNode.name] = keys
				}
			}
		})

		if debugInfo {
			fmt.Printf("\n\n\nglobalPrivateToFileDecl: %+v\n", globalPrivateToFileDecl)
			fmt.Printf("\n\n\nexhaustiveFillingStructs: %+v\n", exhaustiveFillingStructs)
		}

		infosByFile[filepath] = infosFile{
			rootNode:                 rootNode,
			globalPrivateToFileDecl:  globalPrivateToFileDecl,
			exhaustiveFillingStructs: exhaustiveFillingStructs,
		}
	}

	if *verbose {
		fmt.Println("Checking ...")
	}

	if *noColor {
		colorRed = ""
		colorDefault = ""
	}

	//----
	// third pass => check

	var failedAtLeastOnce = false
	for filepath1, infosFile := range infosByFile {
		infosFile.rootNode.visit(func(n *node) {
			if n.name != "" {
				for filepath2, infosFile2 := range infosByFile {

					//----
					// globalPrivateToFileDecl

					if filepath1 != filepath2 {
						if _, ok := infosFile2.globalPrivateToFileDecl[n.name]; ok {
							notPass(fmt.Sprintf("cannot use %s in %s, declared with %s in %s",
								n.name, filepath1, constPrivateToFileComment, filepath2))
							failedAtLeastOnce = true
						}
					}

					//----
					// exhaustiveFillingStructs

					if keysStruct, ok := infosFile2.exhaustiveFillingStructs[n.name]; ok {
						if n.father.typeStr == "CompositeLit" {
							var keys = make(map[string]bool)
							for _, keyValue := range n.father.children {
								if keyValue.typeStr == "KeyValueExpr" {
									if keyValue.children[0].typeStr == "Ident" {
										keys[keyValue.children[0].name] = true
									}
								}
							}
							var missingKeys = make([]string, 0)
							for key := range keysStruct {
								if _, ok := keys[key]; !ok {
									missingKeys = append(missingKeys, key)
								}
							}

							if len(missingKeys) > 0 {
								notPass(fmt.Sprintf("missing key(s) %s in declaration of %s, type declared with %s in %s",
									strings.Join(missingKeys, ", "), n.name, constExaustiveFilling, filepath2))
								failedAtLeastOnce = true
							}
						}
					}
				}
			}
		})
	}
	if failedAtLeastOnce {
		os.Exit(1)
	}
	os.Exit(0)
}

//------------------------------------------------------------------------------

func notPass(message string) {
	fmt.Printf("%sDO NOT PASS%s: %s\n", colorRed, colorDefault, message)
}

//------------------------------------------------------------------------------

func usage() {
	fmt.Println("usage: " + os.Args[0] + " [ -dir <dir> | -pkg <pkg> ]")
	os.Exit(1)
}

//------------------------------------------------------------------------------
