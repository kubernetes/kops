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

package addonmanifests

import (
	"fmt"

	"k8s.io/klog/v2"
	addonsapi "k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/components/addonmanifests/dnscontroller"
	"k8s.io/kops/upup/pkg/fi"
)

func RemapAddonManifest(addon *addonsapi.AddonSpec, context *model.KopsModelContext, assetBuilder *assets.AssetBuilder, manifest []byte) ([]byte, error) {
	name := fi.StringValue(addon.Name)

	{
		objects, err := kubemanifest.LoadObjectsFrom(manifest)
		if err != nil {
			return nil, err
		}

		remapped := false
		if name == "dns-controller.addons.k8s.io" {
			if err := dnscontroller.Remap(context, addon, objects); err != nil {
				return nil, err
			}
			remapped = true
		}

		if remapped {
			b, err := objects.ToYAML()
			if err != nil {
				return nil, err
			}
			manifest = b
		}
	}

	{
		remapped, err := assetBuilder.RemapManifest(manifest)
		if err != nil {
			klog.Infof("invalid manifest: %s", string(manifest))
			return nil, fmt.Errorf("error remapping manifest %s: %v", manifest, err)
		}
		manifest = remapped
	}

	return manifest, nil
}
