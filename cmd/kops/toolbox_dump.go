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

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/util/i18n"
)

var (
	toolbox_dump_long = templates.LongDesc(i18n.T(`
	Displays cluster information.  Includes information about cloud and Kubernetes resources.`))

	toolbox_dump_example = templates.Examples(i18n.T(`
	# Dump cluster information
	kops toolbox dump --name k8s-cluster.example.com
	`))

	toolbox_dump_short = i18n.T(`Dump cluster information`)
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
		Short:   toolbox_dump_short,
		Long:    toolbox_dump_long,
		Example: toolbox_dump_example,
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

	// Todo lets make this smart enough to detect the cloud and switch on the ClusterResources interface
	d := &resources.ClusterResources{}
	d.ClusterName = options.ClusterName
	d.Cloud = cloud

	resources, err := d.ListResources()
	if err != nil {
		return err
	}

	data := make(map[string]interface{})

	dumpedResources := []interface{}{}
	for k, r := range resources {
		if r.Dumper == nil {
			glog.V(8).Infof("skipping dump of %q (no Dumper)", k)
			continue
		}

		o, err := r.Dumper(r)
		if err != nil {
			return fmt.Errorf("error dumping %q: %v", k, err)
		}
		if o != nil {
			dumpedResources = append(dumpedResources, o)
		}
	}
	data["resources"] = dumpedResources

	switch options.Output {
	case OutputYaml:
		b, err := kops.ToRawYaml(data)
		if err != nil {
			return fmt.Errorf("error marshaling yaml: %v", err)
		}
		_, err = out.Write(b)
		if err != nil {
			return fmt.Errorf("error writing to stdout: %v", err)
		}
		return nil

	case OutputJSON:
		b, err := json.MarshalIndent(data, "", "  ")
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
