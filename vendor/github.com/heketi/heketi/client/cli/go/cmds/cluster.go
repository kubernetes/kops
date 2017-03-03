//
// Copyright (c) 2015 The heketi Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package cmds

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/heketi/heketi/client/api/go-client"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(clusterCommand)
	clusterCommand.AddCommand(clusterCreateCommand)
	clusterCommand.AddCommand(clusterDeleteCommand)
	clusterCommand.AddCommand(clusterListCommand)
	clusterCommand.AddCommand(clusterInfoCommand)
	clusterCreateCommand.SilenceUsage = true
	clusterDeleteCommand.SilenceUsage = true
	clusterInfoCommand.SilenceUsage = true
	clusterListCommand.SilenceUsage = true
}

var clusterCommand = &cobra.Command{
	Use:   "cluster",
	Short: "Heketi cluster management",
	Long:  "Heketi Cluster Management",
}

var clusterCreateCommand = &cobra.Command{
	Use:     "create",
	Short:   "Create a cluster",
	Long:    "Create a cluster",
	Example: "  $ heketi-cli cluster create",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create a client to talk to Heketi
		heketi := client.NewClient(options.Url, options.User, options.Key)
		// Create cluster
		cluster, err := heketi.ClusterCreate()
		if err != nil {
			return err
		}

		// Check if JSON should be printed
		if options.Json {
			data, err := json.Marshal(cluster)
			if err != nil {
				return err
			}
			fmt.Fprintf(stdout, string(data))
		} else {
			fmt.Fprintf(stdout, "Cluster id: %v\n", cluster.Id)
		}

		return nil
	},
}

var clusterDeleteCommand = &cobra.Command{
	Use:     "delete [cluster_id]",
	Short:   "Delete the cluster",
	Long:    "Delete the cluster",
	Example: "  $ heketi-cli cluster delete 886a86a868711bef83001",
	RunE: func(cmd *cobra.Command, args []string) error {
		s := cmd.Flags().Args()

		//ensure proper number of args
		if len(s) < 1 {
			return errors.New("Cluster id missing")
		}

		//set clusterId
		clusterId := cmd.Flags().Arg(0)

		// Create a client
		heketi := client.NewClient(options.Url, options.User, options.Key)

		//set url
		err := heketi.ClusterDelete(clusterId)
		if err == nil {
			fmt.Fprintf(stdout, "Cluster %v deleted\n", clusterId)
		}

		return err
	},
}

var clusterInfoCommand = &cobra.Command{
	Use:     "info [cluster_id]",
	Short:   "Retrieves information about cluster",
	Long:    "Retrieves information about cluster",
	Example: "  $ heketi-cli cluster info 886a86a868711bef83001",
	RunE: func(cmd *cobra.Command, args []string) error {
		s := cmd.Flags().Args()
		if len(s) < 1 {
			return errors.New("Cluster id missing")
		}

		//set clusterId
		clusterId := cmd.Flags().Arg(0)

		// Create a client to talk to Heketi
		heketi := client.NewClient(options.Url, options.User, options.Key)

		// Create cluster
		info, err := heketi.ClusterInfo(clusterId)
		if err != nil {
			return err
		}

		// Check if JSON should be printed
		if options.Json {
			data, err := json.Marshal(info)
			if err != nil {
				return err
			}
			fmt.Fprintf(stdout, string(data))
		} else {
			fmt.Fprintf(stdout, "Cluster id: %v\n", info.Id)
			fmt.Fprintf(stdout, "Nodes:\n%v", strings.Join(info.Nodes, "\n"))
			fmt.Fprintf(stdout, "\nVolumes:\n%v", strings.Join(info.Volumes, "\n"))
		}

		return nil
	},
}

var clusterListCommand = &cobra.Command{
	Use:     "list",
	Short:   "Lists the clusters managed by Heketi",
	Long:    "Lists the clusters managed by Heketi",
	Example: "  $ heketi-cli cluster list",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create a client
		heketi := client.NewClient(options.Url, options.User, options.Key)

		// List clusters
		list, err := heketi.ClusterList()
		if err != nil {
			return err
		}

		if options.Json {
			data, err := json.Marshal(list)
			if err != nil {
				return err
			}
			fmt.Fprintf(stdout, string(data))
		} else {
			output := strings.Join(list.Clusters, "\n")
			fmt.Fprintf(stdout, "Clusters:\n%v\n", output)
		}

		return nil
	},
}
