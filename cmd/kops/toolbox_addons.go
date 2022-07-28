/*
Copyright 2022 The Kubernetes Authors.

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
	"io"

	channelscmd "k8s.io/kops/channels/pkg/cmd"
	"k8s.io/kops/cmd/kops/util"

	"github.com/spf13/cobra"
)

func NewCmdToolboxAddons(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "addons",
		Short:         "Manage addons",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	f := util.NewFactory(nil)
	ctx := context.Background()

	// create subcommands
	cmd.AddCommand(&cobra.Command{
		Use:     "apply CHANNEL",
		Short:   "Applies updates from the given channel",
		Example: "kops toolbox addons apply s3://<state_store>/<cluster_name>/addons/bootstrap-channel.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			return channelscmd.RunApplyChannel(ctx, f, out, &channelscmd.ApplyChannelOptions{}, args)
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "Lists installed addons",
		RunE: func(cmd *cobra.Command, args []string) error {
			return channelscmd.RunGetAddons(ctx, f, out, &channelscmd.GetAddonsOptions{})
		},
	})

	return cmd
}
