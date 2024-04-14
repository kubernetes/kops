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

package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/util/pkg/maps"
)

func DeleteVPC(cloud fi.Cloud, r *resources.Resource) error {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	klog.V(2).Infof("Deleting EC2 VPC %q", id)
	request := &ec2.DeleteVpcInput{
		VpcId: &id,
	}
	_, err := c.EC2().DeleteVpc(ctx, request)
	if err != nil {
		if awsup.AWSErrorCode(err) == "InvalidVpcID.NotFound" {
			// Concurrently deleted
			return nil
		}

		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting VPC %q: %v", id, err)
	}
	return nil
}

func DumpVPC(op *resources.DumpOperation, r *resources.Resource) error {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["type"] = ec2types.ResourceTypeVpc
	data["raw"] = r.Obj
	op.Dump.Resources = append(op.Dump.Resources, data)

	ec2VPC := r.Obj.(*ec2types.Vpc)
	vpc := &resources.VPC{
		ID: aws.ToString(ec2VPC.VpcId),
	}
	op.Dump.VPC = vpc

	return nil
}

func DescribeVPC(cloud fi.Cloud, clusterName string) (*ec2types.Vpc, error) {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	vpcs := make(map[string]*ec2types.Vpc)
	klog.V(2).Info("Listing EC2 VPC")
	for _, filters := range buildEC2FiltersForCluster(clusterName) {
		request := &ec2.DescribeVpcsInput{
			Filters: filters,
		}
		response, err := c.EC2().DescribeVpcs(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("error listing VPCs: %v", err)
		}

		for _, vpc := range response.Vpcs {
			vpcs[aws.ToString(vpc.VpcId)] = &vpc
		}
	}

	switch len(vpcs) {
	case 0:
		return nil, nil
	case 1:
		return vpcs[maps.Keys(vpcs)[0]], nil
	default:
		return nil, fmt.Errorf("found multiple VPCs for cluster %q: %v", clusterName, maps.Keys(vpcs))
	}
}

func ListVPCs(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	vpc, err := DescribeVPC(cloud, clusterName)
	if err != nil {
		return nil, err
	}

	var resourceTrackers []*resources.Resource
	if vpc != nil {
		vpcID := aws.ToString(vpc.VpcId)

		resourceTracker := &resources.Resource{
			Name:    FindName(vpc.Tags),
			ID:      vpcID,
			Type:    string(ec2types.ResourceTypeVpc),
			Deleter: DeleteVPC,
			Dumper:  DumpVPC,
			Obj:     vpc,
			Shared:  !HasOwnedTag(string(ec2types.ResourceTypeVpc)+":"+vpcID, vpc.Tags, clusterName),
		}

		var blocks []string
		blocks = append(blocks, "dhcp-options:"+aws.ToString(vpc.DhcpOptionsId))

		resourceTracker.Blocks = blocks

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}
