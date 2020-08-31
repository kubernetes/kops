/*
Copyright 2020 The Kubernetes Authors.

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

package commands

import (
	"io"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/commands/helpers"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	helpersLong = templates.LongDesc(i18n.T(`
	Commands intended for integration with other systems.`))

	helpersShort = i18n.T(`Commands for use with other systems.`)
)

// NewCmdHelpers builds the cobra command tree for the `helpers` subcommand
func NewCmdHelpers(f *util.Factory, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "helpers",
		Short: helpersShort,
		Long:  helpersLong,

		// We hide the command, as it is intended for internal usage
		Hidden: true,
	}

	cmd.AddCommand(helpers.NewCmdHelperKubectlAuth(f, out))

	return cmd
}
