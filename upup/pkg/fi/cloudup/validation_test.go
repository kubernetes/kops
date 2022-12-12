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

package cloudup

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"k8s.io/klog/v2"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/upup/pkg/fi"
)

const testAWSRegion = "us-test-1"

func buildDefaultCluster(t *testing.T, ctx context.Context) *api.Cluster {
	cloud, c := buildMinimalCluster()

	err := PerformAssignments(c, cloud)
	if err != nil {
		t.Fatalf("error from PerformAssignments: %v", err)
	}

	fullSpec, err := mockedPopulateClusterSpec(ctx, c, cloud)
	if err != nil {
		t.Fatalf("error from PopulateClusterSpec: %v", err)
	}

	//// TODO: We should actually just specify the minimums here, and run in though the default logic
	//c.Cluster = &api.Cluster{}
	//c.Cluster.Name = "testcluster.mydomain.com"

	//c.InstanceGroups = append(c.InstanceGroups, buildNodeInstanceGroup("us-test-1a"))
	//c.InstanceGroups = append(c.InstanceGroups, buildMasterInstanceGroup("us-test-1a"))
	//c.SSHPublicKey = path.Join(os.Getenv("HOME"), ".ssh", "id_rsa.pub")
	//
	//c.Cluster.Spec.Kubelet = &api.KubeletConfig{}
	//c.Cluster.Spec.KubeControllerManager = &api.KubeControllerManagerConfig{}
	//c.Cluster.Spec.KubeDNS = &api.KubeDNSConfig{}
	//c.Cluster.Spec.KubeAPIServer = &api.KubeAPIServerConfig{}
	//c.Cluster.Spec.KubeProxy = &api.KubeProxyConfig{}
	//c.Cluster.Spec.Docker = &api.DockerConfig{}
	//
	//c.Cluster.Spec.NetworkCIDR = "172.20.0.0/16"
	//
	//c.Cluster.Spec.NonMasqueradeCIDR = "100.64.0.0/10"
	//c.Cluster.Spec.Kubelet.NonMasqueradeCIDR = c.Cluster.Spec.NonMasqueradeCIDR
	//
	//c.Cluster.Spec.ServiceClusterIPRange = "100.64.1.0/24"
	//c.Cluster.Spec.KubeAPIServer.ServiceClusterIPRange = c.Cluster.Spec.ServiceClusterIPRange
	//
	//c.Cluster.Spec.KubeDNS.ServerIP = "100.64.1.10"
	//c.Cluster.Spec.Kubelet.ClusterDNS = c.Cluster.Spec.KubeDNS.ServerIP
	//
	//c.Cluster.Spec.CloudProvider = "aws"
	//c.Cluster.Spec.Kubelet.CloudProvider = c.Cluster.Spec.CloudProvider
	//c.Cluster.Spec.KubeAPIServer.CloudProvider = c.Cluster.Spec.CloudProvider
	//c.Cluster.Spec.KubeControllerManager.CloudProvider = c.Cluster.Spec.CloudProvider
	//
	//c.Target = "dryrun"
	//
	//basePath := vfs.NewMemFSPath(memfs, "test-statestore")
	//c.ClusterRegistry = api.NewClusterRegistry(basePath)

	return fullSpec
}

func TestValidateFull_Default_Validates(t *testing.T) {
	ctx := context.TODO()

	c := buildDefaultCluster(t, ctx)
	if errs := validation.ValidateCluster(ctx, c, false); len(errs) != 0 {
		klog.Infof("Cluster: %v", c)
		t.Fatalf("Validate gave unexpected error (strict=false): %v", errs.ToAggregate())
	}
	if errs := validation.ValidateCluster(ctx, c, true); len(errs) != 0 {
		t.Fatalf("Validate gave unexpected error (strict=true): %v", errs.ToAggregate())
	}
}

func TestValidateFull_ClusterName_InvalidDNS_NoDot(t *testing.T) {
	ctx := context.TODO()
	c := buildDefaultCluster(t, ctx)
	c.ObjectMeta.Name = "test"
	expectErrorFromValidate(t, ctx, c, "DNS name")
}

func TestValidateFull_ClusterName_InvalidDNS_Invalid(t *testing.T) {
	ctx := context.TODO()
	c := buildDefaultCluster(t, ctx)
	c.ObjectMeta.Name = "test.-"
	expectErrorFromValidate(t, ctx, c, "DNS name")
}

func TestValidateFull_ClusterName_Required(t *testing.T) {
	ctx := context.TODO()
	c := buildDefaultCluster(t, ctx)
	c.ObjectMeta.Name = ""
	expectErrorFromValidate(t, ctx, c, "Name")
}

func TestValidateFull_UpdatePolicy_Valid(t *testing.T) {
	for _, test := range []struct {
		label  string
		policy *string
	}{
		{
			label: "missing",
		},
		{
			label:  "automatic",
			policy: fi.PtrTo(api.UpdatePolicyAutomatic),
		},
		{
			label:  "external",
			policy: fi.PtrTo(api.UpdatePolicyExternal),
		},
	} {
		t.Run(test.label, func(t *testing.T) {
			ctx := context.TODO()
			c := buildDefaultCluster(t, ctx)
			expectNoErrorFromValidate(t, ctx, c)
		})
	}
}

func TestValidateFull_UpdatePolicy_Invalid(t *testing.T) {
	for _, test := range []struct {
		label  string
		policy string
	}{
		{
			label:  "empty",
			policy: "",
		},
		{
			label:  "populated",
			policy: "not-a-real-value",
		},
	} {
		t.Run(test.label, func(t *testing.T) {
			ctx := context.TODO()
			c := buildDefaultCluster(t, ctx)

			c.Spec.UpdatePolicy = &test.policy
			expectErrorFromValidate(t, ctx, c, "spec.updatePolicy")
		})
	}
}

func Test_Validate_Kubenet_With_14(t *testing.T) {
	ctx := context.TODO()

	c := buildDefaultCluster(t, ctx)
	c.Spec.KubernetesVersion = "1.4.1"
	c.Spec.Networking.Kubenet = &api.KubenetNetworkingSpec{}

	expectNoErrorFromValidate(t, ctx, c)
}

func TestValidate_ClusterName_Import(t *testing.T) {
	ctx := context.TODO()

	c := buildDefaultCluster(t, ctx)

	// When we import a cluster, it likely won't have a valid name until we convert it
	c.ObjectMeta.Annotations = make(map[string]string)
	c.ObjectMeta.Annotations[api.AnnotationNameManagement] = api.AnnotationValueManagementImported
	c.ObjectMeta.Name = "kubernetes"

	expectNoErrorFromValidate(t, ctx, c)
}

func TestValidate_ContainerRegistry_and_ContainerProxy_exclusivity(t *testing.T) {
	ctx := context.TODO()

	c := buildDefaultCluster(t, ctx)

	assets := new(api.Assets)
	c.Spec.Assets = assets

	expectNoErrorFromValidate(t, ctx, c)

	registry := "https://registry.example.com/"
	c.Spec.Assets.ContainerRegistry = &registry
	expectNoErrorFromValidate(t, ctx, c)

	proxy := "https://proxy.example.com/"
	c.Spec.Assets.ContainerProxy = &proxy
	expectErrorFromValidate(t, ctx, c, "containerProxy cannot be used in conjunction with containerRegistry")

	c.Spec.Assets.ContainerRegistry = nil
	expectNoErrorFromValidate(t, ctx, c)
}

func expectErrorFromValidate(t *testing.T, ctx context.Context, c *api.Cluster, message string) {
	errs := validation.ValidateCluster(ctx, c, false)
	if len(errs) == 0 {
		t.Fatalf("Expected error from Validate")
	}
	actualMessage := fmt.Sprintf("%v", errs.ToAggregate())
	if !strings.Contains(actualMessage, message) {
		t.Fatalf("Expected error %q, got %q", message, actualMessage)
	}
}

func expectNoErrorFromValidate(t *testing.T, ctx context.Context, c *api.Cluster) {
	errs := validation.ValidateCluster(ctx, c, false)
	if len(errs) != 0 {
		t.Fatalf("Unexpected error from Validate: %v", errs.ToAggregate())
	}
}
