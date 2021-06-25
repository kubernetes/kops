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

package wellknownoperators

import (
	"fmt"
	"net/url"
	"path"

	channelsapi "k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

type WellKnownAddon struct {
	Manifest []byte
	Spec     channelsapi.AddonSpec
}

type Builder struct {
	Cluster *kops.Cluster
}

func (b *Builder) Build() ([]*WellKnownAddon, kubemanifest.ObjectList, error) {
	if !featureflag.UseAddonOperators.Enabled() {
		return nil, nil, nil
	}

	var addons []*WellKnownAddon
	var crds kubemanifest.ObjectList

	if b.Cluster.Spec.KubeDNS != nil && b.Cluster.Spec.KubeDNS.Provider == "CoreDNS" {
		// TODO: Check that we haven't manually loaded a CoreDNS operator
		// TODO: Check that we haven't manually created a CoreDNS CRD

		key := "coredns.addons.x-k8s.io"
		version := "0.1.0-kops.1"
		id := ""

		location := path.Join("operators", key, version, "manifest.yaml")
		channelURL, err := kops.ResolveChannel(b.Cluster.Spec.Channel)
		if err != nil {
			return nil, nil, fmt.Errorf("error resolving channel %q: %v", b.Cluster.Spec.Channel, err)
		}

		locationURL := channelURL.ResolveReference(&url.URL{Path: location}).String()

		manifestBytes, err := vfs.Context.ReadFile(locationURL)
		if err != nil {
			return nil, nil, fmt.Errorf("error reading operator manifest %q: %v", locationURL, err)
		}

		addon := &WellKnownAddon{
			Manifest: manifestBytes,
			Spec: channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			},
		}
		addons = append(addons, addon)

		{
			metadata := map[string]interface{}{
				"namespace": "kube-system",
				"name":      "coredns",
			}
			spec := map[string]interface{}{
				"dnsDomain": b.Cluster.Spec.KubeDNS.Domain,
				"dnsIP":     b.Cluster.Spec.KubeDNS.ServerIP,
			}

			crd := kubemanifest.NewObject(map[string]interface{}{
				"apiVersion": "addons.x-k8s.io/v1alpha1",
				"kind":       "CoreDNS",
				"metadata":   metadata,
				"spec":       spec,
			})
			crds = append(crds, crd)
		}
	}

	return addons, crds, nil
}
