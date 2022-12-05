/*
Copyright 2020 The Kubernetes Authors.

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

package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-05-01/network"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/resolver"
	"k8s.io/kops/protokube/pkg/gossip"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

type client interface {
	ListVMScaleSets(ctx context.Context) ([]compute.VirtualMachineScaleSet, error)
	ListVMSSNetworkInterfaces(ctx context.Context, vmScaleSetName string) ([]network.Interface, error)
}

var _ client = &Client{}

// SeedProvider is an Azure implementation of gossip.SeedProvider.
type SeedProvider struct {
	client client
	tags   map[string]string
}

var _ gossip.SeedProvider = &SeedProvider{}

// NewSeedProvider returns a new SeedProvider.
func NewSeedProvider(client client, tags map[string]string) (*SeedProvider, error) {
	return &SeedProvider{
		client: client,
		tags:   tags,
	}, nil
}

// GetSeeds returns a slice of strings used as seeds of Gossip.
// This follows the implementation of AWS and creates seeds from
// private IPs of VMs in the cluster.
func (p *SeedProvider) GetSeeds() ([]string, error) {
	return p.discover(context.TODO(), nil)
}

func (p *SeedProvider) discover(ctx context.Context, predicate func(*compute.VirtualMachineScaleSet) bool) ([]string, error) {
	vmsses, err := p.client.ListVMScaleSets(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing VM Scale Sets: %s", err)
	}

	var vmssNames []string
	for _, vmss := range vmsses {
		if predicate != nil && !predicate(&vmss) {
			continue
		}
		if p.isVMSSForCluster(&vmss) {
			vmssNames = append(vmssNames, *vmss.Name)
		}
	}
	klog.V(2).Infof("Found %d VM Scale Sets for the cluster (out of %d)", len(vmssNames), len(vmsses))

	var seeds []string
	for _, vmssName := range vmssNames {
		ifaces, err := p.client.ListVMSSNetworkInterfaces(ctx, vmssName)
		if err != nil {
			return nil, fmt.Errorf("error listing VMSS network interfaces: %s", err)
		}
		for _, iface := range ifaces {
			for _, i := range *iface.IPConfigurations {
				seeds = append(seeds, *i.PrivateIPAddress)
			}
		}
	}
	return seeds, nil
}

func (p *SeedProvider) isVMSSForCluster(vmss *compute.VirtualMachineScaleSet) bool {
	found := 0
	for k, v := range vmss.Tags {
		if p.tags[k] == *v {
			found++
		}
	}
	// TODO(kenji): Filter by ProvisioningState if necessary.
	return found == len(p.tags)
}

var _ resolver.Resolver = &SeedProvider{}

// Resolve implements resolver.Resolve, providing name -> address resolution using cloud API discovery.
func (p *SeedProvider) Resolve(ctx context.Context, name string) ([]string, error) {
	klog.Infof("trying to resolve %q using SeedProvider", name)

	// We assume we are trying to resolve a component that runs on the control plane
	isControlPlane := func(vmss *compute.VirtualMachineScaleSet) bool {
		for k := range vmss.Tags {
			switch k {
			case azure.TagNameRolePrefix + kops.InstanceGroupRoleControlPlane.ToLowerString():
				return true
			case azure.TagNameRolePrefix + "master":
				return true
			}
		}
		return false
	}

	// TODO: Can we push the predicate down so we can filter server-side?
	return p.discover(ctx, isControlPlane)
}
