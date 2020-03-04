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

package baremetal

import (
	"fmt"

	"k8s.io/klog"

	v1 "k8s.io/api/core/v1"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

type Cloud struct {
	dns dnsprovider.Interface
}

var _ fi.Cloud = &Cloud{}

func NewCloud(dns dnsprovider.Interface) (*Cloud, error) {
	return &Cloud{dns: dns}, nil
}

func (c *Cloud) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderBareMetal
}

func (c *Cloud) Region() string {
	return ""
}

func (c *Cloud) DNS() (dnsprovider.Interface, error) {
	return c.dns, nil
}

func (c *Cloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return nil, fmt.Errorf("baremetal FindVPCInfo not supported")
}

// GetCloudGroups is not implemented yet, that needs to return the instances and groups that back a kops cluster.
// Baremetal may not support this.
func (c *Cloud) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	klog.V(8).Infof("baremetal cloud GetCloudGroups not implemented yet")
	return nil, fmt.Errorf("baremetal provider does not support getting cloud groups at this time")
}

// DeleteGroup is not implemented yet, is a func that needs to delete a DO instance group.
// Baremetal may not support this.
func (c *Cloud) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	klog.V(8).Infof("baremetal cloud provider DeleteGroup not implemented yet")
	return fmt.Errorf("baremetal cloud provider does not support deleting cloud groups at this time")
}

// DetachInstance is not implemented yet. It needs to cause a cloud instance to no longer be counted against the group's size limits.
// Baremetal may not support this.
func (c *Cloud) DetachInstance(i *cloudinstances.CloudInstanceGroupMember) error {
	klog.V(8).Infof("baremetal cloud provider DetachInstance not implemented")
	return fmt.Errorf("baremetal cloud provider does not support surging")
}

//DeleteInstance is not implemented yet, is func needs to delete a DO instance.
//Baremetal may not support this.
func (c *Cloud) DeleteInstance(instance *cloudinstances.CloudInstanceGroupMember) error {
	klog.V(8).Infof("baremetal cloud provider DeleteInstance not implemented yet")
	return fmt.Errorf("baremetal cloud provider does not support deleting cloud instances at this time")
}
