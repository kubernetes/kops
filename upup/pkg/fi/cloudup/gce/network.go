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
	"encoding/binary"
	"fmt"
	"net"

	context "golang.org/x/net/context"
	compute "google.golang.org/api/compute/v0.beta"
	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
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
func PerformNetworkAssignments(c *kops.Cluster, cloudObj fi.Cloud) error {
	ctx := context.TODO()

	if UsesIPAliases(c) {
		return performNetworkAssignmentsIPAliases(ctx, c, cloudObj)
	}
	return nil
}

func performNetworkAssignmentsIPAliases(ctx context.Context, c *kops.Cluster, cloudObj fi.Cloud) error {
	if len(c.Spec.Subnets) != 1 {
		return fmt.Errorf("expected exactly one subnet with GCE IP Aliases")
	}
	nodeSubnet := &c.Spec.Subnets[0]

	if c.Spec.PodCIDR != "" && c.Spec.ServiceClusterIPRange != "" && nodeSubnet.CIDR != "" {
		return nil
	}

	networkName := c.Spec.NetworkID
	if networkName == "" {
		networkName = "default"
	}

	cloud := cloudObj.(GCECloud)

	var regions []*compute.Region
	if err := cloud.Compute().Regions.List(cloud.Project()).Pages(ctx, func(p *compute.RegionList) error {
		regions = append(regions, p.Items...)
		return nil
	}); err != nil {
		return fmt.Errorf("error listing Regions: %v", err)
	}

	network, err := cloud.Compute().Networks.Get(cloud.Project(), networkName).Do()
	if err != nil {
		return fmt.Errorf("error fetching network name %q: %v", networkName, err)
	}

	subnetURLs := make(map[string]bool)
	for _, subnet := range network.Subnetworks {
		subnetURLs[subnet] = true
	}

	klog.Infof("scanning regions for subnetwork CIDR allocations")

	var subnets []*compute.Subnetwork
	for _, r := range regions {
		if err := cloud.Compute().Subnetworks.List(cloud.Project(), r.Name).Pages(ctx, func(p *compute.SubnetworkList) error {
			subnets = append(subnets, p.Items...)
			return nil
		}); err != nil {
			return fmt.Errorf("error listing Subnetworks: %v", err)
		}
	}

	var used cidrMap
	for _, subnet := range subnets {
		if !subnetURLs[subnet.SelfLink] {
			continue
		}
		if err := used.MarkInUse(subnet.IpCidrRange); err != nil {
			return err
		}

		for _, s := range subnet.SecondaryIpRanges {
			if err := used.MarkInUse(s.IpCidrRange); err != nil {
				return err
			}
		}
	}

	// CIDRs should be in the RFC1918 range, but otherwise we have no constraints
	networkCIDR := "10.0.0.0/8"

	podCIDR, err := used.Allocate(networkCIDR, 14)
	if err != nil {
		return err
	}

	serviceCIDR, err := used.Allocate(networkCIDR, 20)
	if err != nil {
		return err
	}

	nodeCIDR, err := used.Allocate(networkCIDR, 20)
	if err != nil {
		return err
	}

	klog.Infof("Will use %s for Nodes, %s for Pods and %s for Services", nodeCIDR, podCIDR, serviceCIDR)

	nodeSubnet.CIDR = nodeCIDR
	c.Spec.PodCIDR = podCIDR
	c.Spec.ServiceClusterIPRange = serviceCIDR

	return nil
}

// cidrMap is a helper structure to allocate unused CIDRs
type cidrMap struct {
	used []net.IPNet
}

func (c *cidrMap) MarkInUse(s string) error {
	_, cidr, err := net.ParseCIDR(s)
	if err != nil {
		return fmt.Errorf("error parsing network cidr %q: %v", s, err)
	}
	c.used = append(c.used, *cidr)
	return nil
}

func (c *cidrMap) Allocate(from string, mask int) (string, error) {
	_, cidr, err := net.ParseCIDR(from)
	if err != nil {
		return "", fmt.Errorf("error parsing CIDR %q: %v", from, err)
	}

	i := *cidr
	i.Mask = net.CIDRMask(mask, 32)

	for {

		ip4 := i.IP.To4()
		if ip4 == nil {
			return "", fmt.Errorf("expected IPv4 address: %v", from)
		}

		// Note we increment first, so we won't ever use the first range (e.g. 10.0.0.0/n)
		n := binary.BigEndian.Uint32(ip4)
		n += 1 << uint(32-mask)
		binary.BigEndian.PutUint32(i.IP, n)

		if !cidrsOverlap(cidr, &i) {
			break
		}

		if !c.isInUse(&i) {
			if err := c.MarkInUse(i.String()); err != nil {
				return "", err
			}
			return i.String(), nil
		}
	}

	return "", fmt.Errorf("cannot allocate CIDR of size %d", mask)
}

func (c *cidrMap) isInUse(n *net.IPNet) bool {
	for i := range c.used {
		if cidrsOverlap(&c.used[i], n) {
			return true
		}
	}
	return false
}

// cidrsOverlap returns true if and only if the two CIDRs are non-disjoint
func cidrsOverlap(l, r *net.IPNet) bool {
	return l.Contains(r.IP) || r.Contains(l.IP)
}
