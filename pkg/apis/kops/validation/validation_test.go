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
	"testing"

	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
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
		errs := validateCIDR(g.Input, field.NewPath("CIDR"))

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
				{Name: "a", ProviderID: "a", Type: kops.SubnetTypePublic},
				{Name: "b", ProviderID: "b", Type: kops.SubnetTypePublic},
			},
		},
		{
			Input: []kops.ClusterSubnetSpec{
				{Name: "a", ProviderID: "a", Type: kops.SubnetTypePublic},
				{Name: "b", ProviderID: "", Type: kops.SubnetTypePublic},
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
		cluster := &kops.ClusterSpec{
			CloudProvider: kops.CloudProviderSpec{
				AWS: &kops.AWSSpec{},
			},
			Subnets: g.Input,
		}
		errs := validateSubnets(cluster, field.NewPath("subnets"))

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
				AuthorizationMode: fi.String("RBAC"),
			},
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					Authorization: &kops.AuthorizationSpec{
						RBAC: &kops.RBACAuthorizationSpec{},
					},
					KubernetesVersion: "1.19.0",
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
				AuthorizationMode: fi.String("RBAC,Node"),
			},
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					Authorization: &kops.AuthorizationSpec{
						RBAC: &kops.RBACAuthorizationSpec{},
					},
					KubernetesVersion: "1.19.0",
					CloudProvider: kops.CloudProviderSpec{
						AWS: &kops.AWSSpec{},
					},
				},
			},
		},
		{
			Input: kops.KubeAPIServerConfig{
				AuthorizationMode: fi.String("RBAC,Node,Bogus"),
			},
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					Authorization: &kops.AuthorizationSpec{
						RBAC: &kops.RBACAuthorizationSpec{},
					},
					KubernetesVersion: "1.19.0",
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
	}
	for _, g := range grid {
		if g.Cluster == nil {
			g.Cluster = &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "1.20.0",
				},
			}
		}
		errs := validateKubeAPIServer(&g.Input, g.Cluster, field.NewPath("KubeAPIServer"))

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

func Test_Validate_DockerConfig_Storage(t *testing.T) {
	for _, name := range []string{"aufs", "zfs", "overlay"} {
		config := &kops.DockerConfig{Storage: &name}
		errs := validateDockerConfig(config, field.NewPath("docker"))
		if len(errs) != 0 {
			t.Fatalf("Unexpected errors validating DockerConfig %q", errs)
		}
	}

	for _, name := range []string{"overlayfs", "", "au"} {
		config := &kops.DockerConfig{Storage: &name}
		errs := validateDockerConfig(config, field.NewPath("docker"))
		if len(errs) != 1 {
			t.Fatalf("Expected errors validating DockerConfig %+v", config)
		}
		if errs[0].Field != "docker.storage" || errs[0].Type != field.ErrorTypeNotSupported {
			t.Fatalf("Not the expected error validating DockerConfig %q", errs)
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
		networking := &kops.NetworkingSpec{}
		networking.Flannel = &g.Input

		cluster := &kops.Cluster{}
		cluster.Spec.Networking = networking

		errs := validateNetworking(cluster, networking, field.NewPath("networking"))
		testErrors(t, g.Input, errs, g.ExpectedErrors)
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
				"master": `[ { "Action": [ "s3:GetObject" ], "Resource": [ "*" ], "Effect": "Allow" } ]`,
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
				"master": `badjson`,
			},
			ExpectedErrors: []string{"Invalid value::spec.additionalPolicies[master]"},
		},
		{
			Input: map[string]string{
				"master": `[ { "Action": [ "s3:GetObject" ], "Resource": [ "*" ] } ]`,
			},
			ExpectedErrors: []string{"Required value::spec.additionalPolicies[master][0].Effect"},
		},
		{
			Input: map[string]string{
				"master": `[ { "Action": [ "s3:GetObject" ], "Resource": [ "*" ], "Effect": "allow" } ]`,
			},
			ExpectedErrors: []string{"Unsupported value::spec.additionalPolicies[master][0].Effect"},
		},
	}
	for _, g := range grid {
		clusterSpec := &kops.ClusterSpec{
			KubernetesVersion:  "1.17.0",
			AdditionalPolicies: &g.Input,
			Subnets: []kops.ClusterSubnetSpec{
				{Name: "subnet1", Type: kops.SubnetTypePublic},
			},
			EtcdClusters: []kops.EtcdClusterSpec{
				{
					Name: "main",
					Members: []kops.EtcdMemberSpec{
						{
							Name:          "us-test-1a",
							InstanceGroup: fi.String("master-us-test-1a"),
						},
					},
				},
			},
		}
		errs := validateClusterSpec(clusterSpec, &kops.Cluster{Spec: *clusterSpec}, field.NewPath("spec"))
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
				Calico: &kops.CalicoNetworkingSpec{
					AWSSrcDstCheck: "off",
				},
			},
			ExpectedErrors: []string{"Unsupported value::calico.awsSrcDstCheck"},
		},
		{
			Description: "AWS source/destination checks enabled",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					AWSSrcDstCheck: "Enable",
				},
			},
		},
		{
			Description: "AWS source/destination checks disabled",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					AWSSrcDstCheck: "Disable",
				},
			},
		},
		{
			Description: "AWS source/destination checks left as is",
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					AWSSrcDstCheck: "DoNothing",
				},
			},
		},
		{
			Description: "encapsulation none with IPv4",
			Input: caliInput{
				Cluster: &kops.ClusterSpec{
					NonMasqueradeCIDR: "100.64.0.0/10",
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
					NonMasqueradeCIDR: "::/0",
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
					NonMasqueradeCIDR: "::/0",
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
					NonMasqueradeCIDR: "100.64.0.0/10",
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
					NonMasqueradeCIDR: "100.64.0.0/10",
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
					NonMasqueradeCIDR: "100.64.0.0/10",
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
					NonMasqueradeCIDR: "100.64.0.0/10",
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
					NonMasqueradeCIDR: "100.64.0.0/10",
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
					NonMasqueradeCIDR: "100.64.0.0/10",
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
					NonMasqueradeCIDR: "100.64.0.0/10",
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
					NonMasqueradeCIDR: "::/0",
				},
				Calico: &kops.CalicoNetworkingSpec{
					EncapsulationMode: "none",
					IPIPMode:          "Never",
					VXLANMode:         "Never",
				},
			},
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
				Masquerade: fi.Bool(false),
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
				Masquerade: fi.Bool(true),
				IPAM:       "eni",
			},
			Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					AWS: &kops.AWSSpec{},
				},
			},
			ExpectedErrors: []string{"Forbidden::cilium.masquerade"},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				EnableL7Proxy:        fi.Bool(true),
				InstallIptablesRules: fi.Bool(false),
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
				KubernetesVersion: "1.18.0",
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
					Enabled: fi.Bool(true),
				},
			},
			ExpectedErrors: []string{"Forbidden::cilium.hubble.enabled"},
		},
		{
			Cilium: kops.CiliumNetworkingSpec{
				Version: "v1.8.0",
				Hubble: &kops.HubbleSpec{
					Enabled: fi.Bool(true),
				},
			},
			Spec: kops.ClusterSpec{
				CertManager: &kops.CertManagerConfig{
					Enabled: fi.Bool(true),
				},
			},
		},
	}
	for _, g := range grid {
		g.Spec.Networking = &kops.NetworkingSpec{
			Cilium: &g.Cilium,
		}
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
						Enabled: fi.Bool(true),
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
						Enabled: fi.Bool(true),
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
						Enabled: fi.Bool(true),
					},
				},
				Networking: &kops.NetworkingSpec{
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
						Enabled: fi.Bool(true),
						LocalIP: "169.254.20.10",
					},
				},
				Networking: &kops.NetworkingSpec{
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
		ExpectedErrors []string
	}{
		{
			Description: "neither",
			Input:       kops.CloudConfiguration{},
		},
		{
			Description: "all false",
			Input: kops.CloudConfiguration{
				ManageStorageClasses: fi.Bool(false),
			},
		},
		{
			Description: "all true",
			Input: kops.CloudConfiguration{
				ManageStorageClasses: fi.Bool(true),
			},
		},
		{
			Description: "os false",
			Input: kops.CloudConfiguration{
				Openstack: &kops.OpenstackConfiguration{
					BlockStorage: &kops.OpenstackBlockStorageConfig{
						CreateStorageClass: fi.Bool(false),
					},
				},
			},
		},
		{
			Description: "os true",
			Input: kops.CloudConfiguration{
				Openstack: &kops.OpenstackConfiguration{
					BlockStorage: &kops.OpenstackBlockStorageConfig{
						CreateStorageClass: fi.Bool(true),
					},
				},
			},
		},
		{
			Description: "all false, os false",
			Input: kops.CloudConfiguration{
				ManageStorageClasses: fi.Bool(false),
				Openstack: &kops.OpenstackConfiguration{
					BlockStorage: &kops.OpenstackBlockStorageConfig{
						CreateStorageClass: fi.Bool(false),
					},
				},
			},
		},
		{
			Description: "all false, os true",
			Input: kops.CloudConfiguration{
				ManageStorageClasses: fi.Bool(false),
				Openstack: &kops.OpenstackConfiguration{
					BlockStorage: &kops.OpenstackBlockStorageConfig{
						CreateStorageClass: fi.Bool(true),
					},
				},
			},
			ExpectedErrors: []string{"Forbidden::cloudConfig.manageStorageClasses"},
		},
		{
			Description: "all true, os false",
			Input: kops.CloudConfiguration{
				ManageStorageClasses: fi.Bool(true),
				Openstack: &kops.OpenstackConfiguration{
					BlockStorage: &kops.OpenstackBlockStorageConfig{
						CreateStorageClass: fi.Bool(false),
					},
				},
			},
			ExpectedErrors: []string{"Forbidden::cloudConfig.manageStorageClasses"},
		},
		{
			Description: "all true, os true",
			Input: kops.CloudConfiguration{
				ManageStorageClasses: fi.Bool(true),
				Openstack: &kops.OpenstackConfiguration{
					BlockStorage: &kops.OpenstackBlockStorageConfig{
						CreateStorageClass: fi.Bool(true),
					},
				},
			},
		},
	}

	for _, g := range grid {
		fldPath := field.NewPath("cloudConfig")
		t.Run(g.Description, func(t *testing.T) {
			errs := validateCloudConfiguration(&g.Input, fldPath)
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

func Test_Validate_Nvdia(t *testing.T) {
	grid := []struct {
		Input          kops.ClusterSpec
		ExpectedErrors []string
	}{
		{
			Input: kops.ClusterSpec{
				Containerd: &kops.ContainerdConfig{
					NvidiaGPU: &kops.NvidiaGPUConfig{
						Enabled: fi.Bool(true),
					},
				},
				CloudProvider: kops.CloudProviderSpec{
					AWS: &kops.AWSSpec{},
				},
				ContainerRuntime: "containerd",
			},
		},
		{
			Input: kops.ClusterSpec{
				Containerd: &kops.ContainerdConfig{
					NvidiaGPU: &kops.NvidiaGPUConfig{
						Enabled: fi.Bool(true),
					},
				},
				CloudProvider: kops.CloudProviderSpec{
					GCE: &kops.GCESpec{},
				},
				ContainerRuntime: "containerd",
			},
			ExpectedErrors: []string{"Forbidden::containerd.nvidiaGPU"},
		},
		{
			Input: kops.ClusterSpec{
				Containerd: &kops.ContainerdConfig{
					NvidiaGPU: &kops.NvidiaGPUConfig{
						Enabled: fi.Bool(true),
					},
				},
				CloudProvider: kops.CloudProviderSpec{
					AWS: &kops.AWSSpec{},
				},
				ContainerRuntime: "docker",
			},
			ExpectedErrors: []string{"Forbidden::containerd.nvidiaGPU"},
		},
	}
	for _, g := range grid {
		errs := validateNvidiaConfig(&g.Input, g.Input.Containerd.NvidiaGPU, field.NewPath("containerd", "nvidiaGPU"))
		testErrors(t, g.Input, errs, g.ExpectedErrors)
	}
}
