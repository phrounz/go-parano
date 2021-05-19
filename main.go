package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"./src"
	"./src/util"
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
		ignoreGoFiles[file] = true
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

	src.DoAll(*pkgDirPtr, ignoreGoFiles, sqlqo)

	os.Exit(util.GetExitCode())
}

//------------------------------------------------------------------------------

func userFatalError(message string) {
	fmt.Println(message)
	os.Exit(1)
}

//------------------------------------------------------------------------------

func usage() {
	fmt.Println("use for help: " + os.Args[0] + " -h")
	os.Exit(1)
}

//------------------------------------------------------------------------------
