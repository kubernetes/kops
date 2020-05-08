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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/jsonutils"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/pkg/testutils/golden"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"

	"github.com/ghodss/yaml"
	"golang.org/x/crypto/ssh"
)

// updateClusterTestBase is added automatically to the srcDir on all
// tests using runTest, including runTestTerraformAWS, runTestTerraformGCE
const updateClusterTestBase = "../../tests/integration/update_cluster/"

type integrationTest struct {
	clusterName        string
	srcDir             string
	version            string
	private            bool
	zones              int
	expectPolicies     bool
	launchTemplate     bool
	lifecycleOverrides []string
	sshKey             bool
	jsonOutput         bool
	bastionUserData    bool
}

func newIntegrationTest(clusterName, srcDir string) *integrationTest {
	return &integrationTest{
		clusterName:    clusterName,
		srcDir:         srcDir,
		version:        "v1alpha2",
		zones:          1,
		expectPolicies: true,
		sshKey:         true,
	}
}

func (i *integrationTest) withVersion(version string) *integrationTest {
	i.version = version
	return i
}

func (i *integrationTest) withZones(zones int) *integrationTest {
	i.zones = zones
	return i
}

func (i *integrationTest) withoutSSHKey() *integrationTest {
	i.sshKey = false
	return i
}

func (i *integrationTest) withoutPolicies() *integrationTest {
	i.expectPolicies = false
	return i
}

func (i *integrationTest) withLifecycleOverrides(lco []string) *integrationTest {
	i.lifecycleOverrides = lco
	return i
}

func (i *integrationTest) withJSONOutput() *integrationTest {
	i.jsonOutput = true
	return i
}

func (i *integrationTest) withPrivate() *integrationTest {
	i.private = true
	return i
}

func (i *integrationTest) withLaunchTemplate() *integrationTest {
	i.launchTemplate = true
	return i
}

func (i *integrationTest) withBastionUserData() *integrationTest {
	i.bastionUserData = true
	return i
}

// TestMinimal runs the test on a minimum configuration, similar to kops create cluster minimal.example.com --zones us-west-1a
func TestMinimal(t *testing.T) {
	newIntegrationTest("minimal.example.com", "minimal").runTestTerraformAWS(t)
}

// TestMinimalGCE runs tests on a minimal GCE configuration
func TestMinimalGCE(t *testing.T) {
	newIntegrationTest("minimal-gce.example.com", "minimal_gce").runTestTerraformGCE(t)
}

// TestRestrictAccess runs the test on a simple SG configuration, similar to kops create cluster minimal.example.com --ssh-access=$(IPS) --admin-access=$(IPS) --master-count=3
func TestRestrictAccess(t *testing.T) {
	newIntegrationTest("restrictaccess.example.com", "restrict_access").runTestTerraformAWS(t)
}

// TestHA runs the test on a simple HA configuration, similar to kops create cluster minimal.example.com --zones us-west-1a,us-west-1b,us-west-1c --master-count=3
func TestHA(t *testing.T) {
	newIntegrationTest("ha.example.com", "ha").withZones(3).runTestTerraformAWS(t)
}

// TestHighAvailabilityGCE runs the test on a simple HA GCE configuration, similar to kops create cluster ha-gce.example.com
// --zones us-test1-a,us-test1-b,us-test1-c --master-count=3
func TestHighAvailabilityGCE(t *testing.T) {
	newIntegrationTest("ha-gce.example.com", "ha_gce").withZones(3).runTestTerraformGCE(t)
}

// TestComplex runs the test on a more complex configuration, intended to hit more of the edge cases
func TestComplex(t *testing.T) {
	newIntegrationTest("complex.example.com", "complex").runTestTerraformAWS(t)
	newIntegrationTest("complex.example.com", "complex").runTestCloudformation(t)
	newIntegrationTest("complex.example.com", "complex").withVersion("legacy-v1alpha2").runTestTerraformAWS(t)
}

// TestExternalPolicies tests external policies output
func TestExternalPolicies(t *testing.T) {
	newIntegrationTest("externalpolicies.example.com", "externalpolicies").runTestTerraformAWS(t)
}

func TestNoSSHKey(t *testing.T) {
	newIntegrationTest("nosshkey.example.com", "nosshkey").withoutSSHKey().runTestTerraformAWS(t)
	newIntegrationTest("nosshkey.example.com", "nosshkey-cloudformation").withoutSSHKey().runTestCloudformation(t)
}

// TestCrossZone tests that the cross zone setting on the API ELB is set properly
func TestCrossZone(t *testing.T) {
	newIntegrationTest("crosszone.example.com", "api_elb_cross_zone").runTestTerraformAWS(t)
}

// TestMinimalCloudformation runs the test on a minimum configuration, similar to kops create cluster minimal.example.com --zones us-west-1a
func TestMinimalCloudformation(t *testing.T) {
	newIntegrationTest("minimal.example.com", "minimal-cloudformation").runTestCloudformation(t)
}

// TestExistingIAMCloudformation runs the test with existing IAM instance profiles, similar to kops create cluster minimal.example.com --zones us-west-1a
func TestExistingIAMCloudformation(t *testing.T) {
	lifecycleOverrides := []string{"IAMRole=ExistsAndWarnIfChanges", "IAMRolePolicy=ExistsAndWarnIfChanges", "IAMInstanceProfileRole=ExistsAndWarnIfChanges"}
	newIntegrationTest("minimal.example.com", "existing_iam_cloudformation").withLifecycleOverrides(lifecycleOverrides).runTestCloudformation(t)
}

// TestExistingSG runs the test with existing Security Group, similar to kops create cluster minimal.example.com --zones us-west-1a
func TestExistingSG(t *testing.T) {
	newIntegrationTest("existingsg.example.com", "existing_sg").withZones(3).runTestTerraformAWS(t)
}

// TestAdditionalUserData runs the test on passing additional user-data to an instance at bootstrap.
func TestAdditionalUserData(t *testing.T) {
	newIntegrationTest("additionaluserdata.example.com", "additional_user-data").runTestCloudformation(t)
}

// TestBastionAdditionalUserData runs the test on passing additional user-data to a bastion instance group
func TestBastionAdditionalUserData(t *testing.T) {
	newIntegrationTest("bastionuserdata.example.com", "bastionadditional_user-data").withPrivate().withBastionUserData().runTestTerraformAWS(t)
}

// TestMinimalJSON runs the test on a minimal data set and outputs JSON
func TestMinimalJSON(t *testing.T) {
	featureflag.ParseFlags("+TerraformJSON,-Terraform-0.12")
	unsetFeaureFlag := func() {
		featureflag.ParseFlags("-TerraformJSON,+Terraform-0.12")
	}
	defer unsetFeaureFlag()
	newIntegrationTest("minimal-json.example.com", "minimal-json").withJSONOutput().runTestTerraformAWS(t)
}

func TestMinimalTerraform011(t *testing.T) {
	featureflag.ParseFlags("-Terraform-0.12")
	unsetFeaureFlag := func() {
		featureflag.ParseFlags("+Terraform-0.12")
	}
	defer unsetFeaureFlag()
	newIntegrationTest("minimal-tf11.example.com", "minimal-tf11").runTestTerraformAWS(t)
}

// TestPrivateWeave runs the test on a configuration with private topology, weave networking
func TestPrivateWeave(t *testing.T) {
	newIntegrationTest("privateweave.example.com", "privateweave").withPrivate().runTestTerraformAWS(t)
}

// TestPrivateFlannel runs the test on a configuration with private topology, flannel networking
func TestPrivateFlannel(t *testing.T) {
	newIntegrationTest("privateflannel.example.com", "privateflannel").withPrivate().runTestTerraformAWS(t)
}

// TestPrivateCalico runs the test on a configuration with private topology, calico networking
func TestPrivateCalico(t *testing.T) {
	newIntegrationTest("privatecalico.example.com", "privatecalico").withPrivate().runTestTerraformAWS(t)
	newIntegrationTest("privatecalico.example.com", "privatecalico").withPrivate().runTestCloudformation(t)
}

func TestPrivateCilium(t *testing.T) {
	newIntegrationTest("privatecilium.example.com", "privatecilium").withPrivate().runTestTerraformAWS(t)
	newIntegrationTest("privatecilium.example.com", "privatecilium").withPrivate().runTestCloudformation(t)
}

func TestPrivateCiliumAdvanced(t *testing.T) {
	newIntegrationTest("privateciliumadvanced.example.com", "privateciliumadvanced").withPrivate().runTestTerraformAWS(t)
	newIntegrationTest("privateciliumadvanced.example.com", "privateciliumadvanced").withPrivate().runTestCloudformation(t)
}

// TestPrivateCanal runs the test on a configuration with private topology, canal networking
func TestPrivateCanal(t *testing.T) {
	newIntegrationTest("privatecanal.example.com", "privatecanal").withPrivate().runTestTerraformAWS(t)
}

// TestPrivateKopeio runs the test on a configuration with private topology, kopeio networking
func TestPrivateKopeio(t *testing.T) {
	newIntegrationTest("privatekopeio.example.com", "privatekopeio").withPrivate().runTestTerraformAWS(t)
}

// TestUnmanaged is a test where all the subnets opt-out of route management
func TestUnmanaged(t *testing.T) {
	newIntegrationTest("unmanaged.example.com", "unmanaged").withPrivate().runTestTerraformAWS(t)
}

// TestPrivateSharedSubnet runs the test on a configuration with private topology & shared subnets
func TestPrivateSharedSubnet(t *testing.T) {
	newIntegrationTest("private-shared-subnet.example.com", "private-shared-subnet").withPrivate().runTestTerraformAWS(t)
}

// TestPrivateDns1 runs the test on a configuration with private topology, private dns
func TestPrivateDns1(t *testing.T) {
	newIntegrationTest("privatedns1.example.com", "privatedns1").withPrivate().runTestTerraformAWS(t)
}

// TestPrivateDns2 runs the test on a configuration with private topology, private dns, extant vpc
func TestPrivateDns2(t *testing.T) {
	newIntegrationTest("privatedns2.example.com", "privatedns2").withPrivate().runTestTerraformAWS(t)
}

// TestSharedSubnet runs the test on a configuration with a shared subnet (and VPC)
func TestSharedSubnet(t *testing.T) {
	newIntegrationTest("sharedsubnet.example.com", "shared_subnet").runTestTerraformAWS(t)
}

// TestSharedVPC runs the test on a configuration with a shared VPC
func TestSharedVPC(t *testing.T) {
	newIntegrationTest("sharedvpc.example.com", "shared_vpc").runTestTerraformAWS(t)
}

// TestExistingIAM runs the test on a configuration with existing IAM instance profiles
func TestExistingIAM(t *testing.T) {
	lifecycleOverrides := []string{"IAMRole=ExistsAndWarnIfChanges", "IAMRolePolicy=ExistsAndWarnIfChanges", "IAMInstanceProfileRole=ExistsAndWarnIfChanges"}
	newIntegrationTest("existing-iam.example.com", "existing_iam").withZones(3).withoutPolicies().withLifecycleOverrides(lifecycleOverrides).runTestTerraformAWS(t)
}

// TestAdditionalCIDR runs the test on a configuration with a shared VPC
func TestAdditionalCIDR(t *testing.T) {
	newIntegrationTest("additionalcidr.example.com", "additional_cidr").withVersion("v1alpha3").withZones(3).runTestTerraformAWS(t)
	newIntegrationTest("additionalcidr.example.com", "additional_cidr").runTestCloudformation(t)
}

// TestPhaseNetwork tests the output of tf for the network phase
func TestPhaseNetwork(t *testing.T) {
	newIntegrationTest("lifecyclephases.example.com", "lifecycle_phases").runTestPhase(t, cloudup.PhaseNetwork)
}

func TestExternalLoadBalancer(t *testing.T) {
	newIntegrationTest("externallb.example.com", "externallb").runTestTerraformAWS(t)
	newIntegrationTest("externallb.example.com", "externallb").runTestCloudformation(t)
}

// TestPhaseIAM tests the output of tf for the iam phase
func TestPhaseIAM(t *testing.T) {
	t.Skip("unable to test w/o allowing failed validation")
	newIntegrationTest("lifecyclephases.example.com", "lifecycle_phases").runTestPhase(t, cloudup.PhaseSecurity)
}

// TestPhaseCluster tests the output of tf for the cluster phase
func TestPhaseCluster(t *testing.T) {
	// TODO fix tf for phase, and allow override on validation
	t.Skip("unable to test w/o allowing failed validation")
	newIntegrationTest("lifecyclephases.example.com", "lifecycle_phases").runTestPhase(t, cloudup.PhaseCluster)
}

// TestMixedInstancesASG tests ASGs using a mixed instance policy
func TestMixedInstancesASG(t *testing.T) {
	newIntegrationTest("mixedinstances.example.com", "mixed_instances").withZones(3).withLaunchTemplate().runTestTerraformAWS(t)
	newIntegrationTest("mixedinstances.example.com", "mixed_instances").withZones(3).withLaunchTemplate().runTestCloudformation(t)
}

// TestMixedInstancesSpotASG tests ASGs using a mixed instance policy and spot instances
func TestMixedInstancesSpotASG(t *testing.T) {
	newIntegrationTest("mixedinstances.example.com", "mixed_instances_spot").withZones(3).withLaunchTemplate().runTestTerraformAWS(t)
	newIntegrationTest("mixedinstances.example.com", "mixed_instances_spot").withZones(3).withLaunchTemplate().runTestCloudformation(t)
}

// TestContainerdCloudformation runs the test on a containerd configuration
func TestContainerdCloudformation(t *testing.T) {
	newIntegrationTest("containerd.example.com", "containerd-cloudformation").runTestCloudformation(t)
}

// TestLaunchTemplatesASG tests ASGs using launch templates instead of launch configurations
func TestLaunchTemplatesASG(t *testing.T) {
	featureflag.ParseFlags("+EnableLaunchTemplates")
	unsetFeaureFlag := func() {
		featureflag.ParseFlags("-EnableLaunchTemplates")
	}
	defer unsetFeaureFlag()

	newIntegrationTest("launchtemplates.example.com", "launch_templates").withZones(3).withLaunchTemplate().runTestTerraformAWS(t)
	newIntegrationTest("launchtemplates.example.com", "launch_templates").withZones(3).withLaunchTemplate().runTestCloudformation(t)
}

func (i *integrationTest) runTest(t *testing.T, h *testutils.IntegrationTestHarness, expectedDataFilenames []string, tfFileName string, expectedTfFileName string, phase *cloudup.Phase) {
	ctx := context.Background()

	var stdout bytes.Buffer

	i.srcDir = updateClusterTestBase + i.srcDir
	inputYAML := "in-" + i.version + ".yaml"
	testDataTFPath := "kubernetes.tf"
	actualTFPath := "kubernetes.tf"

	if tfFileName != "" {
		testDataTFPath = tfFileName
	}

	if expectedTfFileName != "" {
		actualTFPath = expectedTfFileName
	}

	factoryOptions := &util.FactoryOptions{}
	factoryOptions.RegistryPath = "memfs://tests"

	factory := util.NewFactory(factoryOptions)

	{
		options := &CreateOptions{}
		options.Filenames = []string{path.Join(i.srcDir, inputYAML)}

		err := RunCreate(ctx, factory, &stdout, options)
		if err != nil {
			t.Fatalf("error running %q create: %v", inputYAML, err)
		}
	}

	if i.sshKey {
		options := &CreateSecretPublickeyOptions{}
		options.ClusterName = i.clusterName
		options.Name = "admin"
		options.PublicKeyPath = path.Join(i.srcDir, "id_rsa.pub")

		err := RunCreateSecretPublicKey(ctx, factory, &stdout, options)
		if err != nil {
			t.Fatalf("error running %q create: %v", inputYAML, err)
		}
	}

	{
		options := &UpdateClusterOptions{}
		options.InitDefaults()
		options.Target = "terraform"
		options.OutDir = path.Join(h.TempDir, "out")
		options.RunTasksOptions.MaxTaskDuration = 30 * time.Second
		if phase != nil {
			options.Phase = string(*phase)
		}

		// We don't test it here, and it adds a dependency on kubectl
		options.CreateKubecfg = false

		options.LifecycleOverrides = i.lifecycleOverrides

		_, err := RunUpdateCluster(ctx, factory, i.clusterName, &stdout, options)
		if err != nil {
			t.Fatalf("error running update cluster %q: %v", i.clusterName, err)
		}
	}

	// Compare main files
	{
		files, err := ioutil.ReadDir(path.Join(h.TempDir, "out"))
		if err != nil {
			t.Fatalf("failed to read dir: %v", err)
		}

		var fileNames []string
		for _, f := range files {
			fileNames = append(fileNames, f.Name())
		}
		sort.Strings(fileNames)

		actualFilenames := strings.Join(fileNames, ",")
		expectedFilenames := actualTFPath

		if len(expectedDataFilenames) > 0 {
			expectedFilenames = "data," + actualTFPath
		}

		if actualFilenames != expectedFilenames {
			t.Fatalf("unexpected files.  actual=%q, expected=%q, test=%q", actualFilenames, expectedFilenames, testDataTFPath)
		}

		actualTF, err := ioutil.ReadFile(path.Join(h.TempDir, "out", actualTFPath))
		if err != nil {
			t.Fatalf("unexpected error reading actual terraform output: %v", err)
		}

		golden.AssertMatchesFile(t, string(actualTF), path.Join(i.srcDir, testDataTFPath))
	}

	// Compare data files if they are provided
	if len(expectedDataFilenames) > 0 {
		actualDataPath := path.Join(h.TempDir, "out", "data")
		files, err := ioutil.ReadDir(actualDataPath)
		if err != nil {
			t.Fatalf("failed to read data dir: %v", err)
		}

		var actualDataFilenames []string
		for _, f := range files {
			actualDataFilenames = append(actualDataFilenames, f.Name())
		}

		sort.Strings(expectedDataFilenames)
		if !reflect.DeepEqual(actualDataFilenames, expectedDataFilenames) {
			for j := 0; j < len(actualDataFilenames) && j < len(expectedDataFilenames); j++ {
				if actualDataFilenames[j] != expectedDataFilenames[j] {
					t.Errorf("diff @%d: %q vs %q", j, actualDataFilenames[j], expectedDataFilenames[j])
					break
				}
			}
			t.Fatalf("unexpected data files.  actual=%q, expected=%q", actualDataFilenames, expectedDataFilenames)
		}

		// Some tests might provide _some_ tf data files (not necessarily all that
		// are actually produced), validate that the provided expected data file
		// contents match actual data file content
		expectedDataPath := path.Join(i.srcDir, "data")
		{
			for _, dataFileName := range expectedDataFilenames {
				actualDataContent, err :=
					ioutil.ReadFile(path.Join(actualDataPath, dataFileName))
				if err != nil {
					t.Fatalf("failed to read actual data file: %v", err)
				}
				golden.AssertMatchesFile(t, string(actualDataContent), path.Join(expectedDataPath, dataFileName))
			}
		}
	}
}

func (i *integrationTest) runTestTerraformAWS(t *testing.T) {
	tfFileName := ""
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	if i.jsonOutput {
		tfFileName = "kubernetes.tf.json"
	}

	h.MockKopsVersion("1.15.0")
	h.SetupMockAWS()

	expectedFilenames := []string{}

	if i.launchTemplate {
		expectedFilenames = append(expectedFilenames, "aws_launch_template_nodes."+i.clusterName+"_user_data")
	} else {
		expectedFilenames = append(expectedFilenames, "aws_launch_configuration_nodes."+i.clusterName+"_user_data")
	}
	if i.sshKey {
		expectedFilenames = append(expectedFilenames, "aws_key_pair_kubernetes."+i.clusterName+"-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")
	}

	for j := 0; j < i.zones; j++ {
		zone := "us-test-1" + string([]byte{byte('a') + byte(j)})
		if featureflag.EnableLaunchTemplates.Enabled() {
			expectedFilenames = append(expectedFilenames, "aws_launch_template_master-"+zone+".masters."+i.clusterName+"_user_data")
		} else {
			expectedFilenames = append(expectedFilenames, "aws_launch_configuration_master-"+zone+".masters."+i.clusterName+"_user_data")
		}
	}

	if i.expectPolicies {
		expectedFilenames = append(expectedFilenames, []string{
			"aws_iam_role_masters." + i.clusterName + "_policy",
			"aws_iam_role_nodes." + i.clusterName + "_policy",
			"aws_iam_role_policy_masters." + i.clusterName + "_policy",
			"aws_iam_role_policy_nodes." + i.clusterName + "_policy",
		}...)
		if i.private {
			expectedFilenames = append(expectedFilenames, []string{
				"aws_iam_role_bastions." + i.clusterName + "_policy",
				"aws_iam_role_policy_bastions." + i.clusterName + "_policy",
			}...)
			if i.bastionUserData {
				expectedFilenames = append(expectedFilenames, "aws_launch_configuration_bastion."+i.clusterName+"_user_data")
			}
		}
	}
	i.runTest(t, h, expectedFilenames, tfFileName, tfFileName, nil)
}

func (i *integrationTest) runTestPhase(t *testing.T, phase cloudup.Phase) {
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	h.MockKopsVersion("1.15.0")
	h.SetupMockAWS()
	phaseName := string(phase)
	if phaseName == "" {
		t.Fatalf("phase must be set")
	}
	tfFileName := phaseName + "-kubernetes.tf"

	expectedFilenames := []string{}

	if phase == cloudup.PhaseSecurity {
		expectedFilenames = []string{
			"aws_iam_role_masters." + i.clusterName + "_policy",
			"aws_iam_role_nodes." + i.clusterName + "_policy",
			"aws_iam_role_policy_masters." + i.clusterName + "_policy",
			"aws_iam_role_policy_nodes." + i.clusterName + "_policy",
			"aws_key_pair_kubernetes." + i.clusterName + "-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key",
		}
		if i.private {
			expectedFilenames = append(expectedFilenames, []string{
				"aws_iam_role_bastions." + i.clusterName + "_policy",
				"aws_iam_role_policy_bastions." + i.clusterName + "_policy",
				"aws_launch_configuration_bastion." + i.clusterName + "_user_data",
			}...)
		}
	} else if phase == cloudup.PhaseCluster {
		expectedFilenames = []string{
			"aws_launch_configuration_nodes." + i.clusterName + "_user_data",
		}

		for j := 0; j < i.zones; j++ {
			zone := "us-test-1" + string([]byte{byte('a') + byte(j)})
			s := "aws_launch_configuration_master-" + zone + ".masters." + i.clusterName + "_user_data"
			expectedFilenames = append(expectedFilenames, s)
		}
	}

	i.runTest(t, h, expectedFilenames, tfFileName, "", &phase)
}

func (i *integrationTest) runTestTerraformGCE(t *testing.T) {
	featureflag.ParseFlags("+AlphaAllowGCE")

	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	h.MockKopsVersion("1.15.0")
	h.SetupMockGCE()

	expectedFilenames := []string{
		"google_compute_instance_template_nodes-" + gce.SafeClusterName(i.clusterName) + "_metadata_startup-script",
		"google_compute_instance_template_nodes-" + gce.SafeClusterName(i.clusterName) + "_metadata_ssh-keys",
	}

	for j := 0; j < i.zones; j++ {
		zone := "us-test1-" + string([]byte{byte('a') + byte(j)})
		prefix := "google_compute_instance_template_master-" + zone + "-" + gce.SafeClusterName(i.clusterName) + "_metadata_"

		expectedFilenames = append(expectedFilenames, prefix+"startup-script")
		expectedFilenames = append(expectedFilenames, prefix+"ssh-keys")
	}

	i.runTest(t, h, expectedFilenames, "", "", nil)
}

func (i *integrationTest) runTestCloudformation(t *testing.T) {
	ctx := context.Background()

	i.srcDir = updateClusterTestBase + i.srcDir
	var stdout bytes.Buffer

	inputYAML := "in-" + i.version + ".yaml"
	expectedCfPath := "cloudformation.json"

	factoryOptions := &util.FactoryOptions{}
	factoryOptions.RegistryPath = "memfs://tests"

	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	h.MockKopsVersion("1.15.0")
	h.SetupMockAWS()

	factory := util.NewFactory(factoryOptions)

	{
		options := &CreateOptions{}
		options.Filenames = []string{path.Join(i.srcDir, inputYAML)}

		err := RunCreate(ctx, factory, &stdout, options)
		if err != nil {
			t.Fatalf("error running %q create: %v", inputYAML, err)
		}
	}

	if i.sshKey {
		options := &CreateSecretPublickeyOptions{}
		options.ClusterName = i.clusterName
		options.Name = "admin"
		options.PublicKeyPath = path.Join(i.srcDir, "id_rsa.pub")

		err := RunCreateSecretPublicKey(ctx, factory, &stdout, options)
		if err != nil {
			t.Fatalf("error running %q create: %v", inputYAML, err)
		}
	}

	{
		options := &UpdateClusterOptions{}
		options.InitDefaults()
		options.Target = "cloudformation"
		options.OutDir = path.Join(h.TempDir, "out")
		options.RunTasksOptions.MaxTaskDuration = 30 * time.Second

		// We don't test it here, and it adds a dependency on kubectl
		options.CreateKubecfg = false
		options.LifecycleOverrides = i.lifecycleOverrides

		_, err := RunUpdateCluster(ctx, factory, i.clusterName, &stdout, options)
		if err != nil {
			t.Fatalf("error running update cluster %q: %v", i.clusterName, err)
		}
	}

	// Compare main files
	{
		files, err := ioutil.ReadDir(path.Join(h.TempDir, "out"))
		if err != nil {
			t.Fatalf("failed to read dir: %v", err)
		}

		var fileNames []string
		for _, f := range files {
			fileNames = append(fileNames, f.Name())
		}
		sort.Strings(fileNames)

		actualFilenames := strings.Join(fileNames, ",")
		expectedFilenames := "kubernetes.json"
		if actualFilenames != expectedFilenames {
			t.Fatalf("unexpected files.  actual=%q, expected=%q", actualFilenames, expectedFilenames)
		}

		actualPath := path.Join(h.TempDir, "out", "kubernetes.json")
		actualCF, err := ioutil.ReadFile(actualPath)
		if err != nil {
			t.Fatalf("unexpected error reading actual cloudformation output: %v", err)
		}

		// Expand out the UserData base64 blob, as otherwise testing is painful
		extracted := make(map[string]string)
		var buf bytes.Buffer
		out := jsonutils.NewJSONStreamWriter(&buf)
		in := json.NewDecoder(bytes.NewReader(actualCF))
		for {
			token, err := in.Token()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					t.Fatalf("unexpected error parsing cloudformation output: %v", err)
				}
			}

			if strings.HasSuffix(out.Path(), ".UserData") {
				if s, ok := token.(string); ok {
					vBytes, err := base64.StdEncoding.DecodeString(s)
					if err != nil {
						t.Fatalf("error decoding UserData: %v", err)
					} else {
						extracted[out.Path()] = string(vBytes)
						token = json.Token("extracted")
					}
				}
			}

			if err := out.WriteToken(token); err != nil {
				t.Fatalf("error writing json: %v", err)
			}
		}
		actualCF = buf.Bytes()

		golden.AssertMatchesFile(t, string(actualCF), path.Join(i.srcDir, expectedCfPath))

		// test extracted values
		{
			actual := make(map[string]string)

			for k, v := range extracted {
				// Strip carriage return as expectedValue is stored in a yaml string literal
				// and yaml block quoting doesn't seem to support \r in a string
				v = strings.Replace(v, "\r", "", -1)

				actual[k] = v
			}

			actualExtracted, err := yaml.Marshal(actual)
			if err != nil {
				t.Fatalf("error serializing yaml: %v", err)
			}

			golden.AssertMatchesFile(t, string(actualExtracted), path.Join(i.srcDir, expectedCfPath+".extracted.yaml"))
		}

		golden.AssertMatchesFile(t, string(actualCF), path.Join(i.srcDir, expectedCfPath))
	}
}

func MakeSSHKeyPair(publicKeyPath string, privateKeyPath string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return err
	}

	var privateKeyBytes bytes.Buffer
	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	if err := pem.Encode(&privateKeyBytes, privateKeyPEM); err != nil {
		return err
	}
	if err := ioutil.WriteFile(privateKeyPath, privateKeyBytes.Bytes(), os.FileMode(0700)); err != nil {
		return err
	}

	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(publicKey)
	if err := ioutil.WriteFile(publicKeyPath, publicKeyBytes, os.FileMode(0744)); err != nil {
		return err
	}

	return nil
}
