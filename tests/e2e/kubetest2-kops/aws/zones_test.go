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

package aws

import (
	"fmt"
	"testing"
)

func TestRandomZones(t *testing.T) {
	t.Parallel() // marks TLog as capable of running in parallel with other tests
	tests := []struct {
		count int
		err   error
	}{
		{1, nil},
		{2, nil},
		{3, nil},
		{4, ErrNoEligibleRegion},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v zones", tt.count), func(t *testing.T) {
			zones, err := RandomZones(tt.count)
			if err != tt.err {
				t.Errorf("unexpected error response: %v vs %v. zones: %v", err, tt.err, zones)
				t.Fail()
			} else if tt.err == nil && len(zones) != tt.count {
				t.Errorf("Unexpected number of zones returned: %v vs %v. zones: %v", len(zones), tt.count, zones)
				t.Fail()
			}
		})
	}
}
