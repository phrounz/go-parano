# go-paranoid-broccoli
Experimental Go static analysis and robustness checker - WARNING still in development and messy.

**Feature: private to file**

The first and currently only feature is a "private to file" marker: a function/type/variable declared as "private to file" cannot be used in another file of the same package.

For example:
```
//!PB_PRIVATE_TO_FILE
var i int
```
disallows i to be used in other files of the same package.

**How to test**

You can test with the source code itself:

```
cd go-paranoid-broccoli
go build && ./go-paranoid-broccoli -dir .
```

It should display something like:
```
FATAL: cannot use testVar in main.go, declared with //!PB_PRIVATE_TO_FILE in parse_file.go
```
