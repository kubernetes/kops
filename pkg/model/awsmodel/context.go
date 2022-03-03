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

package awsmodel

import (
	"fmt"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
)

// AWSModelContext provides the context for the aws model
type AWSModelContext struct {
	*model.KopsModelContext
}

func (b *AWSModelContext) LinkToSubnet(z *kops.ClusterSubnetSpec) *awstasks.Subnet {
	name := z.Name + "." + b.ClusterName()

	return &awstasks.Subnet{Name: &name}
}

func (b *AWSModelContext) LinkToPublicSubnetInZone(zoneName string) (*awstasks.Subnet, error) {
	var matches []*kops.ClusterSubnetSpec
	for i := range b.Cluster.Spec.Subnets {
		z := &b.Cluster.Spec.Subnets[i]
		if z.Zone != zoneName {
			continue
		}
		if z.Type != kops.SubnetTypePublic {
			continue
		}
		matches = append(matches, z)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("could not find public subnet in zone: %q", zoneName)
	}
	if len(matches) > 1 {
		// TODO: Support this (arbitrary choice I think, for ELBs)
		return nil, fmt.Errorf("found multiple public subnets in zone: %q", zoneName)
	}

	return b.LinkToSubnet(matches[0]), nil
}

func (b *AWSModelContext) LinkToUtilitySubnetInZone(zoneName string) (*awstasks.Subnet, error) {
	var matches []*kops.ClusterSubnetSpec
	for i := range b.Cluster.Spec.Subnets {
		s := &b.Cluster.Spec.Subnets[i]
		if s.Zone != zoneName {
			continue
		}
		if s.Type != kops.SubnetTypeUtility {
			continue
		}
		matches = append(matches, s)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("could not find utility subnet in zone: %q", zoneName)
	}
	if len(matches) > 1 {
		// TODO: Support this
		return nil, fmt.Errorf("found multiple utility subnets in zone: %q", zoneName)
	}

	return b.LinkToSubnet(matches[0]), nil
}
func (b *AWSModelContext) LinkToPrivateSubnetInZone(zoneName string) ([]*awstasks.Subnet, error) {
	var matches []*kops.ClusterSubnetSpec
	for i := range b.Cluster.Spec.Subnets {
		s := &b.Cluster.Spec.Subnets[i]
		if s.Zone != zoneName {
			continue
		}
		if s.Type != kops.SubnetTypePrivate {
			continue
		}
		matches = append(matches, s)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("could not find private subnet in zone: %q", zoneName)
	}

	var subnets []*awstasks.Subnet

	for _, match := range matches {
		subnets = append(subnets, b.LinkToSubnet(match))
	}

	return subnets, nil
}
