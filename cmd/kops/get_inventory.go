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
	"io/ioutil"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kops"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/apis/kops/v1alpha1"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/tables"
	"k8s.io/kops/util/pkg/vfs"
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

	a, clusterName, err := extractAssets(f, options)
	if err != nil {
		return fmt.Errorf("Error getting inventory: %v", err)
	}

	a.Spec.Cluster = nil

	switch options.output {
	case OutputTable:
		fmt.Fprintf(out, "\n\n")
		fmt.Fprintf(out, "ApplyInventory for cluster %q\n\n", clusterName)
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

		fmt.Fprintf(out, "COMPRESSED FILES\n\n")

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

		fmt.Fprintf(out, "\nCONTAINERS\n\n")
		t = &tables.Table{}
		t.AddColumn("ASSET", func(i *api.ContainerAsset) string {
			return i.String
		})

		// TODO FIXME
		t.AddColumn("NAME", func(i *api.ContainerAsset) string {
			return i.Name
		})
		if err := t.Render(a.Spec.ContainerAssets, out, "NAME", "ASSET"); err != nil {
			return err
		}

		for _, s := range a.Spec.ContainerAssets {
			glog.V(2).Infof("container: %+v", s)
		}

		fmt.Fprintf(out, "\nHOSTS\n\n")
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

		for _, s := range a.Spec.ContainerAssets {
			glog.V(2).Infof("container: %+v", s)
		}
		/*
			t = &tables.Table{}
			t.AddColumn("ASSET", func(i *api.ChannelAsset) string {
				return i.Location
			})
			t.AddColumn("NAME", func(i *api.ChannelAsset) string {
				return i.Name
			})
			if err := t.Render(a.Spec.ChannelAsset, out, "NAME", "ASSET"); err != nil {
				return err
			}*/

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

// readFiles inputs and marshalls YAML files.
func readFiles(options *GetInventoryOptions) (*api.Cluster, []*api.InstanceGroup, error) {

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
func getClientSet(f *util.Factory) (simple.Clientset, error) {
	clientset, err := f.Clientset()
	if err != nil {
		return nil, fmt.Errorf("unable to load client set %v", err)
	}

	return clientset, nil
}

func extractAssets(f *util.Factory, options *GetInventoryOptions) (*api.Inventory, string, error) {
	var cluster *api.Cluster
	var ig []*api.InstanceGroup
	var err error
	var clientset simple.Clientset

	clientset, err = getClientSet(f)
	if err != nil {
		return nil, "", err
	}

	if len(options.Filenames) != 0 {
		cluster, ig, err = readFiles(options)
		if err != nil {
			return nil, "", fmt.Errorf("Error loading file(s) %q, %v", options.Filenames, err)
		}

		options.clusterName = cluster.ObjectMeta.Name
	} else if options.clusterName != "" {

		cluster, err = clientset.Clusters().Get(options.clusterName)

		if err != nil {
			return nil, "", fmt.Errorf("Error getting cluster  %q", err)
		}

		if cluster == nil {
			return nil, "", fmt.Errorf("cluster not found %q", options.clusterName)
		}

	}

	if cluster == nil {
		return nil, "", fmt.Errorf("error getting cluster")
	}

	if cluster.Spec.KubernetesVersion == "" {
		channel, err := api.LoadChannel(options.Channel)
		if err != nil {
			return nil, "", fmt.Errorf("Unable to load channel %q", err)
		}
		kubernetesVersion := api.RecommendedKubernetesVersion(channel, kops.Version)
		if kubernetesVersion != nil {
			cluster.Spec.KubernetesVersion = kubernetesVersion.String()
			options.KubernetesVersion = cluster.Spec.KubernetesVersion
		} else {

			return nil, "", fmt.Errorf("Unable to find kubernetes version")
		}

	}

	// TODO check if the cluster has a key?
	// TODO how do we get this out of here??
	// TODO - Reference: https://github.com/kubernetes/kops/issues/2659
	sshPublicKeys := make(map[string][]byte)
	if options.SSHPublicKey != "" {
		options.SSHPublicKey = utils.ExpandPath(options.SSHPublicKey)
		authorized, err := ioutil.ReadFile(options.SSHPublicKey)
		if err != nil {
			return nil, "", fmt.Errorf("error reading SSH key file %q: %v", options.SSHPublicKey, err)
		}
		sshPublicKeys[fi.SecretNameSSHPrimary] = authorized

	}

	keyStore, err := registry.KeyStore(cluster)
	if err != nil {
		return nil, "", err
	}

	for k, data := range sshPublicKeys {
		err = keyStore.AddSSHPublicKey(k, data)
		if err != nil {
			return nil, "", fmt.Errorf("error addding SSH public key: %v", err)
		}
	}

	applyClusterCmd := &cloudup.ApplyClusterCmd{
		Clientset:      clientset,
		Cluster:        cluster,
		InstanceGroups: ig,
		TargetName:     cloudup.TargetInventory,
		Models:         []string{"config", "proto", "cloudup"},
	}

	err = applyClusterCmd.Run()

	if err != nil {
		return nil, "", fmt.Errorf("error applying cluster build: %v", err)
	}

	a := applyClusterCmd.Inventory

	clusterExists, err := clientset.Clusters().Get(options.clusterName)

	// Need to delete tmp ssh key
	// If we can get ApplyClusterCmd to run w/o ssh key we will not have to create it
	if err != nil || clusterExists == nil {

		glog.V(2).Infof("Deleting cluster resources for %q", options.clusterName)
		configBase, err := registry.ConfigBase(cluster)

		if err != nil {
			return nil, "", err
		}

		err = registry.DeleteAllClusterState(configBase)
		if err != nil {
			return nil, "", fmt.Errorf("error removing cluster from state store: %v", err)
		}

	}

	return a, options.clusterName, nil
}
