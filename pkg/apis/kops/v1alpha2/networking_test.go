/*
Copyright 2026 The Kubernetes Authors.

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

package v1alpha2

import (
	"testing"
)

// TestNetworkingIsEmpty checks that networking options whose support has been
// removed are still recognized by IsEmpty. Otherwise, converting a cluster
// using such an option from the internal API back to v1alpha2 would silently
// drop the networking field instead of round-tripping it.
func TestNetworkingIsEmpty(t *testing.T) {
	grid := []struct {
		name       string
		networking NetworkingSpec
	}{
		{name: "kopeio", networking: NetworkingSpec{Kopeio: &KopeioNetworkingSpec{}}},
		{name: "canal", networking: NetworkingSpec{Canal: &CanalNetworkingSpec{}}},
		{name: "weave", networking: NetworkingSpec{Weave: &WeaveNetworkingSpec{}}},
		{name: "romana", networking: NetworkingSpec{Romana: &RomanaNetworkingSpec{}}},
		{name: "lyftvpc", networking: NetworkingSpec{LyftVPC: &LyftVPCNetworkingSpec{}}},
		{name: "classic", networking: NetworkingSpec{Classic: &ClassicNetworkingSpec{}}},
	}

	if !(&NetworkingSpec{}).IsEmpty() {
		t.Errorf("expected empty NetworkingSpec to be considered empty")
	}

	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			if g.networking.IsEmpty() {
				t.Errorf("expected NetworkingSpec with %s set to not be considered empty", g.name)
			}
		})
	}
}
