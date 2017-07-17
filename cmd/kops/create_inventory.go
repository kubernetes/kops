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

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/util/i18n"
	"strings"
)

type CreateInventoryOptions struct {
	// Maybe we may this a sub command then?
	*CreateOptions
	Registry          string
	FileDestination   string
	StageFiles        bool
	StageContainers   bool
	clusterName       string
	KubernetesVersion string
	Channel           string
	SSHPublicKey      string
}

func (o *CreateInventoryOptions) InitDefaults() {
	o.Channel = api.DefaultChannel
	o.StageContainers = true
	o.StageFiles = true
	o.SSHPublicKey = "~/.ssh/id_rsa.pub"
}

var (
	create_inventory_long = templates.LongDesc(i18n.T(`
		Upload inventory files to specified destinations(Registry/FileDestination).
		
		Note: 
		
		1. This command assumes Docker is installed and the user has the privileges to load and push images.
		2. User is authenticated to the provided Docker registry.`))

	create_inventory_example = templates.Examples(i18n.T(`
		# Stage inventory files from a yaml file
		kops create inventory --registry quay.io/vorstella --fileDestination s3://mybucket -f mycluster.yaml

		`))

	create_inventory_short = i18n.T(`Update inventory files to the specified destinations(Registry/File Destination).`)
	create_inventory_use   = i18n.T("inventory")
)

// FIXME need to document all of the public methods. Follow go standards.
// FIXME need to not write over containers - check to see if the container / binary exists
// FIXME need a force mode which forces the containers / binaries to upload
// FIXME you have some english errors as well
// FIXME make stuff like log messages "FileAssetTransferer.Transfer" more user friendly.  Think IT admin talk
// FIXME we are not uploading SHAs
// FIXME need a dryrun mode --yes

func NewCmdCreateInventory(f *util.Factory, out io.Writer, createOptions *CreateOptions) *cobra.Command {

	options := &CreateInventoryOptions{
		CreateOptions: createOptions,
	}
	options.InitDefaults()

	cmd := &cobra.Command{
		Use:     create_inventory_use,
		Short:   create_inventory_short,
		Example: create_inventory_example,
		Long:    create_inventory_long,
		Run: func(cmd *cobra.Command, args []string) {
			if err := rootCommand.ProcessArgs(args); err != nil {
				exitWithError(err)
			}

			err := rootCommand.ProcessArgs(args)

			if err != nil {
				exitWithError(err)
				return
			}

			if rootCommand.clusterName != "" && len(options.Filenames) == 0 {
				options.clusterName = rootCommand.clusterName
			} else {
				options.clusterName = ""
			}

			if len(options.Filenames) == 0 && options.clusterName == "" {
				exitWithError(fmt.Errorf("--filename or --name option must be used to supply cluster information."))
				return
			}

			if options.FileDestination == "" && options.StageFiles {
				exitWithError(fmt.Errorf("Please provide s3 location via --file-destination flag."))
				return
			}

			if options.Registry == "" && options.StageContainers {
				exitWithError(fmt.Errorf("Please provide registry location via --repository flag."))
				return
			}

			if !options.StageFiles && !options.StageContainers {
				exitWithError(fmt.Errorf("Please choose at least one of --stage-containers or --stage-files flag(s)."))
				return
			}

			err = RunCreateInventory(f, out, options)

			if err != nil {
				exitWithError(err)
				return
			}
		},
	}

	cmd.Flags().StringVarP(&options.Channel, "channel", "c", options.Channel, "Channel for default versions and configuration to use")
	cmd.Flags().StringVarP(&options.KubernetesVersion, "kubernetes-version", "k", options.KubernetesVersion, "Version of kubernetes to run (defaults to version in channel)")
	cmd.Flags().StringArrayVarP(&options.Filenames, "filename", "f", options.Filenames, "Filename to use to create the resource")
	cmd.Flags().StringVarP(&options.Registry, "registry", "r", options.Registry, "Registry location used to stage inventory containers")
	cmd.Flags().StringVarP(&options.FileDestination, "file-repository", "p", options.FileDestination, "File repository location used to stage inventory files")
	cmd.Flags().BoolVar(&options.StageContainers, "stage-containers", options.StageContainers, "Stage containers")
	cmd.Flags().BoolVar(&options.StageFiles, "stage-files", options.StageFiles, "Stage files")
	cmd.MarkFlagRequired("file-destination")
	cmd.MarkFlagRequired("registry")

	return cmd
}

// RunCreateInventory executes the business logic to stage inventory files to the specified repositories.
func RunCreateInventory(f *util.Factory, out io.Writer, options *CreateInventoryOptions) error {

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

	assets, _, err := extractInventory.ExtractAssets()
	if err != nil {
		return fmt.Errorf("Error getting inventory: %v", err)
	}

	options.FileDestination = strings.TrimSuffix(options.FileDestination, "/")

	// FIXME refactor too many parameters now :(
	stageInventory := cloudup.NewStageInventory(options.FileDestination, options.StageFiles, options.Registry, options.StageContainers, assets)
	err = stageInventory.Run()
	if err != nil {
		return fmt.Errorf("Error processing assets file(s) %q, %v", options.Filenames, err)
	}

	return nil
}
