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

package do

import (
	"errors"
	"fmt"

	"github.com/digitalocean/godo"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	dns "k8s.io/kops/dnsprovider/pkg/dnsprovider/providers/do"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

var _ fi.Cloud = (*doCloudMockImplementation)(nil)

type doCloudMockImplementation struct {
	Client *godo.Client

	region string
}

func BuildMockDOCloud(region string) *doCloudMockImplementation {
	return &doCloudMockImplementation{region: region, Client: godo.NewClient(nil)}
}

func (c *doCloudMockImplementation) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderDO
}

// Region returns the DO region we will target
func (c *doCloudMockImplementation) Region() string {
	return c.region
}

func (c *doCloudMockImplementation) DNS() (dnsprovider.Interface, error) {
	provider := dns.NewProvider(c.Client)
	return provider, nil
}

// FindVPCInfo is not implemented, it's only here to satisfy the fi.Cloud interface
func (c *doCloudMockImplementation) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return nil, errors.New("not implemented")
}

// DeleteGroup is not implemented yet, is a func that needs to delete a DO instance group.
func (c *doCloudMockImplementation) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	return fmt.Errorf("digital ocean cloud provider does not support deleting cloud groups at this time")
}

func (c *doCloudMockImplementation) DeleteInstance(instance *cloudinstances.CloudInstance) error {
	return errors.New("not tested")
}

// DetachInstance is not implemented yet. It needs to cause a cloud instance to no longer be counted against the group's size limits.
func (c *doCloudMockImplementation) DetachInstance(i *cloudinstances.CloudInstance) error {
	return fmt.Errorf("digital ocean cloud provider does not support surging")
}

func (c *doCloudMockImplementation) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	return nil, errors.New("not tested")
}

// FindClusterStatus discovers the status of the cluster, by inspecting the cloud objects
func (c *doCloudMockImplementation) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	return nil, errors.New("not tested")
}

func (c *doCloudMockImplementation) GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	return nil, errors.New("not tested")
}

func (c *doCloudMockImplementation) DropletsService() godo.DropletsService {
	return c.Client.Droplets
}

func (c *doCloudMockImplementation) DropletActionService() godo.DropletActionsService {
	return c.Client.DropletActions
}

func (c *doCloudMockImplementation) VolumeService() godo.StorageService {
	return c.Client.Storage
}

func (c *doCloudMockImplementation) VolumeActionService() godo.StorageActionsService {
	return c.Client.StorageActions
}

func (c *doCloudMockImplementation) LoadBalancersService() godo.LoadBalancersService {
	return c.Client.LoadBalancers
}

func (c *doCloudMockImplementation) DomainService() godo.DomainsService {
	return c.Client.Domains
}

func (c *doCloudMockImplementation) ActionsService() godo.ActionsService {
	return c.Client.Actions
}

func (c *doCloudMockImplementation) GetAllLoadBalancers() ([]godo.LoadBalancer, error) {
	return nil, nil
}

func (c *doCloudMockImplementation) GetAllDropletsByTag(tag string) ([]godo.Droplet, error) {
	return nil, nil
}

func (c *doCloudMockImplementation) GetAllVolumesByRegion() ([]godo.Volume, error) {
	return nil, nil
}

func (c *doCloudMockImplementation) GetVPCUUID(networkCIDR string, vpcName string) (string, error) {
	return "", nil
}

func (c *doCloudMockImplementation) GetAllVPCs() ([]*godo.VPC, error) {
	return nil, nil
}

func (c *doCloudMockImplementation) VPCsService() godo.VPCsService {
	return c.Client.VPCs
}
