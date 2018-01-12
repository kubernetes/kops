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
	goflag "flag"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type RootCmd struct {
	configFile string

	cobraCommand *cobra.Command
}

var rootCommand = RootCmd{
	cobraCommand: &cobra.Command{
		Use:   "channels",
		Short: "channels applies software from a channel",
	},
}

func Execute() {
	goflag.Set("logtostderr", "true")
	goflag.CommandLine.Parse([]string{})
	if err := rootCommand.cobraCommand.Execute(); err != nil {
		exitWithError(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	cmd := rootCommand.cobraCommand

	cmd.PersistentFlags().AddGoFlagSet(goflag.CommandLine)

	cmd.PersistentFlags().StringVar(&rootCommand.configFile, "config", "", "config file (default is $HOME/.channels.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if rootCommand.configFile != "" {
		// enable ability to specify config file via flag
		viper.SetConfigFile(rootCommand.configFile)
	}

	viper.SetConfigName(".channels") // name of config file (without extension)
	viper.AddConfigPath("$HOME")     // adding home directory as first search path
	viper.AutomaticEnv()             // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func (c *RootCmd) AddCommand(cmd *cobra.Command) {
	c.cobraCommand.AddCommand(cmd)
}

func (c *RootCmd) KubernetesClient() (kubernetes.Interface, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig

	configOverrides := &clientcmd.ConfigOverrides{
		ClusterDefaults: clientcmd.ClusterDefaults,
	}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("cannot load kubecfg settings: %v", err)
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("cannot build kube client: %v", err)
	}
	return k8sClient, err
}
