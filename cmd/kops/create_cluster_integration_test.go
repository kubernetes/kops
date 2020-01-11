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
	"io/ioutil"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kops/cloudmock/aws/mockec2"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/pkg/testutils/golden"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

var MagicTimestamp = metav1.Time{Time: time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)}

// TestCreateClusterMinimal runs kops create cluster minimal.example.com --zones us-test-1a
func TestCreateClusterMinimal(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/minimal", "v1alpha1")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/minimal", "v1alpha2")
}

// TestCreateClusterOverride tests the override flag
func TestCreateClusterOverride(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/overrides", "v1alpha2")
}

// TestCreateClusterComplex runs kops create cluster, with a grab-bag of edge cases
func TestCreateClusterComplex(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/complex", "v1alpha2")
}

// TestCreateClusterHA runs kops create cluster ha.example.com --zones us-test-1a,us-test-1b,us-test-1c --master-zones us-test-1a,us-test-1b,us-test-1c
func TestCreateClusterHA(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ha", "v1alpha1")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ha", "v1alpha2")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ha_encrypt", "v1alpha1")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ha_encrypt", "v1alpha2")
}

// TestCreateClusterHAGCE runs kops create cluster ha-gce.example.com --cloud gce --zones us-test1-a,us-test1-b,us-test1-c --master-zones us-test1-a,us-test1-b,us-test1-c
func TestCreateClusterHAGCE(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ha_gce", "v1alpha2")
}

// TestCreateClusterHASharedZones tests kops create cluster when the master count is bigger than the number of zones
func TestCreateClusterHASharedZones(t *testing.T) {
	// Cannot be expressed in v1alpha1 API:	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ha_shared_zones", "v1alpha1")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ha_shared_zones", "v1alpha2")
}

// TestCreateClusterPrivate runs kops create cluster private.example.com --zones us-test-1a --master-zones us-test-1a
func TestCreateClusterPrivate(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/private", "v1alpha1")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/private", "v1alpha2")
}

// TestCreateClusterWithNGWSpecified runs kops create cluster private.example.com --zones us-test-1a --master-zones us-test-1a
func TestCreateClusterWithNGWSpecified(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ngwspecified", "v1alpha1")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ngwspecified", "v1alpha2")
}

// TestCreateClusterWithINGWSpecified runs kops create cluster private.example.com --zones us-test-1a --master-zones us-test-1a
func TestCreateClusterWithINGWSpecified(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ingwspecified", "v1alpha1")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/ingwspecified", "v1alpha2")
}

// TestCreateClusterSharedVPC runs kops create cluster vpc.example.com --zones us-test-1a --master-zones us-test-1a --vpc vpc-12345678
func TestCreateClusterSharedVPC(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/shared_vpc", "v1alpha1")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/shared_vpc", "v1alpha2")
}

// TestCreateClusterSharedSubnets runs kops create cluster subnet.example.com --zones us-test-1a --master-zones us-test-1a --vpc vpc-12345678 --subnets subnet-1
func TestCreateClusterSharedSubnets(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/shared_subnets", "v1alpha1")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/shared_subnets", "v1alpha2")
}

// TestCreateClusterSharedSubnetsVpcLookup runs kops create cluster subnet.example.com --zones us-test-1a --master-zones us-test-1a --vpc --subnets subnet-1
func TestCreateClusterSharedSubnetsVpcLookup(t *testing.T) {
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/shared_subnets_vpc_lookup", "v1alpha1")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/shared_subnets_vpc_lookup", "v1alpha2")
}

// TestCreateClusterPrivateSharedSubnets runs kops create cluster private-subnet.example.com --zones us-test-1a --master-zones us-test-1a --vpc vpc-12345678 --subnets subnet-1 --utility-subnets subnet-2
func TestCreateClusterPrivateSharedSubnets(t *testing.T) {
	// Cannot be expressed in v1alpha1 API: runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/private_shared_subnets", "v1alpha1")
	runCreateClusterIntegrationTest(t, "../../tests/integration/create_cluster/private_shared_subnets", "v1alpha2")
}

func runCreateClusterIntegrationTest(t *testing.T, srcDir string, version string) {
	var stdout bytes.Buffer

	optionsYAML := "options.yaml"
	expectedClusterPath := "expected-" + version + ".yaml"

	factoryOptions := &util.FactoryOptions{}
	factoryOptions.RegistryPath = "memfs://tests"

	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	h.SetupMockAWS()
	h.SetupMockGCE()

	cloudTags := map[string]string{}
	awsCloud, _ := awsup.NewAWSCloud("us-test-1", cloudTags)
	(awsCloud.EC2().(*mockec2.MockEC2)).CreateVpcWithId(&ec2.CreateVpcInput{
		CidrBlock: aws.String("10.0.0.0/12"),
	}, "vpc-12345678")

	awsCloud.EC2().CreateSubnet(&ec2.CreateSubnetInput{
		AvailabilityZone: aws.String("us-test-1a"),
		VpcId:            aws.String("vpc-12345678"),
		CidrBlock:        aws.String("10.10.0.0/24"),
	})

	awsCloud.EC2().CreateSubnet(&ec2.CreateSubnetInput{
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
		optionsBytes, err := ioutil.ReadFile(path.Join(srcDir, optionsYAML))
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
			publicKey, err := ioutil.ReadFile(publicKeyPath)
			if err != nil {
				t.Fatalf("error reading public key %q: %v", publicKeyPath, err)
			}
			sshPublicKeys := make(map[string][]byte)
			sshPublicKeys[fi.SecretNameSSHPrimary] = publicKey
			options.SSHPublicKeys = sshPublicKeys
		}

		err = RunCreateCluster(factory, &stdout, options)
		if err != nil {
			t.Fatalf("error running create cluster: %v", err)
		}
	}

	clientset, err := factory.Clientset()
	if err != nil {
		t.Fatalf("error getting clientset: %v", err)
	}

	// Compare cluster
	clusters, err := clientset.ListClusters(metav1.ListOptions{})
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

	instanceGroups, err := clientset.InstanceGroupsFor(&clusters.Items[0]).List(metav1.ListOptions{})
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

	actualYAML := strings.Join(yamlAll, "\n\n---\n\n")
	golden.AssertMatchesFile(t, actualYAML, path.Join(srcDir, expectedClusterPath))
}
