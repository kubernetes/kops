/*
Copyright 2016 The Kubernetes Authors.

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

package instancegroups

import (
	"fmt"
	"testing"
	"time"
)

func TestGetPrefix(t *testing.T) {
	timeStamp := time.Now()
	timeStampLayout := timeStamp.Format(ig_ts_layout)
	grid := []struct {
		Input  string
		Output string
	}{
		{
			Input:  "foo",
			Output: fmt.Sprintf("foo%s%s", ig_prefix, timeStampLayout),
		},
		{
			Input:  "node01-rolling-update-2017-01-02-15-04-05",
			Output: fmt.Sprintf("node01%s%s", ig_prefix, timeStampLayout),
		},
	}

	for _, g := range grid {
		result := getSuffixWithTime(g.Input, timeStamp)
		if result != g.Output {
			t.Fatalf("testing %q failed, expected %q, got %q", g.Input, g.Output, result)
		}
	}
}
