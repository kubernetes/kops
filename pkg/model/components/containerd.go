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
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/pelletier/go-toml"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// ContainerdOptionsBuilder adds options for containerd to the model
type ContainerdOptionsBuilder struct {
	*OptionsContext
}

var _ loader.OptionsBuilder = &ContainerdOptionsBuilder{}

// BuildOptions is responsible for filling in the default setting for containerd daemon
func (b *ContainerdOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)

	if clusterSpec.Containerd == nil {
		clusterSpec.Containerd = &kops.ContainerdConfig{}
	}

	containerd := clusterSpec.Containerd

	if clusterSpec.ContainerRuntime == "containerd" {
		// Set version based on Kubernetes version
		if fi.StringValue(containerd.Version) == "" {
			if b.IsKubernetesGTE("1.19") {
				containerd.Version = fi.String("1.4.11")
			} else {
				containerd.Version = fi.String("1.3.10")
			}
		}
		// Set default log level to INFO
		containerd.LogLevel = fi.String("info")
		// Build config file for containerd running in CRI mode
		if fi.StringValue(containerd.ConfigOverride) == "" {
			config, _ := toml.Load("")
			config.SetPath([]string{"version"}, int64(2))
			for name, endpoints := range containerd.RegistryMirrors {
				config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "registry", "mirrors", name, "endpoint"}, endpoints)
			}
			config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "runtimes", "runc", "runtime_type"}, "io.containerd.runc.v2")
			// only enable systemd cgroups for kubernetes >= 1.20
			config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "runtimes", "runc", "options", "SystemdCgroup"}, b.IsKubernetesGTE("1.20"))
			if UsesKubenet(clusterSpec.Networking) {
				// Using containerd with Kubenet requires special configuration.
				// This is a temporary backwards-compatible solution for kubenet users and will be deprecated when Kubenet is deprecated:
				// https://github.com/containerd/containerd/blob/master/docs/cri/config.md#cni-config-template
				config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "cni", "conf_template"}, "/etc/containerd/config-cni.template")
			}
			containerd.ConfigOverride = fi.String(config.String())
		}

	} else if clusterSpec.ContainerRuntime == "docker" {
		// Docker version should always be available
		dockerVersion := fi.StringValue(clusterSpec.Docker.Version)
		if dockerVersion == "" {
			return fmt.Errorf("docker version is required")
		} else {
			// Skip containerd setup for older versions without containerd service
			sv, err := semver.ParseTolerant(dockerVersion)
			if err != nil {
				return fmt.Errorf("unable to parse version string: %q", dockerVersion)
			}
			if sv.LT(semver.MustParse("18.9.0")) {
				containerd.SkipInstall = true
				return nil
			}
		}
		// Set default log level to INFO
		containerd.LogLevel = fi.String("info")
		// Build config file for containerd running in Docker mode
		config, _ := toml.Load("")
		config.SetPath([]string{"disabled_plugins"}, []string{"cri"})
		containerd.ConfigOverride = fi.String(config.String())

	} else {
		// Unknown container runtime, should not install containerd
		containerd.SkipInstall = true
	}

	return nil
}
