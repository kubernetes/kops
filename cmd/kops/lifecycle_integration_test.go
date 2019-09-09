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
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"k8s.io/kops/cloudmock/aws/mockec2"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

type LifecycleTestOptions struct {
	t           *testing.T
	SrcDir      string
	Version     string
	ClusterName string

	// Shared is a list of resource ids we expect to be tagged as shared
	Shared []string
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
		Shared: []string{"nat-a2345678", "nat-b2345678"},
	})
}

// TestLifecycleSharedVPC runs the test on a shared VPC
func TestLifecycleSharedVPC(t *testing.T) {
	runLifecycleTestAWS(&LifecycleTestOptions{
		t:      t,
		SrcDir: "shared_vpc",
	})
}

// TestLifecycleSharedSubnet runs the test on a shared subnet
func TestLifecycleSharedSubnet(t *testing.T) {
	runLifecycleTestAWS(&LifecycleTestOptions{
		t:      t,
		SrcDir: "shared_subnet",
		Shared: []string{"subnet-12345678"},
	})
}

// TestLifecyclePrivateSharedSubnet runs the test on a shared subnet with private topology
func TestLifecyclePrivateSharedSubnet(t *testing.T) {
	runLifecycleTestAWS(&LifecycleTestOptions{
		t:      t,
		SrcDir: "private-shared-subnet",
		Shared: []string{"subnet-12345678", "subnet-abcdef"},
	})
}

func runLifecycleTest(h *testutils.IntegrationTestHarness, o *LifecycleTestOptions, cloud *awsup.MockAWSCloud) {
	t := o.t

	t.Logf("running lifecycle test for cluster %s", o.ClusterName)

	var stdout bytes.Buffer

	inputYAML := "in-" + o.Version + ".yaml"

	factoryOptions := &util.FactoryOptions{}
	factoryOptions.RegistryPath = "memfs://tests"

	factory := util.NewFactory(factoryOptions)

	beforeResources := AllResources(cloud)

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
		options.RunTasksOptions.MaxTaskDuration = 10 * time.Second
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
		options.RunTasksOptions.MaxTaskDuration = 10 * time.Second

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

	{
		var ids []string
		for id := range AllResources(cloud) {
			ids = append(ids, id)
		}
		sort.Strings(ids)

		for _, id := range ids {
			tags, err := cloud.GetTags(id)
			if err != nil {
				t.Fatalf("error getting tags for %q: %v", id, err)
			}

			dashIndex := strings.Index(id, "-")
			resource := ""
			if dashIndex == -1 {
				t.Errorf("unknown resource type: %q", id)
			} else {
				resource = id[:dashIndex]
			}

			legacy := tags["KubernetesCluster"]
			if legacy != "" && legacy != o.ClusterName {
				t.Errorf("unexpected legacy KubernetesCluster tag: actual=%q cluster=%q", legacy, o.ClusterName)
			}

			ownership := tags["kubernetes.io/cluster/"+o.ClusterName]
			if beforeResources[id] != nil {
				expect := ""
				for _, s := range o.Shared {
					if id == s {
						expect = "shared"
					}
				}
				if ownership != expect {
					t.Errorf("unexpected kubernetes.io/cluster/ tag on %q: actual=%q expected=%q", id, ownership, expect)
				}
				if legacy != "" {
					t.Errorf("added (legacy) KubernetesCluster tag on %q, but it is shared", id)
				}
			} else {
				switch resource {
				case "ami":
				case "sshkey":
					// ignore

				default:
					if ownership == "" {
						t.Errorf("no kubernetes.io/cluster/ tag on %q", id)
					}
					if legacy == "" {
						// We want to deprecate the KubernetesCluster tag, e.g. in IAM
						// but we should probably keep it around for people that may be using it for other purposes
						t.Errorf("no (legacy) KubernetesCluster tag on %q", id)
					}
				}
			}
		}
	}

	{
		options := &DeleteClusterOptions{}
		options.Yes = true
		options.ClusterName = o.ClusterName
		if err := RunDeleteCluster(factory, &stdout, options); err != nil {
			t.Fatalf("error running delete cluster %q: %v", o.ClusterName, err)
		}
	}
}

// AllResources returns all resources
func AllResources(c *awsup.MockAWSCloud) map[string]interface{} {
	all := make(map[string]interface{})
	for k, v := range c.MockEC2.(*mockec2.MockEC2).All() {
		all[k] = v
	}
	return all
}

func runLifecycleTestAWS(o *LifecycleTestOptions) {
	o.AddDefaults()

	t := o.t

	h := testutils.NewIntegrationTestHarness(o.t)
	defer h.Close()

	h.MockKopsVersion("1.8.1")
	cloud := h.SetupMockAWS()

	var beforeIds []string
	for id := range AllResources(cloud) {
		beforeIds = append(beforeIds, id)
	}
	sort.Strings(beforeIds)

	runLifecycleTest(h, o, cloud)

	var afterIds []string
	for id := range AllResources(cloud) {
		afterIds = append(afterIds, id)
	}
	sort.Strings(afterIds)

	if !reflect.DeepEqual(beforeIds, afterIds) {
		t.Fatalf("resources changed by cluster create / destroy: %v -> %v", beforeIds, afterIds)
	}
}
