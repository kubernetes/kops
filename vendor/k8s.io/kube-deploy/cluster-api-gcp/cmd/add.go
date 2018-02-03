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
	"os"
)

type AddOptions struct {
	Machine string
}

var ao = &AddOptions{}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add nodes to cluster",
	Long:  `Add nodes to cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		if ao.Machine == "" {
			glog.Error("Please provide yaml file for machine definition.")
			cmd.Help()
			os.Exit(1)
		}
		if err := RunAdd(ao); err != nil {
			glog.Exit(err)
		}
	},
}

func RunAdd(ao *AddOptions) error {
	machines, err := parseMachinesYaml(ao.Machine)
	if err != nil {
		return err
	}

	d := deploy.NewDeployer(provider, kubeConfig)

	return d.AddNodes(machines)
}
func init() {
	addCmd.Flags().StringVarP(&ao.Machine, "machines", "m", "", "machine yaml file")

	RootCmd.AddCommand(addCmd)
}
