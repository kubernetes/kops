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

	corev1 "k8s.io/api/core/v1"
	addonsapi "k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
)

// Remap remaps the dns-controller addon
func Remap(context *model.KopsModelContext, addon *addonsapi.AddonSpec, objects []*kubemanifest.Object) error {
	if !context.UseServiceAccountExternalPermissions() {
		return nil
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

	if err := iam.AddServiceAccountRole(&context.IAMModelContext, podSpec, &ServiceAccount{}); err != nil {
		return err
	}

	if err := deployments[0].Set(podSpec, "spec", "template", "spec"); err != nil {
		return err
	}

	return nil
}
