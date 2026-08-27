/*
Copyright 2026 The Kubernetes Authors.

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

package install

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	runtimeserializer "k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/randfill"
)

// The fixups below constrain a fuzzed internal object to values the target API
// version is able to represent. Every fixup marks a field, or a combination of
// fields, that does not survive a round trip; see roundtrip_test.go for the
// catalogue of what is not round trippable and why.

// fixupClusterSpec applies the fixups every API version needs: values that no version
// round trips, either because a defaulter rewrites them on decode or because the
// internal type marks the field `json:"-"` and never serializes it at all.
func fixupClusterSpec(spec *kops.ClusterSpec) {
	// SetDefaults_ClusterSpec fills these in on decode, so an unset value is never
	// observable after a round trip.
	if spec.Networking.Topology == nil {
		spec.Networking.Topology = &kops.TopologySpec{}
	}
	if spec.Networking.Topology.DNS == "" {
		spec.Networking.Topology.DNS = kops.DNSTypePublic
	}
	if spec.Authorization == nil || spec.Authorization.IsEmpty() {
		spec.Authorization = &kops.AuthorizationSpec{RBAC: &kops.RBACAuthorizationSpec{}}
	}
	if spec.Networking.Flannel != nil && spec.Networking.Flannel.Backend == "" {
		spec.Networking.Flannel.Backend = "udp"
	}

	// Only one cloud provider can be configured: v1alpha2 stores the choice as a
	// single string, and validation rejects more than one.
	keepOneProvider(&spec.CloudProvider)
}

// keepOneProvider clears all but the first configured cloud provider. v1alpha2
// stores the choice as a single string, so a spec naming two providers cannot round
// trip; validation rejects it too.
func keepOneProvider(spec *kops.CloudProviderSpec) {
	switch {
	case spec.AWS != nil:
		*spec = kops.CloudProviderSpec{AWS: spec.AWS}
	case spec.Azure != nil:
		*spec = kops.CloudProviderSpec{Azure: spec.Azure}
	case spec.DO != nil:
		*spec = kops.CloudProviderSpec{DO: spec.DO}
	case spec.GCE != nil:
		*spec = kops.CloudProviderSpec{GCE: spec.GCE}
	case spec.Hetzner != nil:
		*spec = kops.CloudProviderSpec{Hetzner: spec.Hetzner}
	case spec.Linode != nil:
		*spec = kops.CloudProviderSpec{Linode: spec.Linode}
	case spec.Openstack != nil:
		*spec = kops.CloudProviderSpec{Openstack: spec.Openstack}
	case spec.Scaleway != nil:
		*spec = kops.CloudProviderSpec{Scaleway: spec.Scaleway}
	}
}

// fixupKubeAPIServer clears the kube-apiserver fields that the internal type
// marks `json:"-"`.
func fixupKubeAPIServer(spec *kops.KubeAPIServerConfig) {
	// The legacy OIDC flags exist only so that v1alpha2 can promote them into
	// spec.authentication.oidc on read, which is a one-way migration.
	spec.OIDCClientID = nil
	spec.OIDCGroupsClaim = nil
	spec.OIDCGroupsPrefix = nil
	spec.OIDCIssuerURL = nil
	spec.OIDCRequiredClaim = nil
	spec.OIDCUsernameClaim = nil
	spec.OIDCUsernamePrefix = nil
}

// commonFuzzerFuncs returns the fuzzer functions shared by every API version.
func commonFuzzerFuncs() []any {
	return []any{
		func(spec *kops.EBSCSIDriverSpec, c randfill.Continue) {
			c.FillNoCustom(spec)
			// Enabled is `json:"-"` in the internal type.
			spec.Enabled = nil
		},
		func(sel *corev1.FileKeySelector, c randfill.Continue) {
			c.FillNoCustom(sel)
			// corev1 declares Optional as `+default=false`, so defaulter-gen fills it
			// in on decode.
			if sel.Optional == nil {
				sel.Optional = ptr.To(false)
			}
		},
	}
}

// v1alpha2FuzzerFuncs returns fuzzer functions producing internal objects that
// v1alpha2 can represent.
func v1alpha2FuzzerFuncs(_ runtimeserializer.CodecFactory) []any {
	return append(commonFuzzerFuncs(),
		func(spec *kops.ClusterSpec, c randfill.Continue) {
			c.FillNoCustom(spec)
			fixupClusterSpec(spec)

			// v1alpha2 stashes the AWS and GCE settings it knows about in
			// spec.cloudConfig, creating the struct when it has something to write, so
			// a round trip can never observe it unset. This is applied unconditionally
			// rather than only for the settings that trigger it, which would mean
			// duplicating the conversion's own stashing rules here.
			if spec.CloudConfig == nil {
				spec.CloudConfig = &kops.CloudConfiguration{}
			}

			// v1alpha2 keeps the OIDC settings in spec.kubeAPIServer, and likewise
			// creates that struct when it has something to write.
			if spec.Authentication != nil && spec.Authentication.OIDC != nil && spec.KubeAPIServer == nil {
				spec.KubeAPIServer = &kops.KubeAPIServerConfig{}
			}

			// v1alpha2 has no field for the provider binaries location
			// (kubernetes/kops#18740).
			if spec.CloudProvider.AWS != nil {
				spec.CloudProvider.AWS.BinariesLocation = nil
			}
			if spec.CloudProvider.GCE != nil {
				spec.CloudProvider.GCE.BinariesLocation = nil
			}
		},
		func(spec *kops.KubeAPIServerConfig, c randfill.Continue) {
			c.FillNoCustom(spec)
			fixupKubeAPIServer(spec)
		},
		func(spec *kops.OIDCAuthenticationSpec, c randfill.Continue) {
			c.FillNoCustom(spec)

			// v1alpha2 flattens the OIDC settings into the kube-apiserver flags, and
			// only reconstructs spec.authentication.oidc when at least one is set.
			if spec.ClientID == nil {
				spec.ClientID = ptr.To(c.String(0))
			}

			// GroupsClaims is stored comma-joined in a single flag, so an individual
			// claim may not contain a comma. An empty list joins to "", which splits
			// back to a single empty claim rather than to an empty list, so normalize
			// it to unset.
			for i, claim := range spec.GroupsClaims {
				spec.GroupsClaims[i] = strings.ReplaceAll(claim, ",", "")
			}
			if len(spec.GroupsClaims) == 0 {
				spec.GroupsClaims = nil
			}

			// RequiredClaims is stored as a list of "key=value" flags, so a key may
			// not contain an equals sign. Unlike GroupsClaims it needs no empty-to-nil
			// normalization: omitempty drops an empty slice, and Semantic.DeepEqual
			// equates nil with empty, whereas a non-nil *string is not dropped.
			if spec.RequiredClaims != nil {
				claims := make(map[string]string, len(spec.RequiredClaims))
				for k, v := range spec.RequiredClaims {
					claims[strings.ReplaceAll(k, "=", "")] = v
				}
				spec.RequiredClaims = claims
			}
		},
		func(spec *kops.InstanceGroupSpec, c randfill.Continue) {
			c.FillNoCustom(spec)

			// v1alpha2 flattens RootVolume into individual rootVolume* fields, so an
			// entirely empty RootVolume decodes back as unset.
			if spec.RootVolume != nil && *spec.RootVolume == (kops.InstanceRootVolumeSpec{}) {
				spec.RootVolume = nil
			}
		},
	)
}

// v1alpha3FuzzerFuncs returns fuzzer functions producing internal objects that
// v1alpha3 can represent. Everything cleared here is a setting kOps has removed but
// still carries on the internal type, and which v1alpha3 marks `json:"-"`.
func v1alpha3FuzzerFuncs(_ runtimeserializer.CodecFactory) []any {
	return append(commonFuzzerFuncs(),
		func(spec *kops.ClusterSpec, c randfill.Continue) {
			c.FillNoCustom(spec)
			fixupClusterSpec(spec)

			spec.ContainerRuntime = ""
			spec.Docker = nil
			spec.NodeAuthorization = nil
			if spec.IAM != nil {
				spec.IAM.Legacy = false
			}
			spec.Networking.Classic = nil
			spec.Networking.Romana = nil
			spec.Networking.LyftVPC = nil
			if spec.Networking.Calico != nil {
				spec.Networking.Calico.CrossSubnet = nil
			}
			for i := range spec.EtcdClusters {
				spec.EtcdClusters[i].Provider = ""
				spec.EtcdClusters[i].LeaderElectionTimeout = nil
				spec.EtcdClusters[i].HeartbeatInterval = nil
			}
		},
		func(spec *kops.KubeAPIServerConfig, c randfill.Continue) {
			c.FillNoCustom(spec)
			fixupKubeAPIServer(spec)

			spec.AllowPrivileged = nil
			spec.AuthorizationRBACSuperUser = nil
			spec.BasicAuthFile = ""
			spec.EtcdCAFile = ""
			spec.EtcdCertFile = ""
			spec.EtcdKeyFile = ""
			spec.EtcdQuorumRead = nil
			spec.ExperimentalEncryptionProviderConfig = nil
			spec.KubeletClientCertificate = ""
			spec.KubeletClientKey = ""
			spec.ProxyClientCertFile = nil
			spec.ProxyClientKeyFile = nil
			spec.RequestheaderClientCAFile = ""
			spec.TLSCertFile = ""
			spec.TLSPrivateKeyFile = ""
		},
		func(spec *kops.KubeletConfigSpec, c randfill.Continue) {
			c.FillNoCustom(spec)

			spec.AllowPrivileged = nil
			spec.APIServers = ""
			spec.BabysitDaemons = nil
			spec.ClientCAFile = ""
			spec.ConfigureCBR0 = nil
			spec.DockerDisableSharedPID = nil
			spec.EnableCustomMetrics = nil
			spec.ExperimentalAllowedUnsafeSysctls = nil
			spec.HostnameOverride = ""
			spec.NodeLabels = nil
			spec.NvidiaGPUs = 0
			spec.ReconcileCIDR = nil
			spec.RegisterSchedulable = nil
			spec.RequireKubeconfig = nil
		},
		func(spec *kops.KubeControllerManagerConfig, c randfill.Continue) {
			c.FillNoCustom(spec)

			spec.RootCAFile = ""
			spec.ServiceAccountPrivateKeyFile = ""
		},
		func(spec *kops.KubeProxyConfig, c randfill.Continue) {
			c.FillNoCustom(spec)

			spec.BindAddress = ""
			spec.HostnameOverride = ""
		},
	)
}
