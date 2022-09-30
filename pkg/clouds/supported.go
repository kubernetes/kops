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

package clouds

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
)

func SupportedClouds() []kops.CloudProviderID {
	clouds := []kops.CloudProviderID{
		kops.CloudProviderAWS,
		kops.CloudProviderDO,
		kops.CloudProviderGCE,
		kops.CloudProviderHetzner,
		kops.CloudProviderOpenstack,
	}
	if featureflag.Azure.Enabled() {
		clouds = append(clouds, kops.CloudProviderAzure)
	}
	if featureflag.Scaleway.Enabled() {
		clouds = append(clouds, kops.CloudProviderScaleway)
	}
	if featureflag.Yandex.Enabled() {
		clouds = append(clouds, kops.CloudProviderYandex)
	}

	return clouds
}
