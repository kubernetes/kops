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
	"github.com/blang/semver/v4"
	channelsapi "k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kubemanifest"
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

	// if b.Cluster.Spec.KubeDNS != nil && b.Cluster.Spec.KubeDNS.Provider == "CoreDNS" {
	// 	// TODO: Autopopulate a CoreDNS operator if we don't have one?
	// 	// TODO: Check that we haven't manually created a CoreDNS CRD

	// 	operatorKey := "operator.coredns.addons.x-k8s.io"

	// 	coreDNSVersion := lookupVersion("coredns", b.Cluster.Spec.Channel)
	// 	operatorVersion := lookupVersion(operatorKey, b.Cluster.Spec.Channel)

	// 	//key := "coredns.addons.x-k8s.io"
	// 	//id := ""

	// 	location := path.Join("packages", operatorKey, operatorVersion, "manifest.yaml")
	// 	channelURL, err := kops.ResolveChannel(b.Cluster.Spec.Channel)
	// 	if err != nil {
	// 		return nil, nil, fmt.Errorf("error resolving channel %q: %v", b.Cluster.Spec.Channel, err)
	// 	}

	// 	locationURL := channelURL.ResolveReference(&url.URL{Path: location}).String()

	// 	manifestBytes, err := vfs.Context.ReadFile(locationURL)
	// 	if err != nil {
	// 		return nil, nil, fmt.Errorf("error reading operator manifest %q: %v", locationURL, err)
	// 	}

	// 	addon := &WellKnownAddon{
	// 		Manifest: manifestBytes,
	// 		Spec: channelsapi.AddonSpec{
	// 			Name:     fi.String(operatorKey),
	// 			Selector: map[string]string{"k8s-addon": operatorKey},
	// 			Manifest: fi.String(location),
	// 		},
	// 	}
	// 	addons = append(addons, addon)

	// 	{
	// 		metadata := map[string]interface{}{
	// 			"namespace": "kube-system",
	// 			"name":      "coredns",
	// 		}
	// 		spec := map[string]interface{}{
	// 			"dnsDomain": b.Cluster.Spec.KubeDNS.Domain,
	// 			"dnsIP":     b.Cluster.Spec.KubeDNS.ServerIP,

	// 			//"channel": coreDNSChannel,
	// 			"version": coreDNSVersion,
	// 		}

	// 		crd := kubemanifest.NewObject(map[string]interface{}{
	// 			"apiVersion": "addons.x-k8s.io/v1alpha1",
	// 			"kind":       "CoreDNS",
	// 			"metadata":   metadata,
	// 			"spec":       spec,
	// 		})
	// 		crds = append(crds, crd)
	// 	}
	// }

	return addons, crds, nil
}

func CreateAddons(channel *kops.Channel, kubernetesVersion *semver.Version) (kubemanifest.ObjectList, error) {
	var addons kubemanifest.ObjectList

	if !featureflag.UseAddonOperators.Enabled() {
		return addons, nil
	}

	{
		operatorKey := "operator.coredns.addons.x-k8s.io"

		operatorVersion, err := channel.GetPackageVersion(operatorKey)
		if err != nil {
			return nil, err
		}

		metadata := map[string]interface{}{
			"name": operatorKey,
		}
		spec := map[string]interface{}{
			"version": operatorVersion.String(),
		}

		addonPackage := kubemanifest.NewObject(map[string]interface{}{
			"apiVersion": "kops.io/v1alpha1",
			"kind":       "ClusterPackage",
			"metadata":   metadata,
			"spec":       spec,
		})
		addons = append(addons, addonPackage)
	}

	{
		key := "coredns"
		version, err := channel.GetPackageVersion(key)
		if err != nil {
			return nil, err
		}
		metadata := map[string]interface{}{
			"namespace": "kube-system",
			"name":      "coredns",
		}
		spec := map[string]interface{}{
			// "dnsDomain": b.Cluster.Spec.KubeDNS.Domain,
			// "dnsIP":     b.Cluster.Spec.KubeDNS.ServerIP,

			"version": version.String(),
		}

		crd := kubemanifest.NewObject(map[string]interface{}{
			"apiVersion": "addons.x-k8s.io/v1alpha1",
			"kind":       "CoreDNS",
			"metadata":   metadata,
			"spec":       spec,
		})
		addons = append(addons, crd)
	}
	return addons, nil
}
