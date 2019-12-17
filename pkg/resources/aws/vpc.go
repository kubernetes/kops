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
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"

	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

func DeleteVPC(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	klog.V(2).Infof("Deleting EC2 VPC %q", id)
	request := &ec2.DeleteVpcInput{
		VpcId: &id,
	}
	_, err := c.EC2().DeleteVpc(request)
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
	data["type"] = ec2.ResourceTypeVpc
	data["raw"] = r.Obj
	op.Dump.Resources = append(op.Dump.Resources, data)

	ec2VPC := r.Obj.(*ec2.Vpc)
	vpc := &resources.VPC{
		ID: aws.StringValue(ec2VPC.VpcId),
	}
	op.Dump.VPC = vpc

	return nil
}

func DescribeVPCs(cloud fi.Cloud, clusterName string) (map[string]*ec2.Vpc, error) {
	c := cloud.(awsup.AWSCloud)

	vpcs := make(map[string]*ec2.Vpc)
	klog.V(2).Info("Listing EC2 VPC")
	for _, filters := range buildEC2FiltersForCluster(clusterName) {
		request := &ec2.DescribeVpcsInput{
			Filters: filters,
		}
		response, err := c.EC2().DescribeVpcs(request)
		if err != nil {
			return nil, fmt.Errorf("error listing VPCs: %v", err)
		}

		for _, vpc := range response.Vpcs {
			vpcs[aws.StringValue(vpc.VpcId)] = vpc
		}
	}

	return vpcs, nil
}

func ListVPCs(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	vpcs, err := DescribeVPCs(cloud, clusterName)
	if err != nil {
		return nil, err
	}

	var resourceTrackers []*resources.Resource
	for _, v := range vpcs {
		vpcID := aws.StringValue(v.VpcId)

		resourceTracker := &resources.Resource{
			Name:    FindName(v.Tags),
			ID:      vpcID,
			Type:    ec2.ResourceTypeVpc,
			Deleter: DeleteVPC,
			Dumper:  DumpVPC,
			Obj:     v,
			Shared:  !HasOwnedTag(ec2.ResourceTypeVpc+":"+vpcID, v.Tags, clusterName),
		}

		var blocks []string
		blocks = append(blocks, "dhcp-options:"+aws.StringValue(v.DhcpOptionsId))

		resourceTracker.Blocks = blocks

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}
