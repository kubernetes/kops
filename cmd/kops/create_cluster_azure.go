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
	//"bytes"
	//"encoding/csv"
	//"fmt"
	"io"
	//"io/ioutil"
	//"strconv"
	//"strings"
	//
	//"github.com/golang/glog"
	"github.com/spf13/cobra"
	//"k8s.io/apimachinery/pkg/util/sets"
	//"k8s.io/kops"
	"k8s.io/kops/cmd/kops/util"
	//api "k8s.io/kops/pkg/apis/kops"
	//"k8s.io/kops/pkg/apis/kops/registry"
	//"k8s.io/kops/pkg/apis/kops/validation"
	//"k8s.io/kops/pkg/client/simple/vfsclientset"
	//"k8s.io/kops/pkg/featureflag"
	//"k8s.io/kops/upup/pkg/fi"
	//"k8s.io/kops/upup/pkg/fi/cloudup"
	//"k8s.io/kops/upup/pkg/fi/utils"
	//"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/util/i18n"
)

type CreateClusterAzureOptions struct {
	// Inheritance in Go
	CreateClusterOptions
}

func NewCmdCreateClusterAzure(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateClusterAzureOptions{}
	options.InitDefaults()

	cmd := &cobra.Command{
		Use:     "azure",
		Short:   i18n.T("Create a Kubernetes cluster in Azure"),
		Long:    create_cluster_long,
		Example: create_cluster_example,
		Run: func(cmd *cobra.Command, args []string) {
			err := rootCommand.ProcessArgs(args)
			if err != nil {
				exitWithError(err)
				return
			}
			options.ClusterName = rootCommand.clusterName
			err = RunCreateClusterAzure(f, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}
	return cmd
}

func RunCreateClusterAzure(f *util.Factory, out io.Writer, c *CreateClusterAzureOptions) error {
	return c.RunCreateCluster(f, out)
}