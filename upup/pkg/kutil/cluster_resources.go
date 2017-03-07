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

package kutil

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"strings"
)

// Todo
// After this is stable, we should pull this out of kutil so we can start mapping out the implementation
// for each cloud. Just taking the first steps now into separating these out, but we probably have some
// bigger concerns as this grows.. Just starting to draw some boundaries.

// ClusterResources is a representation of a cluster with abilities to ListResources and DeleteResources
type ClusterResources interface {
	ListResources() (map[string]*ResourceTracker, error)
	DeleteResources(resources map[string]*ResourceTracker) error
}

// AwsCluster is an implementation of ClusterResources
// The algorithm is pretty simple: it discovers all the resources it can (primary using tags)
// There are a few tweaks to that approach, like choosing a default ordering, but it is not much
// smarter.
// Some dependencies are invisible (e.g. ELB dependencies).
//
type AwsCluster struct {
	ClusterName string
	Cloud       fi.Cloud
	Region      string
}

func (c *AwsCluster) ListResources() (map[string]*ResourceTracker, error) {
	switch c.Cloud.ProviderID() {
	case fi.CloudProviderAWS:
		return c.listResourcesAWS()
	case fi.CloudProviderGCE:
		return c.listResourcesGCE()
	default:
		return nil, fmt.Errorf("Delete on clusters on %q not (yet) supported", c.Cloud.ProviderID())
	}
}

func (c *AwsCluster) listResourcesAWS() (map[string]*ResourceTracker, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	resources := make(map[string]*ResourceTracker)

	// These are the functions that are used for looking up
	// cluster resources by their tags.
	listFunctions := []listFn{

		// CloudFormation
		//ListCloudFormationStacks,

		// EC2
		ListInstances,
		ListKeypairs,
		ListSecurityGroups,
		ListVolumes,
		// EC2 VPC
		ListDhcpOptions,
		ListInternetGateways,
		ListRouteTables,
		ListSubnets,
		ListVPCs,
		// ELBs
		ListELBs,
		// ASG
		ListAutoScalingGroups,

		// Route 53
		ListRoute53Records,
		// IAM
		ListIAMInstanceProfiles,
		ListIAMRoles,
	}
	for _, fn := range listFunctions {
		trackers, err := fn(cloud, c.ClusterName)
		if err != nil {
			return nil, err
		}
		for _, t := range trackers {
			resources[t.Type+":"+t.ID] = t
		}
	}

	{
		// Gateways weren't tagged in kube-up
		// If we are deleting the VPC, we should delete the attached gateway
		// (no real reason not to; easy to recreate; no real state etc)

		gateways, err := DescribeInternetGatewaysIgnoreTags(cloud)
		if err != nil {
			return nil, err
		}

		for _, igw := range gateways {
			for _, attachment := range igw.Attachments {
				vpcID := aws.StringValue(attachment.VpcId)
				igwID := aws.StringValue(igw.InternetGatewayId)
				if vpcID == "" || igwID == "" {
					continue
				}
				if resources["vpc:"+vpcID] != nil && resources["internet-gateway:"+igwID] == nil {
					resources["internet-gateway:"+igwID] = &ResourceTracker{
						Name:    FindName(igw.Tags),
						ID:      igwID,
						Type:    "internet-gateway",
						deleter: DeleteInternetGateway,
					}
				}
			}
		}
	}

	{
		// We delete a launch configuration if it is bound to one of the tagged security groups
		securityGroups := sets.NewString()
		for k := range resources {
			if !strings.HasPrefix(k, "security-group:") {
				continue
			}
			id := strings.TrimPrefix(k, "security-group:")
			securityGroups.Insert(id)
		}
		lcs, err := FindAutoScalingLaunchConfigurations(cloud, securityGroups)
		if err != nil {
			return nil, err
		}

		for _, t := range lcs {
			resources[t.Type+":"+t.ID] = t
		}
	}

	if err := addUntaggedRouteTables(cloud, c.ClusterName, resources); err != nil {
		return nil, err
	}

	{
		// We delete a NAT gateway if it is linked to our route table
		routeTableIds := sets.NewString()
		for k := range resources {
			if !strings.HasPrefix(k, ec2.ResourceTypeRouteTable+":") {
				continue
			}
			id := strings.TrimPrefix(k, ec2.ResourceTypeRouteTable+":")
			routeTableIds.Insert(id)
		}
		natGateways, err := FindNatGateways(cloud, routeTableIds)
		if err != nil {
			return nil, err
		}

		for _, t := range natGateways {
			resources[t.Type+":"+t.ID] = t
		}
	}

	for k, t := range resources {
		if t.done {
			delete(resources, k)
		}
	}
	return resources, nil
}
