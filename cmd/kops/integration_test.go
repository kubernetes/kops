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
	"encoding/pem"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/pkg/testutils"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

// TestMinimal runs the test on a minimum configuration, similar to kops create cluster minimal.example.com --zones us-west-1a
func TestMinimal(t *testing.T) {
	runTest(t, "minimal.example.com", "../../tests/integration/minimal", "v1alpha0", false, 1)
	runTest(t, "minimal.example.com", "../../tests/integration/minimal", "v1alpha1", false, 1)
	runTest(t, "minimal.example.com", "../../tests/integration/minimal", "v1alpha2", false, 1)
}

// TestHA runs the test on a simple HA configuration, similar to kops create cluster minimal.example.com --zones us-west-1a,us-west-1b,us-west-1c --master-count=3
func TestHA(t *testing.T) {
	runTest(t, "ha.example.com", "../../tests/integration/ha", "v1alpha1", false, 3)
	runTest(t, "ha.example.com", "../../tests/integration/ha", "v1alpha2", false, 3)
}

// TestComplex runs the test on a more complex configuration, intended to hit more of the edge cases
func TestComplex(t *testing.T) {
	runTest(t, "complex.example.com", "../../tests/integration/complex", "v1alpha2", false, 1)
}

// TestMinimalCloudformation runs the test on a minimum configuration, similar to kops create cluster minimal.example.com --zones us-west-1a
func TestMinimalCloudformation(t *testing.T) {
	//runTestCloudformation(t, "minimal.example.com", "../../tests/integration/minimal", "v1alpha0", false)
	//runTestCloudformation(t, "minimal.example.com", "../../tests/integration/minimal", "v1alpha1", false)
	runTestCloudformation(t, "minimal.example.com", "../../tests/integration/minimal", "v1alpha2", false)
}

// TestMinimal_141 runs the test on a configuration from 1.4.1 release
func TestMinimal_141(t *testing.T) {
	runTest(t, "minimal-141.example.com", "../../tests/integration/minimal-141", "v1alpha0", false, 1)
}

// TestPrivateWeave runs the test on a configuration with private topology, weave networking
func TestPrivateWeave(t *testing.T) {
	runTest(t, "privateweave.example.com", "../../tests/integration/privateweave", "v1alpha1", true, 1)
	runTest(t, "privateweave.example.com", "../../tests/integration/privateweave", "v1alpha2", true, 1)
}

// TestPrivateFlannel runs the test on a configuration with private topology, flannel networking
func TestPrivateFlannel(t *testing.T) {
	runTest(t, "privateflannel.example.com", "../../tests/integration/privateflannel", "v1alpha1", true, 1)
	runTest(t, "privateflannel.example.com", "../../tests/integration/privateflannel", "v1alpha2", true, 1)
}

// TestPrivateCalico runs the test on a configuration with private topology, calico networking
func TestPrivateCalico(t *testing.T) {
	runTest(t, "privatecalico.example.com", "../../tests/integration/privatecalico", "v1alpha1", true, 1)
	runTest(t, "privatecalico.example.com", "../../tests/integration/privatecalico", "v1alpha2", true, 1)
}

// TestPrivateCanal runs the test on a configuration with private topology, canal networking
func TestPrivateCanal(t *testing.T) {
	runTest(t, "privatecanal.example.com", "../../tests/integration/privatecanal", "v1alpha1", true, 1)
	runTest(t, "privatecanal.example.com", "../../tests/integration/privatecanal", "v1alpha2", true, 1)
}

// TestPrivateKopeio runs the test on a configuration with private topology, kopeio networking
func TestPrivateKopeio(t *testing.T) {
	runTest(t, "privatekopeio.example.com", "../../tests/integration/privatekopeio", "v1alpha2", true, 1)
}

// TestPrivateDns runs the test on a configuration with private topology, private dns
func TestPrivateDns1(t *testing.T) {
	runTest(t, "privatedns1.example.com", "../../tests/integration/privatedns1", "v1alpha2", true, 1)
}

// TestPrivateDns runs the test on a configuration with private topology, private dns, extant vpc
func TestPrivateDns2(t *testing.T) {
	runTest(t, "privatedns2.example.com", "../../tests/integration/privatedns2", "v1alpha2", true, 1)
}

func runTest(t *testing.T, clusterName string, srcDir string, version string, private bool, zones int) {
	var stdout bytes.Buffer

	inputYAML := "in-" + version + ".yaml"
	expectedTFPath := "kubernetes.tf"

	factoryOptions := &util.FactoryOptions{}
	factoryOptions.RegistryPath = "memfs://tests"

	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

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
		options.Target = "terraform"
		options.OutDir = path.Join(h.TempDir, "out")
		options.MaxTaskDuration = 30 * time.Second

		// We don't test it here, and it adds a dependency on kubectl
		options.CreateKubecfg = false

		err := RunUpdateCluster(factory, clusterName, &stdout, options)
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
		expectedFilenames := "data,kubernetes.tf"
		if actualFilenames != expectedFilenames {
			t.Fatalf("unexpected files.  actual=%q, expected=%q", actualFilenames, expectedFilenames)
		}

		actualTF, err := ioutil.ReadFile(path.Join(h.TempDir, "out", "kubernetes.tf"))
		if err != nil {
			t.Fatalf("unexpected error reading actual terraform output: %v", err)
		}
		expectedTF, err := ioutil.ReadFile(path.Join(srcDir, expectedTFPath))
		if err != nil {
			t.Fatalf("unexpected error reading expected terraform output: %v", err)
		}

		if !bytes.Equal(actualTF, expectedTF) {
			diffString := diff.FormatDiff(string(expectedTF), string(actualTF))
			t.Logf("diff:\n%s\n", diffString)

			t.Fatalf("terraform output differed from expected")
		}
	}

	// Compare data files
	{
		files, err := ioutil.ReadDir(path.Join(h.TempDir, "out", "data"))
		if err != nil {
			t.Fatalf("failed to read data dir: %v", err)
		}

		var actualFilenames []string
		for _, f := range files {
			actualFilenames = append(actualFilenames, f.Name())
		}

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

				// bastions don't have any userdata
				// "aws_launch_configuration_bastions." + clusterName + "_user_data",
			}...)
		}
		sort.Strings(expectedFilenames)
		if !reflect.DeepEqual(actualFilenames, expectedFilenames) {
			t.Fatalf("unexpected data files.  actual=%q, expected=%q", actualFilenames, expectedFilenames)
		}

		// TODO: any verification of data files?
	}
}

func runTestCloudformation(t *testing.T, clusterName string, srcDir string, version string, private bool) {
	var stdout bytes.Buffer

	inputYAML := "in-" + version + ".yaml"
	expectedCfPath := "cloudformation.json"

	factoryOptions := &util.FactoryOptions{}
	factoryOptions.RegistryPath = "memfs://tests"

	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

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

		err := RunUpdateCluster(factory, clusterName, &stdout, options)
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

		actualCF, err := ioutil.ReadFile(path.Join(h.TempDir, "out", "kubernetes.json"))
		if err != nil {
			t.Fatalf("unexpected error reading actual cloudformation output: %v", err)
		}
		expectedCF, err := ioutil.ReadFile(path.Join(srcDir, expectedCfPath))
		if err != nil {
			t.Fatalf("unexpected error reading expected cloudformation output: %v", err)
		}

		if !bytes.Equal(actualCF, expectedCF) {
			diffString := diff.FormatDiff(string(expectedCF), string(actualCF))
			t.Logf("diff:\n%s\n", diffString)

			t.Fatalf("cloudformation output differed from expected")
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
