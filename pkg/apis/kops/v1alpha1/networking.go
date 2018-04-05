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

package v1alpha1

// NetworkingSpec allows selection and configuration of a networking plugin
type NetworkingSpec struct {
	Classic    *ClassicNetworkingSpec    `json:"classic,omitempty"`
	Kubenet    *KubenetNetworkingSpec    `json:"kubenet,omitempty"`
	External   *ExternalNetworkingSpec   `json:"external,omitempty"`
	CNI        *CNINetworkingSpec        `json:"cni,omitempty"`
	Kopeio     *KopeioNetworkingSpec     `json:"kopeio,omitempty"`
	Weave      *WeaveNetworkingSpec      `json:"weave,omitempty"`
	Flannel    *FlannelNetworkingSpec    `json:"flannel,omitempty"`
	Calico     *CalicoNetworkingSpec     `json:"calico,omitempty"`
	Canal      *CanalNetworkingSpec      `json:"canal,omitempty"`
	Kuberouter *KuberouterNetworkingSpec `json:"kuberouter,omitempty"`
	Romana     *RomanaNetworkingSpec     `json:"romana,omitempty"`
	AmazonVPC  *AmazonVPCNetworkingSpec  `json:"amazonvpc,omitempty"`
	Cilium     *CiliumNetworkingSpec     `json:"cilium,omitempty"`
}

// ClassicNetworkingSpec is the specification of classic networking mode, integrated into kubernetes
type ClassicNetworkingSpec struct {
}

// KubenetNetworkingSpec is the specification for kubenet networking, largely integrated but intended to replace classic
type KubenetNetworkingSpec struct {
}

// ExternalNetworkingSpec is the specification for networking that is implemented by a Daemonset
// It also uses kubenet
type ExternalNetworkingSpec struct {
}

// CNINetworkingSpec is the specification for networking that is implemented by a Daemonset
// Networking is not managed by kops - we can create options here that directly configure e.g. weave
// but this is useful for arbitrary network modes or for modes that don't need additional configuration.
type CNINetworkingSpec struct {
}

// KopeioNetworkingSpec declares that we want Kopeio networking
type KopeioNetworkingSpec struct {
}

// WeaveNetworkingSpec declares that we want Weave networking
type WeaveNetworkingSpec struct {
	MTU       *int32 `json:"mtu,omitempty"`
	ConnLimit *int32 `json:"connLimit,omitempty"`
}

// FlannelNetworkingSpec declares that we want Flannel networking
type FlannelNetworkingSpec struct {
	// Backend is the backend overlay type we want to use (vxlan or udp)
	Backend string `json:"backend,omitempty"`
}

// CalicoNetworkingSpec declares that we want Calico networking
type CalicoNetworkingSpec struct {
	CrossSubnet bool `json:"crossSubnet,omitempty"` // Enables Calico's cross-subnet mode when set to true
	// LogSeverityScreen lets us set the desired log level. (Default: info)
	LogSeverityScreen string `json:"logSeverityScreen,omitempty"`
	// PrometheusMetricsEnabled can be set to enable the experimental Prometheus
	// metrics server (default: false)
	PrometheusMetricsEnabled bool `json:"prometheusMetricsEnabled,omitempty"`
	// PrometheusMetricsPort is the TCP port that the experimental Prometheus
	// metrics server should bind to (default: 9091)
	PrometheusMetricsPort int32 `json:"prometheusMetricsPort,omitempty"`
	// PrometheusGoMetricsEnabled enables Prometheus Go runtime metrics collection
	PrometheusGoMetricsEnabled bool `json:"prometheusGoMetricsEnabled,omitempty"`
	// PrometheusProcessMetricsEnabled enables Prometheus process metrics collection
	PrometheusProcessMetricsEnabled bool `json:"prometheusProcessMetricsEnabled,omitempty"`
}

// CanalNetworkingSpec declares that we want Canal networking
type CanalNetworkingSpec struct {
	// DefaultEndpointToHostAction allows users to configure the default behaviour
	// for traffic between pod to host after calico rules have been processed.
	// Default: ACCEPT (other options: DROP, RETURN)
	DefaultEndpointToHostAction string `json:"defaultEndpointToHostAction,omitempty"`
	// ChainInsertMode controls whether Felix inserts rules to the top of iptables chains, or
	// appends to the bottom. Leaving the default option is safest to prevent accidentally
	// breaking connectivity. Default: 'insert' (other options: 'append')
	ChainInsertMode string `json:"chainInsertMode,omitempty"`
	// PrometheusMetricsEnabled can be set to enable the experimental Prometheus
	// metrics server (default: false)
	PrometheusMetricsEnabled bool `json:"prometheusMetricsEnabled,omitempty"`
	// PrometheusMetricsPort is the TCP port that the experimental Prometheus
	// metrics server should bind to (default: 9091)
	PrometheusMetricsPort int32 `json:"prometheusMetricsPort,omitempty"`
	// PrometheusGoMetricsEnabled enables Prometheus Go runtime metrics collection
	PrometheusGoMetricsEnabled bool `json:"prometheusGoMetricsEnabled,omitempty"`
	// PrometheusProcessMetricsEnabled enables Prometheus process metrics collection
	PrometheusProcessMetricsEnabled bool `json:"prometheusProcessMetricsEnabled,omitempty"`
}

// KuberouterNetworkingSpec declares that we want Kube-router networking
type KuberouterNetworkingSpec struct {
}

// RomanaNetworkingSpec declares that we want Romana networking
type RomanaNetworkingSpec struct {
	// DaemonServiceIP is the Kubernetes Service IP for the romana-daemon pod
	DaemonServiceIP string `json:"daemonServiceIP,omitempty"`
	// EtcdServiceIP is the Kubernetes Service IP for the etcd backend used by Romana
	EtcdServiceIP string `json:"etcdServiceIP,omitempty"`
}

// AmazonVPCNetworkingSpec declares that we want Amazon VPC CNI networking
type AmazonVPCNetworkingSpec struct{}

// CiliumNetworkingSpec declares that we want Cilium networking
type CiliumNetworkingSpec struct {
	Debug  bool   `json:"debug,omitempty"`
	Tunnel string `json:"tunnel,omitempty"`
}
