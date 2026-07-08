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
// nodeup needs and exposes them as concrete types. nodeup contains
// reflect.Value.MethodByName call sites, which prevent the linker from
// pruning methods of any type stored in an interface; keeping the AWS SDK
// clients out of interface values keeps the thousands of unused SDK
// operations out of the nodeup binary.
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

// Cloud holds the AWS clients used by nodeup, for the region we are running in.
type Cloud struct {
	region string

	ec2         *ec2.Client
	autoscaling *autoscaling.Client

	machineTypesMutex sync.Mutex
	machineTypes      map[ec2types.InstanceType]*MachineTypeInfo
}

// NewCloud constructs the AWS clients used by nodeup.
func NewCloud(ctx context.Context, region string) (*Cloud, error) {
	cfg, err := loadAWSConfig(ctx, region)
	if err != nil {
		return nil, fmt.Errorf("failed to load default aws config: %w", err)
	}

	return &Cloud{
		region:      region,
		ec2:         ec2.NewFromConfig(cfg),
		autoscaling: autoscaling.NewFromConfig(cfg),
	}, nil
}

// Region returns the AWS region we are running in.
func (c *Cloud) Region() string {
	return c.region
}

// EC2 returns the EC2 client.
func (c *Cloud) EC2() *ec2.Client {
	return c.ec2
}

// Autoscaling returns the autoscaling client.
func (c *Cloud) Autoscaling() *autoscaling.Client {
	return c.autoscaling
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
