/*
Copyright 2023 The Kubernetes Authors.

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

package azuretasks

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-05-01/network"
	"github.com/Azure/go-autorest/autorest/to"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

// NetworkSecurityGroup is an Azure Cloud Network Security Group
// +kops:fitask
type NetworkSecurityGroup struct {
	Name          *string
	ID            *string
	Lifecycle     fi.Lifecycle
	ResourceGroup *ResourceGroup

	SecurityRules []*NetworkSecurityRule

	Tags map[string]*string
}

var (
	_ fi.CloudupTask          = &NetworkSecurityGroup{}
	_ fi.CompareWithID        = &NetworkSecurityGroup{}
	_ fi.CloudupTaskNormalize = &NetworkSecurityGroup{}
)

// CompareWithID returns the Name of the Network Security Group
func (nsg *NetworkSecurityGroup) CompareWithID() *string {
	return nsg.ID
}

// Find discovers the Network Security Group in the cloud provider
func (nsg *NetworkSecurityGroup) Find(c *fi.CloudupContext) (*NetworkSecurityGroup, error) {
	cloud := c.T.Cloud.(azure.AzureCloud)
	l, err := cloud.NetworkSecurityGroup().List(context.TODO(), *nsg.ResourceGroup.Name)
	if err != nil {
		return nil, err
	}
	var found *network.SecurityGroup
	for _, v := range l {
		if *v.Name == *nsg.Name {
			found = &v
			break
		}
	}
	if found == nil {
		return nil, nil
	}

	actual := &NetworkSecurityGroup{
		Name:      nsg.Name,
		Lifecycle: nsg.Lifecycle,
		ResourceGroup: &ResourceGroup{
			Name: nsg.ResourceGroup.Name,
		},
		ID:   found.ID,
		Tags: found.Tags,
	}
	for _, rule := range *found.SecurityRules {
		nsr := &NetworkSecurityRule{
			Name:                       rule.Name,
			Priority:                   rule.Priority,
			Access:                     rule.Access,
			Direction:                  rule.Direction,
			Protocol:                   rule.Protocol,
			SourceAddressPrefix:        rule.SourceAddressPrefix,
			SourceAddressPrefixes:      rule.SourceAddressPrefixes,
			SourcePortRange:            rule.SourcePortRange,
			DestinationAddressPrefix:   rule.DestinationAddressPrefix,
			DestinationAddressPrefixes: rule.DestinationAddressPrefixes,
			DestinationPortRange:       rule.DestinationPortRange,
		}
		actual.SecurityRules = append(actual.SecurityRules, nsr)
	}

	nsg.ID = found.ID

	return actual, nil
}

func (nsg *NetworkSecurityGroup) Normalize(c *fi.CloudupContext) error {
	c.T.Cloud.(azure.AzureCloud).AddClusterTags(nsg.Tags)
	return nil
}

// Run implements fi.Task.Run.
func (nsg *NetworkSecurityGroup) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(nsg, c)
}

// CheckChanges returns an error if a change is not allowed.
func (*NetworkSecurityGroup) CheckChanges(a, e, changes *NetworkSecurityGroup) error {
	if a == nil {
		// Check if required fields are set when a new resource is created.
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		return nil
	}

	// Check if unchangeable fields won't be changed.
	if changes.Name != nil {
		return fi.CannotChangeField("Name")
	}
	return nil
}

// RenderAzure creates or updates a Network Security Group.
func (*NetworkSecurityGroup) RenderAzure(t *azure.AzureAPITarget, a, e, changes *NetworkSecurityGroup) error {
	if a == nil {
		klog.Infof("Creating a new Network Security Group with name: %s", fi.ValueOf(e.Name))
	} else {
		klog.Infof("Updating a Network Security Group with name: %s", fi.ValueOf(e.Name))
	}

	p := network.SecurityGroup{
		SecurityGroupPropertiesFormat: &network.SecurityGroupPropertiesFormat{
			SecurityRules: &[]network.SecurityRule{},
		},
		Location: to.StringPtr(t.Cloud.Region()),
		Name:     to.StringPtr(*e.Name),
		Tags:     e.Tags,
	}
	for _, nsr := range e.SecurityRules {
		securityRule := network.SecurityRule{
			Name: nsr.Name,
			SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
				Priority:                   nsr.Priority,
				Access:                     nsr.Access,
				Direction:                  nsr.Direction,
				Protocol:                   nsr.Protocol,
				SourceAddressPrefix:        nsr.SourceAddressPrefix,
				SourceAddressPrefixes:      nsr.SourceAddressPrefixes,
				SourcePortRange:            nsr.SourcePortRange,
				DestinationAddressPrefix:   nsr.DestinationAddressPrefix,
				DestinationAddressPrefixes: nsr.DestinationAddressPrefixes,
				DestinationPortRange:       nsr.DestinationPortRange,
			},
		}
		*p.SecurityRules = append(*p.SecurityRules, securityRule)
	}

	nsg, err := t.Cloud.NetworkSecurityGroup().CreateOrUpdate(
		context.TODO(),
		*e.ResourceGroup.Name,
		*e.Name,
		p)
	if err != nil {
		return err
	}

	e.ID = nsg.ID

	return nil
}

// NetworkSecurityRule represents a NetworkSecurityGroup rule.
type NetworkSecurityRule struct {
	Name                       *string
	Priority                   *int32
	Access                     network.SecurityRuleAccess
	Direction                  network.SecurityRuleDirection
	Protocol                   network.SecurityRuleProtocol
	SourceAddressPrefix        *string
	SourceAddressPrefixes      *[]string
	SourcePortRange            *string
	DestinationAddressPrefixes *[]string
	DestinationAddressPrefix   *string
	DestinationPortRange       *string
}

var _ fi.CloudupHasDependencies = &NetworkSecurityRule{}

func (e *NetworkSecurityRule) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	return nil
}
