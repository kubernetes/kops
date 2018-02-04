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
	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"io/ioutil"
	"k8s.io/apiserver/pkg/util/logs"
	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
	"k8s.io/kube-deploy/cluster-api-gcp/util"
)

var RootCmd = &cobra.Command{
	Use:   "cluster-api-gcp",
	Short: "cluster management",
	Long:  `Simple kubernetes cluster management`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
		cmd.Help()
	},
}

var (
	kubeConfig string
	provider   string
)

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		glog.Exit(err)
	}
}

func parseClusterYaml(file string) (*clusterv1.Cluster, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	cluster := &clusterv1.Cluster{}
	err = yaml.Unmarshal(bytes, cluster)
	if err != nil {
		return nil, err
	}

	return cluster, nil
}

func parseMachinesYaml(file string) ([]*clusterv1.Machine, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	machines := &clusterv1.MachineList{}
	err = yaml.Unmarshal(bytes, &machines)
	if err != nil {
		return nil, err
	}

	return util.MachineP(machines.Items), nil
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&provider, "provider", "p", "google", "cloud provider google/azure/aws")
	RootCmd.PersistentFlags().StringVarP(&kubeConfig, "kubecofig", "k", "", "location for the kubernetes config file. If not provided, $HOME/.kube/config is used")
	flag.CommandLine.Parse([]string{})
	logs.InitLogs()
}
