/*
Copyright 2017 The Kubernetes Authors.

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

package codecs

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops/v1alpha3"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/pkg/kopscodecs"
)

func TestSerializeEmptyCluster(t *testing.T) {
	cluster := &v1alpha3.Cluster{}
	cluster.Spec.Kubelet = &v1alpha3.KubeletConfigSpec{}
	cluster.Spec.KubeControllerManager = &v1alpha3.KubeControllerManagerConfig{}
	yaml, err := kopscodecs.ToVersionedYamlWithVersion(cluster, v1alpha3.SchemeGroupVersion)
	if err != nil {
		t.Errorf("unexpected error marshaling Cluster: %v", err)
	}

	yamlString := string(yaml)
	expected := `apiVersion: kops.k8s.io/v1alpha3
kind: Cluster
metadata:
  creationTimestamp: null
spec:
  cloudProvider: {}
  kubeControllerManager: {}
  kubelet: {}
`
	if yamlString != expected {
		diffString := diff.FormatDiff(expected, yamlString)
		t.Logf("diff:\n%s\n", diffString)
		t.Errorf("unexpected yaml from empty Cluster: %q", yamlString)
	}
}
