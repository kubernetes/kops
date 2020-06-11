package kopscontroller

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
)

// Remap remaps the kopscontroller addon
func Remap(context *model.KopsModelContext, objects []*kubemanifest.Object) error {
	var daemonsets []*kubemanifest.Object
	for _, object := range objects {
		if object.Kind() != "DaemonSet" {
			continue
		}
		if object.APIVersion() != "apps/v1" {
			continue
		}
		daemonsets = append(daemonsets, object)
	}

	if len(daemonsets) != 1 {
		return fmt.Errorf("expected exactly one daemonset in kops-controller manifest, found %d", len(daemonsets))
	}

	podSpec := &corev1.PodSpec{}
	if err := daemonsets[0].Reparse(podSpec, "spec", "template", "spec"); err != nil {
		return fmt.Errorf("failed to parse spec.template.spec from Daemonset: %v", err)
	}

	containers := podSpec.Containers
	if len(containers) != 1 {
		return fmt.Errorf("expected exactly one container in kops-controller Daemonset, found %d", len(containers))
	}
	container := &containers[0]

	if featureflag.UsePodIAM.Enabled() {
		if err := iam.AddPodRole(&context.IAMModelContext, podSpec, container, iam.PodRoleKopsController); err != nil {
			return err
		}

		if err := daemonsets[0].Set(podSpec, "spec", "template", "spec"); err != nil {
			return err
		}
	}

	return nil
}
