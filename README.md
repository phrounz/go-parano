# go-paranoid-broccoli
Experimental Go static analysis and robustness checker - WARNING still in development and messy.

The first feature is a "private to file" marker which ensure that a function/type/variable declared after this marker cannot be used in another file of the same package.

Test with:

```
cd go-paranoid-broccoli
go build && ./go-paranoid-broccoli
```

It should display something like:
```
FATAL: cannot call testFunction in main.go, declared in parse_file.go after privateToFileMarker
```

