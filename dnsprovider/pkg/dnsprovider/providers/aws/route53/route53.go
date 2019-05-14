/*
Copyright 2016 The Kubernetes Authors.

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
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/golang/glog"
	"gopkg.in/gcfg.v1"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
)

const (
	ProviderName = "aws-route53"
)

// MaxBatchSize is used to limit the max size of resource record changesets
var MaxBatchSize = 900

// Config to override defaults
type Config struct {
	Global struct {
		Route53Region      string `gcfg:"route53-region"`
		Route53EndpointURL string `gcfg:"route53-endpoint-url"`
	}
}

func init() {
	dnsprovider.RegisterDnsProvider(ProviderName, func(config io.Reader) (dnsprovider.Interface, error) {
		return newRoute53(config)
	})
}

// route53HandlerLogger is a request handler for aws-sdk-go that logs route53 requests
func route53HandlerLogger(req *request.Request) {
	service := req.ClientInfo.ServiceName

	name := "?"
	if req.Operation != nil {
		name = req.Operation.Name
	}

	glog.V(4).Infof("AWS request: %s %s", service, name)
}

// newRoute53 creates a new instance of an AWS Route53 DNS Interface.
func newRoute53(config io.Reader) (*Interface, error) {
	// Connect to AWS Route53 - TODO: Do more sophisticated auth

	awsConfig := aws.NewConfig()

	// This avoids a confusing error message when we fail to get credentials
	// e.g. https://github.com/kubernetes/kops/issues/605
	awsConfig = awsConfig.WithCredentialsChainVerboseErrors(true)

	// read config in case of using route53-China
	if config != nil {
		var cfg Config
		if err := gcfg.ReadInto(&cfg, config); err != nil {
			glog.Errorf("Couldn't read config: %v", err)
			return nil, err
		}
		glog.V(4).Infof("[newRoute53] got config for route53 China, route53Region: %s, route53EndpointURL: %s", cfg.Global.Route53Region, cfg.Global.Route53EndpointURL)
		awsConfig.Region = &cfg.Global.Route53Region
		awsConfig.Endpoint = &cfg.Global.Route53EndpointURL
	}

	svc := route53.New(session.New(), awsConfig)

	// Add our handler that will log requests
	svc.Handlers.Sign.PushFrontNamed(request.NamedHandler{
		Name: "k8s/logger",
		Fn:   route53HandlerLogger,
	})

	return New(svc), nil
}
