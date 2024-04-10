/*
Copyright 2021 The Kubernetes Authors.

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
)

func DeleteENI(cloud fi.Cloud, r *resources.Resource) error {
	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	klog.V(2).Infof("Deleting EC2 ENI %q", id)
	request := &ec2.DeleteNetworkInterfaceInput{
		NetworkInterfaceId: &id,
	}
	_, err := c.EC2().DeleteNetworkInterface(ctx, request)
	if err != nil {
		if awsup.AWSErrorCode(err) == "InvalidNetworkInterfaceID.NotFound" {
			// Concurrently deleted
			return nil
		}

		if IsDependencyViolation(err) {
			return err
		}
		return fmt.Errorf("error deleting ENI %q: %v", id, err)
	}
	return nil
}

func DumpENI(op *resources.DumpOperation, r *resources.Resource) error {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["type"] = ec2types.ResourceTypeNetworkInterface
	data["raw"] = r.Obj

	op.Dump.Resources = append(op.Dump.Resources, data)

	return nil
}

func DescribeENIs(cloud fi.Cloud, vpcID, clusterName string) (map[string]ec2types.NetworkInterface, error) {
	if vpcID == "" {
		return nil, nil
	}

	ctx := context.TODO()
	c := cloud.(awsup.AWSCloud)

	vpcFilter := awsup.NewEC2Filter("vpc-id", vpcID)
	statusFilter := awsup.NewEC2Filter("status", string(ec2types.NetworkInterfaceStatusAvailable))
	enis := make(map[string]ec2types.NetworkInterface)
	klog.V(2).Info("Listing ENIs")
	for _, filters := range buildEC2FiltersForCluster(clusterName) {
		request := &ec2.DescribeNetworkInterfacesInput{
			Filters: append(filters, vpcFilter, statusFilter),
		}
		paginator := ec2.NewDescribeNetworkInterfacesPaginator(c.EC2(), request)
		for paginator.HasMorePages() {
			dnio, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, fmt.Errorf("error listing ENIs: %v", err)
			}
			for _, eni := range dnio.NetworkInterfaces {
				enis[aws.ToString(eni.NetworkInterfaceId)] = eni
			}
		}
	}

	return enis, nil
}

func ListENIs(cloud fi.Cloud, vpcID, clusterName string) ([]*resources.Resource, error) {
	enis, err := DescribeENIs(cloud, vpcID, clusterName)
	if err != nil {
		return nil, err
	}

	var resourceTrackers []*resources.Resource
	for _, v := range enis {
		eniID := aws.ToString(v.NetworkInterfaceId)

		resourceTracker := &resources.Resource{
			ID:      eniID,
			Type:    string(ec2types.ResourceTypeNetworkInterface),
			Deleter: DeleteENI,
			Dumper:  DumpENI,
			Obj:     v,
			Shared:  !HasOwnedTag(string(ec2types.ResourceTypeNetworkInterface)+":"+eniID, v.TagSet, clusterName),
		}

		var blocks []string
		blocks = append(blocks, string(ec2types.ResourceTypeVpc)+":"+aws.ToString(v.VpcId))

		resourceTracker.Blocks = blocks

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}
