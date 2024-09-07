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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/dump"
	"k8s.io/kops/pkg/resources"
	resourceops "k8s.io/kops/pkg/resources/ops"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	toolboxDumpLong = templates.LongDesc(i18n.T(`
	Displays cluster information.  Includes information about cloud and Kubernetes resources.`))

	toolboxDumpExample = templates.Examples(i18n.T(`
	# Dump cluster information
	kops toolbox dump --name k8s-cluster.example.com
	`))

	toolboxDumpShort = i18n.T(`Dump cluster information`)

	k8sResources = os.Getenv("KOPS_TOOLBOX_DUMP_K8S_RESOURCES")
)

type ToolboxDumpOptions struct {
	Output string

	ClusterName string

	Dir          string
	PrivateKey   string
	SSHUser      string
	MaxNodes     int
	K8sResources bool
}

func (o *ToolboxDumpOptions) InitDefaults() {
	o.Output = OutputYaml
	o.PrivateKey = "~/.ssh/id_rsa"
	o.SSHUser = "ubuntu"
	o.MaxNodes = 500
	o.K8sResources = k8sResources != ""
}

func NewCmdToolboxDump(f commandutils.Factory, out io.Writer) *cobra.Command {
	options := &ToolboxDumpOptions{}
	options.InitDefaults()

	cmd := &cobra.Command{
		Use:               "dump [CLUSTER]",
		Short:             toolboxDumpShort,
		Long:              toolboxDumpLong,
		Example:           toolboxDumpExample,
		Args:              rootCommand.clusterNameArgs(&options.ClusterName),
		ValidArgsFunction: commandutils.CompleteClusterName(f, true, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunToolboxDump(cmd.Context(), f, out, options)
		},
	}

	cmd.Flags().StringVarP(&options.Output, "output", "o", options.Output, "Output format.  One of json or yaml")
	cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "yaml"}, cobra.ShellCompDirectiveNoFileComp
	})

	cmd.Flags().StringVar(&options.Dir, "dir", options.Dir, "Target directory; if specified will collect logs and other information.")
	cmd.MarkFlagDirname("dir")
	cmd.Flags().BoolVar(&options.K8sResources, "k8s-resources", options.K8sResources, "Include k8s resources in the dump")
	cmd.Flags().IntVar(&options.MaxNodes, "max-nodes", options.MaxNodes, "The maximum number of nodes from which to dump logs")
	cmd.Flags().StringVar(&options.PrivateKey, "private-key", options.PrivateKey, "File containing private key to use for SSH access to instances")
	cmd.Flags().StringVar(&options.SSHUser, "ssh-user", options.SSHUser, "The remote user for SSH access to instances")
	cmd.RegisterFlagCompletionFunc("ssh-user", cobra.NoFileCompletions)

	return cmd
}

func RunToolboxDump(ctx context.Context, f commandutils.Factory, out io.Writer, options *ToolboxDumpOptions) error {
	clientset, err := f.KopsClient()
	if err != nil {
		return err
	}

	cluster, err := clientset.GetCluster(ctx, options.ClusterName)
	if err != nil {
		return err
	}

	if cluster == nil {
		return fmt.Errorf("cluster not found %q", options.ClusterName)
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return err
	}

	resourceMap, err := resourceops.ListResources(cloud, cluster)
	if err != nil {
		return err
	}
	d, err := resources.BuildDump(ctx, cloud, resourceMap)
	if err != nil {
		return err
	}

	if options.Dir != "" {
		privateKeyPath := options.PrivateKey
		if strings.HasPrefix(privateKeyPath, "~/") {
			privateKeyPath = filepath.Join(os.Getenv("HOME"), privateKeyPath[2:])
		}
		key, err := os.ReadFile(privateKeyPath)
		if err != nil {
			return fmt.Errorf("reading private key %q: %v", privateKeyPath, err)
		}

		parsedKey, err := ssh.ParseRawPrivateKey(key)
		if err != nil {
			return fmt.Errorf("parsing private key %q: %v", privateKeyPath, err)
		}

		signer, err := ssh.NewSignerFromKey(parsedKey)
		if err != nil {
			return fmt.Errorf("creating signer for private key %q: %v", privateKeyPath, err)
		}

		contextName := cluster.ObjectMeta.Name
		clientGetter := genericclioptions.NewConfigFlags(true)
		clientGetter.Context = &contextName

		var nodes corev1.NodeList

		kubeConfig, err := clientGetter.ToRESTConfig()
		if err != nil {
			klog.Warningf("cannot load kubeconfig settings for %q: %v", contextName, err)
		} else {
			k8sClient, err := kubernetes.NewForConfig(kubeConfig)
			if err != nil {
				klog.Warningf("cannot build kube client for %q: %v", contextName, err)
			} else {

				nodeList, err := k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
				if err != nil {
					klog.Warningf("error listing nodes in cluster: %v", err)
				} else {
					nodes = *nodeList
				}
			}
		}

		err = truncateNodeList(&nodes, options.MaxNodes)
		if err != nil {
			klog.Warningf("not limiting number of nodes dumped: %v", err)
		}

		sshConfig := &ssh.ClientConfig{
			Config: ssh.Config{},
			User:   options.SSHUser,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		keyRing := agent.NewKeyring()
		defer func(keyRing agent.Agent) {
			_ = keyRing.RemoveAll()
		}(keyRing)
		err = keyRing.Add(agent.AddedKey{
			PrivateKey: parsedKey,
		})
		if err != nil {
			return fmt.Errorf("adding key to SSH agent: %w", err)
		}

		// look for a bastion instance and use it if exists
		// Prefer a bastion load balancer if exists
		bastionAddress := ""
		for _, lb := range d.LoadBalancers {
			if strings.Contains(lb.Name, "bastion") && lb.DNSName != "" {
				bastionAddress = lb.DNSName
			}
		}
		if bastionAddress == "" {
			for _, instance := range d.Instances {
				if strings.Contains(instance.Name, "bastion") {
					bastionAddress = instance.PublicAddresses[0]
				}
			}
		}
		dumper := dump.NewLogDumper(bastionAddress, sshConfig, keyRing, options.Dir)

		var additionalIPs []string
		var additionalPrivateIPs []string
		for _, instance := range d.Instances {
			if len(instance.PublicAddresses) != 0 {
				additionalIPs = append(additionalIPs, instance.PublicAddresses[0])
			} else if len(instance.PrivateAddresses) != 0 {
				additionalPrivateIPs = append(additionalPrivateIPs, instance.PrivateAddresses[0])
			} else {
				klog.Warningf("no IP for instance %q", instance.Name)
			}
		}

		if err := dumper.DumpAllNodes(ctx, nodes, options.MaxNodes, additionalIPs, additionalPrivateIPs); err != nil {
			klog.Warningf("error dumping nodes: %v", err)
		}

		if kubeConfig != nil && options.K8sResources {
			dumper, err := dump.NewResourceDumper(kubeConfig, options.Output, options.Dir)
			if err != nil {
				return fmt.Errorf("error creating resource dumper: %w", err)
			}
			if err := dumper.DumpResources(ctx); err != nil {
				klog.Warningf("error dumping resources: %v", err)
			}

			logDumper, err := dump.NewPodLogDumper(kubeConfig, options.Dir)
			if err != nil {
				return fmt.Errorf("error creating pod log dumper: %w", err)
			}
			if err := logDumper.DumpLogs(ctx); err != nil {
				klog.Warningf("error dumping pod logs: %v", err)
			}
		}
	}

	switch options.Output {
	case OutputYaml:
		b, err := kops.ToRawYaml(d)
		if err != nil {
			return fmt.Errorf("error marshaling yaml: %v", err)
		}
		_, err = out.Write(b)
		if err != nil {
			return fmt.Errorf("error writing to stdout: %v", err)
		}
		return nil

	case OutputJSON:
		b, err := json.MarshalIndent(d, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshaling json: %v", err)
		}
		_, err = out.Write(b)
		if err != nil {
			return fmt.Errorf("error writing to stdout: %v", err)
		}
		return nil

	default:
		return fmt.Errorf("unsupported output format: %q", options.Output)
	}
}

func truncateNodeList(nodes *corev1.NodeList, max int) error {
	if max < 0 {
		return errors.New("--max-nodes must be greater than zero")
	}
	// Move control plane nodes to the start of the list and truncate the remainder
	slices.SortFunc[[]corev1.Node](nodes.Items, func(a corev1.Node, e corev1.Node) int {
		if role := util.GetNodeRole(&a); role == "control-plane" || role == "apiserver" {
			return -1
		}
		return 1
	})
	if len(nodes.Items) > max {
		nodes.Items = nodes.Items[:max]
	}
	return nil
}
