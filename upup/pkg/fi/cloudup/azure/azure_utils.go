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
	"fmt"
	"strings"
)

// ZoneToLocation extracts the location from a zone of the
// form <location>-<available-zone-number>..
func ZoneToLocation(zone string) (string, error) {
	l := strings.Split(zone, "-")
	if len(l) != 2 {
		return "", fmt.Errorf("invalid Azure zone: %q ", zone)
	}
	return l[0], nil
}

// ZoneToAvailabilityZoneNumber extracts the availability zone number from a zone of the
// form <location>-<available-zone-number>..
func ZoneToAvailabilityZoneNumber(zone string) (string, error) {
	l := strings.Split(zone, "-")
	if len(l) != 2 {
		return "", fmt.Errorf("invalid Azure zone: %q ", zone)
	}
	return l[1], nil
}

// SubnetID contains the resource ID/names required to construct a subnet ID.
type SubnetID struct {
	SubscriptionID     string
	ResourceGroupName  string
	VirtualNetworkName string
	SubnetName         string
}

// String returns the subnet ID in the path format.
func (s *SubnetID) String() string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/%s/subnets/%s",
		s.SubscriptionID,
		s.ResourceGroupName,
		s.VirtualNetworkName,
		s.SubnetName)
}

// ParseSubnetID parses a given subnet ID string and returns a SubnetID.
func ParseSubnetID(s string) (*SubnetID, error) {
	l := strings.Split(s, "/")
	if len(l) != 11 {
		return nil, fmt.Errorf("malformed format of subnet ID: %s, %d", s, len(l))
	}
	return &SubnetID{
		SubscriptionID:     l[2],
		ResourceGroupName:  l[4],
		VirtualNetworkName: l[8],
		SubnetName:         l[10],
	}, nil
}

// NetworkSecurityGroupID contains the resource ID/names required to construct a NetworkSecurityGroup ID.
type NetworkSecurityGroupID struct {
	SubscriptionID           string
	ResourceGroupName        string
	NetworkSecurityGroupName string
}

// String returns the NetworkSecurityGroup ID in the path format.
func (s *NetworkSecurityGroupID) String() string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/networkSecurityGroups/%s",
		s.SubscriptionID,
		s.ResourceGroupName,
		s.NetworkSecurityGroupName)
}

// ParseNetworkSecurityGroupID parses a given NetworkSecurityGroup ID string and returns a NetworkSecurityGroup ID.
func ParseNetworkSecurityGroupID(s string) (*NetworkSecurityGroupID, error) {
	l := strings.Split(s, "/")
	if len(l) != 9 {
		return nil, fmt.Errorf("malformed format of NetworkSecurityGroup ID: %s, %d", s, len(l))
	}
	return &NetworkSecurityGroupID{
		SubscriptionID:           l[2],
		ResourceGroupName:        l[4],
		NetworkSecurityGroupName: l[8],
	}, nil
}

// ApplicationSecurityGroupID contains the resource ID/names required to construct a ApplicationSecurityGroup ID.
type ApplicationSecurityGroupID struct {
	SubscriptionID               string
	ResourceGroupName            string
	ApplicationSecurityGroupName string
}

// String returns the ApplicationSecurityGroup ID in the path format.
func (s *ApplicationSecurityGroupID) String() string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/applicationSecurityGroups/%s",
		s.SubscriptionID,
		s.ResourceGroupName,
		s.ApplicationSecurityGroupName)
}

// ParseApplicationSecurityGroupID parses a given ApplicationSecurityGroup ID string and returns a ApplicationSecurityGroup ID.
func ParseApplicationSecurityGroupID(s string) (*ApplicationSecurityGroupID, error) {
	l := strings.Split(s, "/")
	if len(l) != 9 {
		return nil, fmt.Errorf("malformed format of ApplicationSecurityGroup ID: %s, %d", s, len(l))
	}
	return &ApplicationSecurityGroupID{
		SubscriptionID:               l[2],
		ResourceGroupName:            l[4],
		ApplicationSecurityGroupName: l[8],
	}, nil
}

// LoadBalancerID contains the resource ID/names required to construct a load balancer ID.
type LoadBalancerID struct {
	SubscriptionID    string
	ResourceGroupName string
	LoadBalancerName  string
}

// String returns the load balancer ID in the path format.
func (lb *LoadBalancerID) String() string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/loadbalancers/%s/backendAddressPools/LoadBalancerBackEnd",
		lb.SubscriptionID,
		lb.ResourceGroupName,
		lb.LoadBalancerName,
	)
}

// ParseLoadBalancerID parses a given load balancer ID string and returns a LoadBalancerID.
func ParseLoadBalancerID(lb string) (*LoadBalancerID, error) {
	l := strings.Split(lb, "/")
	if len(l) != 11 {
		return nil, fmt.Errorf("malformed format of loadbalancer ID: %s, %d", lb, len(l))
	}
	return &LoadBalancerID{
		SubscriptionID:    l[2],
		ResourceGroupName: l[4],
		LoadBalancerName:  l[8],
	}, nil
}

// PublicIPAddressID contains the resource ID/names required to construct a PublicIPAddress ID.
type PublicIPAddressID struct {
	SubscriptionID      string
	ResourceGroupName   string
	PublicIPAddressName string
}

// String returns the PublicIPAddress ID in the path format.
func (s *PublicIPAddressID) String() string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/publicIPAddresss/%s",
		s.SubscriptionID,
		s.ResourceGroupName,
		s.PublicIPAddressName)
}

// ParsePublicIPAddressID parses a given PublicIPAddress ID string and returns a PublicIPAddress ID.
func ParsePublicIPAddressID(s string) (*PublicIPAddressID, error) {
	l := strings.Split(s, "/")
	if len(l) != 9 {
		return nil, fmt.Errorf("malformed format of PublicIPAddress ID: %s, %d", s, len(l))
	}
	return &PublicIPAddressID{
		SubscriptionID:      l[2],
		ResourceGroupName:   l[4],
		PublicIPAddressName: l[8],
	}, nil
}
