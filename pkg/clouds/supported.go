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
	"fmt"
	"os"
	"strings"

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

	return clouds
}

// GuessCloudForPath tries to infer the cloud provider from a VFS path
func GuessCloudForPath(path string) (kops.CloudProviderID, error) {
	switch {
	case strings.HasPrefix(path, "azureblob://"):
		return kops.CloudProviderAzure, nil
	case strings.HasPrefix(path, "do://"):
		return kops.CloudProviderDO, nil
	case strings.HasPrefix(path, "gs://"):
		return kops.CloudProviderGCE, nil
	case strings.HasPrefix(path, "scw://"):
		return kops.CloudProviderScaleway, nil
	case strings.HasPrefix(path, "swift://"):
		return kops.CloudProviderOpenstack, nil
	case strings.HasPrefix(path, "s3://"):
		if os.Getenv("HCLOUD_TOKEN") != "" {
			return kops.CloudProviderHetzner, nil
		}
		return kops.CloudProviderAWS, nil
	default:
		return "", fmt.Errorf("cannot infer cloud provider from path: %q", path)
	}
}
