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

package validation

import (
	"net"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/utils/ptr"
)

func Test_Validate_DNS(t *testing.T) {
	for _, name := range []string{"test.-", "!", "-"} {
		errs := validation.IsDNS1123Subdomain(name)
		if len(errs) == 0 {
			t.Fatalf("Expected errors validating name %q", name)
		}
	}
}

func TestValidateCIDR(t *testing.T) {
	grid := []struct {
		Input          string
		ExpectedErrors []string
		ExpectedDetail string
	}{
		{
			Input: "192.168.0.1/32",
		},
		{
			Input:          "192.168.0.1",
			ExpectedErrors: []string{"Invalid value::CIDR"},
			ExpectedDetail: "Could not be parsed as a CIDR (did you mean \"192.168.0.1/32\")",
		},
		{
			Input:          "10.128.0.0/8",
			ExpectedErrors: []string{"Invalid value::CIDR"},
			ExpectedDetail: "Network contains bits outside prefix (did you mean \"10.0.0.0/8\")",
		},
		{
			Input:          "",
			ExpectedErrors: []string{"Invalid value::CIDR"},
		},
		{
			Input:          "invalid.example.com",
			ExpectedErrors: []string{"Invalid value::CIDR"},
			ExpectedDetail: "Could not be parsed as a CIDR",
		},
	}
	for _, g := range grid {
		errs := validateCIDR(field.NewPath("CIDR"), g.Input)

		testErrors(t, g.Input, errs, g.ExpectedErrors)

		if g.ExpectedDetail != "" {
			found := false
			for _, err := range errs {
				if err.Detail == g.ExpectedDetail {
					found = true
				}
			}
			if !found {
				for _, err := range errs {
					t.Logf("found detail: %q", err.Detail)
				}

				t.Errorf("did not find expected error %q", g.ExpectedDetail)
			}
		}
	}
}

func testErrors(t *testing.T, context interface{}, actual field.ErrorList, expectedErrors []string) {
	t.Helper()
	if len(expectedErrors) == 0 {
		if len(actual) != 0 {
			t.Errorf("unexpected errors from %q: %v", context, actual)
		}
	} else {
		errStrings := sets.NewString()
		for _, err := range actual {
			errStrings.Insert(err.Type.String() + "::" + err.Field)
		}

		for _, expected := range expectedErrors {
			if !errStrings.Has(expected) {
				t.Errorf("expected error %q from %v, was not found in %q", expected, context, errStrings.List())
			}
		}
	}
}

func TestValidateKubeletTaints(t *testing.T) {
	grid := []struct {
		taints   []string
		expected []string
	}{
		{
			taints: []string{
				"dedicated=gpu:NoSchedule",
				"spot:PreferNoSchedule",
				"drain:NoExecute",
			},
		},
		{
			taints: []string{
				"dedicated=gpu:ScheduleSometimes",
			},
			expected: []string{"Invalid value::spec.kubelet.taints[0]"},
		},
	}

	for _, g := range grid {
		errs := validateKubelet(&kops.KubeletConfigSpec{Taints: g.taints}, &kops.Cluster{}, field.NewPath("spec", "kubelet"))
		testErrors(t, g.taints, errs, g.expected)
	}
}

func TestValidateSubnets(t *testing.T) {
	grid := []struct {
		Input          []kops.ClusterSubnetSpec
		ExpectedErrors []string
	}{
		{
			Input: []kops.ClusterSubnetSpec{
				{Name: "a", Type: kops.SubnetTypePublic},
			},
		},
		{
			Input: []kops.ClusterSubnetSpec{
				{Name: "a", CIDR: "10.0.0.0/8", Type: kops.SubnetTypePublic},
			},
		},
		{
			Input: []kops.ClusterSubnetSpec{
				{Name: "", Type: kops.SubnetTypePublic},
			},
			ExpectedErrors: []string{"Required value::subnets[0].name"},
		},
		{
			Input: []kops.ClusterSubnetSpec{
				{Name: "a", Type: kops.SubnetTypePublic},
				{Name: "a", Type: kops.SubnetTypePublic},
			},
			ExpectedErrors: []string{"Duplicate value::subnets[1].name"},
		},
		{
			Input: []kops.ClusterSubnetSpec{
				{Name: "a", ID: "a", Type: kops.SubnetTypePublic},
				{Name: "b", ID: "b", Type: kops.SubnetTypePublic},
			},
		},
		{
			Input: []kops.ClusterSubnetSpec{
				{Name: "a", ID: "a", Type: kops.SubnetTypePublic},
				{Name: "b", ID: "", Type: kops.SubnetTypePublic},
			},
			ExpectedErrors: []string{"Forbidden::subnets[1].id"},
		},
		{
			Input: []kops.ClusterSubnetSpec{
				{Name: "a", CIDR: "10.128.0.0/8", Type: kops.SubnetTypePublic},
			},
			ExpectedErrors: []string{"Invalid value::subnets[0].cidr"},
		},
		{
			Input: []kops.ClusterSubnetSpec{
				{Name: "a", IPv6CIDR: "2001:db8::/56", Type: kops.SubnetTypePublic},
			},
		},
		{
			Input: []kops.ClusterSubnetSpec{
				{Name: "a", IPv6CIDR: "10.0.0.0/8", Type: kops.SubnetTypePublic},
			},
			ExpectedErrors: []string{"Invalid value::subnets[0].ipv6CIDR"},
		},
		{
			Input: []kops.ClusterSubnetSpec{
				{Name: "a", IPv6CIDR: "::ffff:10.128.0.0", Type: kops.SubnetTypePublic},
			},
			ExpectedErrors: []string{"Invalid value::subnets[0].ipv6CIDR"},
		},
		{
			Input: []kops.ClusterSubnetSpec{
				{Name: "a", IPv6CIDR: "::ffff:10.128.0.0/8", Type: kops.SubnetTypePublic},
			},
			ExpectedErrors: []string{"Invalid value::subnets[0].ipv6CIDR"},
		},
		{
			Input: []kops.ClusterSubnetSpec{
				{Name: "a", CIDR: "::ffff:10.128.0.0/8", Type: kops.SubnetTypePublic},
			},
			ExpectedErrors: []string{"Invalid value::subnets[0].cidr"},
		},
	}
	for _, g := range grid {
		cluster := &kops.Cluster{}
		cluster.Spec = kops.ClusterSpec{
			CloudProvider: kops.CloudProviderSpec{
				AWS: &kops.AWSSpec{},
			},
			Networking: kops.NetworkingSpec{
				NetworkCIDR: "10.0.0.0/8",
				Subnets:     g.Input,
			},
		}
		_, ipNet, _ := net.ParseCIDR(cluster.Spec.Networking.NetworkCIDR)
		errs := validateSubnets(cluster, cluster.Spec.Networking.Subnets, field.NewPath("subnets"), true, &cloudProviderConstraints{}, []*net.IPNet{ipNet}, nil, nil)

		testErrors(t, g.Input, errs, g.ExpectedErrors)
	}
}

func TestValidateKubeAPIServer(t *testing.T) {
	str := "foobar"
	authzMode := "RBAC,Webhook"

	grid := []struct {
		Input          kops.KubeAPIServerConfig
		Cluster        *kops.Cluster
		ExpectedErrors []string
		ExpectedDetail string
	}{
		{
			Input: kops.KubeAPIServerConfig{
				ProxyClientCertFile: &str,
			},
			ExpectedErrors: []string{
				"Forbidden::KubeAPIServer",
			},
			ExpectedDetail: "proxyClientCertFile and proxyClientKeyFile must both be specified (or neither)",
		},
		{
			Input: kops.KubeAPIServerConfig{
				ProxyClientKeyFile: &str,
			},
			ExpectedErrors: []string{
				"Forbidden::KubeAPIServer",
			},
			ExpectedDetail: "proxyClientCertFile and proxyClientKeyFile must both be specified (or neither)",
		},
		{
			Input: kops.KubeAPIServerConfig{
				ServiceNodePortRange: str,
			},
			ExpectedErrors: []string{
				"Invalid value::KubeAPIServer.serviceNodePortRange",
			},
		},
		{
			Input: kops.KubeAPIServerConfig{
				AuthorizationMode: &authzMode,
			},
			ExpectedErrors: []string{
				"Required value::KubeAPIServer.authorizationWebhookConfigFile",
			},
			ExpectedDetail: "Authorization mode Webhook requires authorizationWebhookConfigFile to be specified",
		},
		{
			Input: kops.KubeAPIServerConfig{
				AuthorizationMode: new("RBAC"),
			},
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					Authorization: &kops.AuthorizationSpec{
						RBAC: &kops.RBACAuthorizationSpec{},
					},
					KubernetesVersion: "1.35.0",
					CloudProvider: kops.CloudProviderSpec{
						AWS: &kops.AWSSpec{},
					},
				},
			},
			ExpectedErrors: []string{
				"Required value::KubeAPIServer.authorizationMode",
			},
		},
		{
			Input: kops.KubeAPIServerConfig{
				AuthorizationMode: new("RBAC,Node"),
			},
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					Authorization: &kops.AuthorizationSpec{
						RBAC: &kops.RBACAuthorizationSpec{},
					},
					KubernetesVersion: "1.35.0",
					CloudProvider: kops.CloudProviderSpec{
						AWS: &kops.AWSSpec{},
					},
				},
			},
		},
		{
			Input: kops.KubeAPIServerConfig{
				AuthorizationMode: new("RBAC,Node,Bogus"),
			},
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					Authorization: &kops.AuthorizationSpec{
						RBAC: &kops.RBACAuthorizationSpec{},
					},
					KubernetesVersion: "1.35.0",
					CloudProvider: kops.CloudProviderSpec{
						AWS: &kops.AWSSpec{},
					},
				},
			},
			ExpectedErrors: []string{
				"Unsupported value::KubeAPIServer.authorizationMode",
			},
		},
		{
			Input: kops.KubeAPIServerConfig{
				LogFormat: "no-json",
			},
			ExpectedErrors: []string{"Unsupported value::KubeAPIServer.logFormat"},
		},
		{
			Input: kops.KubeAPIServerConfig{
				AuthenticationConfigFile: "/foo/bar",
			},
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					Authentication: &kops.AuthenticationSpec{
						OIDC: &kops.OIDCAuthenticationSpec{
							ClientID: new("foo"),
						},
					},
				},
			},
			ExpectedErrors: []string{"Forbidden::KubeAPIServer.authenticationConfigFile"},
		},
	}
	for _, g := range grid {
		if g.Cluster == nil {
			g.Cluster = &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "1.35.0",
				},
			}
		}
		errs := validateKubeAPIServer(&g.Input, g.Cluster, field.NewPath("KubeAPIServer"), true)

		testErrors(t, g.Input, errs, g.ExpectedErrors)

		if g.ExpectedDetail != "" {
			found := false
			for _, err := range errs {
				if err.Detail == g.ExpectedDetail {
					found = true
				}
			}
			if !found {
				for _, err := range errs {
					t.Logf("found detail: %q", err.Detail)
				}

				t.Errorf("did not find expected error %q", g.ExpectedDetail)
			}
		}
	}
}

func TestValidateKubeControllermanager(t *testing.T) {
	grid := []struct {
		Input          kops.KubeControllerManagerConfig
		Cluster        *kops.Cluster
		ExpectedErrors []string
		ExpectedDetail string
	}{
		{
			Input: kops.KubeControllerManagerConfig{
				ExperimentalClusterSigningDuration: &metav1.Duration{Duration: time.Hour},
			},
			ExpectedErrors: []string{
				"Forbidden::kubeControllerManager.experimentalClusterSigningDuration",
			},
			ExpectedDetail: "experimentalClusterSigningDuration has been replaced with clusterSigningDuration as of kubernetes 1.25",
		},
	}
	for _, g := range grid {
		if g.Cluster == nil {
			g.Cluster = &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "1.35.0",
				},
			}
		}
		errs := validateKubeControllerManager(&g.Input, g.Cluster, field.NewPath("kubeControllerManager"), true)

		testErrors(t, g.Input, errs, g.ExpectedErrors)

		if g.ExpectedDetail != "" {
			found := false
			for _, err := range errs {
				if err.Detail == g.ExpectedDetail {
					found = true
				}
			}
			if !found {
				for _, err := range errs {
					t.Logf("found detail: %q", err.Detail)
				}

				t.Errorf("did not find expected error %q", g.ExpectedDetail)
			}
		}
	}
}

func Test_Validate_Networking_Flannel(t *testing.T) {
	grid := []struct {
		Input          kops.FlannelNetworkingSpec
		ExpectedErrors []string
	}{
		{
			Input: kops.FlannelNetworkingSpec{
				Backend: "udp",
			},
		},
		{
			Input: kops.FlannelNetworkingSpec{
				Backend: "vxlan",
			},
		},
		{
			Input: kops.FlannelNetworkingSpec{
				Backend: "",
			},
			ExpectedErrors: []string{"Required value::networking.flannel.backend"},
		},
		{
			Input: kops.FlannelNetworkingSpec{
				Backend: "nope",
			},
			ExpectedErrors: []string{"Unsupported value::networking.flannel.backend"},
		},
	}
	for _, g := range grid {
		cluster := &kops.Cluster{
			Spec: kops.ClusterSpec{
				KubernetesVersion: "1.35.0",
				Networking: kops.NetworkingSpec{
					NetworkCIDR:           "10.0.0.0/8",
					NonMasqueradeCIDR:     "100.64.0.0/10",
					PodCIDR:               "100.96.0.0/11",
					ServiceClusterIPRange: "100.64.0.0/13",
					Subnets: []kops.ClusterSubnetSpec{
						{
							Name: "sg-test",
							CIDR: "10.11.0.0/16",
							Type: "Public",
						},
					},
					Flannel: &g.Input,
				},
			},
		}

		errs := validateNetworking(cluster, &cluster.Spec.Networking, field.NewPath("networking"), true, &cloudProviderConstraints{})
		testErrors(t, g.Input, errs, g.ExpectedErrors)
	}
}

func Test_Validate_Networking_Kindnet(t *testing.T) {
	grid := []struct {
		Input          kops.KindnetNetworkingSpec
		ExpectedErrors []string
	}{
		{
			Input: kops.KindnetNetworkingSpec{
				Masquerade: &kops.KindnetMasqueradeSpec{
					Enabled: ptr.To(true),
				},
			},
		},
		{
			Input: kops.KindnetNetworkingSpec{
				Masquerade: &kops.KindnetMasqueradeSpec{
					Enabled:            ptr.To(true),
					NonMasqueradeCIDRs: []string{"10.0.0.0/24", "2001:db8::/64"},
				},
			},
		},
		{
			Input: kops.KindnetNetworkingSpec{
				Masquerade: &kops.KindnetMasqueradeSpec{
					Enabled:            ptr.To(true),
					NonMasqueradeCIDRs: []string{"a.b.c.d/24", "2001:db8::/64"},
				},
			},
			ExpectedErrors: []string{"Invalid value::networking.kindnet"},
		},
		{
			Input: kops.KindnetNetworkingSpec{
				Masquerade: &kops.KindnetMasqueradeSpec{
					Enabled:            ptr.To(false),
					NonMasqueradeCIDRs: []string{"a.b.c.d/24", "2001:db8::/64"},
				},
			},
			ExpectedErrors: []string{},
		},
	}

	for _, g := range grid {
		cluster := &kops.Cluster{
			Spec: kops.ClusterSpec{
				KubernetesVersion: "1.35.0",
				Networking: kops.NetworkingSpec{
					NetworkCIDR:           "10.0.0.0/8",
					NonMasqueradeCIDR:     "100.64.0.0/10",
					PodCIDR:               "100.96.0.0/11",
					ServiceClusterIPRange: "100.64.0.0/13",
					Subnets: []kops.ClusterSubnetSpec{
						{
							Name: "sg-test",
							CIDR: "10.11.0.0/16",
							Type: "Public",
						},
					},
					Kindnet: &g.Input,
				},
			},
		}

		errs := validateNetworking(cluster, &cluster.Spec.Networking, field.NewPath("networking"), true, &cloudProviderConstraints{})
		testErrors(t, g.Input, errs, g.ExpectedErrors)
	}
}

func Test_Validate_Networking_OverlappingCIDR(t *testing.T) {
	grid := []struct {
		Name           string
		Networking     kops.NetworkingSpec
		ExpectedErrors []*field.Error
	}{
		{
			Name: "no-overlap",
			Networking: kops.NetworkingSpec{
				NetworkCIDR:           "10.0.0.0/8",
				NonMasqueradeCIDR:     "100.64.0.0/10",
				PodCIDR:               "100.64.10.0/24",
				ServiceClusterIPRange: "100.64.20.0/24",
				Subnets: []kops.ClusterSubnetSpec{
					{
						Name: "subnet-test",
						CIDR: "10.10.0.0/16",
						Type: "Public",
					},
				},
			},
		},
		{
			Name: "overlap-podcidr-and-servicecidr",
			Networking: kops.NetworkingSpec{
				NetworkCIDR:           "10.0.0.0/8",
				NonMasqueradeCIDR:     "100.64.0.0/10",
				PodCIDR:               "100.64.0.0/10",
				ServiceClusterIPRange: "100.64.0.0/13",
				Subnets: []kops.ClusterSubnetSpec{
					{
						Name: "subnet-test",
						CIDR: "10.10.0.0/16",
						Type: "Public",
					},
				},
			},
		},
		{
			Name: "overlap-servicecidr-and-subnetcidr",
			Networking: kops.NetworkingSpec{
				NetworkCIDR:           "10.0.0.0/8",
				NonMasqueradeCIDR:     "100.64.0.0/10",
				PodCIDR:               "100.64.10.0/24",
				ServiceClusterIPRange: "100.64.20.0/24",
				Subnets: []kops.ClusterSubnetSpec{
					{
						Name: "subnet-test",
						CIDR: "100.64.20.0/28",
						Type: "Public",
					},
				},
			},
			ExpectedErrors: []*field.Error{
				{
					Type:   field.ErrorTypeForbidden,
					Detail: `subnet "subnet-test" cidr "100.64.20.0/28" is not a subnet of the networkCIDR "10.0.0.0/8"`,
					Field:  "networking.subnets[0].cidr",
				},
				{
					Type:   field.ErrorTypeForbidden,
					Detail: `subnet "subnet-test" cidr "100.64.20.0/28" must not overlap serviceClusterIPRange "100.64.20.0/24"`,
					Field:  "networking.subnets[0].cidr",
				},
			},
		},
	}
	for _, g := range grid {
		t.Run(g.Name, func(t *testing.T) {
			cluster := &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "1.35.0",
				},
			}
			cluster.Spec.Networking = g.Networking

			errs := validateNetworking(cluster, &cluster.Spec.Networking, field.NewPath("networking"), true, &cloudProviderConstraints{})
			testFieldErrors(t, errs, g.ExpectedErrors)
		})
	}
}

func testFieldErrors(t *testing.T, actual field.ErrorList, expectedErrors []*field.Error) {
	t.Helper()

	if len(actual) > len(expectedErrors) {
		t.Errorf("found unexpected errors: %+v", actual)
	}

	for _, expected := range expectedErrors {
		found := false
		for _, err := range actual {
			if expected.Type != "" && expected.Type != err.Type {
				continue
			}
			if expected.Detail != "" && expected.Detail != err.Detail {
				continue
			}
			if expected.Field != "" && expected.Field != err.Field {
				continue
			}
			found = true
		}
		if !found {
			t.Errorf("expected error %+v, was not found in errors: %+v", expected, actual)
		}
	}
}

func Test_Validate_AdditionalPolicies(t *testing.T) {
	grid := []struct {
		Input          map[string]string
		ExpectedErrors []string
	}{
		{
			Input: map[string]string{},
		},
		{
			Input: map[string]string{
				"control-plane": `[ { "Action": [ "s3:GetObject" ], "Resource": [ "*" ], "Effect": "Allow" } ]`,
			},
		},
		{
			Input: map[string]string{
				"notarole": `[ { "Action": [ "s3:GetObject" ], "Resource": [ "*" ], "Effect": "Allow" } ]`,
			},
			ExpectedErrors: []string{"Unsupported value::spec.additionalPolicies"},
		},
		{
			Input: map[string]string{
				"control-plane": `badjson`,
			},
			ExpectedErrors: []string{"Invalid value::spec.additionalPolicies[control-plane]"},
		},
		{
			Input: map[string]string{
				"control-plane": `[ { "Action": [ "s3:GetObject" ], "Resource": [ "*" ] } ]`,
			},
			ExpectedErrors: []string{"Required value::spec.additionalPolicies[control-plane][0].Effect"},
		},
		{
			Input: map[string]string{
				"control-plane": `[ { "Action": [ "s3:GetObject" ], "Resource": [ "*" ], "Effect": "allow" } ]`,
			},
			ExpectedErrors: []string{"Unsupported value::spec.additionalPolicies[control-plane][0].Effect"},
		},
	}
	for _, g := range grid {
		clusterSpec := &kops.ClusterSpec{
			KubernetesVersion:  "1.35.0",
			AdditionalPolicies: g.Input,
			CloudProvider: kops.CloudProviderSpec{
				AWS: &kops.AWSSpec{},
			},
			Networking: kops.NetworkingSpec{
				NetworkCIDR:           "10.10.0.0/16",
				NonMasqueradeCIDR:     "100.64.0.0/10",
				PodCIDR:               "100.96.0.0/11",
				ServiceClusterIPRange: "100.64.0.0/13",
				Subnets: []kops.ClusterSubnetSpec{
					{
						Name: "subnet1",
						Type: kops.SubnetTypePublic,
						CIDR: "10.10.10.0/24",
					},
				},
			},
			EtcdClusters: []kops.EtcdClusterSpec{
				{
					Name: "main",
					Members: []kops.EtcdMemberSpec{
						{
							Name:          "us-test-1a",
							InstanceGroup: new("master-us-test-1a"),
						},
					},
				},
			},
		}
		errs := validateClusterSpec(clusterSpec, &kops.Cluster{Spec: *clusterSpec}, field.NewPath("spec"), true)
		testErrors(t, g.Input, errs, g.ExpectedErrors)
	}
}

func Test_Validate_Addons(t *testing.T) {
	grid := []struct {
		Input          []kops.AddonSpec
		ExpectedErrors []string
	}{
		{},
		{
			Input: []kops.AddonSpec{{Manifest: "s3://somebucket/example.yaml"}},
		},
		{
			Input:          []kops.AddonSpec{{Manifest: "file:///etc/kubernetes/kops/config/addons/extra.yaml"}},
			ExpectedErrors: []string{"Invalid value::spec.addons[0].manifest"},
		},
	}
	for _, g := range grid {
		clusterSpec := &kops.ClusterSpec{
			KubernetesVersion: "1.35.0",
			Addons:            g.Input,
			CloudProvider: kops.CloudProviderSpec{
				AWS: &kops.AWSSpec{},
			},
			Networking: kops.NetworkingSpec{
				NetworkCIDR:           "10.10.0.0/16",
				NonMasqueradeCIDR:     "100.64.0.0/10",
				PodCIDR:               "100.96.0.0/11",
				ServiceClusterIPRange: "100.64.0.0/13",
				Subnets: []kops.ClusterSubnetSpec{
					{Name: "subnet1", Type: kops.SubnetTypePublic, CIDR: "10.10.10.0/24"},
				},
			},
			EtcdClusters: []kops.EtcdClusterSpec{
				{
					Name: "main",
					Members: []kops.EtcdMemberSpec{
						{Name: "us-test-1a", InstanceGroup: new("master-us-test-1a")},
					},
				},
			},
		}
		errs := validateClusterSpec(clusterSpec, &kops.Cluster{Spec: *clusterSpec}, field.NewPath("spec"), true)
		testErrors(t, g.Input, errs, g.ExpectedErrors)
	}
}

type caliInput struct {
	Cluster *kops.ClusterSpec
	Calico  *kops.CalicoNetworkingSpec
}

func Test_Validate_Calico(t *testing.T) {
	grid := []struct {
		Description    string
		Input          caliInput
		ExpectedErrors []string
	}{
		{
			Description: "empty specs",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{},
			},
		},
		{
			Description: "positive Typha replica count",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					TyphaReplicas: 3,
				},
			},
		},
		{
			Description: "negative Typha replica count",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					TyphaReplicas: -1,
				},
			},
			ExpectedErrors: []string{"Invalid value::calico.typhaReplicas"},
		},
		{
			Description: "with etcd version",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{},
			},
		},
		{
			Description: "IPv4 autodetection method",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					IPv4AutoDetectionMethod: "first-found",
				},
			},
		},
		{
			Description: "IPv6 autodetection method",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					IPv6AutoDetectionMethod: "first-found",
				},
			},
		},
		{
			Description: "IPv4 autodetection method with parameter",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					IPv4AutoDetectionMethod: "can-reach=8.8.8.8",
				},
			},
		},
		{
			Description: "IPv6 autodetection method with parameter",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					IPv6AutoDetectionMethod: "can-reach=2001:4860:4860::8888",
				},
			},
		},
		{
			Description: "invalid IPv4 autodetection method",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					IPv4AutoDetectionMethod: "bogus",
				},
			},
			ExpectedErrors: []string{"Invalid value::calico.ipv4AutoDetectionMethod"},
		},
		{
			Description: "invalid IPv6 autodetection method",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					IPv6AutoDetectionMethod: "bogus",
				},
			},
			ExpectedErrors: []string{"Invalid value::calico.ipv6AutoDetectionMethod"},
		},
		{
			Description: "invalid IPv6 autodetection method missing parameter",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					IPv6AutoDetectionMethod: "interface=",
				},
			},
			ExpectedErrors: []string{"Invalid value::calico.ipv6AutoDetectionMethod"},
		},
		{
			Description: "IPv4 autodetection method with parameter list",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					IPv4AutoDetectionMethod: "interface=en.*,eth0",
				},
			},
		},
		{
			Description: "IPv6 autodetection method with parameter list",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					IPv6AutoDetectionMethod: "skip-interface=en.*,eth0",
				},
			},
		},
		{
			Description: "invalid IPv4 autodetection method parameter (parenthesis)",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					IPv4AutoDetectionMethod: "interface=(,en1",
				},
			},
			ExpectedErrors: []string{"Invalid value::calico.ipv4AutoDetectionMethod"},
		},
		{
			Description: "invalid IPv4 autodetection method parameter (equals)",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					IPv4AutoDetectionMethod: "interface=foo=bar",
				},
			},
			ExpectedErrors: []string{"Invalid value::calico.ipv4AutoDetectionMethod"},
		},
		{
			Description: "invalid IPv4 autodetection method parameter with no name",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					IPv4AutoDetectionMethod: "=en0,eth.*",
				},
			},
			ExpectedErrors: []string{"Invalid value::calico.ipv4AutoDetectionMethod"},
		},
		{
			Description: "AWS source/destination checks off",
			Input: caliInput{
				Cluster: &kops.ClusterSpec{},
				Calico: &kops.CalicoNetworkingSpec{
					AWSSrcDstCheck: "off",
				},
			},
			ExpectedErrors: []string{"Unsupported value::calico.awsSrcDstCheck"},
		},
		{
			Description: "AWS source/destination checks enabled",
			Input: caliInput{
				Cluster: &kops.ClusterSpec{},
				Calico: &kops.CalicoNetworkingSpec{
					AWSSrcDstCheck: "Enable",
				},
			},
		},
		{
			Description: "AWS source/destination checks disabled",
			Input: caliInput{
				Cluster: &kops.ClusterSpec{},
				Calico: &kops.CalicoNetworkingSpec{
					AWSSrcDstCheck: "Disable",
				},
			},
		},
		{
			Description: "AWS source/destination checks left as is",
			Input: caliInput{
				Cluster: &kops.ClusterSpec{},
				Calico: &kops.CalicoNetworkingSpec{
					AWSSrcDstCheck: "DoNothing",
				},
			},
		},
		{
			Description: "encapsulation none with IPv4",
			Input: caliInput{
				Cluster: &kops.ClusterSpec{
					Networking: kops.NetworkingSpec{
						NonMasqueradeCIDR: "100.64.0.0/10",
					},
				},
				Calico: &kops.CalicoNetworkingSpec{
					EncapsulationMode: "none",
				},
			},
			ExpectedErrors: []string{"Forbidden::calico.encapsulationMode"},
		},
		{
			Description: "encapsulation mode IPIP for IPv6",
			Input: caliInput{
				Cluster: &kops.ClusterSpec{
					Networking: kops.NetworkingSpec{
						NonMasqueradeCIDR: "::/0",
					},
				},
				Calico: &kops.CalicoNetworkingSpec{
					EncapsulationMode: "ipip",
				},
			},
			ExpectedErrors: []string{"Forbidden::calico.encapsulationMode"},
		},
		{
			Description: "encapsulation mode VXLAN for IPv6",
			Input: caliInput{
				Cluster: &kops.ClusterSpec{
					Networking: kops.NetworkingSpec{
						NonMasqueradeCIDR: "::/0",
					},
				},
				Calico: &kops.CalicoNetworkingSpec{
					EncapsulationMode: "vxlan",
				},
			},
			ExpectedErrors: []string{"Forbidden::calico.encapsulationMode"},
		},
		{
			Description: "unknown Calico IPIP mode",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					IPIPMode: "unknown",
				},
			},
			ExpectedErrors: []string{"Unsupported value::calico.ipipMode"},
		},
		// You can't use per-IPPool IP-in-IP encapsulation unless you're using the "ipip"
		// encapsulation mode.
		{
			Description: "Calico IPIP encapsulation mode (implicit) with IPIP IPPool mode (always)",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					IPIPMode: "Always",
				},
			},
		},
		{
			Description: "Calico IPIP encapsulation mode (implicit) with IPIP IPPool mode (cross-subnet)",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					IPIPMode: "CrossSubnet",
				},
			},
		},
		{
			Description: "Calico IPIP encapsulation mode (implicit) with IPIP IPPool mode (never)",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					IPIPMode: "Never",
				},
			},
		},
		{
			Description: "Calico IPIP encapsulation mode (explicit) with IPIP IPPool mode (always)",
			Input: caliInput{
				Cluster: &kops.ClusterSpec{
					Networking: kops.NetworkingSpec{
						NonMasqueradeCIDR: "100.64.0.0/10",
					},
				},
				Calico: &kops.CalicoNetworkingSpec{
					EncapsulationMode: "ipip",
					IPIPMode:          "Always",
				},
			},
		},
		{
			Description: "Calico IPIP encapsulation mode (explicit) with IPIP IPPool mode (cross-subnet)",
			Input: caliInput{
				Cluster: &kops.ClusterSpec{
					Networking: kops.NetworkingSpec{
						NonMasqueradeCIDR: "100.64.0.0/10",
					},
				},
				Calico: &kops.CalicoNetworkingSpec{
					EncapsulationMode: "ipip",
					IPIPMode:          "CrossSubnet",
				},
			},
		},
		{
			Description: "Calico IPIP encapsulation mode (explicit) with IPIP IPPool mode (never)",
			Input: caliInput{
				Cluster: &kops.ClusterSpec{
					Networking: kops.NetworkingSpec{
						NonMasqueradeCIDR: "100.64.0.0/10",
					},
				},
				Calico: &kops.CalicoNetworkingSpec{
					EncapsulationMode: "ipip",
					IPIPMode:          "Never",
				},
			},
		},
		{
			Description: "Calico VXLAN encapsulation mode with IPIP IPPool mode",
			Input: caliInput{
				Cluster: &kops.ClusterSpec{
					Networking: kops.NetworkingSpec{
						NonMasqueradeCIDR: "100.64.0.0/10",
					},
				},
				Calico: &kops.CalicoNetworkingSpec{
					EncapsulationMode: "vxlan",
					IPIPMode:          "Always",
				},
			},
			ExpectedErrors: []string{`Forbidden::calico.ipipMode`},
		},
		{
			Description: "Calico VXLAN encapsulation mode with IPIP IPPool mode (always)",
			Input: caliInput{
				Cluster: &kops.ClusterSpec{
					Networking: kops.NetworkingSpec{
						NonMasqueradeCIDR: "100.64.0.0/10",
					},
				},
				Calico: &kops.CalicoNetworkingSpec{
					EncapsulationMode: "vxlan",
					IPIPMode:          "Always",
				},
			},
			ExpectedErrors: []string{`Forbidden::calico.ipipMode`},
		},
		{
			Description: "Calico VXLAN encapsulation mode with IPIP IPPool mode (cross-subnet)",
			Input: caliInput{
				Cluster: &kops.ClusterSpec{
					Networking: kops.NetworkingSpec{
						NonMasqueradeCIDR: "100.64.0.0/10",
					},
				},
				Calico: &kops.CalicoNetworkingSpec{
					EncapsulationMode: "vxlan",
					IPIPMode:          "CrossSubnet",
				},
			},
			ExpectedErrors: []string{`Forbidden::calico.ipipMode`},
		},
		{
			Description: "Calico VXLAN encapsulation mode with IPIP IPPool mode (never)",
			Input: caliInput{
				Cluster: &kops.ClusterSpec{
					Networking: kops.NetworkingSpec{
						NonMasqueradeCIDR: "100.64.0.0/10",
					},
				},
				Calico: &kops.CalicoNetworkingSpec{
					EncapsulationMode: "vxlan",
					IPIPMode:          "Never",
				},
			},
		},
		{
			Description: "Calico IPv6 without encapsulation",
			Input: caliInput{
				Cluster: &kops.ClusterSpec{
					Networking: kops.NetworkingSpec{
						NonMasqueradeCIDR: "::/0",
					},
				},
				Calico: &kops.CalicoNetworkingSpec{
					EncapsulationMode: "none",
					IPIPMode:          "Never",
					VXLANMode:         "Never",
				},
			},
		},
		{
			Description: "Calico BPF with kube-proxy explicitly disabled",
			Input: caliInput{
				Cluster: &kops.ClusterSpec{
					KubeProxy: &kops.KubeProxyConfig{Enabled: new(false)},
				},
				Calico: &kops.CalicoNetworkingSpec{BPFEnabled: true},
			},
		},
		{
			Description: "Calico BPF with kube-proxy implicitly enabled",
			Input: caliInput{
				Cluster: &kops.ClusterSpec{
					KubeProxy: &kops.KubeProxyConfig{},
				},
				Calico: &kops.CalicoNetworkingSpec{BPFEnabled: true},
			},
			ExpectedErrors: []string{"Forbidden::spec.kubeProxy.enabled"},
		},
		{
			Description: "Calico BPF with kube-proxy explicitly enabled",
			Input: caliInput{
				Cluster: &kops.ClusterSpec{
					KubeProxy: &kops.KubeProxyConfig{Enabled: new(true)},
				},
				Calico: &kops.CalicoNetworkingSpec{BPFEnabled: true},
			},
			ExpectedErrors: []string{"Forbidden::spec.kubeProxy.enabled"},
		},
	}
	rootFieldPath := field.NewPath("calico")
	for _, g := range grid {
		t.Run(g.Description, func(t *testing.T) {
			errs := validateNetworkingCalico(g.Input.Cluster, g.Input.Calico, rootFieldPath)
			testErrors(t, g.Input, errs, g.ExpectedErrors)
		})
	}
}

func Test_Validate_Cilium(t *testing.T) {
	grid := []struct {
		Cilium         kops.CiliumNetworkingSpec
		Spec           kops.ClusterSpec
		ExpectedErrors []string
	}{
		{
			Cilium: kops.CiliumNetworkingSpec{},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				IPAM: "crd",
			},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				IPAM: "eni",
			},
			Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					AWS: &kops.AWSSpec{},
				},
			},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				ClusterID: 253,
			},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				Masquerade: new(true),
				IPAM:       "eni",
			},
			Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					AWS: &kops.AWSSpec{},
				},
			},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				IPAM: "foo",
			},
			ExpectedErrors: []string{"Unsupported value::cilium.ipam"},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				Masquerade: new(false),
				IPAM:       "eni",
			},
			Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					AWS: &kops.AWSSpec{},
				},
			},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				EnableL7Proxy:        new(true),
				InstallIptablesRules: new(false),
			},
			Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					AWS: &kops.AWSSpec{},
				},
			},
			ExpectedErrors: []string{"Forbidden::cilium.enableL7Proxy"},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				IPAM: "eni",
			},
			Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					GCE: &kops.GCESpec{},
				},
			},
			ExpectedErrors: []string{"Forbidden::cilium.ipam"},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				IdentityAllocationMode: "kvstore",
			},
			ExpectedErrors: []string{"Forbidden::cilium.identityAllocationMode"},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				Version: "v1.0.0",
			},
			Spec: kops.ClusterSpec{
				KubernetesVersion: "1.35.0",
			},
			ExpectedErrors: []string{"Invalid value::cilium.version"},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				Version: "1.7.0",
			},
			ExpectedErrors: []string{"Invalid value::cilium.version"},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				Version: "v1.8.0",
				Hubble: &kops.HubbleSpec{
					Enabled: new(true),
				},
			},
			ExpectedErrors: []string{"Forbidden::cilium.hubble.enabled"},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				Version: "v1.18.0",
				Ingress: &kops.CiliumIngressSpec{
					Enabled:                 new(true),
					DefaultLoadBalancerMode: "bad-value",
				},
			},
			ExpectedErrors: []string{"Unsupported value::cilium.ingress.defaultLoadBalancerMode"},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				Version: "v1.18.0",
				Ingress: &kops.CiliumIngressSpec{
					Enabled:                 new(true),
					DefaultLoadBalancerMode: "dedicated",
				},
			},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				Version: "v1.18.0",
				GatewayAPI: &kops.CiliumGatewayAPISpec{
					Enabled:           new(true),
					EnableSecretsSync: new(true),
				},
			},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				Version: "v1.18.0",
				Hubble: &kops.HubbleSpec{
					Enabled: new(true),
				},
			},
			Spec: kops.ClusterSpec{
				CertManager: &kops.CertManagerConfig{
					Enabled: new(true),
				},
			},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				BPFLBSock:           false,
				BPFLBSockHostNSOnly: true,
			},
			ExpectedErrors: []string{"Forbidden::cilium.bpfLBSockHostNSOnly"},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				BPFLBSock:           true,
				BPFLBSockHostNSOnly: true,
			},
		},
	}
	for _, g := range grid {
		g.Spec.Networking.Cilium = &g.Cilium
		if g.Spec.KubernetesVersion == "" {
			g.Spec.KubernetesVersion = "1.17.0"
		}
		cluster := &kops.Cluster{
			Spec: g.Spec,
		}
		errs := validateNetworkingCilium(cluster, g.Spec.Networking.Cilium, field.NewPath("cilium"))
		testErrors(t, g.Spec, errs, g.ExpectedErrors)
	}
}

func Test_Validate_RollingUpdate(t *testing.T) {
	grid := []struct {
		Input          kops.RollingUpdate
		OnMasterIG     bool
		ExpectedErrors []string
	}{
		{
			Input: kops.RollingUpdate{},
		},
		{
			Input: kops.RollingUpdate{
				MaxUnavailable: intStr(intstr.FromInt(0)),
			},
		},
		{
			Input: kops.RollingUpdate{
				MaxUnavailable: intStr(intstr.FromString("0%")),
			},
		},
		{
			Input: kops.RollingUpdate{
				MaxUnavailable: intStr(intstr.FromString("nope")),
			},
			ExpectedErrors: []string{"Invalid value::testField.maxUnavailable"},
		},
		{
			Input: kops.RollingUpdate{
				MaxUnavailable: intStr(intstr.FromInt(-1)),
			},
			ExpectedErrors: []string{"Invalid value::testField.maxUnavailable"},
		},
		{
			Input: kops.RollingUpdate{
				MaxUnavailable: intStr(intstr.FromString("-1%")),
			},
			ExpectedErrors: []string{"Invalid value::testField.maxUnavailable"},
		},
		{
			Input: kops.RollingUpdate{
				MaxSurge: intStr(intstr.FromInt(0)),
			},
		},
		{
			Input: kops.RollingUpdate{
				MaxSurge: intStr(intstr.FromString("0%")),
			},
		},
		{
			Input: kops.RollingUpdate{
				MaxSurge: intStr(intstr.FromInt(1)),
			},
		},
		{
			Input: kops.RollingUpdate{
				MaxSurge: intStr(intstr.FromString("1%")),
			},
		},
		{
			Input: kops.RollingUpdate{
				MaxSurge: intStr(intstr.FromString("nope")),
			},
			ExpectedErrors: []string{"Invalid value::testField.maxSurge"},
		},
		{
			Input: kops.RollingUpdate{
				MaxSurge: intStr(intstr.FromInt(-1)),
			},
			ExpectedErrors: []string{"Invalid value::testField.maxSurge"},
		},
		{
			Input: kops.RollingUpdate{
				MaxSurge: intStr(intstr.FromString("-1%")),
			},
			ExpectedErrors: []string{"Invalid value::testField.maxSurge"},
		},
		{
			Input: kops.RollingUpdate{
				MaxSurge: intStr(intstr.FromInt(0)),
			},
			OnMasterIG: true,
		},
		{
			Input: kops.RollingUpdate{
				MaxSurge: intStr(intstr.FromString("0%")),
			},
			OnMasterIG: true,
		},
		{
			Input: kops.RollingUpdate{
				MaxSurge: intStr(intstr.FromInt(1)),
			},
			OnMasterIG:     true,
			ExpectedErrors: []string{"Forbidden::testField.maxSurge"},
		},
		{
			Input: kops.RollingUpdate{
				MaxSurge: intStr(intstr.FromString("1%")),
			},
			OnMasterIG:     true,
			ExpectedErrors: []string{"Forbidden::testField.maxSurge"},
		},
		{
			Input: kops.RollingUpdate{
				MaxSurge: intStr(intstr.FromString("nope")),
			},
			OnMasterIG:     true,
			ExpectedErrors: []string{"Invalid value::testField.maxSurge"},
		},
		{
			Input: kops.RollingUpdate{
				MaxSurge: intStr(intstr.FromInt(-1)),
			},
			OnMasterIG:     true,
			ExpectedErrors: []string{"Forbidden::testField.maxSurge"},
		},
		{
			Input: kops.RollingUpdate{
				MaxSurge: intStr(intstr.FromString("-1%")),
			},
			OnMasterIG:     true,
			ExpectedErrors: []string{"Forbidden::testField.maxSurge"},
		},
		{
			Input: kops.RollingUpdate{
				MaxUnavailable: intStr(intstr.FromInt(0)),
				MaxSurge:       intStr(intstr.FromInt(0)),
			},
			ExpectedErrors: []string{"Forbidden::testField.maxSurge"},
		},
	}
	for _, g := range grid {
		errs := validateRollingUpdate(&g.Input, field.NewPath("testField"), g.OnMasterIG)
		testErrors(t, g.Input, errs, g.ExpectedErrors)
	}
}

func intStr(i intstr.IntOrString) *intstr.IntOrString {
	return &i
}

func Test_Validate_NodeLocalDNS(t *testing.T) {
	grid := []struct {
		Input          kops.ClusterSpec
		ExpectedErrors []string
	}{
		{
			Input: kops.ClusterSpec{
				KubeProxy: &kops.KubeProxyConfig{
					ProxyMode: "iptables",
				},
				KubeDNS: &kops.KubeDNSConfig{
					Provider: "CoreDNS",
					NodeLocalDNS: &kops.NodeLocalDNSConfig{
						Enabled: new(true),
					},
				},
			},
			ExpectedErrors: []string{},
		},
		{
			Input: kops.ClusterSpec{
				Kubelet: &kops.KubeletConfigSpec{
					ClusterDNS: "100.64.0.10",
				},
				KubeProxy: &kops.KubeProxyConfig{
					ProxyMode: "ipvs",
				},
				KubeDNS: &kops.KubeDNSConfig{
					Provider: "CoreDNS",
					NodeLocalDNS: &kops.NodeLocalDNSConfig{
						Enabled: new(true),
					},
				},
			},
			ExpectedErrors: []string{"Forbidden::spec.kubelet.clusterDNS"},
		},
		{
			Input: kops.ClusterSpec{
				Kubelet: &kops.KubeletConfigSpec{
					ClusterDNS: "100.64.0.10",
				},
				KubeProxy: &kops.KubeProxyConfig{
					ProxyMode: "ipvs",
				},
				KubeDNS: &kops.KubeDNSConfig{
					Provider: "CoreDNS",
					NodeLocalDNS: &kops.NodeLocalDNSConfig{
						Enabled: new(true),
					},
				},
				Networking: kops.NetworkingSpec{
					Cilium: &kops.CiliumNetworkingSpec{},
				},
			},
			ExpectedErrors: []string{"Forbidden::spec.kubelet.clusterDNS"},
		},
		{
			Input: kops.ClusterSpec{
				Kubelet: &kops.KubeletConfigSpec{
					ClusterDNS: "169.254.20.10",
				},
				KubeProxy: &kops.KubeProxyConfig{
					ProxyMode: "iptables",
				},
				KubeDNS: &kops.KubeDNSConfig{
					Provider: "CoreDNS",
					NodeLocalDNS: &kops.NodeLocalDNSConfig{
						Enabled: new(true),
						LocalIP: "169.254.20.10",
					},
				},
				Networking: kops.NetworkingSpec{
					Cilium: &kops.CiliumNetworkingSpec{},
				},
			},
			ExpectedErrors: []string{},
		},
	}

	for _, g := range grid {
		errs := validateNodeLocalDNS(&g.Input, field.NewPath("spec"))
		testErrors(t, g.Input, errs, g.ExpectedErrors)
	}
}

func Test_Validate_CloudConfiguration(t *testing.T) {
	grid := []struct {
		Description    string
		Input          kops.CloudConfiguration
		CloudProvider  kops.CloudProviderSpec
		ExpectedErrors []string
	}{
		{
			Description: "neither",
			Input:       kops.CloudConfiguration{},
		},
		{
			Description: "all false",
			Input: kops.CloudConfiguration{
				ManageStorageClasses: new(false),
			},
		},
		{
			Description: "all true",
			Input: kops.CloudConfiguration{
				ManageStorageClasses: new(true),
			},
		},
		{
			Description: "os false",
			Input:       kops.CloudConfiguration{},
			CloudProvider: kops.CloudProviderSpec{
				Openstack: &kops.OpenstackSpec{
					BlockStorage: &kops.OpenstackBlockStorageConfig{
						CreateStorageClass: new(false),
					},
				},
			},
		},
		{
			Description: "os true",
			Input:       kops.CloudConfiguration{},
			CloudProvider: kops.CloudProviderSpec{
				Openstack: &kops.OpenstackSpec{
					BlockStorage: &kops.OpenstackBlockStorageConfig{
						CreateStorageClass: new(true),
					},
				},
			},
		},
		{
			Description: "all false, os false",
			Input: kops.CloudConfiguration{
				ManageStorageClasses: new(false),
			},
			CloudProvider: kops.CloudProviderSpec{
				Openstack: &kops.OpenstackSpec{
					BlockStorage: &kops.OpenstackBlockStorageConfig{
						CreateStorageClass: new(false),
					},
				},
			},
		},
		{
			Description: "all false, os true",
			Input: kops.CloudConfiguration{
				ManageStorageClasses: new(false),
			},
			CloudProvider: kops.CloudProviderSpec{
				Openstack: &kops.OpenstackSpec{
					BlockStorage: &kops.OpenstackBlockStorageConfig{
						CreateStorageClass: new(true),
					},
				},
			},
			ExpectedErrors: []string{"Forbidden::cloudConfig.manageStorageClasses"},
		},
		{
			Description: "all true, os false",
			Input: kops.CloudConfiguration{
				ManageStorageClasses: new(true),
			},
			CloudProvider: kops.CloudProviderSpec{
				Openstack: &kops.OpenstackSpec{
					BlockStorage: &kops.OpenstackBlockStorageConfig{
						CreateStorageClass: new(false),
					},
				},
			},
			ExpectedErrors: []string{"Forbidden::cloudConfig.manageStorageClasses"},
		},
		{
			Description: "all true, os true",
			Input: kops.CloudConfiguration{
				ManageStorageClasses: new(true),
			},
			CloudProvider: kops.CloudProviderSpec{
				Openstack: &kops.OpenstackSpec{
					BlockStorage: &kops.OpenstackBlockStorageConfig{
						CreateStorageClass: new(true),
					},
				},
			},
		},
	}

	for _, g := range grid {
		fldPath := field.NewPath("cloudConfig")
		t.Run(g.Description, func(t *testing.T) {
			spec := &kops.ClusterSpec{
				CloudProvider: g.CloudProvider,
			}
			errs := validateCloudConfiguration(&g.Input, spec, fldPath)
			testErrors(t, g.Input, errs, g.ExpectedErrors)
		})
	}
}

func TestValidateSAExternalPermissions(t *testing.T) {
	grid := []struct {
		Description    string
		Input          []kops.ServiceAccountExternalPermission
		ExpectedErrors []string
	}{
		{
			Description: "Duplicate SA",
			Input: []kops.ServiceAccountExternalPermission{
				{
					Name:      "MySA",
					Namespace: "MyNS",
					AWS: &kops.AWSPermission{
						PolicyARNs: []string{"-"},
					},
				},
				{
					Name:      "MySA",
					Namespace: "MyNS",
					AWS: &kops.AWSPermission{
						PolicyARNs: []string{"-"},
					},
				},
			},
			ExpectedErrors: []string{"Duplicate value::iam.serviceAccountExternalPermissions[MyNS/MySA]"},
		},
		{
			Description: "Missing permissions",
			Input: []kops.ServiceAccountExternalPermission{
				{
					Name:      "MySA",
					Namespace: "MyNS",
				},
			},
			ExpectedErrors: []string{"Required value::iam.serviceAccountExternalPermissions[MyNS/MySA].aws"},
		},
		{
			Description: "Setting both arn and inline",
			Input: []kops.ServiceAccountExternalPermission{
				{
					Name:      "MySA",
					Namespace: "MyNS",
					AWS: &kops.AWSPermission{
						PolicyARNs:   []string{"-"},
						InlinePolicy: "-",
					},
				},
			},
			ExpectedErrors: []string{"Forbidden::iam.serviceAccountExternalPermissions[MyNS/MySA].aws"},
		},
		{
			Description: "Empty SA name",
			Input: []kops.ServiceAccountExternalPermission{
				{
					Namespace: "MyNS",
					AWS: &kops.AWSPermission{
						PolicyARNs:   []string{"-"},
						InlinePolicy: "-",
					},
				},
			},
			ExpectedErrors: []string{"Required value::iam.serviceAccountExternalPermissions[MyNS/].name"},
		},
		{
			Description: "Empty SA namespace",
			Input: []kops.ServiceAccountExternalPermission{
				{
					Name: "MySA",
					AWS: &kops.AWSPermission{
						PolicyARNs:   []string{"-"},
						InlinePolicy: "-",
					},
				},
			},
			ExpectedErrors: []string{"Required value::iam.serviceAccountExternalPermissions[/MySA].namespace"},
		},
	}

	for _, g := range grid {
		fldPath := field.NewPath("iam.serviceAccountExternalPermissions")
		t.Run(g.Description, func(t *testing.T) {
			errs := validateSAExternalPermissions(g.Input, fldPath)
			testErrors(t, g.Input, errs, g.ExpectedErrors)
		})
	}
}

func Test_Validate_Nvidia_Cluster(t *testing.T) {
	grid := []struct {
		Input          kops.ClusterSpec
		ExpectedErrors []string
	}{
		{
			Input: kops.ClusterSpec{
				Containerd: &kops.ContainerdConfig{
					NvidiaGPU: &kops.NvidiaGPUConfig{
						Enabled: new(true),
					},
				},
				CloudProvider: kops.CloudProviderSpec{
					AWS: &kops.AWSSpec{},
				},
			},
		},
		{
			Input: kops.ClusterSpec{
				Containerd: &kops.ContainerdConfig{
					NvidiaGPU: &kops.NvidiaGPUConfig{
						Enabled: new(true),
					},
				},
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
			},
			ExpectedErrors: []string{"Forbidden::containerd.nvidiaGPU"},
		},
		{
			Input: kops.ClusterSpec{
				Containerd: &kops.ContainerdConfig{
					NvidiaGPU: &kops.NvidiaGPUConfig{
						Enabled: new(true),
					},
				},
				CloudProvider: kops.CloudProviderSpec{
					GCE: &kops.GCESpec{},
				},
			},
			ExpectedErrors: []string{"Forbidden::containerd.nvidiaGPU"},
		},
	}
	for _, g := range grid {
		cluster := &kops.Cluster{}
		cluster.Spec = g.Input
		errs := validateNvidiaConfig(cluster, g.Input.Containerd.NvidiaGPU, field.NewPath("containerd", "nvidiaGPU"), true)
		testErrors(t, g.Input, errs, g.ExpectedErrors)
	}
}

func Test_Validate_Nvidia_Ig(t *testing.T) {
	grid := []struct {
		Input          kops.ClusterSpec
		ExpectedErrors []string
	}{
		{
			Input: kops.ClusterSpec{
				Containerd: &kops.ContainerdConfig{
					NvidiaGPU: &kops.NvidiaGPUConfig{
						Enabled: new(true),
					},
				},
				CloudProvider: kops.CloudProviderSpec{
					AWS: &kops.AWSSpec{},
				},
			},
		},
		{
			Input: kops.ClusterSpec{
				Containerd: &kops.ContainerdConfig{
					NvidiaGPU: &kops.NvidiaGPUConfig{
						Enabled: new(true),
					},
				},
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
			},
		},
		{
			Input: kops.ClusterSpec{
				Containerd: &kops.ContainerdConfig{
					NvidiaGPU: &kops.NvidiaGPUConfig{
						Enabled: new(true),
					},
				},
				CloudProvider: kops.CloudProviderSpec{
					GCE: &kops.GCESpec{},
				},
			},
			ExpectedErrors: []string{"Forbidden::containerd.nvidiaGPU"},
		},
	}
	for _, g := range grid {
		cluster := &kops.Cluster{}
		cluster.Spec = g.Input
		errs := validateNvidiaConfig(cluster, g.Input.Containerd.NvidiaGPU, field.NewPath("containerd", "nvidiaGPU"), false)
		testErrors(t, g.Input, errs, g.ExpectedErrors)
	}
}

func Test_Validate_GVisor(t *testing.T) {
	grid := []struct {
		name            string
		inClusterConfig bool
		enabled         *bool
		expectedErrors  []string
	}{
		{
			name:            "enabled in cluster config",
			inClusterConfig: true,
			enabled:         new(true),
			expectedErrors:  []string{"Forbidden::containerd.gvisor"},
		},
		{
			name:            "disabled in cluster config",
			inClusterConfig: true,
			enabled:         new(false),
			expectedErrors:  []string{"Forbidden::containerd.gvisor"},
		},
		{
			name:    "enabled in instance group config",
			enabled: new(true),
		},
	}
	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			containerd := &kops.ContainerdConfig{
				GVisor: &kops.GVisorConfig{
					Enabled: g.enabled,
				},
			}
			errs := validateContainerdConfig(&kops.Cluster{}, containerd, field.NewPath("containerd"), g.inClusterConfig)
			testErrors(t, g.name, errs, g.expectedErrors)
		})
	}
}

func Test_Validate_NriConfig(t *testing.T) {
	unsupportedContainerdVersion := "1.6.0"
	supportedContainerdVersion := "1.7.0"
	grid := []struct {
		Input          kops.ClusterSpec
		ExpectedErrors []string
	}{
		{
			Input: kops.ClusterSpec{
				Containerd: &kops.ContainerdConfig{
					NRI: &kops.NRIConfig{
						Enabled: new(true),
					},
					Version: &unsupportedContainerdVersion,
				},
			},
			ExpectedErrors: []string{"Forbidden::containerd.nri"},
		},
		{
			Input: kops.ClusterSpec{
				Containerd: &kops.ContainerdConfig{
					NRI:     &kops.NRIConfig{},
					Version: &unsupportedContainerdVersion,
				},
			},
			ExpectedErrors: []string{},
		},
		{
			Input: kops.ClusterSpec{
				Containerd: &kops.ContainerdConfig{
					NRI: &kops.NRIConfig{
						Enabled: nil,
					},
					Version: &unsupportedContainerdVersion,
				},
			},
			ExpectedErrors: []string{},
		},
		{
			Input: kops.ClusterSpec{
				Containerd: &kops.ContainerdConfig{
					NRI: &kops.NRIConfig{
						Enabled: new(false),
					},
					Version: &unsupportedContainerdVersion,
				},
			},
			ExpectedErrors: []string{},
		},
		{
			Input: kops.ClusterSpec{
				Containerd: &kops.ContainerdConfig{
					NRI: &kops.NRIConfig{
						Enabled: new(true),
					},
					Version: &supportedContainerdVersion,
				},
			},
			ExpectedErrors: []string{},
		},
	}
	for _, g := range grid {
		errs := validateNriConfig(g.Input.Containerd, field.NewPath("containerd", "nri"))
		testErrors(t, g.Input.Containerd, errs, g.ExpectedErrors)
	}
}

func newLinodeClusterForNetworkingValidation(networking kops.NetworkingSpec) *kops.Cluster {
	return &kops.Cluster{
		Spec: kops.ClusterSpec{
			CloudProvider: kops.CloudProviderSpec{
				Linode: &kops.LinodeSpec{},
			},
			Networking: networking,
		},
	}
}

func validLinodeNetworkingSpec() kops.NetworkingSpec {
	return kops.NetworkingSpec{
		NetworkCIDR:           "10.0.0.0/8",
		NonMasqueradeCIDR:     "100.64.0.0/10",
		PodCIDR:               "100.96.0.0/11",
		ServiceClusterIPRange: "100.64.0.0/13",
		Subnets: []kops.ClusterSubnetSpec{
			{
				Name:   "subnet-us-east",
				CIDR:   "10.11.0.0/16",
				Type:   kops.SubnetTypePublic,
				Region: "us-east",
			},
		},
	}
}

func TestValidateNetworkingLinode(t *testing.T) {
	tests := []struct {
		name     string
		network  kops.NetworkingSpec
		expected []*field.Error
	}{
		{
			name:    "accepts private network CIDR",
			network: validLinodeNetworkingSpec(),
		},
		{
			name: "rejects public network CIDR",
			network: func() kops.NetworkingSpec {
				n := validLinodeNetworkingSpec()
				n.NetworkCIDR = "8.8.8.0/24"
				n.Subnets[0].CIDR = "8.8.8.0/25"
				return n
			}(),
			expected: []*field.Error{
				{
					Type:   field.ErrorTypeInvalid,
					Field:  "networking.networkCIDR",
					Detail: "networkCIDR must be within a private IP range",
				},
			},
		},
		{
			name: "rejects networkID with networkCIDR",
			network: func() kops.NetworkingSpec {
				n := validLinodeNetworkingSpec()
				n.NetworkID = "123456"
				return n
			}(),
			expected: []*field.Error{
				{
					Type:   field.ErrorTypeForbidden,
					Field:  "networking.networkCIDR",
					Detail: "Linode (Akamai) doesn't support specifying both NetworkID and NetworkCIDR",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := newLinodeClusterForNetworkingValidation(tt.network)
			errList := validateNetworking(cluster, &cluster.Spec.Networking, field.NewPath("networking"), true, &cloudProviderConstraints{})
			testFieldErrors(t, errList, tt.expected)
		})
	}
}

func TestValidateAzureBlobAccountUniformity(t *testing.T) {
	tests := []struct {
		name     string
		spec     kops.ClusterSpec
		expected []*field.Error
	}{
		{
			name: "all matching azureblob URLs",
			spec: kops.ClusterSpec{
				ConfigStore: kops.ConfigStoreSpec{
					Base:     "azureblob://kopsstate/state/cluster.example.com",
					Keypairs: "azureblob://kopsstate/state/cluster.example.com/pki",
					Secrets:  "azureblob://kopsstate/state/cluster.example.com/secrets",
				},
				EtcdClusters: []kops.EtcdClusterSpec{{
					Backups: &kops.EtcdBackupSpec{
						BackupStore: "azureblob://kopsstate/state/cluster.example.com/backups/etcd/main",
					},
				}},
			},
		},
		{
			name: "non-azure cluster is unaffected",
			spec: kops.ClusterSpec{
				ConfigStore: kops.ConfigStoreSpec{
					Base:     "s3://my-bucket/cluster.example.com",
					Keypairs: "s3://my-bucket/cluster.example.com/pki",
				},
				EtcdClusters: []kops.EtcdClusterSpec{{
					Backups: &kops.EtcdBackupSpec{
						BackupStore: "s3://my-bucket/cluster.example.com/backups/etcd/main",
					},
				}},
			},
		},
		{
			name: "keypairs uses different storage account",
			spec: kops.ClusterSpec{
				ConfigStore: kops.ConfigStoreSpec{
					Base:     "azureblob://kopsstate/state/cluster.example.com",
					Keypairs: "azureblob://otheracct/state/cluster.example.com/pki",
				},
			},
			expected: []*field.Error{
				{
					Type:  field.ErrorTypeInvalid,
					Field: "spec.configStore.keypairs",
				},
			},
		},
		{
			name: "secrets uses different storage account",
			spec: kops.ClusterSpec{
				ConfigStore: kops.ConfigStoreSpec{
					Base:    "azureblob://kopsstate/state/cluster.example.com",
					Secrets: "azureblob://otheracct/state/cluster.example.com/secrets",
				},
			},
			expected: []*field.Error{
				{
					Type:  field.ErrorTypeInvalid,
					Field: "spec.configStore.secrets",
				},
			},
		},
		{
			name: "etcd backupStore uses different storage account",
			spec: kops.ClusterSpec{
				ConfigStore: kops.ConfigStoreSpec{
					Base: "azureblob://kopsstate/state/cluster.example.com",
				},
				EtcdClusters: []kops.EtcdClusterSpec{{
					Backups: &kops.EtcdBackupSpec{
						BackupStore: "azureblob://otheracct/backups/etcd/main",
					},
				}},
			},
			expected: []*field.Error{
				{
					Type:  field.ErrorTypeInvalid,
					Field: "spec.etcdClusters[0].backups.backupStore",
				},
			},
		},
		{
			name: "azureblob backupStore with non-azure configStore.base is rejected",
			spec: kops.ClusterSpec{
				ConfigStore: kops.ConfigStoreSpec{
					Base: "s3://my-bucket/cluster.example.com",
				},
				EtcdClusters: []kops.EtcdClusterSpec{{
					Backups: &kops.EtcdBackupSpec{
						BackupStore: "azureblob://kopsstate/backups/etcd/main",
					},
				}},
			},
			expected: []*field.Error{
				{
					Type:  field.ErrorTypeInvalid,
					Field: "spec.etcdClusters[0].backups.backupStore",
				},
			},
		},
		{
			name: "malformed azureblob configStore.base is rejected",
			spec: kops.ClusterSpec{
				ConfigStore: kops.ConfigStoreSpec{
					Base: "azureblob://kopsstate",
				},
			},
			expected: []*field.Error{
				{
					Type:  field.ErrorTypeInvalid,
					Field: "spec.configStore.base",
				},
			},
		},
		{
			name: "malformed azureblob keypairs is rejected",
			spec: kops.ClusterSpec{
				ConfigStore: kops.ConfigStoreSpec{
					Base:     "azureblob://kopsstate/state/cluster.example.com",
					Keypairs: "azureblob://kopsstate",
				},
			},
			expected: []*field.Error{
				{
					Type:  field.ErrorTypeInvalid,
					Field: "spec.configStore.keypairs",
				},
			},
		},
		{
			name: "non-azure backup store with azure config base is allowed",
			spec: kops.ClusterSpec{
				ConfigStore: kops.ConfigStoreSpec{
					Base: "azureblob://kopsstate/state/cluster.example.com",
				},
				EtcdClusters: []kops.EtcdClusterSpec{{
					Backups: &kops.EtcdBackupSpec{
						BackupStore: "memfs://tests/cluster.example.com/backups/etcd/main",
					},
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errList := validateAzureBlobAccountUniformity(&tt.spec, field.NewPath("spec"))
			testFieldErrors(t, errList, tt.expected)
		})
	}
}

func TestValidateFileRepository(t *testing.T) {
	grid := []struct {
		Input          string
		ExpectedErrors []string
	}{
		{
			Input: "https://example.com/files",
		},
		{
			Input: "http://example.com/files",
		},
		{
			Input:          "s3://example-k8s-assets/kops",
			ExpectedErrors: []string{"Invalid value::spec.assets.fileRepository"},
		},
		{
			Input:          "gs://example-k8s-assets/kops",
			ExpectedErrors: []string{"Invalid value::spec.assets.fileRepository"},
		},
		{
			Input:          "example.com/files",
			ExpectedErrors: []string{"Invalid value::spec.assets.fileRepository"},
		},
		{
			Input:          "",
			ExpectedErrors: []string{"Invalid value::spec.assets.fileRepository"},
		},
		{
			Input:          "https://",
			ExpectedErrors: []string{"Invalid value::spec.assets.fileRepository"},
		},
	}
	for _, g := range grid {
		errs := validateFileRepository(g.Input, field.NewPath("spec", "assets", "fileRepository"))
		testErrors(t, g.Input, errs, g.ExpectedErrors)
	}
}
