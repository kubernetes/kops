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

package stringorslice

import (
	"encoding/json"
	"testing"

	"k8s.io/klog"
)

func TestRoundTrip(t *testing.T) {
	grid := []struct {
		Value StringOrSlice
		JSON  string
	}{
		{
			Value: String("a"),
			JSON:  "\"a\"",
		},
		{
			Value: Of("a"),
			JSON:  "\"a\"",
		},
		{
			Value: Slice([]string{"a"}),
			JSON:  "[\"a\"]",
		},
		{
			Value: Of("a", "b"),
			JSON:  "[\"a\",\"b\"]",
		},
		{
			Value: Slice([]string{"a", "b"}),
			JSON:  "[\"a\",\"b\"]",
		},
		{
			Value: Of(),
			JSON:  "[]",
		},
		{
			Value: Slice(nil),
			JSON:  "[]",
		},
	}
	for _, g := range grid {
		actualJson, err := json.Marshal(g.Value)
		if err != nil {
			t.Errorf("error encoding StringOrSlice %s to json: %v", g.Value, err)
		}

		klog.V(8).Infof("marshalled %s -> %q", g.Value, actualJson)

		if g.JSON != string(actualJson) {
			t.Errorf("Unexpected JSON encoding.  Actual=%q, Expected=%q", string(actualJson), g.JSON)
		}

		parsed := &StringOrSlice{}
		err = json.Unmarshal([]byte(g.JSON), parsed)
		if err != nil {
			t.Errorf("error decoding StringOrSlice %s to json: %v", g.JSON, err)
		}

		if !parsed.Equal(g.Value) {
			t.Errorf("Unexpected JSON decoded value for %q.  Actual=%v, Expected=%v", g.JSON, parsed, g.Value)
		}

	}
}
