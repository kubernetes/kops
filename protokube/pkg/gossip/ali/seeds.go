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

package ali

import (
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"k8s.io/kops/protokube/pkg/gossip"
)

type SeedProvider struct {
	ecs    *ecs.Client
	region string
	tag    map[string]string
}

var _ gossip.SeedProvider = &SeedProvider{}

func (p *SeedProvider) GetSeeds() ([]string, error) {
	var seeds []string

	// We could query at most 50 instances at a time on Aliyun ECS
	maxPageSize := 50
	args := &ecs.DescribeInstancesArgs{
		// TODO: pending? starting?
		Status:   ecs.Running,
		RegionId: common.Region(p.region),
		Pagination: common.Pagination{
			PageNumber: 1,
			PageSize:   maxPageSize,
		},
		Tag: p.tag,
	}

	var instances []ecs.InstanceAttributesType
	for {
		resp, page, err := p.ecs.DescribeInstances(args)
		if err != nil {
			return nil, err
		}
		instances = append(instances, resp...)

		if page.NextPage() == nil {
			break
		}
		args.Pagination = *(page.NextPage())
	}

	for _, instance := range instances {
		// TODO: Multiple IP addresses?
		seeds = append(seeds, instance.VpcAttributes.PrivateIpAddress.IpAddress...)
	}

	return seeds, nil
}

func NewSeedProvider(c *ecs.Client, region string, tag map[string]string) (*SeedProvider, error) {
	return &SeedProvider{
		ecs:    c,
		region: region,
		tag:    tag,
	}, nil
}
