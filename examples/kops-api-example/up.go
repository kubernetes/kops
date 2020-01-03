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
	"fmt"
	"io/ioutil"

	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple/vfsclientset"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/utils"
)

func up() error {
	allowList := true
	clientset := vfsclientset.NewVFSClientset(registryBase, allowList)

	cluster := &kopsapi.Cluster{}
	cluster.ObjectMeta.Name = clusterName
	cluster.Spec = kopsapi.ClusterSpec{
		Channel:       "stable",
		CloudProvider: "aws",
		ConfigBase:    registryBase.Join(cluster.ObjectMeta.Name).Path(),
		Topology:      &kopsapi.TopologySpec{},
	}
	cluster.Spec.Topology.Masters = kopsapi.TopologyPublic
	cluster.Spec.Topology.Nodes = kopsapi.TopologyPublic

	for _, z := range nodeZones {
		cluster.Spec.Subnets = append(cluster.Spec.Subnets, kopsapi.ClusterSubnetSpec{
			Name: z,
			Zone: z,
			Type: kopsapi.SubnetTypePublic,
		})
	}

	for _, etcdClusterName := range cloudup.EtcdClusters {
		etcdCluster := &kopsapi.EtcdClusterSpec{
			Name: etcdClusterName,
		}
		for _, masterZone := range masterZones {
			etcdMember := &kopsapi.EtcdMemberSpec{
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

	_, err := clientset.CreateCluster(cluster)
	if err != nil {
		return err
	}

	// Create master ig
	{
		ig := &kopsapi.InstanceGroup{}
		ig.ObjectMeta.Name = "master"
		ig.Spec = kopsapi.InstanceGroupSpec{
			Role:    kopsapi.InstanceGroupRoleMaster,
			Subnets: masterZones,
		}
		_, err := clientset.InstanceGroupsFor(cluster).Create(ig)
		if err != nil {
			return err
		}
	}

	// Create node ig
	{
		ig := &kopsapi.InstanceGroup{}
		ig.ObjectMeta.Name = "nodes"
		ig.Spec = kopsapi.InstanceGroupSpec{
			Role:    kopsapi.InstanceGroupRoleNode,
			Subnets: nodeZones,
		}

		_, err := clientset.InstanceGroupsFor(cluster).Create(ig)
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
