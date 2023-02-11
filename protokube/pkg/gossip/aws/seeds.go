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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/resolver"
	"k8s.io/kops/protokube/pkg/gossip"
)

type SeedProvider struct {
	ec2  ec2iface.EC2API
	tags map[string]string
}

var _ gossip.SeedProvider = &SeedProvider{}

func (p *SeedProvider) GetSeeds() ([]string, error) {
	return p.discover(context.TODO(), nil, nil)
}

func (p *SeedProvider) discover(ctx context.Context, hasTags []string, predicate func(*ec2.Instance) bool) ([]string, error) {
	request := &ec2.DescribeInstancesInput{}
	for k, v := range p.tags {
		filter := &ec2.Filter{
			Name:   aws.String("tag:" + k),
			Values: aws.StringSlice([]string{v}),
		}
		request.Filters = append(request.Filters, filter)
	}
	for _, k := range hasTags {
		filter := &ec2.Filter{
			Name:   aws.String("tag-key"),
			Values: aws.StringSlice([]string{k}),
		}
		request.Filters = append(request.Filters, filter)
	}
	request.Filters = append(request.Filters, &ec2.Filter{
		Name:   aws.String("instance-state-name"),
		Values: aws.StringSlice([]string{"running", "pending"}),
	})

	var seeds []string
	err := p.ec2.DescribeInstancesPagesWithContext(ctx, request, func(p *ec2.DescribeInstancesOutput, lastPage bool) (shouldContinue bool) {
		for _, r := range p.Reservations {
			for _, i := range r.Instances {
				if predicate != nil && !predicate(i) {
					continue
				}
				ip := aws.StringValue(i.PrivateIpAddress)
				if ip != "" {
					seeds = append(seeds, ip)
				}
			}
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("error querying for EC2 instances: %v", err)
	}

	return seeds, nil
}

func NewSeedProvider(ec2 ec2iface.EC2API, tags map[string]string) (*SeedProvider, error) {
	return &SeedProvider{
		ec2:  ec2,
		tags: tags,
	}, nil
}

var _ resolver.Resolver = &SeedProvider{}

// Resolve implements resolver.Resolve, providing name -> address resolution using cloud API discovery.
func (p *SeedProvider) Resolve(ctx context.Context, name string) ([]string, error) {
	klog.Infof("trying to resolve %q using SeedProvider", name)

	// We assume we are trying to resolve a component that runs on the control plane
	hasTags := []string{
		"k8s.io/role/control-plane",
	}
	return p.discover(ctx, hasTags, nil)
}
