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
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog/v2"

	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

func DeleteENI(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(awsup.AWSCloud)

	id := r.ID

	klog.V(2).Infof("Deleting EC2 ENI %q", id)
	request := &ec2.DeleteNetworkInterfaceInput{
		NetworkInterfaceId: &id,
	}
	_, err := c.EC2().DeleteNetworkInterface(request)
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
	data["type"] = ec2.ResourceTypeNetworkInterface
	data["raw"] = r.Obj

	op.Dump.Resources = append(op.Dump.Resources, data)

	return nil
}

func DescribeENIs(cloud fi.Cloud, clusterName string) (map[string]*ec2.NetworkInterface, error) {
	c := cloud.(awsup.AWSCloud)

	enis := make(map[string]*ec2.NetworkInterface)
	klog.V(2).Info("Listing ENIs")
	for _, filters := range buildEC2FiltersForCluster(clusterName) {
		request := &ec2.DescribeNetworkInterfacesInput{
			Filters: filters,
		}
		response, err := c.EC2().DescribeNetworkInterfaces(request)
		if err != nil {
			return nil, fmt.Errorf("error listing ENIs: %v", err)
		}

		for _, eni := range response.NetworkInterfaces {
			// Skip ENIs that are attached
			if eni.Attachment != nil {
				continue
			}
			enis[aws.StringValue(eni.NetworkInterfaceId)] = eni
		}
	}

	return enis, nil
}

func ListENIs(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	enis, err := DescribeENIs(cloud, clusterName)
	if err != nil {
		return nil, err
	}

	var resourceTrackers []*resources.Resource
	for _, v := range enis {
		eniID := aws.StringValue(v.NetworkInterfaceId)

		resourceTracker := &resources.Resource{
			ID:      eniID,
			Type:    ec2.ResourceTypeNetworkInterface,
			Deleter: DeleteENI,
			Dumper:  DumpENI,
			Obj:     v,
			Shared:  !HasOwnedTag(ec2.ResourceTypeNetworkInterface+":"+eniID, v.TagSet, clusterName),
		}

		var blocks []string
		blocks = append(blocks, ec2.ResourceTypeVpc+":"+aws.StringValue(v.VpcId))

		resourceTracker.Blocks = blocks

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}
