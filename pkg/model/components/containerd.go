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

	if fi.StringValue(clusterSpec.Containerd.Version) == "" {
		containerdVersion := ""

		if clusterSpec.ContainerRuntime == "containerd" {
			if clusterSpec.KubernetesVersion == "" {
				return fmt.Errorf("Kubernetes version is required")
			}

			sv, err := KubernetesVersion(clusterSpec)
			if err != nil {
				return fmt.Errorf("unable to determine kubernetes version from %q", clusterSpec.KubernetesVersion)
			}

			if sv.Major == 1 && sv.Minor >= 11 {
				// Containerd 1.2 is validated against Kubernetes v1.11+
				// https://github.com/containerd/containerd/blob/master/releases/v1.2.0.toml#L34
				containerdVersion = "1.2.10"
			} else {
				return fmt.Errorf("unknown version of kubernetes %q (cannot infer containerd version)", clusterSpec.KubernetesVersion)
			}

		} else if clusterSpec.ContainerRuntime == "docker" {
			if fi.StringValue(clusterSpec.Docker.Version) == "" {
				return fmt.Errorf("Docker version is required")
			}

			// Set containerd version for known Docker versions
			dockerVersion := fi.StringValue(clusterSpec.Docker.Version)
			switch dockerVersion {
			case "19.03.4":
				containerdVersion = "1.2.10"
			case "18.09.9":
				containerdVersion = "1.2.10"
			case "18.09.3":
				containerdVersion = "1.2.10"
			default:
				// Older version of docker
				containerd.SkipInstall = true
			}

		} else {
			// Unknown container runtime, should not install containerd
			containerd.SkipInstall = true
		}

		if containerdVersion != "" {
			containerd.Version = &containerdVersion
		}
	}

	// Apply global containerd defaults
	containerd.LogLevel = fi.String("warn")

	configFile := ""
	if clusterSpec.ContainerRuntime == "docker" {
		configFile += "disabled_plugins = [\"cri\"]\n"
	}
	containerd.ConfigFile = &configFile

	return nil
}
