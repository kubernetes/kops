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

package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"k8s.io/kops/cmd/kops/util"
	"os"
)

type ToolboxCompletionOptions struct {
	// No options yet
}

func NewCmdToolboxCompletion(f *util.Factory, out io.Writer) *cobra.Command {
	options := &ToolboxCompletionOptions{}

	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Helper function for shell autocompletion",
		Run: func(cmd *cobra.Command, args []string) {
			err := RunToolboxCompletion(f, cmd, args, os.Stdout, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}

func RunToolboxCompletion(f *util.Factory, cmd *cobra.Command, args []string, out io.Writer, options *ToolboxCompletionOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("No args were provided (this command is normally called from a bash completion script)")
	}

	lastCommand := args[0]
	switch lastCommand {
	case "kops_get_clusters":
		fmt.Fprintf(out, "kopeio.awsdata.com\n")
		fmt.Fprintf(out, "kopeio2.awsdata.com\n")
		return nil

	case "kops_get_instancegroups":
		fmt.Fprintf(out, "node\n")
		fmt.Fprintf(out, "masters\n")
		return nil

	default:
		return fmt.Errorf("unhandled completion %q", lastCommand)
	}
}

func buildBashCompletionFunction() string {
	return bash_completion_func
}

const (
	bash_completion_func = `__kops_toolbox_completion()
{
    local kops_output out
    if kops_output=$(kops toolbox completion ${last_command} 2>/dev/null); then
        # TODO: remove
        out=($(echo "${kops_output}" | awk '{print $1}'))
        COMPREPLY=( $( compgen -W "${out[*]}" -- "$cur" ) )
      return 0
    fi
    return 1
}

__custom_func() {
    case ${last_command} in
        kops_get_clusters | kops_get_instancegroups)
            __kops_toolbox_completion
            return
            ;;
        *)
            ;;
    esac
}
`
)
