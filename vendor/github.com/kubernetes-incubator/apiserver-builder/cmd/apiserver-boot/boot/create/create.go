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

package create

import (
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Command group for bootstrapping new resources.",
	Long:  `Command group for bootstrapping new resources.`,
	Example: `# Create new resource "Bee" in the "insect" group with version "v1beta"
# Will automatically the group and version if they do not exist
apiserver-boot create group version resource --group insect --version v1beta1 --kind Bee

# Create a new group "insect"
apiserver-boot create group --group insect

# Create a new version "v1beta" of group "insect"
# Will automatically create group if it does not exist
apiserver-boot create group --group insect --version v1beta`,
	Run: RunCreate,
}

var copyright string

func AddCreate(cmd *cobra.Command) {
	cmd.AddCommand(createCmd)
	cmd.Flags().StringVar(&copyright, "copyright", "boilerplate.go.txt", "Location of copyright boilerplate file.")
	AddCreateGroup(createCmd)
	AddCreateResource(createCmd)
	AddCreateSubresource(createCmd)
	AddCreateVersion(createCmd)
}

func RunCreate(cmd *cobra.Command, args []string) {
	cmd.Help()
}
