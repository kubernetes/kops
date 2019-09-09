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

package backoff

import (
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	expected := []time.Duration{
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		16 * time.Second,
		32 * time.Second,
		64 * time.Second,
		128 * time.Second,
		256 * time.Second,
		5 * time.Minute,
		5 * time.Minute,
		5 * time.Minute,
		5 * time.Minute,
	}

	for i := range expected {
		actual := computeBackoff()
		if actual != expected[i] {
			t.Fatalf("unexpected backoff @%d: %v", i, actual)
		}
	}
}
