/*
Copyright 2026 The Kubernetes Authors.

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

package dump

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"sigs.k8s.io/yaml"
)

// mockCloud reports a fixed set of groups, along with the optional provisioning
// error and managed instance capabilities the dumper looks for.
type mockCloud struct {
	awsup.MockAWSCloud

	groups           map[string]*cloudinstances.CloudInstanceGroup
	errors           map[string][]cloudinstances.GroupFailure
	managedInstances map[string][]cloudinstances.GroupInstanceStatus
}

var (
	_ cloudinstances.GroupFailureReporter        = (*mockCloud)(nil)
	_ cloudinstances.GroupInstanceStatusReporter = (*mockCloud)(nil)
)

func (c *mockCloud) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	return c.groups, nil
}

func (c *mockCloud) GetGroupFailures(_ context.Context, group *cloudinstances.CloudInstanceGroup) ([]cloudinstances.GroupFailure, error) {
	return c.errors[group.HumanName], nil
}

func (c *mockCloud) GetGroupInstanceStatuses(_ context.Context, group *cloudinstances.CloudInstanceGroup) ([]cloudinstances.GroupInstanceStatus, error) {
	return c.managedInstances[group.HumanName], nil
}

func TestDumpCloudInstanceGroups(t *testing.T) {
	ctx := t.Context()

	cluster := &kops.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "testcluster.k8s.local"}}
	controlPlaneIG := &kops.InstanceGroup{
		ObjectMeta: metav1.ObjectMeta{Name: "control-plane-1"},
		Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleControlPlane},
	}
	nodeIG := &kops.InstanceGroup{
		ObjectMeta: metav1.ObjectMeta{Name: "nodes-1"},
		Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleNode},
	}

	launched := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	cloud := &mockCloud{
		MockAWSCloud: *awsup.BuildMockAWSCloud("us-east-1", "abc"),
		groups: map[string]*cloudinstances.CloudInstanceGroup{
			// A group that launched nothing: the fingerprint this file exists for.
			"cp-1.testcluster.k8s.local": {
				HumanName:     "cp-1.testcluster.k8s.local",
				InstanceGroup: controlPlaneIG,
				MinSize:       1,
				TargetSize:    1,
				MaxSize:       1,
			},
			"nodes-1.testcluster.k8s.local": {
				HumanName:     "nodes-1.testcluster.k8s.local",
				InstanceGroup: nodeIG,
				MinSize:       1,
				TargetSize:    1,
				MaxSize:       1,
				Ready: []*cloudinstances.CloudInstance{
					{
						ID:                "i-00001",
						MachineType:       "t3.medium",
						Status:            cloudinstances.CloudInstanceStatusUpToDate,
						PrivateIP:         "10.0.1.5",
						CreationTimestamp: launched,
						Node:              &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-1a"}},
					},
				},
			},
		},
		errors: map[string][]cloudinstances.GroupFailure{
			"cp-1.testcluster.k8s.local": {
				{
					Code:      "InsufficientInstanceCapacity",
					Message:   "We currently do not have sufficient t4g.large capacity in the Availability Zone you requested",
					Count:     12,
					FirstSeen: launched,
					LastSeen:  launched.Add(10 * time.Minute),
				},
			},
		},
		managedInstances: map[string][]cloudinstances.GroupInstanceStatus{
			"cp-1.testcluster.k8s.local": {
				{
					Name:          "cp-1-abcd",
					CurrentAction: "CREATING",
					Failures: []cloudinstances.GroupFailure{
						{
							Code:    "ZONE_RESOURCE_POOL_EXHAUSTED_WITH_DETAILS",
							Message: "The zone does not have enough resources available",
							Count:   1,
						},
					},
				},
			},
		},
	}

	dir := t.TempDir()
	err := DumpCloudInstanceGroups(ctx, cloud, cluster, []*kops.InstanceGroup{controlPlaneIG, nodeIG}, nil, dir)
	require.NoError(t, err)

	b, err := os.ReadFile(path.Join(dir, CloudInstanceGroupsFileName))
	require.NoError(t, err)

	var dumps []CloudInstanceGroupDump
	require.NoError(t, yaml.Unmarshal(b, &dumps))
	require.Len(t, dumps, 2)

	cp := dumps[0]
	assert.Equal(t, "cp-1.testcluster.k8s.local", cp.Name)
	assert.Equal(t, "control-plane-1", cp.InstanceGroup)
	assert.Equal(t, 1, cp.TargetSize)
	assert.Empty(t, cp.Instances)
	require.Len(t, cp.Failures, 1)
	assert.Equal(t, "InsufficientInstanceCapacity", cp.Failures[0].Code)
	assert.Equal(t, 12, cp.Failures[0].Count)
	require.Len(t, cp.ManagedInstances, 1)
	assert.Equal(t, "CREATING", cp.ManagedInstances[0].CurrentAction)
	require.Len(t, cp.ManagedInstances[0].Failures, 1)
	assert.Equal(t, "ZONE_RESOURCE_POOL_EXHAUSTED_WITH_DETAILS", cp.ManagedInstances[0].Failures[0].Code)

	nodes := dumps[1]
	assert.Equal(t, "nodes-1.testcluster.k8s.local", nodes.Name)
	assert.Empty(t, nodes.Failures)
	require.Len(t, nodes.Instances, 1)
	assert.Equal(t, "i-00001", nodes.Instances[0].ID)
	assert.Equal(t, "node-1a", nodes.Instances[0].Node)
	assert.Equal(t, launched, nodes.Instances[0].CreationTimestamp)
}

// TestDumpCloudInstanceGroupsWithoutOptionalCapabilities covers a cloud that
// implements neither optional interface, e.g. Azure or OpenStack.
func TestDumpCloudInstanceGroupsWithoutOptionalCapabilities(t *testing.T) {
	ctx := t.Context()

	cluster := &kops.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "testcluster.k8s.local"}}
	ig := &kops.InstanceGroup{
		ObjectMeta: metav1.ObjectMeta{Name: "nodes-1"},
		Spec:       kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleNode},
	}

	cloud := &basicMockCloud{
		groups: map[string]*cloudinstances.CloudInstanceGroup{
			"nodes-1.testcluster.k8s.local": {
				HumanName:     "nodes-1.testcluster.k8s.local",
				InstanceGroup: ig,
				MinSize:       1,
				TargetSize:    1,
				MaxSize:       1,
			},
		},
	}

	var c any = cloud
	_, isErrReporter := c.(cloudinstances.GroupFailureReporter)
	require.False(t, isErrReporter, "fixture must not implement GroupFailureReporter")
	_, isStatusReporter := c.(cloudinstances.GroupInstanceStatusReporter)
	require.False(t, isStatusReporter, "fixture must not implement GroupInstanceStatusReporter")

	dir := t.TempDir()
	require.NoError(t, DumpCloudInstanceGroups(ctx, cloud, cluster, []*kops.InstanceGroup{ig}, nil, dir))

	b, err := os.ReadFile(path.Join(dir, CloudInstanceGroupsFileName))
	require.NoError(t, err)

	var dumps []CloudInstanceGroupDump
	require.NoError(t, yaml.Unmarshal(b, &dumps))
	require.Len(t, dumps, 1)
	assert.Empty(t, dumps[0].Failures)
	assert.Empty(t, dumps[0].ManagedInstances)
}

// basicMockCloud implements neither optional reporter interface, matching a
// provider such as Azure or OpenStack. It embeds the fi.Cloud interface rather
// than awsup.MockAWSCloud so that no reporter method is promoted onto it.
type basicMockCloud struct {
	fi.Cloud

	groups map[string]*cloudinstances.CloudInstanceGroup
}

func (c *basicMockCloud) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	return c.groups, nil
}
