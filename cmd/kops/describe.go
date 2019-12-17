/*
Copyright 2019 The Kubernetes Authors.

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
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
	"k8s.io/kubernetes/pkg/kubectl/util/templates"
)

var (
	describeLong = templates.LongDesc(i18n.T(`
	Get additional information about cloud and cluster resources.
	`))

	describeExample = templates.Examples(i18n.T(`
	`))
	describeShort = i18n.T(`Describe a resource.`)
)

// DescribeCmd represents the describe command
type DescribeCmd struct {
	cobraCommand *cobra.Command
}

var describeCmd = DescribeCmd{
	cobraCommand: &cobra.Command{
		Use:     "describe",
		Short:   describeShort,
		Long:    describeLong,
		Example: describeExample,
	},
}

func init() {
	cmd := describeCmd.cobraCommand

	rootCommand.AddCommand(cmd)
}
