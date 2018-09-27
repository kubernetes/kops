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
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/credentials"
	"github.com/spotinst/spotinst-sdk-go/spotinst/log"
	"github.com/spotinst/spotinst-sdk-go/spotinst/session"
	kopsv "k8s.io/kops"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

// NewCloud returns Cloud instance for given ClusterSpec.
func NewCloud(cluster *kops.Cluster) (fi.Cloud, error) {
	glog.V(2).Info("Creating Spotinst cloud")

	cloudProvider := GuessCloudFromClusterSpec(&cluster.Spec)
	if cloudProvider == "" {
		return nil, fmt.Errorf("spotinst: unable to infer cloud provider from cluster spec")
	}

	return newCloud(cloudProvider, cluster, newService(newConfig()))
}

func newCloud(cloudProvider kops.CloudProviderID, cluster *kops.Cluster, svc elastigroup.Service) (fi.Cloud, error) {
	var cloud fi.Cloud
	var err error

	switch cloudProvider {
	case kops.CloudProviderAWS:
		cloud, err = newAWSCloud(cluster, svc)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("spotinst: unsupported cloud provider: %s", cloudProvider)
	}

	return cloud, nil
}

func newAWSCloud(cluster *kops.Cluster, svc elastigroup.Service) (fi.Cloud, error) {
	region, err := awsup.FindRegion(cluster)
	if err != nil {
		return nil, err
	}

	tags := map[string]string{
		awsup.TagClusterName: cluster.ObjectMeta.Name,
	}

	cloud, err := awsup.NewAWSCloud(region, tags)
	if err != nil {
		return nil, err
	}

	return &awsCloud{
		AWSCloud: cloud.(awsup.AWSCloud),
		svc:      svc.CloudProviderAWS(),
	}, nil
}

func newService(config *spotinst.Config) elastigroup.Service {
	return elastigroup.New(session.New(config))
}

func newConfig() *spotinst.Config {
	config := spotinst.DefaultConfig()
	config.WithCredentials(newChainCredentials())
	config.WithUserAgent("Kubernetes-Kops/" + kopsv.Version)
	config.WithLogger(newStdLogger())

	return config
}

func newChainCredentials() *credentials.Credentials {
	return credentials.NewChainCredentials(
		new(credentials.EnvProvider),
		new(credentials.FileProvider),
	)
}

func newStdLogger() log.Logger {
	return log.LoggerFunc(func(format string, args ...interface{}) {
		glog.V(2).Infof(format, args...)
	})
}
