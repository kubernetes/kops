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

package spotinst

import (
	"fmt"

	"github.com/spotinst/spotinst-sdk-go/service/elastigroup"
	awseg "github.com/spotinst/spotinst-sdk-go/service/elastigroup/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/service/ocean"
	awsoc "github.com/spotinst/spotinst-sdk-go/service/ocean/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/credentials"
	"github.com/spotinst/spotinst-sdk-go/spotinst/log"
	"github.com/spotinst/spotinst-sdk-go/spotinst/session"
	"k8s.io/klog"
	kopsv "k8s.io/kops"
	"k8s.io/kops/pkg/apis/kops"
)

// NewCloud returns a Cloud interface for the specified cloud provider.
func NewCloud(cloudProviderID kops.CloudProviderID) (Cloud, error) {
	var (
		cloud Cloud
		sess  = session.New(NewConfig())
		eg    = elastigroup.New(sess)
		oc    = ocean.New(sess)
	)

	switch cloudProviderID {
	case kops.CloudProviderAWS:
		cloud = &awsCloud{
			eg: &awsElastigroupService{eg.CloudProviderAWS()},
			oc: &awsOceanService{oc.CloudProviderAWS()},
			ls: &awsOceanLaunchSpecService{oc.CloudProviderAWS()},
		}
	default:
		return nil, fmt.Errorf("spotinst: unsupported cloud provider: %s", cloudProviderID)
	}

	return cloud, nil
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
		klog.V(2).Infof(format, args...)
	})
}

// NewInstanceGroups returns an InstanceGroup wrapper for the specified cloud provider.
func NewInstanceGroup(cloudProviderID kops.CloudProviderID,
	instanceGroupType InstanceGroupType, obj interface{}) (InstanceGroup, error) {

	switch cloudProviderID {
	case kops.CloudProviderAWS:
		{
			switch instanceGroupType {
			case InstanceGroupElastigroup:
				return &awsElastigroupInstanceGroup{obj.(*awseg.Group)}, nil
			case InstanceGroupOcean:
				return &awsOceanInstanceGroup{obj.(*awsoc.Cluster)}, nil
			default:
				return nil, fmt.Errorf("spotinst: unsupported instance group type: %s", instanceGroupType)
			}
		}
	default:
		return nil, fmt.Errorf("spotinst: unsupported cloud provider: %s", cloudProviderID)
	}
}

// NewElastigroup returns an Elastigroup wrapper for the specified cloud provider.
func NewElastigroup(cloudProviderID kops.CloudProviderID,
	obj interface{}) (InstanceGroup, error) {

	return NewInstanceGroup(
		cloudProviderID,
		InstanceGroupElastigroup,
		obj)
}

// NewOcean returns an Ocean wrapper for the specified cloud provider.
func NewOcean(cloudProviderID kops.CloudProviderID,
	obj interface{}) (InstanceGroup, error) {

	return NewInstanceGroup(
		cloudProviderID,
		InstanceGroupOcean,
		obj)
}

// NewLaunchSpec returns a LaunchSpec wrapper for the specified cloud provider.
func NewLaunchSpec(cloudProviderID kops.CloudProviderID, obj interface{}) (LaunchSpec, error) {
	switch cloudProviderID {
	case kops.CloudProviderAWS:
		return &awsOceanLaunchSpec{obj.(*awsoc.LaunchSpec)}, nil
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
