package main

import (
	"fmt"

	"./node"
)

func paranoSqllintVisit(n *node.Node, functionName string, filename string) bool {
	if n.TypeStr == "BasicLit" && n.Father.TypeStr == "CallExpr" && n.Father.Name == functionName {
		var sqlQueryStr = n.Name
		if sqlQueryStr[len(sqlQueryStr)-1] != ';' {
			sqlQueryStr += ";"
		}
		if debugInfo {
			fmt.Printf("paranoSqllintVisit: %s %s %s\n", sqlQueryStr, n.Father.TypeStr, n.Father.Name)
		}
		var out, exitCode = runCmdWithStdin(sqlQueryStr, sqlQueryLintBinary)
		//fmt.Printf("out: %s\n", out)
		if out != "" && exitCode != 0 {
			var sqlQueryStrBeginning = sqlQueryStr
			if len(sqlQueryStrBeginning) > 20 {
				sqlQueryStrBeginning = sqlQueryStrBeginning[:18] + "..."
			}
			notPass(fmt.Sprintf("invalid sql query in %s: %s\n%s", filename, sqlQueryStrBeginning, out))
			return false
		}
	}
	return true
}
