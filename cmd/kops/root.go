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
	"context"
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
	"k8s.io/klog/v2"
	"k8s.io/kops/cmd/kops/util"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/commands"
	"k8s.io/kops/pkg/commands/commandutils"
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
	kOps is Kubernetes Operations.

	kOps is the easiest way to get a production grade Kubernetes cluster up and running.
	We like to think of it as kubectl for clusters.

	kOps helps you create, destroy, upgrade and maintain production-grade, highly available,
	Kubernetes clusters from the command line. AWS (Amazon Web Services) is currently
	officially supported, with Digital Ocean and OpenStack in beta support.
	`))

	rootShort = i18n.T(`kOps is Kubernetes Operations.`)
)

type RootCmd struct {
	util.FactoryOptions

	factory *util.Factory

	configFile string

	clusterName string

	cobraCommand *cobra.Command
}

var rootCommand = RootCmd{
	cobraCommand: &cobra.Command{
		Use:   "kops",
		Short: rootShort,
		Long:  rootLong,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmd.SilenceUsage = true
		},
	},
}

func Execute() {
	goflag.Set("logtostderr", "true")
	goflag.CommandLine.Parse([]string{})
	if err := rootCommand.cobraCommand.Execute(); err != nil {
		os.Exit(1)
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

	// cmd.PersistentFlags().AddGoFlagSet(goflag.CommandLine)
	goflag.CommandLine.VisitAll(func(goflag *goflag.Flag) {
		switch goflag.Name {
		case "cloud-provider-gce-lb-src-cidrs":
		case "cloud-provider-gce-l7lb-src-cidrs":
			// Skip; these is dragged in by the google cloudprovider dependency

		// Hide klog flags that just clutter the --help output; they are still supported, we just don't show them
		case "add_dir_header",
			"alsologtostderr",
			"log_backtrace_at",
			"log_dir",
			"log_file",
			"log_file_max_size",
			"logtostderr",
			"one_output",
			"skip_headers",
			"skip_log_headers",
			"stderrthreshold",
			"vmodule":
			// We keep "v" as that flag is generally useful
			cmd.PersistentFlags().AddGoFlag(goflag)
			cmd.PersistentFlags().Lookup(goflag.Name).Hidden = true

		default:
			cmd.PersistentFlags().AddGoFlag(goflag)
		}
	})

	cmd.PersistentFlags().StringVar(&rootCommand.configFile, "config", "", "yaml config file (default is $HOME/.kops.yaml)")
	viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config"))
	viper.SetDefault("config", "$HOME/.kops.yaml")
	cmd.RegisterFlagCompletionFunc("config", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"yaml", "json"}, cobra.ShellCompDirectiveFilterFileExt
	})

	cmd.PersistentFlags().StringVar(&rootCommand.RegistryPath, "state", "", "Location of state storage (kops 'config' file). Overrides KOPS_STATE_STORE environment variable")
	viper.BindPFlag("KOPS_STATE_STORE", cmd.PersistentFlags().Lookup("state"))
	viper.BindEnv("KOPS_STATE_STORE")
	// TODO implement completion against VFS

	defaultClusterName := os.Getenv("KOPS_CLUSTER_NAME")
	cmd.PersistentFlags().StringVarP(&rootCommand.clusterName, "name", "", defaultClusterName, "Name of cluster. Overrides KOPS_CLUSTER_NAME environment variable")
	cmd.RegisterFlagCompletionFunc("name", commandutils.CompleteClusterName(rootCommand.factory, false, false))

	// create subcommands
	cmd.AddCommand(NewCmdCreate(f, out))
	cmd.AddCommand(NewCmdDelete(f, out))
	cmd.AddCommand(NewCmdDistrust(f, out))
	cmd.AddCommand(NewCmdEdit(f, out))
	cmd.AddCommand(NewCmdExport(f, out))
	cmd.AddCommand(NewCmdGenCLIDocs(f, out))
	cmd.AddCommand(NewCmdGet(f, out))
	cmd.AddCommand(commands.NewCmdHelpers(f, out))
	cmd.AddCommand(NewCmdPromote(f, out))
	cmd.AddCommand(NewCmdReplace(f, out))
	cmd.AddCommand(NewCmdRollingUpdate(f, out))
	cmd.AddCommand(NewCmdToolbox(f, out))
	cmd.AddCommand(NewCmdTrust(f, out))
	cmd.AddCommand(NewCmdUpdate(f, out))
	cmd.AddCommand(NewCmdUpgrade(f, out))
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

func (c *RootCmd) clusterNameArgs(clusterName *string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := c.ProcessArgs(args); err != nil {
			return err
		}

		*clusterName = c.ClusterName(true)
		if *clusterName == "" {
			return fmt.Errorf("--name is required")
		}

		return nil
	}
}

func (c *RootCmd) clusterNameArgsNoKubeconfig(clusterName *string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := c.ProcessArgs(args); err != nil {
			return err
		}

		*clusterName = c.clusterName
		if *clusterName == "" {
			return fmt.Errorf("--name is required")
		}

		return nil
	}
}

func (c *RootCmd) clusterNameArgsAllowNoCluster(clusterName *string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := c.ProcessArgs(args); err != nil {
			return err
		}

		*clusterName = c.clusterName
		return nil
	}
}

// ProcessArgs will parse the positional args.  It assumes one of these formats:
//   - <no arguments at all>
//   - <clustername> (and --name not specified)
//
// Everything else is an error.
func (c *RootCmd) ProcessArgs(args []string) error {
	if len(args) > 0 {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "\nClusterName as positional argument is deprecated and will be removed in a future version\n")
		fmt.Fprintf(os.Stderr, "\n")
	}

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

func (c *RootCmd) ClusterName(verbose bool) string {
	if c.clusterName != "" {
		return c.clusterName
	}

	// Read from kubeconfig
	pathOptions := clientcmd.NewDefaultPathOptions()

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
			if verbose {
				fmt.Fprintf(os.Stderr, "Using cluster from kubectl context: %s\n\n", context.Cluster)
			}
			c.clusterName = context.Cluster
		}
	}

	return c.clusterName
}

func GetCluster(ctx context.Context, factory commandutils.Factory, clusterName string) (*kopsapi.Cluster, error) {
	if clusterName == "" {
		return nil, field.Required(field.NewPath("clusterName"), "Cluster name is required")
	}

	clientset, err := factory.KopsClient()
	if err != nil {
		return nil, err
	}

	cluster, err := clientset.GetCluster(ctx, clusterName)
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

func GetClusterNameForCompletionNoKubeconfig(clusterArgs []string) (clusterName string, completions []string, directive cobra.ShellCompDirective) {
	if len(clusterArgs) > 0 {
		return clusterArgs[0], nil, 0
	}

	if rootCommand.clusterName != "" {
		return rootCommand.clusterName, nil, 0
	}

	return "", []string{"--name"}, cobra.ShellCompDirectiveNoFileComp
}

func GetClusterForCompletion(ctx context.Context, factory commandutils.Factory, clusterArgs []string) (cluster *kopsapi.Cluster, clientSet simple.Clientset, completions []string, directive cobra.ShellCompDirective) {
	clusterName := ""

	if len(clusterArgs) > 0 {
		clusterName = clusterArgs[0]
	} else {
		clusterName = rootCommand.ClusterName(false)
	}

	if clusterName == "" {
		return nil, nil, []string{"--name"}, cobra.ShellCompDirectiveNoFileComp
	}

	cluster, err := GetCluster(ctx, factory, clusterName)
	if err != nil {
		completions, directive := commandutils.CompletionError("getting cluster", err)
		return nil, nil, completions, directive
	}

	clientSet, err = factory.KopsClient()
	if err != nil {
		completions, directive := commandutils.CompletionError("getting clientset", err)
		return nil, nil, completions, directive
	}

	return cluster, clientSet, nil, 0
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
