package goutils

import (
	"fmt"
	"testing"
)

// ****************************** TESTS ********************************************

func TestWrapNormalWord(t *testing.T) {

	in := "Bob Manuel Bob Manuel"
	out := "Bob Manuel\nBob Manuel"
	wrapLength := 10

	if x := Wrap(in, wrapLength); x != out {
		t.Errorf("Wrap(%v) = %v, want %v", in, x, out)
	}
}

func TestWrapCustomLongWordFalse(t *testing.T) {

	in := "BobManuelBob Bob"
	out := "BobManuelBob<br\\>Bob"
	wrapLength := 10
	newLineStr := "<br\\>"
	wrapLongWords := false

	if x := WrapCustom(in, wrapLength, newLineStr, wrapLongWords); x != out {
		t.Errorf("Wrap(%v) = %v, want %v", in, x, out)
	}
}

func TestWrapCustomLongWordTrue(t *testing.T) {

	in := "BobManuelBob Bob"
	out := "BobManuelB<br\\>ob Bob"
	wrapLength := 10
	newLineStr := "<br\\>"
	wrapLongWords := true

	if x := WrapCustom(in, wrapLength, newLineStr, wrapLongWords); x != out {
		t.Errorf("WrapCustom(%v) = %v, want %v", in, x, out)
	}
}

func TestCapitalize(t *testing.T) {

	// Test 1: Checks if function works with 1 parameter, and default whitespace delimiter
	in := "test is going.well.thank.you.for inquiring"
	out := "Test Is Going.well.thank.you.for Inquiring"

	if x := Capitalize(in); x != out {
		t.Errorf("Capitalize(%v) = %v, want %v", in, x, out)
	}

	// Test 2: Checks if function works with both parameters, with param 2 containing whitespace and '.'
	out = "Test Is Going.Well.Thank.You.For Inquiring"
	delimiters := []rune{' ', '.'}

	if x := Capitalize(in, delimiters...); x != out {
		t.Errorf("Capitalize(%v) = %v, want %v", in, x, out)
	}
}

func TestCapitalizeFully(t *testing.T) {

	// Test 1
	in := "tEsT iS goiNG.wELL.tHaNk.yOU.for inqUIrING"
	out := "Test Is Going.well.thank.you.for Inquiring"

	if x := CapitalizeFully(in); x != out {
		t.Errorf("CapitalizeFully(%v) = %v, want %v", in, x, out)
	}

	// Test 2
	out = "Test Is Going.Well.Thank.You.For Inquiring"
	delimiters := []rune{' ', '.'}

	if x := CapitalizeFully(in, delimiters...); x != out {
		t.Errorf("CapitalizeFully(%v) = %v, want %v", in, x, out)
	}
}

func TestUncapitalize(t *testing.T) {

	// Test 1: Checks if function works with 1 parameter, and default whitespace delimiter
	in := "This Is A.Test"
	out := "this is a.Test"

	if x := Uncapitalize(in); x != out {
		t.Errorf("Uncapitalize(%v) = %v, want %v", in, x, out)
	}

	// Test 2: Checks if function works with both parameters, with param 2 containing whitespace and '.'
	out = "this is a.test"
	delimiters := []rune{' ', '.'}

	if x := Uncapitalize(in, delimiters...); x != out {
		t.Errorf("Uncapitalize(%v) = %v, want %v", in, x, out)
	}
}

func TestSwapCase(t *testing.T) {

	in := "This Is A.Test"
	out := "tHIS iS a.tEST"

	if x := SwapCase(in); x != out {
		t.Errorf("SwapCase(%v) = %v, want %v", in, x, out)
	}
}

func TestInitials(t *testing.T) {

	// Test 1
	in := "John Doe.Ray"
	out := "JD"

	if x := Initials(in); x != out {
		t.Errorf("Initials(%v) = %v, want %v", in, x, out)
	}

	// Test 2
	out = "JDR"
	delimiters := []rune{' ', '.'}

	if x := Initials(in, delimiters...); x != out {
		t.Errorf("Initials(%v) = %v, want %v", in, x, out)
	}

}

// ****************************** EXAMPLES ********************************************

func ExampleWrap() {

	in := "Bob Manuel Bob Manuel"
	wrapLength := 10

	fmt.Println(Wrap(in, wrapLength))
	// Output:
	// Bob Manuel
	// Bob Manuel
}

func ExampleWrapCustom_1() {

	in := "BobManuelBob Bob"
	wrapLength := 10
	newLineStr := "<br\\>"
	wrapLongWords := false

	fmt.Println(WrapCustom(in, wrapLength, newLineStr, wrapLongWords))
	// Output:
	// BobManuelBob<br\>Bob
}

func ExampleWrapCustom_2() {

	in := "BobManuelBob Bob"
	wrapLength := 10
	newLineStr := "<br\\>"
	wrapLongWords := true

	fmt.Println(WrapCustom(in, wrapLength, newLineStr, wrapLongWords))
	// Output:
	// BobManuelB<br\>ob Bob
}

func ExampleCapitalize() {

	in := "test is going.well.thank.you.for inquiring" // Compare input to CapitalizeFully example
	delimiters := []rune{' ', '.'}

	fmt.Println(Capitalize(in))
	fmt.Println(Capitalize(in, delimiters...))
	// Output:
	// Test Is Going.well.thank.you.for Inquiring
	// Test Is Going.Well.Thank.You.For Inquiring
}

func ExampleCapitalizeFully() {

	in := "tEsT iS goiNG.wELL.tHaNk.yOU.for inqUIrING" // Notice scattered capitalization
	delimiters := []rune{' ', '.'}

	fmt.Println(CapitalizeFully(in))
	fmt.Println(CapitalizeFully(in, delimiters...))
	// Output:
	// Test Is Going.well.thank.you.for Inquiring
	// Test Is Going.Well.Thank.You.For Inquiring
}

func ExampleUncapitalize() {

	in := "This Is A.Test"
	delimiters := []rune{' ', '.'}

	fmt.Println(Uncapitalize(in))
	fmt.Println(Uncapitalize(in, delimiters...))
	// Output:
	// this is a.Test
	// this is a.test
}

func ExampleSwapCase() {

	in := "This Is A.Test"
	fmt.Println(SwapCase(in))
	// Output:
	// tHIS iS a.tEST
}

func ExampleInitials() {

	in := "John Doe.Ray"
	delimiters := []rune{' ', '.'}

	fmt.Println(Initials(in))
	fmt.Println(Initials(in, delimiters...))
	// Output:
	// JD
	// JDR
}
