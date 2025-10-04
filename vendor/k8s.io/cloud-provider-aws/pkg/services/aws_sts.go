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
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"k8s.io/klog/v2"
)

const headerSourceArn = "x-amz-source-arn"
const headerSourceAccount = "x-amz-source-account"

type withStsHeadersMiddleware struct {
	headers map[string]string
}

func (*withStsHeadersMiddleware) ID() string {
	return "withStsHeadersMiddleware"
}

func (m *withStsHeadersMiddleware) HandleBuild(ctx context.Context, in middleware.BuildInput, next middleware.BuildHandler) (
	out middleware.BuildOutput, metadata middleware.Metadata, err error,
) {
	req, ok := in.Request.(*smithyhttp.Request)
	if !ok {
		return out, metadata, fmt.Errorf("unrecognized transport type %T", in.Request)
	}

	for k, v := range m.headers {
		req.Header.Set(k, v)
	}
	return next.HandleBuild(ctx, in)
}

// WithStsHeadersMiddleware provides middleware to set custom headers for STS calls
func WithStsHeadersMiddleware(headers map[string]string) func(*sts.Options) {
	return func(o *sts.Options) {
		o.APIOptions = append(o.APIOptions, func(s *middleware.Stack) error {
			return s.Build.Add(&withStsHeadersMiddleware{
				headers: headers,
			}, middleware.After)
		})
	}
}

// NewStsClient provides a new STS client.
func NewStsClient(ctx context.Context, region, roleARN, sourceARN string) (*sts.Client, error) {
	klog.Infof("Using AWS assumed role %v", roleARN)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	parsedSourceArn, err := arn.Parse(roleARN)
	if err != nil {
		return nil, err
	}

	sourceAcct := parsedSourceArn.AccountID

	reqHeaders := map[string]string{
		headerSourceAccount: sourceAcct,
	}
	if sourceARN != "" {
		reqHeaders[headerSourceArn] = sourceARN
	}

	// Create the STS client with the custom middleware
	// svc := s3.NewFromConfig(cfg, WithHeader("x-user-header", "..."))
	stsClient := sts.NewFromConfig(cfg, func(o *sts.Options) {
		o.Region = region
	}, WithStsHeadersMiddleware(reqHeaders))

	klog.V(4).Infof("configuring STS client with extra headers, %v", reqHeaders)
	return stsClient, nil
}
