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

package components

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// ContainerdOptionsBuilder adds options for containerd to the model
type ContainerdOptionsBuilder struct {
	*OptionsContext
}

var _ loader.ClusterOptionsBuilder = &ContainerdOptionsBuilder{}

// BuildOptions is responsible for filling in the default setting for containerd daemon
func (b *ContainerdOptionsBuilder) BuildOptions(o *kops.Cluster) error {
	clusterSpec := &o.Spec

	if clusterSpec.Containerd == nil {
		clusterSpec.Containerd = &kops.ContainerdConfig{}
	}

	containerd := clusterSpec.Containerd

	// Set version based on Kubernetes version
	if fi.ValueOf(containerd.Version) == "" {
		switch {
		case b.IsKubernetesLT("1.27.2"):
			containerd.Version = fi.PtrTo("1.6.20")
			containerd.Runc = &kops.Runc{
				Version: fi.PtrTo("1.1.5"),
			}
		default:
			containerd.Version = fi.PtrTo("1.7.25")
			containerd.Runc = &kops.Runc{
				Version: fi.PtrTo("1.2.4"),
			}
		}
	}
	// Set default log level to INFO
	containerd.LogLevel = fi.PtrTo("info")

	if containerd.NvidiaGPU != nil && fi.ValueOf(containerd.NvidiaGPU.Enabled) {
		if containerd.NvidiaGPU.DriverPackage == "" {
			containerd.NvidiaGPU.DriverPackage = kops.NvidiaDefaultDriverPackage
		}

		if containerd.NvidiaGPU.ImageTag == "" {
			containerd.NvidiaGPU.ImageTag = kops.NvidiaDevicePluginImageTag
		}
	}

	return nil
}
