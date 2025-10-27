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

var _ loader.ClusterOptionsBuilder = &GCPPDCSIDriverOptionsBuilder{}

func (b *GCPPDCSIDriverOptionsBuilder) BuildOptions(o *kops.Cluster) error {
	gce := o.Spec.CloudProvider.GCE
	if gce == nil {
		return nil
	}

	if gce.PDCSIDriver == nil {
		gce.PDCSIDriver = &kops.PDCSIDriver{
			Enabled:                 fi.PtrTo(true),
			DefaultStorageClassName: fi.PtrTo("balanced-csi"),
			Version:                 fi.PtrTo("v1.21.6"),
		}
	}

	return nil
}
