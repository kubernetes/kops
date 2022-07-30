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
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"k8s.io/kops/cloudmock/aws/mockec2"
	gcemock "k8s.io/kops/cloudmock/gce"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/commands"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
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

// TestLifecycleMinimalAWS runs the test on a minimum configuration, similar to kops create cluster minimal.example.com --zones us-west-1a
func TestLifecycleMinimalAWS(t *testing.T) {
	runLifecycleTestAWS(&LifecycleTestOptions{
		t:      t,
		SrcDir: "minimal",
	})
}

func TestLifecycleMinimalOpenstack(t *testing.T) {
	runLifecycleTestOpenstack(&LifecycleTestOptions{
		t:           t,
		SrcDir:      "minimal_openstack",
		ClusterName: "minimal-openstack.k8s.local",
	})
}

func TestLifecycleMinimalGCE(t *testing.T) {
	runLifecycleTestGCE(&LifecycleTestOptions{
		t:           t,
		SrcDir:      "minimal_gce",
		ClusterName: "minimal-gce.example.com",
	})
}

func TestLifecycleFloatingIPOpenstack(t *testing.T) {
	runLifecycleTestOpenstack(&LifecycleTestOptions{
		t:           t,
		SrcDir:      "openstack_floatingip",
		ClusterName: "floatingip-openstack.k8s.local",
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

// TestLifecycleIPv6 runs the test on a IPv6 topology
func TestLifecycleIPv6(t *testing.T) {
	runLifecycleTestAWS(&LifecycleTestOptions{
		t:      t,
		SrcDir: "minimal-ipv6",
	})
}

// TestLifecycleSharedVPC runs the test on a shared VPC
func TestLifecycleSharedVPC(t *testing.T) {
	runLifecycleTestAWS(&LifecycleTestOptions{
		t:      t,
		SrcDir: "shared_vpc",
	})
}

// TestLifecycleComplex runs the test on a complex cluster
func TestLifecycleComplex(t *testing.T) {
	runLifecycleTestAWS(&LifecycleTestOptions{
		t:      t,
		SrcDir: "complex",
	})
}

// TestLifecycleExternlLB runs the test on a cluster with external load balancers and target groups attached
func TestLifecycleExternalLB(t *testing.T) {
	runLifecycleTestAWS(&LifecycleTestOptions{
		t:      t,
		SrcDir: "externallb",
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

// TestLifecyclePrivateSharedIP runs the test on a subnet with private topology and shared IP
func TestLifecyclePrivateSharedIP(t *testing.T) {
	runLifecycleTestAWS(&LifecycleTestOptions{
		t:      t,
		SrcDir: "private-shared-ip",
		Shared: []string{"eipalloc-12345678"},
	})
}

// TestLifecycleNodeTerminationHandlerQueueProcessor runs the test on a cluster with requisite resources for NTH Queue Processor
func TestLifecycleNodeTerminationHandlerQueueProcessor(t *testing.T) {
	runLifecycleTestAWS(&LifecycleTestOptions{
		t:           t,
		SrcDir:      "nth_sqs_resources",
		ClusterName: "nthsqsresources.longclustername.example.com",
	})
}

func runLifecycleTest(h *testutils.IntegrationTestHarness, o *LifecycleTestOptions, cloud *awsup.MockAWSCloud) {
	ctx := context.Background()

	featureflag.ParseFlags("+SpecOverrideFlag")
	unsetFeatureFlags := func() {
		featureflag.ParseFlags("-SpecOverrideFlag")
	}
	defer unsetFeatureFlags()

	t := o.t

	t.Logf("running lifecycle test for cluster %s", o.ClusterName)

	var stdout bytes.Buffer

	inputYAML := "in-" + o.Version + ".yaml"

	beforeResources := AllAWSResources(cloud)

	factory := newIntegrationTest(o.ClusterName, o.SrcDir).
		setupCluster(t, inputYAML, ctx, stdout)

	updateEnsureNoChanges(ctx, t, factory, o.ClusterName, stdout)

	// Overrides
	{
		cluster, err := GetCluster(ctx, factory, o.ClusterName)
		if err != nil {
			t.Fatalf("error getting cluster: %v", err)
		}
		clientset, err := factory.KopsClient()
		if err != nil {
			t.Fatalf("error getting clientset: %v", err)
		}

		overrides, err := loadOverrides(path.Join(o.SrcDir, "cluster.overrides.txt"))
		if err != nil {
			t.Fatalf("error loading overrides file: %v", err)
		}
		for _, overrideBatch := range overrides {
			t.Logf("overriding cluster values %v\n", overrideBatch)
			instanceGroups, err := commands.ReadAllInstanceGroups(ctx, clientset, cluster)
			if err != nil {
				t.Fatalf("error reading instance groups: %v", err)
			}

			if err := commands.SetClusterFields(overrideBatch, cluster); err != nil {
				t.Fatalf("error setting cluster fields: %v", err)
			}

			if err := commands.UpdateCluster(ctx, clientset, cluster, instanceGroups); err != nil {
				t.Fatalf("error updating cluster: %v", err)
			}
			updateEnsureNoChanges(ctx, t, factory, o.ClusterName, stdout)
		}

		instanceGroups, err := commands.ReadAllInstanceGroups(ctx, clientset, cluster)
		if err != nil {
			t.Fatalf("error reading instance groups: %v", err)
		}
		for _, ig := range instanceGroups {
			overrideFile := path.Join(o.SrcDir, fmt.Sprintf("instancegroup.%v.overrides.txt", ig.Name))
			overrides, err := loadOverrides(overrideFile)
			if err != nil {
				t.Fatalf("error loading overrides file: %v", err)
			}

			for _, overrideBatch := range overrides {
				t.Logf("overriding instance group values (%v) %v\n", ig.Name, overrideBatch)
				instanceGroups, err := commands.ReadAllInstanceGroups(ctx, clientset, cluster)
				if err != nil {
					t.Fatalf("error reading instance groups: %v", err)
				}
				var instanceGroupToUpdate *kops.InstanceGroup
				for _, instanceGroup := range instanceGroups {
					if instanceGroup.GetName() == ig.Name {
						instanceGroupToUpdate = instanceGroup
					}
				}
				if instanceGroupToUpdate == nil {
					t.Fatalf("unable to find instance group with name %q", ig.Name)
				}

				err = commands.SetInstancegroupFields(overrideBatch, instanceGroupToUpdate)
				if err != nil {
					t.Fatalf("error applying overrides: %v", err)
				}

				err = commands.UpdateInstanceGroup(ctx, clientset, cluster, instanceGroups, instanceGroupToUpdate)
				if err != nil {
					t.Fatalf("error updating instance groups: %v", err)
				}
				updateEnsureNoChanges(ctx, t, factory, o.ClusterName, stdout)
			}
		}
	}

	{
		var ids []string
		for id := range AllAWSResources(cloud) {
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
				case "lt":
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
		if err := RunDeleteCluster(ctx, factory, &stdout, options); err != nil {
			t.Fatalf("error running delete cluster %q: %v", o.ClusterName, err)
		}
	}
}

// AllAWSResources returns all resources
func AllAWSResources(c *awsup.MockAWSCloud) map[string]interface{} {
	all := make(map[string]interface{})
	for k, v := range c.MockEC2.(*mockec2.MockEC2).All() {
		all[k] = v
	}
	return all
}

// AllOpenstackResources returns all resources
func AllOpenstackResources(c *openstack.MockCloud) map[string]interface{} {
	all := make(map[string]interface{})
	for k, v := range c.MockNovaClient.All() {
		all[k] = v
	}
	return all
}

// AllGCEResources returns all resources
func AllGCEResources(c *gcemock.MockGCECloud) map[string]interface{} {
	all := make(map[string]interface{})
	for k, v := range c.AllResources() {
		all[k] = v
	}
	return all
}

func runLifecycleTestAWS(o *LifecycleTestOptions) {
	o.AddDefaults()

	t := o.t

	h := testutils.NewIntegrationTestHarness(o.t)
	defer h.Close()

	h.MockKopsVersion("1.21.0-alpha.1")
	cloud := h.SetupMockAWS()

	var beforeIds []string
	for id := range AllAWSResources(cloud) {
		beforeIds = append(beforeIds, id)
	}
	sort.Strings(beforeIds)

	runLifecycleTest(h, o, cloud)

	var afterIds []string
	for id := range AllAWSResources(cloud) {
		afterIds = append(afterIds, id)
	}
	sort.Strings(afterIds)

	if !reflect.DeepEqual(beforeIds, afterIds) {
		t.Fatalf("resources changed by cluster create / destroy: %v -> %v", beforeIds, afterIds)
	}
}

func runLifecycleTestOpenstack(o *LifecycleTestOptions) {
	o.AddDefaults()

	t := o.t

	h := testutils.NewIntegrationTestHarness(o.t)
	defer h.Close()

	origRegion := os.Getenv("OS_REGION_NAME")
	os.Setenv("OS_REGION_NAME", "us-test1")
	defer func() {
		os.Setenv("OS_REGION_NAME", origRegion)
	}()

	h.MockKopsVersion("1.21.0-alpha.1")
	cloud := testutils.SetupMockOpenstack()

	var beforeIds []string
	for id := range AllOpenstackResources(cloud) {
		beforeIds = append(beforeIds, id)
	}
	sort.Strings(beforeIds)

	ctx := context.Background()

	t.Logf("running lifecycle test for cluster %s", o.ClusterName)

	var stdout bytes.Buffer

	inputYAML := "in-" + o.Version + ".yaml"

	factory := newIntegrationTest(o.ClusterName, o.SrcDir).
		setupCluster(t, inputYAML, ctx, stdout)

	updateEnsureNoChanges(ctx, t, factory, o.ClusterName, stdout)

	{
		options := &DeleteClusterOptions{}
		options.Yes = true
		options.ClusterName = o.ClusterName
		if err := RunDeleteCluster(ctx, factory, &stdout, options); err != nil {
			t.Fatalf("error running delete cluster %q: %v", o.ClusterName, err)
		}
	}

	{
		var afterIds []string
		for id := range AllOpenstackResources(cloud) {
			afterIds = append(afterIds, id)
		}
		sort.Strings(afterIds)

		if !reflect.DeepEqual(beforeIds, afterIds) {
			t.Fatalf("resources changed by cluster create / destroy: %v -> %v", beforeIds, afterIds)
		}
	}
}

func runLifecycleTestGCE(o *LifecycleTestOptions) {
	o.AddDefaults()

	t := o.t

	h := testutils.NewIntegrationTestHarness(o.t)
	defer h.Close()

	h.MockKopsVersion("1.21.0-alpha.1")

	cloud := h.SetupMockGCE()

	var beforeIds []string
	for id := range AllGCEResources(cloud) {
		beforeIds = append(beforeIds, id)
	}
	sort.Strings(beforeIds)

	ctx := context.Background()

	t.Logf("running lifecycle test for cluster %s", o.ClusterName)

	var stdout bytes.Buffer
	inputYAML := "in-" + o.Version + ".yaml"

	factory := newIntegrationTest(o.ClusterName, o.SrcDir).
		setupCluster(t, inputYAML, ctx, stdout)

	updateEnsureNoChanges(ctx, t, factory, o.ClusterName, stdout)

	{
		options := &DeleteClusterOptions{}
		options.Yes = true
		options.ClusterName = o.ClusterName
		if err := RunDeleteCluster(ctx, factory, &stdout, options); err != nil {
			t.Fatalf("error running delete cluster %q: %v", o.ClusterName, err)
		}
	}

	var afterIds []string
	for id := range AllGCEResources(cloud) {
		afterIds = append(afterIds, id)
	}
	sort.Strings(afterIds)

	if !reflect.DeepEqual(beforeIds, afterIds) {
		t.Fatalf("resources changed by cluster create / destroy: %v -> %v", beforeIds, afterIds)
	}
}

func updateEnsureNoChanges(ctx context.Context, t *testing.T, factory *util.Factory, clusterName string, stdout bytes.Buffer) {
	t.Helper()
	options := &UpdateClusterOptions{}
	options.InitDefaults()
	options.RunTasksOptions.MaxTaskDuration = 10 * time.Second
	options.Yes = true

	// We don't test it here, and it adds a dependency on kubectl
	options.CreateKubecfg = false
	options.ClusterName = clusterName

	_, err := RunUpdateCluster(ctx, factory, &stdout, options)
	if err != nil {
		t.Fatalf("error running update cluster %q: %v", clusterName, err)
	}

	// Now perform another dryrun update and ensure no changes are reported

	options = &UpdateClusterOptions{}
	options.InitDefaults()
	options.Target = cloudup.TargetDryRun
	options.RunTasksOptions.MaxTaskDuration = 10 * time.Second

	// We don't test it here, and it adds a dependency on kubectl
	options.CreateKubecfg = false
	options.ClusterName = clusterName

	results, err := RunUpdateCluster(ctx, factory, &stdout, options)
	if err != nil {
		t.Fatalf("error running update cluster %q: %v", clusterName, err)
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

// Returns a list of lists of overrides. each list of overrides will be applied in a batch
func loadOverrides(filepath string) ([][]string, error) {
	f, err := os.Open(filepath)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	overrides := make([][]string, 0)
	overrides = append(overrides, make([]string, 0))
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			overrides = append(overrides, make([]string, 0))
			continue
		}
		overrides[len(overrides)-1] = append(overrides[len(overrides)-1], line)
	}
	return overrides, nil
}
