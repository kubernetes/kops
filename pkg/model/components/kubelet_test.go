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

package components

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
)

func buildSpec() *kops.ClusterSpec {
	spec := kops.ClusterSpec{
		KubernetesVersion:     "1.6.2",
		ServiceClusterIPRange: "10.10.0.0/16",
		Kubelet:               &kops.KubeletConfigSpec{},
	}

	return &spec
}

func buildOptions(spec *kops.ClusterSpec) error {
	ab := assets.NewAssetBuilder(nil)

	ver, err := KubernetesVersion(spec)
	if err != nil {
		return err
	}

	builder := KubeletOptionsBuilder{
		Context: &OptionsContext{
			AssetBuilder:      ab,
			KubernetesVersion: *ver,
		},
	}

	err = builder.BuildOptions(spec)
	if err != nil {
		return nil
	}

	return nil
}

func TestFeatureGates(t *testing.T) {
	spec := buildSpec()
	err := buildOptions(spec)
	if err != nil {
		t.Fatal(err)
	}

	gates := spec.Kubelet.FeatureGates
	if gates["ExperimentalCriticalPodAnnotation"] != "true" {
		t.Errorf("ExperimentalCriticalPodAnnotation feature gate should be enabled by default")
	}
}

func TestFeatureGatesKubernetesVersion(t *testing.T) {
	spec := buildSpec()
	spec.KubernetesVersion = "1.4.0"
	err := buildOptions(spec)
	if err != nil {
		t.Fatal(err)
	}

	gates := spec.Kubelet.FeatureGates
	if _, found := gates["ExperimentalCriticalPodAnnotation"]; found {
		t.Errorf("ExperimentalCriticalPodAnnotation feature gate should not be added on Kubernetes < 1.5.2")
	}
}

func TestFeatureGatesOverride(t *testing.T) {
	spec := buildSpec()
	spec.Kubelet.FeatureGates = map[string]string{
		"ExperimentalCriticalPodAnnotation": "false",
	}

	err := buildOptions(spec)
	if err != nil {
		t.Fatal(err)
	}

	gates := spec.Kubelet.FeatureGates
	if gates["ExperimentalCriticalPodAnnotation"] != "false" {
		t.Errorf("ExperimentalCriticalPodAnnotation feature should be disalbled")
	}
}
