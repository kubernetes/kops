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

	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/kutil"
)

type ImportClusterCmd struct {
	Region string
}

var importCluster ImportClusterCmd

func init() {
	cmd := &cobra.Command{
		Use:     "cluster",
		Short:   importShort,
		Long:    importLong,
		Example: importExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := importCluster.Run()
			if err != nil {
				exitWithError(err)
			}
		},
	}

	importCmd.AddCommand(cmd)

	cmd.Flags().StringVar(&importCluster.Region, "region", "", "region")
}

func (c *ImportClusterCmd) Run() error {
	if c.Region == "" {
		return fmt.Errorf("--region is required")
	}
	clusterName := rootCommand.clusterName
	if clusterName == "" {
		return fmt.Errorf("--name is required")
	}

	tags := map[string]string{awsup.TagClusterName: clusterName}
	cloud, err := awsup.NewAWSCloud(c.Region, tags)
	if err != nil {
		return fmt.Errorf("error initializing AWS client: %v", err)
	}

	clientset, err := rootCommand.Clientset()
	if err != nil {
		return err
	}

	cluster, err := clientset.GetCluster(clusterName)
	if err != nil {
		return err
	}
	if cluster != nil {
		return fmt.Errorf("cluster already exists: %q", clusterName)
	}

	d := &kutil.ImportCluster{}
	d.ClusterName = clusterName
	d.Cloud = cloud
	d.Clientset = clientset

	err = d.ImportAWSCluster()
	if err != nil {
		return err
	}

	fmt.Printf("\nImported settings for cluster %q\n", clusterName)

	return nil
}
