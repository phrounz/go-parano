/*
	https://medium.com/justforfunc/understanding-go-programs-with-go-parser-c4e88a6edb87
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//------------------------------------------------------------------------------

type infosFile struct {
	rootNode                 *node
	privateToFileDecl        map[string]bool
	exhaustiveFillingStructs map[string]map[string]bool
}

const constPrivateToFileComment = "//!PARANO__PRIVATE_TO_FILE"
const constExaustiveFilling = "//!PARANO__EXHAUSTIVE_FILLING"
const constLocalPrivateStuffLineRegexp1 = "\n//\\s+LOCAL PRIVATE STUFF\\s+\n"
const constLocalPrivateStuffLineRegexp2 = "\n//\\s+PRIVATE LOCAL STUFF\\s+\n"

var colorRed = "\033[31m"
var colorDefault = "\033[39m"

var verbose bool
var noColor bool

//------------------------------------------------------------------------------

func main() {

	if len(os.Args) == 1 {
		usage()
	}

	var noColorPtr = flag.Bool("nocolor", false, "disable color")
	var pkgDir = flag.String("dir", "", "source directory")
	var pkg = flag.String("pkg", "", "source package")
	var verbosePtr = flag.Bool("v", false, "verbose info")
	var debug = flag.Bool("debug", false, "debug info")
	flag.Parse()
	noColor = *noColorPtr
	verbose = *verbosePtr

	debugInfo = *debug

	if *pkg != "" {
		*pkgDir = os.Getenv("GOPATH") + "/src/" + *pkg
	}

	_, err := os.Stat(*pkgDir)
	if os.IsNotExist(err) {
		panic("Folder " + *pkgDir + " does not exist.")
	}

	recurseDir(*pkgDir)
}

//------------------------------------------------------------------------------

func recurseDir(pkgDir string) {
	var items, err = filepath.Glob(pkgDir + "/*")
	if err != nil {
		panic(err)
	}
	var srcFiles []string
	for _, item := range items {
		if len(item) >= 3 && item[len(item)-3:] == ".go" {
			srcFiles = append(srcFiles, item)
		} else {
			info, err := os.Stat(item)
			if os.IsNotExist(err) {
				panic("File does not exist: " + item)
			}
			if info.IsDir() {
				recurseDir(item)
			}
		}
	}
	// if len(srcFiles) == 0 {
	// 	fmt.Println("WARNING: no sources files in " + pkgDir)
	// }
	if verbose {
		fmt.Println("Processing package: " + pkgDir)
	}

	processPkgFiles(srcFiles)
}

//------------------------------------------------------------------------------

func processPkgFiles(files []string) {

	var infosByFile = make(map[string]infosFile)
	for _, filename := range files {
		infosByFile[filename] = processFile(filename)
	}

	if debugInfo || verbose {
		fmt.Println("  Checking ...")
	}

	if noColor {
		colorRed = ""
		colorDefault = ""
	}

	//----
	// third pass => check

	var failedAtLeastOnce = false
	for filename1, infosFile := range infosByFile {
		infosFile.rootNode.visit(func(n *node) {
			if n.name != "" {
				for filename2, infosFile2 := range infosByFile {

					//----
					// privateToFileDecl

					if filename1 != filename2 {
						if _, ok := infosFile2.privateToFileDecl[n.name]; ok {
							notPass(fmt.Sprintf("cannot use %s in %s, declared as private to file in %s",
								n.name, filename1, filename2))
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
									strings.Join(missingKeys, ", "), n.name, constExaustiveFilling, filename2))
								failedAtLeastOnce = true
							}
						}
					}
				}
			}
		})
	}
	if failedAtLeastOnce {
		fmt.Println("Stopping program now.")
		os.Exit(1)
	}
}

//------------------------------------------------------------------------------

func processFile(filename string) infosFile {
	if debugInfo {
		fmt.Println("===============================")
	}
	if verbose {
		fmt.Println("  Scanning: " + filepath.Base(filename) + " ...")
	}

	//----
	// first pass => load file and get nodes

	var locationLocalPrivateStuff = -1

	var rootNode = readFile(filename)

	var loc = regexp.MustCompile(constLocalPrivateStuffLineRegexp1).FindIndex(fileBytes)
	if len(loc) > 0 {
		locationLocalPrivateStuff = loc[1]
	} else {
		loc = regexp.MustCompile(constLocalPrivateStuffLineRegexp2).FindIndex(fileBytes)
		if len(loc) > 0 {
			locationLocalPrivateStuff = loc[1]
		}
	}
	if debugInfo {
		fmt.Printf("  locationLocalPrivateStuff: %d\n", locationLocalPrivateStuff)
	}

	//----
	// second pass => get privateToFileDecl/exhaustiveFillingStructs

	var privateToFileDecl = make(map[string]bool)
	var exhaustiveFillingStructs = make(map[string]map[string]bool)

	rootNode.visit(func(n *node) {
		if n.typeStr == "Ident" && n.depthLevel <= 4 && locationLocalPrivateStuff != -1 && n.bytesIndexBegin > locationLocalPrivateStuff {
			privateToFileDecl[n.name] = true
		}
		if isCommentGroupWithComment(n, constPrivateToFileComment) && n.father != nil {
			if n.father.typeStr == "GenDecl" {
				for _, n2 := range n.father.children {
					if n2.typeStr == "ValueSpec" {
						if len(n2.children) >= 2 {
							var name = n2.children[0].bytes
							if debugInfo {
								fmt.Printf("CCCC >=%s <=\n", name)
							}
							privateToFileDecl[name] = true
							break
						}
					}
				}
			} else if n.father.typeStr == "FuncDecl" {
				if debugInfo {
					fmt.Printf("AAAA >=%s %s<=\n", n.father.name, n.father.typeStr)
				}
				privateToFileDecl[n.father.name] = true
			} else {
				var nextNode = n.nextNode()
				if nextNode != nil && nextNode.typeStr == "TypeSpec" {
					if debugInfo {
						fmt.Printf("BBBB >=%s %s<=\n", nextNode.name, nextNode.typeStr)
					}
					privateToFileDecl[nextNode.name] = true
				}
			}
		}
		if isCommentGroupWithComment(n, constExaustiveFilling) && n.father != nil {
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
		fmt.Printf("\n\n\nprivateToFileDecl: %+v\n", privateToFileDecl)
		fmt.Printf("\n\n\nexhaustiveFillingStructs: %+v\n", exhaustiveFillingStructs)
	}

	return infosFile{
		rootNode:                 rootNode,
		privateToFileDecl:        privateToFileDecl,
		exhaustiveFillingStructs: exhaustiveFillingStructs,
	}
}

func isCommentGroupWithComment(n *node, comment string) bool {
	if n.typeStr == "CommentGroup" {
		for _, child := range n.children {
			if child.bytes == comment {
				return true
			}
		}
	}
	return false
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
