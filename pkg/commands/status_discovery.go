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

package commands

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/cloudup"
)

// CloudDiscoveryStatusStore implements status.Store by inspecting cloud objects.
// Likely temporary until we validate our status usage
type CloudDiscoveryStatusStore struct {
}

var _ kops.StatusStore = &CloudDiscoveryStatusStore{}

func (s *CloudDiscoveryStatusStore) GetApiIngressStatus(cluster *kops.Cluster) ([]kops.ApiIngressStatus, error) {
	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return nil, err
	}

	return cloud.GetApiIngressStatus(cluster)
}

// FindClusterStatus discovers the status of the cluster, by inspecting the cloud objects
func (s *CloudDiscoveryStatusStore) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return nil, err
	}

	return cloud.FindClusterStatus(cluster)
}
