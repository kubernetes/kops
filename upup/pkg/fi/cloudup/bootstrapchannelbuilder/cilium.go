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
	"github.com/blang/semver/v4"
	"k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/upup/pkg/fi"
)

func addCiliumAddon(b *BootstrapChannelBuilder, addons *api.Addons) {

	cilium := b.Cluster.Spec.Networking.Cilium
	if cilium != nil {
		ver, _ := semver.ParseTolerant(cilium.Version)
		key := "networking.cilium.io"
		if ver.Minor < 8 {
			version := "1.7.3-kops.1"

			{
				id := "k8s-1.12"
				location := key + "/" + id + ".yaml"

				addons.Spec.Addons = append(addons.Spec.Addons, &api.AddonSpec{
					Name:     fi.String(key),
					Version:  fi.String(version),
					Selector: networkingSelector(),
					Manifest: fi.String(location),
					Id:       id,
				})
			}
		} else if ver.Minor == 8 {
			version := "1.8.0-kops.1"
			{
				id := "k8s-1.12"
				location := key + "/" + id + "-v1.8.yaml"

				addons.Spec.Addons = append(addons.Spec.Addons, &api.AddonSpec{
					Name:               fi.String(key),
					Version:            fi.String(version),
					Selector:           networkingSelector(),
					Manifest:           fi.String(location),
					Id:                 id,
					NeedsRollingUpdate: "all",
				})
			}
		} else if ver.Minor == 9 {
			version := "1.9.0-kops.1"
			{
				id := "k8s-1.12"
				location := key + "/" + id + "-v1.9.yaml"

				addons.Spec.Addons = append(addons.Spec.Addons, &api.AddonSpec{
					Name:               fi.String(key),
					Version:            fi.String(version),
					Selector:           networkingSelector(),
					Manifest:           fi.String(location),
					Id:                 id,
					NeedsRollingUpdate: "all",
				})
			}
		}
	}

}
