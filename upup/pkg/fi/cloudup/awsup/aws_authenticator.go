/*
Copyright 2020 The Kubernetes Authors.

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
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"k8s.io/kops/pkg/bootstrap"
)

const AWSAuthenticationTokenPrefix = "x-aws-sts "

type awsAuthenticator struct {
	sts *sts.Client
}

var _ bootstrap.Authenticator = &awsAuthenticator{}

// RegionFromMetadata returns the current region from the aws metdata
func RegionFromMetadata(ctx context.Context) (string, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load default aws config: %w", err)
	}
	metadata := imds.NewFromConfig(cfg)

	resp, err := metadata.GetRegion(ctx, &imds.GetRegionInput{})
	if err != nil {
		return "", fmt.Errorf("failed to get region from ec2 metadata: %w", err)
	}
	return resp.Region, nil
}

func NewAWSAuthenticator(ctx context.Context, region string) (bootstrap.Authenticator, error) {
	config, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}
	return &awsAuthenticator{
		sts: sts.NewFromConfig(config),
	}, nil
}

func (a *awsAuthenticator) CreateToken(body []byte) (string, error) {
	sha := sha256.Sum256(body)

	presignClient := sts.NewPresignClient(a.sts)

	// Ensure the signature is only valid for this particular body content.
	stsRequest, _ := presignClient.PresignGetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{}, func(po *sts.PresignOptions) {
		po.ClientOptions = append(po.ClientOptions, func(o *sts.Options) {
			o.APIOptions = append(o.APIOptions, smithyhttp.AddHeaderValue("X-Kops-Request-SHA", base64.RawStdEncoding.EncodeToString(sha[:])))
		})
	})

	headers, _ := json.Marshal(stsRequest.SignedHeader)
	return AWSAuthenticationTokenPrefix + base64.StdEncoding.EncodeToString(headers), nil
}
