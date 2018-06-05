/*
Copyright 2016 The Kubernetes Authors.

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
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/jsonutils"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"

	"github.com/ghodss/yaml"
	"golang.org/x/crypto/ssh"
)

// updateClusterTestBase is added automatically to the srcDir on all
// tests using runTest, including runTestAWS, runTestGCE
const updateClusterTestBase = "../../tests/integration/update_cluster/"

// TestMinimal runs the test on a minimum configuration, similar to kops create cluster minimal.example.com --zones us-west-1a
func TestMinimal(t *testing.T) {
	runTestAWS(t, "minimal.example.com", "minimal", "v1alpha0", false, 1)
	runTestAWS(t, "minimal.example.com", "minimal", "v1alpha1", false, 1)
	runTestAWS(t, "minimal.example.com", "minimal", "v1alpha2", false, 1)
}

// TestHA runs the test on a simple HA configuration, similar to kops create cluster minimal.example.com --zones us-west-1a,us-west-1b,us-west-1c --master-count=3
func TestHA(t *testing.T) {
	runTestAWS(t, "ha.example.com", "ha", "v1alpha1", false, 3)
	runTestAWS(t, "ha.example.com", "ha", "v1alpha2", false, 3)
}

// TestHighAvailabilityGCE runs the test on a simple HA GCE configuration, similar to kops create cluster ha-gce.example.com
// --zones us-test1-a,us-test1-b,us-test1-c --master-count=3
func TestHighAvailabilityGCE(t *testing.T) {
	runTestGCE(t, "ha-gce.example.com", "ha_gce", "v1alpha2", false, 3)
}

// TestComplex runs the test on a more complex configuration, intended to hit more of the edge cases
func TestComplex(t *testing.T) {
	runTestAWS(t, "complex.example.com", "complex", "v1alpha2", false, 1)
}

// TestMinimalCloudformation runs the test on a minimum configuration, similar to kops create cluster minimal.example.com --zones us-west-1a
func TestMinimalCloudformation(t *testing.T) {
	runTestCloudformation(t, "minimal.example.com", "minimal-cloudformation", "v1alpha2", false)
}

// TestAdditionalUserData runs the test on passing additional user-data to an instance at bootstrap.
func TestAdditionalUserData(t *testing.T) {
	runTestCloudformation(t, "additionaluserdata.example.com", "additional_user-data", "v1alpha2", false)
}

// TestBastionAdditionalUserData runs the test on passing additional user-data to a bastion instance group
func TestBastionAdditionalUserData(t *testing.T) {
	runTestAWS(t, "bastionuserdata.example.com", "bastionadditional_user-data", "v1alpha2", true, 1)
}

// TestMinimal_141 runs the test on a configuration from 1.4.1 release
func TestMinimal_141(t *testing.T) {
	runTestAWS(t, "minimal-141.example.com", "minimal-141", "v1alpha0", false, 1)
}

// TestPrivateWeave runs the test on a configuration with private topology, weave networking
func TestPrivateWeave(t *testing.T) {
	runTestAWS(t, "privateweave.example.com", "privateweave", "v1alpha1", true, 1)
	runTestAWS(t, "privateweave.example.com", "privateweave", "v1alpha2", true, 1)
}

// TestPrivateFlannel runs the test on a configuration with private topology, flannel networking
func TestPrivateFlannel(t *testing.T) {
	runTestAWS(t, "privateflannel.example.com", "privateflannel", "v1alpha1", true, 1)
	runTestAWS(t, "privateflannel.example.com", "privateflannel", "v1alpha2", true, 1)
}

// TestPrivateCalico runs the test on a configuration with private topology, calico networking
func TestPrivateCalico(t *testing.T) {
	runTestAWS(t, "privatecalico.example.com", "privatecalico", "v1alpha1", true, 1)
	runTestAWS(t, "privatecalico.example.com", "privatecalico", "v1alpha2", true, 1)
}

// TestPrivateCanal runs the test on a configuration with private topology, canal networking
func TestPrivateCanal(t *testing.T) {
	runTestAWS(t, "privatecanal.example.com", "privatecanal", "v1alpha1", true, 1)
	runTestAWS(t, "privatecanal.example.com", "privatecanal", "v1alpha2", true, 1)
}

// TestPrivateKopeio runs the test on a configuration with private topology, kopeio networking
func TestPrivateKopeio(t *testing.T) {
	runTestAWS(t, "privatekopeio.example.com", "privatekopeio", "v1alpha2", true, 1)
}

// TestPrivateSharedSubnet runs the test on a configuration with private topology & shared subnets
func TestPrivateSharedSubnet(t *testing.T) {
	runTestAWS(t, "private-shared-subnet.example.com", "private-shared-subnet", "v1alpha2", true, 1)
}

// TestPrivateDns1 runs the test on a configuration with private topology, private dns
func TestPrivateDns1(t *testing.T) {
	runTestAWS(t, "privatedns1.example.com", "privatedns1", "v1alpha2", true, 1)
}

// TestPrivateDns2 runs the test on a configuration with private topology, private dns, extant vpc
func TestPrivateDns2(t *testing.T) {
	runTestAWS(t, "privatedns2.example.com", "privatedns2", "v1alpha2", true, 1)
}

// TestSharedSubnet runs the test on a configuration with a shared subnet (and VPC)
func TestSharedSubnet(t *testing.T) {
	runTestAWS(t, "sharedsubnet.example.com", "shared_subnet", "v1alpha2", false, 1)
}

// TestSharedVPC runs the test on a configuration with a shared VPC
func TestSharedVPC(t *testing.T) {
	runTestAWS(t, "sharedvpc.example.com", "shared_vpc", "v1alpha2", false, 1)
}

// TestPhaseNetwork tests the output of tf for the network phase
func TestPhaseNetwork(t *testing.T) {
	runTestPhase(t, "lifecyclephases.example.com", "lifecycle_phases", "v1alpha2", true, 1, cloudup.PhaseNetwork)
}

// TestPhaseIAM tests the output of tf for the iam phase
func TestPhaseIAM(t *testing.T) {
	t.Skip("unable to test w/o allowing failed validation")
	runTestPhase(t, "lifecyclephases.example.com", "lifecycle_phases", "v1alpha2", true, 1, cloudup.PhaseSecurity)
}

// TestPhaseCluster tests the output of tf for the cluster phase
func TestPhaseCluster(t *testing.T) {
	// TODO fix tf for phase, and allow override on validation
	t.Skip("unable to test w/o allowing failed validation")
	runTestPhase(t, "lifecyclephases.example.com", "lifecycle_phases", "v1alpha2", true, 1, cloudup.PhaseCluster)
}

func runTest(t *testing.T, h *testutils.IntegrationTestHarness, clusterName string, srcDir string, version string, private bool, zones int, expectedDataFilenames []string, tfFileName string, phase *cloudup.Phase) {
	var stdout bytes.Buffer

	srcDir = updateClusterTestBase + srcDir
	inputYAML := "in-" + version + ".yaml"
	testDataTFPath := "kubernetes.tf"
	actualTFPath := "kubernetes.tf"

	if tfFileName != "" {
		testDataTFPath = tfFileName
	}

	factoryOptions := &util.FactoryOptions{}
	factoryOptions.RegistryPath = "memfs://tests"

	factory := util.NewFactory(factoryOptions)

	{
		options := &CreateOptions{}
		options.Filenames = []string{path.Join(srcDir, inputYAML)}

		err := RunCreate(factory, &stdout, options)
		if err != nil {
			t.Fatalf("error running %q create: %v", inputYAML, err)
		}
	}

	{
		options := &CreateSecretPublickeyOptions{}
		options.ClusterName = clusterName
		options.Name = "admin"
		options.PublicKeyPath = path.Join(srcDir, "id_rsa.pub")

		err := RunCreateSecretPublicKey(factory, &stdout, options)
		if err != nil {
			t.Fatalf("error running %q create: %v", inputYAML, err)
		}
	}

	{
		options := &UpdateClusterOptions{}
		options.InitDefaults()
		options.Target = "terraform"
		options.OutDir = path.Join(h.TempDir, "out")
		options.MaxTaskDuration = 30 * time.Second
		if phase != nil {
			options.Phase = string(*phase)
		}

		// We don't test it here, and it adds a dependency on kubectl
		options.CreateKubecfg = false

		_, err := RunUpdateCluster(factory, clusterName, &stdout, options)
		if err != nil {
			t.Fatalf("error running update cluster %q: %v", clusterName, err)
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
		expectedFilenames := "kubernetes.tf"

		if len(expectedDataFilenames) > 0 {
			expectedFilenames = "data,kubernetes.tf"
		}

		if actualFilenames != expectedFilenames {
			t.Fatalf("unexpected files.  actual=%q, expected=%q, test=%q", actualFilenames, expectedFilenames, testDataTFPath)
		}

		actualTF, err := ioutil.ReadFile(path.Join(h.TempDir, "out", actualTFPath))
		if err != nil {
			t.Fatalf("unexpected error reading actual terraform output: %v", err)
		}
		expectedTF, err := ioutil.ReadFile(path.Join(srcDir, testDataTFPath))
		if err != nil {
			t.Fatalf("unexpected error reading expected terraform output: %v", err)
		}
		expectedTF = bytes.Replace(expectedTF, []byte("\r\n"), []byte("\n"), -1)

		if !bytes.Equal(actualTF, expectedTF) {
			diffString := diff.FormatDiff(string(expectedTF), string(actualTF))
			t.Logf("diff:\n%s\n", diffString)

			if os.Getenv("HACK_UPDATE_TF_IN_PLACE") != "" {
				if err := ioutil.WriteFile(path.Join(srcDir, testDataTFPath), actualTF, 0644); err != nil {
					t.Errorf("error writing terraform output: %v", err)
				}
				t.Errorf("terraform output differed from expected")
				return // Avoid Fatalf as we want to keep going and update all files
			}
			t.Fatalf("terraform output differed from expected")
		}
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
			t.Fatalf("unexpected data files.  actual=%q, expected=%q", actualDataFilenames, expectedDataFilenames)
		}

		// Some tests might provide _some_ tf data files (not necessarilly all that
		// are actually produced), validate that the provided expected data file
		// contents match actual data file content
		expectedDataPath := path.Join(srcDir, "data")
		if _, err := os.Stat(expectedDataPath); err == nil {
			expectedDataFiles, err := ioutil.ReadDir(expectedDataPath)
			if err != nil {
				t.Fatalf("failed to read expected data dir: %v", err)
			}
			for _, expectedDataFile := range expectedDataFiles {
				dataFileName := expectedDataFile.Name()
				expectedDataContent, err :=
					ioutil.ReadFile(path.Join(expectedDataPath, dataFileName))
				if err != nil {
					t.Fatalf("failed to read expected data file: %v", err)
				}
				actualDataContent, err :=
					ioutil.ReadFile(path.Join(actualDataPath, dataFileName))
				if err != nil {
					t.Fatalf("failed to read actual data file: %v", err)
				}
				if string(expectedDataContent) != string(actualDataContent) {
					t.Fatalf(
						"actual data file (%s) did not match the content of expected data file (%s). "+
							"NOTE: If outputs seem identical, check for end-of-line differences, "+
							"especially if the file is in multipart MIME format!"+
							"\nBEGIN_ACTUAL:\n%s\nEND_ACTUAL\nBEGIN_EXPECTED:\n%s\nEND_EXPECTED",
						path.Join(actualDataPath, dataFileName),
						path.Join(expectedDataPath, dataFileName),
						actualDataContent,
						expectedDataContent,
					)
				}
			}
		}
	}
}

func runTestAWS(t *testing.T, clusterName string, srcDir string, version string, private bool, zones int) {
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	h.MockKopsVersion("1.8.1")
	h.SetupMockAWS()

	expectedFilenames := []string{
		"aws_iam_role_masters." + clusterName + "_policy",
		"aws_iam_role_nodes." + clusterName + "_policy",
		"aws_iam_role_policy_masters." + clusterName + "_policy",
		"aws_iam_role_policy_nodes." + clusterName + "_policy",
		"aws_key_pair_kubernetes." + clusterName + "-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key",
		"aws_launch_configuration_nodes." + clusterName + "_user_data",
	}

	for i := 0; i < zones; i++ {
		zone := "us-test-1" + string([]byte{byte('a') + byte(i)})
		s := "aws_launch_configuration_master-" + zone + ".masters." + clusterName + "_user_data"
		expectedFilenames = append(expectedFilenames, s)
	}

	if private {
		expectedFilenames = append(expectedFilenames, []string{
			"aws_iam_role_bastions." + clusterName + "_policy",
			"aws_iam_role_policy_bastions." + clusterName + "_policy",

			// bastions usually don't have any userdata
			// "aws_launch_configuration_bastions." + clusterName + "_user_data",
		}...)
	}
	// Special case that tests a bastion with user-data
	if srcDir == "bastionadditional_user-data" {
		expectedFilenames = append(expectedFilenames, "aws_launch_configuration_bastion."+clusterName+"_user_data")
	}
	runTest(t, h, clusterName, srcDir, version, private, zones, expectedFilenames, "", nil)
}

func runTestPhase(t *testing.T, clusterName string, srcDir string, version string, private bool, zones int, phase cloudup.Phase) {
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	h.MockKopsVersion("1.8.1")
	h.SetupMockAWS()
	phaseName := string(phase)
	if phaseName == "" {
		t.Fatalf("phase must be set")
	}
	tfFileName := phaseName + "-kubernetes.tf"

	expectedFilenames := []string{}

	if phase == cloudup.PhaseSecurity {
		expectedFilenames = []string{
			"aws_iam_role_masters." + clusterName + "_policy",
			"aws_iam_role_nodes." + clusterName + "_policy",
			"aws_iam_role_policy_masters." + clusterName + "_policy",
			"aws_iam_role_policy_nodes." + clusterName + "_policy",
			"aws_key_pair_kubernetes." + clusterName + "-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key",
		}
		if private {
			expectedFilenames = append(expectedFilenames, []string{
				"aws_iam_role_bastions." + clusterName + "_policy",
				"aws_iam_role_policy_bastions." + clusterName + "_policy",

				// bastions don't have any userdata
				// "aws_launch_configuration_bastions." + clusterName + "_user_data",
			}...)
		}
	} else if phase == cloudup.PhaseCluster {
		expectedFilenames = []string{
			"aws_launch_configuration_nodes." + clusterName + "_user_data",
		}

		for i := 0; i < zones; i++ {
			zone := "us-test-1" + string([]byte{byte('a') + byte(i)})
			s := "aws_launch_configuration_master-" + zone + ".masters." + clusterName + "_user_data"
			expectedFilenames = append(expectedFilenames, s)
		}
	}

	runTest(t, h, clusterName, srcDir, version, private, zones, expectedFilenames, tfFileName, &phase)
}

func runTestGCE(t *testing.T, clusterName string, srcDir string, version string, private bool, zones int) {
	featureflag.ParseFlags("+AlphaAllowGCE")

	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	h.MockKopsVersion("1.8.1")
	h.SetupMockGCE()

	expectedFilenames := []string{
		"google_compute_instance_template_nodes-" + gce.SafeClusterName(clusterName) + "_metadata_cluster-name",
		"google_compute_instance_template_nodes-" + gce.SafeClusterName(clusterName) + "_metadata_startup-script",
		"google_compute_instance_template_nodes-" + gce.SafeClusterName(clusterName) + "_metadata_ssh-keys",
	}

	for i := 0; i < zones; i++ {
		zone := "us-test1-" + string([]byte{byte('a') + byte(i)})
		prefix := "google_compute_instance_template_master-" + zone + "-" + gce.SafeClusterName(clusterName) + "_metadata_"

		expectedFilenames = append(expectedFilenames, prefix+"cluster-name")
		expectedFilenames = append(expectedFilenames, prefix+"startup-script")
		expectedFilenames = append(expectedFilenames, prefix+"ssh-keys")
	}

	runTest(t, h, clusterName, srcDir, version, private, zones, expectedFilenames, "", nil)
}

func runTestCloudformation(t *testing.T, clusterName string, srcDir string, version string, private bool) {
	srcDir = updateClusterTestBase + srcDir
	var stdout bytes.Buffer

	inputYAML := "in-" + version + ".yaml"
	expectedCfPath := "cloudformation.json"

	factoryOptions := &util.FactoryOptions{}
	factoryOptions.RegistryPath = "memfs://tests"

	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	h.MockKopsVersion("1.8.1")
	h.SetupMockAWS()

	factory := util.NewFactory(factoryOptions)

	{
		options := &CreateOptions{}
		options.Filenames = []string{path.Join(srcDir, inputYAML)}

		err := RunCreate(factory, &stdout, options)
		if err != nil {
			t.Fatalf("error running %q create: %v", inputYAML, err)
		}
	}

	{
		options := &CreateSecretPublickeyOptions{}
		options.ClusterName = clusterName
		options.Name = "admin"
		options.PublicKeyPath = path.Join(srcDir, "id_rsa.pub")

		err := RunCreateSecretPublicKey(factory, &stdout, options)
		if err != nil {
			t.Fatalf("error running %q create: %v", inputYAML, err)
		}
	}

	{
		options := &UpdateClusterOptions{}
		options.InitDefaults()
		options.Target = "cloudformation"
		options.OutDir = path.Join(h.TempDir, "out")
		options.MaxTaskDuration = 30 * time.Second

		// We don't test it here, and it adds a dependency on kubectl
		options.CreateKubecfg = false

		_, err := RunUpdateCluster(factory, clusterName, &stdout, options)
		if err != nil {
			t.Fatalf("error running update cluster %q: %v", clusterName, err)
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
		expectedCF, err := ioutil.ReadFile(path.Join(srcDir, expectedCfPath))
		if err != nil {
			t.Fatalf("unexpected error reading expected cloudformation output: %v", err)
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

		expectedCFTrimmed := strings.Replace(strings.TrimSpace(string(expectedCF)), "\r\n", "\n", -1)
		actualCFTrimmed := strings.TrimSpace(string(actualCF))
		if actualCFTrimmed != expectedCFTrimmed {
			diffString := diff.FormatDiff(expectedCFTrimmed, actualCFTrimmed)
			t.Logf("diff:\n%s\n", diffString)

			if os.Getenv("KEEP_TEMP_DIR") == "" {
				t.Logf("(hint: setting KEEP_TEMP_DIR will preserve test output")
			} else {
				t.Logf("actual terraform output in %s", actualPath)
			}

			t.Fatalf("cloudformation output differed from expected. Test file: %s", path.Join(srcDir, expectedCfPath))
		}

		expectedExtracted, err := ioutil.ReadFile(path.Join(srcDir, expectedCfPath+".extracted.yaml"))
		if err != nil {
			t.Fatalf("unexpected error reading expected extracted cloudformation output: %v", err)
		}

		expected := make(map[string]string)
		err = yaml.Unmarshal(expectedExtracted, &expected)
		if err != nil {
			t.Fatalf("unexpected error unmarshal expected extracted cloudformation output: %v", err)
		}

		if len(extracted) != len(expected) {
			t.Fatalf("error differed number of cloudformation in expected and extracted: %v", err)
		}

		for key, expectedValue := range expected {
			extractedValue, ok := extracted[key]
			if !ok {
				t.Fatalf("unexpected error expected cloudformation not found for k: %v", key)
			}

			// Strip cariage return as expectedValue is stored in a yaml string literal
			// and golang will automaticaly strip CR from any string literal
			extractedValueTrimmed := strings.Replace(extractedValue, "\r", "", -1)
			if expectedValue != extractedValueTrimmed {

				diffString := diff.FormatDiff(expectedValue, extractedValueTrimmed)
				t.Logf("diff for key %s:\n%s\n\n\n\n\n\n", key, diffString)
				t.Fatalf("cloudformation output differed from expected. Test file: %s", path.Join(srcDir, expectedCfPath+".extracted.yaml"))
			}
		}
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
