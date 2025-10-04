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

package aws

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	stscredsv2 "github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/kms"

	smithymiddleware "github.com/aws/smithy-go/middleware"

	"k8s.io/client-go/pkg/version"
	"k8s.io/cloud-provider-aws/pkg/providers/v1/config"
	"k8s.io/cloud-provider-aws/pkg/providers/v1/iface"
	"k8s.io/klog/v2"
)

type awsSDKProvider struct {
	creds aws.CredentialsProvider
	cfg   awsCloudConfigProvider

	mutex          sync.Mutex
	regionDelayers map[string]*CrossRequestRetryDelay
}

func newAWSSDKProvider(creds aws.CredentialsProvider, cfg *config.CloudConfig) *awsSDKProvider {
	return &awsSDKProvider{
		creds:          creds,
		cfg:            cfg,
		regionDelayers: make(map[string]*CrossRequestRetryDelay),
	}
}

// Adds middleware to AWS SDK Go V2 clients.
func (p *awsSDKProvider) AddMiddleware(ctx context.Context, regionName string, cfg *aws.Config) {
	cfg.APIOptions = append(cfg.APIOptions,
		middleware.AddUserAgentKeyValue("kubernetes", version.Get().String()),
		func(stack *smithymiddleware.Stack) error {
			return stack.Finalize.Add(awsHandlerLoggerMiddleware(), smithymiddleware.Before)
		},
	)

	delayer := p.getCrossRequestRetryDelay(regionName)
	if delayer != nil {
		cfg.APIOptions = append(cfg.APIOptions,
			func(stack *smithymiddleware.Stack) error {
				stack.Finalize.Add(delayPreSign(delayer), smithymiddleware.Before)
				stack.Finalize.Insert(delayAfterRetry(delayer), "Retry", smithymiddleware.Before)
				return nil
			},
		)
	}

	p.addAPILoggingMiddleware(cfg)
}

// Adds logging middleware for AWS SDK Go V2 clients
func (p *awsSDKProvider) addAPILoggingMiddleware(cfg *aws.Config) {
	cfg.APIOptions = append(cfg.APIOptions,
		func(stack *smithymiddleware.Stack) error {
			stack.Serialize.Add(awsSendHandlerLoggerMiddleware(), smithymiddleware.After)
			stack.Deserialize.Add(awsValidateResponseHandlerLoggerMiddleware(), smithymiddleware.Before)
			return nil
		},
	)
}

// Get a CrossRequestRetryDelay, scoped to the region, not to the request.
// This means that when we hit a limit on a call, we will delay _all_ calls to the API.
// We do this to protect the AWS account from becoming overloaded and effectively locked.
// We also log when we hit request limits.
// Note that this delays the current goroutine; this is bad behaviour and will
// likely cause k8s to become slow or unresponsive for cloud operations.
// However, this throttle is intended only as a last resort.  When we observe
// this throttling, we need to address the root cause (e.g. add a delay to a
// controller retry loop)
func (p *awsSDKProvider) getCrossRequestRetryDelay(regionName string) *CrossRequestRetryDelay {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	delayer, found := p.regionDelayers[regionName]
	if !found {
		delayer = NewCrossRequestRetryDelay()
		p.regionDelayers[regionName] = delayer
	}
	return delayer
}

func (p *awsSDKProvider) Compute(ctx context.Context, regionName string, assumeRoleProvider *stscredsv2.AssumeRoleProvider) (iface.EC2, error) {
	cfg, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithDefaultsMode(aws.DefaultsModeInRegion),
		awsConfig.WithRegion(regionName),
	)
	if assumeRoleProvider != nil {
		cfg.Credentials = aws.NewCredentialsCache(assumeRoleProvider)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to initialize AWS config: %v", err)
	}

	p.AddMiddleware(ctx, regionName, &cfg)
	var opts []func(*ec2.Options) = p.cfg.GetEC2EndpointOpts(regionName)
	opts = append(opts, func(o *ec2.Options) {
		o.Retryer = &customRetryer{
			retry.NewStandard(),
		}
		o.EndpointResolverV2 = p.cfg.GetCustomEC2Resolver()
	})

	ec2Client := ec2.NewFromConfig(cfg, opts...)

	ec2 := &awsSdkEC2{
		ec2: ec2Client,
	}
	return ec2, nil
}

func (p *awsSDKProvider) LoadBalancing(ctx context.Context, regionName string, assumeRoleProvider *stscredsv2.AssumeRoleProvider) (ELB, error) {
	cfg, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithDefaultsMode(aws.DefaultsModeInRegion),
		awsConfig.WithRegion(regionName),
	)
	if assumeRoleProvider != nil {
		cfg.Credentials = aws.NewCredentialsCache(assumeRoleProvider)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to initialize AWS config: %v", err)
	}

	p.AddMiddleware(ctx, regionName, &cfg)
	var opts []func(*elb.Options) = p.cfg.GetELBEndpointOpts(regionName)
	opts = append(opts, func(o *elb.Options) {
		o.Retryer = &customRetryer{
			retry.NewStandard(),
		}
		o.EndpointResolverV2 = p.cfg.GetCustomELBResolver()
	})

	elbClient := elb.NewFromConfig(cfg, opts...)

	return elbClient, nil
}

func (p *awsSDKProvider) LoadBalancingV2(ctx context.Context, regionName string, assumeRoleProvider *stscredsv2.AssumeRoleProvider) (ELBV2, error) {
	cfg, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithDefaultsMode(aws.DefaultsModeInRegion),
		awsConfig.WithRegion(regionName),
	)
	if assumeRoleProvider != nil {
		cfg.Credentials = aws.NewCredentialsCache(assumeRoleProvider)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to initialize AWS config: %v", err)
	}

	p.AddMiddleware(ctx, regionName, &cfg)
	var opts []func(*elbv2.Options) = p.cfg.GetELBV2EndpointOpts(regionName)
	opts = append(opts, func(o *elbv2.Options) {
		o.Retryer = &customRetryer{
			retry.NewStandard(),
		}
		o.EndpointResolverV2 = p.cfg.GetCustomELBV2Resolver()
	})

	elbv2Client := elbv2.NewFromConfig(cfg, opts...)

	return elbv2Client, nil
}

func (p *awsSDKProvider) Metadata(ctx context.Context) (config.EC2Metadata, error) {
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithDefaultsMode(aws.DefaultsModeInRegion))
	if err != nil {
		return nil, fmt.Errorf("unable to initialize AWS config: %v", err)
	}

	p.addAPILoggingMiddleware(&cfg)

	// Unlike other SDK clients, the IMDS client does not support signing, so any overrides of the signing region and name
	// from awsSDKProvider.cfg will not be recognized.
	// Standard SDK clients use SigV4: https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_sigv-create-signed-request.html
	// But IMDS uses a different request pattern: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/configuring-instance-metadata-service.html
	var opts []func(*imds.Options) = p.cfg.GetIMDSEndpointOpts()
	imdsClient := imds.NewFromConfig(cfg, opts...)
	// opts = append(opts, func(o *imds.Options) {
	// 	o.ClientEnableState = imds.ClientEnabled
	// })

	getInstanceIdentityDocumentOutput, err := imdsClient.GetInstanceIdentityDocument(ctx, &imds.GetInstanceIdentityDocumentInput{})
	if err == nil {
		identity := getInstanceIdentityDocumentOutput.InstanceIdentityDocument
		klog.InfoS("instance metadata identity",
			"region", identity.Region,
			"availability-zone", identity.AvailabilityZone,
			"instance-type", identity.InstanceType,
			"architecture", identity.Architecture,
			"instance-id", identity.InstanceID,
			"private-ip", identity.PrivateIP,
			"account-id", identity.AccountID,
			"image-id", identity.ImageID)
	}
	return imdsClient, nil
}

func (p *awsSDKProvider) KeyManagement(ctx context.Context, regionName string, assumeRoleProvider *stscredsv2.AssumeRoleProvider) (KMS, error) {
	cfg, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithDefaultsMode(aws.DefaultsModeInRegion),
		awsConfig.WithRegion(regionName),
	)
	if assumeRoleProvider != nil {
		cfg.Credentials = aws.NewCredentialsCache(assumeRoleProvider)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to initialize AWS config: %v", err)
	}

	p.AddMiddleware(ctx, regionName, &cfg)
	var opts []func(*kms.Options) = p.cfg.GetKMSEndpointOpts(regionName)
	opts = append(opts, func(o *kms.Options) {
		o.Retryer = &customRetryer{
			retry.NewStandard(),
		}
		o.EndpointResolverV2 = p.cfg.GetCustomKMSResolver()
	})

	kmsClient := kms.NewFromConfig(cfg, opts...)

	return kmsClient, nil
}
