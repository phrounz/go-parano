# go-paranoid-broccoli
Experimental Go static analysis and robustness checker - WARNING still in development and messy.

The first feature is a "private to file" marker which ensure that a function/type/variable declared after this marker cannot be used in another file of the same package.

Test with the source code itself:

```
cd go-paranoid-broccoli
go build && ./go-paranoid-broccoli -dir .
```

It should display something like:
```
FATAL: cannot use testVar in main.go, declared with //!PB_PRIVATE_TO_FILE in parse_file.go
```
