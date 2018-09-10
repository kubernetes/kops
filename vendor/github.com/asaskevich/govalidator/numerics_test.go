package govalidator

import "testing"

func TestAbs(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		param    float64
		expected float64
	}{
		{0, 0},
		{-1, 1},
		{10, 10},
		{3.14, 3.14},
		{-96, 96},
		{-10e-12, 10e-12},
	}
	for _, test := range tests {
		actual := Abs(test.param)
		if actual != test.expected {
			t.Errorf("Expected Abs(%v) to be %v, got %v", test.param, test.expected, actual)
		}
	}
}

func TestSign(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		param    float64
		expected float64
	}{
		{0, 0},
		{-1, -1},
		{10, 1},
		{3.14, 1},
		{-96, -1},
		{-10e-12, -1},
	}
	for _, test := range tests {
		actual := Sign(test.param)
		if actual != test.expected {
			t.Errorf("Expected Sign(%v) to be %v, got %v", test.param, test.expected, actual)
		}
	}
}

func TestIsNegative(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		param    float64
		expected bool
	}{
		{0, false},
		{-1, true},
		{10, false},
		{3.14, false},
		{-96, true},
		{-10e-12, true},
	}
	for _, test := range tests {
		actual := IsNegative(test.param)
		if actual != test.expected {
			t.Errorf("Expected IsNegative(%v) to be %v, got %v", test.param, test.expected, actual)
		}
	}
}

func TestIsNonNegative(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		param    float64
		expected bool
	}{
		{0, true},
		{-1, false},
		{10, true},
		{3.14, true},
		{-96, false},
		{-10e-12, false},
	}
	for _, test := range tests {
		actual := IsNonNegative(test.param)
		if actual != test.expected {
			t.Errorf("Expected IsNonNegative(%v) to be %v, got %v", test.param, test.expected, actual)
		}
	}
}

func TestIsPositive(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		param    float64
		expected bool
	}{
		{0, false},
		{-1, false},
		{10, true},
		{3.14, true},
		{-96, false},
		{-10e-12, false},
	}
	for _, test := range tests {
		actual := IsPositive(test.param)
		if actual != test.expected {
			t.Errorf("Expected IsPositive(%v) to be %v, got %v", test.param, test.expected, actual)
		}
	}
}

func TestIsNonPositive(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		param    float64
		expected bool
	}{
		{0, true},
		{-1, true},
		{10, false},
		{3.14, false},
		{-96, true},
		{-10e-12, true},
	}
	for _, test := range tests {
		actual := IsNonPositive(test.param)
		if actual != test.expected {
			t.Errorf("Expected IsNonPositive(%v) to be %v, got %v", test.param, test.expected, actual)
		}
	}
}

func TestIsWhole(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		param    float64
		expected bool
	}{
		{0, true},
		{-1, true},
		{10, true},
		{3.14, false},
		{-96, true},
		{-10e-12, false},
	}
	for _, test := range tests {
		actual := IsWhole(test.param)
		if actual != test.expected {
			t.Errorf("Expected IsWhole(%v) to be %v, got %v", test.param, test.expected, actual)
		}
	}
}

func TestIsNatural(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		param    float64
		expected bool
	}{
		{0, false},
		{-1, false},
		{10, true},
		{3.14, false},
		{96, true},
		{-10e-12, false},
	}
	for _, test := range tests {
		actual := IsNatural(test.param)
		if actual != test.expected {
			t.Errorf("Expected IsNatural(%v) to be %v, got %v", test.param, test.expected, actual)
		}
	}
}

func TestInRangeInt(t *testing.T) {
	t.Parallel()

	var testAsInts = []struct {
		param    int
		left     int
		right    int
		expected bool
	}{
		{0, 0, 0, true},
		{1, 0, 0, false},
		{-1, 0, 0, false},
		{0, -1, 1, true},
		{0, 0, 1, true},
		{0, -1, 0, true},
		{0, 0, -1, true},
		{0, 10, 5, false},
	}
	for _, test := range testAsInts {
		actual := InRangeInt(test.param, test.left, test.right)
		if actual != test.expected {
			t.Errorf("Expected InRangeInt(%v, %v, %v) to be %v, got %v using type int", test.param, test.left, test.right, test.expected, actual)
		}
	}

	var testAsInt8s = []struct {
		param    int8
		left     int8
		right    int8
		expected bool
	}{
		{0, 0, 0, true},
		{1, 0, 0, false},
		{-1, 0, 0, false},
		{0, -1, 1, true},
		{0, 0, 1, true},
		{0, -1, 0, true},
		{0, 0, -1, true},
		{0, 10, 5, false},
	}
	for _, test := range testAsInt8s {
		actual := InRangeInt(test.param, test.left, test.right)
		if actual != test.expected {
			t.Errorf("Expected InRangeInt(%v, %v, %v) to be %v, got %v using type int8", test.param, test.left, test.right, test.expected, actual)
		}
	}

	var testAsInt16s = []struct {
		param    int16
		left     int16
		right    int16
		expected bool
	}{
		{0, 0, 0, true},
		{1, 0, 0, false},
		{-1, 0, 0, false},
		{0, -1, 1, true},
		{0, 0, 1, true},
		{0, -1, 0, true},
		{0, 0, -1, true},
		{0, 10, 5, false},
	}
	for _, test := range testAsInt16s {
		actual := InRangeInt(test.param, test.left, test.right)
		if actual != test.expected {
			t.Errorf("Expected InRangeInt(%v, %v, %v) to be %v, got %v using type int16", test.param, test.left, test.right, test.expected, actual)
		}
	}

	var testAsInt32s = []struct {
		param    int32
		left     int32
		right    int32
		expected bool
	}{
		{0, 0, 0, true},
		{1, 0, 0, false},
		{-1, 0, 0, false},
		{0, -1, 1, true},
		{0, 0, 1, true},
		{0, -1, 0, true},
		{0, 0, -1, true},
		{0, 10, 5, false},
	}
	for _, test := range testAsInt32s {
		actual := InRangeInt(test.param, test.left, test.right)
		if actual != test.expected {
			t.Errorf("Expected InRangeInt(%v, %v, %v) to be %v, got %v using type int32", test.param, test.left, test.right, test.expected, actual)
		}
	}

	var testAsInt64s = []struct {
		param    int64
		left     int64
		right    int64
		expected bool
	}{
		{0, 0, 0, true},
		{1, 0, 0, false},
		{-1, 0, 0, false},
		{0, -1, 1, true},
		{0, 0, 1, true},
		{0, -1, 0, true},
		{0, 0, -1, true},
		{0, 10, 5, false},
	}
	for _, test := range testAsInt64s {
		actual := InRangeInt(test.param, test.left, test.right)
		if actual != test.expected {
			t.Errorf("Expected InRangeInt(%v, %v, %v) to be %v, got %v using type int64", test.param, test.left, test.right, test.expected, actual)
		}
	}

	var testAsUInts = []struct {
		param    uint
		left     uint
		right    uint
		expected bool
	}{
		{0, 0, 0, true},
		{1, 0, 0, false},
		{0, 0, 1, true},
		{0, 10, 5, false},
	}
	for _, test := range testAsUInts {
		actual := InRangeInt(test.param, test.left, test.right)
		if actual != test.expected {
			t.Errorf("Expected InRangeInt(%v, %v, %v) to be %v, got %v using type uint", test.param, test.left, test.right, test.expected, actual)
		}
	}

	var testAsUInt8s = []struct {
		param    uint8
		left     uint8
		right    uint8
		expected bool
	}{
		{0, 0, 0, true},
		{1, 0, 0, false},
		{0, 0, 1, true},
		{0, 10, 5, false},
	}
	for _, test := range testAsUInt8s {
		actual := InRangeInt(test.param, test.left, test.right)
		if actual != test.expected {
			t.Errorf("Expected InRangeInt(%v, %v, %v) to be %v, got %v using type uint", test.param, test.left, test.right, test.expected, actual)
		}
	}

	var testAsUInt16s = []struct {
		param    uint16
		left     uint16
		right    uint16
		expected bool
	}{
		{0, 0, 0, true},
		{1, 0, 0, false},
		{0, 0, 1, true},
		{0, 10, 5, false},
	}
	for _, test := range testAsUInt16s {
		actual := InRangeInt(test.param, test.left, test.right)
		if actual != test.expected {
			t.Errorf("Expected InRangeInt(%v, %v, %v) to be %v, got %v using type uint", test.param, test.left, test.right, test.expected, actual)
		}
	}

	var testAsUInt32s = []struct {
		param    uint32
		left     uint32
		right    uint32
		expected bool
	}{
		{0, 0, 0, true},
		{1, 0, 0, false},
		{0, 0, 1, true},
		{0, 10, 5, false},
	}
	for _, test := range testAsUInt32s {
		actual := InRangeInt(test.param, test.left, test.right)
		if actual != test.expected {
			t.Errorf("Expected InRangeInt(%v, %v, %v) to be %v, got %v using type uint", test.param, test.left, test.right, test.expected, actual)
		}
	}

	var testAsUInt64s = []struct {
		param    uint64
		left     uint64
		right    uint64
		expected bool
	}{
		{0, 0, 0, true},
		{1, 0, 0, false},
		{0, 0, 1, true},
		{0, 10, 5, false},
	}
	for _, test := range testAsUInt64s {
		actual := InRangeInt(test.param, test.left, test.right)
		if actual != test.expected {
			t.Errorf("Expected InRangeInt(%v, %v, %v) to be %v, got %v using type uint", test.param, test.left, test.right, test.expected, actual)
		}
	}

	var testAsStrings = []struct {
		param    string
		left     string
		right    string
		expected bool
	}{
		{"0", "0", "0", true},
		{"1", "0", "0", false},
		{"-1", "0", "0", false},
		{"0", "-1", "1", true},
		{"0", "0", "1", true},
		{"0", "-1", "0", true},
		{"0", "0", "-1", true},
		{"0", "10", "5", false},
	}
	for _, test := range testAsStrings {
		actual := InRangeInt(test.param, test.left, test.right)
		if actual != test.expected {
			t.Errorf("Expected InRangeInt(%v, %v, %v) to be %v, got %v using type string", test.param, test.left, test.right, test.expected, actual)
		}
	}
}

func TestInRangeFloat32(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		param    float32
		left     float32
		right    float32
		expected bool
	}{
		{0, 0, 0, true},
		{1, 0, 0, false},
		{-1, 0, 0, false},
		{0, -1, 1, true},
		{0, 0, 1, true},
		{0, -1, 0, true},
		{0, 0, -1, true},
		{0, 10, 5, false},
	}
	for _, test := range tests {
		actual := InRangeFloat32(test.param, test.left, test.right)
		if actual != test.expected {
			t.Errorf("Expected InRangeFloat32(%v, %v, %v) to be %v, got %v", test.param, test.left, test.right, test.expected, actual)
		}
	}
}

func TestInRangeFloat64(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		param    float64
		left     float64
		right    float64
		expected bool
	}{
		{0, 0, 0, true},
		{1, 0, 0, false},
		{-1, 0, 0, false},
		{0, -1, 1, true},
		{0, 0, 1, true},
		{0, -1, 0, true},
		{0, 0, -1, true},
		{0, 10, 5, false},
	}
	for _, test := range tests {
		actual := InRangeFloat64(test.param, test.left, test.right)
		if actual != test.expected {
			t.Errorf("Expected InRangeFloat64(%v, %v, %v) to be %v, got %v", test.param, test.left, test.right, test.expected, actual)
		}
	}
}

func TestInRange(t *testing.T) {
	t.Parallel()

	var testsInt = []struct {
		param    int
		left     int
		right    int
		expected bool
	}{
		{0, 0, 0, true},
		{1, 0, 0, false},
		{-1, 0, 0, false},
		{0, -1, 1, true},
		{0, 0, 1, true},
		{0, -1, 0, true},
		{0, 0, -1, true},
		{0, 10, 5, false},
	}
	for _, test := range testsInt {
		actual := InRange(test.param, test.left, test.right)
		if actual != test.expected {
			t.Errorf("Expected InRange(%v, %v, %v) to be %v, got %v", test.param, test.left, test.right, test.expected, actual)
		}
	}

	var testsFloat32 = []struct {
		param    float32
		left     float32
		right    float32
		expected bool
	}{
		{0, 0, 0, true},
		{1, 0, 0, false},
		{-1, 0, 0, false},
		{0, -1, 1, true},
		{0, 0, 1, true},
		{0, -1, 0, true},
		{0, 0, -1, true},
		{0, 10, 5, false},
	}
	for _, test := range testsFloat32 {
		actual := InRange(test.param, test.left, test.right)
		if actual != test.expected {
			t.Errorf("Expected InRange(%v, %v, %v) to be %v, got %v", test.param, test.left, test.right, test.expected, actual)
		}
	}

	var testsFloat64 = []struct {
		param    float64
		left     float64
		right    float64
		expected bool
	}{
		{0, 0, 0, true},
		{1, 0, 0, false},
		{-1, 0, 0, false},
		{0, -1, 1, true},
		{0, 0, 1, true},
		{0, -1, 0, true},
		{0, 0, -1, true},
		{0, 10, 5, false},
	}
	for _, test := range testsFloat64 {
		actual := InRange(test.param, test.left, test.right)
		if actual != test.expected {
			t.Errorf("Expected InRange(%v, %v, %v) to be %v, got %v", test.param, test.left, test.right, test.expected, actual)
		}
	}

	var testsTypeMix = []struct {
		param    int
		left     float64
		right    float64
		expected bool
	}{
		{0, 0, 0, false},
		{1, 0, 0, false},
		{-1, 0, 0, false},
		{0, -1, 1, false},
		{0, 0, 1, false},
		{0, -1, 0, false},
		{0, 0, -1, false},
		{0, 10, 5, false},
	}
	for _, test := range testsTypeMix {
		actual := InRange(test.param, test.left, test.right)
		if actual != test.expected {
			t.Errorf("Expected InRange(%v, %v, %v) to be %v, got %v", test.param, test.left, test.right, test.expected, actual)
		}
	}
}
