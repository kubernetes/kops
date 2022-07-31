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
	"os"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	getLong = templates.LongDesc(i18n.T(`
	Display one or many resources.` + validResources))

	getExample = templates.Examples(i18n.T(`
	# Get a cluster and its instance groups
	kops get k8s-cluster.example.com

	# Get a cluster and its instancegroups' YAML desired configuration
	kops get k8s-cluster.example.com -o yaml

	# Save a cluster and its instancegroups' desired configuration to YAML file
	kops get k8s-cluster.example.com -o yaml > cluster-desired-config.yaml
	`))

	getShort = i18n.T(`Get one or many resources.`)
)

type GetOptions struct {
	ClusterName string
	Output      string
}

const (
	OutputYaml  = "yaml"
	OutputTable = "table"
	OutputJSON  = "json"
)

func NewCmdGet(f *util.Factory, out io.Writer) *cobra.Command {
	options := &GetOptions{
		Output: OutputTable,
	}

	cmd := &cobra.Command{
		Use:               "get",
		SuggestFor:        []string{"list"},
		Short:             getShort,
		Long:              getLong,
		Example:           getExample,
		Args:              rootCommand.clusterNameArgs(&options.ClusterName),
		ValidArgsFunction: commandutils.CompleteClusterName(f, true, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunGet(context.TODO(), f, out, options)
		},
	}

	cmd.PersistentFlags().StringVarP(&options.Output, "output", "o", options.Output, "output format. One of: table, yaml, json")
	cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{OutputTable, OutputJSON, OutputYaml}, cobra.ShellCompDirectiveNoFileComp
	})

	// create subcommands
	cmd.AddCommand(NewCmdGetAssets(f, out, options))
	cmd.AddCommand(NewCmdGetCluster(f, out, options))
	cmd.AddCommand(NewCmdGetInstanceGroups(f, out, options))
	cmd.AddCommand(NewCmdGetInstances(f, out, options))
	cmd.AddCommand(NewCmdGetKeypairs(f, out, options))
	cmd.AddCommand(NewCmdGetSecrets(f, out, options))
	cmd.AddCommand(NewCmdGetSSHPublicKeys(f, out, options))

	return cmd
}

func RunGet(ctx context.Context, f commandutils.Factory, out io.Writer, options *GetOptions) error {
	client, err := f.KopsClient()
	if err != nil {
		return err
	}

	cluster, err := client.GetCluster(ctx, options.ClusterName)
	if err != nil {
		return err
	}

	if cluster == nil {
		return fmt.Errorf("No cluster found")
	}

	igList, err := client.InstanceGroupsFor(cluster).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	if igList == nil || igList.Items == nil || len(igList.Items) == 0 {
		fmt.Fprintf(os.Stderr, "No instance groups found\n")
	}

	var instancegroups []*api.InstanceGroup
	for i := range igList.Items {
		instancegroups = append(instancegroups, &igList.Items[i])
	}

	var addonObjects []*unstructured.Unstructured
	{
		addons, err := client.AddonsFor(cluster).List()
		if err != nil {
			return err
		}
		for _, addon := range addons {
			addonObjects = append(addonObjects, addon.ToUnstructured())
		}
	}

	var allObjects []runtime.Object
	if options.Output != OutputTable {
		allObjects = append(allObjects, cluster)
		for _, group := range instancegroups {
			allObjects = append(allObjects, group)
		}
		for _, additionalObject := range addonObjects {
			allObjects = append(allObjects, additionalObject)
		}
	}

	switch options.Output {
	case OutputYaml:
		if err := fullOutputYAML(out, allObjects...); err != nil {
			return fmt.Errorf("error writing yaml to stdout: %v", err)
		}

		return nil

	case OutputJSON:
		if err := fullOutputJSON(out, false, allObjects...); err != nil {
			return fmt.Errorf("error writing json to stdout: %v", err)
		}
		return nil

	case OutputTable:
		fmt.Fprintf(out, "Cluster\n")
		err = clusterOutputTable([]*api.Cluster{cluster}, out)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "\nInstance Groups\n")
		err = igOutputTable(cluster, instancegroups, out)
		if err != nil {
			return err
		}
		if len(addonObjects) != 0 {
			fmt.Fprintf(out, "\nAddon Objects\n")
			err = addonsOutputTable(cluster, addonObjects, out)
			if err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("Unknown output format: %q", options.Output)
	}

	return nil
}

func writeYAMLSep(out io.Writer) error {
	_, err := out.Write([]byte("\n---\n\n"))
	if err != nil {
		return fmt.Errorf("error writing to stdout: %v", err)
	}
	return nil
}

type marshalFunc func(obj runtime.Object) ([]byte, error)

func marshalToWriter(obj runtime.Object, marshal marshalFunc, w io.Writer) error {
	b, err := marshal(obj)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	if err != nil {
		return fmt.Errorf("error writing to stdout: %v", err)
	}
	return nil
}

// obj must be a pointer to a marshalable object
func marshalYaml(obj runtime.Object) ([]byte, error) {
	y, err := kopscodecs.ToVersionedYaml(obj)
	if err != nil {
		return nil, fmt.Errorf("error marshaling yaml: %v", err)
	}
	return y, nil
}

// obj must be a pointer to a marshalable object
func marshalJSON(obj runtime.Object) ([]byte, error) {
	j, err := kopscodecs.ToVersionedJSON(obj)
	if err != nil {
		return nil, fmt.Errorf("error marshaling json: %v", err)
	}
	return j, nil
}
