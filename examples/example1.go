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

	var banana = 4

	localPrivateStuffTest = true

	var _ = examplesub.Query(true, "SELECT \"\\1\", el.foo "+
		"FROM elements "+` el `+ // +strconv.Itoa(testOkFill.foo3)
		"WHERE el.bar=12",
		true,
		"test")
	var _ = examplesub.Query(true, "SELECT FROM JOIN \"1\"", true, "test")
	examplesub.QueryNoAnswer("INSERT INTO elements (`foo`,`bar`) VALUES (1,2)")
	examplesub.QueryNoAnswer("INSERT INTO elements typo mistake (`foo`,`bar`) VALUES (1,2)") //comment
	foo()
	examplesub.Query(true, "SELECTYYYY", //!PARANO__IGNORE_CHECK_SQL_QUERY
		true, "test")
	examplesub.QueryNoAnswer( //!PARANO__IGNORE_CHECK_SQL_QUERY
		"SELECTYYYY")

	//const constElements = "elements"
	var varElements = "elements"
	examplesub.Query(true, "SELECT * FROM "+string("elements"), true, "test")
	examplesub.Query(true, "SELECT * FROM "+constElements, true, "test")
	examplesub.Query(true, "SELECT * FROM "+varElements, true, "test")
}

//!PARANO__IGNORE_CHECK_SQL_QUERIES
func foo() {
	examplesub.QueryNoAnswer("SELECTZZZZ")
}
