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

package openstack

import (
	"fmt"

	"k8s.io/api/core/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
)

type OpenstackCloud interface {
	fi.Cloud
}

type openstackCloud struct {
}

var _ fi.Cloud = &openstackCloud{}

func NewOpenstackCloud() (OpenstackCloud, error) {
	return &openstackCloud{}, nil
}

func (c *openstackCloud) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderOpenstack
}

func (c *openstackCloud) DNS() (dnsprovider.Interface, error) {
	return nil, fmt.Errorf("has not implemented this method")
}

func (c *openstackCloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return nil, fmt.Errorf("has not implemented this method")
}

func (c *openstackCloud) DeleteInstance(i *cloudinstances.CloudInstanceGroupMember) error {
	return fmt.Errorf("has not implemented this method")
}

func (c *openstackCloud) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	return fmt.Errorf("has not implemented this method")
}

func (c *openstackCloud) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	return nil, fmt.Errorf("has not implemented this method")
}
