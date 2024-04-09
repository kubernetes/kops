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

// route53 is the implementation of pkg/dnsprovider interface for AWS Route53
package route53

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/util/pkg/awslog"
)

const (
	ProviderName = "aws-route53"
)

// MaxBatchSize is used to limit the max size of resource record changesets
var MaxBatchSize = 900

func init() {
	dnsprovider.RegisterDNSProvider(ProviderName, func(config io.Reader) (dnsprovider.Interface, error) {
		return newRoute53()
	})
}

// newRoute53 creates a new instance of an AWS Route53 DNS Interface.
func newRoute53() (*Interface, error) {
	ctx := context.TODO()

	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithClientLogMode(aws.LogRetries),
		awslog.WithAWSLogger(),
		awsconfig.WithRetryer(func() aws.Retryer {
			return retry.AddWithMaxAttempts(retry.NewStandard(), 5)
		}))
	if err != nil {
		return nil, fmt.Errorf("failed to load default aws config: %w", err)
	}

	// AWS_REGION, IMDS, or config profiles can override this in LoadDefaultConfig above.
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}

	svc := route53.NewFromConfig(cfg)

	return New(svc), nil
}
