/*
Copyright 2021 The Kubernetes Authors.

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

package commandutils

import (
	"context"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

// CompleteClusterName returns a Cobra completion function for cluster names.
func CompleteClusterName(f Factory, suppressIfArgs bool, suppressArgs bool) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if suppressIfArgs && len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		ConfigureKlogForCompletion()

		client, err := f.KopsClient()
		if err != nil {
			return CompletionError("getting clientset", err)
		}

		list, err := client.ListClusters(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return CompletionError("listing clusters", err)
		}

		var clusterNames []string
		alreadySelected := sets.NewString()
		if suppressArgs {
			alreadySelected = alreadySelected.Insert(args...)
		}
		for _, cluster := range list.Items {
			if !alreadySelected.Has(cluster.Name) {
				clusterNames = append(clusterNames, cluster.Name)
			}
		}

		return clusterNames, cobra.ShellCompDirectiveNoFileComp
	}
}
