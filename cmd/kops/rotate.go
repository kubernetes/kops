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

package main

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	rotateLong = templates.LongDesc(i18n.T(`
		Rotate a secret.
	`))

	rotateExample = templates.Examples(i18n.T(`
	# Rotate the service-account key
	kops rotate secret service-account
	`))
	rotateShort = i18n.T("Rotate a secret.")
)

// DescribeCmd represents the describe command
type RotateCmd struct {
	cobraCommand *cobra.Command
}

var rotateCmd = RotateCmd{
	cobraCommand: &cobra.Command{
		Use:     "rotate",
		Short:   rotateShort,
		Long:    rotateLong,
		Example: rotateExample,
	},
}

func init() {
	cmd := rotateCmd.cobraCommand

	rootCommand.AddCommand(cmd)
}
