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
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	network "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
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

	// ApplicationSecurityGroups are the ASGs referenced by security rules.
	// This field is used for dependency ordering and is not rendered to the cloud.
	ApplicationSecurityGroups []*ApplicationSecurityGroup

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
			found = v
			break
		}
	}
	if found == nil {
		return nil, nil
	}
	if found.Properties != nil && found.Properties.ProvisioningState != nil && *found.Properties.ProvisioningState == network.ProvisioningStateFailed {
		klog.Warningf("found network security group %q in failed provisioning state", *nsg.Name)
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
		// ApplicationSecurityGroups is for dependency ordering only and is not rendered to the cloud.
		ApplicationSecurityGroups: nsg.ApplicationSecurityGroups,
	}
	for _, rule := range found.Properties.SecurityRules {
		nsr := &NetworkSecurityRule{
			Name:                     rule.Name,
			Priority:                 rule.Properties.Priority,
			Access:                   *rule.Properties.Access,
			Direction:                *rule.Properties.Direction,
			Protocol:                 *rule.Properties.Protocol,
			SourceAddressPrefix:      rule.Properties.SourceAddressPrefix,
			SourcePortRange:          rule.Properties.SourcePortRange,
			DestinationAddressPrefix: rule.Properties.DestinationAddressPrefix,
			DestinationPortRange:     rule.Properties.DestinationPortRange,
		}
		if len(rule.Properties.SourceAddressPrefixes) > 0 {
			nsr.SourceAddressPrefixes = rule.Properties.SourceAddressPrefixes
		}
		if len(rule.Properties.SourceApplicationSecurityGroups) > 0 {
			var sasgs []*string
			for _, sasg := range rule.Properties.SourceApplicationSecurityGroups {
				asg, err := arm.ParseResourceID(*sasg.ID)
				if err != nil {
					return nil, err
				}
				sasgs = append(sasgs, to.Ptr(strings.ToLower(asg.Name)))
			}
			nsr.SourceApplicationSecurityGroupNames = sasgs
		}
		if len(rule.Properties.DestinationAddressPrefixes) > 0 {
			nsr.DestinationAddressPrefixes = rule.Properties.DestinationAddressPrefixes
		}
		if len(rule.Properties.DestinationApplicationSecurityGroups) > 0 {
			var dasgs []*string
			for _, dasg := range rule.Properties.DestinationApplicationSecurityGroups {
				asg, err := arm.ParseResourceID(*dasg.ID)
				if err != nil {
					return nil, err
				}
				dasgs = append(dasgs, to.Ptr(strings.ToLower(asg.Name)))
			}
			nsr.DestinationApplicationSecurityGroupNames = dasgs
		}
		actual.SecurityRules = append(actual.SecurityRules, nsr)
	}

	nsg.ID = found.ID

	return actual, nil
}

func (nsg *NetworkSecurityGroup) Normalize(c *fi.CloudupContext) error {
	c.T.Cloud.(azure.AzureCloud).AddClusterTags(nsg.Tags)
	if _, ok := c.Target.(*azure.AzureAPITarget); ok {
		if err := nsg.resolvePublicIPAddressSourcePrefixes(); err != nil {
			return err
		}
	}
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
		Properties: &network.SecurityGroupPropertiesFormat{
			SecurityRules: []*network.SecurityRule{},
		},
		Location: to.Ptr(t.Cloud.Region()),
		Name:     to.Ptr(*e.Name),
		Tags:     e.Tags,
	}
	for _, nsr := range e.SecurityRules {
		securityRule := network.SecurityRule{
			Name: nsr.Name,
			Properties: &network.SecurityRulePropertiesFormat{
				Priority:                   nsr.Priority,
				Access:                     &nsr.Access,
				Direction:                  &nsr.Direction,
				Protocol:                   &nsr.Protocol,
				SourceAddressPrefix:        nsr.SourceAddressPrefix,
				SourceAddressPrefixes:      nsr.SourceAddressPrefixes,
				SourcePortRange:            nsr.SourcePortRange,
				DestinationAddressPrefix:   nsr.DestinationAddressPrefix,
				DestinationAddressPrefixes: nsr.DestinationAddressPrefixes,
				DestinationPortRange:       nsr.DestinationPortRange,
			},
		}
		if nsr.SourceApplicationSecurityGroupNames != nil {
			var sasgs []*network.ApplicationSecurityGroup
			for _, name := range nsr.SourceApplicationSecurityGroupNames {
				id := azure.ApplicationSecurityGroupID{
					SubscriptionID:               t.Cloud.SubscriptionID(),
					ResourceGroupName:            *e.ResourceGroup.Name,
					ApplicationSecurityGroupName: *name,
				}
				idStr := id.String()
				sasgs = append(sasgs, &network.ApplicationSecurityGroup{ID: &idStr})
			}
			securityRule.Properties.SourceApplicationSecurityGroups = sasgs
		}
		if nsr.DestinationApplicationSecurityGroupNames != nil {
			var dasgs []*network.ApplicationSecurityGroup
			for _, name := range nsr.DestinationApplicationSecurityGroupNames {
				id := azure.ApplicationSecurityGroupID{
					SubscriptionID:               t.Cloud.SubscriptionID(),
					ResourceGroupName:            *e.ResourceGroup.Name,
					ApplicationSecurityGroupName: *name,
				}
				idStr := id.String()
				dasgs = append(dasgs, &network.ApplicationSecurityGroup{ID: &idStr})
			}
			securityRule.Properties.DestinationApplicationSecurityGroups = dasgs
		}
		p.Properties.SecurityRules = append(p.Properties.SecurityRules, &securityRule)
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

func (nsg *NetworkSecurityGroup) resolvePublicIPAddressSourcePrefixes() error {
	for _, rule := range nsg.SecurityRules {
		if err := rule.resolvePublicIPAddressSourcePrefix(); err != nil {
			return err
		}
	}
	return nil
}

// NetworkSecurityRule represents a NetworkSecurityGroup rule.
type NetworkSecurityRule struct {
	Name                                     *string
	Priority                                 *int32
	Access                                   network.SecurityRuleAccess
	Direction                                network.SecurityRuleDirection
	Protocol                                 network.SecurityRuleProtocol
	SourceAddressPrefix                      *string
	SourceAddressPrefixes                    []*string
	SourceApplicationSecurityGroupNames      []*string
	SourcePortRange                          *string
	DestinationAddressPrefixes               []*string
	DestinationAddressPrefix                 *string
	DestinationApplicationSecurityGroupNames []*string
	DestinationPortRange                     *string

	SourcePublicIPAddress *PublicIPAddress
}

var _ fi.CloudupHasDependencies = (*NetworkSecurityRule)(nil)

func (e *NetworkSecurityRule) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	if e.SourcePublicIPAddress == nil {
		return nil
	}
	return []fi.CloudupTask{e.SourcePublicIPAddress}
}

func (e *NetworkSecurityRule) resolvePublicIPAddressSourcePrefix() error {
	if e.SourcePublicIPAddress == nil {
		return nil
	}
	if e.SourceAddressPrefix != nil {
		return fmt.Errorf("cannot specify both SourceAddressPrefix and SourcePublicIPAddress for network security rule %q", fi.ValueOf(e.Name))
	}
	if len(e.SourceAddressPrefixes) != 0 {
		return fmt.Errorf("cannot specify both SourceAddressPrefixes and SourcePublicIPAddress for network security rule %q", fi.ValueOf(e.Name))
	}
	if e.SourcePublicIPAddress.IPAddress == nil || fi.ValueOf(e.SourcePublicIPAddress.IPAddress) == "" {
		return fmt.Errorf("public IP address %q has no assigned IP address for network security rule %q", fi.ValueOf(e.SourcePublicIPAddress.Name), fi.ValueOf(e.Name))
	}
	e.SourceAddressPrefixes = []*string{e.SourcePublicIPAddress.IPAddress}
	e.SourcePublicIPAddress = nil
	return nil
}
