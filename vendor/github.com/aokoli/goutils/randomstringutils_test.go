package goutils

import (
	"fmt"
	"math/rand"
	"testing"
)

// ****************************** TESTS ********************************************

func TestRandomSeed(t *testing.T) {

	// count, start, end, letters, numbers := 5, 0, 0, true, true
	random := rand.New(rand.NewSource(10))
	out := "3ip9v"

	// Test 1: Simulating RandomAlphaNumeric(count int)
	if x, _ := RandomSeed(5, 0, 0, true, true, nil, random); x != out {
		t.Errorf("RandomSeed(%v, %v, %v, %v, %v, %v, %v) = %v, want %v", 5, 0, 0, true, true, nil, random, x, out)
	}

	// Test 2: Simulating RandomAlphabetic(count int)
	out = "MBrbj"

	if x, _ := RandomSeed(5, 0, 0, true, false, nil, random); x != out {
		t.Errorf("RandomSeed(%v, %v, %v, %v, %v, %v, %v) = %v, want %v", 5, 0, 0, true, false, nil, random, x, out)
	}

	// Test 3: Simulating RandomNumeric(count int)
	out = "88935"

	if x, _ := RandomSeed(5, 0, 0, false, true, nil, random); x != out {
		t.Errorf("RandomSeed(%v, %v, %v, %v, %v, %v, %v) = %v, want %v", 5, 0, 0, false, true, nil, random, x, out)
	}

	// Test 4: Simulating RandomAscii(count int)
	out = "H_I;E"

	if x, _ := RandomSeed(5, 32, 127, false, false, nil, random); x != out {
		t.Errorf("RandomSeed(%v, %v, %v, %v, %v, %v, %v) = %v, want %v", 5, 32, 127, false, false, nil, random, x, out)
	}

	// Test 5: Simulating RandomSeed(...) with custom chars
	chars := []rune{'1', '2', '3', 'a', 'b', 'c'}
	out = "2b2ca"

	if x, _ := RandomSeed(5, 0, 0, false, false, chars, random); x != out {
		t.Errorf("RandomSeed(%v, %v, %v, %v, %v, %v, %v) = %v, want %v", 5, 0, 0, false, false, chars, random, x, out)
	}

}

// ****************************** EXAMPLES ********************************************

func ExampleRandomSeed() {

	var seed int64 = 10 // If you change this seed #, the random sequence below will change
	random := rand.New(rand.NewSource(seed))
	chars := []rune{'1', '2', '3', 'a', 'b', 'c'}

	rand1, _ := RandomSeed(5, 0, 0, true, true, nil, random)      // RandomAlphaNumeric (Alphabets and numbers possible)
	rand2, _ := RandomSeed(5, 0, 0, true, false, nil, random)     // RandomAlphabetic (Only alphabets)
	rand3, _ := RandomSeed(5, 0, 0, false, true, nil, random)     // RandomNumeric (Only numbers)
	rand4, _ := RandomSeed(5, 32, 127, false, false, nil, random) // RandomAscii (Alphabets, numbers, and other ASCII chars)
	rand5, _ := RandomSeed(5, 0, 0, true, true, chars, random)    // RandomSeed with custom characters

	fmt.Println(rand1)
	fmt.Println(rand2)
	fmt.Println(rand3)
	fmt.Println(rand4)
	fmt.Println(rand5)
	// Output:
	// 3ip9v
	// MBrbj
	// 88935
	// H_I;E
	// 2b2ca
}
