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
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"k8s.io/kops"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/util/pkg/vfs"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	versionLong = templates.LongDesc(i18n.T(`
	Print the kOps version and git SHA.`))

	versionExample = templates.Examples(i18n.T(`
	kops version`))

	versionShort = i18n.T(`Print the kOps version information.`)
)

// NewCmdVersion builds a cobra command for the kops version command
func NewCmdVersion(f *util.Factory, out io.Writer) *cobra.Command {
	options := &VersionOptions{}

	cmd := &cobra.Command{
		Use:               "version",
		Short:             versionShort,
		Long:              versionLong,
		Example:           versionExample,
		Args:              rootCommand.clusterNameArgsAllowNoCluster(&options.ClusterName),
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunVersion(f, out, options)
		},
	}

	cmd.Flags().BoolVar(&options.short, "short", options.short, "only print the main kOps version. Useful for scripting.")
	cmd.Flags().BoolVar(&options.server, "server", options.server, "show the kOps version that made the last change to the state store.")

	return cmd
}

type VersionOptions struct {
	short       bool
	server      bool
	ClusterName string
}

// RunVersion implements the version command logic
func RunVersion(f *util.Factory, out io.Writer, options *VersionOptions) error {
	if options.short {
		s := kops.Version
		_, err := fmt.Fprintf(out, "%s\n", s)
		if err != nil {
			return err
		}
		if options.server {
			server := serverVersion(f, options)

			_, err := fmt.Fprintf(out, "%s\n", server)
			return err
		}

		return nil
	} else {
		client := kops.Version
		if kops.GitVersion != "" {
			client += " (git-" + kops.GitVersion + ")"
		}

		{
			_, err := fmt.Fprintf(out, "Client version: %s\n", client)
			if err != nil {
				return err
			}
		}
		if options.server {
			server := serverVersion(f, options)

			_, err := fmt.Fprintf(out, "Last applied server version: %s\n", server)
			return err
		}
		return nil
	}
}

func serverVersion(f *util.Factory, options *VersionOptions) string {
	if options.ClusterName == "" {
		return "No cluster selected"
	}

	ctx := context.Background()
	cluster, err := GetCluster(ctx, f, options.ClusterName)
	if err != nil {
		return "could not fetch cluster"
	}
	configBase, err := vfs.Context.BuildVfsPath(cluster.Spec.ConfigBase)
	if err != nil {
		return "could not talk to vfs"
	}

	kopsVersionUpdatedBytes, err := configBase.Join(registry.PathKopsVersionUpdated).ReadFile()
	if err != nil {
		return "could get cluster version"
	}
	return string(kopsVersionUpdatedBytes)
}
