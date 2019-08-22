/*
Copyright 2016 The Kubernetes Authors.

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

// DockerConfig is the configuration for docker
type DockerConfig struct {
	// AuthorizationPlugins is a list of authorization plugins
	AuthorizationPlugins []string `json:"authorizationPlugins,omitempty" flag:"authorization-plugin,repeat"`
	// Bridge is the network interface containers should bind onto
	Bridge *string `json:"bridge,omitempty" flag:"bridge"`
	// BridgeIP is a specific IP address and netmask for the docker0 bridge, using standard CIDR notation
	BridgeIP *string `json:"bridgeIP,omitempty" flag:"bip"`
	// DataRoot is the root directory of persistent docker state (default "/var/lib/docker")
	DataRoot *string `json:"dataRoot,omitempty" flag:"data-root"`
	// DefaultUlimit is the ulimits for containers
	DefaultUlimit []string `json:"defaultUlimit,omitempty" flag:"default-ulimit,repeat"`
	// ExecOpt is a series of options passed to the runtime
	ExecOpt []string `json:"execOpt,omitempty" flag:"exec-opt,repeat"`
	// ExecRoot is the root directory for execution state files (default "/var/run/docker")
	ExecRoot *string `json:"execRoot,omitempty" flag:"exec-root"`
	// Experimental features permits enabling new features such as dockerd metrics
	Experimental *bool `json:"experimental,omitempty" flag:"experimental"`
	// Hosts enables you to configure the endpoints the docker daemon listens on i.e. tcp://0.0.0.0.2375 or unix:///var/run/docker.sock etc
	Hosts []string `json:"hosts,omitempty" flag:"host,repeat"`
	// IPMasq enables ip masquerading for containers
	IPMasq *bool `json:"ipMasq,omitempty" flag:"ip-masq"`
	// IPtables enables addition of iptables rules
	IPTables *bool `json:"ipTables,omitempty" flag:"iptables"`
	// InsecureRegistry enable insecure registry communication @question according to dockers this a list??
	InsecureRegistry *string `json:"insecureRegistry,omitempty" flag:"insecure-registry"`
	// InsecureRegistries enables multiple insecure docker registry communications
	InsecureRegistries []string `json:"insecureRegistries,omitempty" flag:"insecure-registry"`
	// LiveRestore enables live restore of docker when containers are still running
	LiveRestore *bool `json:"liveRestore,omitempty" flag:"live-restore"`
	// LogDriver is the default driver for container logs (default "json-file")
	LogDriver *string `json:"logDriver,omitempty" flag:"log-driver"`
	// LogLevel is the logging level ("debug", "info", "warn", "error", "fatal") (default "info")
	LogLevel *string `json:"logLevel,omitempty" flag:"log-level"`
	// Logopt is a series of options given to the log driver options for containers
	LogOpt []string `json:"logOpt,omitempty" flag:"log-opt,repeat"`
	// Metrics address is the endpoint to serve with Prometheus format metrics
	MetricsAddress *string `json:"metricsAddress,omitempty" flag:"metrics-addr"`
	// MTU is the containers network MTU
	MTU *int32 `json:"mtu,omitempty" flag:"mtu"`
	// RegistryMirrors is a referred list of docker registry mirror
	RegistryMirrors []string `json:"registryMirrors,omitempty" flag:"registry-mirror,repeat"`
	// Storage is the docker storage driver to use
	Storage *string `json:"storage,omitempty" flag:"storage-driver"`
	// StorageOpts is a series of options passed to the storage driver
	StorageOpts []string `json:"storageOpts,omitempty" flag:"storage-opt,repeat"`
	// UserNamespaceRemap sets the user namespace remapping option for the docker daemon
	UserNamespaceRemap string `json:"userNamespaceRemap,omitempty" flag:"userns-remap"`
	// Version is consumed by the nodeup and used to pick the docker version
	Version *string `json:"version,omitempty"`
}
