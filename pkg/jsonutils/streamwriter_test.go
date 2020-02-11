/*
Copyright 2020 The Kubernetes Authors.

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

package jsonutils

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc"
)

func TestWriteToken(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		indent   string
		expected string
	}{
		{
			name:   "WriteJson",
			input:  heredoc.Doc(`{"key1": 123,"key2" : true   }  `),
			indent: "",
			expected: heredoc.Doc(`{
			  "key1": 123,
			  "key2": true
			}`),
		},
		{
			name:   "WriteJsonWithIndent",
			input:  heredoc.Doc(`{"key1" :123, "999" : [123.5678,"abc"]  }`),
			indent: "###",
			expected: heredoc.Doc(`###{
			###  "key1": 123,
			###  "999": [
			###    123.5678,
			###    "abc"
			###  ]
			###}`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			out := &JSONStreamWriter{
				out:    &buf,
				indent: test.indent,
			}
			in := json.NewDecoder(strings.NewReader(test.input))
			for {
				token, err := in.Token()
				if err != nil {
					if err == io.EOF { // No more token
						break
					} else {
						t.Fatalf("unexpected error parsing json input: %v", err)
					}
				}
				if err := out.WriteToken(token); err != nil {
					t.Fatalf("error writing json: %v", err)
				}
			}
			if buf.String() != test.expected {
				t.Fatalf("Actual result:\n%s\nExpect:\n%s", buf.String(), test.expected)
			}
		})
	}
}
