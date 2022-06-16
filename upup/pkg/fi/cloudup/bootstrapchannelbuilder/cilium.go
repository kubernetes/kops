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
	"k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/upup/pkg/fi"
)

func addCiliumAddon(b *BootstrapChannelBuilder, addons *AddonList) error {
	cilium := b.Cluster.Spec.Networking.Cilium
	if cilium != nil {
		key := "networking.cilium.io"

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
	}
	return nil
}
