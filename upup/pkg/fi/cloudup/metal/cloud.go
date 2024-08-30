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

package metal

import (
	"fmt"
	"net"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

var _ fi.Cloud = &Cloud{}

// Cloud holds the fi.Cloud implementation for metal resources.
type Cloud struct {
}

// NewCloud returns a Cloud for metal resources.
func NewCloud() (*Cloud, error) {
	cloud := &Cloud{}
	return cloud, nil
}

func (c *Cloud) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderMetal
}
func (c *Cloud) DNS() (dnsprovider.Interface, error) {
	return nil, fmt.Errorf("method not implemented")
}

// FindVPCInfo looks up the specified VPC by id, returning info if found, otherwise (nil, nil).
func (c *Cloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return nil, fmt.Errorf("method not implemented")
}

// DeleteInstance deletes a cloud instance.
func (c *Cloud) DeleteInstance(instance *cloudinstances.CloudInstance) error {
	return fmt.Errorf("method not implemented")
}

// // DeregisterInstance drains a cloud instance and loadbalancers.
func (c *Cloud) DeregisterInstance(instance *cloudinstances.CloudInstance) error {
	return fmt.Errorf("method not implemented")
}

// DeleteGroup deletes the cloud resources that make up a CloudInstanceGroup, including the instances.
func (c *Cloud) DeleteGroup(group *cloudinstances.CloudInstanceGroup) error {
	return fmt.Errorf("method not implemented")
}

// DetachInstance causes a cloud instance to no longer be counted against the group's size limits.
func (c *Cloud) DetachInstance(instance *cloudinstances.CloudInstance) error {
	return fmt.Errorf("method not implemented")
}

// GetCloudGroups returns a map of cloud instances that back a kops cluster.
// Detached instances must be returned in the NeedUpdate slice.
func (c *Cloud) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	return nil, fmt.Errorf("method not implemented")
}

// Region returns the cloud region bound to the cloud instance.
// If the region concept does not apply, returns "".
func (c *Cloud) Region() string {
	return ""
}

// FindClusterStatus discovers the status of the cluster, by inspecting the cloud objects
func (c *Cloud) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	// etcdStatus, err := findEtcdStatus(c, cluster)
	// if err != nil {
	//      return nil, err
	// }
	klog.Warningf("method metal.Cloud::FindClusterStatus stub-implemented")
	return &kops.ClusterStatus{
		// EtcdClusters: etcdStatus,
	}, nil
}

func (c *Cloud) GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	var ret []fi.ApiIngressStatus
	publicName := cluster.Spec.API.PublicName
	if publicName == "" {
		return ret, fmt.Errorf("spec.api.publicName must be set for bare metal")
	}
	ip := net.ParseIP(publicName)
	if ip == nil {
		ret = append(ret, fi.ApiIngressStatus{
			Hostname: publicName,
		})
	} else {
		ret = append(ret, fi.ApiIngressStatus{
			IP: publicName,
		})
	}
	return ret, nil
}
