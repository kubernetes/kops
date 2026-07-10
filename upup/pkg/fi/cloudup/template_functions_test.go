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
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"text/template"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gcemock "k8s.io/kops/cloudmock/gce"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
)

func Test_TemplateFunctions_CloudControllerConfigArgv(t *testing.T) {
	tests := []struct {
		desc          string
		cluster       *kops.Cluster
		expectedArgv  []string
		expectedError error
	}{
		{
			desc: "Default Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{},
			}},
			expectedArgv: []string{
				"--v=2",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "Log Level Configuration",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					CloudProvider: kops.CloudProviderSpec{
						Openstack: &kops.OpenstackSpec{},
					},
					ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
						LogLevel: 3,
					},
				},
			},
			expectedArgv: []string{
				"--v=3",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "ExternalCloudControllerManager CloudProvider Configuration",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
						CloudProvider: string(kops.CloudProviderOpenstack),
						LogLevel:      3,
					},
				},
			},
			expectedArgv: []string{
				"--cloud-provider=openstack",
				"--v=3",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "No CloudProvider Configuration",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
						LogLevel: 3,
					},
				},
			},
			expectedError: fmt.Errorf("Cloud Provider is not set"),
		},
		{
			desc: "k8s cluster name",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					ClusterName: "k8s",
				},
			}},
			expectedArgv: []string{
				"--cluster-name=k8s",
				"--v=2",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "Default Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					Master: "127.0.0.1",
				},
			}},
			expectedArgv: []string{
				"--master=127.0.0.1",
				"--v=2",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "Cluster-cidr Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					ClusterCIDR: "10.0.0.0/24",
				},
			}},
			expectedArgv: []string{
				"--cluster-cidr=10.0.0.0/24",
				"--v=2",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "AllocateNodeCIDRs Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					AllocateNodeCIDRs: new(true),
				},
			}},
			expectedArgv: []string{
				"--allocate-node-cidrs=true",
				"--v=2",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "ConfigureCloudRoutes Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					ConfigureCloudRoutes: new(true),
				},
			}},
			expectedArgv: []string{
				"--configure-cloud-routes=true",
				"--v=2",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "CIDRAllocatorType Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					CIDRAllocatorType: new("RangeAllocator"),
				},
			}},
			expectedArgv: []string{
				"--cidr-allocator-type=RangeAllocator",
				"--v=2",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "CIDRAllocatorType Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					UseServiceAccountCredentials: new(false),
				},
			}},
			expectedArgv: []string{
				"--use-service-account-credentials=false",
				"--v=2",
				"--cloud-provider=openstack",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "Leader Election",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					LeaderElection: &kops.LeaderElectionConfiguration{LeaderElect: new(true)},
				},
			}},
			expectedArgv: []string{
				"--leader-elect=true",
				"--v=2",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "Disable Controller",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					Controllers: []string{"*", "-nodeipam"},
				},
			}},
			expectedArgv: []string{
				"--controllers=*,-nodeipam",
				"--v=2",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.desc, func(t *testing.T) {
			tf := &TemplateFunctions{}
			tf.Cluster = testCase.cluster

			actual, error := tf.CloudControllerConfigArgv()
			if !reflect.DeepEqual(error, testCase.expectedError) {
				t.Errorf("Error differs: %+v instead of %+v", error, testCase.expectedError)
			}
			if !reflect.DeepEqual(actual, testCase.expectedArgv) {
				t.Errorf("Argv differs: %+v instead of %+v", actual, testCase.expectedArgv)
			}
		})
	}
}

func TestKopsFeatureEnabled(t *testing.T) {
	tests := []struct {
		name          string
		featureFlags  string
		featureName   string
		expectedValue bool
		expectError   bool
	}{
		{
			name:         "Missing feature",
			featureFlags: "",
			featureName:  "NonExistingFeature",
			expectError:  true,
		},
		{
			name:          "Existing feature",
			featureFlags:  "+Scaleway",
			featureName:   "Scaleway",
			expectError:   false,
			expectedValue: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			featureflag.ParseFlags(tc.featureFlags)
			tf := &TemplateFunctions{}
			value, err := tf.kopsFeatureEnabled(tc.featureName)
			if err != nil && !tc.expectError {
				t.Errorf("unexpected error: %s", err)
			}
			if err == nil && tc.expectError {
				t.Errorf("expected error, got nil")
			}
			if value != tc.expectedValue {
				t.Errorf("expected value %t, got %t", tc.expectedValue, value)
			}
		})
	}
}

func TestHasHighlyAvailableControlPlane(t *testing.T) {
	tests := []struct {
		name              string
		allInstanceGroups []*kops.InstanceGroup
		instanceGroups    []*kops.InstanceGroup
		expectedHA        bool
	}{
		{
			name: "Single control plane node",
			allInstanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "master-us-east-1a"},
					Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleControlPlane},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "nodes"},
					Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleNode},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "nodes"},
					Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleNode},
				},
			},
			expectedHA: false,
		},
		{
			name: "Multiple control plane nodes",
			allInstanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "master-us-east-1a"},
					Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleControlPlane},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "master-us-east-1b"},
					Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleControlPlane},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "nodes"},
					Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleNode},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "nodes"},
					Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleNode},
				},
			},
			expectedHA: true,
		},
		{
			name: "Multiple control plane nodes with filtered instance groups (regression test)",
			allInstanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "master-us-east-1a"},
					Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleControlPlane},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "master-us-east-1b"},
					Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleControlPlane},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "master-us-east-1c"},
					Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleControlPlane},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "nodes"},
					Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleNode},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "nodes"},
					Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleNode},
				},
			},
			expectedHA: true,
		},
		{
			name: "Three control plane nodes",
			allInstanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "master-us-east-1a"},
					Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleControlPlane},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "master-us-east-1b"},
					Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleControlPlane},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "master-us-east-1c"},
					Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleControlPlane},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "master-us-east-1a"},
					Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleControlPlane},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "master-us-east-1b"},
					Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleControlPlane},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "master-us-east-1c"},
					Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleControlPlane},
				},
			},
			expectedHA: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tf := &TemplateFunctions{}
			tf.AllInstanceGroups = tc.allInstanceGroups
			tf.InstanceGroups = tc.instanceGroups

			actual := tf.HasHighlyAvailableControlPlane()
			if actual != tc.expectedHA {
				t.Errorf("expected HA to be %t, got %t", tc.expectedHA, actual)
			}
		})
	}
}

func TestTemplateFunctions_TaskHelpers(t *testing.T) {
	tf := &TemplateFunctions{}
	tf.Cluster = &kops.Cluster{}
	tf.tasks = map[string]fi.CloudupTask{
		"ManagedFile/zeta": &fitasks.ManagedFile{
			Name:     new("zeta"),
			Location: new("addons/zeta.yaml"),
		},
		"ManagedFile/alpha": &fitasks.ManagedFile{
			Name:     new("alpha"),
			Location: new("addons/alpha.yaml"),
		},
	}

	if !tf.HasTask("ManagedFile", "alpha") {
		t.Fatalf("expected alpha task to exist")
	}
	if tf.HasTask("ManagedFile", "missing") {
		t.Fatalf("did not expect missing task to exist")
	}

	task, err := tf.Task("ManagedFile", "alpha")
	if err != nil {
		t.Fatalf("Task returned error: %v", err)
	}
	if key, err := tf.TaskKey(task); err != nil || key != "ManagedFile/alpha" {
		t.Fatalf("unexpected task key %q, err=%v", key, err)
	}

	tasks, err := tf.TasksByType("ManagedFile")
	if err != nil {
		t.Fatalf("TasksByType returned error: %v", err)
	}
	var gotKeys []string
	for _, task := range tasks {
		key, err := tf.TaskKey(task)
		if err != nil {
			t.Fatalf("TaskKey returned error: %v", err)
		}
		gotKeys = append(gotKeys, key)
	}
	expectedKeys := []string{"ManagedFile/alpha", "ManagedFile/zeta"}
	if !reflect.DeepEqual(gotKeys, expectedKeys) {
		t.Fatalf("unexpected task order %v", gotKeys)
	}

	funcMap := template.FuncMap{}
	if err := tf.AddTo(funcMap, nil); err != nil {
		t.Fatalf("AddTo returned error: %v", err)
	}

	tmpl, err := template.New("tasks").Funcs(funcMap).Parse(`{{ TaskKey (Task "ManagedFile" "alpha") }}|{{ range $task := TasksByType "ManagedFile" }}{{ TaskKey $task }};{{ end }}`)
	if err != nil {
		t.Fatalf("error parsing template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, tf.Cluster.Spec); err != nil {
		t.Fatalf("error executing template: %v", err)
	}

	if actual := buf.String(); actual != "ManagedFile/alpha|ManagedFile/alpha;ManagedFile/zeta;" {
		t.Fatalf("unexpected rendered template %q", actual)
	}
}

func TestGetClusterAutoscalerNodeGroupsGCE(t *testing.T) {
	cloud := gcemock.InstallMockGCECloud("us-test1", "testproject")

	cluster := &kops.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "minimal.example.com"},
		Spec: kops.ClusterSpec{
			CloudProvider: kops.CloudProviderSpec{
				GCE: &kops.GCESpec{},
			},
		},
	}

	newIG := func(name string, minSize, maxSize int32, zones []string) *kops.InstanceGroup {
		return &kops.InstanceGroup{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Spec: kops.InstanceGroupSpec{
				Role:    kops.InstanceGroupRoleNode,
				MinSize: new(minSize),
				MaxSize: new(maxSize),
				Zones:   zones,
			},
		}
	}

	migURL := func(zone, name string) string {
		return fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/testproject/zones/%s/instanceGroups/%s", zone, name)
	}

	grid := []struct {
		desc     string
		ig       *kops.InstanceGroup
		expected map[string]ClusterAutoscalerNodeGroup
	}{
		{
			desc: "single zone",
			ig:   newIG("nodes", 1, 10, []string{"us-test1-a"}),
			expected: map[string]ClusterAutoscalerNodeGroup{
				"nodes": {MinSize: 1, MaxSize: 10, Other: migURL("us-test1-a", "a-nodes-minimal-example-com")},
			},
		},
		{
			desc: "multiple zones split min and max",
			ig:   newIG("nodes", 1, 3, []string{"us-test1-a", "us-test1-b", "us-test1-c"}),
			expected: map[string]ClusterAutoscalerNodeGroup{
				"nodes-us-test1-a": {MinSize: 1, MaxSize: 1, Other: migURL("us-test1-a", "a-nodes-minimal-example-com")},
				"nodes-us-test1-b": {MinSize: 0, MaxSize: 1, Other: migURL("us-test1-b", "b-nodes-minimal-example-com")},
				"nodes-us-test1-c": {MinSize: 0, MaxSize: 1, Other: migURL("us-test1-c", "c-nodes-minimal-example-com")},
			},
		},
		{
			desc: "max size smaller than zone count skips empty zones",
			ig:   newIG("nodes", 1, 1, []string{"us-test1-a", "us-test1-b", "us-test1-c"}),
			expected: map[string]ClusterAutoscalerNodeGroup{
				"nodes-us-test1-a": {MinSize: 1, MaxSize: 1, Other: migURL("us-test1-a", "a-nodes-minimal-example-com")},
			},
		},
	}

	for _, g := range grid {
		t.Run(g.desc, func(t *testing.T) {
			tf := &TemplateFunctions{cloud: cloud}
			tf.Cluster = cluster
			tf.InstanceGroups = []*kops.InstanceGroup{g.ig}

			actual, err := tf.GetClusterAutoscalerNodeGroups()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(actual, g.expected) {
				t.Errorf("expected %+v, got %+v", g.expected, actual)
			}
		})
	}
}
