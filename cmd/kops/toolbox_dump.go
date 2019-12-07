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

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/resources"
	resourceops "k8s.io/kops/pkg/resources/ops"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	toolboxDumpLong = templates.LongDesc(i18n.T(`
	Displays cluster information.  Includes information about cloud and Kubernetes resources.`))

	toolboxDumpExample = templates.Examples(i18n.T(`
	# Dump cluster information
	kops toolbox dump --name k8s-cluster.example.com
	`))

	toolboxDumpShort = i18n.T(`Dump cluster information`)
)

type ToolboxDumpOptions struct {
	Output string

	ClusterName string
}

func (o *ToolboxDumpOptions) InitDefaults() {
	o.Output = OutputYaml
}

func NewCmdToolboxDump(f *util.Factory, out io.Writer) *cobra.Command {
	options := &ToolboxDumpOptions{}
	options.InitDefaults()

	cmd := &cobra.Command{
		Use:     "dump",
		Short:   toolboxDumpShort,
		Long:    toolboxDumpLong,
		Example: toolboxDumpExample,
		Run: func(cmd *cobra.Command, args []string) {
			if err := rootCommand.ProcessArgs(args); err != nil {
				exitWithError(err)
			}

			options.ClusterName = rootCommand.ClusterName()

			err := RunToolboxDump(f, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	// TODO: Push up to top-level command?
	// Yes please! (@kris-nova)
	cmd.Flags().StringVarP(&options.Output, "output", "o", options.Output, "output format.  One of: yaml, json")

	return cmd
}

func RunToolboxDump(f *util.Factory, out io.Writer, options *ToolboxDumpOptions) error {
	clientset, err := f.Clientset()
	if err != nil {
		return err
	}

	if options.ClusterName == "" {
		return fmt.Errorf("ClusterName is required")
	}

	cluster, err := clientset.GetCluster(options.ClusterName)
	if err != nil {
		return err
	}

	if cluster == nil {
		return fmt.Errorf("cluster not found %q", options.ClusterName)
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return err
	}

	region := "" // Use default
	resourceMap, err := resourceops.ListResources(cloud, options.ClusterName, region)
	if err != nil {
		return err
	}
	dump, err := resources.BuildDump(context.TODO(), cloud, resourceMap)
	if err != nil {
		return err
	}

	switch options.Output {
	case OutputYaml:
		b, err := kops.ToRawYaml(dump)
		if err != nil {
			return fmt.Errorf("error marshaling yaml: %v", err)
		}
		_, err = out.Write(b)
		if err != nil {
			return fmt.Errorf("error writing to stdout: %v", err)
		}
		return nil

	case OutputJSON:
		b, err := json.MarshalIndent(dump, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshaling json: %v", err)
		}
		_, err = out.Write(b)
		if err != nil {
			return fmt.Errorf("error writing to stdout: %v", err)
		}
		return nil

	default:
		return fmt.Errorf("Unsupported output format: %q", options.Output)
	}
}
