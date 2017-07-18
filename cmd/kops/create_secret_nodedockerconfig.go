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
	create_secret_nodedockerconfig_long = templates.LongDesc(i18n.T(`
	Create a new node docker config, and store it in the state store. Use update
	to update it, this command will only create a new entry.`))

	create_secret_nodedockerconfig_example = templates.Examples(i18n.T(`
	# Create an new node docker config.
	kops create secret nodedockerconfig -i /path/to/docker/config.json \
		--name k8s-cluster.example.com --state s3://example.com
	`))

	create_secret_nodedockerconfig_short = i18n.T(`Create a node docker config.`)
)

type CreateSecretDockercfgOptions struct {
	ClusterName   string
	DockerCfgPath string
}

func NewCmdCreateSecretNodeDockerConfig(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateSecretDockercfgOptions{}

	cmd := &cobra.Command{
		Use:     "nodedockercfg",
		Short:   create_secret_nodedockerconfig_short,
		Long:    create_secret_nodedockerconfig_long,
		Example: create_secret_nodedockerconfig_example,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 0 {
				exitWithError(fmt.Errorf("syntax: -i <DockerCfgPath>"))
			}

			err := rootCommand.ProcessArgs(args[0:])
			if err != nil {
				exitWithError(err)
			}

			options.ClusterName = rootCommand.ClusterName()

			err = RunCreateSecretDockerCfg(f, os.Stdout, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().StringVarP(&options.DockerCfgPath, "", "i", "", "Path to node docker config")

	return cmd
}

func RunCreateSecretDockerCfg(f *util.Factory, out io.Writer, options *CreateSecretDockercfgOptions) error {
	if options.DockerCfgPath == "" {
		return fmt.Errorf("docker config path is required (use -i)")
	}
	secret, err := fi.CreateSecret()
	if err != nil {
		return fmt.Errorf("error creating node docker config secret %v: %v", secret, err)
	}

	cluster, err := GetCluster(f, options.ClusterName)
	if err != nil {
		return err
	}

	secretStore, err := registry.SecretStore(cluster)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(options.DockerCfgPath)
	if err != nil {
		return fmt.Errorf("error reading node docker config %v: %v", options.DockerCfgPath, err)
	}

	secret.Data = data

	_, _, err = secretStore.GetOrCreateSecret("nodedockercfg", secret)
	if err != nil {
		return fmt.Errorf("error adding node docker config: %v", err)
	}

	return nil
}
