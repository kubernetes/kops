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
	"fmt"
	"io"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/util/pkg/tables"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/util/i18n"
)

type GetInventoryOptions struct {
	KubernetesVersion string
	Channel           string
	*GetOptions
	resource.FilenameOptions
	SSHPublicKey string
}

func (o *GetInventoryOptions) InitDefaults() {
	o.Channel = api.DefaultChannel
	o.output = OutputTable
	o.Channel = "stable"
	o.SSHPublicKey = "~/.ssh/id_rsa.pub"
}

var (
	get_inventory_long = templates.LongDesc(i18n.T(`
		Output a list of IoTk - inventory of all things kops.  Bill of materials (BOM) for a kops installation; containers, binaries, etc.`))

	get_inventory_example = templates.Examples(i18n.T(`
		# Get a inventory list from a YAML file
		kops get inventory -f k8s.example.com.yaml --state s3://k8s.example.com

		# Get a inventory list from a cluster
		kops get inventory k8s.example.com --state s3://k8s.example.com

		# Get a inventory list from a cluster as YAML
		kops get inventory k8s.example.com --state s3://k8s.example.com -o YAML

		`))

	get_inventory_short = i18n.T(`Output a list of IoTk - inventory of all things kops. `)
	get_inventory_use   = i18n.T("inventory")
)

// NewCmdGetInventory sets up a new corbra command.
func NewCmdGetInventory(f *util.Factory, out io.Writer, getOptions *GetOptions) *cobra.Command {
	options := &GetInventoryOptions{
		GetOptions: getOptions,
	}
	options.InitDefaults()

	cmd := &cobra.Command{
		Use:     get_inventory_use,
		Short:   get_inventory_short,
		Example: get_inventory_example,
		Long:    get_inventory_long,
		Run: func(cmd *cobra.Command, args []string) {

			if rootCommand.clusterName != "" && len(options.Filenames) == 0 {
				options.clusterName = rootCommand.clusterName
			} else {
				options.clusterName = ""
			}

			err := RunToolboxInventory(f, out, options)

			if err != nil {
				exitWithError(err)
				return
			}
		},
	}

	cmd.Flags().StringVarP(&options.output, "output", "o", options.output, "output format.  One of: yaml, json, table")
	cmd.Flags().StringVar(&options.Channel, "channel", options.Channel, "Channel for default versions and configuration to use")
	cmd.Flags().StringVar(&options.KubernetesVersion, "kubernetes-version", options.KubernetesVersion, "Version of kubernetes to run (defaults to version in channel)")
	cmd.Flags().StringSliceVarP(&options.Filenames, "filename", "f", options.Filenames, "Filename to use to create the resource")
	return cmd
}

// RunToolboxInventory executes the business logic to generate a BOM for a kops intallation.
func RunToolboxInventory(f *util.Factory, out io.Writer, options *GetInventoryOptions) error {

	clientset, err := getClientSet(f)
	if err != nil {
		return err
	}

	extractInventory := &cloudup.ExtractInventory{
		Clientset:         clientset,
		FilenameOptions:   options.FilenameOptions,
		ClusterName:       options.clusterName,
		KubernetesVersion: options.KubernetesVersion,
		SSHPublicKey:      options.SSHPublicKey,
	}
	a, clusterName, err := extractInventory.ExtractAssets()
	if err != nil {
		return fmt.Errorf("Error getting inventory: %v", err)
	}

	// TODO do we want cluster on the inventory?
	a.Spec.Cluster = nil

	switch options.output {
	case OutputTable:
		fmt.Fprintf(out, "\n\n")
		fmt.Fprintf(out, "Inventory for cluster %q\n\n", clusterName)
		fmt.Fprintf(out, "\n")

		fmt.Fprintf(out, "FILES\n\n")

		t := &tables.Table{}
		t.AddColumn("ASSET", func(i *api.ExecutableFileAsset) string {
			return i.Location
		})
		t.AddColumn("NAME", func(i *api.ExecutableFileAsset) string {
			return i.Name
		})
		if err := t.Render(a.Spec.ExecutableFileAsset, out, "NAME", "ASSET"); err != nil {
			return err
		}

		fmt.Fprintf(out, "\n\nFILE SHAS\n\n")

		t = &tables.Table{}
		t.AddColumn("NAME", func(i *api.ExecutableFileAsset) string {
			return i.Name
		})
		t.AddColumn("SHA", func(i *api.ExecutableFileAsset) string {
			return i.SHA
		})
		if err := t.Render(a.Spec.ExecutableFileAsset, out, "NAME", "SHA"); err != nil {
			return err
		}

		fmt.Fprintf(out, "\n\nCOMPRESSED FILES\n\n")

		t = &tables.Table{}
		t.AddColumn("ASSET", func(i *api.CompressedFileAsset) string {
			return i.Location
		})
		t.AddColumn("NAME", func(i *api.CompressedFileAsset) string {
			return i.Name
		})
		if err := t.Render(a.Spec.CompressedFileAssets, out, "NAME", "ASSET"); err != nil {
			return err
		}

		fmt.Fprintf(out, "\n\nCOMPRESSED FILE SHAS\n\n")

		t = &tables.Table{}
		t.AddColumn("NAME", func(i *api.CompressedFileAsset) string {
			return i.Name
		})
		t.AddColumn("SHA", func(i *api.CompressedFileAsset) string {
			return i.SHA
		})
		if err := t.Render(a.Spec.CompressedFileAssets, out, "NAME", "SHA"); err != nil {
			return err
		}

		fmt.Fprintf(out, "\n\nCONTAINERS\n\n")
		t = &tables.Table{}
		t.AddColumn("ASSET", func(i *api.ContainerAsset) string {
			if i.String != "" {
				return i.String
			} else if i.Location != "" {
				return i.Location
			}

			glog.Errorf("unable to print container asset %q", i)

			return "asset name not set correctly"
		})

		t.AddColumn("NAME", func(i *api.ContainerAsset) string {
			return i.Name
		})
		if err := t.Render(a.Spec.ContainerAssets, out, "NAME", "ASSET"); err != nil {
			return err
		}

		fmt.Fprintf(out, "\n\nHOSTS\n\n")
		t = &tables.Table{}
		t.AddColumn("IMAGE", func(i *api.HostAsset) string {
			return i.Name
		})

		t.AddColumn("INSTANCE GROUP", func(i *api.HostAsset) string {
			return i.InstanceGroup
		})
		if err := t.Render(a.Spec.HostAssets, out, "INSTANCE GROUP", "IMAGE"); err != nil {
			return err
		}

	case OutputJSON:
		if err := marshalToWriter(a, marshalJSON, out); err != nil {
			return err
		}

	case OutputYaml:
		if err := marshalToWriter(a, marshalYaml, out); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unknown output format: %q", options.output)
	}

	return nil

}

// getClientSet returns a clientset.
func getClientSet(f *util.Factory) (simple.Clientset, error) {
	clientset, err := f.Clientset()
	if err != nil {
		return nil, fmt.Errorf("unable to load client set %v", err)
	}

	return clientset, nil
}
