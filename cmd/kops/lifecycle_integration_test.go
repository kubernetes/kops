/*
Copyright 2018 The Kubernetes Authors.

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
	"path"
	"strings"
	"testing"
	"time"

	"k8s.io/kops/upup/pkg/fi"

	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi/cloudup"
)

type LifecycleTestOptions struct {
	t           *testing.T
	SrcDir      string
	Version     string
	ClusterName string
}

func (o *LifecycleTestOptions) AddDefaults() {
	if o.Version == "" {
		o.Version = "v1alpha2"
	}
	if o.ClusterName == "" {
		o.ClusterName = strings.Replace(o.SrcDir, "_", "", -1) + ".example.com"
	}

	o.SrcDir = "../../tests/integration/update_cluster/" + o.SrcDir
}

// TestLifecycleMinimal runs the test on a minimum configuration, similar to kops create cluster minimal.example.com --zones us-west-1a
func TestLifecycleMinimal(t *testing.T) {
	runLifecycleTestAWS(&LifecycleTestOptions{
		t:      t,
		SrcDir: "minimal",
	})
}

// TestLifecyclePrivateCalico runs the test on a private topology
func TestLifecyclePrivateCalico(t *testing.T) {
	runLifecycleTestAWS(&LifecycleTestOptions{
		t:      t,
		SrcDir: "privatecalico",
	})
}

// TestLifecyclePrivateKopeio runs the test on a private topology, with kopeio networking
func TestLifecyclePrivateKopeio(t *testing.T) {
	runLifecycleTestAWS(&LifecycleTestOptions{
		t:      t,
		SrcDir: "privatekopeio",
	})
}

// TestLifecycleSharedVPC runs the test on a shared VPC
func TestLifecycleSharedVPC(t *testing.T) {
	runLifecycleTestAWS(&LifecycleTestOptions{
		t:      t,
		SrcDir: "shared_vpc",
	})
}

func runLifecycleTest(h *testutils.IntegrationTestHarness, o *LifecycleTestOptions) {
	t := o.t

	var stdout bytes.Buffer

	inputYAML := "in-" + o.Version + ".yaml"

	factoryOptions := &util.FactoryOptions{}
	factoryOptions.RegistryPath = "memfs://tests"

	factory := util.NewFactory(factoryOptions)

	{
		options := &CreateOptions{}
		options.Filenames = []string{path.Join(o.SrcDir, inputYAML)}

		err := RunCreate(factory, &stdout, options)
		if err != nil {
			t.Fatalf("error running %q create: %v", inputYAML, err)
		}
	}

	{
		options := &CreateSecretPublickeyOptions{}
		options.ClusterName = o.ClusterName
		options.Name = "admin"
		options.PublicKeyPath = path.Join(o.SrcDir, "id_rsa.pub")

		err := RunCreateSecretPublicKey(factory, &stdout, options)
		if err != nil {
			t.Fatalf("error running %q create: %v", inputYAML, err)
		}
	}

	{
		options := &UpdateClusterOptions{}
		options.InitDefaults()
		options.MaxTaskDuration = 10 * time.Second
		options.Yes = true

		// We don't test it here, and it adds a dependency on kubectl
		options.CreateKubecfg = false

		_, err := RunUpdateCluster(factory, o.ClusterName, &stdout, options)
		if err != nil {
			t.Fatalf("error running update cluster %q: %v", o.ClusterName, err)
		}
	}

	{
		options := &UpdateClusterOptions{}
		options.InitDefaults()
		options.Target = cloudup.TargetDryRun
		options.MaxTaskDuration = 10 * time.Second

		// We don't test it here, and it adds a dependency on kubectl
		options.CreateKubecfg = false

		results, err := RunUpdateCluster(factory, o.ClusterName, &stdout, options)
		if err != nil {
			t.Fatalf("error running update cluster %q: %v", o.ClusterName, err)
		}

		target := results.Target.(*fi.DryRunTarget)
		if target.HasChanges() {
			var b bytes.Buffer
			if err := target.PrintReport(results.TaskMap, &b); err != nil {
				t.Fatalf("error building report: %v", err)
			}
			t.Fatalf("Target had changes after executing: %v", b.String())
		}
	}
}

func runLifecycleTestAWS(o *LifecycleTestOptions) {
	o.AddDefaults()

	h := testutils.NewIntegrationTestHarness(o.t)
	defer h.Close()

	h.MockKopsVersion("1.8.1")
	h.SetupMockAWS()

	runLifecycleTest(h, o)
}
