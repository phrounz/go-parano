package main

import (
	"fmt"

	"./node"
)

//------------------------------------------------------------------------------

func paranoSqllintVisit(n *node.Node, functionsNames []string, filename string) bool {
	if n.Father != nil && n.Father.TypeStr == "CallExpr" && inSlice(n.Father.Name, functionsNames) {

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

		if n.TypeStr == "BinaryExpr" || n.TypeStr == "BasicLit" {
			if debugInfo {
				fmt.Printf("paranoSqllintVisit: %s %s %s %s\n", n.TypeStr, n.Name, n.Father.TypeStr, n.Father.Name)
			}
			var strQuery, abort = concatStringsRecursive(n)
			if abort {
				warn("Cannot fully check query: " + strQuery)
				return false
			}
			return checkQuery(strQuery, filename)

		}

	}
	return false
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

func checkQuery(sqlQueryStr string, filename string) (failed bool) {
	if debugInfo {
		fmt.Printf("checkQuery: %s\n", sqlQueryStr)
	}
	if len(sqlQueryStr) > 0 && sqlQueryStr[len(sqlQueryStr)-1] != ';' {
		sqlQueryStr += ";"
	}
	var out, exitCode = runCmdWithStdin(sqlQueryStr, sqlQueryLintBinary)
	//fmt.Printf("out: %s\n", out)
	if out != "" && exitCode != 0 {
		var sqlQueryStrBeginning = sqlQueryStr
		if len(sqlQueryStrBeginning) > 20 {
			sqlQueryStrBeginning = sqlQueryStrBeginning[:18] + "..."
		}
		notPass(fmt.Sprintf("invalid SQL query in %s: %s\n%s", filename, sqlQueryStrBeginning, out))
		failed = true
		return
	}
	return
}

//------------------------------------------------------------------------------
