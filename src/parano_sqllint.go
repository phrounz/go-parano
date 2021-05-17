package main

import (
	"fmt"
	"strings"

	"./node"
	"./util"
)

//------------------------------------------------------------------------------

// used in sqlQueryAllInOne==true
type queryInfo struct {
	strQuery string
	filename string
}

var sqlQueriesSlice []queryInfo

//------------------------------------------------------------------------------

const constIgnoreGoCheckDBQueries = "//!PARANO__IGNORE_CHECK_SQL_QUERIES"
const constIgnoreGoCheckDBQuery = "//!PARANO__IGNORE_CHECK_SQL_QUERY"

//------------------------------------------------------------------------------

func paranoSqllintVisit(n *node.Node, functionsNames []string, filename string) bool {
	if n.Father != nil && n.Father.TypeStr == "CallExpr" && util.InSlice(n.Father.Name, functionsNames) &&
		(n.TypeStr == "BinaryExpr" || n.TypeStr == "BasicLit") {

		if debugInfo {
			util.DebugPrintf("paranoSqllintVisit: %s %s %s %s", n.TypeStr, n.Name, n.Father.TypeStr, n.Father.Name)
		}
		var strQuery, abort = concatStringsRecursive(n)

		if strings.Index(n.Father.Bytes, constIgnoreGoCheckDBQuery) != -1 {
			if debugInfo || verbose {
				util.Info("    Ignoring SQL query in '%s': %s", filename, getSQLQueryTruncated(strQuery))
			}
			return false
		}

		var nFather = n.Father.Father
		for nFather != nil {
			if nFather.TypeStr == "FuncDecl" {
				for _, subn := range nFather.Children {
					if subn.IsCommentGroupWithComment(constIgnoreGoCheckDBQueries) {
						if debugInfo || verbose {
							util.Info("    Ignoring SQL query in '%s' within function %s: %s", filename, nFather.Name, getSQLQueryTruncated(strQuery))
						}
						return false
					}
				}
				//break
			}
			nFather = nFather.Father
		}

		var firstBasicLit = true
		for _, subn := range n.Father.Children {
			if subn.TypeStr == "BasicLit" || subn.TypeStr == "BinaryExpr" {
				if firstBasicLit && n != subn {
					//fmt.Printf(" %+v %+v\n", n, subn)
					return false // we only check if it is the first BasicLit-type argument of the function
				}
				firstBasicLit = false
			}
		}

		if abort {
			util.Warn("Cannot fully check query in file '%s': %s", filename, strQuery)
			return false
		}
		if len(strQuery) > 0 && strQuery[len(strQuery)-1] != ';' {
			strQuery += ";"
		}
		var qi = queryInfo{strQuery: strQuery, filename: filename}
		if sqlQueryAllInOne {
			sqlQueriesSlice = append(sqlQueriesSlice, qi)
		} else {
			return checkQuery(qi)
		}

	}
	return false
}

//------------------------------------------------------------------------------

func paranoSqllintCheckQueries() {
	if len(sqlQueriesSlice) > 0 {
		var sqlQueriesAll string
		for _, qi := range sqlQueriesSlice {
			sqlQueriesAll += qi.strQuery + "\n"
		}
		if verbose {
			util.Info("Checking %d SQL queries (%d characters)...", len(sqlQueriesSlice), len(sqlQueriesAll))
			//fmt.Printf("%s\n", sqlQueriesAll)
		}
		checkQuery(queryInfo{strQuery: sqlQueriesAll, filename: "???"})
		if verbose {
			util.Info("Checking %d SQL queries done.", len(sqlQueriesSlice))
		}
	}
}

//------------------------------------------------------------------------------

func concatStringsRecursive(n *node.Node) (strQuery string, abort bool) {
	if n.TypeStr == "BasicLit" {
		strQuery = n.Name
	} else if n.TypeStr == "BinaryExpr" {
		for _, subn := range n.Children {
			var subStrQuery, abortSub = concatStringsRecursive(subn)
			strQuery += subStrQuery
			if abortSub {
				abort = true
			}
		}
	} else {
		strQuery += "???"
		abort = true
	}
	return
}

//------------------------------------------------------------------------------

func checkQuery(qi queryInfo) (failed bool) {
	if debugInfo {
		util.DebugPrintf("checkQuery: %s", qi.strQuery)
	}

	var sqlQueryLintBinaryWithArgs = strings.Split(sqlQueryLintBinary, " ")
	var out, exitCode = util.RunCmdWithStdin(qi.strQuery, sqlQueryLintBinaryWithArgs[0], sqlQueryLintBinaryWithArgs[1:])
	//fmt.Printf("out: %s\n", out)
	if out != "" && exitCode != 0 {
		util.NotPass("Invalid SQL query in %s: %s\n%s", qi.filename, getSQLQueryTruncated(qi.strQuery), out)
		failed = true
		return
	} else if out != "" {
		fmt.Printf("%s\n", out)
	}
	return
}

//------------------------------------------------------------------------------

func getSQLQueryTruncated(sqlQueryStr string) string {
	if len(sqlQueryStr) > 50 {
		return sqlQueryStr[:48] + "..."
	}
	return sqlQueryStr
}

//------------------------------------------------------------------------------
