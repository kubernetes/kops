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
	"os"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	createSecretDockerConfigLong = templates.LongDesc(i18n.T(`
	Create a new Docker config and store it in the state store.
	Used to configure Docker authentication on each node.
	
	After creating a dockerconfig secret a /root/.docker/config.json file
    will be added to newly created nodes. This file will be used by Kubernetes
    to authenticate to container registries.

	This will also work when using containerd as the container runtime.`))

	createSecretDockerConfigExample = templates.Examples(i18n.T(`
	# Create a new Docker config.
	kops create secret dockerconfig -f /path/to/docker/config.json \
		--name k8s-cluster.example.com --state s3://my-state-store

	# Create a docker config via stdin.
	generate-docker-config.sh | kops create secret dockerconfig -f - \
		--name k8s-cluster.example.com --state s3://my-state-store

	# Replace an existing docker config secret.
	kops create secret dockerconfig -f /path/to/docker/config.json --force \
		--name k8s-cluster.example.com --state s3://my-state-store
	`))

	createSecretDockerConfigShort = i18n.T(`Create a Docker config.`)
)

type CreateSecretDockerConfigOptions struct {
	ClusterName      string
	DockerConfigPath string
	Force            bool
}

func NewCmdCreateSecretDockerConfig(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateSecretDockerConfigOptions{}

	cmd := &cobra.Command{
		Use:               "dockerconfig [CLUSTER] -f FILENAME",
		Short:             createSecretDockerConfigShort,
		Long:              createSecretDockerConfigLong,
		Example:           createSecretDockerConfigExample,
		Args:              rootCommand.clusterNameArgs(&options.ClusterName),
		ValidArgsFunction: commandutils.CompleteClusterName(f, true, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunCreateSecretDockerConfig(context.TODO(), f, out, options)
		},
	}

	cmd.Flags().StringVarP(&options.DockerConfigPath, "filename", "f", "", "Path to Docker config JSON file")
	cmd.MarkFlagRequired("filename")
	cmd.RegisterFlagCompletionFunc("filename", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json"}, cobra.ShellCompDirectiveFilterFileExt
	})
	cmd.Flags().BoolVar(&options.Force, "force", options.Force, "Force replace the secret if it already exists")

	return cmd
}

func RunCreateSecretDockerConfig(ctx context.Context, f commandutils.Factory, out io.Writer, options *CreateSecretDockerConfigOptions) error {
	cluster, err := GetCluster(ctx, f, options.ClusterName)
	if err != nil {
		return err
	}

	clientset, err := f.KopsClient()
	if err != nil {
		return err
	}

	secretStore, err := clientset.SecretStore(cluster)
	if err != nil {
		return err
	}

	var data []byte
	if options.DockerConfigPath == "-" {
		data, err = ConsumeStdin()
		if err != nil {
			return fmt.Errorf("reading Docker config from stdin: %v", err)
		}
	} else {
		data, err = os.ReadFile(options.DockerConfigPath)
		if err != nil {
			return fmt.Errorf("reading Docker config %v: %v", options.DockerConfigPath, err)
		}
	}

	var parsedData map[string]interface{}
	err = json.Unmarshal(data, &parsedData)
	if err != nil {
		return fmt.Errorf("unable to parse JSON %v: %v", options.DockerConfigPath, err)
	}

	secret := &fi.Secret{
		Data: data,
	}

	if !options.Force {
		_, created, err := secretStore.GetOrCreateSecret("dockerconfig", secret)
		if err != nil {
			return fmt.Errorf("adding dockerconfig secret: %v", err)
		}
		if !created {
			return fmt.Errorf("failed to create the dockerconfig secret as it already exists. Pass the `--force` flag to replace an existing secret")
		}
	} else {
		_, err := secretStore.ReplaceSecret("dockerconfig", secret)
		if err != nil {
			return fmt.Errorf("updating dockerconfig secret: %v", err)
		}
	}

	return nil
}
