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

package validation

import (
	"net"
	"testing"
)

func Test_isSubnet(t *testing.T) {
	grid := []struct {
		L        string
		R        string
		IsSubnet bool
	}{
		{
			L:        "192.168.1.0/24",
			R:        "192.168.0.0/24",
			IsSubnet: false,
		},
		{
			L:        "192.168.0.0/16",
			R:        "192.168.0.0/24",
			IsSubnet: true,
		},
		{
			L:        "192.168.0.0/24",
			R:        "192.168.0.0/16",
			IsSubnet: false,
		},
		{
			L:        "192.168.0.0/16",
			R:        "192.168.0.0/16",
			IsSubnet: true, // Not a strict subnet
		},
		{
			L:        "192.168.0.1/16",
			R:        "192.168.0.0/24",
			IsSubnet: true,
		},
		{
			L:        "0.0.0.0/0",
			R:        "101.0.1.0/32",
			IsSubnet: true,
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
		actual := isSubnet(l, r)
		if actual != g.IsSubnet {
			t.Errorf("isSubnet(%q, %q) = %v, expected %v", g.L, g.R, actual, g.IsSubnet)
		}
	}
}
