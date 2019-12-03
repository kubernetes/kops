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

// DockerOptionsBuilder adds options for docker to the model
type DockerOptionsBuilder struct {
	*OptionsContext
}

var _ loader.OptionsBuilder = &DockerOptionsBuilder{}

// BuildOptions is responsible for filling in the default setting for docker daemon
func (b *DockerOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)

	sv, err := KubernetesVersion(clusterSpec)
	if err != nil {
		return fmt.Errorf("unable to determine kubernetes version from %q", clusterSpec.KubernetesVersion)
	}

	if clusterSpec.Docker == nil {
		clusterSpec.Docker = &kops.DockerConfig{}
	}

	docker := clusterSpec.Docker

	// Container runtime is not Docker, should not install
	if clusterSpec.ContainerRuntime != "docker" {
		docker.SkipInstall = true
	}

	if fi.StringValue(clusterSpec.Docker.Version) == "" {
		if clusterSpec.KubernetesVersion == "" {
			return fmt.Errorf("KubernetesVersion is required")
		}

		dockerVersion := ""
		if sv.Major == 1 && sv.Minor >= 17 {
			dockerVersion = "19.03.4"
		} else if sv.Major == 1 && sv.Minor >= 16 {
			dockerVersion = "18.09.9"
		} else if sv.Major == 1 && sv.Minor >= 12 {
			dockerVersion = "18.06.3"
		} else if sv.Major == 1 && sv.Minor >= 9 {
			dockerVersion = "17.03.2"
		} else if sv.Major == 1 && sv.Minor >= 8 {
			dockerVersion = "1.13.1"
		} else if sv.Major == 1 && sv.Minor >= 6 {
			dockerVersion = "1.12.6"
		} else if sv.Major == 1 && sv.Minor >= 5 {
			dockerVersion = "1.12.3"
		} else if sv.Major == 1 && sv.Minor <= 4 {
			dockerVersion = "1.11.2"
		}

		if dockerVersion == "" {
			return fmt.Errorf("unknown version of kubernetes %q (cannot infer docker version)", clusterSpec.KubernetesVersion)
		}

		clusterSpec.Docker.Version = &dockerVersion
	}

	if sv.Major == 1 && sv.Minor >= 6 {
		if len(clusterSpec.Docker.LogOpt) == 0 && clusterSpec.Docker.LogDriver == nil {
			// Use built-in docker logging, if not configured otherwise (by the user)
			logDriver := "json-file"
			clusterSpec.Docker.LogDriver = &logDriver
			clusterSpec.Docker.LogOpt = append(clusterSpec.Docker.LogOpt, "max-size=10m")
			clusterSpec.Docker.LogOpt = append(clusterSpec.Docker.LogOpt, "max-file=5")
		}
	}

	docker.LogLevel = fi.String("warn")
	docker.IPTables = fi.Bool(false)
	docker.IPMasq = fi.Bool(false)

	// Note the alternative syntax... with a comma nodeup will try each of the filesystems in turn
	if b.IsKubernetesGTE("1.11") {
		// TODO(justinsb): figure out whether to use overlay2 on AWS jessie:
		// The ContainerOS image now has docker configured to use overlay2 out-of-the-box
		// and it is an error to specify the flag twice.
		// But Jessie (still our default AWS image) isn't recommended by docker with overlay2
		// (though that may be a kernel issue, and we run a custom kernel on our default image)
		// But we still need to worry about users running generic AMIs (e.g. stock jessie)
		docker.Storage = fi.String("overlay2,overlay,aufs")
	} else {
		docker.Storage = fi.String("overlay,aufs")
	}

	networking := clusterSpec.Networking
	if networking == nil || networking.Classic != nil {
		klog.Warningf("using deprecated (classic) networking")
		docker.Bridge = fi.String("cbr0")
	}

	return nil
}
