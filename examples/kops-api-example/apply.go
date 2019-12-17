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
	"k8s.io/kops/pkg/client/simple/vfsclientset"
	"k8s.io/kops/upup/pkg/fi/cloudup"
)

func apply() error {
	allowList := true
	clientset := vfsclientset.NewVFSClientset(registryBase, allowList)

	cluster, err := clientset.GetCluster(clusterName)
	if err != nil {
		return err
	}

	applyCmd := &cloudup.ApplyClusterCmd{
		Cluster:    cluster,
		Clientset:  clientset,
		TargetName: cloudup.TargetDirect,
	}
	err = applyCmd.Run()
	if err != nil {
		return err
	}

	return nil
}
