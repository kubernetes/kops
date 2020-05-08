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
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"

	"github.com/blang/semver"
)

// KubeAPIServerOptionsBuilder adds options for the apiserver to the model
type KubeAPIServerOptionsBuilder struct {
	*OptionsContext
}

var _ loader.OptionsBuilder = &KubeAPIServerOptionsBuilder{}

// BuildOptions is responsible for filling in the default settings for the kube apiserver
func (b *KubeAPIServerOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)
	if clusterSpec.KubeAPIServer == nil {
		clusterSpec.KubeAPIServer = &kops.KubeAPIServerConfig{}
	}
	c := clusterSpec.KubeAPIServer

	if c.APIServerCount == nil {
		count := b.buildAPIServerCount(clusterSpec)
		if count == 0 {
			return fmt.Errorf("no instance groups found")
		}
		c.APIServerCount = fi.Int32(int32(count))
	}

	// @question: should the question every be able to set this?
	if c.StorageBackend == nil {
		// @note: we can use the first version as we enforce both running the same versions.
		// albeit feels a little weird to do this
		sem, err := semver.Parse(strings.TrimPrefix(clusterSpec.EtcdClusters[0].Version, "v"))
		if err != nil {
			return err
		}
		c.StorageBackend = fi.String(fmt.Sprintf("etcd%d", sem.Major))
	}

	if c.KubeletPreferredAddressTypes == nil {
		if b.IsKubernetesGTE("1.5") {
			// We prioritize the internal IP above the hostname
			c.KubeletPreferredAddressTypes = []string{
				string(v1.NodeInternalIP),
				string(v1.NodeHostName),
				string(v1.NodeExternalIP),
			}

			if b.IsKubernetesLT("1.7") {
				// NodeLegacyHostIP was removed in 1.7; we add it to prior versions with lowest precedence
				c.KubeletPreferredAddressTypes = append(c.KubeletPreferredAddressTypes, "LegacyHostIP")
			}
		}
	}

	if clusterSpec.Authentication != nil {
		if clusterSpec.Authentication.Kopeio != nil {
			c.AuthenticationTokenWebhookConfigFile = fi.String("/etc/kubernetes/authn.config")
		}
	}

	if clusterSpec.Authorization == nil || clusterSpec.Authorization.IsEmpty() {
		// Do nothing - use the default as defined by the apiserver
		// (this won't happen anyway because of our default logic)
	} else if clusterSpec.Authorization.AlwaysAllow != nil {
		clusterSpec.KubeAPIServer.AuthorizationMode = fi.String("AlwaysAllow")
	} else if clusterSpec.Authorization.RBAC != nil {
		var modes []string

		if b.IsKubernetesGTE("1.10") {
			if fi.BoolValue(clusterSpec.KubeAPIServer.EnableBootstrapAuthToken) {
				// Enable the Node authorizer, used for special per-node RBAC policies
				modes = append(modes, "Node")
			}
		}
		modes = append(modes, "RBAC")

		clusterSpec.KubeAPIServer.AuthorizationMode = fi.String(strings.Join(modes, ","))
	}

	if clusterSpec.KubeAPIServer.EtcdQuorumRead == nil {
		if b.IsKubernetesGTE("1.9") {
			// 1.9 changed etcd-quorum-reads default to true
			// There's a balance between some bugs which are attributed to not having etcd-quorum-reads,
			// and the poor implementation of quorum-reads in etcd2.

			etcdHA := false
			etcdV2 := true
			for _, c := range clusterSpec.EtcdClusters {
				if len(c.Members) > 1 {
					etcdHA = true
				}
				if c.Version != "" && !strings.HasPrefix(c.Version, "2.") {
					etcdV2 = false
				}
			}

			if !etcdV2 {
				// etcd3 quorum reads are cheap.  Stick with default (which is to enable quorum reads)
				clusterSpec.KubeAPIServer.EtcdQuorumRead = nil
			} else {
				// etcd2 quorum reads go through raft => write to disk => expensive
				if !etcdHA {
					// Turn off quorum reads - they still go through raft, but don't serve any purpose in non-HA clusters.
					clusterSpec.KubeAPIServer.EtcdQuorumRead = fi.Bool(false)
				} else {
					// The problematic case.  We risk exposing more bugs, but against that we have to balance performance.
					// For now we turn off quorum reads - it's a bad enough performance regression
					// We'll likely make this default to true once we can set IOPS on the etcd volume and can easily upgrade to etcd3
					clusterSpec.KubeAPIServer.EtcdQuorumRead = fi.Bool(false)
				}
			}
		}
	}

	if err := b.configureAggregation(clusterSpec); err != nil {
		return nil
	}

	image, err := Image("kube-apiserver", b.Architecture(), clusterSpec, b.AssetBuilder)
	if err != nil {
		return err
	}
	c.Image = image

	switch kops.CloudProviderID(clusterSpec.CloudProvider) {
	case kops.CloudProviderAWS:
		c.CloudProvider = "aws"
	case kops.CloudProviderGCE:
		c.CloudProvider = "gce"
	case kops.CloudProviderDO:
		c.CloudProvider = "external"
	case kops.CloudProviderVSphere:
		c.CloudProvider = "vsphere"
	case kops.CloudProviderBareMetal:
		// for baremetal, we don't specify a cloudprovider to apiserver
	case kops.CloudProviderOpenstack:
		c.CloudProvider = "openstack"
	case kops.CloudProviderALI:
		c.CloudProvider = "alicloud"
	default:
		return fmt.Errorf("unknown cloudprovider %q", clusterSpec.CloudProvider)
	}

	if clusterSpec.ExternalCloudControllerManager != nil {
		c.CloudProvider = "external"
	}

	c.LogLevel = 2
	c.SecurePort = 443

	if b.IsKubernetesGTE("1.10") {
		c.BindAddress = "0.0.0.0"
		c.InsecureBindAddress = "127.0.0.1"
	} else {
		c.Address = "127.0.0.1"
	}

	c.AllowPrivileged = fi.Bool(true)
	c.ServiceClusterIPRange = clusterSpec.ServiceClusterIPRange
	c.EtcdServers = []string{"http://127.0.0.1:4001"}
	c.EtcdServersOverrides = []string{"/events#http://127.0.0.1:4002"}

	// TODO: We can probably rewrite these more clearly in descending order
	if b.IsKubernetesGTE("1.3") && b.IsKubernetesLT("1.4") {
		c.AdmissionControl = []string{
			"NamespaceLifecycle",
			"LimitRanger",
			"ServiceAccount",
			"PersistentVolumeLabel",
			"ResourceQuota",
		}
	}
	if b.IsKubernetesGTE("1.4") && b.IsKubernetesLT("1.5") {
		c.AdmissionControl = []string{
			"NamespaceLifecycle",
			"LimitRanger",
			"ServiceAccount",
			"PersistentVolumeLabel",
			"DefaultStorageClass",
			"ResourceQuota",
		}
	}
	if b.IsKubernetesGTE("1.5") && b.IsKubernetesLT("1.6") {
		c.AdmissionControl = []string{
			"NamespaceLifecycle",
			"LimitRanger",
			"ServiceAccount",
			"PersistentVolumeLabel",
			"DefaultStorageClass",
			"ResourceQuota",
		}
	}
	if b.IsKubernetesGTE("1.6") && b.IsKubernetesLT("1.7") {
		c.AdmissionControl = []string{
			"NamespaceLifecycle",
			"LimitRanger",
			"ServiceAccount",
			"PersistentVolumeLabel",
			"DefaultStorageClass",
			"DefaultTolerationSeconds",
			"ResourceQuota",
		}
	}
	if b.IsKubernetesGTE("1.7") && b.IsKubernetesLT("1.9") {
		c.AdmissionControl = []string{
			"Initializers",
			"NamespaceLifecycle",
			"LimitRanger",
			"ServiceAccount",
			"PersistentVolumeLabel",
			"DefaultStorageClass",
			"DefaultTolerationSeconds",
			"NodeRestriction",
			"ResourceQuota",
		}
	}
	if b.IsKubernetesGTE("1.9") && b.IsKubernetesLT("1.10") {
		c.AdmissionControl = []string{
			"Initializers",
			"NamespaceLifecycle",
			"LimitRanger",
			"ServiceAccount",
			"PersistentVolumeLabel",
			"DefaultStorageClass",
			"DefaultTolerationSeconds",
			"MutatingAdmissionWebhook",
			"ValidatingAdmissionWebhook",
			"NodeRestriction",
			"ResourceQuota",
		}
	}
	// Based on recommendations from:
	// https://kubernetes.io/docs/admin/admission-controllers/#is-there-a-recommended-set-of-admission-controllers-to-use
	if b.IsKubernetesGTE("1.10") && b.IsKubernetesLT("1.12") {
		c.EnableAdmissionPlugins = []string{
			"Initializers",
			"NamespaceLifecycle",
			"LimitRanger",
			"ServiceAccount",
			"PersistentVolumeLabel",
			"DefaultStorageClass",
			"DefaultTolerationSeconds",
			"MutatingAdmissionWebhook",
			"ValidatingAdmissionWebhook",
			"NodeRestriction",
			"ResourceQuota",
		}
		c.EnableAdmissionPlugins = append(c.EnableAdmissionPlugins, c.AppendAdmissionPlugins...)
	}
	// Based on recommendations from:
	// https://kubernetes.io/docs/admin/admission-controllers/#is-there-a-recommended-set-of-admission-controllers-to-use
	if b.IsKubernetesGTE("1.12") {
		c.EnableAdmissionPlugins = []string{
			"NamespaceLifecycle",
			"LimitRanger",
			"ServiceAccount",
			"PersistentVolumeLabel",
			"DefaultStorageClass",
			"DefaultTolerationSeconds",
			"MutatingAdmissionWebhook",
			"ValidatingAdmissionWebhook",
			"NodeRestriction",
			"ResourceQuota",
		}
		c.EnableAdmissionPlugins = append(c.EnableAdmissionPlugins, c.AppendAdmissionPlugins...)
	}

	// We make sure to disable AnonymousAuth from when it was introduced
	if b.IsKubernetesGTE("1.5") {
		c.AnonymousAuth = fi.Bool(false)
	}

	if b.IsKubernetesGTE("1.17") {
		// We query via the kube-apiserver-healthcheck proxy, which listens on port 8080
		c.InsecurePort = 0
	} else {
		// Older versions of kubernetes continue to rely on the insecure port: kubernetes issue #43784
		c.InsecurePort = 8080
	}

	return nil
}

// buildAPIServerCount calculates the count of the api servers, essentially the number of node marked as Master role
func (b *KubeAPIServerOptionsBuilder) buildAPIServerCount(clusterSpec *kops.ClusterSpec) int {
	// The --apiserver-count flag is (generally agreed) to be something we need to get rid of in k8s

	// We should do something like this:

	//count := 0
	//for _, ig := range b.InstanceGroups {
	//	if !ig.IsMaster() {
	//		continue
	//	}
	//	size := fi.IntValue(ig.Spec.MaxSize)
	//	if size == 0 {
	//		size = fi.IntValue(ig.Spec.MinSize)
	//	}
	//	count += size
	//}

	// But if we do, we end up with a weird dependency on InstanceGroups.  We actually could tolerate
	// that in kops, but we don't really want to.

	// So instead, we assume that the etcd cluster size is the API Server Count.
	// We can re-examine this when we allow separate etcd clusters - at which time hopefully
	// the flag won't exist

	counts := make(map[string]int)
	for _, etcdCluster := range clusterSpec.EtcdClusters {
		counts[etcdCluster.Name] = len(etcdCluster.Members)
	}

	count := counts["main"]

	return count
}

// configureAggregation sets up the aggregation options
func (b *KubeAPIServerOptionsBuilder) configureAggregation(clusterSpec *kops.ClusterSpec) error {
	if b.IsKubernetesGTE("1.7") {
		clusterSpec.KubeAPIServer.RequestheaderAllowedNames = []string{"aggregator"}
		clusterSpec.KubeAPIServer.RequestheaderExtraHeaderPrefixes = []string{"X-Remote-Extra-"}
		clusterSpec.KubeAPIServer.RequestheaderGroupHeaders = []string{"X-Remote-Group"}
		clusterSpec.KubeAPIServer.RequestheaderUsernameHeaders = []string{"X-Remote-User"}
	}

	return nil
}
