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

package v1alpha2

// ContainerdConfig is the configuration for containerd
type ContainerdConfig struct {
	// Address of containerd's GRPC server (default "/run/containerd/containerd.sock")
	Address *string `json:"address,omitempty" flag:"address"`
	// Complete containerd config file provided by the user
	ConfigFile *string `json:"configFile,omitempty"`
	// Logging level [trace, debug, info, warn, error, fatal, panic] (default "warn")
	LogLevel *string `json:"logLevel,omitempty" flag:"log-level"`
	// Directory for persistent data (default "/var/lib/containerd")
	Root *string `json:"root,omitempty" flag:"root"`
	// Prevents kops from installing and modifying containerd in any way (default "false")
	SkipInstall bool `json:"skipInstall,omitempty"`
	// Directory for execution state files (default "/run/containerd")
	State *string `json:"state,omitempty" flag:"state"`
	// Consumed by nodeup and used to pick the containerd version
	Version *string `json:"version,omitempty"`
}
