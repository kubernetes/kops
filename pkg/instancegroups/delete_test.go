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

package instancegroups

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	compute "google.golang.org/api/compute/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

func TestDeleteInstanceGroup_GCEWaitOnInstanceDeletion(t *testing.T) {
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	gce.PollingInterval = 5 * time.Millisecond
	defer func() {
		gce.PollingInterval = 5 * time.Second
	}()

	clusterName := "test.k8s.io"
	cloud := h.SetupMockGCE()

	zones, err := cloud.Zones()
	require.NoError(t, err)
	require.NotEmpty(t, zones)
	zone := zones[0]

	ctx := context.Background()
	f := util.NewFactory(&util.FactoryOptions{
		RegistryPath: "memfs://tests",
	})

	cluster := testutils.BuildMinimalClusterGCE(clusterName, cloud.Project())

	clientset, err := f.KopsClient()
	require.NoError(t, err)
	_, err = clientset.CreateCluster(ctx, cluster)
	require.NoError(t, err)

	ig := testutils.BuildMinimalNodeInstanceGroup("nodes-1", zone)

	_, err = clientset.InstanceGroupsFor(cluster).Create(context.TODO(), &ig, metav1.CreateOptions{})
	require.NoError(t, err)

	templateName := "test-template"
	templateURL := "https://www.googleapis.com/compute/v1/projects/testproject/global/instanceTemplates/" + templateName

	migName := gce.NameForInstanceGroupManager(clusterName, "nodes-1", zone)

	_, err = cloud.Compute().InstanceGroupManagers().Insert(cloud.Project(), zone, &compute.InstanceGroupManager{
		Name:             migName,
		InstanceTemplate: templateURL,
		Zone:             zone,
	})
	require.NoError(t, err)

	_, err = cloud.Compute().InstanceTemplates().Insert(cloud.Project(),
		&compute.InstanceTemplate{
			Name: templateName,
			Properties: &compute.InstanceProperties{
				Metadata: &compute.Metadata{
					Items: []*compute.MetadataItems{
						{
							Key:   "cluster-name",
							Value: fi.PtrTo(clusterName),
						},
					},
				},
			},
		})
	require.NoError(t, err)

	d := &DeleteInstanceGroup{
		Cluster:   cluster,
		Cloud:     cloud,
		Clientset: clientset,
	}

	err = d.DeleteInstanceGroup(ctx, &ig, false /*force*/)
	assert.NoError(t, err)

	// Check that all resources related to the CloudInstanceGroup were successfully deleted
	instances, err := cloud.Compute().Instances().List(ctx, cloud.Project(), zone)
	assert.NoError(t, err)
	assert.Len(t, instances, 0)

	_, err = cloud.Compute().InstanceGroupManagers().Get(cloud.Project(), zone, migName)
	assert.True(t, gce.IsNotFound(err))

	migs, err := cloud.Compute().InstanceGroupManagers().List(ctx, cloud.Project(), zone)
	assert.NoError(t, err)
	assert.Len(t, migs, 0)

	instanceTemplates, err := cloud.Compute().InstanceTemplates().List(ctx, cloud.Project())
	assert.NoError(t, err)
	assert.Len(t, instanceTemplates, 0)
}

func TestDeleteInstanceGroup(t *testing.T) {
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	gce.PollingInterval = 5 * time.Millisecond
	defer func() {
		gce.PollingInterval = 5 * time.Second
	}()

	clusterName := "test.k8s.io"
	cloud := h.SetupMockGCE()

	zones, err := cloud.Zones()
	require.NoError(t, err)
	require.NotEmpty(t, zones)
	zone := zones[0]

	ctx := context.Background()
	f := util.NewFactory(&util.FactoryOptions{
		RegistryPath: "memfs://tests",
	})

	cluster := testutils.BuildMinimalClusterGCE(clusterName, cloud.Project())

	clientset, err := f.KopsClient()
	assert.NoError(t, err, "error getting clientset")
	_, err = clientset.CreateCluster(ctx, cluster)
	assert.NoError(t, err, "error creating cluster")

	ig := testutils.BuildMinimalNodeInstanceGroup("nodes-1")

	templateName := "test-template"
	templateURL := "https://www.googleapis.com/compute/v1/projects/testproject/global/instanceTemplates/" + templateName

	migName := gce.NameForInstanceGroupManager(clusterName, "nodes-1", zone)

	template := &compute.InstanceTemplate{
		Name: templateName,
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

	igm := &compute.InstanceGroupManager{
		Name:             migName,
		InstanceTemplate: templateURL,
		Zone:             zone,
	}

	_, err = cloud.Compute().InstanceGroupManagers().Insert(cloud.Project(), zone, igm)
	assert.NoError(t, err, "error inserting InstanceGroupManager")

	_, err = clientset.InstanceGroupsFor(cluster).Create(ctx, &ig, metav1.CreateOptions{})
	assert.NoError(t, err, "error creating InstanceGroup")

	deleteIG := &DeleteInstanceGroup{
		Cluster:   cluster,
		Cloud:     cloud,
		Clientset: clientset,
	}

	assert.NoError(t, deleteIG.DeleteInstanceGroup(ctx, &ig, false /*force*/))

	// Verify that the instance group was deleted from the clientset
	_, err = clientset.InstanceGroupsFor(cluster).Get(ctx, ig.Name, metav1.GetOptions{})
	assert.Error(t, err)
	assert.True(t, errors.IsNotFound(err), "unexpected error when getting deleted instance group: %v", err)
}

func TestDeleteInstanceGroup_MissingInstance(t *testing.T) {
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
	assert.NoError(t, err, "error inserting InstanceGroupManager")

	_, err = clientset.InstanceGroupsFor(cluster).Create(ctx, &ig, metav1.CreateOptions{})
	assert.NoError(t, err, "error creating InstanceGroup")

	deleteIG := &DeleteInstanceGroup{
		Cluster:   cluster,
		Cloud:     cloud,
		Clientset: clientset,
	}

	// Simulate missing Instance
	_, err = cloud.Compute().Instances().Delete(cloud.Project(), igm.Zone, igm.Name)
	assert.NoError(t, err, "error deleting Instance")

	assert.ErrorContains(t, deleteIG.DeleteInstanceGroup(ctx, &ig, false /*force*/), "error getting Instance")
	// Verify that the instance group was not deleted
	_, err = clientset.InstanceGroupsFor(cluster).Get(ctx, ig.Name, metav1.GetOptions{})
	assert.NoError(t, err)

	assert.NoError(t, deleteIG.DeleteInstanceGroup(ctx, &ig, true /*force*/))
	// Verify that the instance group was deleted from the clientset
	_, err = clientset.InstanceGroupsFor(cluster).Get(ctx, ig.Name, metav1.GetOptions{})
	assert.Error(t, err)
	assert.True(t, errors.IsNotFound(err), "unexpected error when getting deleted instance group: %v", err)

	// Verify that the cloud resources for the InstanceGroup are also deleted
	groups, err := cloud.GetCloudGroups(cluster, []*api.InstanceGroup{&ig}, &fi.GetCloudGroupsOptions{}, nil)
	assert.NoError(t, err)
	assert.Len(t, groups, 0)
}
