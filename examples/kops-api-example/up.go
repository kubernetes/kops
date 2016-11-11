/*
Copyright 2016 The Kubernetes Authors.

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

	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/client/simple/vfsclientset"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/utils"
)

func up() error {
	clientset := vfsclientset.NewVFSClientset(registryBase)

	cluster := &api.Cluster{}

	cluster.Name = clusterName
	cluster.Spec = api.ClusterSpec{
		Channel:       "stable",
		CloudProvider: "aws",
		ConfigBase:    registryBase.Join(cluster.Name).Path(),
	}

	for _, z := range nodeZones {
		cluster.Spec.Zones = append(cluster.Spec.Zones, &api.ClusterZoneSpec{
			Name: z,
		})
	}

	for _, etcdClusterName := range cloudup.EtcdClusters {
		etcdCluster := &api.EtcdClusterSpec{
			Name: etcdClusterName,
		}
		for _, masterZone := range masterZones {
			etcdMember := &api.EtcdMemberSpec{
				Name: masterZone,
				Zone: fi.String(masterZone),
			}
			etcdCluster.Members = append(etcdCluster.Members, etcdMember)
		}
		cluster.Spec.EtcdClusters = append(cluster.Spec.EtcdClusters, etcdCluster)
	}

	if err := cluster.PerformAssignments(); err != nil {
		return err
	}

	_, err := clientset.Clusters().Create(cluster)
	if err != nil {
		return err
	}

	// Create master ig
	{
		ig := &api.InstanceGroup{}
		ig.Name = "master"
		ig.Spec = api.InstanceGroupSpec{
			Role:  api.InstanceGroupRoleMaster,
			Zones: masterZones,
		}
		_, err := clientset.InstanceGroups(cluster.Name).Create(ig)
		if err != nil {
			return err
		}
	}

	// Create node ig
	{
		ig := &api.InstanceGroup{}
		ig.Name = "nodes"
		ig.Spec = api.InstanceGroupSpec{
			Role:  api.InstanceGroupRoleNode,
			Zones: nodeZones,
		}

		_, err := clientset.InstanceGroups(cluster.Name).Create(ig)
		if err != nil {
			return err
		}
	}

	keyStore, err := registry.KeyStore(cluster)
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
		err = keyStore.AddSSHPublicKey(fi.SecretNameSSHPrimary, pubKey)
		if err != nil {
			return fmt.Errorf("error addding SSH public key: %v", err)
		}
	}

	return nil
}
