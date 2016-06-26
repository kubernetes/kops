package nodeup

import (
//"k8s.io/kube-deploy/upup/pkg/fi"
)

type NodeUpConfig struct {
	// Tags enable/disable chunks of the model
	Tags []string `json:",omitempty"`
	// Assets are locations where we can find files to be installed
	// TODO: Remove once everything is in containers?
	Assets []string `json:",omitempty"`

	// ClusterLocation is the VFS path to the cluster spec
	ClusterLocation string `json:",omitempty"`
}

// Our client configuration structure
// Wherever possible, we try to use the types & names in https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/componentconfig/types.go

//type NodeConfig struct {
//	Kubelet               *KubeletConfig               `json:",omitempty"`
//	KubeProxy             *KubeProxyConfig             `json:",omitempty"`
//	KubeControllerManager *KubeControllerManagerConfig `json:",omitempty"`
//	KubeScheduler         *KubeSchedulerConfig         `json:",omitempty"`
//	Docker                *DockerConfig                `json:",omitempty"`
//	APIServer             *APIServerConfig             `json:",omitempty"`
//
//	DNS *DNSConfig `json:",omitempty"`
//
//	// NodeConfig can directly access a store of secrets, keys or configuration
//	// (for example on S3) and then configure based on that
//	// This supports (limited) dynamic reconfiguration also
//	SecretStore string `json:",omitempty"`
//	KeyStore    string `json:",omitempty"`
//	ConfigStore string `json:",omitempty"`
//
//	KubeUser string `json:",omitempty"`
//
//	Tags   []string `json:",omitempty"`
//	Assets []string `json:",omitempty"`
//
//	MasterInternalName string `json:",omitempty"`
//
//	// The DNS zone to use if configuring a cloud provided DNS zone
//	DNSZone string `json:",omitempty"`
//
//	// Deprecated in favor of KeyStore / SecretStore
//	Tokens       map[string]string          `json:",omitempty"`
//	Certificates map[string]*fi.Certificate `json:",omitempty"`
//	PrivateKeys  map[string]*fi.PrivateKey  `json:",omitempty"`
//}
