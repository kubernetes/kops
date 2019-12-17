/*
Copyright 2019 The Kubernetes Authors.

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

package text

import (
	"testing"
)

func TestSplitToSections(t *testing.T) {
	tests := []struct {
		content     []byte
		numSections int
	}{
		{
			content:     []byte(""),
			numSections: 1,
		},
		{
			content:     []byte("section 1"),
			numSections: 1,
		},
		{
			content:     []byte("section 1\n"),
			numSections: 1,
		},
		{
			content:     []byte("section 1\r\n"),
			numSections: 1,
		},
		{
			content:     []byte("section 1\nanother line\n"),
			numSections: 1,
		},
		{
			content:     []byte("section 1\r\nanother line\r\n"),
			numSections: 1,
		},
		{
			content:     []byte("section 1\n---\nsection 2"),
			numSections: 2,
		},
		{
			content:     []byte("section 1\r\n---\r\nsection 2"),
			numSections: 2,
		},
		{
			content:     []byte("section 1\n\n---\n\nsection 2"),
			numSections: 2,
		},
		{
			content:     []byte("section 1\r\n\r\n---\r\n\r\nsection 2"),
			numSections: 2,
		},
	}
	for _, test := range tests {
		ns := len(SplitContentToSections(test.content))
		if ns != test.numSections {
			t.Errorf("Expected %v, got %v sections for content %q", test.numSections, ns, string(test.content))
		}
	}
}
