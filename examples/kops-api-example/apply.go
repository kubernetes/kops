package main

import (
	"k8s.io/kops/pkg/client/simple/vfsclientset"
	"k8s.io/kops/upup/pkg/fi/cloudup"
)

func apply() error {
	clientset := vfsclientset.NewVFSClientset(registryBase)

	cluster, err := clientset.Clusters().Get(clusterName)
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