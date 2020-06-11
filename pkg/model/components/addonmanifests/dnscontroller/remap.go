package dnscontroller

import (
	"fmt"

	"github.com/blang/semver"
	corev1 "k8s.io/api/core/v1"
	addonsapi "k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
)

// Remap remaps the dns-controller addon
func Remap(context *model.KopsModelContext, addon *addonsapi.AddonSpec, objects []*kubemanifest.Object) error {
	if !featureflag.UsePodIAM.Enabled() {
		return nil
	}

	if addon.KubernetesVersion != "" {
		versionRange, err := semver.ParseRange(addon.KubernetesVersion)
		if err != nil {
			return fmt.Errorf("cannot parse KubernetesVersion=%q", addon.KubernetesVersion)
		}

		// TODO: do we need to loop through 19, 20, 21 ... to infinity???
		if !versionRange(semver.Version{Major: 1, Minor: 19}) {
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

	if featureflag.UsePodIAM.Enabled() {
		if err := iam.AddPodRole(&context.IAMModelContext, podSpec, container, iam.PodRoleDNSController); err != nil {
			return err
		}

		if err := deployments[0].Set(podSpec, "spec", "template", "spec"); err != nil {
			return err
		}
	}

	return nil
}
