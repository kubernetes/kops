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
	"os"

	"github.com/spf13/cobra"
)

var kubeconfig string

var rootCmd = &cobra.Command{
	Use: "machineset",
	Long: `This is an example client-side implementation of MachineSets.

For this demo, a "set" of machines is defined by using a label selector. All
machines matching the label selector are considered to be part of the same set.
You can get the machines in a set or scale it up using subcommands.

Show all machines in a set:

  $ machineset get set=node

Scale the set:

  $ machineset scale set=node --replicas 10`,
}

func main() {
	rootCmd.PersistentFlags().StringVarP(&kubeconfig, "kubeconfig", "k",
		homeDirOrPanic()+"/.kube/config", "path to kubeconfig file")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
