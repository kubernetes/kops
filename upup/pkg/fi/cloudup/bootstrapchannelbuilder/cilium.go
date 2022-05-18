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

package bootstrapchannelbuilder

import (
	"fmt"

	"github.com/blang/semver/v4"

	"k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/upup/pkg/fi"
)

func addCiliumAddon(b *BootstrapChannelBuilder, addons *AddonList) error {
	cilium := b.Cluster.Spec.Networking.Cilium
	if cilium != nil {
		ver, err := semver.ParseTolerant(cilium.Version)
		if err != nil {
			return fmt.Errorf("Failed to parse cilium version: %w", err)
		}

		key := "networking.cilium.io"
		if ver.Minor < 9 {
			{
				id := "k8s-1.12"
				location := key + "/" + id + "-v1.8.yaml"

				addons.Add(&api.AddonSpec{
					Name:               fi.String(key),
					Selector:           networkingSelector(),
					Manifest:           fi.String(location),
					Id:                 id,
					NeedsRollingUpdate: "all",
				})
			}
		} else if ver.Minor == 9 {
			{
				id := "k8s-1.12"
				location := key + "/" + id + "-v1.9.yaml"

				addon := &api.AddonSpec{
					Name:               fi.String(key),
					Selector:           networkingSelector(),
					Manifest:           fi.String(location),
					Id:                 id,
					NeedsRollingUpdate: "all",
				}
				if cilium.Hubble != nil && fi.BoolValue(cilium.Hubble.Enabled) {
					addon.NeedsPKI = true
				}
				addons.Add(addon)
			}
		} else if ver.Minor == 10 || (ver.Minor == 11 && ver.Patch < 5) {
			{
				id := "k8s-1.16"
				location := key + "/" + id + "-v1.10.yaml"

				addon := &api.AddonSpec{
					Name:               fi.String(key),
					Selector:           networkingSelector(),
					Manifest:           fi.String(location),
					Id:                 id,
					NeedsRollingUpdate: "all",
				}
				if cilium.Hubble != nil && fi.BoolValue(cilium.Hubble.Enabled) {
					addon.NeedsPKI = true
				}
				addons.Add(addon)
			}
		} else if ver.Minor == 11 {
			{
				id := "k8s-1.16"
				location := key + "/" + id + "-v1.11.yaml"

				addon := &api.AddonSpec{
					Name:               fi.String(key),
					Selector:           networkingSelector(),
					Manifest:           fi.String(location),
					Id:                 id,
					NeedsRollingUpdate: "all",
				}
				if cilium.Hubble != nil && fi.BoolValue(cilium.Hubble.Enabled) {
					addon.NeedsPKI = true
				}
				addons.Add(addon)
			}
		} else {
			return fmt.Errorf("unknown cilium version: %q", cilium.Version)
		}
	}
	return nil
}
