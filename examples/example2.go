package main

var testVarOkay bool

//!PB_PRIVATE_TO_FILE
var testVarNotOkay bool

//!PB_PRIVATE_TO_FILE
func testFunctionNotOkay() {
}

//!PB_PRIVATE_TO_FILE
type testType2 struct {
	foo int
}

//!PB_PRIVATE_TO_FILE
type testType3 int
