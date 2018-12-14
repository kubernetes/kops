package sprig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTuple(t *testing.T) {
	tpl := `{{$t := tuple 1 "a" "foo"}}{{index $t 2}}{{index $t 0 }}{{index $t 1}}`
	if err := runt(tpl, "foo1a"); err != nil {
		t.Error(err)
	}
}

func TestList(t *testing.T) {
	tpl := `{{$t := list 1 "a" "foo"}}{{index $t 2}}{{index $t 0 }}{{index $t 1}}`
	if err := runt(tpl, "foo1a"); err != nil {
		t.Error(err)
	}
}

func TestPush(t *testing.T) {
	// Named `append` in the function map
	tests := map[string]string{
		`{{ $t := tuple 1 2 3  }}{{ append $t 4 | len }}`:        "4",
		`{{ $t := tuple 1 2 3 4  }}{{ append $t 5 | join "-" }}`: "1-2-3-4-5",
	}
	for tpl, expect := range tests {
		assert.NoError(t, runt(tpl, expect))
	}
}
func TestPrepend(t *testing.T) {
	tests := map[string]string{
		`{{ $t := tuple 1 2 3  }}{{ prepend $t 0 | len }}`:        "4",
		`{{ $t := tuple 1 2 3 4  }}{{ prepend $t 0 | join "-" }}`: "0-1-2-3-4",
	}
	for tpl, expect := range tests {
		assert.NoError(t, runt(tpl, expect))
	}
}

func TestFirst(t *testing.T) {
	tests := map[string]string{
		`{{ list 1 2 3 | first }}`: "1",
		`{{ list | first }}`:       "<no value>",
	}
	for tpl, expect := range tests {
		assert.NoError(t, runt(tpl, expect))
	}
}
func TestLast(t *testing.T) {
	tests := map[string]string{
		`{{ list 1 2 3 | last }}`: "3",
		`{{ list | last }}`:       "<no value>",
	}
	for tpl, expect := range tests {
		assert.NoError(t, runt(tpl, expect))
	}
}

func TestInitial(t *testing.T) {
	tests := map[string]string{
		`{{ list 1 2 3 | initial | len }}`:   "2",
		`{{ list 1 2 3 | initial | last }}`:  "2",
		`{{ list 1 2 3 | initial | first }}`: "1",
		`{{ list | initial }}`:               "[]",
	}
	for tpl, expect := range tests {
		assert.NoError(t, runt(tpl, expect))
	}
}

func TestRest(t *testing.T) {
	tests := map[string]string{
		`{{ list 1 2 3 | rest | len }}`:   "2",
		`{{ list 1 2 3 | rest | last }}`:  "3",
		`{{ list 1 2 3 | rest | first }}`: "2",
		`{{ list | rest }}`:               "[]",
	}
	for tpl, expect := range tests {
		assert.NoError(t, runt(tpl, expect))
	}
}

func TestReverse(t *testing.T) {
	tests := map[string]string{
		`{{ list 1 2 3 | reverse | first }}`:        "3",
		`{{ list 1 2 3 | reverse | rest | first }}`: "2",
		`{{ list 1 2 3 | reverse | last }}`:         "1",
		`{{ list 1 2 3 4 | reverse }}`:              "[4 3 2 1]",
		`{{ list 1 | reverse }}`:                    "[1]",
		`{{ list | reverse }}`:                      "[]",
	}
	for tpl, expect := range tests {
		assert.NoError(t, runt(tpl, expect))
	}
}

func TestUniq(t *testing.T) {
	tests := map[string]string{
		`{{ list 1 2 3 4 | uniq }}`:                   `[1 2 3 4]`,
		`{{ list "a" "b" "c" "d" | uniq }}`:           `[a b c d]`,
		`{{ list 1 1 1 1 2 2 2 2 | uniq }}`:           `[1 2]`,
		`{{ list "foo" 1 1 1 1 "foo" "foo" | uniq }}`: `[foo 1]`,
		`{{ list | uniq }}`:                           `[]`,
	}
	for tpl, expect := range tests {
		assert.NoError(t, runt(tpl, expect))
	}
}

func TestWithout(t *testing.T) {
	tests := map[string]string{
		`{{ without (list 1 2 3 4) 1 }}`:           `[2 3 4]`,
		`{{ without (list "a" "b" "c" "d") "a" }}`: `[b c d]`,
		`{{ without (list 1 1 1 1 2) 1 }}`:         `[2]`,
		`{{ without (list) 1 }}`:                   `[]`,
		`{{ without (list 1 2 3) }}`:               `[1 2 3]`,
		`{{ without list }}`:                       `[]`,
	}
	for tpl, expect := range tests {
		assert.NoError(t, runt(tpl, expect))
	}
}

func TestHas(t *testing.T) {
	tests := map[string]string{
		`{{ list 1 2 3 | has 1 }}`: `true`,
		`{{ list 1 2 3 | has 4 }}`: `false`,
	}
	for tpl, expect := range tests {
		assert.NoError(t, runt(tpl, expect))
	}
}
