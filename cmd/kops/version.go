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
	"encoding/json"
	"fmt"
	"io"

	"sigs.k8s.io/yaml"

	"github.com/spf13/cobra"
	"k8s.io/kops"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops/registry"
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
	cmd.Flags().StringVarP(&options.Output, "output", "o", options.Output, "One of 'yaml' or 'json'.")

	return cmd
}

type VersionOptions struct {
	short       bool
	server      bool
	ClusterName string
	Output      string
}

// Version is a struct for version information
type Version struct {
	ClientVersion *kops.Info `json:"clientVersion,omitempty" yaml:"clientVersion,omitempty"`
	ServerVersion string     `json:"serverVersion,omitempty" yaml:"serverVersion,omitempty"`
}

// RunVersion implements the version command logic
func RunVersion(f *util.Factory, out io.Writer, options *VersionOptions) error {
	var versionInfo Version
	clientVersion := kops.Get()
	versionInfo.ClientVersion = &clientVersion
	if options.server {
		versionInfo.ServerVersion = serverVersion(f, options)
	}
	switch options.Output {
	case "":
		if options.short {
			_, err := fmt.Fprintf(out, "%s\n", versionInfo.ClientVersion.Version)
			if err != nil {
				return err
			}
			if options.server {
				_, err := fmt.Fprintf(out, "%s\n", versionInfo.ServerVersion)
				return err
			}
			return nil
		} else {
			client := versionInfo.ClientVersion.Version
			if versionInfo.ClientVersion.GitVersion != "" {
				client += " (git-" + versionInfo.ClientVersion.GitVersion + ")"
			}

			{
				_, err := fmt.Fprintf(out, "Client Version: %s\n", client)
				if err != nil {
					return err
				}
			}
			if options.server {
				_, err := fmt.Fprintf(out, "Last applied server version: %s\n", versionInfo.ServerVersion)
				return err
			}
			return nil
		}
	case OutputYaml:
		marshalled, err := yaml.Marshal(&versionInfo)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(out, string(marshalled))
		return err
	case OutputJSON:
		marshalled, err := json.MarshalIndent(&versionInfo, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(out, string(marshalled))
		return err
	default:
		return fmt.Errorf("VersionOptions were not validated: --output=%q should have been rejected", options.Output)
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
	configBase, err := f.VFSContext().BuildVfsPath(cluster.Spec.ConfigStore.Base)
	if err != nil {
		return "could not talk to vfs"
	}

	kopsVersionUpdatedBytes, err := configBase.Join(registry.PathKopsVersionUpdated).ReadFile(ctx)
	if err != nil {
		return "could get cluster version"
	}
	return string(kopsVersionUpdatedBytes)
}
