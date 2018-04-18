/*
Copyright 2016 The Kubernetes Authors.

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
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/kubernetes-incubator/apiserver-builder/cmd/apiserver-boot/boot/build"
	"github.com/kubernetes-incubator/apiserver-builder/cmd/apiserver-boot/boot/util"
)

var runInClusterCmd = &cobra.Command{
	Use:   "in-cluster",
	Short: "run the etcd, apiserver and the controller-manager as an aggegrated apiserver in a cluster",
	Long:  `run the etcd, apiserver and the controller-manager as an aggegrated apiserver in a cluster`,
	Example: `
# Build a new image and run the apiserver and controller-manager in the cluster
apiserver-boot run in-cluster --name example --namespace default --image gcr.io/myrepo/myimage:mytag

# Clear the discovery cache for kubectl
rm -rf ~/.kube/cache/discovery/

# Run kubectl and check for the new version
kubectl api-versions

# Create an instance and fetch it
nano -w samples/<type>.yaml
kubectl apply -f samples/<type>.yaml
kubectl get <type>`,
	Run: RunInCluster,
}

var buildImage bool

func AddInCluster(cmd *cobra.Command) {
	cmd.AddCommand(runInClusterCmd)

	build.AddBuildResourceConfigFlags(runInClusterCmd)
	runInClusterCmd.Flags().BoolVar(&build.GenerateForBuild, "generate", true, "if true, generate code before building the container image")
	runInClusterCmd.Flags().BoolVar(&buildImage, "build-image", true, "if true, build the container image and push it to the image repo.")
}

func RunInCluster(cmd *cobra.Command, args []string) {
	if buildImage {
		// Build the container first
		build.RunBuildContainer(cmd, args)

		// Push the image
		util.DoCmd("docker", "push", build.Image)
	}

	// Build the resource config
	os.Remove(filepath.Join(build.ResourceConfigDir, "apiserver.yaml"))
	build.RunBuildResourceConfig(cmd, args)

	// Apply the new config
	util.DoCmd("kubectl", "apply", "-f", filepath.Join(build.ResourceConfigDir, "apiserver.yaml"))
}
