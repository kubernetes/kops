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

package baremetal

import (
	"fmt"

	"github.com/golang/glog"

	"k8s.io/api/core/v1"
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

func (c *Cloud) DNS() (dnsprovider.Interface, error) {
	return c.dns, nil
}

func (c *Cloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return nil, fmt.Errorf("baremetal FindVPCInfo not supported")
}

// GetCloudGroups is not implemented yet, that needs to return the instances and groups that back a kops cluster.
// Baremetal may not support this.
func (c *Cloud) GetCloudGroups(*kops.Cluster, []*kops.InstanceGroup, bool, []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	glog.V(8).Infof("baremetal cloud GetCloudGroups not implemented yet")
	return nil, fmt.Errorf("baremetal provider does not support getting cloud groups at this time")
}

// GetCloudGroupStatus is not implemented yet, that needs to return the instances and groups that back a kops cluster.
// Baremetal may not support this.
func (c *Cloud) GetCloudGroupStatus(*kops.Cluster, string) (int, int, error) {
	glog.V(8).Infof("baremetal cloud GetCloudGroupStatus not implemented yet")
	return 0, 0, fmt.Errorf("baremetal provider does not support getting cloud groups at this time")
}

// SetTerminationPolicy is not implemented yet
// Baremetal may not support this.
func (c *Cloud) SetTerminationPolicy(*kops.Cluster, string, []cloudinstances.TerminationPolicy) error {
	glog.V(8).Infof("baremetal cloud SetTerminationPolicy not implemented yet")
	return fmt.Errorf("baremetal provider does not support setting termination policy at this time")
}

// DeleteGroup is not implemented yet, is a func that needs to delete a DO instance group.
// Baremetal may not support this.
func (c *Cloud) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	glog.V(8).Infof("baremetal cloud provider DeleteGroup not implemented yet")
	return fmt.Errorf("baremetal cloud provider does not support deleting cloud groups at this time")
}

//DeleteInstance is not implemented yet, is func needs to delete a DO instance.
//Baremetal may not support this.
func (c *Cloud) DeleteInstance(instance *cloudinstances.CloudInstanceGroupMember) error {
	glog.V(8).Infof("baremetal cloud provider DeleteInstance not implemented yet")
	return fmt.Errorf("baremetal cloud provider does not support deleting cloud instances at this time")
}
