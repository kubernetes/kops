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
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/kms"
	"k8s.io/client-go/pkg/version"
	"k8s.io/klog/v2"

	"k8s.io/cloud-provider-aws/pkg/providers/v1/config"
	"k8s.io/cloud-provider-aws/pkg/providers/v1/iface"
)

type awsSDKProvider struct {
	creds *credentials.Credentials
	cfg   awsCloudConfigProvider

	mutex          sync.Mutex
	regionDelayers map[string]*CrossRequestRetryDelay
}

func newAWSSDKProvider(creds *credentials.Credentials, cfg *config.CloudConfig) *awsSDKProvider {
	return &awsSDKProvider{
		creds:          creds,
		cfg:            cfg,
		regionDelayers: make(map[string]*CrossRequestRetryDelay),
	}
}

func (p *awsSDKProvider) AddHandlers(regionName string, h *request.Handlers) {
	h.Build.PushFrontNamed(request.NamedHandler{
		Name: "k8s/user-agent",
		Fn:   request.MakeAddToUserAgentHandler("kubernetes", version.Get().String()),
	})

	h.Sign.PushFrontNamed(request.NamedHandler{
		Name: "k8s/logger",
		Fn:   awsHandlerLogger,
	})

	delayer := p.getCrossRequestRetryDelay(regionName)
	if delayer != nil {
		h.Sign.PushFrontNamed(request.NamedHandler{
			Name: "k8s/delay-presign",
			Fn:   delayer.BeforeSign,
		})

		h.AfterRetry.PushFrontNamed(request.NamedHandler{
			Name: "k8s/delay-afterretry",
			Fn:   delayer.AfterRetry,
		})
	}

	p.addAPILoggingHandlers(h)
}

func (p *awsSDKProvider) addAPILoggingHandlers(h *request.Handlers) {
	h.Send.PushBackNamed(request.NamedHandler{
		Name: "k8s/api-request",
		Fn:   awsSendHandlerLogger,
	})

	h.ValidateResponse.PushFrontNamed(request.NamedHandler{
		Name: "k8s/api-validate-response",
		Fn:   awsValidateResponseHandlerLogger,
	})
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

func (p *awsSDKProvider) Compute(regionName string) (iface.EC2, error) {
	awsConfig := &aws.Config{
		Region:      &regionName,
		Credentials: p.creds,
	}
	awsConfig = awsConfig.WithCredentialsChainVerboseErrors(true).
		WithEndpointResolver(p.cfg.GetResolver())
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            *awsConfig,
		SharedConfigState: session.SharedConfigEnable,
	})

	if err != nil {
		return nil, fmt.Errorf("unable to initialize AWS session: %v", err)
	}
	service := ec2.New(sess)

	p.AddHandlers(regionName, &service.Handlers)

	ec2 := &awsSdkEC2{
		ec2: service,
	}
	return ec2, nil
}

func (p *awsSDKProvider) LoadBalancing(regionName string) (ELB, error) {
	awsConfig := &aws.Config{
		Region:      &regionName,
		Credentials: p.creds,
	}
	awsConfig = awsConfig.WithCredentialsChainVerboseErrors(true).
		WithEndpointResolver(p.cfg.GetResolver())
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            *awsConfig,
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to initialize AWS session: %v", err)
	}
	elbClient := elb.New(sess)
	p.AddHandlers(regionName, &elbClient.Handlers)

	return elbClient, nil
}

func (p *awsSDKProvider) LoadBalancingV2(regionName string) (ELBV2, error) {
	awsConfig := &aws.Config{
		Region:      &regionName,
		Credentials: p.creds,
	}
	awsConfig = awsConfig.WithCredentialsChainVerboseErrors(true).
		WithEndpointResolver(p.cfg.GetResolver())
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            *awsConfig,
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to initialize AWS session: %v", err)
	}
	elbClient := elbv2.New(sess)

	p.AddHandlers(regionName, &elbClient.Handlers)

	return elbClient, nil
}

func (p *awsSDKProvider) Autoscaling(regionName string) (ASG, error) {
	awsConfig := &aws.Config{
		Region:      &regionName,
		Credentials: p.creds,
	}
	awsConfig = awsConfig.WithCredentialsChainVerboseErrors(true).
		WithEndpointResolver(p.cfg.GetResolver())
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            *awsConfig,
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to initialize AWS session: %v", err)
	}
	client := autoscaling.New(sess)

	p.AddHandlers(regionName, &client.Handlers)

	return client, nil
}

func (p *awsSDKProvider) Metadata() (config.EC2Metadata, error) {
	sess, err := session.NewSession(&aws.Config{
		EndpointResolver: p.cfg.GetResolver(),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to initialize AWS session: %v", err)
	}
	client := ec2metadata.New(sess)
	p.addAPILoggingHandlers(&client.Handlers)

	identity, err := client.GetInstanceIdentityDocument()
	if err != nil {
		return nil, fmt.Errorf("unable to get instance identity document: %v", err)
	}
	klog.InfoS("instance metadata identity",
		"region", identity.Region,
		"availability-zone", identity.AvailabilityZone,
		"instance-type", identity.InstanceType,
		"architecture", identity.Architecture,
		"instance-id", identity.InstanceID,
		"private-ip", identity.PrivateIP,
		"account-id", identity.AccountID,
		"image-id", identity.ImageID)
	return client, nil
}

func (p *awsSDKProvider) KeyManagement(regionName string) (KMS, error) {
	awsConfig := &aws.Config{
		Region:      &regionName,
		Credentials: p.creds,
	}
	awsConfig = awsConfig.WithCredentialsChainVerboseErrors(true).
		WithEndpointResolver(p.cfg.GetResolver())
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            *awsConfig,
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to initialize AWS session: %v", err)
	}
	kmsClient := kms.New(sess)

	p.AddHandlers(regionName, &kmsClient.Handlers)

	return kmsClient, nil
}
