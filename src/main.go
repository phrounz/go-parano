package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"./fileparser"
	"./util"
)

//------------------------------------------------------------------------------

type packageInfos struct {
	packageName      string
	packageDir       string
	subPackagesInfos []*packageInfos
	infosByFile      map[string]infosFile
}

type infosFile struct {
	packageName              string
	rootNode                 *fileparser.Node
	fileConstants            []fileparser.ConstValue
	featurePrivateToFile     *featurePrivateToFile
	featureExhaustiveFilling *featureExhaustiveFilling
}

var verbose bool
var debugInfo bool
var noWarn bool
var noColor bool
var sqlQueryFunctionsNames util.WildcardMap
var sqlQueryLintBinary string
var sqlQueryAllInOne bool
var ignoreGoFiles = make(map[string]bool)

//------------------------------------------------------------------------------

func main() {

	if len(os.Args) == 1 {
		usage()
	}

	var noColorPtr = flag.Bool("nocolor", false, "Disables color in output.")
	var pkgDir = flag.String("dir", "", "Source directory.")
	var pkg = flag.String("pkg", "", "Source package.")
	var verbosePtr = flag.Bool("v", false, "Shows INFO messages.")
	var debug = flag.Bool("debug", false, "Shows DEBUG messages.")
	var noWarnPtr = flag.Bool("no-warn", false, "Hide WARNING messages (those messages usually show something the program cannot check because of its limitations).")
	var sqlQueryFunctionNamePtr = flag.String("sql-query-func-name", "", "Name of the function used for queries in the source code.\n"+
		"- You may provide several function names, separated by comma.\n"+
		"- Accepts up to one wilcard character '*' by function name.\n"+
		"- Each function name must contain as suffix a colon followed by the argument index \n"+
		"  (starting from 1) containing the query, e.g. \":2\" is the second function argument.")
	var sqlQueryLintBinaryPtr = flag.String("sql-query-lint-binary", "", "SQL query lint program")
	var sqlQueryAllInOnePtr = flag.Bool("sql-query-all-in-one", false, "If set, run the SQL query lint program once with all the queries as argument, instead of running once by query.")
	var ignoreGoFilesPtr = flag.String("ignore-go-files", "", "List of files to ignore when using -dir or -pkg, comma-separated.")
	flag.Parse()
	noColor = *noColorPtr
	verbose = *verbosePtr
	if *sqlQueryFunctionNamePtr != "" {
		sqlQueryFunctionsNames = util.NewWildcardMap()
		for _, el := range strings.Split(*sqlQueryFunctionNamePtr, ",") {
			var elSplitted = strings.Split(el, ":")
			if len(elSplitted) != 2 {
				panic("bad argument: " + *sqlQueryFunctionNamePtr)
			}
			var i, err = strconv.Atoi(elSplitted[1])
			if err != nil {
				panic("bad argument: " + *sqlQueryFunctionNamePtr + ": " + err.Error())
			}
			sqlQueryFunctionsNames.Add(elSplitted[0], i)
		}
	}
	sqlQueryLintBinary = *sqlQueryLintBinaryPtr
	sqlQueryAllInOne = *sqlQueryAllInOnePtr
	for _, file := range strings.Split(*ignoreGoFilesPtr, ",") {
		ignoreGoFiles[file] = true
	}

	debugInfo = *debug
	fileparser.DebugInfo = debugInfo
	noWarn = *noWarnPtr

	if *pkg != "" {
		*pkgDir = os.Getenv("GOPATH") + "/src/" + *pkg
	}

	_, err := os.Stat(*pkgDir)
	if os.IsNotExist(err) {
		panic("Folder " + *pkgDir + " does not exist.")
	}

	var rootPkg = recurseDir(*pkgDir)

	paranoSqllintCheckQueries()

	var mInfosByPackageName = make(map[string]*packageInfos)
	processPkgRecursiveAndMakeMap(rootPkg, mInfosByPackageName)

	if verbose {
		util.Info("\"Fourth\" pass")
	}

	processPkgAgain(mInfosByPackageName)

	os.Exit(util.GetExitCode())
}

//------------------------------------------------------------------------------

func processPkgRecursiveAndMakeMap(pkgInfos *packageInfos, mInfosByPackageName map[string]*packageInfos) {
	mInfosByPackageName[pkgInfos.packageName] = pkgInfos
	for _, subPackageInfos := range pkgInfos.subPackagesInfos {
		processPkgRecursiveAndMakeMap(subPackageInfos, mInfosByPackageName)
	}
}

//------------------------------------------------------------------------------

func recurseDir(pkgDir string) *packageInfos {

	var subPackagesInfos = make([]*packageInfos, 0)

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
				subPackagesInfos = append(subPackagesInfos, recurseDir(item))
			}
		}
	}
	// if len(srcFiles) == 0 {
	// 	fmt.Println("WARNING: no sources files in " + pkgDir)
	// }
	if verbose {
		util.Info("Processing package: %s", pkgDir)
	}

	var infosByFile = processPkgFiles(srcFiles)
	var packageName string // note: remains empty string if no source files
	for _, infosFile := range infosByFile {
		packageName = infosFile.packageName
		break
	}
	return &packageInfos{
		packageName:      packageName,
		packageDir:       pkgDir,
		subPackagesInfos: subPackagesInfos,
		infosByFile:      infosByFile,
	}
}

//------------------------------------------------------------------------------

func processPkgFiles(files []string) (infosByFile map[string]infosFile) {

	infosByFile = make(map[string]infosFile)
	for _, filename := range files {
		if _, ok := ignoreGoFiles[filename]; !ok {
			infosByFile[filename] = processFile(filename)
		}
	}

	if debugInfo || verbose {
		util.Info("  Checking ...")
	}

	if noColor {
		util.DisableColor()
	}

	//----
	// third pass => check

	var failedAtLeastOnce = false
	for filename1, fileInfos := range infosByFile { // for each input file
		fileInfos.rootNode.Visit(func(n *fileparser.Node) {
			if len(sqlQueryFunctionsNames) > 0 {
				failedAtLeastOnce = paranoSqllintVisit(n, filename1, fileInfos.fileConstants) && failedAtLeastOnce
			}
			if n.Name != "" {
				for filename2, fileInfos2 := range infosByFile { // for each file
					failedAtLeastOnce = paranoPrivateToFileCheck(n, fileInfos2.featurePrivateToFile, filename1, filename2) && failedAtLeastOnce
					failedAtLeastOnce = paranoExhaustiveFillingCheck(n, fileInfos2.packageName, fileInfos2.featureExhaustiveFilling, filename1, filename2) && failedAtLeastOnce
				}
			}
		})
	}
	if failedAtLeastOnce {
		fmt.Println("Stopping program now.")
		os.Exit(1)
	}

	return
}

//------------------------------------------------------------------------------

func processFile(filename string) infosFile {
	if debugInfo {
		util.DebugPrintf("===============================> %s", filepath.Base(filename))
	}
	if verbose {
		util.Info("  Scanning: " + filepath.Base(filename) + " ...")
	}

	//----
	// first pass => load file and get informations

	var fileInfo = fileparser.ReadFile(filename)

	//----
	// init

	var featurePrivateToFile = paranoPrivateToFileInit(fileInfo.FileBuffer)
	var featureExhaustiveFilling = paranoExhaustiveFillingInit()

	//----
	// second pass => gather informations about nodes of this file

	fileInfo.RootNode.Visit(func(n *fileparser.Node) {
		if debugInfo {
			n.Display()
		}
		paranoPrivateToFileVisit(n, featurePrivateToFile)
		paranoExhaustiveFillingVisit(n, featureExhaustiveFilling)
	})

	var infosf = infosFile{
		packageName:              fileInfo.PackageName,
		rootNode:                 fileInfo.RootNode,
		fileConstants:            fileInfo.FileConstants,
		featurePrivateToFile:     featurePrivateToFile,
		featureExhaustiveFilling: featureExhaustiveFilling,
	}

	if debugInfo {
		util.DebugPrintf("\n\n\n%+v\n", infosf)
	}

	return infosf
}

//------------------------------------------------------------------------------

func usage() {
	fmt.Println("use for help: " + os.Args[0] + " -h")
	os.Exit(1)
}

//------------------------------------------------------------------------------

func processPkgAgain(mInfosByPackageName map[string]*packageInfos) {

	//----
	// fourth pass => check

	paranoExhaustiveFillingCheckGlobal(mInfosByPackageName)
}

//------------------------------------------------------------------------------
