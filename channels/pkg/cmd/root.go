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

package cmd

import (
	goflag "flag"
	"fmt"

	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type CmdRootOptions struct {
	configFile string
}

func Execute(f Factory, out io.Writer) error {
	cobra.OnInitialize(initConfig)

	cmd := NewCmdRoot(f, out)

	goflag.Set("logtostderr", "true")
	goflag.CommandLine.Parse([]string{})
	return cmd.Execute()
}

func NewCmdRoot(f Factory, out io.Writer) *cobra.Command {
	options := &CmdRootOptions{}

	cmd := &cobra.Command{
		Use:   "channels",
		Short: "channels applies software from a channel",
	}

	cmd.PersistentFlags().AddGoFlagSet(goflag.CommandLine)

	cmd.PersistentFlags().StringVar(&options.configFile, "config", "", "config file (default is $HOME/.channels.yaml)")

	// create subcommands
	cmd.AddCommand(NewCmdApply(f, out))
	cmd.AddCommand(NewCmdGet(f, out))

	return cmd
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigName(".channels") // name of config file (without extension)
	viper.AddConfigPath("$HOME")     // adding home directory as first search path
	viper.AutomaticEnv()             // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
