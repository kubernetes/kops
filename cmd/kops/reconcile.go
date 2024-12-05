/*
Copyright 2024 The Kubernetes Authors.

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
	"io"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kubectl/pkg/util/i18n"
)

var reconcileShort = i18n.T("Reconcile a cluster.")

func NewCmdReconcile(f *util.Factory, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reconcile",
		Short: reconcileShort,
	}

	//  subcommands
	cmd.AddCommand(NewCmdReconcileCluster(f, out))

	return cmd
}
