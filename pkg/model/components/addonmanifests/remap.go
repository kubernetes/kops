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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/klog/v2"
	addonsapi "k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/components/addonmanifests/awsebscsidriver"
	"k8s.io/kops/pkg/model/components/addonmanifests/awsloadbalancercontroller"
	"k8s.io/kops/pkg/model/components/addonmanifests/clusterautoscaler"
	"k8s.io/kops/pkg/model/components/addonmanifests/dnscontroller"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi"
)

func RemapAddonManifest(addon *addonsapi.AddonSpec, context *model.KopsModelContext, assetBuilder *assets.AssetBuilder, manifest []byte) ([]byte, error) {
	name := fi.StringValue(addon.Name)

	{
		objects, err := kubemanifest.LoadObjectsFrom(manifest)
		if err != nil {
			return nil, err
		}

		if name == "dns-controller.addons.k8s.io" {
			if err := dnscontroller.Remap(context, addon, objects); err != nil {
				return nil, err
			}
		}

		err = addLabels(addon, objects)
		if err != nil {
			return nil, fmt.Errorf("failed to annotate %q: %w", name, err)
		}

		err = addServiceAccountRole(context, objects)
		if err != nil {
			return nil, fmt.Errorf("failed to add service account for %q: %w", name, err)
		}

		b, err := objects.ToYAML()
		if err != nil {
			return nil, err
		}
		manifest = b
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

func addServiceAccountRole(context *model.KopsModelContext, objects kubemanifest.ObjectList) error {
	if !featureflag.UseServiceAccountIAM.Enabled() {
		return nil
	}

	for _, object := range objects {
		if object.Kind() != "Deployment" {
			continue
		}
		if object.APIVersion() != "apps/v1" {
			continue
		}
		podSpec := &corev1.PodSpec{}

		if err := object.Reparse(podSpec, "spec", "template", "spec"); err != nil {
			return fmt.Errorf("failed to parse spec.template.spec from Deployment: %v", err)
		}
		sa := podSpec.ServiceAccountName
		subject := getWellknownServiceAccount(sa)
		if subject == nil {
			continue
		}

		if err := iam.AddServiceAccountRole(&context.IAMModelContext, podSpec, subject); err != nil {
			return err
		}

		if err := object.Set(podSpec, "spec", "template", "spec"); err != nil {
			return fmt.Errorf("failed to set object: %w", err)
		}

	}
	return nil
}

func getWellknownServiceAccount(name string) iam.Subject {
	switch name {
	case "aws-load-balancer-controller":
		return &awsloadbalancercontroller.ServiceAccount{}
	case "cluster-autoscaler":
		return &clusterautoscaler.ServiceAccount{}
	case "ebs-csi-controller-sa":
		return &awsebscsidriver.ServiceAccount{}
	case "external-dns":
		return &awsebscsidriver.ServiceAccount{}
	default:
		return nil
	}
}

func addLabels(addon *addonsapi.AddonSpec, objects kubemanifest.ObjectList) error {

	for _, object := range objects {
		meta := &metav1.ObjectMeta{}
		err := object.Reparse(meta, "metadata")
		if err != nil {
			return fmt.Errorf("Failed to annotate %T", object)
		}

		if meta.Labels == nil {
			meta.Labels = make(map[string]string)
		}

		meta.Labels["app.kubernetes.io/managed-by"] = "kops"
		meta.Labels["addon.kops.k8s.io/name"] = *addon.Name

		// ensure selector is set where applicable
		for key, val := range addon.Selector {
			existingVal, ok := meta.Labels[key]
			if ok && existingVal != val {
				return fmt.Errorf("label %q already set to %q while it should be %q", key, meta.Labels[key], val)
			}

			meta.Labels[key] = val
		}
		object.Set(meta, "metadata")
	}
	return nil
}
