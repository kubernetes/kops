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

	"k8s.io/klog"
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
		if b.IsKubernetesLT("1.11") {
			// Containerd 1.2 is validated against Kubernetes v1.11+
			// https://github.com/containerd/containerd/blob/master/releases/v1.2.0.toml#L34
			return fmt.Errorf("kubernetes %s is not compatible with containerd", clusterSpec.KubernetesVersion)
		} else if b.IsKubernetesLT("1.18") {
			klog.Warningf("kubernetes %s is untested with containerd", clusterSpec.KubernetesVersion)
		}

		// Set containerd based on Kubernetes version
		if fi.StringValue(containerd.Version) == "" {
			if b.IsKubernetesGTE("1.17") {
				containerd.Version = fi.String("1.3.2")
			} else if b.IsKubernetesGTE("1.11") {
				return fmt.Errorf("containerd version is required")
			}
		}

		// Apply defaults for containerd running in container runtime mode
		containerd.LogLevel = fi.String("warn")
		containerd.ConfigOverride = fi.String("")

	} else if clusterSpec.ContainerRuntime == "docker" {
		if fi.StringValue(containerd.Version) == "" {
			// Docker version should always be available
			if fi.StringValue(clusterSpec.Docker.Version) == "" {
				return fmt.Errorf("docker version is required")
			}

			// Set the containerd version for known Docker versions
			switch fi.StringValue(clusterSpec.Docker.Version) {
			case "19.03.4":
				containerd.Version = fi.String("1.2.10")
			case "18.09.9":
				containerd.Version = fi.String("1.2.10")
			case "18.09.3":
				containerd.Version = fi.String("1.2.4")
			default:
				// Old version of docker, single package
				containerd.SkipInstall = true
				return nil
			}
		}

		// Apply defaults for containerd running in Docker mode
		containerd.LogLevel = fi.String("warn")
		containerd.ConfigOverride = fi.String("disabled_plugins = [\"cri\"]\n")

	} else {
		// Unknown container runtime, should not install containerd
		containerd.SkipInstall = true
	}

	return nil
}
