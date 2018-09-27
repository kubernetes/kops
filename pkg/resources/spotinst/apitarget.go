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
	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

// NewAPITarget returns APITarget instance for given Cloud.
func NewAPITarget(cluster *kops.Cluster, cloud fi.Cloud) fi.Target {
	glog.V(2).Info("Creating Spotinst target")

	cloudProvider := GuessCloudFromClusterSpec(&cluster.Spec)
	var target fi.Target

	switch cloudProvider {
	case kops.CloudProviderAWS:
		{
			glog.V(2).Infof("Cloud provider detected: %s", cloudProvider)
			target = NewAWSAPITarget(cloud.(*awsCloud))
		}
	}

	return target
}
