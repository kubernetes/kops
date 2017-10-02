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

package run

import (
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Command group for launching instances.",
	Long:  `Command group for launching instances.`,
	Example: `# Run a local etcd, apiserver and controller-manager.
apiserver-boot run local

# Check the api versions of the locally running server
kubectl --kubeconfig kubeconfig api-versions

# Run a etcd, apiserver and controller-manager remotely in a Kubernetes cluster as an aggregated apiserver
apiserver-boot run in-cluster

# Clear the discovery service cache
rm -rf ~/.kube/cache/discovery/

# Check the api versions of the remotely running server
kubectl api-versions`,
	Run: RunRun,
}

func AddRun(cmd *cobra.Command) {
	cmd.AddCommand(runCmd)
	AddInCluster(runCmd)
	AddLocal(runCmd)
	AddLocalMinikube(runCmd)
}

func RunRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}
