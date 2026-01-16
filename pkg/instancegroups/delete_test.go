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

package instancegroups

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/compute/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops"
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

	err = d.DeleteInstanceGroup(&ig)
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

	clusterName := "test.k8s.io"

	cloud := h.SetupMockGCE()

	ctx := context.Background()
	f := util.NewFactory(&util.FactoryOptions{
		RegistryPath: "memfs://tests",
	})

	cluster := testutils.BuildMinimalClusterGCE(clusterName, cloud.Project())

	clientset, err := f.KopsClient()

	_, err = clientset.CreateCluster(ctx, cluster)
	if err != nil {
		t.Fatalf("error creating cluster: %v", err)
	}

	i := &compute.InstanceTemplate{
		Name: "test-template",
		Properties: &compute.InstanceProperties{
			Metadata: &compute.Metadata{
				Kind: "compute#metadata",
				Items: []*compute.MetadataItems{
					{
						Key:   "cluster-name",
						Value: &clusterName,
					},
				},
			},
		},
	}

	op, err := cloud.Compute().InstanceTemplates().Insert(cloud.Project(), i)
	if err != nil {
		t.Fatalf("error creating InstanceTemplate: %v", err)
	}

	if err := cloud.WaitForOp(op); err != nil {
		t.Fatalf("error creating InstanceTemplate: %v", err)
	}

	ig := &kops.InstanceGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-ig",
			Labels: map[string]string{
				kops.LabelClusterName: clusterName,
			},
		},
		Spec: kops.InstanceGroupSpec{
			Role:        kops.InstanceGroupRoleNode,
			MinSize:     fi.PtrTo(int32(1)),
			MaxSize:     fi.PtrTo(int32(1)),
			MachineType: "e2-medium",
			Zones:       []string{"us-central1-a"},
		},
	}
	igm := &compute.InstanceGroupManager{
		Name:             "a-test-ig-test-k8s-io",
		Zone:             "us-central1-a",
		InstanceTemplate: i.SelfLink,
	}

	op, err = cloud.Compute().InstanceGroupManagers().Insert(cloud.Project(), "us-test1-a", igm)
	if err != nil {
		t.Fatalf("error creating InstanceGroupManager: %v", err)
	}

	if err := cloud.WaitForOp(op); err != nil {
		t.Fatalf("error creating InstanceGroupManager: %v", err)
	}

	_, err = clientset.InstanceGroupsFor(cluster).Create(ctx, ig, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error creating instance group: %v", err)
	}

	op, err = cloud.Compute().InstanceGroupManagers().RecreateInstances("testproject", "us-central1-a", "a-test-ig-test-k8s-io", "0")
	if err != nil {
		t.Fatalf("error recreating Instance: %v", err)
	}

	list, _ := clientset.InstanceGroupsFor(cluster).List(ctx, metav1.ListOptions{})
	fmt.Println(list)

	deleteIG := &DeleteInstanceGroup{
		Cluster:   cluster,
		Cloud:     cloud,
		Clientset: clientset,
	}

	err = deleteIG.DeleteInstanceGroup(ig)
	if err != nil {
		t.Fatalf("error deleting instance group: %v", err)
	}

	// Verify that the instance group is deleted from the clientset
	_, err = clientset.InstanceGroupsFor(cluster).Get(ctx, ig.Name, metav1.GetOptions{})
	if err == nil {
		t.Fatalf("instance group %q was not deleted from clientset", ig.Name)
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("unexpected error when getting deleted instance group: %v", err)
	}

	// // Verify that the cloud resources are deleted (MIG and InstanceTemplate)
	// // This requires inspecting the mock cloud's internal state.
	// allResources := cloud.AllResources()
	// for name, resource := range allResources {
	// 	switch v := resource.(type) {
	// 	case *compute.InstanceGroupManager:
	// 		if strings.Contains(v.Name, ig.Name) {
	// 			t.Fatalf("InstanceGroupManager %q was not deleted from cloud", v.Name)
	// 		}
	// 	case *compute.InstanceTemplate:
	// 		if strings.Contains(v.Name, ig.Name) {
	// 			t.Fatalf("InstanceTemplate %q was not deleted from cloud", v.Name)
	// 		}
	// 	}
	// }
}
