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

package init_repo

import (
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Command group for bootstrapping new go projects.",
	Long:  `Command group for bootstrapping new go projects.`,
	Example: `
# Bootstrap a new repo
apiserver-boot init repo --domain example.com
`,
	Run: RunInit,
}

func AddInit(cmd *cobra.Command) {
	cmd.AddCommand(initCmd)
	AddInitRepo(initCmd)
	AddVendorInstallCmd(initCmd)
}

func RunInit(cmd *cobra.Command, args []string) {
	cmd.Help()
}
