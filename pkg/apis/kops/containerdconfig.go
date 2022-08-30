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

package kops

// NvidiaDefaultDriverPackage is the nvidia driver default version
const NvidiaDefaultDriverPackage = "nvidia-headless-515-server"

// ContainerdConfig is the configuration for containerd
type ContainerdConfig struct {
	// Address of containerd's GRPC server (default "/run/containerd/containerd.sock").
	Address *string `json:"address,omitempty" flag:"address"`
	// ConfigOverride is the complete containerd config file provided by the user.
	ConfigOverride *string `json:"configOverride,omitempty"`
	// LogLevel controls the logging details [trace, debug, info, warn, error, fatal, panic] (default "info").
	LogLevel *string `json:"logLevel,omitempty" flag:"log-level"`
	// Packages overrides the URL and hash for the packages.
	Packages *PackagesConfig `json:"packages,omitempty"`
	// RegistryMirrors is list of image registries
	RegistryMirrors map[string][]string `json:"registryMirrors,omitempty"`
	// Root directory for persistent data (default "/var/lib/containerd").
	Root *string `json:"root,omitempty" flag:"root"`
	// SkipInstall prevents kOps from installing and modifying containerd in any way (default "false").
	SkipInstall bool `json:"skipInstall,omitempty"`
	// State directory for execution state files (default "/run/containerd").
	State *string `json:"state,omitempty" flag:"state"`
	// Version used to pick the containerd package.
	Version *string `json:"version,omitempty"`
	// NvidiaGPU configures the Nvidia GPU runtime.
	NvidiaGPU *NvidiaGPUConfig `json:"nvidiaGPU,omitempty"`
	// Runc configures the runc runtime.
	Runc *Runc `json:"runc,omitempty"`
}

type NvidiaGPUConfig struct {
	// Package is the name of the nvidia driver package that will be installed.
	// Default is "nvidia-headless-510-server".
	DriverPackage string `json:"package,omitempty"`
	// Enabled determines if kOps will install the Nvidia GPU runtime and drivers.
	// They will only be installed on intances that has an Nvidia GPU.
	Enabled *bool `json:"enabled,omitempty"`
	// DCGMExporterConfig configures the DCGM exporter
	DCGMExporter *DCGMExporterConfig `json:"dcgmExporter,omitempty"`
}

// DCGMExporterConfig configures the DCGM exporter.
// Only the DCGMExporterConfig in the cluster level takes effect. Configurations on the Instance Group are ignored.
type DCGMExporterConfig struct {
	// Enabled determines if kOps will install the DCGM exporter
	Enabled bool `json:"enabled,omitempty"`
}

type Runc struct {
	// Version used to pick the runc package.
	Version *string `json:"version,omitempty"`
	// Packages overrides the URL and hash for the packages.
	Packages *PackagesConfig `json:"packages,omitempty"`
}
