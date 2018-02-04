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
	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
	"k8s.io/kube-deploy/cluster-api/client"
)

var replicas int

var scaleCmd = &cobra.Command{
	Use:   "scale <label>",
	Short: "Scale a set to a desired number of replicas",
	Long: `This is an example client-side implementation of MachineSets. You can specify a
label selector to define the set, and a number of replicas to scale to. If
there are fewer Machines that match the label selector than --replicas, it will
create more Machines by cloning entries in the set. If there are more Machines
than --replicas, it will randomly delete Machines down to the correct number.

  $ machineset scale set=master --replicas 3
  $ machineset scale set=node   --replicas 10
  $ machineset scale gpu=true   --replicas 50

At least one Machine must exist that matches the label selector.

Labels are always key-value pairs in the form "key=value", but can be any
arbitrary labels on your machines. If your machines don't have labels yet, you
can add them like so:

  $ kubectl label machine <name> key=value`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		machines, err := machinesClient(kubeconfig)
		if err != nil {
			die("Error creating machines controller: %v", err)
		}

		if err := scale(machines, args[0], replicas); err != nil {
			die("Problem during scaling: %v", err)
		}
	},
}

func init() {
	scaleCmd.Flags().IntVarP(&replicas, "replicas", "r", -1, "number of replicas to scale to")
	scaleCmd.MarkFlagRequired("replicas")

	rootCmd.AddCommand(scaleCmd)
}

func scale(machines client.MachinesInterface, labelSelector string, replicas int) error {
	list, err := machines.List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return err
	}

	if len(list.Items) == 0 {
		return fmt.Errorf("could not find existing machines with label %s", labelSelector)
	}

	if len(list.Items) == replicas {
		fmt.Printf("Already have %d machines matching %q. Nothing to do.\n", replicas, labelSelector)
		return nil
	}

	if len(list.Items) > replicas {
		numToDelete := len(list.Items) - replicas
		fmt.Printf("Scaling down matchines matching %q from %d to %d; deleting %d machines.\n",
			labelSelector, len(list.Items), replicas, numToDelete)
		fmt.Println("Machines deleted:")
		for _, machine := range list.Items[0:numToDelete] {
			if err := machines.Delete(machine.ObjectMeta.Name, nil); err != nil {
				return err
			} else {
				fmt.Printf("  %s\n", machine.ObjectMeta.Name)
			}
		}

		return nil
	}

	if len(list.Items) < replicas {
		numToCreate := replicas - len(list.Items)
		fmt.Printf("Scaling up machines matching %q from %d to %d; creating %d machines.\n",
			labelSelector, len(list.Items), replicas, numToCreate)
		fmt.Println("Machines created:")

		newMachine := clone(list.Items[0])

		for i := 0; i < numToCreate; i++ {
			created, err := machines.Create(newMachine)
			if err != nil {
				return err
			} else {
				fmt.Printf("  %s\n", created.ObjectMeta.Name)
			}
		}

		return nil
	}

	return nil
}

func clone(old clusterv1.Machine) *clusterv1.Machine {
	// Make sure we get the full Spec
	newMachine := old.DeepCopy()

	// but sanitize the metadata so we only use meaningful fields.
	// TODO: set GenerateName ourselves if the target object doesn't have one.
	newMachine.ObjectMeta = metav1.ObjectMeta{}
	newMachine.ObjectMeta.GenerateName = old.ObjectMeta.GenerateName
	newMachine.ObjectMeta.Labels = old.ObjectMeta.Labels
	// Do not copy annotations as they currently contain info like machine name/location etc.
	// newMachine.ObjectMeta.Annotations = old.ObjectMeta.Annotations
	newMachine.ObjectMeta.ClusterName = old.ObjectMeta.ClusterName

	// Completely wipe out the status as well
	newMachine.Status = clusterv1.MachineStatus{}
	return newMachine
}
