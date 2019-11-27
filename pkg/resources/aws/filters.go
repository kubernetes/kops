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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

// buildEc2FiltersForCluster returns the set of filters we must use to find all resources
func buildEC2FiltersForCluster(clusterName string) [][]*ec2.Filter {
	var filterSets [][]*ec2.Filter

	// TODO: We could look for tag-key on the old & new tags, and then post-filter (we do this in k/k cloudprovider)

	filterSets = append(filterSets, []*ec2.Filter{
		{Name: aws.String("tag:" + awsup.TagClusterName), Values: aws.StringSlice([]string{clusterName})},
	})

	filterSets = append(filterSets, []*ec2.Filter{
		{Name: aws.String("tag-key"), Values: aws.StringSlice([]string{"kubernetes.io/cluster/" + clusterName})},
	})

	return filterSets
}
