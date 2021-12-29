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

package components

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// GCPPDCSIDriverOptionsBuilder adds options for the GCP PD CSI driver to the model
type GCPPDCSIDriverOptionsBuilder struct {
	*OptionsContext
}

var _ loader.OptionsBuilder = &GCPPDCSIDriverOptionsBuilder{}

func (b *GCPPDCSIDriverOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)
	if clusterSpec.GetCloudProvider() != kops.CloudProviderGCE {
		return nil
	}

	cc := clusterSpec.CloudConfig
	if cc.GCPPDCSIDriver == nil {
		cc.GCPPDCSIDriver = &kops.GCPPDCSIDriver{
			Enabled: fi.Bool(b.IsKubernetesGTE("1.23")),
		}
	}

	return nil
}
