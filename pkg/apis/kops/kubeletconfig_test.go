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

package kops

import (
	"testing"
)

func Test_InstanceGroupKubeletMerge(t *testing.T) {
	var cluster = &Cluster{}
	cluster.Spec.Kubelet = &KubeletConfigSpec{}
	cluster.Spec.Kubelet.NvidiaGPUs = 0

	var instanceGroup = &InstanceGroup{}
	instanceGroup.Spec.Kubelet = &KubeletConfigSpec{}
	instanceGroup.Spec.Kubelet.NvidiaGPUs = 1

	var mergedKubeletSpec, err = BuildKubeletConfigSpec(cluster, instanceGroup)
	if err != nil {
		t.Error(err)
	}
	if mergedKubeletSpec == nil {
		t.Error("Returned nil kubelet spec")
	}

	if mergedKubeletSpec.NvidiaGPUs != instanceGroup.Spec.Kubelet.NvidiaGPUs {
		t.Errorf("InstanceGroup kubelet value (%d) should be reflected in merged output", instanceGroup.Spec.Kubelet.NvidiaGPUs)
	}
}
