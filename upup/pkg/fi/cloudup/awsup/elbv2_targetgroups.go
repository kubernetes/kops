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
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"k8s.io/klog/v2"
)

type TargetGroupInfo struct {
	TargetGroup *elbv2.TargetGroup
	Tags        []*elbv2.Tag
	arn         string
}

// ARN returns the ARN of the load balancer.
func (i *TargetGroupInfo) ARN() string {
	return i.arn
}

// NameTag returns the value of the tag with the key "Name".
func (i *TargetGroupInfo) NameTag() string {
	s, _ := i.GetTag("Name")
	return s
}

// GetTag returns the value of the tag with the given key.
func (i *TargetGroupInfo) GetTag(key string) (string, bool) {
	for _, tag := range i.Tags {
		if aws.StringValue(tag.Key) == key {
			return aws.StringValue(tag.Value), true
		}
	}
	return "", false
}

func ListELBV2TargetGroups(ctx context.Context, cloud AWSCloud) ([]*TargetGroupInfo, error) {
	klog.V(2).Infof("Listing all target groups")

	request := &elbv2.DescribeTargetGroupsInput{}
	// ELBV2 DescribeTags has a limit of 20 names, so we set the page size here to 20 also
	request.PageSize = aws.Int64(20)

	byARN := make(map[string]*TargetGroupInfo)

	var errs []error
	err := cloud.ELBV2().DescribeTargetGroupsPagesWithContext(ctx, request, func(p *elbv2.DescribeTargetGroupsOutput, lastPage bool) bool {
		if len(p.TargetGroups) == 0 {
			return true
		}

		tagRequest := &elbv2.DescribeTagsInput{}

		for _, tg := range p.TargetGroups {
			arn := aws.StringValue(tg.TargetGroupArn)
			byARN[arn] = &TargetGroupInfo{TargetGroup: tg, arn: arn}

			tagRequest.ResourceArns = append(tagRequest.ResourceArns, tg.TargetGroupArn)
		}

		tagResponse, err := cloud.ELBV2().DescribeTagsWithContext(ctx, tagRequest)
		if err != nil {
			errs = append(errs, fmt.Errorf("listing ELB tags: %w", err))
			return false
		}

		for _, t := range tagResponse.TagDescriptions {
			arn := aws.StringValue(t.ResourceArn)

			info := byARN[arn]
			if info == nil {
				klog.Fatalf("found tag for load balancer we didn't ask for %q", arn)
			}

			info.Tags = append(info.Tags, t.Tags...)
		}

		return true
	})
	if err != nil {
		return nil, fmt.Errorf("listing ELB TargetGroups: %w", err)
	}
	if len(errs) != 0 {
		return nil, fmt.Errorf("listing ELB TargetGroups: %w", errors.Join(errs...))
	}

	cloudTags := cloud.Tags()

	var results []*TargetGroupInfo
	for _, v := range byARN {
		if !MatchesElbV2Tags(cloudTags, v.Tags) {
			continue
		}
		results = append(results, v)
	}
	return results, nil
}
