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
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	createSecretDockerconfigLong = templates.LongDesc(i18n.T(`
	Create a new docker config, and store it in the state store.
	Used to configure docker on each master or node (i.e. for auth)
	Use update to modify it, this command will only create a new entry.`))

	createSecretDockerconfigExample = templates.Examples(i18n.T(`
	# Create a new docker config.
	kops create secret dockerconfig -f /path/to/docker/config.json \
		--name k8s-cluster.example.com --state s3://example.com
	# Create a docker config via stdin.
	generate-docker-config.sh | kops create secret dockerconfig -f - \
		--name k8s-cluster.example.com --state s3://example.com
	# Replace an existing docker config secret.
	kops create secret dockerconfig -f /path/to/docker/config.json --force \
		--name k8s-cluster.example.com --state s3://example.com
	`))

	createSecretDockerconfigShort = i18n.T(`Create a docker config.`)
)

type CreateSecretDockerConfigOptions struct {
	ClusterName      string
	DockerConfigPath string
	Force            bool
}

func NewCmdCreateSecretDockerConfig(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateSecretDockerConfigOptions{}

	cmd := &cobra.Command{
		Use:     "dockerconfig",
		Short:   createSecretDockerconfigShort,
		Long:    createSecretDockerconfigLong,
		Example: createSecretDockerconfigExample,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.TODO()

			if len(args) != 0 {
				exitWithError(fmt.Errorf("syntax: -f <DockerConfigPath>"))
			}

			err := rootCommand.ProcessArgs(args[0:])
			if err != nil {
				exitWithError(err)
			}

			options.ClusterName = rootCommand.ClusterName()

			err = RunCreateSecretDockerConfig(ctx, f, os.Stdout, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().StringVarP(&options.DockerConfigPath, "", "f", "", "Path to docker config JSON file")
	cmd.Flags().BoolVar(&options.Force, "force", options.Force, "Force replace the kops secret if it already exists")

	return cmd
}

func RunCreateSecretDockerConfig(ctx context.Context, f *util.Factory, out io.Writer, options *CreateSecretDockerConfigOptions) error {
	if options.DockerConfigPath == "" {
		return fmt.Errorf("docker config path is required (use -f)")
	}
	secret, err := fi.CreateSecret()
	if err != nil {
		return fmt.Errorf("error creating docker config secret: %v", err)
	}

	cluster, err := GetCluster(ctx, f, options.ClusterName)
	if err != nil {
		return err
	}

	clientset, err := f.Clientset()
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
			return fmt.Errorf("error reading docker config from stdin: %v", err)
		}
	} else {
		data, err = ioutil.ReadFile(options.DockerConfigPath)
		if err != nil {
			return fmt.Errorf("error reading docker config %v: %v", options.DockerConfigPath, err)
		}
	}

	var parsedData map[string]interface{}
	err = json.Unmarshal(data, &parsedData)
	if err != nil {
		return fmt.Errorf("Unable to parse JSON %v: %v", options.DockerConfigPath, err)
	}

	secret.Data = data

	if !options.Force {
		_, created, err := secretStore.GetOrCreateSecret("dockerconfig", secret)
		if err != nil {
			return fmt.Errorf("error adding dockerconfig secret: %v", err)
		}
		if !created {
			return fmt.Errorf("failed to create the dockerconfig secret as it already exists. The `--force` flag can be passed to replace an existing secret.")
		}
	} else {
		_, err := secretStore.ReplaceSecret("dockerconfig", secret)
		if err != nil {
			return fmt.Errorf("error updating dockerconfig secret: %v", err)
		}
	}

	return nil
}
