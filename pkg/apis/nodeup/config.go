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

package nodeup

import (
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/nodelabels"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/reflectutils"
)

// Config is the configuration for the nodeup binary
type Config struct {
	// Assets are locations where we can find files to be installed
	// TODO: Remove once everything is in containers?
	Assets map[architectures.Architecture][]string `json:",omitempty"`
	// Images are a list of images we should preload
	Images map[architectures.Architecture][]*Image `json:"images,omitempty"`
	// ClusterName is the name of the cluster
	ClusterName string `json:",omitempty"`
	// Channels is a list of channels that we should apply
	Channels []string `json:"channels,omitempty"`
	// ApiserverAdditionalIPs are additional IP address to put in the apiserver server cert.
	ApiserverAdditionalIPs []string `json:",omitempty"`
	// WarmPoolImages are the container images to pre-pull during instance pre-initialization
	WarmPoolImages []string `json:"warmPoolImages,omitempty"`

	// Manifests for running etcd
	EtcdManifests []string `json:"etcdManifests,omitempty"`

	// CAs are the CA certificates to trust.
	CAs map[string]string
	// KeypairIDs are the IDs of keysets used to sign things.
	KeypairIDs map[string]string
	// DefaultMachineType is the first-listed instance machine type, used if querying instance metadata fails.
	DefaultMachineType *string `json:",omitempty"`
	// EnableLifecycleHook defines whether we need to complete a lifecycle hook.
	EnableLifecycleHook bool `json:",omitempty"`
	// StaticManifests describes generic static manifests
	// Using this allows us to keep complex logic out of nodeup
	StaticManifests []*StaticManifest `json:"staticManifests,omitempty"`
	// KubeletConfig defines the kubelet configuration.
	KubeletConfig kops.KubeletConfigSpec
	// SysctlParameters will configure kernel parameters using sysctl(8). When
	// specified, each parameter must follow the form variable=value, the way
	// it would appear in sysctl.conf.
	SysctlParameters []string `json:",omitempty"`
	// UpdatePolicy determines the policy for applying upgrades automatically.
	UpdatePolicy string
	// VolumeMounts are a collection of volume mounts.
	VolumeMounts []kops.VolumeMountSpec `json:",omitempty"`

	// FileAssets are a collection of file assets for this instance group.
	FileAssets []kops.FileAssetSpec `json:",omitempty"`
	// Hooks are for custom actions, for example on first installation.
	Hooks [][]kops.HookSpec
	// ContainerdConfig config holds the configuration for containerd
	ContainerdConfig *kops.ContainerdConfig `json:"containerdConfig,omitempty"`

	// APIServerConfig is additional configuration for nodes running an APIServer.
	APIServerConfig *APIServerConfig `json:",omitempty"`
	// Nvidia contains the configuration for nvidia
	Nvidia *kops.NvidiaConfig `json:",omitempty"`
}

// BootConfig is the configuration for the nodeup binary that might be too big to fit in userdata.
type BootConfig struct {
	// CloudProvider is the cloud provider in use.
	CloudProvider string
	// ConfigBase is the base VFS path for config objects.
	ConfigBase *string `json:",omitempty"`
	// ConfigServer holds the configuration for the configuration server.
	ConfigServer *ConfigServerOptions `json:",omitempty"`
	// InstanceGroupName is the name of the instance group.
	InstanceGroupName string `json:",omitempty"`
	// InstanceGroupRole is the instance group role.
	InstanceGroupRole kops.InstanceGroupRole
	// NodeupConfigHash holds a secure hash of the nodeup.Config.
	NodeupConfigHash string
}

type ConfigServerOptions struct {
	// Server is the address of the configuration server to use (kops-controller)
	Server string `json:"server,omitempty"`
	// CACertificates are the certificates to trust for fi.CertificateIDCA.
	CACertificates string
}

// Image is a docker image we should pre-load
type Image struct {
	// This is the name we would pass to "docker run", whereas source could be a URL from which we would download an image.
	Name string `json:"name,omitempty"`
	// Sources is a list of URLs from which we should download the image
	Sources []string `json:"sources,omitempty"`
	// Hash is the hash of the file, to verify image integrity (even over http)
	Hash string `json:"hash,omitempty"`
}

// StaticManifest is a generic static manifest
type StaticManifest struct {
	// Key identifies the static manifest
	Key string `json:"key,omitempty"`
	// Path is the path to the manifest
	Path string `json:"path,omitempty"`
}

// APIServerConfig is additional configuration for nodes running an APIServer.
type APIServerConfig struct {
	// KubeAPIServer is a copy of the KubeAPIServerConfig from the cluster spec.
	KubeAPIServer *kops.KubeAPIServerConfig
	// EncryptionConfigSecretHash is a hash of the encryptionconfig secret.
	// It is empty if EncryptionConfig is not enabled.
	// TODO: give secrets IDs and look them up like we do keypairs.
	EncryptionConfigSecretHash string `json:",omitempty"`
	// ServiceAccountPublicKeys are the service-account public keys to trust.
	ServiceAccountPublicKeys string
}

func NewConfig(cluster *kops.Cluster, instanceGroup *kops.InstanceGroup) (*Config, *BootConfig) {
	role := instanceGroup.Spec.Role
	isMaster := role == kops.InstanceGroupRoleMaster

	clusterHooks := filterHooks(cluster.Spec.Hooks, instanceGroup.Spec.Role)
	igHooks := filterHooks(instanceGroup.Spec.Hooks, instanceGroup.Spec.Role)

	config := Config{
		ClusterName:      cluster.ObjectMeta.Name,
		CAs:              map[string]string{},
		KeypairIDs:       map[string]string{},
		SysctlParameters: instanceGroup.Spec.SysctlParameters,
		VolumeMounts:     instanceGroup.Spec.VolumeMounts,
		FileAssets:       append(filterFileAssets(instanceGroup.Spec.FileAssets, role), filterFileAssets(cluster.Spec.FileAssets, role)...),
		Hooks:            [][]kops.HookSpec{igHooks, clusterHooks},
	}

	bootConfig := BootConfig{
		CloudProvider:     cluster.Spec.CloudProvider,
		InstanceGroupName: instanceGroup.ObjectMeta.Name,
		InstanceGroupRole: role,
	}

	warmPool := cluster.Spec.WarmPool.ResolveDefaults(instanceGroup)
	if warmPool.IsEnabled() && warmPool.EnableLifecycleHook {
		config.EnableLifecycleHook = true
	}

	if isMaster {
		reflectutils.JSONMergeStruct(&config.KubeletConfig, cluster.Spec.MasterKubelet)

		// A few settings in Kubelet override those in MasterKubelet. I'm not sure why.
		if cluster.Spec.Kubelet != nil && cluster.Spec.Kubelet.AnonymousAuth != nil && !*cluster.Spec.Kubelet.AnonymousAuth {
			config.KubeletConfig.AnonymousAuth = fi.Bool(false)
		}
	} else {
		reflectutils.JSONMergeStruct(&config.KubeletConfig, cluster.Spec.Kubelet)
	}

	if isMaster || role == kops.InstanceGroupRoleAPIServer {
		config.APIServerConfig = &APIServerConfig{
			KubeAPIServer: cluster.Spec.KubeAPIServer,
		}
	}

	if instanceGroup.Spec.Kubelet != nil {
		useSecureKubelet := config.KubeletConfig.AnonymousAuth != nil && !*config.KubeletConfig.AnonymousAuth

		reflectutils.JSONMergeStruct(&config.KubeletConfig, instanceGroup.Spec.Kubelet)

		if useSecureKubelet {
			config.KubeletConfig.AnonymousAuth = fi.Bool(false)
		}
	}

	// We include the NodeLabels in the userdata even for Kubernetes 1.16 and later so that
	// rolling update will still replace nodes when they change.
	config.KubeletConfig.NodeLabels = nodelabels.BuildNodeLabels(cluster, instanceGroup)

	config.KubeletConfig.Taints = append(config.KubeletConfig.Taints, instanceGroup.Spec.Taints...)

	if instanceGroup.Spec.UpdatePolicy != nil {
		config.UpdatePolicy = *instanceGroup.Spec.UpdatePolicy
	} else if cluster.Spec.UpdatePolicy != nil {
		config.UpdatePolicy = *cluster.Spec.UpdatePolicy
	} else {
		config.UpdatePolicy = kops.UpdatePolicyAutomatic
	}

	if cluster.Spec.Networking != nil && cluster.Spec.Networking.AmazonVPC != nil {
		config.DefaultMachineType = fi.String(strings.Split(instanceGroup.Spec.MachineType, ",")[0])
	}

	return &config, &bootConfig
}

func filterFileAssets(f []kops.FileAssetSpec, role kops.InstanceGroupRole) []kops.FileAssetSpec {
	var fileAssets []kops.FileAssetSpec
	for _, fileAsset := range f {
		if len(fileAsset.Roles) > 0 && !containsRole(role, fileAsset.Roles) {
			continue
		}
		fileAsset.Roles = nil
		fileAssets = append(fileAssets, fileAsset)
	}
	return fileAssets
}

func filterHooks(h []kops.HookSpec, role kops.InstanceGroupRole) []kops.HookSpec {
	var hooks []kops.HookSpec
	for _, hook := range h {
		if len(hook.Roles) > 0 && !containsRole(role, hook.Roles) {
			continue
		}
		hook.Roles = nil
		hooks = append(hooks, hook)
	}
	return hooks
}

func containsRole(v kops.InstanceGroupRole, list []kops.InstanceGroupRole) bool {
	for _, x := range list {
		if v == x {
			return true
		}
	}

	return false
}
