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
	"os"
	"strings"

	"k8s.io/kops/util/pkg/ui"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

var confirmDelete bool

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:        "delete",
	Short:      "delete clusters",
	Long:       `Delete clusters`,
	SuggestFor: []string{"rm"},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		flag.Parse()
		args = flag.Args()
		if !confirmDelete {
			fmt.Printf(
				"Do you really want to %s? This action cannot be undone. (Y/n)\n",
				strings.Join(args, " "),
			)
			if !ui.GetConfirm() {
				os.Exit(1)
			}

		}
	},
}

func init() {
	deleteCmd.PersistentFlags().BoolVarP(&confirmDelete, "yes", "y", false, "Auto confirm deletetion.")

	rootCommand.AddCommand(deleteCmd)
}
