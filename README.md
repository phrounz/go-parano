# go-parano
Experimental Go static analysis / robustness checker tool.

go-parano is organized into different *features* (see below), 
each feature checks a different thing. go-parano scans the source code 
(argument -dir or -pkg), including specific keywords in comments, 
and alerts (and returns exit code 2) if something is wrong.

WARNING: still in development. Tested only on Linux, but there is no reason it would not work on other operating systems.

## How to build and test

If you don't need the SQL linter feature:
```
$ go build -o go-parano.out ./src/ && ./go-parano.out -dir examples/
```

If you need the SQL linter feature (version installing and using phpmyadmin/sql-parser using composer):
```
$ go build -o go-parano.out ./src/ && sh test_phpmyadmin_sql-parser.sh
```
If you need the SQL linter feature (version installing and using sqlfluff using pip):
```
$ go build -o go-parano.out ./src/ && sh test_sqlfluff.sh
```

It should display something like (example below is the version with phpmyadmin/sql-parser):
```
INVALID: missing fields(s) foo3, foo4 in declaration "testType1{}" in examples/example1.go, type declared with //!PARANO__EXHAUSTIVE_FILLING in examples/example1.go
INVALID: Cannot use testVarNotOkay in examples/example1.go, declared as private to file in examples/example2.go
INVALID: Cannot use testType2 in examples/example1.go, declared as private to file in examples/example2.go
INVALID: Cannot use testType3 in examples/example1.go, declared as private to file in examples/example2.go
INVALID: Cannot use testFunctionNotOkay in examples/example1.go, declared as private to file in examples/example2.go
INVALID: Cannot use testFunctionNotOkay in examples/example1.go, declared as private to file in examples/example2.go
INVALID: Cannot use localPrivateStuffTest in examples/example1.go, declared as private to file in examples/example2.go
INVALID: Invalid SQL query in examples/example1.go: SELECT FROM JOIN "1";
INVALID:      |_ #1: An expression was expected. (near "FROM" at position 7)
INVALID:      |_ #2: An expression was expected. (near "JOIN" at position 12)
INVALID:      |_ 
INVALID: Invalid SQL query in examples/example1.go: INSERT INTO elements typo mistake (`foo`,`bar`) ...
INVALID:      |_ #1: Unexpected token. (near "typo" at position 21)
INVALID:      |_ #2: Unexpected beginning of statement. (near "typo" at position 21)
INVALID:      |_ #3: Unexpected beginning of statement. (near "mistake" at position 26)
INVALID:      |_ #4: Unexpected beginning of statement. (near "`foo`" at position 35)
INVALID:      |_ #5: Unexpected beginning of statement. (near "`bar`" at position 41)
INVALID:      |_ #6: Unrecognized statement type. (near "VALUES" at position 48)
INVALID:      |_ 
WARNING: Cannot fully check query in file 'examples/example1.go': SELECT * FROM ???
INVALID: missing fields(s) Foo2 in declaration "examplesub.TestTypeSub{}" in examples/example1.go, type declared with //!PARANO__EXHAUSTIVE_FILLING in ???
```
## Features:

### Feature: private to file

This gives a way to ensure that a function/type/variable is not 
used in another file of the same package.

For example:
```
//!PARANO__PRIVATE_TO_FILE
var i int // ---> i cannot be used in other files of the same package.
```

Also everything below the line `// LOCAL PRIVATE STUFF` 
until the end of file, is also _private to file_.

### Feature: struct exhaustive filling

This gives a way to check that all the fields of a Go struct are informed 
when that struct is instancied.
```
//!PARANO__EXHAUSTIVE_FILLING
type testType1 struct {
	foo1 int
	foo2 int
}
...
var testNokFill = testType1{
  foo1: 3, // ---> foo2 is detected as unaffected
}
```

### Feature: SQL linter

This is a way to check that the SQL queries in the Go code are correct.
 
Basically you define in the arguments of go-parano:
 * the function(s) used for all of your queries, and the argument index in 
 this(these) function(s) containing the query 
 (like `examplesub.Query:1`) (with `-sql-query-func-name`)
 * a linter program (with `-sql-query-lint-binary`)
	
and it will check all the queries in the functions calls in the source code.

```
examplesub.Query("SELECT 1") // -> will run the linter program on "SELECT 1"
```

You can use any linter program as long as:
 * it accepts the query as stdin.
 * it returns nonzero code and a message if there is an error in the query.

Current features:
 * Supports if the query is splitted into several strings concatenated 
 with '+', or even if it contains a constant declared in the current source file.

Current limitations (TODO): 
 * What if the function uses a struct as argument and the query is a field in the struct.
 * What if this is a method and not a function e.g. `myDBHandle.Query()`.
 * Limitations depending of the linter programs I tested:
 	* '?' (static queries) won't work?
 	* Does not check whether the tables and fields exist.
 
If you don't want to check queries in a function (e.g. false positives), 
put as a comment `//!PARANO__IGNORE_CHECK_SQL_QUERIES` on top of the 
function declaration with the function call, for example:
```
//!PARANO__IGNORE_CHECK_SQL_QUERIES
func foo() {
	examplesub.Query("SELECT 1")
}
```

Or use `//!PARANO__IGNORE_CHECK_SQL_QUERY` within the function call, 
for example:
```
examplesub.Query( //!PARANO__IGNORE_CHECK_SQL_QUERY
	"SELECTYYYY")
}
```

