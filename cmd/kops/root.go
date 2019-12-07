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
	"bytes"
	goflag "flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog"
	"k8s.io/kops/cmd/kops/util"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/upup/pkg/kutil"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

const (
	validResources = `

	* cluster
	* instancegroup
	* secret

	`
)

var (
	rootLong = templates.LongDesc(i18n.T(`
	kops is Kubernetes ops.

	kops is the easiest way to get a production grade Kubernetes cluster up and running.
	We like to think of it as kubectl for clusters.

	kops helps you create, destroy, upgrade and maintain production-grade, highly available,
	Kubernetes clusters from the command line. AWS (Amazon Web Services) is currently
	officially supported, with GCE and VMware vSphere in alpha support.
	`))

	rootShort = i18n.T(`kops is Kubernetes ops.`)
)

type Factory interface {
	Clientset() (simple.Clientset, error)
}

type RootCmd struct {
	util.FactoryOptions

	factory *util.Factory

	configFile string

	clusterName string

	cobraCommand *cobra.Command
}

var _ Factory = &RootCmd{}

var rootCommand = RootCmd{
	cobraCommand: &cobra.Command{
		Use:   "kops",
		Short: rootShort,
		Long:  rootLong,
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

	klog.InitFlags(nil)

	factory := util.NewFactory(&rootCommand.FactoryOptions)
	rootCommand.factory = factory

	NewCmdRoot(factory, os.Stdout)
}

func NewCmdRoot(f *util.Factory, out io.Writer) *cobra.Command {

	cmd := rootCommand.cobraCommand

	//cmd.PersistentFlags().AddGoFlagSet(goflag.CommandLine)
	goflag.CommandLine.VisitAll(func(goflag *goflag.Flag) {
		switch goflag.Name {
		case "cloud-provider-gce-lb-src-cidrs":
			// Skip; this is dragged in by the google cloudprovider dependency

		default:
			cmd.PersistentFlags().AddGoFlag(goflag)
		}
	})

	cmd.PersistentFlags().StringVar(&rootCommand.configFile, "config", "", "yaml config file (default is $HOME/.kops.yaml)")
	viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config"))
	viper.SetDefault("config", "$HOME/.kops.yaml")

	cmd.PersistentFlags().StringVar(&rootCommand.RegistryPath, "state", "", "Location of state storage (kops 'config' file). Overrides KOPS_STATE_STORE environment variable")
	viper.BindPFlag("KOPS_STATE_STORE", cmd.PersistentFlags().Lookup("state"))
	viper.BindEnv("KOPS_STATE_STORE")

	defaultClusterName := os.Getenv("KOPS_CLUSTER_NAME")
	cmd.PersistentFlags().StringVarP(&rootCommand.clusterName, "name", "", defaultClusterName, "Name of cluster. Overrides KOPS_CLUSTER_NAME environment variable")

	// create subcommands
	cmd.AddCommand(NewCmdCompletion(f, out))
	cmd.AddCommand(NewCmdCreate(f, out))
	cmd.AddCommand(NewCmdDelete(f, out))
	cmd.AddCommand(NewCmdEdit(f, out))
	cmd.AddCommand(NewCmdExport(f, out))
	cmd.AddCommand(NewCmdGet(f, out))
	cmd.AddCommand(NewCmdUpdate(f, out))
	cmd.AddCommand(NewCmdReplace(f, out))
	cmd.AddCommand(NewCmdRollingUpdate(f, out))
	cmd.AddCommand(NewCmdSet(f, out))
	cmd.AddCommand(NewCmdToolbox(f, out))
	cmd.AddCommand(NewCmdValidate(f, out))
	cmd.AddCommand(NewCmdVersion(f, out))

	return cmd
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Config file precedence: --config flag, ${HOME}/.kops.yaml ${HOME}/.kops/config
	configFile := rootCommand.configFile
	if configFile == "" {
		home := homedir.HomeDir()
		configPaths := []string{
			filepath.Join(home, ".kops.yaml"),
			filepath.Join(home, ".kops", "config"),
		}
		for _, p := range configPaths {
			_, err := os.Stat(p)
			if err == nil {
				configFile = p
				break
			} else if !os.IsNotExist(err) {
				klog.V(2).Infof("error checking for file %s: %v", p, err)
			}
		}
	}

	if configFile != "" {
		viper.SetConfigFile(configFile)
		viper.SetConfigType("yaml")

		if err := viper.ReadInConfig(); err != nil {
			klog.Warningf("error reading config: %v", err)
		}
	}

	rootCommand.RegistryPath = viper.GetString("KOPS_STATE_STORE")

	// Tolerate multiple slashes at end
	rootCommand.RegistryPath = strings.TrimSuffix(rootCommand.RegistryPath, "/")
}

func (c *RootCmd) AddCommand(cmd *cobra.Command) {
	c.cobraCommand.AddCommand(cmd)
}

// ProcessArgs will parse the positional args.  It assumes one of these formats:
//  * <no arguments at all>
//  * <clustername> (and --name not specified)
// Everything else is an error.
func (c *RootCmd) ProcessArgs(args []string) error {
	if len(args) == 0 {
		return nil
	}

	if len(args) == 1 {
		// Assume <clustername>
		if c.clusterName == "" {
			c.clusterName = args[0]
			return nil
		}
	}

	fmt.Printf("\nFound multiple arguments which look like a cluster name\n")
	if c.clusterName != "" {
		fmt.Printf("\t%q (via flag)\n", c.clusterName)
	}
	for _, arg := range args {
		fmt.Printf("\t%q (as argument)\n", arg)
	}
	fmt.Printf("\n")
	fmt.Printf("This often happens if you specify an argument to a boolean flag without using =\n")
	fmt.Printf("For example: use `--bastion=true` or `--bastion`, not `--bastion true`\n\n")

	if len(args) == 1 {
		return fmt.Errorf("cannot specify cluster via --name and positional argument")
	}
	return fmt.Errorf("expected a single <clustername> to be passed as an argument")
}

func (c *RootCmd) ClusterName() string {
	if c.clusterName != "" {
		return c.clusterName
	}

	c.clusterName = ClusterNameFromKubecfg()

	return c.clusterName
}

func ClusterNameFromKubecfg() string {
	// Read from kubeconfig
	pathOptions := clientcmd.NewDefaultPathOptions()

	clusterName := ""

	config, err := pathOptions.GetStartingConfig()
	if err != nil {
		klog.Warningf("error reading kubecfg: %v", err)
	} else if config.CurrentContext == "" {
		klog.Warningf("no context set in kubecfg")
	} else {
		context := config.Contexts[config.CurrentContext]
		if context == nil {
			klog.Warningf("context %q in kubecfg not found", config.CurrentContext)
		} else if context.Cluster == "" {
			klog.Warningf("context %q in kubecfg did not have a cluster", config.CurrentContext)
		} else {
			fmt.Fprintf(os.Stderr, "Using cluster from kubectl context: %s\n\n", context.Cluster)
			clusterName = context.Cluster
		}
	}

	//config, err := readKubectlClusterConfig()
	//if err != nil {
	//	klog.Warningf("error reading kubecfg: %v", err)
	//} else if config != nil && config.Name != "" {
	//	fmt.Fprintf(os.Stderr, "Using cluster from kubectl context: %s\n\n", config.Name)
	//	c.clusterName = config.Name
	//}

	return clusterName
}

func (c *RootCmd) Clientset() (simple.Clientset, error) {
	return c.factory.Clientset()
}

func (c *RootCmd) Cluster() (*kopsapi.Cluster, error) {
	clusterName := c.ClusterName()
	if clusterName == "" {
		return nil, fmt.Errorf("--name is required")
	}

	return GetCluster(c.factory, clusterName)
}

func GetCluster(factory Factory, clusterName string) (*kopsapi.Cluster, error) {
	if clusterName == "" {
		return nil, field.Required(field.NewPath("ClusterName"), "Cluster name is required")
	}

	clientset, err := factory.Clientset()
	if err != nil {
		return nil, err
	}

	cluster, err := clientset.GetCluster(clusterName)
	if err != nil {
		return nil, fmt.Errorf("error reading cluster configuration: %v", err)
	}
	if cluster == nil {
		return nil, fmt.Errorf("cluster %q not found", clusterName)
	}

	if clusterName != cluster.ObjectMeta.Name {
		return nil, fmt.Errorf("cluster name did not match expected name: %v vs %v", clusterName, cluster.ObjectMeta.Name)
	}
	return cluster, nil
}

// ConsumeStdin reads all the bytes available from stdin
func ConsumeStdin() ([]byte, error) {
	file := os.Stdin
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(file)
	if err != nil {
		return nil, fmt.Errorf("error reading stdin: %v", err)
	}
	return buf.Bytes(), nil
}
