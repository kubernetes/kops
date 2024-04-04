/*
Copyright 2017 The Kubernetes Authors.

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
	"k8s.io/kops/protokube/pkg/gossip"
)

type SeedProvider struct {
	ec2  ec2.DescribeInstancesAPIClient
	tags map[string]string
}

var _ gossip.SeedProvider = &SeedProvider{}

func (p *SeedProvider) GetSeeds() ([]string, error) {
	ctx := context.TODO()

	request := &ec2.DescribeInstancesInput{}
	for k, v := range p.tags {
		filter := ec2types.Filter{
			Name:   aws.String("tag:" + k),
			Values: []string{v},
		}
		request.Filters = append(request.Filters, filter)
	}
	request.Filters = append(request.Filters, ec2types.Filter{
		Name:   aws.String("instance-state-name"),
		Values: []string{"running", "pending"},
	})

	var seeds []string
	paginator := ec2.NewDescribeInstancesPaginator(p.ec2, request)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error querying for EC2 instances: %v", err)
		}
		for _, r := range page.Reservations {
			for _, i := range r.Instances {
				ip := aws.ToString(i.PrivateIpAddress)
				if ip != "" {
					seeds = append(seeds, ip)
				}
			}
		}
	}

	return seeds, nil
}

func NewSeedProvider(ec2 ec2.DescribeInstancesAPIClient, tags map[string]string) (*SeedProvider, error) {
	return &SeedProvider{
		ec2:  ec2,
		tags: tags,
	}, nil
}
