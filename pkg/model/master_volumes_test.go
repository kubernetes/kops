/*
Copyright 2021 The Kubernetes Authors.

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

package model

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linodetasks"
)

func TestValidateAWSVolumeAllow50ratio(t *testing.T) {
	volumeName := "a"
	volumeType := "io1"
	volumeIops := 1000
	volumeThroughput := 0
	volumeSize := 20

	err := validateAWSVolume(volumeName, volumeType, int32(volumeSize), int32(volumeIops), int32(volumeThroughput))
	if err != nil {
		t.Errorf("Failed to validate valid etcd member spec: %v", err)
	}
}

func TestMasterVolumeBuilderBuildLinode(t *testing.T) {
	cluster := testutils.BuildMinimalClusterAWS("linode.k8s.local")
	cluster.Spec.CloudProvider = kops.CloudProviderSpec{Linode: &kops.LinodeSpec{}}

	var instanceGroups []*kops.InstanceGroup
	for _, subnet := range cluster.Spec.Networking.Subnets {
		ig := testutils.BuildMinimalMasterInstanceGroup(subnet.Name)
		instanceGroups = append(instanceGroups, &ig)
	}

	b := &MasterVolumeBuilder{
		KopsModelContext: &KopsModelContext{
			IAMModelContext:   iam.IAMModelContext{Cluster: cluster},
			AllInstanceGroups: instanceGroups,
			InstanceGroups:    instanceGroups,
		},
	}

	c := &fi.CloudupModelBuilderContext{Tasks: map[string]fi.CloudupTask{}}
	if err := b.Build(c); err != nil {
		t.Fatalf("unexpected error from Build(): %v", err)
	}

	expectedTasks := 0
	for _, etcd := range cluster.Spec.EtcdClusters {
		expectedTasks += len(etcd.Members)
	}

	if got := len(c.Tasks); got != expectedTasks {
		t.Fatalf("expected %d master volume tasks for linode, got %d", expectedTasks, got)
	}

	for key, task := range c.Tasks {
		if _, ok := task.(*linodetasks.Volume); !ok {
			t.Fatalf("expected task %q to be *linodetasks.Volume, got %T", key, task)
		}
	}
}
