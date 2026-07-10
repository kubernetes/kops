/*
Copyright 2026 The Kubernetes Authors.

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

// Package awsup holds the AWS clients used by nodeup.
//
// Unlike the cloudup awsup package, it constructs only the clients that
// nodeup needs, and it keeps them out of reach of reflection. nodeup
// contains reflect.Value.MethodByName call sites, so the linker keeps every
// exported method of any type that is convertible to an interface value via
// reflection; that property propagates from interface-stored types (such as
// NodeupModelContext) through all struct fields, transitively. An AWS SDK
// client reachable that way keeps every operation of its service in the
// binary, which is tens of megabytes for EC2 alone.
//
// Cloud is therefore only a handle: the SDK clients live in unexported
// package-level state, which is referenced from code but not from any type
// descriptor, so the linker can discard the unused SDK operations. Do not
// add SDK clients (or structs containing them) as fields on Cloud.
package awsup

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// ClientMaxRetries is the number of retries for AWS API calls, matching the cloudup awsup package.
const ClientMaxRetries = 13

// Cloud is a handle to the AWS clients, for the region we are running in.
type Cloud struct {
	region string
}

// clients holds the AWS clients used by nodeup. See the package comment for
// why they are package state rather than fields on Cloud.
var clients struct {
	mutex sync.Mutex

	region      string
	ec2         *ec2.Client
	autoscaling *autoscaling.Client

	machineTypes map[ec2types.InstanceType]*MachineTypeInfo
}

// NewCloud constructs the AWS clients used by nodeup.
func NewCloud(ctx context.Context, region string) (*Cloud, error) {
	clients.mutex.Lock()
	defer clients.mutex.Unlock()

	if clients.ec2 != nil {
		if clients.region != region {
			return nil, fmt.Errorf("attempt to build AWS clients for region %q, already built for %q", region, clients.region)
		}
		return &Cloud{region: region}, nil
	}

	cfg, err := loadAWSConfig(ctx, region)
	if err != nil {
		return nil, fmt.Errorf("failed to load default aws config: %w", err)
	}

	clients.region = region
	clients.ec2 = ec2.NewFromConfig(cfg)
	clients.autoscaling = autoscaling.NewFromConfig(cfg)

	return &Cloud{region: region}, nil
}

// Region returns the AWS region we are running in.
func (c *Cloud) Region() string {
	return c.region
}

// AssignIpv6Addresses calls ec2.AssignIpv6Addresses.
func (c *Cloud) AssignIpv6Addresses(ctx context.Context, input *ec2.AssignIpv6AddressesInput) (*ec2.AssignIpv6AddressesOutput, error) {
	return clients.ec2.AssignIpv6Addresses(ctx, input)
}

// DescribeLifecycleHooks calls autoscaling.DescribeLifecycleHooks.
func (c *Cloud) DescribeLifecycleHooks(ctx context.Context, input *autoscaling.DescribeLifecycleHooksInput) (*autoscaling.DescribeLifecycleHooksOutput, error) {
	return clients.autoscaling.DescribeLifecycleHooks(ctx, input)
}

// CompleteLifecycleAction calls autoscaling.CompleteLifecycleAction.
func (c *Cloud) CompleteLifecycleAction(ctx context.Context, input *autoscaling.CompleteLifecycleActionInput) (*autoscaling.CompleteLifecycleActionOutput, error) {
	return clients.autoscaling.CompleteLifecycleAction(ctx, input)
}

func loadAWSConfig(ctx context.Context, region string) (aws.Config, error) {
	loadOptions := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
		awsconfig.WithClientLogMode(aws.LogRetries),
		awsconfig.WithLogger(awsLogger{}),
		awsconfig.WithRetryer(func() aws.Retryer {
			return retry.NewAdaptiveMode(func(ao *retry.AdaptiveModeOptions) {
				ao.StandardOptions = append(ao.StandardOptions, func(so *retry.StandardOptions) {
					so.MaxAttempts = ClientMaxRetries
				})
			})
		}),
	}

	// assumes the role before executing commands
	roleARN := os.Getenv("KOPS_AWS_ROLE_ARN")
	if roleARN != "" {
		cfg, err := awsconfig.LoadDefaultConfig(ctx, loadOptions...)
		if err != nil {
			return aws.Config{}, fmt.Errorf("failed to load default aws config: %w", err)
		}
		stsClient := sts.NewFromConfig(cfg)
		assumeRoleProvider := stscreds.NewAssumeRoleProvider(stsClient, roleARN)

		loadOptions = append(loadOptions, awsconfig.WithCredentialsProvider(assumeRoleProvider))
	}

	return awsconfig.LoadDefaultConfig(ctx, loadOptions...)
}
