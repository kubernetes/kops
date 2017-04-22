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
	"bytes"
	"fmt"
	"io"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kops"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/apis/kops/v1alpha1"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/util/pkg/tables"
	"k8s.io/kops/util/pkg/vfs"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/util/i18n"
)

type ToolboxInventoryOptions struct {
	ClusterName       string
	KubernetesVersion string
	Output            string
	Filenames         []string
	// Channel is the location of the api.Channel to use for our defaults
	Channel string
}

func (o *ToolboxInventoryOptions) InitDefaults() {
	o.Channel = api.DefaultChannel
	o.Output = OutputTable
	o.Channel = "stable"
}

var (
	toolbox_inventory_long = templates.LongDesc(i18n.T(`
		Output a list of IoTk - inventory of all things kops.  Bill of materials (BOM) for a kops installation; containers, binaries, etc.`))

	toolbox_inventory_example = templates.Examples(i18n.T(`
		# Get a inventory list from a yaml file
		kops toolbox inventory -f k8s.example.com.yaml --state s3://k8s.example.com

		# Get a inventory list from a cluster
		kops toolbox inventory k8s.example.com --state s3://k8s.example.com

		`))

	toolbox_inventory_short = i18n.T(`Output a list of IoTk - inventory of all things kops. `)
	toolbox_inventory_use   = i18n.T("inventory")
)

func NewCmdToolboxInventory(f *util.Factory, out io.Writer) *cobra.Command {
	options := &ToolboxInventoryOptions{}
	options.InitDefaults()

	options.ClusterName = rootCommand.ClusterName()

	cmd := &cobra.Command{
		Use:     toolbox_inventory_use,
		Short:   toolbox_inventory_short,
		Example: toolbox_inventory_example,
		Long:    toolbox_inventory_long,
		Run: func(cmd *cobra.Command, args []string) {
			if err := rootCommand.ProcessArgs(args); err != nil {
				exitWithError(err)
			}

			err := rootCommand.ProcessArgs(args)

			if err != nil {
				exitWithError(err)
				return
			}

			options.ClusterName = rootCommand.clusterName

			err = RunToolboxInventory(f, out, options)

			if err != nil {
				exitWithError(err)
				return
			}
		},
	}

	cmd.Flags().StringVarP(&options.Output, "output", "o", options.Output, "output format.  One of: yaml, json")
	cmd.Flags().StringVar(&options.Channel, "channel", options.Channel, "Channel for default versions and configuration to use")
	cmd.Flags().StringVar(&options.KubernetesVersion, "kubernetes-version", options.KubernetesVersion, "Version of kubernetes to run (defaults to version in channel)")
	cmd.Flags().StringArrayVarP(&options.Filenames, "filename", "f", options.Filenames, "Filename to use to create the resource")
	return cmd
}

// RunToolboxInventory executes the business logic to generate a BOM for a kops intallation.
func RunToolboxInventory(f *util.Factory, out io.Writer, options *ToolboxInventoryOptions) error {

	var cluster *api.Cluster
	var ig []*api.InstanceGroup
	var err error
	var clientset simple.Clientset

	clientset, err = options.getClientSet(f)
	if err != nil {
		return err
	}

	if len(options.Filenames) != 0 {
		cluster, ig, err = options.readFiles(options)
		if err != nil {
			return fmt.Errorf("Error loading file(s) %q, %v", options.Filenames, err)
		}
	} else if options.ClusterName != "" {

		cluster, err = clientset.Clusters().Get(options.ClusterName)

		if err != nil {
			return fmt.Errorf("Error getting cluster  %q", err)
		}

		if cluster == nil {
			return fmt.Errorf("cluster not found %q", options.ClusterName)
		}

		configBase, err := registry.ConfigBase(cluster)
		if err != nil {
			return fmt.Errorf("error reading full cluster spec for %q: %v", cluster.ObjectMeta.Name, err)
		}

		err = registry.ReadConfigDeprecated(configBase.Join(registry.PathClusterCompleted), cluster)
		if err != nil {
			return fmt.Errorf("error reading full cluster spec for %q: %v", cluster.ObjectMeta.Name, err)
		}

	}

	if cluster.Spec.KubernetesVersion == "" {
		channel, err := api.LoadChannel(options.Channel)
		if err != nil {
			return fmt.Errorf("Unable to load channel %q", err)
		}
		kubernetesVersion := api.RecommendedKubernetesVersion(channel, kops.Version)
		if kubernetesVersion != nil {
			cluster.Spec.KubernetesVersion = kubernetesVersion.String()
			options.KubernetesVersion = cluster.Spec.KubernetesVersion
		} else {

			return fmt.Errorf("Unable to find kubernetes version")
		}

	}

	inventory := &cloudup.Inventory{}

	a, err := inventory.Build(cluster, ig, clientset)

	if err != nil {
		return fmt.Errorf("error building inventory assests: %v", err)
	}

	switch options.Output {
	case OutputTable:
		fmt.Fprintf(out, "\n\n")
		fmt.Fprintf(out, "Inventory for cluster %q\n\n", cluster.Name)
		fmt.Fprintf(out, "\n")

		t := &tables.Table{}
		t.AddColumn("TYPE", func(i *cloudup.InventoryAsset) string {
			return i.Type
		})
		t.AddColumn("ASSET", func(i *cloudup.InventoryAsset) string {
			return i.Data
		})
		return t.Render(a, out, "TYPE", "ASSET")
	default:
		return fmt.Errorf("Unknown output format: %q", options.Output)
	}
}

// readFiles inputs and marshalls YAML files.
func (o *ToolboxInventoryOptions) readFiles(options *ToolboxInventoryOptions) (*api.Cluster, []*api.InstanceGroup, error) {

	codec := api.Codecs.UniversalDecoder(api.SchemeGroupVersion)

	var cluster *api.Cluster
	var ig []*api.InstanceGroup

	for _, f := range options.Filenames {
		var sb bytes.Buffer
		fmt.Fprintf(&sb, "\n")

		contents, err := vfs.Context.ReadFile(f)
		if err != nil {
			return nil, nil, fmt.Errorf("error reading file %q: %v", f, err)
		}

		sections := bytes.Split(contents, []byte("\n---\n"))
		for _, section := range sections {
			defaults := &schema.GroupVersionKind{
				Group:   v1alpha1.SchemeGroupVersion.Group,
				Version: v1alpha1.SchemeGroupVersion.Version,
			}
			o, _, err := codec.Decode(section, defaults, nil)
			if err != nil {
				return nil, nil, fmt.Errorf("error parsing file %q: %v", f, err)
			}

			switch v := o.(type) {
			case *api.Cluster:
				err = cloudup.PerformAssignments(v)
				if err != nil {
					return nil, nil, fmt.Errorf("error populating configuration: %v", err)
				}
				cluster = v
			case *api.InstanceGroup:
				ig = append(ig, v)
			default:
				glog.V(8).Infof("Type of object was %T", v)
			}
		}
	}

	return cluster, ig, nil
}

// getClientSet returns a clientset.
func (o *ToolboxInventoryOptions) getClientSet(f *util.Factory) (simple.Clientset, error) {
	clientset, err := f.Clientset()
	if err != nil {
		return nil, fmt.Errorf("unable to load client set %v", err)
	}

	return clientset, nil
}
