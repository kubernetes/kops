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
	"k8s.io/kube-deploy/cluster-api/tools/repair/util"
)

type RepairOptions struct {
	dryRun     bool
	kubeConfig string
}

var ro = &RepairOptions{}

var rootCmd = &cobra.Command{
	Use:   "repair",
	Short: "Repair node",
	Long:  `Repairs given node`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunRepair(ro); err != nil {
			glog.Exit(err)
		}
	},
}

func RunRepair(ro *RepairOptions) error {
	r, err := util.NewRepairer(ro.dryRun, ro.kubeConfig)
	if err != nil {
		return err
	}
	return r.RepairNode()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		glog.Exit(err)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&ro.dryRun, "dryrun", "", false, "dry run mode.")
	rootCmd.PersistentFlags().StringVarP(&ro.kubeConfig, "kubecofig", "k", "", "location for the kubernetes config file. If not provided, $HOME/.kube/config is used")
	flag.CommandLine.Parse([]string{})
	logs.InitLogs()
}
