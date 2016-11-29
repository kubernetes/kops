package util

import "testing"

func TestStringSlicesEqual(t *testing.T) {
	cases := []struct {
		l, r     []string
		expected bool
	}{
		{
			[]string{"a", "b", "c"},
			[]string{"a", "b", "c"},
			true,
		},
		{
			[]string{"a", "b", "c"},
			[]string{"x", "y", "z"},
			false,
		},
		{
			[]string{"a", "b"},
			[]string{"a", "b", "c"},
			false,
		},
		{
			[]string{"", "", ""},
			[]string{"", "", ""},
			true,
		},
	}

	for _, c := range cases {
		if actual := StringSlicesEqual(c.l, c.r); actual != c.expected {
			t.Errorf("StringSlicesEqual(%#v, %#v) expected %#v, but got %#v", c.l, c.r, c.expected, actual)
		}
	}
}
