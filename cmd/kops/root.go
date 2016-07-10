package main

import (
	goflag "flag"
	"fmt"
	"os"

	"encoding/json"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/vfs"
	"k8s.io/kops/upup/pkg/kutil"
)

type RootCmd struct {
	configFile string

	clusterRegistry *api.ClusterRegistry

	stateLocation string
	clusterName   string

	cobraCommand *cobra.Command
}

var rootCommand = RootCmd{
	cobraCommand: &cobra.Command{
		Use:   "kops",
		Short: "kops is kubernetes ops",
		Long: `kops is kubernetes ops.
It allows you to create, destroy, upgrade and maintain clusters.`,
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

	cmd.PersistentFlags().StringVar(&rootCommand.configFile, "config", "", "config file (default is $HOME/.kops.yaml)")

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

	viper.SetConfigName(".kops") // name of config file (without extension)
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

func (c *RootCmd) ClusterName() string {
	if c.clusterName != "" {
		return c.clusterName
	}

	config, err := readKubectlClusterConfig()
	if err != nil {
		glog.Warningf("error reading kubecfg: %v", err)
	} else if config != nil && config.Name != "" {
		glog.V(2).Infof("Got cluster name from current kubectl context: %s", config.Name)
		c.clusterName = config.Name
	}
	return c.clusterName
}

func readKubectlClusterConfig() (*kubectlClusterWithName, error) {
	kubectl := &kutil.Kubectl{}
	context, err := kubectl.GetCurrentContext()
	if err != nil {
		return nil, fmt.Errorf("error getting current context from kubectl: %v", err)
	}
	glog.V(4).Infof("context = %q", context)

	configString, err := kubectl.GetConfig(true, "json")
	if err != nil {
		return nil, fmt.Errorf("error getting current config from kubectl: %v", err)
	}
	glog.V(8).Infof("config = %q", configString)

	config := &kubectlConfig{}
	err = json.Unmarshal([]byte(configString), config)
	if err != nil {
		return nil, fmt.Errorf("cannot parse current config from kubectl: %v", err)
	}

	if len(config.Clusters) != 1 {
		return nil, fmt.Errorf("expected exactly one cluster in kubectl config, found %d", len(config.Clusters))
	}

	return config.Clusters[0], nil
}

func (c *RootCmd) ClusterRegistry() (*api.ClusterRegistry, error) {
	if c.clusterRegistry != nil {
		return c.clusterRegistry, nil
	}

	if c.stateLocation == "" {
		return nil, fmt.Errorf("--state is required (or export KOPS_STATE_STORE)")
	}

	basePath, err := vfs.Context.BuildVfsPath(c.stateLocation)
	if err != nil {
		return nil, fmt.Errorf("error building state store path for %q: %v", c.stateLocation, err)
	}

	clusterRegistry := api.NewClusterRegistry(basePath)
	c.clusterRegistry = clusterRegistry
	return clusterRegistry, nil
}

func (c *RootCmd) Cluster() (*api.ClusterRegistry, *api.Cluster, error) {
	clusterRegistry, err := c.ClusterRegistry()
	if err != nil {
		return nil, nil, err
	}

	clusterName := c.ClusterName()
	if clusterName == "" {
		return nil, nil, fmt.Errorf("--name is required")
	}

	cluster, err := clusterRegistry.Find(clusterName)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading cluster configuration: %v", err)
	}
	if cluster == nil {
		return nil, nil, fmt.Errorf("cluster %q not found", clusterName)
	}

	if clusterName != cluster.Name {
		return nil, nil, fmt.Errorf("cluster name did not match expected name: %v vs %v", clusterName, cluster.Name)
	}
	return clusterRegistry, cluster, nil
}

func (c *RootCmd) InstanceGroupRegistry() (*api.InstanceGroupRegistry, error) {
	clusterStore, err := c.ClusterRegistry()
	if err != nil {
		return nil, err
	}

	clusterName := c.ClusterName()
	if clusterName == "" {
		return nil, fmt.Errorf("--name is required")
	}

	return clusterStore.InstanceGroups(clusterName)
}

func (c *RootCmd) SecretStore() (fi.SecretStore, error) {
	clusterStore, err := c.ClusterRegistry()
	if err != nil {
		return nil, err
	}

	clusterName := c.ClusterName()
	if clusterName == "" {
		return nil, fmt.Errorf("--name is required")
	}

	return clusterStore.SecretStore(clusterName), nil
}

func (c *RootCmd) KeyStore() (fi.CAStore, error) {
	clusterStore, err := c.ClusterRegistry()
	if err != nil {
		return nil, err
	}

	clusterName := c.ClusterName()
	if clusterName == "" {
		return nil, fmt.Errorf("--name is required")
	}

	return clusterStore.KeyStore(clusterName), nil
}
