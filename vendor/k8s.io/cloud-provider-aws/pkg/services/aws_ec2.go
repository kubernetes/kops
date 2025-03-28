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

package services

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// EC2ClientV2 is an interface to allow it to be mocked.
type EC2ClientV2 interface {
	DescribeInstanceTopology(ctx context.Context, params *ec2.DescribeInstanceTopologyInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstanceTopologyOutput, error)
}

// Ec2SdkV2 is an implementation of the EC2 v2 interface, backed by aws-sdk-go-v2
type Ec2SdkV2 interface {
	DescribeInstanceTopology(ctx context.Context, request *ec2.DescribeInstanceTopologyInput) ([]types.InstanceTopology, error)
}

// ec2SdkV2 is an implementation of the EC2 v2 interface, backed by aws-sdk-go-v2
type ec2SdkV2 struct {
	Ec2 EC2ClientV2
}

// NewEc2SdkV2 is a constructor for Ec2SdkV2 that creates a default EC2 client.
func NewEc2SdkV2(ctx context.Context, region string, assumeRoleProvider *stscreds.AssumeRoleProvider) (Ec2SdkV2, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	// Don't override the default creds if the assume role provider isn't set.
	if assumeRoleProvider != nil {
		cfg.Credentials = aws.NewCredentialsCache(assumeRoleProvider)
	}

	client := ec2.NewFromConfig(cfg, func(o *ec2.Options) {
		o.Region = region
	})

	return &ec2SdkV2{Ec2: client}, nil
}

// DescribeInstanceTopology paginates calls to EC2 DescribeInstanceTopology API.
func (s *ec2SdkV2) DescribeInstanceTopology(ctx context.Context, request *ec2.DescribeInstanceTopologyInput) ([]types.InstanceTopology, error) {
	var topologies []types.InstanceTopology

	paginator := ec2.NewDescribeInstanceTopologyPaginator(s.Ec2, request)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		topologies = append(topologies, output.Instances...)
	}

	return topologies, nil
}
