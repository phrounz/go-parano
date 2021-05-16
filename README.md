# go-parano
Experimental Go static analysis / robustness checker tool.

It scans specific keywords in comments and alerts if something is not in agreement with them.

WARNING still in development.

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

This is a way to check that your SQL queries are correct.

## How to test

You can test with:

Without the SQL linter feature:
```
go build -o go-parano.out ./src/ && ./go-parano.out -dir examples/
```

With the SQL linter feature (installs phpmyadmin/sql-parser dependency in ./vendor using composer):
```
sh build_and_test.sh
```

It should display something like:
```
DO NOT PASS: missing field(s) foo3, foo4 in declaration of testType1, declared with //!PARANO__EXHAUSTIVE_FILLING in examples/example1.go
DO NOT PASS: missing field(s) foo3, foo4 in declaration of testType1, type declared with //!PARANO__EXHAUSTIVE_FILLING in examples/example1.go
DO NOT PASS: cannot use testVarNotOkay in examples/example1.go, declared as private to file in examples/example2.go
DO NOT PASS: cannot use testType2 in examples/example1.go, declared as private to file in examples/example2.go
DO NOT PASS: cannot use testType3 in examples/example1.go, declared as private to file in examples/example2.go
DO NOT PASS: cannot use testFunctionNotOkay in examples/example1.go, declared as private to file in examples/example2.go
DO NOT PASS: cannot use testFunctionNotOkay in examples/example1.go, declared as private to file in examples/example2.go
DO NOT PASS: cannot use localPrivateStuffTest in examples/example1.go, declared as private to file in examples/example2.go
DO NOT PASS: missing fields(s) Foo2 in declaration "examplesub.TestTypeSub{}" in examples/example1.go, type declared with //!PARANO__EXHAUSTIVE_FILLING in ???
```

