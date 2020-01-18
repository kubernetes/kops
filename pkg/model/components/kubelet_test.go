/*
Copyright 2019 The Kubernetes Authors.

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

func buildKubeletTestCluster() *kops.Cluster {
	return &kops.Cluster{
		Spec: kops.ClusterSpec{
			KubernetesVersion:     "1.6.2",
			ServiceClusterIPRange: "10.10.0.0/16",
			Kubelet:               &kops.KubeletConfigSpec{},
		},
	}
}

func buildOptions(cluster *kops.Cluster) error {
	ab := assets.NewAssetBuilder(cluster, "")

	ver, err := KubernetesVersion(&cluster.Spec)
	if err != nil {
		return err
	}

	builder := KubeletOptionsBuilder{
		Context: &OptionsContext{
			AssetBuilder:      ab,
			KubernetesVersion: *ver,
		},
	}

	err = builder.BuildOptions(&cluster.Spec)
	if err != nil {
		return nil
	}

	return nil
}

func TestFeatureGates(t *testing.T) {
	cluster := buildKubeletTestCluster()
	err := buildOptions(cluster)
	if err != nil {
		t.Fatal(err)
	}

	gates := cluster.Spec.Kubelet.FeatureGates
	if gates["ExperimentalCriticalPodAnnotation"] != "true" {
		t.Errorf("ExperimentalCriticalPodAnnotation feature gate should be enabled by default")
	}
}

func TestFeatureGatesKubernetesVersion(t *testing.T) {
	cluster := buildKubeletTestCluster()
	cluster.Spec.KubernetesVersion = "1.16.0"
	err := buildOptions(cluster)
	if err != nil {
		t.Fatal(err)
	}

	gates := cluster.Spec.Kubelet.FeatureGates
	if _, found := gates["ExperimentalCriticalPodAnnotation"]; found {
		t.Errorf("ExperimentalCriticalPodAnnotation feature gate should not be added on Kubernetes >= 1.16.0")
	}
}

func TestFeatureGatesOverride(t *testing.T) {
	cluster := buildKubeletTestCluster()
	cluster.Spec.Kubelet.FeatureGates = map[string]string{
		"ExperimentalCriticalPodAnnotation": "false",
	}

	err := buildOptions(cluster)
	if err != nil {
		t.Fatal(err)
	}

	gates := cluster.Spec.Kubelet.FeatureGates
	if gates["ExperimentalCriticalPodAnnotation"] != "false" {
		t.Errorf("ExperimentalCriticalPodAnnotation feature should be disalbled")
	}
}
