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

const (
	DefaultSandboxImage = "registry.k8s.io/pause:3.10.1"
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

	// Set the default version
	if fi.ValueOf(containerd.Version) == "" {
		// Stay on containerd 2.2.x rather than 2.3.x to avoid a sandbox-image
		// regression in 2.3 (https://github.com/containerd/containerd/issues/13529).
		containerd.Version = new("2.2.4")
		containerd.Runc = &kops.Runc{
			Version: new("1.3.5"),
		}
	}
	// Set the default log level to INFO
	containerd.LogLevel = new("info")

	// Set the sandbox image used to scope pod shared resources used by the pod's containers.
	if fi.ValueOf(containerd.SandboxImage) == "" {
		containerd.SandboxImage = new(b.AssetBuilder.RemapImage(DefaultSandboxImage))
	}

	if containerd.NvidiaGPU != nil && fi.ValueOf(containerd.NvidiaGPU.Enabled) {
		if containerd.NvidiaGPU.DriverPackage == "" {
			containerd.NvidiaGPU.DriverPackage = kops.NvidiaDefaultDriverPackage
		}

		if containerd.NvidiaGPU.DevicePluginImage == "" {
			containerd.NvidiaGPU.DevicePluginImage = kops.NvidiaDevicePluginImage
		}
	}

	if containerd.GVisor != nil && fi.ValueOf(containerd.GVisor.Enabled) {
		if containerd.GVisor.Platform == "" {
			containerd.GVisor.Platform = kops.GVisorDefaultPlatform
		}
	}

	return nil
}
