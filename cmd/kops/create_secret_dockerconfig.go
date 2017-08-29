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
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/util/i18n"
)

var (
	create_secret_dockerconfig_long = templates.LongDesc(i18n.T(`
	Create a new docker config, and store it in the state store. 
	Used to configure docker on each master or node (ie. for auth)
	Use update to modify it, this command will only create a new entry.`))

	create_secret_dockerconfig_example = templates.Examples(i18n.T(`
	# Create an new docker config.
	kops create secret dockerconfig -f /path/to/docker/config.json \
		--name k8s-cluster.example.com --state s3://example.com
	`))

	create_secret_dockerconfig_short = i18n.T(`Create a docker config.`)
)

type CreateSecretDockerConfigOptions struct {
	ClusterName      string
	DockerConfigPath string
}

func NewCmdCreateSecretDockerConfig(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateSecretDockerConfigOptions{}

	cmd := &cobra.Command{
		Use:     "dockerconfig",
		Short:   create_secret_dockerconfig_short,
		Long:    create_secret_dockerconfig_long,
		Example: create_secret_dockerconfig_example,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 0 {
				exitWithError(fmt.Errorf("syntax: -f <DockerConfigPath>"))
			}

			err := rootCommand.ProcessArgs(args[0:])
			if err != nil {
				exitWithError(err)
			}

			options.ClusterName = rootCommand.ClusterName()

			err = RunCreateSecretDockerConfig(f, os.Stdout, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().StringVarP(&options.DockerConfigPath, "", "f", "", "Path to docker config JSON file")

	return cmd
}

func RunCreateSecretDockerConfig(f *util.Factory, out io.Writer, options *CreateSecretDockerConfigOptions) error {
	if options.DockerConfigPath == "" {
		return fmt.Errorf("docker config path is required (use -f)")
	}
	secret, err := fi.CreateSecret()
	if err != nil {
		return fmt.Errorf("error creating docker config secret: %v", err)
	}

	cluster, err := GetCluster(f, options.ClusterName)
	if err != nil {
		return err
	}

	secretStore, err := registry.SecretStore(cluster)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(options.DockerConfigPath)
	if err != nil {
		return fmt.Errorf("error reading docker config %v: %v", options.DockerConfigPath, err)
	}

	var parsedData map[string]interface{}
	err = json.Unmarshal(data, &parsedData)
	if err != nil {
		return fmt.Errorf("Unable to parse JSON %v: %v", options.DockerConfigPath, err)
	}

	secret.Data = data

	_, _, err = secretStore.GetOrCreateSecret("dockerconfig", secret)
	if err != nil {
		return fmt.Errorf("error adding docker config secret: %v", err)
	}

	return nil
}
