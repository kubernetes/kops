package goutils

import (
	"fmt"
	"testing"
)

// ****************************** TESTS ********************************************

func TestAbbreviate(t *testing.T) {

	// Test 1
	in := "abcdefg"
	out := "abc..."
	maxWidth := 6

	if x, _ := Abbreviate(in, maxWidth); x != out {
		t.Errorf("Abbreviate(%v, %v) = %v, want %v", in, maxWidth, x, out)
	}

	// Test 2
	out = "abcdefg"
	maxWidth = 7

	if x, _ := Abbreviate(in, maxWidth); x != out {
		t.Errorf("Abbreviate(%v, %v) = %v, want %v", in, maxWidth, x, out)
	}

	// Test 3
	out = "a..."
	maxWidth = 4

	if x, _ := Abbreviate(in, maxWidth); x != out {
		t.Errorf("Abbreviate(%v, %v) = %v, want %v", in, maxWidth, x, out)
	}
}

func TestAbbreviateFull(t *testing.T) {

	// Test 1
	in := "abcdefghijklmno"
	out := "abcdefg..."
	offset := -1
	maxWidth := 10

	if x, _ := AbbreviateFull(in, offset, maxWidth); x != out {
		t.Errorf("AbbreviateFull(%v, %v, %v) = %v, want %v", in, offset, maxWidth, x, out)
	}

	// Test 2
	out = "...fghi..."
	offset = 5
	maxWidth = 10

	if x, _ := AbbreviateFull(in, offset, maxWidth); x != out {
		t.Errorf("AbbreviateFull(%v, %v, %v) = %v, want %v", in, offset, maxWidth, x, out)
	}

	// Test 3
	out = "...ijklmno"
	offset = 12
	maxWidth = 10

	if x, _ := AbbreviateFull(in, offset, maxWidth); x != out {
		t.Errorf("AbbreviateFull(%v, %v, %v) = %v, want %v", in, offset, maxWidth, x, out)
	}
}

func TestIndexOf(t *testing.T) {

	// Test 1
	str := "abcafgka"
	sub := "a"
	start := 0
	out := 0

	if x := IndexOf(str, sub, start); x != out {
		t.Errorf("IndexOf(%v, %v, %v) = %v, want %v", str, sub, start, x, out)
	}

	// Test 2
	start = 1
	out = 3

	if x := IndexOf(str, sub, start); x != out {
		t.Errorf("IndexOf(%v, %v, %v) = %v, want %v", str, sub, start, x, out)
	}

	// Test 3
	start = 4
	out = 7

	if x := IndexOf(str, sub, start); x != out {
		t.Errorf("IndexOf(%v, %v, %v) = %v, want %v", str, sub, start, x, out)
	}

	// Test 4
	sub = "z"
	out = -1

	if x := IndexOf(str, sub, start); x != out {
		t.Errorf("IndexOf(%v, %v, %v) = %v, want %v", str, sub, start, x, out)
	}

}

func TestIsBlank(t *testing.T) {

	// Test 1
	str := ""
	out := true

	if x := IsBlank(str); x != out {
		t.Errorf("IndexOf(%v) = %v, want %v", str, x, out)
	}

	// Test 2
	str = "   "
	out = true

	if x := IsBlank(str); x != out {
		t.Errorf("IndexOf(%v) = %v, want %v", str, x, out)
	}

	// Test 3
	str = " abc "
	out = false

	if x := IsBlank(str); x != out {
		t.Errorf("IndexOf(%v) = %v, want %v", str, x, out)
	}
}

func TestDeleteWhiteSpace(t *testing.T) {

	// Test 1
	str := " a b c "
	out := "abc"

	if x := DeleteWhiteSpace(str); x != out {
		t.Errorf("IndexOf(%v) = %v, want %v", str, x, out)
	}

	// Test 2
	str = "    "
	out = ""

	if x := DeleteWhiteSpace(str); x != out {
		t.Errorf("IndexOf(%v) = %v, want %v", str, x, out)
	}
}

func TestIndexOfDifference(t *testing.T) {

	str1 := "abc"
	str2 := "a_c"
	out := 1

	if x := IndexOfDifference(str1, str2); x != out {
		t.Errorf("IndexOfDifference(%v, %v) = %v, want %v", str1, str2, x, out)
	}
}

// ****************************** EXAMPLES ********************************************

func ExampleAbbreviate() {

	str := "abcdefg"
	out1, _ := Abbreviate(str, 6)
	out2, _ := Abbreviate(str, 7)
	out3, _ := Abbreviate(str, 8)
	out4, _ := Abbreviate(str, 4)
	_, err1 := Abbreviate(str, 3)

	fmt.Println(out1)
	fmt.Println(out2)
	fmt.Println(out3)
	fmt.Println(out4)
	fmt.Println(err1)
	// Output:
	// abc...
	// abcdefg
	// abcdefg
	// a...
	// stringutils illegal argument: Minimum abbreviation width is 4
}

func ExampleAbbreviateFull() {

	str := "abcdefghijklmno"
	str2 := "abcdefghij"
	out1, _ := AbbreviateFull(str, -1, 10)
	out2, _ := AbbreviateFull(str, 0, 10)
	out3, _ := AbbreviateFull(str, 1, 10)
	out4, _ := AbbreviateFull(str, 4, 10)
	out5, _ := AbbreviateFull(str, 5, 10)
	out6, _ := AbbreviateFull(str, 6, 10)
	out7, _ := AbbreviateFull(str, 8, 10)
	out8, _ := AbbreviateFull(str, 10, 10)
	out9, _ := AbbreviateFull(str, 12, 10)
	_, err1 := AbbreviateFull(str2, 0, 3)
	_, err2 := AbbreviateFull(str2, 5, 6)

	fmt.Println(out1)
	fmt.Println(out2)
	fmt.Println(out3)
	fmt.Println(out4)
	fmt.Println(out5)
	fmt.Println(out6)
	fmt.Println(out7)
	fmt.Println(out8)
	fmt.Println(out9)
	fmt.Println(err1)
	fmt.Println(err2)
	// Output:
	// abcdefg...
	// abcdefg...
	// abcdefg...
	// abcdefg...
	// ...fghi...
	// ...ghij...
	// ...ijklmno
	// ...ijklmno
	// ...ijklmno
	// stringutils illegal argument: Minimum abbreviation width is 4
	// stringutils illegal argument: Minimum abbreviation width with offset is 7
}

func ExampleIsBlank() {

	out1 := IsBlank("")
	out2 := IsBlank(" ")
	out3 := IsBlank("bob")
	out4 := IsBlank("  bob  ")

	fmt.Println(out1)
	fmt.Println(out2)
	fmt.Println(out3)
	fmt.Println(out4)
	// Output:
	// true
	// true
	// false
	// false
}

func ExampleDeleteWhiteSpace() {

	out1 := DeleteWhiteSpace(" ")
	out2 := DeleteWhiteSpace("bob")
	out3 := DeleteWhiteSpace("bob   ")
	out4 := DeleteWhiteSpace("  b  o    b  ")

	fmt.Println(out1)
	fmt.Println(out2)
	fmt.Println(out3)
	fmt.Println(out4)
	// Output:
	//
	// bob
	// bob
	// bob
}

func ExampleIndexOf() {

	str := "abcdefgehije"
	out1 := IndexOf(str, "e", 0)
	out2 := IndexOf(str, "e", 5)
	out3 := IndexOf(str, "e", 8)
	out4 := IndexOf(str, "eh", 0)
	out5 := IndexOf(str, "eh", 22)
	out6 := IndexOf(str, "z", 0)
	out7 := IndexOf(str, "", 0)

	fmt.Println(out1)
	fmt.Println(out2)
	fmt.Println(out3)
	fmt.Println(out4)
	fmt.Println(out5)
	fmt.Println(out6)
	fmt.Println(out7)
	// Output:
	// 4
	// 7
	// 11
	// 7
	// -1
	// -1
	// -1
}

func ExampleIndexOfDifference() {

	out1 := IndexOfDifference("abc", "abc")
	out2 := IndexOfDifference("ab", "abxyz")
	out3 := IndexOfDifference("", "abc")
	out4 := IndexOfDifference("abcde", "abxyz")

	fmt.Println(out1)
	fmt.Println(out2)
	fmt.Println(out3)
	fmt.Println(out4)
	// Output:
	// -1
	// 2
	// 0
	// 2
}
