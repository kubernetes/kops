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
	"io"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
)

const boilerPlate = `
# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
`

type CompletionOptions struct {
	Shell string
}

var (
	longDescription = `Output shell completion code for the given shell (bash).

This command prints shell code which must be evaluation to provide interactive
completion of kops commands.`

	example = `
# load in the kops completion code for bash (depends on the bash-completion framework).
source <(kops completion bash)`
)

func NewCmdCompletion(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CompletionOptions{}

	cmd := &cobra.Command{
		Use:     "completion",
		Short:   "Output shell completion code for the given shell (bash).",
		Long:    longDescription,
		Example: example,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunCompletion(f, cmd, args, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().StringVar(&options.Shell, "shell", "", "target shell (bash).")

	return cmd
}

func RunCompletion(f *util.Factory, cmd *cobra.Command, args []string, out io.Writer, c *CompletionOptions) error {
	if len(args) != 0 {
		if c.Shell != "" {
			return fmt.Errorf("cannot specify shell both as a flag and a positional argument")
		}
		c.Shell = args[0]
	}

	if c.Shell == "" {
		return fmt.Errorf("shell is required")
	}

	if c.Shell != "bash" {
		return fmt.Errorf("only bash shell is supported for kops completion")
	}

	_, err := out.Write([]byte(boilerPlate))
	if err != nil {
		return err
	}

	err = rootCommand.cobraCommand.GenBashCompletion(out)
	if err != nil {
		return err
	}

	return nil
}
