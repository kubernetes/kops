/*
Copyright 2017 The Kubernetes Authors.

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

package cmd

import (
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kube-deploy/cluster-api-gcp/deploy"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete kubernetes cluster",
	Long:  `Delete a kubernetes cluster with one command`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunDelete(); err != nil {
			glog.Exit(err)
		}
	},
}

func RunDelete() error {
	d := deploy.NewDeployer(provider, kubeConfig)
	return d.DeleteCluster()
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}
