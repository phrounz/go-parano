# go-paranoid-broccoli
Experimental Go static analysis and robustness checker - WARNING still in development and messy.

**Feature: private to file**

The first feature is a _private to file_ marker: a function/type/variable declared as "private to file" cannot be used in another file of the same package.

For example:
```
//!PB_PRIVATE_TO_FILE
var i int
```
disallows i to be used in other files of the same package.

Also everything below the line:
```
// LOCAL PRIVATE STUFF
```
until the end of file, is also _private to file_.

**Feature: struct exhaustive filling**

The second feature is a way to check that a structure is completely filled.
```
//!PB_EXHAUSTIVE_FILLING
type testType1 struct {
	foo1 int
	foo2 int
}
...
var testNokFill = testType1{
  foo1: 3, // ---> foo2 is detected as missing
}
```

**How to test**

You can test with the source code itself:

```
cd go-paranoid-broccoli
go build -o go-paranoid-broccoli.out ./src/*
./go-paranoid-broccoli.out -dir examples/
```

It should display something like:
```
DO NOT PASS: missing key(s) foo3, foo4 in declaration of testType1, declared with //!PB_EXHAUSTIVE_FILLING in examples/example1.go
DO NOT PASS: missing key(s) foo3, foo4 in declaration of testType1, type declared with //!PB_EXHAUSTIVE_FILLING in examples/example1.go
DO NOT PASS: cannot use testVarNotOkay in examples/example1.go, declared as private to file in examples/example2.go
DO NOT PASS: cannot use testType2 in examples/example1.go, declared as private to file in examples/example2.go
DO NOT PASS: cannot use testType3 in examples/example1.go, declared as private to file in examples/example2.go
DO NOT PASS: cannot use testFunctionNotOkay in examples/example1.go, declared as private to file in examples/example2.go
DO NOT PASS: cannot use testFunctionNotOkay in examples/example1.go, declared as private to file in examples/example2.go
DO NOT PASS: cannot use localPrivateStuffTest in examples/example1.go, declared as private to file in examples/example2.go
```

