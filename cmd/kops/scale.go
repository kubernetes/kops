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
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/util/i18n"
)

var (
	//TODO add comments
	scale_long = templates.LongDesc(i18n.T(`Scale resources in/out
	`))

	scale_example = templates.Examples(i18n.T(`kops scale ig --name $NAME nodes --replicas=2
kops scale ig --name $NAME nodes --replicas=0
	`))
)

var scaleCmd = &cobra.Command{
	Use:     "scale",
	Short:   i18n.T(`Scale instancegroups and other resources`),
	Long:    scale_long,
	Example: scale_example,
}

func init() {
	rootCommand.AddCommand(scaleCmd)
}
