# go-parano
Experimental Go static analysis / robustness checker tool.

It scans specific keywords in comments and alerts if something is not in agreement with them.

WARNING: still in development. Tested only on Linux, but there is no reason it would not work on other operating systems.

## How to build and test

Without the SQL linter feature:
```
$ go build -o go-parano.out ./src/ && ./go-parano.out -dir examples/
```

With the SQL linter feature (installs and uses phpmyadmin/sql-parser using composer):
```
$ go build -o go-parano.out ./src/ && sh test_phpmyadmin_sql-parser.sh
```
With the SQL linter feature (installs and uses sqlfluff using pip):
```
$ go build -o go-parano.out ./src/ && sh test_sqlfluff.sh
```

It should display something like (example with phpmyadmin/sql-parser):
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

## Feature: private to file

The first feature is a _private to file_ marker: a function/type/variable declared as "private to file" cannot be used in another file of the same package.

For example:
```
//!PARANO__PRIVATE_TO_FILE
var i int
```
disallows i to be used in other files of the same package.

Also everything below the line:
```
// LOCAL PRIVATE STUFF
```
until the end of file, is also _private to file_.

## Feature: struct exhaustive filling

The second feature is a way to check that a structure is completely filled.
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

## Feature: SQL linter

This is a way to check that the SQL queries in the Go code are correct.
 
Basically you define in the arguments of go-parano:
 * the function that is used for all of your queries 
 (like `examplesub.Query`) (with `-sql-query-func-name`)
 * a linter program (with `-sql-query-lint-binary`)
	
and it will check all the queries in the functions calls in the source code.

```
examplesub.Query("SELECT 1") // -> will run the linter program on "SELECT 1"
```

You can use any linter program as long as:
 * it accepts the query as stdin.
 * it returns nonzero code and a message if there is an error in the query.

Current features:
 * Supports if there are several arguments in the function. *Note:* the query 
 shall be the first string-type argument of the function.
 * Supports if the query is splitted into several strings concatenated 
 with '+'. 

Current limitations (TODO): 
 * Does not work if the query is a variable or even a constant 
 e.g. `examplesub.Query(foo)`
 * What if this is a method and not a function e.g. `dbHandle.Query()`.
 * '?' (static queries) won't work? (depends of the linter, though)
 * Do not check whether the tables and fields exist (should be configurable 
	 with an SQL file).
 
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

