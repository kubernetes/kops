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

package channels

import (
	"testing"

	"github.com/blang/semver"
	"k8s.io/kops/channels/pkg/api"
)

func Test_Filtering(t *testing.T) {
	grid := []struct {
		Input             api.AddonSpec
		KubernetesVersion string
		Expected          bool
	}{
		{
			Input: api.AddonSpec{
				KubernetesVersion: ">=1.6.0",
			},
			KubernetesVersion: "1.6.0",
			Expected:          true,
		},
		{
			Input: api.AddonSpec{
				KubernetesVersion: "<1.6.0",
			},
			KubernetesVersion: "1.6.0",
			Expected:          false,
		},
		{
			Input: api.AddonSpec{
				KubernetesVersion: ">=1.6.0",
			},
			KubernetesVersion: "1.5.9",
			Expected:          false,
		},
		{
			Input: api.AddonSpec{
				KubernetesVersion: ">=1.4.0 <1.6.0",
			},
			KubernetesVersion: "1.5.9",
			Expected:          true,
		},
		{
			Input: api.AddonSpec{
				KubernetesVersion: ">=1.4.0 <1.6.0",
			},
			KubernetesVersion: "1.6.0",
			Expected:          false,
		},
	}
	for _, g := range grid {
		k8sVersion := semver.MustParse(g.KubernetesVersion)
		addon := &Addon{
			Spec: &g.Input,
		}
		actual := addon.matches(k8sVersion)
		if actual != g.Expected {
			t.Errorf("unexpected result from %v, %s.  got %v", g.Input.KubernetesVersion, g.KubernetesVersion, actual)
		}
	}
}

func Test_Replacement(t *testing.T) {
	grid := []struct {
		Old      *ChannelVersion
		New      *ChannelVersion
		Replaces bool
	}{
		// With no id, update if and only if newer semver
		{
			Old:      &ChannelVersion{Version: s("1.0.0"), Id: "", ManifestHash: ""},
			New:      &ChannelVersion{Version: s("1.0.0"), Id: "", ManifestHash: ""},
			Replaces: false,
		},
		{
			Old:      &ChannelVersion{Version: s("1.0.0"), Id: "", ManifestHash: ""},
			New:      &ChannelVersion{Version: s("1.0.1"), Id: "", ManifestHash: ""},
			Replaces: true,
		},
		{
			Old:      &ChannelVersion{Version: s("1.0.1"), Id: "", ManifestHash: ""},
			New:      &ChannelVersion{Version: s("1.0.0"), Id: "", ManifestHash: ""},
			Replaces: false,
		},
		{
			Old:      &ChannelVersion{Version: s("1.1.0"), Id: "", ManifestHash: ""},
			New:      &ChannelVersion{Version: s("1.1.1"), Id: "", ManifestHash: ""},
			Replaces: true,
		},
		{
			Old:      &ChannelVersion{Version: s("1.1.1"), Id: "", ManifestHash: ""},
			New:      &ChannelVersion{Version: s("1.1.0"), Id: "", ManifestHash: ""},
			Replaces: false,
		},

		// With id, update if different id and same version, otherwise follow semver
		{
			Old:      &ChannelVersion{Version: s("1.0.0"), Id: "a", ManifestHash: ""},
			New:      &ChannelVersion{Version: s("1.0.0"), Id: "a", ManifestHash: ""},
			Replaces: false,
		},
		{
			Old:      &ChannelVersion{Version: s("1.0.0"), Id: "a", ManifestHash: ""},
			New:      &ChannelVersion{Version: s("1.0.0"), Id: "b", ManifestHash: ""},
			Replaces: true,
		},
		{
			Old:      &ChannelVersion{Version: s("1.0.0"), Id: "b", ManifestHash: ""},
			New:      &ChannelVersion{Version: s("1.0.0"), Id: "a", ManifestHash: ""},
			Replaces: true,
		},
		{
			Old:      &ChannelVersion{Version: s("1.0.0"), Id: "a", ManifestHash: ""},
			New:      &ChannelVersion{Version: s("1.0.1"), Id: "a", ManifestHash: ""},
			Replaces: true,
		},
		{
			Old:      &ChannelVersion{Version: s("1.0.0"), Id: "a", ManifestHash: ""},
			New:      &ChannelVersion{Version: s("1.0.1"), Id: "a", ManifestHash: ""},
			Replaces: true,
		},
		{
			Old:      &ChannelVersion{Version: s("1.0.0"), Id: "a", ManifestHash: ""},
			New:      &ChannelVersion{Version: s("1.0.1"), Id: "a", ManifestHash: ""},
			Replaces: true,
		},
		//Test ManifestHash Changes
		{
			Old:      &ChannelVersion{Version: s("1.0.0"), Id: "a", ManifestHash: "3544de6578b2b582c0323b15b7b05a28c60b9430"},
			New:      &ChannelVersion{Version: s("1.0.0"), Id: "a", ManifestHash: "3544de6578b2b582c0323b15b7b05a28c60b9430"},
			Replaces: false,
		},
		{
			Old:      &ChannelVersion{Version: s("1.0.0"), Id: "a", ManifestHash: ""},
			New:      &ChannelVersion{Version: s("1.0.0"), Id: "a", ManifestHash: "3544de6578b2b582c0323b15b7b05a28c60b9430"},
			Replaces: true,
		},
		{
			Old:      &ChannelVersion{Version: s("1.0.0"), Id: "a", ManifestHash: "3544de6578b2b582c0323b15b7b05a28c60b9430"},
			New:      &ChannelVersion{Version: s("1.0.0"), Id: "a", ManifestHash: "ea9e79bf29adda450446487d65a8fc6b3fdf8c2b"},
			Replaces: true,
		},
	}
	for _, g := range grid {
		actual := g.New.replaces(g.Old)
		if actual != g.Replaces {
			t.Errorf("unexpected result from %v -> %v, expect %t.  actual %v", g.Old, g.New, g.Replaces, actual)
		}
	}
}

func s(v string) *string {
	return &v
}
