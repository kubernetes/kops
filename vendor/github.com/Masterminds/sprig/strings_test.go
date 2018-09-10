package sprig

import (
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"math/rand"
	"testing"

	"github.com/aokoli/goutils"
	"github.com/stretchr/testify/assert"
)

func TestSubstr(t *testing.T) {
	tpl := `{{"fooo" | substr 0 3 }}`
	if err := runt(tpl, "foo"); err != nil {
		t.Error(err)
	}
}

func TestTrunc(t *testing.T) {
	tpl := `{{ "foooooo" | trunc 3 }}`
	if err := runt(tpl, "foo"); err != nil {
		t.Error(err)
	}
}

func TestQuote(t *testing.T) {
	tpl := `{{quote "a" "b" "c"}}`
	if err := runt(tpl, `"a" "b" "c"`); err != nil {
		t.Error(err)
	}
	tpl = `{{quote "\"a\"" "b" "c"}}`
	if err := runt(tpl, `"\"a\"" "b" "c"`); err != nil {
		t.Error(err)
	}
	tpl = `{{quote 1 2 3 }}`
	if err := runt(tpl, `"1" "2" "3"`); err != nil {
		t.Error(err)
	}
}
func TestSquote(t *testing.T) {
	tpl := `{{squote "a" "b" "c"}}`
	if err := runt(tpl, `'a' 'b' 'c'`); err != nil {
		t.Error(err)
	}
	tpl = `{{squote 1 2 3 }}`
	if err := runt(tpl, `'1' '2' '3'`); err != nil {
		t.Error(err)
	}
}

func TestContains(t *testing.T) {
	// Mainly, we're just verifying the paramater order swap.
	tests := []string{
		`{{if contains "cat" "fair catch"}}1{{end}}`,
		`{{if hasPrefix "cat" "catch"}}1{{end}}`,
		`{{if hasSuffix "cat" "ducat"}}1{{end}}`,
	}
	for _, tt := range tests {
		if err := runt(tt, "1"); err != nil {
			t.Error(err)
		}
	}
}

func TestTrim(t *testing.T) {
	tests := []string{
		`{{trim "   5.00   "}}`,
		`{{trimAll "$" "$5.00$"}}`,
		`{{trimPrefix "$" "$5.00"}}`,
		`{{trimSuffix "$" "5.00$"}}`,
	}
	for _, tt := range tests {
		if err := runt(tt, "5.00"); err != nil {
			t.Error(err)
		}
	}
}

func TestSplit(t *testing.T) {
	tpl := `{{$v := "foo$bar$baz" | split "$"}}{{$v._0}}`
	if err := runt(tpl, "foo"); err != nil {
		t.Error(err)
	}
}

func TestToString(t *testing.T) {
	tpl := `{{ toString 1 | kindOf }}`
	assert.NoError(t, runt(tpl, "string"))
}

func TestToStrings(t *testing.T) {
	tpl := `{{ $s := list 1 2 3 | toStrings }}{{ index $s 1 | kindOf }}`
	assert.NoError(t, runt(tpl, "string"))
}

func TestJoin(t *testing.T) {
	assert.NoError(t, runt(`{{ tuple "a" "b" "c" | join "-" }}`, "a-b-c"))
	assert.NoError(t, runt(`{{ tuple 1 2 3 | join "-" }}`, "1-2-3"))
	assert.NoError(t, runtv(`{{ join "-" .V }}`, "a-b-c", map[string]interface{}{"V": []string{"a", "b", "c"}}))
	assert.NoError(t, runtv(`{{ join "-" .V }}`, "abc", map[string]interface{}{"V": "abc"}))
	assert.NoError(t, runtv(`{{ join "-" .V }}`, "1-2-3", map[string]interface{}{"V": []int{1, 2, 3}}))
}

func TestSortAlpha(t *testing.T) {
	// Named `append` in the function map
	tests := map[string]string{
		`{{ list "c" "a" "b" | sortAlpha | join "" }}`: "abc",
		`{{ list 2 1 4 3 | sortAlpha | join "" }}`:     "1234",
	}
	for tpl, expect := range tests {
		assert.NoError(t, runt(tpl, expect))
	}
}
func TestBase64EncodeDecode(t *testing.T) {
	magicWord := "coffee"
	expect := base64.StdEncoding.EncodeToString([]byte(magicWord))

	if expect == magicWord {
		t.Fatal("Encoder doesn't work.")
	}

	tpl := `{{b64enc "coffee"}}`
	if err := runt(tpl, expect); err != nil {
		t.Error(err)
	}
	tpl = fmt.Sprintf("{{b64dec %q}}", expect)
	if err := runt(tpl, magicWord); err != nil {
		t.Error(err)
	}
}
func TestBase32EncodeDecode(t *testing.T) {
	magicWord := "coffee"
	expect := base32.StdEncoding.EncodeToString([]byte(magicWord))

	if expect == magicWord {
		t.Fatal("Encoder doesn't work.")
	}

	tpl := `{{b32enc "coffee"}}`
	if err := runt(tpl, expect); err != nil {
		t.Error(err)
	}
	tpl = fmt.Sprintf("{{b32dec %q}}", expect)
	if err := runt(tpl, magicWord); err != nil {
		t.Error(err)
	}
}

func TestGoutils(t *testing.T) {
	tests := map[string]string{
		`{{abbrev 5 "hello world"}}`:           "he...",
		`{{abbrevboth 5 10 "1234 5678 9123"}}`: "...5678...",
		`{{nospace "h e l l o "}}`:             "hello",
		`{{untitle "First Try"}}`:              "first try", //https://youtu.be/44-RsrF_V_w
		`{{initials "First Try"}}`:             "FT",
		`{{wrap 5 "Hello World"}}`:             "Hello\nWorld",
		`{{wrapWith 5 "\t" "Hello World"}}`:    "Hello\tWorld",
	}
	for k, v := range tests {
		t.Log(k)
		if err := runt(k, v); err != nil {
			t.Errorf("Error on tpl %s: %s", err)
		}
	}
}

func TestRandom(t *testing.T) {
	// One of the things I love about Go:
	goutils.RANDOM = rand.New(rand.NewSource(1))

	// Because we're using a random number generator, we need these to go in
	// a predictable sequence:
	if err := runt(`{{randAlphaNum 5}}`, "9bzRv"); err != nil {
		t.Errorf("Error on tpl %s: %s", err)
	}
	if err := runt(`{{randAlpha 5}}`, "VjwGe"); err != nil {
		t.Errorf("Error on tpl %s: %s", err)
	}
	if err := runt(`{{randAscii 5}}`, "1KA5p"); err != nil {
		t.Errorf("Error on tpl %s: %s", err)
	}
	if err := runt(`{{randNumeric 5}}`, "26018"); err != nil {
		t.Errorf("Error on tpl %s: %s", err)
	}

}

func TestCat(t *testing.T) {
	tpl := `{{$b := "b"}}{{"c" | cat "a" $b}}`
	if err := runt(tpl, "a b c"); err != nil {
		t.Error(err)
	}
}

func TestIndent(t *testing.T) {
	tpl := `{{indent 4 "a\nb\nc"}}`
	if err := runt(tpl, "    a\n    b\n    c"); err != nil {
		t.Error(err)
	}
}

func TestReplace(t *testing.T) {
	tpl := `{{"I Am Henry VIII" | replace " " "-"}}`
	if err := runt(tpl, "I-Am-Henry-VIII"); err != nil {
		t.Error(err)
	}
}

func TestPlural(t *testing.T) {
	tpl := `{{$num := len "two"}}{{$num}} {{$num | plural "1 char" "chars"}}`
	if err := runt(tpl, "3 chars"); err != nil {
		t.Error(err)
	}
	tpl = `{{len "t" | plural "cheese" "%d chars"}}`
	if err := runt(tpl, "cheese"); err != nil {
		t.Error(err)
	}
}
