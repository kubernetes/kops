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

package dnscontroller

import (
	"fmt"

	"github.com/blang/semver/v4"
	corev1 "k8s.io/api/core/v1"
	addonsapi "k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
)

// Remap remaps the dns-controller addon
func Remap(context *model.KopsModelContext, addon *addonsapi.AddonSpec, objects []*kubemanifest.Object) error {
	if !context.UseServiceAccountIAM() {
		return nil
	}

	if addon.KubernetesVersion != "" {
		versionRange, err := semver.ParseRange(addon.KubernetesVersion)
		if err != nil {
			return fmt.Errorf("cannot parse KubernetesVersion=%q", addon.KubernetesVersion)
		}

		if !kubernetesRangesIntersect(versionRange, semver.MustParseRange(">= 1.19.0")) {
			// Skip; this is an older manifest
			return nil
		}
	}

	var deployments []*kubemanifest.Object
	for _, object := range objects {
		if object.Kind() != "Deployment" {
			continue
		}
		if object.APIVersion() != "apps/v1" {
			continue
		}
		deployments = append(deployments, object)
	}

	if len(deployments) != 1 {
		return fmt.Errorf("expected exactly one Deployment in dns-controller manifest, found %d", len(deployments))
	}

	podSpec := &corev1.PodSpec{}
	if err := deployments[0].Reparse(podSpec, "spec", "template", "spec"); err != nil {
		return fmt.Errorf("failed to parse spec.template.spec from Deployment: %v", err)
	}

	containers := podSpec.Containers
	if len(containers) != 1 {
		return fmt.Errorf("expected exactly one container in dns-controller Deployment, found %d", len(containers))
	}
	container := &containers[0]

	if err := iam.AddServiceAccountRole(&context.IAMModelContext, podSpec, container, iam.ServiceAccountRoleDNSController); err != nil {
		return err
	}

	if err := deployments[0].Set(podSpec, "spec", "template", "spec"); err != nil {
		return err
	}

	return nil
}

// kubernetesRangesIntersect returns true if the two semver ranges overlap
// Sadly there's no actual function to do this.
// Instead we restrict to kubernetes versions, and just probe with 1.1, 1.2, 1.3 etc.
// This will therefore be inaccurate if there's a patch specifier
func kubernetesRangesIntersect(r1, r2 semver.Range) bool {
	for minor := 1; minor < 99; minor++ {
		v := semver.Version{Major: 1, Minor: uint64(minor), Patch: 0}
		if r1(v) && r2(v) {
			return true
		}
	}
	return false
}
