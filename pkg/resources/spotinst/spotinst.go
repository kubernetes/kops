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

package spotinst

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/spotinst/spotinst-sdk-go/service/elastigroup"
	"github.com/spotinst/spotinst-sdk-go/service/elastigroup/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/credentials"
	"github.com/spotinst/spotinst-sdk-go/spotinst/log"
	"github.com/spotinst/spotinst-sdk-go/spotinst/session"
	kopsv "k8s.io/kops"
	"k8s.io/kops/pkg/apis/kops"
)

// NewService returns a Service interface for the specified cloud provider.
func NewService(cloudProviderID kops.CloudProviderID) (Service, error) {
	svc := elastigroup.New(session.New(NewConfig()))

	switch cloudProviderID {
	case kops.CloudProviderAWS:
		return &awsService{svc.CloudProviderAWS()}, nil
	default:
		return nil, fmt.Errorf("spotinst: unsupported cloud provider: %s", cloudProviderID)
	}
}

// NewConfig returns a new configuration object.
func NewConfig() *spotinst.Config {
	config := spotinst.DefaultConfig()

	config.WithCredentials(NewCredentials())
	config.WithLogger(NewStdLogger())
	config.WithUserAgent("kubernetes-kops/" + kopsv.Version)

	return config
}

// NewCredentials returns a new chain-credentials object.
func NewCredentials() *credentials.Credentials {
	return credentials.NewChainCredentials(
		new(credentials.EnvProvider),
		new(credentials.FileProvider),
	)
}

// NewStdLogger returns a new Logger.
func NewStdLogger() log.Logger {
	return log.LoggerFunc(func(format string, args ...interface{}) {
		glog.V(2).Infof(format, args...)
	})
}

// NewElastigroup returns an Elastigroup wrapper for the specified cloud provider.
func NewElastigroup(cloudProviderID kops.CloudProviderID,
	obj interface{}) (Elastigroup, error) {

	switch cloudProviderID {
	case kops.CloudProviderAWS:
		return &awsElastigroup{obj.(*aws.Group)}, nil
	default:
		return nil, fmt.Errorf("spotinst: unsupported cloud provider: %s", cloudProviderID)
	}
}

// LoadCredentials attempts to load credentials using the default chain.
func LoadCredentials() (credentials.Value, error) {
	var (
		chain = NewCredentials()
		creds credentials.Value
		err   error
	)

	// Attempt to load the credentials.
	creds, err = chain.Get()
	if err != nil {
		return creds, fmt.Errorf("spotinst: unable to load credentials: %v", err)
	}

	return creds, nil
}
