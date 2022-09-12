/*
Copyright 2021 The Kubernetes Authors.

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

package truncate

import (
	"fmt"
	"testing"
)

func TestTruncateString(t *testing.T) {
	grid := []struct {
		Input         string
		Expected      string
		MaxLength     int
		AlwaysAddHash bool
	}{
		{
			Input:     "foo",
			Expected:  "foo",
			MaxLength: 64,
		},
		{
			Input:     "this_string_is_33_characters_long",
			Expected:  "this_string_is_33_characters_long",
			MaxLength: 64,
		},
		{
			Input:         "this_string_is_33_characters_long",
			Expected:      "this_string_is_33_characters_long-t4mk8d",
			MaxLength:     64,
			AlwaysAddHash: true,
		},
		{
			Input:     "this_string_is_longer_it_is_46_characters_long",
			Expected:  "this_string_is_longer_it_-ha2gug",
			MaxLength: 32,
		},
		{
			Input:         "this_string_is_longer_it_is_46_characters_long",
			Expected:      "this_string_is_longer_it_-ha2gug",
			MaxLength:     32,
			AlwaysAddHash: true,
		},
		{
			Input:     "this_string_is_even_longer_due_to_extreme_verbosity_it_is_in_fact_84_characters_long",
			Expected:  "this_string_is_even_longer_due_to_extreme_verbosity_it_is-7mc0g6",
			MaxLength: 64,
		},
	}

	for _, g := range grid {
		t.Run(fmt.Sprintf("input:%s/maxLength:%d/alwaysAddHash:%v", g.Input, g.MaxLength, g.AlwaysAddHash), func(t *testing.T) {
			opt := TruncateStringOptions{MaxLength: g.MaxLength, AlwaysAddHash: g.AlwaysAddHash}
			actual := TruncateString(g.Input, opt)
			if actual != g.Expected {
				t.Errorf("TruncateString(%q, %+v) => %q, expected %q", g.Input, opt, actual, g.Expected)
			}
		})
	}
}
