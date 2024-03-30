/*
Copyright 2024 The Kubernetes Authors.

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

package awsup

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"k8s.io/klog/v2"
)

type LoadBalancerInfo struct {
	LoadBalancer elbv2types.LoadBalancer
	Tags         []elbv2types.Tag
	arn          string
}

// ARN returns the ARN of the load balancer.
func (i *LoadBalancerInfo) ARN() string {
	return i.arn
}

// NameTag returns the value of the tag with the key "Name".
func (i *LoadBalancerInfo) NameTag() string {
	s, _ := i.GetTag("Name")
	return s
}

// GetTag returns the value of the tag with the given key.
func (i *LoadBalancerInfo) GetTag(key string) (string, bool) {
	for _, tag := range i.Tags {
		if aws.ToString(tag.Key) == key {
			return aws.ToString(tag.Value), true
		}
	}
	return "", false
}

func ListELBV2LoadBalancers(ctx context.Context, cloud AWSCloud) ([]*LoadBalancerInfo, error) {
	// TODO: Any way around this?
	klog.V(2).Infof("Listing all NLBs for ListELBV2LoadBalancers")

	request := &elbv2.DescribeLoadBalancersInput{}
	// ELBV2 DescribeTags has a limit of 20 names, so we set the page size here to 20 also
	request.PageSize = aws.Int32(20)

	byARN := make(map[string]*LoadBalancerInfo)

	paginator := elbv2.NewDescribeLoadBalancersPaginator(cloud.ELBV2(), request)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing ELB LoadBalancers: %w", err)
		}
		if len(page.LoadBalancers) == 0 {
			break
		}

		tagRequest := &elbv2.DescribeTagsInput{}

		for _, elb := range page.LoadBalancers {
			arn := aws.ToString(elb.LoadBalancerArn)
			byARN[arn] = &LoadBalancerInfo{LoadBalancer: elb, arn: arn}

			// TODO: Any way to filter by cluster here?

			tagRequest.ResourceArns = append(tagRequest.ResourceArns, aws.ToString(elb.LoadBalancerArn))
		}

		tagResponse, err := cloud.ELBV2().DescribeTags(ctx, tagRequest)
		if err != nil {
			return nil, fmt.Errorf("listing ELB tags: %w", err)
		}

		for _, t := range tagResponse.TagDescriptions {
			arn := aws.ToString(t.ResourceArn)

			info := byARN[arn]
			if info == nil {
				klog.Fatalf("found tag for load balancer we didn't ask for %q", arn)
			}

			info.Tags = append(info.Tags, t.Tags...)
		}
	}

	cloudTags := cloud.Tags()

	var results []*LoadBalancerInfo
	for _, v := range byARN {
		if !MatchesElbV2Tags(cloudTags, v.Tags) {
			continue
		}
		results = append(results, v)
	}
	return results, nil
}

func MatchesElbV2Tags(tags map[string]string, actual []elbv2types.Tag) bool {
	for k, v := range tags {
		found := false
		for _, a := range actual {
			if aws.ToString(a.Key) == k {
				if aws.ToString(a.Value) == v {
					found = true
					break
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}
