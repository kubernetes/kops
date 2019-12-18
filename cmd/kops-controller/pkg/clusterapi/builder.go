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

package clusterapi

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	kopsv1alpha2 "k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
)

type Builder struct {
	Clientset simple.Clientset
}

// ClusterAPIBuilder is implemented by the model task that knows how to map an InstanceGroup to k8s objects
type ClusterAPIBuilder interface {
	MapToClusterAPI(cluster *kops.Cluster, ig *kops.InstanceGroup) ([]*unstructured.Unstructured, error)
}

func (b *Builder) BuildMachineDeployment(clusterObj *kopsv1alpha2.Cluster, igObj *kopsv1alpha2.InstanceGroup) ([]*unstructured.Unstructured, error) {
	cloudup.AlphaAllowGCE.SetEnabled(true)

	cluster := &kops.Cluster{}
	{
		if err := kopscodecs.Scheme.Convert(clusterObj, cluster, nil); err != nil {
			return nil, fmt.Errorf("error converting cluster to internal form: %v", err)
		}
	}

	ig := &kops.InstanceGroup{}
	{
		if err := kopscodecs.Scheme.Convert(igObj, ig, nil); err != nil {
			return nil, fmt.Errorf("error converting InstanceGroup to internal form: %v", err)
		}
	}

	phase := cloudup.PhaseCluster

	var secretStore fi.SecretStore

	/*{
		configBase, err := registry.ConfigBase(cluster)
		if err != nil {
			return nil, err
		}
		basedir := configBase.Join("secrets")
		secretStore = secrets.NewVFSSecretStore(cluster, basedir)
	}*/

	var keyStore fi.CAStore
	/*{
		configBase, err := registry.ConfigBase(cluster)
		if err != nil {
			return nil, err
		}
		basedir := configBase.Join("pki")
		allowList := true
		keyStore = fi.NewVFSCAStore(cluster, basedir, allowList)
	}*/

	var sshCredentialStore fi.SSHCredentialStore
	{
		configBase, err := registry.ConfigBase(cluster)
		if err != nil {
			return nil, err
		}
		basedir := configBase.Join("pki")
		sshCredentialStore = fi.NewVFSSSHCredentialStore(cluster, basedir)
	}

	{
		/*	assetBuilder := assets.NewAssetBuilder(cluster, string(phase))

			fullCluster, err := cloudup.PopulateClusterSpec(cluster, keyStore, secretStore, assetBuilder)
			if err != nil {
				return nil, err
			}
			cluster = fullCluster
		*/

		// The instance group is populated in place; no need to hydrate
		/*
			fullGroup, err := cloudup.PopulateInstanceGroupSpec(fullCluster, g, c.channel)
			if err != nil {
				return nil, err
			}
			ig = fullGroup
		*/
	}

	applyCmd := &cloudup.ApplyClusterCmd{
		Cluster:        cluster,
		Clientset:      b.Clientset,
		InstanceGroups: []*kops.InstanceGroup{ig},
		Phase:          phase,
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return nil, err
	}

	l, err := applyCmd.BuildLoader(cloud, keyStore, secretStore, sshCredentialStore)
	if err != nil {
		return nil, err
	}

	var objects []*unstructured.Unstructured
	for _, b := range l.Builders {
		clusterAPIBuilder, ok := b.(ClusterAPIBuilder)
		if !ok {
			continue
		}

		objs, err := clusterAPIBuilder.MapToClusterAPI(cluster, ig)
		if err != nil {
			return nil, err
		}

		objects = append(objects, objs...)
	}

	for _, obj := range objects {
		owners := obj.GetOwnerReferences()
		if len(owners) != 0 {
			// We could implement this, but for now we guard against it
			return nil, fmt.Errorf("existing owner refs not yet supported")
		}

		blockOwnerDeletion := true
		controller := true
		owners = append(owners, metav1.OwnerReference{
			Name:               ig.Name,
			Kind:               "InstanceGroup",
			APIVersion:         kopsv1alpha2.SchemeGroupVersion.String(),
			UID:                ig.UID,
			BlockOwnerDeletion: &blockOwnerDeletion,
			Controller:         &controller,
		})

		obj.SetOwnerReferences(owners)
	}

	return objects, nil
}
