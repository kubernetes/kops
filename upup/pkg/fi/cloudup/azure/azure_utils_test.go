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

package azure

import (
	"fmt"
	"testing"
)

func TestZoneToLocation(t *testing.T) {
	testCases := []struct {
		zone     string
		success  bool
		location string
	}{
		{
			zone:     "eastus-1",
			success:  true,
			location: "eastus",
		},
		{
			zone:    "eastus",
			success: false,
		},
		{
			zone:    "eastus-1-2",
			success: false,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			loc, err := ZoneToLocation(tc.zone)
			if !tc.success {
				if err == nil {
					t.Fatalf("unexpected success")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if loc != tc.location {
				t.Errorf("expected %s but got %s", tc.location, loc)
			}
		})
	}
}
