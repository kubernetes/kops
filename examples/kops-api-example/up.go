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
	"context"
	"fmt"
	"io/ioutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple/vfsclientset"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/utils"
)

func up(ctx context.Context) error {
	clientset := vfsclientset.NewVFSClientset(registryBase)

	cluster := &api.Cluster{}
	cluster.ObjectMeta.Name = clusterName
	cluster.Spec = api.ClusterSpec{
		Channel:       "stable",
		CloudProvider: "aws",
		ConfigBase:    registryBase.Join(cluster.ObjectMeta.Name).Path(),
		Topology:      &api.TopologySpec{},
	}
	cluster.Spec.Topology.Masters = api.TopologyPublic
	cluster.Spec.Topology.Nodes = api.TopologyPublic

	for _, z := range nodeZones {
		cluster.Spec.Subnets = append(cluster.Spec.Subnets, api.ClusterSubnetSpec{
			Name: z,
			Zone: z,
			Type: api.SubnetTypePublic,
		})
	}

	for _, etcdClusterName := range cloudup.EtcdClusters {
		etcdCluster := &api.EtcdClusterSpec{
			Name: etcdClusterName,
		}
		for _, masterZone := range masterZones {
			etcdMember := &api.EtcdMemberSpec{
				Name:          masterZone,
				InstanceGroup: fi.String(masterZone),
			}
			etcdCluster.Members = append(etcdCluster.Members, etcdMember)
		}
		cluster.Spec.EtcdClusters = append(cluster.Spec.EtcdClusters, etcdCluster)
	}

	if err := cloudup.PerformAssignments(cluster); err != nil {
		return err
	}

	_, err := clientset.CreateCluster(ctx, cluster)
	if err != nil {
		return err
	}

	// Create master ig
	{
		ig := &api.InstanceGroup{}
		ig.ObjectMeta.Name = "master"
		ig.Spec = api.InstanceGroupSpec{
			Role:    api.InstanceGroupRoleMaster,
			Subnets: masterZones,
		}
		_, err := clientset.InstanceGroupsFor(cluster).Create(ctx, ig, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	// Create node ig
	{
		ig := &api.InstanceGroup{}
		ig.ObjectMeta.Name = "nodes"
		ig.Spec = api.InstanceGroupSpec{
			Role:    api.InstanceGroupRoleNode,
			Subnets: nodeZones,
		}

		_, err := clientset.InstanceGroupsFor(cluster).Create(ctx, ig, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	sshCredentialStore, err := clientset.SSHCredentialStore(cluster)
	if err != nil {
		return err
	}

	// Add a public key
	{
		f := utils.ExpandPath(sshPublicKey)
		pubKey, err := ioutil.ReadFile(f)
		if err != nil {
			return fmt.Errorf("error reading SSH key file %q: %v", f, err)
		}
		err = sshCredentialStore.AddSSHPublicKey(fi.SecretNameSSHPrimary, pubKey)
		if err != nil {
			return fmt.Errorf("error adding SSH public key: %v", err)
		}
	}

	return nil
}
