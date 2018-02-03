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
	"flag"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/apiserver/pkg/util/logs"
	"k8s.io/kube-deploy/cluster-api/tools/upgrader/util"
)

type UpgradeOptions struct {
	KubernetesVersion string
	kubeConfig        string
}

var uo = &UpgradeOptions{}

var RootCmd = &cobra.Command{
	Use:   "upgrader",
	Short: "cluster upgrader",
	Long:  `A single tool to upgrade kubernetes cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		if uo.KubernetesVersion == "" {
			glog.Exit("Please provide new kubernetes version.")
		}
		if err := RunUpgrade(uo); err != nil {
			glog.Exit(err.Error())
		}
	},
}

func RunUpgrade(uo *UpgradeOptions) error {
	err := util.UpgradeCluster(uo.KubernetesVersion, uo.kubeConfig)
	if err != nil {
		glog.Errorf("Failed to upgrade cluster with error : %v", err)
	}
	return err
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		glog.Exit(err)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&uo.kubeConfig, "kubecofig", "k", "", "location for the kubernetes config file. If not provided, $HOME/.kube/config is used")
	RootCmd.PersistentFlags().StringVarP(&uo.KubernetesVersion, "version", "v", "", "target kubernets version")
	flag.CommandLine.Parse([]string{})
	logs.InitLogs()
}
