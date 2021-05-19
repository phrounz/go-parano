package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/phrounz/go-parano/src"
	"github.com/phrounz/go-parano/src/util"
)

//------------------------------------------------------------------------------

func main() {

	//---
	// parse arguments

	if len(os.Args) == 1 {
		usage()
	}

	var noColorPtr = flag.Bool("nocolor", false, "Disables color in output.")
	var pkgDirPtr = flag.String("dir", "", "Source directory.")
	var pkgPtr = flag.String("pkg", "", "Source package.")
	var verbosePtr = flag.Bool("v", false, "Shows INFO messages.")
	var debugPtr = flag.Bool("debug", false, "Shows DEBUG messages.")
	var noWarnPtr = flag.Bool("no-warn", false, "Hide WARNING messages (those messages usually show something the program cannot check because of its limitations).")
	var sqlQueryFunctionNamePtr = flag.String("sql-query-func-name", "", "Name of the function used for queries in the source code.\n"+
		"- You may provide several function names, separated by comma.\n"+
		"- Accepts up to one wilcard character '*' by function name.\n"+
		"- Each function name must contain as suffix a colon followed by the argument index \n"+
		"  (starting from 1) containing the query, e.g. \":2\" is the second function argument.")
	var sqlQueryLintBinaryPtr = flag.String("sql-query-lint-binary", "", "SQL query lint program")
	var sqlQueryAllInOnePtr = flag.Bool("sql-query-all-in-one", false, "If set, run the SQL query lint program once with all the queries as argument, instead of running once by query.")
	var ignoreGoFilesPtr = flag.String("ignore-go-files", "", "List of files to ignore when using -dir or -pkg, comma-separated.")
	var ignorePrivateToFilePtr = flag.String("ignore-private-to-file", "", "List of functions/variables which shall be ignored when checking private-to-file, comma-separated.")
	flag.Parse()

	//---
	// -nocolor / -v / -debug / -no-warn

	if *noColorPtr {
		util.DisableColor()
	}
	util.SetVerbosity(*verbosePtr, *debugPtr, !(*noWarnPtr))

	//---
	// -sql-query-XXX

	var sqlQueryFunctionsNames util.WildcardMap
	if *sqlQueryFunctionNamePtr != "" {
		sqlQueryFunctionsNames = util.NewWildcardMap()
		for _, el := range strings.Split(*sqlQueryFunctionNamePtr, ",") {
			var elSplitted = strings.Split(el, ":")
			if len(elSplitted) != 2 {
				userFatalError("Invalid argument: " + *sqlQueryFunctionNamePtr)
			}
			var i, err = strconv.Atoi(elSplitted[1])
			if err != nil {
				userFatalError("Invalid argument: " + *sqlQueryFunctionNamePtr + ": " + err.Error())
			}
			sqlQueryFunctionsNames.Add(elSplitted[0], i)
		}
	}
	var sqlqo = src.SQLQueryOptions{
		FunctionsNames: sqlQueryFunctionsNames,
		AllInOne:       *sqlQueryAllInOnePtr,
		LintBinary:     *sqlQueryLintBinaryPtr,
	}

	//---
	// -ignore-go-files

	var ignoreGoFiles = make(map[string]bool)
	for _, file := range strings.Split(*ignoreGoFilesPtr, ",") {
		if len(file) > 2 && file[:2] == "./" {
			file = file[2:] // remove ./ because it causes a problem when matching files of -dir or -pkg
		}
		ignoreGoFiles[file] = true
	}

	//---
	// -ignore-private-to-file

	var ignorePrivateToFile = make(map[string]bool)
	for _, name := range strings.Split(*ignorePrivateToFilePtr, ",") {
		ignorePrivateToFile[name] = true
	}

	//---
	// -dir / -pkg

	if *pkgPtr != "" {
		*pkgDirPtr = os.Getenv("GOPATH") + "/src/" + *pkgPtr
	}

	_, err := os.Stat(*pkgDirPtr)
	if os.IsNotExist(err) {
		userFatalError("Folder " + *pkgDirPtr + " does not exist.")
	}

	//---
	// do stuff

	src.DoAll(*pkgDirPtr, src.Options{
		IgnoreGoFiles:       ignoreGoFiles,
		IgnorePrivateToFile: ignorePrivateToFile,
		Sqlqo:               sqlqo,
	})

	os.Exit(util.GetExitCode())
}

//------------------------------------------------------------------------------

func userFatalError(message string) {
	fmt.Println(message)
	os.Exit(1)
}

//------------------------------------------------------------------------------

func usage() {
	fmt.Printf(
		"Use for help: " + os.Args[0] + " -h\n\n" +
			"More informations at http://github.com/phrounz/go-parano\n")
	os.Exit(1)
}

//------------------------------------------------------------------------------
