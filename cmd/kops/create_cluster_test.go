/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"testing"
)

func TestParseCloudLabels(t *testing.T) {
	expect := map[string]string{"foo": "bar", "fib": "baz"}
	checkParse(t, "", map[string]string{}, false)
	checkParse(t, "foo=bar,fib=baz", expect, false)
	checkParse(t, `foo=bar,"fib"="baz"`, expect, false)
	checkParse(t, `"fo\""o"=bar,"fi\b"="baz"`,
		map[string]string{`fo\"o`: "bar", `fi\b`: "baz"}, false)
	checkParse(t, `fo"o=bar,fib=baz`, expect, true)
	checkParse(t, `fo,o=bar,fib=baz`, expect, true)
}

func checkParse(t *testing.T, s string, expect map[string]string, shouldErr bool) {
	m, err := parseCloudLabels(s)
	if err != nil {
		if shouldErr {
			return
		}
		t.Errorf(err.Error())
	}

	for k, v := range expect {
		if m[k] != v {
			t.Errorf("Expected: %v, Got: %v", expect, m)
		}
	}
}
