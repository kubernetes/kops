/*
Copyright The Kubernetes Authors.

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
	"testing"

	"github.com/stretchr/testify/assert"
	compute "google.golang.org/api/compute/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/testutils"
)

func TestRunDeleteInstanceGroup(t *testing.T) {
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	clusterName := "test.k8s.io"

	cloud := h.SetupMockGCE()

	ctx := context.Background()
	f := util.NewFactory(&util.FactoryOptions{
		RegistryPath: "memfs://tests",
	})

	cluster := testutils.BuildMinimalClusterGCE(clusterName, cloud.Project())

	clientset, err := f.KopsClient()
	assert.NoError(t, err, "error getting clientset")

	_, err = clientset.CreateCluster(ctx, cluster)
	assert.NoError(t, err, "error creating cluster")

	template := &compute.InstanceTemplate{
		Name: "test-template",
		Properties: &compute.InstanceProperties{
			Metadata: &compute.Metadata{
				Items: []*compute.MetadataItems{
					{
						Key:   "cluster-name",
						Value: &clusterName,
					},
				},
			},
		},
	}

	_, err = cloud.Compute().InstanceTemplates().Insert(cloud.Project(), template)
	assert.NoError(t, err, "error creating InstanceTemplate")

	ig := testutils.BuildMinimalNodeInstanceGroup("test-ig")

	igm := &compute.InstanceGroupManager{
		Name:             "a-test-ig-test-k8s-io",
		Zone:             "us-test1-a",
		InstanceTemplate: template.SelfLink,
	}

	_, err = cloud.Compute().InstanceGroupManagers().Insert(cloud.Project(), igm.Zone, igm)
	assert.NoError(t, err, "error creating InstanceGroupManager")

	_, err = clientset.InstanceGroupsFor(cluster).Create(ctx, &ig, metav1.CreateOptions{})
	assert.NoError(t, err, "error creating instance group")

	var stdout bytes.Buffer
	options := &DeleteInstanceGroupOptions{
		ClusterName: clusterName,
		GroupName:   ig.Name,
		Yes:         true,
		Force:       false,
	}

	assert.NoError(t, RunDeleteInstanceGroup(ctx, f, &stdout, options))

	// Verify that the instance group was deleted
	_, err = clientset.InstanceGroupsFor(cluster).Get(ctx, ig.Name, metav1.GetOptions{})
	assert.Error(t, err)
	assert.True(t, errors.IsNotFound(err), "unexpected error when getting deleted instance group")
}

func TestRunDeleteInstanceGroup_MissingInstance(t *testing.T) {
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	clusterName := "test.k8s.io"

	cloud := h.SetupMockGCE()

	ctx := context.Background()
	f := util.NewFactory(&util.FactoryOptions{
		RegistryPath: "memfs://tests",
	})

	cluster := testutils.BuildMinimalClusterGCE(clusterName, cloud.Project())

	clientset, err := f.KopsClient()
	assert.NoError(t, err, "error getting clientset")

	_, err = clientset.CreateCluster(ctx, cluster)
	assert.NoError(t, err, "error creating cluster")

	template := &compute.InstanceTemplate{
		Name: "test-template",
		Properties: &compute.InstanceProperties{
			Metadata: &compute.Metadata{
				Items: []*compute.MetadataItems{
					{
						Key:   "cluster-name",
						Value: &clusterName,
					},
				},
			},
		},
	}

	_, err = cloud.Compute().InstanceTemplates().Insert(cloud.Project(), template)
	assert.NoError(t, err, "error creating InstanceTemplate")

	ig := testutils.BuildMinimalNodeInstanceGroup("test-ig")

	igm := &compute.InstanceGroupManager{
		Name:             "a-test-ig-test-k8s-io",
		Zone:             "us-test1-a",
		InstanceTemplate: template.SelfLink,
	}

	_, err = cloud.Compute().InstanceGroupManagers().Insert(cloud.Project(), igm.Zone, igm)
	assert.NoError(t, err, "error creating InstanceGroupManager")

	_, err = clientset.InstanceGroupsFor(cluster).Create(ctx, &ig, metav1.CreateOptions{})
	assert.NoError(t, err, "error creating instance group")

	var stdout bytes.Buffer
	options := &DeleteInstanceGroupOptions{
		ClusterName: clusterName,
		GroupName:   ig.Name,
		Yes:         true,
		Force:       false,
	}

	// Simulate missing Instance
	_, err = cloud.Compute().Instances().Delete(cloud.Project(), igm.Zone, igm.Name)
	assert.NoError(t, err, "error deleting Instance")

	assert.ErrorContains(t, RunDeleteInstanceGroup(ctx, f, &stdout, options), "error getting Instance")

	// Verify that the instance group was not deleted
	_, err = clientset.InstanceGroupsFor(cluster).Get(ctx, ig.Name, metav1.GetOptions{})
	assert.NoError(t, err)

	options.Force = true
	assert.NoError(t, RunDeleteInstanceGroup(ctx, f, &stdout, options))

	// Verify that the instance group is deleted
	_, err = clientset.InstanceGroupsFor(cluster).Get(ctx, ig.Name, metav1.GetOptions{})
	assert.Error(t, err)
	assert.True(t, errors.IsNotFound(err), "unexpected error when getting deleted instance group")
}
