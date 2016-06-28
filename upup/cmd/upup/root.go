package main

import (
	goflag "flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/vfs"
	"strings"
)

type RootCmd struct {
	configFile string

	stateStore    fi.StateStore
	stateLocation string
	clusterName   string

	cobraCommand *cobra.Command
}

var rootCommand = RootCmd{
	cobraCommand: &cobra.Command{
		Use:   "upup",
		Short: "upup manages kubernetes clusters",
		Long: `upup manages kubernetes clusters.
It allows you to create, destroy, upgrade and maintain them.`,
	},
}

func Execute() {
	goflag.CommandLine.Parse([]string{})
	if err := rootCommand.cobraCommand.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	cmd := rootCommand.cobraCommand

	cmd.PersistentFlags().AddGoFlagSet(goflag.CommandLine)

	cmd.PersistentFlags().StringVar(&rootCommand.configFile, "config", "", "config file (default is $HOME/.upup.yaml)")

	defaultStateStore := os.Getenv("KOPS_STATE_STORE")
	cmd.PersistentFlags().StringVarP(&rootCommand.stateLocation, "state", "", defaultStateStore, "Location of state storage")

	cmd.PersistentFlags().StringVarP(&rootCommand.clusterName, "name", "", "", "Name of cluster")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if rootCommand.configFile != "" {
		// enable ability to specify config file via flag
		viper.SetConfigFile(rootCommand.configFile)
	}

	viper.SetConfigName(".upup") // name of config file (without extension)
	viper.AddConfigPath("$HOME") // adding home directory as first search path
	viper.AutomaticEnv()         // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func (c *RootCmd) AddCommand(cmd *cobra.Command) {
	c.cobraCommand.AddCommand(cmd)
}

func (c *RootCmd) StateStore() (fi.StateStore, error) {
	if c.clusterName == "" {
		return nil, fmt.Errorf("--name is required")
	}

	if c.stateStore != nil {
		return c.stateStore, nil
	}
	stateStore, err := c.StateStoreForCluster(c.clusterName)
	if err != nil {
		return nil, err
	}
	c.stateStore = stateStore
	return stateStore, nil
}

func (c *RootCmd) StateStoreForCluster(clusterName string) (fi.StateStore, error) {
	if c.stateLocation == "" {
		return nil, fmt.Errorf("--state is required")
	}
	if clusterName == "" {
		return nil, fmt.Errorf("clusterName is required")
	}

	statePath, err := vfs.Context.BuildVfsPath(c.stateLocation)
	if err != nil {
		return nil, fmt.Errorf("error building state store path: %v", err)
	}

	isDryrun := false
	stateStore, err := fi.NewVFSStateStore(statePath, clusterName, isDryrun)
	if err != nil {
		return nil, fmt.Errorf("error building state store: %v", err)
	}
	return stateStore, nil
}

func (c *RootCmd) ListClusters() ([]string, error) {
	if c.stateLocation == "" {
		return nil, fmt.Errorf("--state is required")
	}

	statePath, err := vfs.Context.BuildVfsPath(c.stateLocation)
	if err != nil {
		return nil, fmt.Errorf("error building state store path: %v", err)
	}

	paths, err := statePath.ReadTree()
	if err != nil {
		return nil, fmt.Errorf("error reading state store: %v", err)
	}

	var keys []string
	for _, p := range paths {
		relativePath, err := vfs.RelativePath(statePath, p)
		if err != nil {
			return nil, err
		}
		if !strings.HasSuffix(relativePath, "/config") {
			continue
		}
		key := strings.TrimSuffix(relativePath, "/config")
		keys = append(keys, key)
	}
	return keys, nil
}

func (c *RootCmd) Secrets() (fi.SecretStore, error) {
	s, err := c.StateStore()
	if err != nil {
		return nil, err
	}
	return s.Secrets(), nil
}

func (c *RootCmd) CA() (fi.CAStore, error) {
	s, err := c.StateStore()
	if err != nil {
		return nil, err
	}
	return s.CA(), nil
}
