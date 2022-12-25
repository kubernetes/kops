/*
Copyright 2021 The Kubernetes Authors.

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

package components

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// AWSOptionsBuilder prepares settings related to the AWS cloud provider.
type AWSOptionsBuilder struct {
	OptionsContext *OptionsContext
}

var _ loader.OptionsBuilder = &AWSOptionsBuilder{}

func (b *AWSOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)
	aws := clusterSpec.CloudProvider.AWS
	if aws == nil {
		return nil
	}

	if clusterSpec.IsIPv6Only() && len(aws.NodeIPFamilies) == 0 {
		aws.NodeIPFamilies = []string{"ipv6", "ipv4"}
	}

	return nil
}
