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

package subnet

import (
	"net"
	"reflect"
	"testing"
)

func Test_BelongsTo(t *testing.T) {
	grid := []struct {
		L       string
		R       string
		Belongs bool
	}{
		{
			L:       "192.168.1.0/24",
			R:       "192.168.0.0/24",
			Belongs: false,
		},
		{
			L:       "192.168.0.0/16",
			R:       "192.168.0.0/24",
			Belongs: true,
		},
		{
			L:       "192.168.0.0/24",
			R:       "192.168.0.0/16",
			Belongs: false,
		},
		{
			L:       "192.168.0.0/16",
			R:       "192.168.0.0/16",
			Belongs: true, // Not a strict subnet
		},
		{
			L:       "192.168.0.1/16",
			R:       "192.168.0.0/24",
			Belongs: true,
		},
		{
			L:       "0.0.0.0/0",
			R:       "101.0.1.0/32",
			Belongs: true,
		},
	}
	for _, g := range grid {
		_, l, err := net.ParseCIDR(g.L)
		if err != nil {
			t.Fatalf("error parsing %q: %v", g.L, err)
		}
		_, r, err := net.ParseCIDR(g.R)
		if err != nil {
			t.Fatalf("error parsing %q: %v", g.R, err)
		}
		actual := BelongsTo(l, r)
		if actual != g.Belongs {
			t.Errorf("isSubnet(%q, %q) = %v, expected %v", g.L, g.R, actual, g.Belongs)
		}
	}
}

func Test_SplitInto8(t *testing.T) {
	tests := []struct {
		parent   string
		expected []string
	}{
		{
			parent:   "1.2.3.0/24",
			expected: []string{"1.2.3.0/27", "1.2.3.32/27", "1.2.3.64/27", "1.2.3.96/27", "1.2.3.128/27", "1.2.3.160/27", "1.2.3.192/27", "1.2.3.224/27"},
		},
		{
			parent:   "1.2.3.0/27",
			expected: []string{"1.2.3.0/30", "1.2.3.4/30", "1.2.3.8/30", "1.2.3.12/30", "1.2.3.16/30", "1.2.3.20/30", "1.2.3.24/30", "1.2.3.28/30"},
		},
	}
	for _, test := range tests {
		_, parent, err := net.ParseCIDR(test.parent)
		if err != nil {
			t.Fatalf("error parsing parent cidr %q: %v", test.parent, err)
		}

		subnets, err := SplitInto8(parent)
		if err != nil {
			t.Fatalf("error splitting parent cidr %q: %v", parent, err)
		}

		var actual []string
		for _, subnet := range subnets {
			actual = append(actual, subnet.String())
		}
		if !reflect.DeepEqual(actual, test.expected) {
			t.Fatalf("unexpected result of split: actual=%v, expected=%v", actual, test.expected)
		}
	}
}
