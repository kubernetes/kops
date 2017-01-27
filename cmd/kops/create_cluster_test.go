package main

import (
	"testing"
)

func TestParseCloudLabels(t *testing.T) {
	expect := map[string]string{"foo":"bar", "fib":"baz"}
	checkParse(t, "", map[string]string{}, false)
	checkParse(t, "foo=bar,fib=baz", expect, false)
	checkParse(t, `foo=bar,"fib"="baz"`, expect, false)
	checkParse(t, `"fo\""o"=bar,"fi\b"="baz"`,
		map[string]string{`fo\"o`:"bar", `fi\b`:"baz"}, false)
	checkParse(t, `fo"o=bar,fib=baz`, expect, true)
	checkParse(t, `fo,o=bar,fib=baz`, expect, true)
}

func checkParse(t *testing.T, s string, expect map[string]string, shouldErr bool) {
	m, err := parseCloudLabels(s)
	if err != nil {
		if shouldErr {
			return
		} else {
			t.Errorf(err.Error())
		}
	}

	for k, v := range expect {
		if m[k] != v {
			t.Errorf("Expected: %v, Got: %v", expect, m)
		}
	}
}
