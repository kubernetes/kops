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
	"encoding/json"
	"fmt"
	"io"

	"os"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/util/i18n"
)

var (
	get_long = templates.LongDesc(i18n.T(`
	Display one or many resources.` + validResources))

	get_example = templates.Examples(i18n.T(`
	# Get all resource in a single cluster as yaml
	kops get --name k8s-cluster.example.com -o yaml

	# Get all clusters in a state store
	kops get clusters

	# Get a cluster
	kops get cluster k8s-cluster.example.com

	# Get a cluster YAML cluster spec
	kops get cluster k8s-cluster.example.com -o yaml

	# Get an instancegroup
	kops get ig --name k8s-cluster.example.com nodes

	# Get a secret
	kops get secrets kube -oplaintext

	# Get the admin password for a cluster
	kops get secrets admin -oplaintext`))

	get_short = i18n.T(`Get one or many resources.`)
)

type GetOptions struct {
	output      string
	clusterName string
}

const (
	OutputYaml  = "yaml"
	OutputTable = "table"
	OutputJSON  = "json"
)

func NewCmdGet(f *util.Factory, out io.Writer) *cobra.Command {
	options := &GetOptions{
		output: OutputTable,
	}

	cmd := &cobra.Command{
		Use:        "get",
		SuggestFor: []string{"list"},
		Short:      get_short,
		Long:       get_long,
		Example:    get_example,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 0 {
				options.clusterName = args[0]
			}

			if rootCommand.clusterName != "" {
				if len(args) != 0 {
					exitWithError(fmt.Errorf("cannot mix --name for cluster with positional arguments"))
				}

				options.clusterName = rootCommand.clusterName
			}

			err := RunGet(&rootCommand, os.Stdout, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&options.output, "output", "o", options.output, "output format.  One of: table, yaml, json")

	// create subcommands
	cmd.AddCommand(NewCmdGetCluster(f, out, options))
	cmd.AddCommand(NewCmdGetFederations(f, out, options))
	cmd.AddCommand(NewCmdGetInstanceGroups(f, out, options))
	cmd.AddCommand(NewCmdGetSecrets(f, out, options))

	return cmd
}

func RunGet(context Factory, out io.Writer, options *GetOptions) error {

	client, err := context.Clientset()
	if err != nil {
		return err
	}

	cluster, err := client.GetCluster(options.clusterName)
	if err != nil {
		return err
	}

	if cluster == nil {
		fmt.Fprintf(os.Stderr, "No cluster found\n")
		return nil
	}

	clusterList := &api.ClusterList{}
	clusterList.Items = make([]api.Cluster, 1)
	clusterList.Items[0] = *cluster

	args := make([]string, 0)

	clusters, err := buildClusters(args, clusterList)

	ig, err := client.InstanceGroupsFor(cluster).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	if ig == nil || ig.Items == nil || len(ig.Items) == 0 {
		fmt.Fprintf(os.Stderr, "No instance groups found\n")
	}

	instancegroups, err := buildInstanceGroups(args, ig)
	if err != nil {
		return err
	}

	switch options.output {
	case OutputYaml:

		err = clusterOutputYAML(clusters, out)
		if err != nil {
			return err
		}

		if err := writeYAMLSep(out); err != nil {
			return err
		}

		err = igOutputYAML(instancegroups, out)
		if err != nil {
			return err
		}

	case OutputJSON:
		return fmt.Errorf("not implemented")
		// TODO this is not outputing valid json.  Not sure what cluster and instance groups should look like
		/*
			err = clusterOutputJson(clusters,out)
			if err != nil {
				return err
			}
			err = igOutputJson(instancegroups,out)
			if err != nil {
				return err
			}*/

	case OutputTable:
		fmt.Fprintf(os.Stdout, "Cluster\n")
		err = clusterOutputTable(clusters, out)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "\nInstance Groups\n")
		err = igOutputTable(instancegroups, out)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("Unknown output format: %q", options.output)
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
	y, err := api.ToVersionedYaml(obj)
	if err != nil {
		return nil, fmt.Errorf("error marshaling yaml: %v", err)
	}
	return y, nil
}

// obj must be a pointer to a marshalable object
func marshalJSON(obj runtime.Object) ([]byte, error) {
	j, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshaling json: %v", err)
	}
	return j, nil
}
