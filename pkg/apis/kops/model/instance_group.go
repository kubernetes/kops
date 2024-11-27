/*
Copyright 2024 The Kubernetes Authors.

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
	"fmt"

	"k8s.io/kops/pkg/apis/kops"
)

// InstanceGroup is a subset of the full Cluster and InstanceGroup functionality,
// that gives us some abstraction over the raw types.
type InstanceGroup interface {
	// KubernetesVersion returns the Kubernetes version for the instance group
	KubernetesVersion() *KubernetesVersion

	// GetCloudProvider returns the cloud provider for the instance group
	GetCloudProvider() kops.CloudProviderID

	// RawClusterSpec returns the cluster spec for the instance group.
	// If possible, prefer abstracted methods over accessing this data directly.
	RawClusterSpec() *kops.ClusterSpec
}

// ForInstanceGroup creates an InstanceGroup model for the given cluster and instance group.
func ForInstanceGroup(cluster *kops.Cluster, ig *kops.InstanceGroup) (InstanceGroup, error) {
	kubernetesVersionString := cluster.Spec.KubernetesVersion
	kubernetesVersion, err := ParseKubernetesVersion(kubernetesVersionString)
	if err != nil {
		return nil, fmt.Errorf("error parsing Kubernetes version %q: %v", kubernetesVersionString, err)
	}

	return &instanceGroupModel{cluster: cluster, ig: ig, kubernetesVersion: kubernetesVersion}, nil
}

// instanceGroupModel is a concrete implementation of InstanceGroup.
type instanceGroupModel struct {
	cluster           *kops.Cluster
	ig                *kops.InstanceGroup
	kubernetesVersion *KubernetesVersion
}

var _ InstanceGroup = &instanceGroupModel{}

func (m *instanceGroupModel) KubernetesVersion() *KubernetesVersion {
	return m.kubernetesVersion
}

func (m *instanceGroupModel) GetCloudProvider() kops.CloudProviderID {
	return m.cluster.GetCloudProvider()
}

func (m *instanceGroupModel) RawClusterSpec() *kops.ClusterSpec {
	return &m.cluster.Spec
}
