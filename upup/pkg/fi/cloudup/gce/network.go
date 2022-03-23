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

package gce

import (
	"context"
	"fmt"
	"net"
	"strings"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/util/subnet"
	"k8s.io/kops/upup/pkg/fi"
)

// UsesIPAliases checks if the cluster uses IP aliases for network connectivity
func UsesIPAliases(c *kops.Cluster) bool {
	if c.Spec.Networking != nil && c.Spec.Networking.GCE != nil {
		return true
	}
	return false
}

// PerformNetworkAssignments assigns suitable pod and service assignments for GCE,
// in particular for IP alias support.
func PerformNetworkAssignments(ctx context.Context, c *kops.Cluster, cloudObj fi.Cloud) error {
	if UsesIPAliases(c) {
		return performNetworkAssignmentsIPAliases(ctx, c, cloudObj)
	} else {
		return performSubnetAssignments(ctx, c, cloudObj)
	}
}

// ParseNameAndProjectFromNetworkID will take in the GCE-flavored network ID,
// and return the project and name of the resource.  The permitted formats are
// "network-name", or "project-id/network-name".  Empty string is also accepted.
func ParseNameAndProjectFromNetworkID(networkID string) (string, string, error) {
	var name, project string
	if networkID == "" {
		return "", "", nil
	}
	name = networkID
	// If the network ID has a slash, then we take the part before the / as a project ID.
	// Otherwise, we assume the entire provided value is the network ID.
	if strings.Contains(name, "/") {
		nameParts := strings.Split(name, "/")
		if len(nameParts) > 2 {
			return "", "", fmt.Errorf("cannot parse network name %q as either project/network or network", name)
		}
		name = nameParts[1]
		project = nameParts[0]
	}
	return name, project, nil
}

func buildUsed(ctx context.Context, c *kops.Cluster, cloudObj fi.Cloud) (*subnet.CIDRMap, error) {
	networkName := c.Spec.NetworkID
	if networkName == "" {
		networkName = SafeClusterName(c.Name)
	}

	cloud := cloudObj.(GCECloud)
	networkName, projectName, err := ParseNameAndProjectFromNetworkID(networkName)
	if err != nil {
		return nil, err
	}
	if projectName == "" {
		projectName = cloud.Project()
	}

	network, err := cloud.Compute().Networks().Get(projectName, networkName)
	if err != nil {
		if IsNotFound(err) {
			network = nil
		} else {
			return nil, fmt.Errorf("error fetching network %q: %w", networkName, err)
		}
	}
	used := &subnet.CIDRMap{}

	if network == nil {
		return used, nil
	}

	subnetURLs := make(map[string]bool)
	for _, subnet := range network.Subnetworks {
		subnetURLs[subnet] = true
	}
	if len(subnetURLs) == 0 {
		return used, nil
	}

	klog.Infof("scanning regions for subnetwork CIDR allocations")

	regions := make(map[string]bool)
	for subnetURL := range subnetURLs {
		u, err := ParseGoogleCloudURL(subnetURL)
		if err != nil {
			return nil, fmt.Errorf("error parsing subnet url %q: %w", subnetURL, err)
		}
		regions[u.Region] = true
	}

	var subnets []*compute.Subnetwork
	for region := range regions {
		l, err := cloud.Compute().Subnetworks().List(ctx, cloud.Project(), region)
		if err != nil {
			return nil, fmt.Errorf("error listing Subnetworks in region %q: %w", region, err)
		}
		subnets = append(subnets, l...)
	}

	for _, subnet := range subnets {
		if !subnetURLs[subnet.SelfLink] {
			continue
		}
		if err := used.MarkInUse(subnet.IpCidrRange); err != nil {
			return nil, err
		}

		for _, s := range subnet.SecondaryIpRanges {
			if err := used.MarkInUse(s.IpCidrRange); err != nil {
				return nil, err
			}
		}
	}

	return used, nil
}

func performNetworkAssignmentsIPAliases(ctx context.Context, c *kops.Cluster, cloudObj fi.Cloud) error {
	if len(c.Spec.Subnets) != 1 {
		return fmt.Errorf("expected exactly one subnet with GCE IP Aliases")
	}
	nodeSubnet := &c.Spec.Subnets[0]

	if c.Spec.PodCIDR != "" && c.Spec.ServiceClusterIPRange != "" && nodeSubnet.CIDR != "" {
		return nil
	}

	used, err := buildUsed(ctx, c, cloudObj)
	if err != nil {
		return err
	}

	// CIDRs should be in the RFC1918 range, but otherwise we have no constraints
	networkCIDR := "10.0.0.0/8"

	podCIDR, err := used.Allocate(networkCIDR, net.CIDRMask(14, 32))
	if err != nil {
		return err
	}

	serviceCIDR, err := used.Allocate(networkCIDR, net.CIDRMask(20, 32))
	if err != nil {
		return err
	}

	nodeCIDR, err := used.Allocate(networkCIDR, net.CIDRMask(20, 32))
	if err != nil {
		return err
	}

	klog.Infof("Will use %v for Nodes, %v for Pods and %v for Services", nodeCIDR, podCIDR, serviceCIDR)

	nodeSubnet.CIDR = nodeCIDR.String()
	c.Spec.PodCIDR = podCIDR.String()
	c.Spec.ServiceClusterIPRange = serviceCIDR.String()

	return nil
}

func performSubnetAssignments(ctx context.Context, c *kops.Cluster, cloudObj fi.Cloud) error {
	needCIDR := 0
	for i := range c.Spec.Subnets {
		subnet := &c.Spec.Subnets[i]
		if subnet.ProviderID != "" {
			continue
		}
		if subnet.CIDR == "" {
			needCIDR++
		}
	}

	if needCIDR == 0 {
		return nil
	}

	used, err := buildUsed(ctx, c, cloudObj)
	if err != nil {
		return err
	}

	// CIDRs should be in the RFC1918 range, but otherwise we have no constraints
	networkCIDR := "10.0.0.0/8"

	for i := range c.Spec.Subnets {
		subnet := &c.Spec.Subnets[i]
		if subnet.ProviderID != "" {
			continue
		}
		if subnet.CIDR != "" {
			continue
		}

		subnetCIDR, err := used.Allocate(networkCIDR, net.CIDRMask(20, 32))
		if err != nil {
			return err
		}
		subnet.CIDR = subnetCIDR.String()

		klog.Infof("assigned %v to subnet %v", subnetCIDR, subnet.Name)
	}

	return nil
}
