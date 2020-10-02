/*
Copyright 2020 The Kubernetes Authors.

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
	"os"
	"testing"
	"time"

	"k8s.io/kops/upup/pkg/fi"

	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"

	"k8s.io/kops/util/pkg/vfs"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/client/simple/vfsclientset"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

func getTestSetupOS(t *testing.T) (*RollingUpdateCluster, *openstack.MockCloud) {
	vfs.Context.ResetMemfsContext(true)

	k8sClient := fake.NewSimpleClientset()

	mockcloud := testutils.SetupMockOpenstack()

	inCluster := testutils.BuildMinimalCluster("test.k8s.local")

	inCluster.Spec.CloudProvider = "openstack"
	inCluster.Name = "test.k8s.local"

	inCluster.Spec.Topology.Masters = kopsapi.TopologyPrivate
	inCluster.Spec.Topology.Nodes = kopsapi.TopologyPrivate

	err := cloudup.PerformAssignments(inCluster, mockcloud)
	if err != nil {
		t.Fatalf("Failed to perform assignments: %v", err)
	}

	assetBuilder := assets.NewAssetBuilder(inCluster, "")
	basePath, _ := vfs.Context.BuildVfsPath(inCluster.Spec.ConfigBase)
	clientset := vfsclientset.NewVFSClientset(basePath)
	cluster, err := cloudup.PopulateClusterSpec(clientset, inCluster, mockcloud, assetBuilder)

	if err != nil {
		t.Fatalf("Failed to populate cluster spec: %v", err)
	}

	sshPublicKey := []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDF2sghZsClUBXJB4mBMIw8rb0hJWjg1Vz4eUeXwYmTdi92Gf1zNc5xISSip9Y+PWX/jJokPB7tgPnMD/2JOAKhG1bi4ZqB15pYRmbbBekVpM4o4E0dx+czbqjiAm6wlccTrINK5LYenbucAAQt19eH+D0gJwzYUK9SYz1hWnlGS+qurt2bz7rrsG73lN8E2eiNvGtIXqv3GabW/Hea3acOBgCUJQWUDTRu0OmmwxzKbFN/UpNKeRaHlCqwZWjVAsmqA8TX8LIocq7Np7MmIBwt7EpEeZJxThcmC8DEJs9ClAjD+jlLIvMPXKC3JWCPgwCLGxHjy7ckSGFCSzbyPduh")
	sshCredentialStore, err := clientset.SSHCredentialStore(cluster)
	if err != nil {
		t.Fatalf("Failed to get credential store: %v", err)
	}

	sshCredentialStore.AddSSHPublicKey(fi.SecretNameSSHPrimary, sshPublicKey)

	c := &RollingUpdateCluster{
		Cloud:                   mockcloud,
		MasterInterval:          1 * time.Millisecond,
		NodeInterval:            1 * time.Millisecond,
		BastionInterval:         1 * time.Millisecond,
		Force:                   false,
		K8sClient:               k8sClient,
		ClusterValidator:        &successfulClusterValidator{},
		FailOnValidate:          true,
		ValidateTickDuration:    1 * time.Millisecond,
		ValidateSuccessDuration: 5 * time.Millisecond,
		ValidateCount:           2,
		Ctx:                     context.Background(),
		Cluster:                 cluster,
		Clientset:               clientset,
	}

	return c, mockcloud
}

func TestRollingUpdateDisabledSurgeOS(t *testing.T) {
	origRegion := os.Getenv("OS_REGION_NAME")
	os.Setenv("OS_REGION_NAME", "us-test1")
	defer func() {
		os.Setenv("OS_REGION_NAME", origRegion)
	}()

	c, cloud := getTestSetupOS(t)

	groups, igList := getGroupsAllNeedUpdateOS(t, c)
	err := c.RollingUpdate(groups, igList)
	assert.NoError(t, err, "rolling update")

	assertGroupInstanceCountOS(t, cloud, "node-1", 3)
	assertGroupInstanceCountOS(t, cloud, "node-2", 3)
	assertGroupInstanceCountOS(t, cloud, "master-1", 2)
	assertGroupInstanceCountOS(t, cloud, "bastion-1", 1)
}

func makeGroupOS(t *testing.T, groups map[string]*cloudinstances.CloudInstanceGroup, igList *kopsapi.InstanceGroupList,
	c *RollingUpdateCluster, subnet string, role kopsapi.InstanceGroupRole, count int, needUpdate int) {
	cloud := c.Cloud.(*openstack.MockCloud)
	igif := c.Clientset.InstanceGroupsFor(c.Cluster)
	fakeClient := c.K8sClient.(*fake.Clientset)

	var newIg kopsapi.InstanceGroup
	switch role {
	case kopsapi.InstanceGroupRoleNode:
		newIg = testutils.BuildMinimalNodeInstanceGroup("nodes-"+subnet, subnet)
	case kopsapi.InstanceGroupRoleMaster:
		newIg = testutils.BuildMinimalMasterInstanceGroup(subnet)
	case kopsapi.InstanceGroupRoleBastion:
		newIg = testutils.BuildMinimalBastionInstanceGroup("bastion-"+subnet, subnet)
	}

	newIg.Spec.MachineType = "n1-standard-2"
	newIg.Spec.Image = "Ubuntu-20.04"

	igList.Items = append(igList.Items, newIg)

	ig, err := igif.Create(c.Ctx, &newIg, v1meta.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create ig %v: %v", subnet, err)
	}

	groups[subnet] = &cloudinstances.CloudInstanceGroup{
		HumanName:     ig.ObjectMeta.Name,
		InstanceGroup: ig,
	}
	for i := 0; i < count; i++ {
		name := subnet + string(rune('a'+i))
		port, err := cloud.CreatePort(ports.CreateOpts{
			Name:      name,
			NetworkID: "test",
		})
		if err != nil {
			t.Fatalf("Failed to make port: %v", err)
		}
		server, err := cloud.CreateInstance(servers.CreateOpts{
			Name: name,
			Networks: []servers.Network{
				{
					Port: port.ID,
				},
			},
		})
		if err != nil {
			t.Fatalf("Failed to make group: %v", err)
		}
		id := server.ID
		var node *v1.Node
		if role != kopsapi.InstanceGroupRoleBastion {
			node = &v1.Node{
				ObjectMeta: v1meta.ObjectMeta{Name: id + ".local"},
			}
			_ = fakeClient.Tracker().Add(node)
		}
		member := cloudinstances.CloudInstance{
			ID:                 id,
			Node:               node,
			CloudInstanceGroup: groups[subnet],
		}
		if i < needUpdate {
			groups[subnet].NeedUpdate = append(groups[subnet].NeedUpdate, &member)
		} else {
			groups[subnet].Ready = append(groups[subnet].Ready, &member)
		}
	}
}

func getGroupsAllNeedUpdateOS(t *testing.T, c *RollingUpdateCluster) (map[string]*cloudinstances.CloudInstanceGroup, *kopsapi.InstanceGroupList) {
	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	igList := &kopsapi.InstanceGroupList{}
	makeGroupOS(t, groups, igList, c, c.Cluster.Spec.Subnets[0].Name, kopsapi.InstanceGroupRoleNode, 3, 3)
	makeGroupOS(t, groups, igList, c, c.Cluster.Spec.Subnets[1].Name, kopsapi.InstanceGroupRoleNode, 3, 3)
	makeGroupOS(t, groups, igList, c, c.Cluster.Spec.Subnets[0].Name, kopsapi.InstanceGroupRoleMaster, 1, 1)
	makeGroupOS(t, groups, igList, c, c.Cluster.Spec.Subnets[1].Name, kopsapi.InstanceGroupRoleMaster, 1, 1)
	makeGroupOS(t, groups, igList, c, c.Cluster.Spec.Subnets[2].Name, kopsapi.InstanceGroupRoleMaster, 1, 1)
	makeGroupOS(t, groups, igList, c, c.Cluster.Spec.Subnets[0].Name, kopsapi.InstanceGroupRoleBastion, 1, 1)
	return groups, igList
}

func assertGroupInstanceCountOS(t *testing.T, cloud *openstack.MockCloud, groupName string, expected int) {

	groups, _ := cloud.ListServerGroups()
	for _, g := range groups {
		if g.Name == groupName {
			assert.Lenf(t, g.Members, expected, "%s instances", groupName)
		}
	}
}
