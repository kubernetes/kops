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

package cloudup

import (
	"fmt"

	channelsapi "k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/templates"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/upup/pkg/fi/utils"
)

type BootstrapChannelBuilder struct {
	cluster      *kops.Cluster
	Lifecycle    *fi.Lifecycle
	templates    *templates.Templates
	assetBuilder *assets.AssetBuilder
}

var _ fi.ModelBuilder = &BootstrapChannelBuilder{}

func (b *BootstrapChannelBuilder) Build(c *fi.ModelBuilderContext) error {

	addons, manifests, err := b.buildManifest()
	if err != nil {
		return err
	}

	addonsYAML, err := utils.YamlMarshal(addons)
	if err != nil {
		return fmt.Errorf("error serializing addons yaml: %v", err)
	}

	name := b.cluster.ObjectMeta.Name + "-addons-bootstrap"

	tasks := c.Tasks

	tasks[name] = &fitasks.ManagedFile{
		Name:      fi.String(name),
		Lifecycle: b.Lifecycle,

		Location: fi.String("addons/bootstrap-channel.yaml"),
		Contents: fi.WrapResource(fi.NewBytesResource(addonsYAML)),
	}

	for key, manifest := range manifests {
		name := b.cluster.ObjectMeta.Name + "-addons-" + key

		manifestResource := b.templates.Find(manifest)
		if manifestResource == nil {
			return fmt.Errorf("unable to find manifest %s", manifest)
		}

		manifestBytes, err := fi.ResourceAsBytes(manifestResource)
		if err != nil {
			return fmt.Errorf("error reading manifest %s: %v", manifest, err)
		}

		manifestBytes, err = b.assetBuilder.RemapManifest(manifestBytes)
		if err != nil {
			return fmt.Errorf("error remapping manifest %s: %v", manifest, err)
		}

		tasks[name] = &fitasks.ManagedFile{
			Name:      fi.String(name),
			Lifecycle: b.Lifecycle,

			Location: fi.String(manifest),
			Contents: fi.WrapResource(fi.NewBytesResource(manifestBytes)),
		}
	}

	return nil
}

func (b *BootstrapChannelBuilder) buildManifest() (*channelsapi.Addons, map[string]string, error) {
	manifests := make(map[string]string)

	addons := &channelsapi.Addons{}
	addons.Kind = "Addons"
	addons.ObjectMeta.Name = "bootstrap"

	{
		key := "core.addons.k8s.io"
		version := "1.4.0"

		location := key + "/v" + version + ".yaml"

		addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
			Name:     fi.String(key),
			Version:  fi.String(version),
			Selector: map[string]string{"k8s-addon": key},
			Manifest: fi.String(location),
		})
		manifests[key] = "addons/" + location
	}

	{
		key := "kube-dns.addons.k8s.io"
		version := "1.14.4"

		{
			location := key + "/pre-k8s-1.6.yaml"
			id := "pre-k8s-1.6"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:              fi.String(key),
				Version:           fi.String(version),
				Selector:          map[string]string{"k8s-addon": key},
				Manifest:          fi.String(location),
				KubernetesVersion: "<1.6.0",
				Id:                id,
			})
			manifests[key+"-"+id] = "addons/" + location
		}

		{
			location := key + "/k8s-1.6.yaml"
			id := "k8s-1.6"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:              fi.String(key),
				Version:           fi.String(version),
				Selector:          map[string]string{"k8s-addon": key},
				Manifest:          fi.String(location),
				KubernetesVersion: ">=1.6.0",
				Id:                id,
			})
			manifests[key+"-"+id] = "addons/" + location
		}
	}

	{
		key := "limit-range.addons.k8s.io"
		version := "1.5.0"

		location := key + "/v" + version + ".yaml"

		addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
			Name:     fi.String(key),
			Version:  fi.String(version),
			Selector: map[string]string{"k8s-addon": key},
			Manifest: fi.String(location),
		})
		manifests[key] = "addons/" + location
	}

	{
		key := "dns-controller.addons.k8s.io"
		version := "1.7.1"

		{
			location := key + "/pre-k8s-1.6.yaml"
			id := "pre-k8s-1.6"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:              fi.String(key),
				Version:           fi.String(version),
				Selector:          map[string]string{"k8s-addon": key},
				Manifest:          fi.String(location),
				KubernetesVersion: "<1.6.0",
				Id:                id,
			})
			manifests[key+"-"+id] = "addons/" + location
		}

		{
			location := key + "/k8s-1.6.yaml"
			id := "k8s-1.6"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:              fi.String(key),
				Version:           fi.String(version),
				Selector:          map[string]string{"k8s-addon": key},
				Manifest:          fi.String(location),
				KubernetesVersion: ">=1.6.0",
				Id:                id,
			})
			manifests[key+"-"+id] = "addons/" + location
		}
	}

	if featureflag.EnableExternalDNS.Enabled() {
		{
			key := "external-dns.addons.k8s.io"
			version := "0.3.0"

			{
				location := key + "/pre-k8s-1.6.yaml"
				id := "pre-k8s-1.6"

				addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
					Name:              fi.String(key),
					Version:           fi.String(version),
					Selector:          map[string]string{"k8s-addon": key},
					Manifest:          fi.String(location),
					KubernetesVersion: "<1.6.0",
					Id:                id,
				})
				manifests[key+"-"+id] = "addons/" + location
			}

			{
				location := key + "/k8s-1.6.yaml"
				id := "k8s-1.6"

				addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
					Name:              fi.String(key),
					Version:           fi.String(version),
					Selector:          map[string]string{"k8s-addon": key},
					Manifest:          fi.String(location),
					KubernetesVersion: ">=1.6.0",
					Id:                id,
				})
				manifests[key+"-"+id] = "addons/" + location
			}
		}
	}

	{
		key := "storage-aws.addons.k8s.io"
		version := "1.6.0"

		location := key + "/v" + version + ".yaml"

		addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
			Name:     fi.String(key),
			Version:  fi.String(version),
			Selector: map[string]string{"k8s-addon": key},
			Manifest: fi.String(location),
		})
		manifests[key] = "addons/" + location
	}

	// The role.kubernetes.io/networking is used to label anything related to a networking addin,
	// so that if we switch networking plugins (e.g. calico -> weave or vice-versa), we'll replace the
	// old networking plugin, and there won't be old pods "floating around".

	// This means whenever we create or update a networking plugin, we should be sure that:
	// 1. the selector is role.kubernetes.io/networking=1
	// 2. every object in the manifest is labeleled with role.kubernetes.io/networking=1

	// TODO: Some way to test/enforce this?

	// TODO: Create "empty" configurations for others, so we can delete e.g. the kopeio configuration
	// if we switch to kubenet?

	// TODO: Create configuration object for cni providers (maybe create it but orphan it)?

	networkingSelector := map[string]string{"role.kubernetes.io/networking": "1"}

	if b.cluster.Spec.Networking.Kopeio != nil {
		key := "networking.kope.io"
		version := "1.0.20170406"

		{
			location := key + "/pre-k8s-1.6.yaml"
			id := "pre-k8s-1.6"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:              fi.String(key),
				Version:           fi.String(version),
				Selector:          networkingSelector,
				Manifest:          fi.String(location),
				KubernetesVersion: "<1.6.0",
				Id:                id,
			})
			manifests[key+"-"+id] = "addons/" + location
		}

		{
			location := key + "/k8s-1.6.yaml"
			id := "k8s-1.6"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:              fi.String(key),
				Version:           fi.String(version),
				Selector:          networkingSelector,
				Manifest:          fi.String(location),
				KubernetesVersion: ">=1.6.0",
				Id:                id,
			})
			manifests[key+"-"+id] = "addons/" + location
		}
	}

	if b.cluster.Spec.Networking.Weave != nil {
		key := "networking.weave"

		version := "2.0.1"

		{
			location := key + "/pre-k8s-1.6.yaml"
			id := "pre-k8s-1.6"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:              fi.String(key),
				Version:           fi.String(version),
				Selector:          networkingSelector,
				Manifest:          fi.String(location),
				KubernetesVersion: "<1.6.0",
				Id:                id,
			})
			manifests[key+"-"+id] = "addons/" + location
		}

		{
			location := key + "/k8s-1.6.yaml"
			id := "k8s-1.6"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:              fi.String(key),
				Version:           fi.String(version),
				Selector:          networkingSelector,
				Manifest:          fi.String(location),
				KubernetesVersion: ">=1.6.0",
				Id:                id,
			})
			manifests[key+"-"+id] = "addons/" + location
		}
	}

	if b.cluster.Spec.Networking.Flannel != nil {
		key := "networking.flannel"

		// 0.7.2-kops.1 = 0.7.1 + hairpinMode fix
		version := "0.7.2-kops.1"

		{
			location := key + "/pre-k8s-1.6.yaml"
			id := "pre-k8s-1.6"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:              fi.String(key),
				Version:           fi.String(version),
				Selector:          networkingSelector,
				Manifest:          fi.String(location),
				KubernetesVersion: "<1.6.0",
				Id:                id,
			})
			manifests[key+"-"+id] = "addons/" + location
		}

		{
			location := key + "/k8s-1.6.yaml"
			id := "k8s-1.6"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:              fi.String(key),
				Version:           fi.String(version),
				Selector:          networkingSelector,
				Manifest:          fi.String(location),
				KubernetesVersion: ">=1.6.0",
				Id:                id,
			})
			manifests[key+"-"+id] = "addons/" + location
		}
	}

	if b.cluster.Spec.Networking.Calico != nil {
		key := "networking.projectcalico.org"
		version := "2.4.1"

		{
			location := key + "/pre-k8s-1.6.yaml"
			id := "pre-k8s-1.6"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:              fi.String(key),
				Version:           fi.String(version),
				Selector:          networkingSelector,
				Manifest:          fi.String(location),
				KubernetesVersion: "<1.6.0",
				Id:                id,
			})
			manifests[key+"-"+id] = "addons/" + location
		}

		{
			location := key + "/k8s-1.6.yaml"
			id := "k8s-1.6"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:              fi.String(key),
				Version:           fi.String(version),
				Selector:          networkingSelector,
				Manifest:          fi.String(location),
				KubernetesVersion: ">=1.6.0",
				Id:                id,
			})
			manifests[key+"-"+id] = "addons/" + location
		}
	}

	if b.cluster.Spec.Networking.Canal != nil {
		key := "networking.projectcalico.org.canal"
		// Locking canal addon version to 2.4.1 (same as Calico node). Best to maintain lockstep for sanity
		version := "2.4.1"

		{
			location := key + "/pre-k8s-1.6.yaml"
			id := "pre-k8s-1.6"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:              fi.String(key),
				Version:           fi.String(version),
				Selector:          networkingSelector,
				Manifest:          fi.String(location),
				KubernetesVersion: "<1.6.0",
				Id:                id,
			})
			manifests[key+"-"+id] = "addons/" + location
		}

		{
			location := key + "/k8s-1.6.yaml"
			id := "k8s-1.6"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:              fi.String(key),
				Version:           fi.String(version),
				Selector:          networkingSelector,
				Manifest:          fi.String(location),
				KubernetesVersion: ">=1.6.0",
				Id:                id,
			})
			manifests[key+"-"+id] = "addons/" + location
		}
	}

	if b.cluster.Spec.Networking.Kuberouter != nil {
		key := "networking.kuberouter"

		version := "0.1.0"

		{
			location := key + "/k8s-1.6.yaml"
			id := "k8s-1.6"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:              fi.String(key),
				Version:           fi.String(version),
				Selector:          networkingSelector,
				Manifest:          fi.String(location),
				KubernetesVersion: ">=1.6.0",
				Id:                id,
			})
			manifests[key+"-"+id] = "addons/" + location
		}
	}

	if b.cluster.Spec.Networking.Romana != nil {
		key := "networking.romana"

		version := "v2.0-preview.2"

		{
			location := key + "/k8s-1.6.yaml"
			id := "k8s-1.6"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:              fi.String(key),
				Version:           fi.String(version),
				Selector:          networkingSelector,
				Manifest:          fi.String(location),
				KubernetesVersion: ">=1.6.0",
				Id:                id,
			})
			manifests[key+"-"+id] = "addons/" + location
		}
	}

	authenticationSelector := map[string]string{"role.kubernetes.io/authentication": "1"}

	if b.cluster.Spec.Authentication != nil && b.cluster.Spec.Authentication.Kopeio != nil {
		key := "authentication.kope.io"
		version := "1.0.20170619"

		{
			location := key + "/k8s-1.6.yaml"
			id := "k8s-1.6"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:              fi.String(key),
				Version:           fi.String(version),
				Selector:          authenticationSelector,
				Manifest:          fi.String(location),
				KubernetesVersion: ">=1.6.0",
				Id:                id,
			})
			manifests[key+"-"+id] = "addons/" + location
		}
	}

	return addons, manifests, nil
}
