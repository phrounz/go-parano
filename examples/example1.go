package main

import (
	"./examplesub"
)

//!PARANO__PRIVATE_TO_FILE
//!PARANO__EXHAUSTIVE_FILLING
type testType1 struct {
	foo1 int
	foo2 int
	foo3 int
	foo4 string
}

func main() {

	var tts = examplesub.TestTypeSub{Foo1: 1}
	tts.Foo2 = 2

	// test exhaustive filling
	var testNokFill = testType1{
		foo1: 3,
		foo2: 3,
	}
	testNokFill.foo2 = 4
	var testOkFill = testType1{
		foo3: 3,
		foo2: 0,
		foo1: 3,
		foo4: "",
	}
	testOkFill.foo3 = 4

	// test private to file
	testVarOkay = true
	testVarNotOkay = true
	var tata testType2
	tata.foo = 1
	var tata3 testType3
	if tata3 == 456 {
	}
	testFunctionNotOkay()

	localPrivateStuffTest = true

	examplesub.Query("SELECT \"\\1\", el.foo FROM elements el")
	examplesub.Query("SELECT FROM JOIN \"1\"")
}
