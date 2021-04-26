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

package v1alpha2

// CrioConfig is the configuration for cri-o runtime
type CrioConfig struct {
	// ConfigOverride is the complete crio config file provided by the user
	// TODO: since now that we are using a TOML based file generator, we can afford to specify extra options to configure cri-o with less effort
	ConfigOverride *string `json:"configOverride,omitempty"`
	// ContainerPolicyOverride is the image signature verification policy provided by the user. It goes to /etc/containers/
	ContainerPolicyOverride *string `json:"containerPolicyOverride,omitempty"`
	// ContainerRegistriesOverride contains config for the /etc/containers/registries.conf
	ContainerRegistriesOverride *string `json:"containerRegistriesOverride,omitempty"`
	// LogLevel determines the level of logging by the the crio daemon (default "info")
	LogLevel *string `json:"logLevel,omitempty" flag:"log-level"`
	// SkipInstall prevents kops from installing and modifying crio in any way (default "false")
	SkipInstall bool `json:"skipInstall,omitempty"`
	// Packages overrides the URL and hash for the packages.
	Packages *PackagesConfig `json:"packages,omitempty"`
	// Version determines the version of crio daemon which will be installed.
	Version *string `json:"version,omitempty"`
}
