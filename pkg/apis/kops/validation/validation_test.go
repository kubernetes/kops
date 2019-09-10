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

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/pkg/apis/kops"
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
				t.Errorf("expected error %v from %v, was not found in %q", expected, context, errStrings.List())
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
				{Name: "a"},
			},
		},
		{
			Input: []kops.ClusterSubnetSpec{
				{Name: ""},
			},
			ExpectedErrors: []string{"Required value::Subnets[0].Name"},
		},
		{
			Input: []kops.ClusterSubnetSpec{
				{Name: "a"},
				{Name: "a"},
			},
			ExpectedErrors: []string{"Invalid value::Subnets"},
		},
		{
			Input: []kops.ClusterSubnetSpec{
				{Name: "a", ProviderID: "a"},
				{Name: "b", ProviderID: "b"},
			},
		},
		{
			Input: []kops.ClusterSubnetSpec{
				{Name: "a", ProviderID: "a"},
				{Name: "b", ProviderID: ""},
			},
			ExpectedErrors: []string{"Invalid value::Subnets"},
		},
	}
	for _, g := range grid {
		errs := validateSubnets(g.Input, field.NewPath("Subnets"))

		testErrors(t, g.Input, errs, g.ExpectedErrors)
	}
}

func TestValidateKubeAPIServer(t *testing.T) {
	str := "foobar"
	authzMode := "RBAC,Webhook"

	grid := []struct {
		Input          kops.KubeAPIServerConfig
		ExpectedErrors []string
		ExpectedDetail string
	}{
		{
			Input: kops.KubeAPIServerConfig{
				ProxyClientCertFile: &str,
			},
			ExpectedErrors: []string{
				"Invalid value::KubeAPIServer",
			},
			ExpectedDetail: "ProxyClientCertFile and ProxyClientKeyFile must both be specified (or not all)",
		},
		{
			Input: kops.KubeAPIServerConfig{
				ProxyClientKeyFile: &str,
			},
			ExpectedErrors: []string{
				"Invalid value::KubeAPIServer",
			},
			ExpectedDetail: "ProxyClientCertFile and ProxyClientKeyFile must both be specified (or not all)",
		},
		{
			Input: kops.KubeAPIServerConfig{
				ServiceNodePortRange: str,
			},
			ExpectedErrors: []string{
				"Invalid value::KubeAPIServer",
			},
		},
		{
			Input: kops.KubeAPIServerConfig{
				AuthorizationMode: &authzMode,
			},
			ExpectedErrors: []string{
				"Invalid value::KubeAPIServer",
			},
			ExpectedDetail: "Authorization mode Webhook requires AuthorizationWebhookConfigFile to be specified",
		},
	}
	for _, g := range grid {
		errs := validateKubeAPIServer(&g.Input, field.NewPath("KubeAPIServer"))

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
		errs := ValidateDockerConfig(config, field.NewPath("docker"))
		if len(errs) != 0 {
			t.Fatalf("Unexpected errors validating DockerConfig %q", errs)
		}
	}

	for _, name := range []string{"overlayfs", "", "au"} {
		config := &kops.DockerConfig{Storage: &name}
		errs := ValidateDockerConfig(config, field.NewPath("docker"))
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
			ExpectedErrors: []string{"Required value::Networking.Flannel.Backend"},
		},
		{
			Input: kops.FlannelNetworkingSpec{
				Backend: "nope",
			},
			ExpectedErrors: []string{"Unsupported value::Networking.Flannel.Backend"},
		},
	}
	for _, g := range grid {
		networking := &kops.NetworkingSpec{}
		networking.Flannel = &g.Input

		cluster := &kops.Cluster{}
		cluster.Spec.Networking = networking

		errs := validateNetworking(&cluster.Spec, networking, field.NewPath("Networking"))
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
			ExpectedErrors: []string{"Invalid value::spec.additionalPolicies"},
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
			ExpectedErrors: []string{"Invalid value::spec.additionalPolicies[master][0].Effect"},
		},
	}
	for _, g := range grid {
		clusterSpec := &kops.ClusterSpec{
			AdditionalPolicies: &g.Input,
			Subnets: []kops.ClusterSubnetSpec{
				{Name: "subnet1"},
			},
		}
		errs := validateClusterSpec(clusterSpec, field.NewPath("spec"))
		testErrors(t, g.Input, errs, g.ExpectedErrors)
	}
}

type caliInput struct {
	Calico *kops.CalicoNetworkingSpec
	Etcd   *kops.EtcdClusterSpec
}

func Test_Validate_Calico(t *testing.T) {
	grid := []struct {
		Input          caliInput
		ExpectedErrors []string
	}{
		{
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{},
				Etcd:   &kops.EtcdClusterSpec{},
			},
		},
		{
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					TyphaReplicas: 3,
				},
				Etcd: &kops.EtcdClusterSpec{},
			},
		},
		{
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					TyphaReplicas: -1,
				},
				Etcd: &kops.EtcdClusterSpec{},
			},
			ExpectedErrors: []string{"Invalid value::Calico.TyphaReplicas"},
		},
		{
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					MajorVersion: "v3",
				},
				Etcd: &kops.EtcdClusterSpec{
					Version: "3.2.18",
				},
			},
		},
		{
			Input: caliInput{
				Calico: &kops.CalicoNetworkingSpec{
					MajorVersion: "v3",
				},
				Etcd: &kops.EtcdClusterSpec{
					Version: "2.2.18",
				},
			},
			ExpectedErrors: []string{"Invalid value::Calico.MajorVersion"},
		},
	}
	for _, g := range grid {
		errs := validateNetworkingCalico(g.Input.Calico, g.Input.Etcd, field.NewPath("Calico"))
		testErrors(t, g.Input, errs, g.ExpectedErrors)
	}
}
