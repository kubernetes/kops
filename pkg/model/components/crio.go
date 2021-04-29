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
	"fmt"

	"github.com/blang/semver"
	"github.com/pelletier/go-toml"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// CrioOptionsBuilder adds options for crio to the model
type CrioOptionsBuilder struct {
	*OptionsContext
}

var _ loader.OptionsBuilder = &CrioOptionsBuilder{}

// BuildOptions is responsible for filling in the default setting for crio daemon
func (c *CrioOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)

	if clusterSpec.Crio == nil {
		clusterSpec.Crio = &kops.CrioConfig{}
	}

	crio := clusterSpec.Crio

	if clusterSpec.ContainerRuntime == "crio" {
		if c.IsKubernetesLT("1.21") {
			return fmt.Errorf("kOps does not support crio with k8s version less than 1.21")
		}

		// Unless user specifies version, use the cri-o version matching the cluster's kubernetes version
		if crio.Version == nil {
			crioVersion, err := c.getCrioVersion(clusterSpec.KubernetesVersion)
			if err != nil {
				return err
			}

			crio.Version = crioVersion
		}

		if fi.StringValue(crio.ConfigOverride) == "" {
			crio.ConfigOverride = c.buildCrioDefaultConfig()
		}

		if fi.StringValue(crio.ContainerPolicyOverride) == "" {
			crio.ContainerPolicyOverride = c.buildCrioDefaultContainerPolicy()
		}

		if fi.StringValue(crio.ContainerRegistriesOverride) == "" {
			crio.ContainerRegistriesOverride = c.buildCrioDefaultContainerRegistriesPolicy()
		}

		if crio.LogLevel == nil {
			crio.LogLevel = fi.String("info")
		}

	} else {
		crio.SkipInstall = true
	}

	return nil
}

func (c *CrioOptionsBuilder) buildCrioDefaultConfig() *string {
	config, _ := toml.Load(``)

	config.SetPath([]string{"crio", "api", "listen"}, "/run/crio/crio.sock")

	config.SetPath([]string{"crio", "runtime", "selinux"}, true)
	config.SetPath([]string{"crio", "runtime", "hooks_dir"}, []string{"/usr/share/containers/oci/hooks.d"})
	config.SetPath([]string{"crio", "runtime", "log_size_max"}, int64(8192))
	config.SetPath([]string{"crio", "runtime", "container_exits_dir"}, "/run/crio/exits")
	config.SetPath([]string{"crio", "runtime", "container_attach_socket_dir"}, "/run/crio")
	config.SetPath([]string{"crio", "runtime", "namespaces_dir"}, "/run")

	config.SetPath([]string{"crio", "runtime", "runtimes", "runc", "runtime_path"}, "")
	config.SetPath([]string{"crio", "runtime", "runtimes", "runc", "runtime_type"}, "oci")
	config.SetPath([]string{"crio", "runtime", "runtimes", "runc", "runtime_root"}, "/run/runc")

	config.SetPath([]string{"crio", "metrics", "enable_metrics"}, true)
	config.SetPath([]string{"crio", "metrics", "metrics_port"}, int64(9090))

	return fi.String(config.String())
}

func (c *CrioOptionsBuilder) buildCrioDefaultContainerPolicy() *string {
	return fi.String(`{ "default": [{ "type": "insecureAcceptAnything" }] }`)
}

func (c *CrioOptionsBuilder) buildCrioDefaultContainerRegistriesPolicy() *string {
	config, _ := toml.Load(``)

	config.SetPath([]string{"unqualified-search-registries"}, []string{"docker.io", "k8s.gcr.io", "quay.io"})

	return fi.String(config.String())
}

func (c *CrioOptionsBuilder) getCrioVersion(kubernetesVersion string) (*string, error) {
	sv, err := semver.ParseTolerant(kubernetesVersion)
	if err != nil {
		return nil, fmt.Errorf("Cannot parse kubernetes version %s", kubernetesVersion)
	}

	return fi.String(fmt.Sprintf("1.%d.0", sv.Minor)), nil
}
