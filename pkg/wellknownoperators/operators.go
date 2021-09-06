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

	"github.com/blang/semver/v4"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	channelsapi "k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

type Package struct {
	Manifest []byte
	Spec     channelsapi.AddonSpec
}

type Builder struct {
	Cluster *kops.Cluster
}

func (b *Builder) Build(objects kubemanifest.ObjectList) ([]*Package, kubemanifest.ObjectList, error) {
	if !featureflag.UseAddonOperators.Enabled() {
		return nil, objects, nil
	}

	var packages []*Package
	var keepObjects kubemanifest.ObjectList

	for _, object := range objects {
		keep := true

		// We may in future move ClusterPackage to the server-side,
		// however remapping it client-side allow us to use our existing manifest logic
		// including image remapping.
		if object.Kind() == "ClusterPackage" {
			u := object.ToUnstructured()
			if u.GroupVersionKind().Group == "addons.x-k8s.io" {
				pkg, err := b.loadClusterPackage(u)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to load package: %w", err)
				}
				packages = append(packages, pkg)
				keep = false
			}
		}

		if keep {
			keepObjects = append(keepObjects, object)
		}
	}
	return packages, keepObjects, nil
}

func (b *Builder) loadClusterPackage(u *unstructured.Unstructured) (*Package, error) {
	operatorKey := u.GetName()

	operatorVersion, _, err := unstructured.NestedString(u.Object, "spec", "version")
	if err != nil || operatorVersion == "" {
		return nil, fmt.Errorf("could not get spec.version from ClusterPackage")
	}

	location := path.Join("packages", operatorKey, operatorVersion, "manifest.yaml")
	channelURL, err := kops.ResolveChannel(b.Cluster.Spec.Channel)
	if err != nil {
		return nil, fmt.Errorf("error resolving channel %q: %v", b.Cluster.Spec.Channel, err)
	}

	locationURL := channelURL.ResolveReference(&url.URL{Path: location}).String()

	manifestBytes, err := vfs.Context.ReadFile(locationURL)
	if err != nil {
		return nil, fmt.Errorf("error reading operator manifest %q: %v", locationURL, err)
	}

	addon := &Package{
		Manifest: manifestBytes,
		Spec: channelsapi.AddonSpec{
			Name:     fi.String(operatorKey),
			Selector: map[string]string{"k8s-addon": operatorKey},
			Manifest: fi.String(location),
		},
	}
	return addon, nil
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
			"apiVersion": "addons.x-k8s.io/v1alpha1",
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
