package src

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/phrounz/go-parano/src/fileparser"
	"github.com/phrounz/go-parano/src/util"
)

//------------------------------------------------------------------------------

type SQLQueryOptions struct {
	FunctionsNames util.WildcardMap
	AllInOne       bool
	LintBinary     string
	IgnoreGoFiles  util.WildcardMap
}

//------------------------------------------------------------------------------

type queryInfo struct {
	strQuery string
	filename string
}

var sqlQueriesSlice []queryInfo

//------------------------------------------------------------------------------

const constIgnoreGoCheckDBQueries = "//!PARANO__IGNORE_CHECK_SQL_QUERIES"
const constIgnoreGoCheckDBQueriesAlt = "// !PARANO__IGNORE_CHECK_SQL_QUERIES"
const constIgnoreGoCheckDBQuery = "//!PARANO__IGNORE_CHECK_SQL_QUERY"
const constIgnoreGoCheckDBQueryAlt = "// !PARANO__IGNORE_CHECK_SQL_QUERY"

const constDisclaimerGoCheckDB = "## To ignore this(these) error(s) (e.g. if you think this is a false positive), put " +
	constIgnoreGoCheckDBQuery + " or " + constIgnoreGoCheckDBQueryAlt + " in the function call,\n" +
	"## or " + constIgnoreGoCheckDBQueries + " or " + constIgnoreGoCheckDBQueriesAlt + " on top of the function where this function call is done."

//------------------------------------------------------------------------------

func ParanoSqllintVisit(nCaller *fileparser.Node, filename string, constantValues []fileparser.ConstValue, sqlqo SQLQueryOptions) bool {
	if nCaller != nil && nCaller.TypeStr == "CallExpr" {

		var value, ok = sqlqo.FunctionsNames.Find(nCaller.Name)
		if !ok {
			return false
		}
		var argumentIndex, ok2 = value.(int)
		if !ok2 {
			panic("value not int in sqlQueryFunctionsNames")
		}

		// fmt.Printf("---\n")
		var countShift = 1
		// for _, subn := range nCaller.Children {
		// 	// fmt.Printf("%s\n", subn.TypeStr)
		// 	if subn.TypeStr == "SelectorExpr" { or "Ident"
		// 		countShift++
		// 	} else {
		// 		break
		// 	}
		// }

		var argIndex = argumentIndex + countShift - 1

		if argIndex >= len(nCaller.Children) {
			panic("bad index argument " + strconv.Itoa(countShift) + " for " + nCaller.Name)
		}
		var goodN = nCaller.Children[argIndex]
		if !(goodN.TypeStr == "BinaryExpr" || goodN.TypeStr == "BasicLit") {
			// panic("Arg " + strconv.Itoa(argumentIndex) + " (child node " + strconv.Itoa(argIndex) +
			// 	") is not a BinaryExpr or BasicLit but " + goodN.TypeStr + " for " + nCaller.Name + ": " + nCaller.Bytes)
			if util.IsWarn() {
				util.Warn("File '%s': Cannot check arg %d in function call: %s", filename, argumentIndex, nCaller.Bytes)
			}
		}

		if util.IsDebug() {
			util.DebugPrintf("paranoSqllintVisit: %s %s %s %s", goodN.TypeStr, goodN.Name, nCaller.TypeStr, nCaller.Name)
		}
		var strQuery, abort = goodN.ComputeStringExpression(constantValues)
		if abort {
			if util.IsWarn() {
				util.Warn("File '%s': Cannot check query in function call %s: %s", filename, nCaller.Name, goodN.Bytes)
			}
			return false
		}

		if strings.Index(nCaller.Bytes, constIgnoreGoCheckDBQuery) != -1 || strings.Index(nCaller.Bytes, constIgnoreGoCheckDBQueryAlt) != -1 {
			if util.IsDebug() || util.IsInfo() {
				util.Info("    Ignoring SQL query in '%s': %s", filename, getStrTruncated(strQuery))
			}
			return false
		}

		var nFather = nCaller.Father
		for nFather != nil {
			if nFather.TypeStr == "FuncDecl" {
				for _, subn := range nFather.Children {
					if subn.IsCommentGroupWithComment(constIgnoreGoCheckDBQueries) || subn.IsCommentGroupWithComment(constIgnoreGoCheckDBQueriesAlt) {
						if util.IsDebug() || util.IsInfo() {
							util.Info("    Ignoring SQL query in '%s' within function %s: %s", filename, nFather.Name, getStrTruncated(strQuery))
						}
						return false
					}
				}
				//break
			}
			nFather = nFather.Father
		}

		if len(strQuery) > 0 && strQuery[len(strQuery)-1] != ';' {
			strQuery += ";"
		}
		var qi = queryInfo{strQuery: strQuery, filename: filename}
		if sqlqo.AllInOne {
			sqlQueriesSlice = append(sqlQueriesSlice, qi)
		} else {
			return checkQuery(qi, false, sqlqo.LintBinary)
		}

	}
	return false
}

//------------------------------------------------------------------------------

func ParanoSqllintCheckQueries(sqlqo SQLQueryOptions) {
	if len(sqlQueriesSlice) > 0 {
		var sqlQueriesAll string
		for _, qi := range sqlQueriesSlice {
			sqlQueriesAll += qi.strQuery + "\n"
		}
		if util.IsInfo() {
			util.Info("Checking %d SQL queries (%d characters)...", len(sqlQueriesSlice), len(sqlQueriesAll))
			//fmt.Printf("%s\n", sqlQueriesAll)
		}
		checkQuery(queryInfo{strQuery: sqlQueriesAll, filename: "???"}, true, sqlqo.LintBinary)
		if util.IsInfo() {
			util.Info("Checking %d SQL queries done.", len(sqlQueriesSlice))
		}
	}
}

//------------------------------------------------------------------------------

func checkQuery(qi queryInfo, isGroupOfQueries bool, sqlQueryLintBinary string) (failed bool) {
	if util.IsDebug() {
		util.DebugPrintf("checkQuery: %s", qi.strQuery)
	}

	var sqlQueryLintBinaryWithArgs = strings.Split(sqlQueryLintBinary, " ")
	var out, exitCode = util.RunCmdWithStdin(qi.strQuery, sqlQueryLintBinaryWithArgs[0], sqlQueryLintBinaryWithArgs[1:])
	//fmt.Printf("out: %s\n", out)
	if out != "" && exitCode != 0 {
		if isGroupOfQueries {
			util.NotPass("Invalid SQL query in %s:\n%s\n%s", qi.filename, out, constDisclaimerGoCheckDB)
		} else {
			util.NotPass("Invalid SQL query in %s: %s\n%s\n%s", qi.filename, getStrTruncated(qi.strQuery), out, constDisclaimerGoCheckDB)
		}
		failed = true
		return
	} else if out != "" {
		fmt.Printf("%s\n", out)
	}
	return
}

//------------------------------------------------------------------------------

func getStrTruncated(str string) string {
	if len(str) > 50 {
		return str[:48] + "..."
	}
	return str
}

//------------------------------------------------------------------------------
