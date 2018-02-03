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

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/kube-deploy/cluster-api/client"
)

func init() {
	rootCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get <label>",
	Short: "Print the machines in a current set",
	Long: `Get all machines matching the given label.

Labels are always key-value pairs in the form "key=value", but can be any
arbitrary labels on your machines. For example:

  $ machineset get set=node
  $ machineset get gpu=true
  $ machineset get zone=us-central1-c

If your machines don't have labels yet, you can add them like so:

  $ kubectl label machine <name> key=value`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		machines, err := machinesClient(kubeconfig)
		if err != nil {
			die("Error creating machines controller: %v\n", err)
		}

		if err := get(machines, args[0]); err != nil {
			die("Error getting machines: %v\n", err)
		}
	},
}

func get(machines client.MachinesInterface, labelSelector string) error {
	list, err := machines.List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return err
	}

	if len(list.Items) == 0 {
		return fmt.Errorf("could not find existing machines with label %q", labelSelector)
	}

	fmt.Printf("There are %d machines with the label %q:\n", len(list.Items), labelSelector)
	for _, machine := range list.Items {
		fmt.Println(machine.ObjectMeta.Name)
	}

	return nil
}
