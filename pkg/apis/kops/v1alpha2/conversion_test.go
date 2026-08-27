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

package v1alpha2_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/install"
	"k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/utils/ptr"
)

// The conversions between the internal type and v1alpha2 are hand written, and the
// round trip test in pkg/apis/kops/install cannot cover the cases where they are
// deliberately lossy or where they rename a value. Those are pinned here instead.

// testScheme is shared: Install registers several hundred conversion functions per
// call, and the resulting scheme is read only once registration completes.
var testScheme = sync.OnceValue(func() *runtime.Scheme {
	scheme := runtime.NewScheme()
	install.Install(scheme)

	return scheme
})

func toV1alpha2(t *testing.T, spec kops.ClusterSpec) *v1alpha2.ClusterSpec {
	t.Helper()

	out := &v1alpha2.Cluster{}
	if err := testScheme().Convert(&kops.Cluster{Spec: spec}, out, nil); err != nil {
		t.Fatalf("converting to v1alpha2: %v", err)
	}

	return &out.Spec
}

// toInternal converts a v1alpha2 spec back to the internal type. It drops
// spec.authentication.oidc, which v1alpha2 declares `json:"-"` and types as the internal
// *kops.OIDCAuthenticationSpec, so a real decode never carries one; the generated
// conversion copies that pointer straight across, which would otherwise hand the
// assertion back the very object the test passed in. The AuthenticationSpec is copied
// rather than modified in place, because the by-value spec still shares that pointer
// with the caller.
func toInternal(t *testing.T, spec v1alpha2.ClusterSpec) *kops.ClusterSpec {
	t.Helper()

	if spec.Authentication != nil {
		auth := *spec.Authentication
		auth.OIDC = nil
		spec.Authentication = &auth
	}

	out := &kops.Cluster{}
	if err := testScheme().Convert(&v1alpha2.Cluster{Spec: spec}, out, nil); err != nil {
		t.Fatalf("converting to internal: %v", err)
	}

	return &out.Spec
}

// TestConvertInvertedBooleans covers the fields v1alpha2 spells as the negation of
// the internal field.
func TestConvertInvertedBooleans(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		t.Run(fmt.Sprintf("enabled=%v", enabled), func(t *testing.T) {
			internal := kops.ClusterSpec{
				Networking: kops.NetworkingSpec{
					TagSubnets: ptr.To(enabled),
					Canal:      &kops.CanalNetworkingSpec{FlanneldIptablesForwardRules: ptr.To(enabled)},
					Cilium: &kops.CiliumNetworkingSpec{
						InstallIptablesRules: ptr.To(enabled),
						Masquerade:           ptr.To(enabled),
					},
				},
				Hooks: []kops.HookSpec{{Enabled: ptr.To(enabled)}},
			}

			external := toV1alpha2(t, internal)
			if external.LegacyNetworking == nil {
				t.Fatal("networking was not populated")
			}
			if external.LegacyNetworking.Canal == nil || external.LegacyNetworking.Cilium == nil {
				t.Fatalf("canal and cilium were not populated: %#v", external.LegacyNetworking)
			}

			want := ptr.To(!enabled)
			for _, tc := range []struct {
				name string
				got  *bool
			}{
				{"tagSubnets", external.TagSubnets},
				{"canal.flanneldIptablesForwardRules", external.LegacyNetworking.Canal.FlanneldIptablesForwardRules},
				{"cilium.installIptablesRules", external.LegacyNetworking.Cilium.InstallIptablesRules},
				{"cilium.masquerade", external.LegacyNetworking.Cilium.Masquerade},
			} {
				if diff := cmp.Diff(want, tc.got); diff != "" {
					t.Errorf("%s (-want +got):\n%s", tc.name, diff)
				}
			}
			if diff := cmp.Diff([]v1alpha2.HookSpec{{Enabled: want}}, external.Hooks); diff != "" {
				t.Errorf("hooks (-want +got):\n%s", diff)
			}
		})
	}
}

// TestConvertControlPlaneNaming covers the places where v1alpha2 still says "master".
func TestConvertControlPlaneNaming(t *testing.T) {
	internal := kops.ClusterSpec{
		AdditionalPolicies: map[string]string{"control-plane": "cp-policy", "node": "node-policy"},
		ExternalPolicies:   map[string][]string{"control-plane": {"cp-arn"}, "node": {"node-arn"}},
		Hooks:              []kops.HookSpec{{Roles: []kops.InstanceGroupRole{kops.InstanceGroupRoleControlPlane, kops.InstanceGroupRoleNode}}},
	}

	external := toV1alpha2(t, internal)
	if diff := cmp.Diff(map[string]string{"master": "cp-policy", "node": "node-policy"}, external.AdditionalPolicies); diff != "" {
		t.Errorf("toV1alpha2(%v).AdditionalPolicies (-want +got):\n%s", internal.AdditionalPolicies, diff)
	}
	if diff := cmp.Diff(map[string][]string{"master": {"cp-arn"}, "node": {"node-arn"}}, external.ExternalPolicies); diff != "" {
		t.Errorf("toV1alpha2(%v).ExternalPolicies (-want +got):\n%s", internal.ExternalPolicies, diff)
	}
	if diff := cmp.Diff([]v1alpha2.HookSpec{{Roles: []v1alpha2.InstanceGroupRole{"Master", "Node"}}}, external.Hooks); diff != "" {
		t.Errorf("toV1alpha2(%v).Hooks (-want +got):\n%s", internal.Hooks, diff)
	}

	back := toInternal(t, *external)
	if diff := cmp.Diff(internal.AdditionalPolicies, back.AdditionalPolicies); diff != "" {
		t.Errorf("additionalPolicies round trip (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(internal.ExternalPolicies, back.ExternalPolicies); diff != "" {
		t.Errorf("externalPolicies round trip (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(internal.Hooks, back.Hooks); diff != "" {
		t.Errorf("hooks round trip (-want +got):\n%s", diff)
	}
}

// TestConvertInstanceGroupRootVolume covers the flattening of spec.rootVolume into the
// individual v1alpha2 rootVolume* fields.
func TestConvertInstanceGroupRootVolume(t *testing.T) {
	internal := &kops.InstanceGroup{
		Spec: kops.InstanceGroupSpec{
			Role: kops.InstanceGroupRoleControlPlane,
			RootVolume: &kops.InstanceRootVolumeSpec{
				Size:          ptr.To(int32(64)),
				Type:          ptr.To("gp3"),
				IOPS:          ptr.To(int32(3000)),
				Throughput:    ptr.To(int32(125)),
				Optimization:  ptr.To(true),
				Encryption:    ptr.To(true),
				EncryptionKey: ptr.To("alias/kops"),
			},
		},
	}

	external := &v1alpha2.InstanceGroup{}
	if err := testScheme().Convert(internal, external, nil); err != nil {
		t.Fatalf("converting to v1alpha2: %v", err)
	}

	// Drop spec.rootVolume for the same reason toInternal drops
	// spec.authentication.oidc: v1alpha2 declares it `json:"-"` and types it as the
	// internal struct, so a real decode never carries one and the generated pointer copy
	// would otherwise let the round trip below pass through the alias rather than
	// through the flattened rootVolume* fields.
	external.Spec.RootVolume = nil

	// Compare the whole spec: a round trip alone cannot catch a symmetric error such as
	// writing IOPS into rootVolumeThroughput and reading it back from there.
	want := v1alpha2.InstanceGroupSpec{
		Role:                    "Master",
		RootVolumeSize:          ptr.To(int32(64)),
		RootVolumeType:          ptr.To("gp3"),
		RootVolumeIOPS:          ptr.To(int32(3000)),
		RootVolumeThroughput:    ptr.To(int32(125)),
		RootVolumeOptimization:  ptr.To(true),
		RootVolumeEncryption:    ptr.To(true),
		RootVolumeEncryptionKey: ptr.To("alias/kops"),
	}
	if diff := cmp.Diff(want, external.Spec); diff != "" {
		t.Errorf("Convert(%+v) to v1alpha2 (-want +got):\n%s", internal.Spec, diff)
	}

	back := &kops.InstanceGroup{}
	if err := testScheme().Convert(external, back, nil); err != nil {
		t.Fatalf("converting to internal: %v", err)
	}
	if diff := cmp.Diff(internal.Spec, back.Spec); diff != "" {
		t.Errorf("Convert(%+v) to internal (-want +got):\n%s", external.Spec, diff)
	}
}

// TestConvertOIDC covers the promotion of the legacy kube-apiserver OIDC flags into
// spec.authentication.oidc. wantExternal pins the v1alpha2 encoding itself, which is the
// backward compatibility contract every stored cluster spec depends on: a symmetric
// change to the separators or the loss of the sort would round trip fine while
// reinterpreting every spec already on disk.
func TestConvertOIDC(t *testing.T) {
	for _, tc := range []struct {
		name         string
		in           kops.OIDCAuthenticationSpec
		wantExternal v1alpha2.KubeAPIServerConfig
		want         kops.OIDCAuthenticationSpec
	}{
		{
			name: "round trips",
			in: kops.OIDCAuthenticationSpec{
				ClientID:       ptr.To("kubernetes"),
				IssuerURL:      ptr.To("https://example.com"),
				GroupsClaims:   []string{"groups", "roles"},
				RequiredClaims: map[string]string{"hd": "example.com", "aud": "kops"},
			},
			wantExternal: v1alpha2.KubeAPIServerConfig{
				OIDCClientID:  ptr.To("kubernetes"),
				OIDCIssuerURL: ptr.To("https://example.com"),
				// Comma joined, and sorted so that ranging the map stays deterministic.
				OIDCGroupsClaim:   ptr.To("groups,roles"),
				OIDCRequiredClaim: []string{"aud=kops", "hd=example.com"},
			},
			want: kops.OIDCAuthenticationSpec{
				ClientID:       ptr.To("kubernetes"),
				IssuerURL:      ptr.To("https://example.com"),
				GroupsClaims:   []string{"groups", "roles"},
				RequiredClaims: map[string]string{"hd": "example.com", "aud": "kops"},
			},
		},
		{
			// This documents a known defect, not desired behaviour: groupsClaims is
			// stored comma joined in a single flag, so a claim containing a comma comes
			// back as two claims.
			name: "a groups claim containing a comma is split in two",
			in: kops.OIDCAuthenticationSpec{
				ClientID:     ptr.To("kubernetes"),
				GroupsClaims: []string{"a,b"},
			},
			wantExternal: v1alpha2.KubeAPIServerConfig{
				OIDCClientID:    ptr.To("kubernetes"),
				OIDCGroupsClaim: ptr.To("a,b"),
			},
			want: kops.OIDCAuthenticationSpec{
				ClientID:     ptr.To("kubernetes"),
				GroupsClaims: []string{"a", "b"},
			},
		},
		{
			// Likewise a known defect: requiredClaims is stored as a list of
			// "key=value" flags, so an equals sign in a key moves into the value.
			name: "a required claim key containing an equals sign moves into the value",
			in: kops.OIDCAuthenticationSpec{
				ClientID:       ptr.To("kubernetes"),
				RequiredClaims: map[string]string{"a=b": "c"},
			},
			wantExternal: v1alpha2.KubeAPIServerConfig{
				OIDCClientID:      ptr.To("kubernetes"),
				OIDCRequiredClaim: []string{"a=b=c"},
			},
			want: kops.OIDCAuthenticationSpec{
				ClientID:       ptr.To("kubernetes"),
				RequiredClaims: map[string]string{"a": "b=c"},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			oidc := tc.in.DeepCopy()
			external := toV1alpha2(t, kops.ClusterSpec{Authentication: &kops.AuthenticationSpec{OIDC: oidc}})
			if external.KubeAPIServer == nil {
				t.Fatal("expected the OIDC settings to be written to spec.kubeAPIServer")
			}
			if diff := cmp.Diff(tc.wantExternal, *external.KubeAPIServer); diff != "" {
				t.Errorf("toV1alpha2(oidc=%+v).KubeAPIServer (-want +got):\n%s", tc.in, diff)
			}

			back := toInternal(t, *external)
			if back.Authentication == nil || back.Authentication.OIDC == nil {
				t.Fatal("expected spec.authentication.oidc to be reconstructed")
			}
			if diff := cmp.Diff(tc.want, *back.Authentication.OIDC); diff != "" {
				t.Errorf("oidc round trip (-want +got):\n%s", diff)
			}
		})
	}
}

// TestConvertOIDCFlagsWinOverInMemorySpec pins the precedence when both OIDC sources are
// populated. This is only reachable in memory, because v1alpha2 declares
// spec.authentication.oidc as `json:"-"`.
func TestConvertOIDCFlagsWinOverInMemorySpec(t *testing.T) {
	external := &v1alpha2.Cluster{
		Spec: v1alpha2.ClusterSpec{
			Authentication: &v1alpha2.AuthenticationSpec{
				OIDC: &kops.OIDCAuthenticationSpec{ClientID: ptr.To("in-memory")},
			},
			KubeAPIServer: &v1alpha2.KubeAPIServerConfig{OIDCClientID: ptr.To("from-flags")},
		},
	}

	out := &kops.Cluster{}
	if err := testScheme().Convert(external, out, nil); err != nil {
		t.Fatalf("converting to internal: %v", err)
	}
	if out.Spec.Authentication == nil || out.Spec.Authentication.OIDC == nil {
		t.Fatal("expected spec.authentication.oidc to be populated")
	}
	if diff := cmp.Diff(ptr.To("from-flags"), out.Spec.Authentication.OIDC.ClientID); diff != "" {
		t.Errorf("authentication.oidc.clientID (-want +got):\n%s", diff)
	}
}

// TestConvertOIDCRequiredClaimWithoutValue covers a malformed oidcRequiredClaim entry,
// which used to panic the conversion.
func TestConvertOIDCRequiredClaimWithoutValue(t *testing.T) {
	external := &v1alpha2.Cluster{
		Spec: v1alpha2.ClusterSpec{
			KubeAPIServer: &v1alpha2.KubeAPIServerConfig{OIDCRequiredClaim: []string{"hd=example.com", "novalue"}},
		},
	}

	err := testScheme().Convert(external, &kops.Cluster{}, nil)

	var fieldErr *field.Error
	if !errors.As(err, &fieldErr) {
		t.Fatalf("error = %v (%T); want a *field.Error", err, err)
	}
	if got, want := fieldErr.Type, field.ErrorTypeInvalid; got != want {
		t.Errorf("error type = %v; want %v", got, want)
	}
	if got, want := fieldErr.Field, "spec.kubeAPIServer.oidcRequiredClaim[1]"; got != want {
		t.Errorf("error field = %q; want %q", got, want)
	}
	if got, want := fieldErr.BadValue, "novalue"; got != want {
		t.Errorf("error value = %v; want %q", got, want)
	}
}

// TestConvertCloudProvider covers the single "cloudProvider" string that v1alpha2 uses
// in place of spec.cloudProvider, and the settings it stashes in spec.cloudConfig.
func TestConvertCloudProvider(t *testing.T) {
	for _, tc := range []struct {
		name          string
		in            kops.CloudProviderSpec
		wantLegacy    string
		checkExternal func(*testing.T, *v1alpha2.ClusterSpec)
	}{
		{
			name:       "aws settings move to cloudConfig",
			in:         kops.CloudProviderSpec{AWS: &kops.AWSSpec{ElbSecurityGroup: ptr.To("sg-1")}},
			wantLegacy: "aws",
			checkExternal: func(t *testing.T, spec *v1alpha2.ClusterSpec) {
				if spec.CloudConfig == nil {
					t.Fatal("cloudConfig was not populated")
				}
				if diff := cmp.Diff(ptr.To("sg-1"), spec.CloudConfig.ElbSecurityGroup); diff != "" {
					t.Errorf("cloudConfig.elbSecurityGroup (-want +got):\n%s", diff)
				}
			},
		},
		{
			name:       "gce project moves to spec.project",
			in:         kops.CloudProviderSpec{GCE: &kops.GCESpec{Project: "my-project"}},
			wantLegacy: "gce",
			checkExternal: func(t *testing.T, spec *v1alpha2.ClusterSpec) {
				if got, want := spec.Project, "my-project"; got != want {
					t.Errorf("project = %q; want %q", got, want)
				}
			},
		},
		{
			name:       "openstack settings move to cloudConfig",
			in:         kops.CloudProviderSpec{Openstack: &kops.OpenstackSpec{Router: &kops.OpenstackRouter{ExternalNetwork: ptr.To("public")}}},
			wantLegacy: "openstack",
			checkExternal: func(t *testing.T, spec *v1alpha2.ClusterSpec) {
				if spec.CloudConfig == nil || spec.CloudConfig.Openstack == nil || spec.CloudConfig.Openstack.Router == nil {
					t.Fatalf("cloudConfig.openstack.router was not populated: %#v", spec.CloudConfig)
				}
				if diff := cmp.Diff(ptr.To("public"), spec.CloudConfig.Openstack.Router.ExternalNetwork); diff != "" {
					t.Errorf("cloudConfig.openstack.router.externalNetwork (-want +got):\n%s", diff)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			internal := kops.ClusterSpec{CloudProvider: tc.in}

			external := toV1alpha2(t, internal)
			if got := external.LegacyCloudProvider; got != tc.wantLegacy {
				t.Errorf("cloudProvider = %q; want %q", got, tc.wantLegacy)
			}
			tc.checkExternal(t, external)

			back := toInternal(t, *external)
			if diff := cmp.Diff(tc.in, back.CloudProvider); diff != "" {
				t.Errorf("cloudProvider round trip (-want +got):\n%s", diff)
			}
		})
	}
}

// TestConvertCloudConfigRejectsMismatchedProvider covers the settings that v1alpha2
// stores in the provider agnostic spec.cloudConfig but that only one provider accepts.
func TestConvertCloudConfigRejectsMismatchedProvider(t *testing.T) {
	external := &v1alpha2.Cluster{
		Spec: v1alpha2.ClusterSpec{
			LegacyCloudProvider: "gce",
			CloudConfig:         &v1alpha2.CloudConfiguration{ElbSecurityGroup: ptr.To("sg-1")},
		},
	}

	err := testScheme().Convert(external, &kops.Cluster{}, nil)

	var fieldErr *field.Error
	if !errors.As(err, &fieldErr) {
		t.Fatalf("error = %v (%T); want a *field.Error", err, err)
	}
	if got, want := fieldErr.Type, field.ErrorTypeForbidden; got != want {
		t.Errorf("error type = %v; want %v", got, want)
	}
	if got, want := fieldErr.Field, "spec.cloudConfig.elbSecurityGroup"; got != want {
		t.Errorf("error field = %q; want %q", got, want)
	}
}

// TestConvertExternalDNSDisable covers spec.externalDNS.disable, which v1alpha3 spells
// as a "none" provider.
func TestConvertExternalDNSDisable(t *testing.T) {
	internal := toInternal(t, v1alpha2.ClusterSpec{ExternalDNS: &v1alpha2.ExternalDNSConfig{Disable: true}})
	if internal.ExternalDNS == nil {
		t.Fatal("expected spec.externalDNS to be populated")
	}
	if got, want := internal.ExternalDNS.Provider, kops.ExternalDNSProviderNone; got != want {
		t.Errorf("externalDNS.provider = %q; want %q", got, want)
	}

	external := toV1alpha2(t, *internal)
	if external.ExternalDNS == nil {
		t.Fatal("expected spec.externalDNS to be populated")
	}
	if !external.ExternalDNS.Disable {
		t.Error("externalDNS.disable = false; want true")
	}
	if got := external.ExternalDNS.Provider; got != "" {
		t.Errorf("externalDNS.provider = %q; want it to be cleared", got)
	}
}

// TestConvertTopologyDNS covers spec.topology.dns, which v1alpha2 nests one level deeper.
func TestConvertTopologyDNS(t *testing.T) {
	internal := kops.ClusterSpec{
		Networking: kops.NetworkingSpec{Topology: &kops.TopologySpec{DNS: kops.DNSTypePrivate}},
	}

	external := toV1alpha2(t, internal)
	if external.Topology == nil || external.Topology.LegacyDNS == nil {
		t.Fatal("expected topology.dns to be set")
	}
	if got, want := external.Topology.LegacyDNS.Type, v1alpha2.DNSType(kops.DNSTypePrivate); got != want {
		t.Errorf("topology.dns.type = %q; want %q", got, want)
	}

	back := toInternal(t, *external)
	if back.Networking.Topology == nil {
		t.Fatal("expected topology to be set")
	}
	if got, want := back.Networking.Topology.DNS, kops.DNSTypePrivate; got != want {
		t.Errorf("topology.dns = %q; want %q", got, want)
	}
}
