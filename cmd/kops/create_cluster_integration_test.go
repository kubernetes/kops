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

package main

import (
	"bytes"
	"context"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/kops/cloudmock/aws/mockec2"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/pkg/testutils/golden"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

var MagicTimestamp = metav1.Time{Time: time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)}

// TestCreateClusterMinimal runs kops create cluster minimal.example.com --zones us-test-1a
func TestCreateClusterMinimal(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/minimal-1.29", "v1alpha2")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/minimal-1.30", "v1alpha2")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/minimal-1.31", "v1alpha2")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/minimal-1.32", "v1alpha2")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/minimal-arm64", "v1alpha2")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/minimal-irsa", "v1alpha2")
}

// TestCreateClusterHetzner runs kops create cluster minimal.k8s.local --zones fsn1
func TestCreateClusterHetzner(t *testing.T) {
	t.Setenv("HCLOUD_TOKEN", "REDACTED")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ha_hetzner", "v1alpha2")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/minimal_hetzner", "v1alpha2")
}

func TestCreateClusterOpenStack(t *testing.T) {
	t.Setenv("OS_REGION_NAME", "us-test1")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ha_openstack", "v1alpha2")
}

func TestCreateClusterOpenStackOctavia(t *testing.T) {
	t.Setenv("OS_REGION_NAME", "us-test1")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ha_openstack_octavia", "v1alpha2")
}

func TestCreateClusterOpenStackNoDNS(t *testing.T) {
	t.Setenv("OS_REGION_NAME", "us-test1")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ha_openstack_nodns", "v1alpha2")
}

// TestCreateClusterCilium runs kops with the cilium networking flags
func TestCreateClusterCilium(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/cilium-eni", "v1alpha2")
}

// TestCreateClusterOverride tests the override flag
func TestCreateClusterOverride(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/overrides", "v1alpha2")
}

// TestCreateClusterKubernetesFeatureGates tests the override flag
func TestCreateClusterKubernetesFeatureGates(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/minimal_feature-gates", "v1alpha2")
}

// TestCreateClusterComplex runs kops create cluster, with a grab-bag of edge cases
func TestCreateClusterComplex(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/complex", "v1alpha2")
}

// TestCreateClusterComplexPrivate runs kops create cluster, with a grab-bag of edge cases
func TestCreateClusterComplexPrivate(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/complex-private", "v1alpha2")
}

// TestCreateClusterHA runs kops create cluster ha.example.com --zones us-test-1a,us-test-1b,us-test-1c --master-zones us-test-1a,us-test-1b,us-test-1c
func TestCreateClusterHA(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ha", "v1alpha2")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ha_encrypt", "v1alpha2")
}

// TestCreateClusterMinimalGCE runs kops create cluster minimal.example.com --cloud gce --zones us-test1-a
func TestCreateClusterMinimalGCE(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/minimal-gce", "v1alpha2")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/minimal-gce-dns-none", "v1alpha2")
}

// TestCreateClusterHAGCE runs kops create cluster ha-gce.example.com --cloud gce --zones us-test1-a,us-test1-b,us-test1-c --master-zones us-test1-a,us-test1-b,us-test1-c
func TestCreateClusterHAGCE(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ha_gce", "v1alpha2")
}

// TestCreateClusterGCE runs kops create cluster gce.example.com --cloud gce --zones us-test1-a --gce-service-account=test-account@testproject.iam.gserviceaccounts.com
func TestCreateClusterGCE(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/gce_byo_sa", "v1alpha2")
}

// TestCreateClusterHASharedZone tests kops create cluster when the master count is bigger than the number of zones
func TestCreateClusterHASharedZone(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ha_shared_zone", "v1alpha2")
}

// TestCreateClusterHASharedZones tests kops create cluster when the master count is bigger than the number of zones
func TestCreateClusterHASharedZones(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ha_shared_zones", "v1alpha2")
}

// TestCreateClusterPrivate runs kops create cluster private.example.com --zones us-test-1a --master-zones us-test-1a
func TestCreateClusterPrivate(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/private", "v1alpha2")
}

// TestCreateClusterPrivateGCE runs kops create cluster private.example.com --cloud gce --zones us-test1-a --master-zones us-test-1a --topology private --bastion
func TestCreateClusterPrivateGCE(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/private_gce", "v1alpha2")
}

// TestCreateClusterWithNGWSpecified runs kops create cluster private.example.com --zones us-test-1a --master-zones us-test-1a
func TestCreateClusterWithNGWSpecified(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ngwspecified", "v1alpha2")
}

// TestCreateClusterWithINGWSpecified runs kops create cluster private.example.com --zones us-test-1a --master-zones us-test-1a
func TestCreateClusterWithINGWSpecified(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ingwspecified", "v1alpha2")
}

// TestCreateClusterSharedVPC runs kops create cluster vpc.example.com --zones us-test-1a --master-zones us-test-1a --vpc vpc-12345678
func TestCreateClusterSharedVPC(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/shared_vpc", "v1alpha2")
}

// TestCreateClusterSharedSubnets runs kops create cluster subnet.example.com --zones us-test-1a --master-zones us-test-1a --vpc vpc-12345678 --subnets subnet-1
func TestCreateClusterSharedSubnets(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/shared_subnets", "v1alpha2")
}

// TestCreateClusterSharedSubnetsVpcLookup runs kops create cluster subnet.example.com --zones us-test-1a --master-zones us-test-1a --vpc --subnets subnet-1
func TestCreateClusterSharedSubnetsVpcLookup(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/shared_subnets_vpc_lookup", "v1alpha2")
}

// TestCreateClusterPrivateSharedSubnets runs kops create cluster private-subnet.example.com --zones us-test-1a --master-zones us-test-1a --vpc vpc-12345678 --subnets subnet-1 --utility-subnets subnet-2
func TestCreateClusterPrivateSharedSubnets(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/private_shared_subnets", "v1alpha2")
}

// TestCreateClusterIPv6 runs kops create cluster --zones us-test-1a --master-zones us-test-1a --ipv6
func TestCreateClusterIPv6(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ipv6", "v1alpha2")
}

// TestCreateClusterDifferentAMIs runs kops create cluster with different AMI inputs
func TestCreateClusterDifferentAMIs(t *testing.T) {
	featureflag.ParseFlags("+APIServerNodes")
	unsetFeatureFlags := func() {
		featureflag.ParseFlags("-APIServerNodes")
	}
	defer unsetFeatureFlags()
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/different-amis", "v1alpha2")
}

// TestCreateClusterKarpenter runs kops create cluster --instance-manager=karpenter
func TestCreateClusterKarpenter(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/karpenter", "v1alpha2")
}

// TestCreateClusterZeroNodes runs kops create cluster --node-count=0
func TestCreateClusterZeroNodes(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/zero-nodes", "v1alpha2")
}

func runCreateClusterIntegrationTest(t *testing.T, srcDir string, version string) {
	ctx := context.Background()

	var stdout bytes.Buffer

	optionsYAML := "options.yaml"
	expectedClusterPath := "expected-" + version + ".yaml"

	factoryOptions := &util.FactoryOptions{}
	factoryOptions.RegistryPath = "memfs://tests"

	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	h.SetupMockAWS()
	h.SetupMockGCE()
	testutils.SetupMockOpenstack()

	cloudTags := map[string]string{}
	awsCloud, _ := awsup.NewAWSCloud("us-test-1", cloudTags)
	(awsCloud.EC2().(*mockec2.MockEC2)).CreateVpcWithId(&ec2.CreateVpcInput{
		CidrBlock: aws.String("10.0.0.0/12"),
	}, "vpc-12345678")

	awsCloud.EC2().CreateSubnet(ctx, &ec2.CreateSubnetInput{
		AvailabilityZone: aws.String("us-test-1a"),
		VpcId:            aws.String("vpc-12345678"),
		CidrBlock:        aws.String("10.10.0.0/24"),
	})

	awsCloud.EC2().CreateSubnet(ctx, &ec2.CreateSubnetInput{
		AvailabilityZone: aws.String("us-test-1a"),
		VpcId:            aws.String("vpc-12345678"),
		CidrBlock:        aws.String("10.11.0.0/24"),
	})

	publicKeyPath := path.Join(h.TempDir, "id_rsa.pub")
	privateKeyPath := path.Join(h.TempDir, "id_rsa")
	{
		if err := MakeSSHKeyPair(publicKeyPath, privateKeyPath); err != nil {
			t.Fatalf("error making SSH keypair: %v", err)
		}
	}

	factory := util.NewFactory(factoryOptions)

	{
		optionsBytes, err := os.ReadFile(path.Join(srcDir, optionsYAML))
		if err != nil {
			t.Fatalf("error reading options file: %v", err)
		}

		options := &CreateClusterOptions{}
		options.InitDefaults()

		err = kops.ParseRawYaml(optionsBytes, options)
		if err != nil {
			t.Fatalf("error parsing options: %v", err)
		}

		// No preview
		options.Target = ""

		// Use the public key we produced
		{
			publicKey, err := os.ReadFile(publicKeyPath)
			if err != nil {
				t.Fatalf("error reading public key %q: %v", publicKeyPath, err)
			}
			sshPublicKeys := make(map[string][]byte)
			sshPublicKeys[fi.SecretNameSSHPrimary] = publicKey
			options.SSHPublicKeys = sshPublicKeys
		}

		err = RunCreateCluster(ctx, factory, &stdout, options)
		if err != nil {
			t.Fatalf("error running create cluster: %v", err)
		}
	}

	clientset, err := factory.KopsClient()
	if err != nil {
		t.Fatalf("error getting clientset: %v", err)
	}

	// Compare cluster
	clusters, err := clientset.ListClusters(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("error listing clusters: %v", err)
	}

	if len(clusters.Items) != 1 {
		t.Fatalf("expected one cluster, found %d", len(clusters.Items))
	}

	var yamlAll []string

	for _, cluster := range clusters.Items {
		cluster.ObjectMeta.CreationTimestamp = MagicTimestamp
		actualYAMLBytes, err := kopscodecs.ToVersionedYamlWithVersion(&cluster, schema.GroupVersion{Group: "kops.k8s.io", Version: version})
		if err != nil {
			t.Fatalf("unexpected error serializing cluster: %v", err)
		}
		actualYAML := strings.TrimSpace(string(actualYAMLBytes))

		yamlAll = append(yamlAll, actualYAML)
	}

	// Compare instance groups

	instanceGroups, err := clientset.InstanceGroupsFor(&clusters.Items[0]).List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("error listing instance groups: %v", err)
	}

	for _, ig := range instanceGroups.Items {
		ig.ObjectMeta.CreationTimestamp = MagicTimestamp

		actualYAMLBytes, err := kopscodecs.ToVersionedYamlWithVersion(&ig, schema.GroupVersion{Group: "kops.k8s.io", Version: version})
		if err != nil {
			t.Fatalf("unexpected error serializing InstanceGroup: %v", err)
		}

		actualYAML := strings.TrimSpace(string(actualYAMLBytes))

		yamlAll = append(yamlAll, actualYAML)
	}

	// Compare additional objects
	addons, err := clientset.AddonsFor(&clusters.Items[0]).List(ctx)
	if err != nil {
		t.Fatalf("error listing addons: %v", err)
	}

	for _, addon := range addons {
		u := addon.ToUnstructured()

		actualYAMLBytes, err := kopscodecs.ToVersionedYamlWithVersion(u, schema.GroupVersion{Group: "kops.k8s.io", Version: version})
		if err != nil {
			t.Fatalf("unexpected error serializing Addon: %v", err)
		}

		actualYAML := strings.TrimSpace(string(actualYAMLBytes))

		yamlAll = append(yamlAll, actualYAML)
	}

	actualYAML := strings.Join(yamlAll, "\n\n---\n\n")
	golden.AssertMatchesFile(t, actualYAML, path.Join(srcDir, expectedClusterPath))
}
