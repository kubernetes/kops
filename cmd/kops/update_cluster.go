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
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/upup/pkg/kutil"
	"os"
	"strings"
	"k8s.io/kops/pkg/apis/kops"
)

type UpdateClusterOptions struct {
	Yes          bool
	Target       string
	Models       string
	OutDir       string
	SSHPublicKey string
}

func NewCmdUpdateCluster(f *util.Factory, out io.Writer) *cobra.Command {
	options := &UpdateClusterOptions{}

	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Update cluster",
		Long:  `Updates a k8s cluster.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunUpdateCluster(f, cmd, args, os.Stdout, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().BoolVar(&options.Yes, "yes", false, "Actually create cloud resources")
	cmd.Flags().StringVar(&options.Target, "target", "direct", "Target - direct, terraform")
	cmd.Flags().StringVar(&options.Models, "model", strings.Join(cloudup.CloudupModels, ","), "Models to apply (separate multiple models with commas)")
	cmd.Flags().StringVar(&options.SSHPublicKey, "ssh-public-key", "", "SSH public key to use (deprecated: use kops create secret instead)")
	cmd.Flags().StringVar(&options.OutDir, "out", "", "Path to write any local output")

	return cmd
}

func RunUpdateCluster(f *util.Factory, cmd *cobra.Command, args []string, out io.Writer, c *UpdateClusterOptions) error {
	err := rootCommand.ProcessArgs(args)
	if err != nil {
		return err
	}

	isDryrun := false
	targetName := c.Target

	// direct requires --yes (others do not, because they don't do anything!)
	if c.Target == cloudup.TargetDirect {
		if !c.Yes {
			isDryrun = true
			targetName = cloudup.TargetDryRun
		}
	}
	if c.Target == cloudup.TargetDryRun {
		isDryrun = true
		targetName = cloudup.TargetDryRun
	}

	if c.OutDir == "" {
		if c.Target == cloudup.TargetTerraform {
			c.OutDir = "out/terraform"
		} else {
			c.OutDir = "out"
		}
	}

	cluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	keyStore, err := registry.KeyStore(cluster)
	if err != nil {
		return err
	}

	secretStore, err := registry.SecretStore(cluster)
	if err != nil {
		return err
	}

	clientset, err := f.Clientset()
	if err != nil {
		return err
	}

	if c.SSHPublicKey != "" {
		fmt.Fprintf(out, "--ssh-public-key on update is deprecated - please use `kops create secret --name %s sshpublickey admin -i ~/.ssh/id_rsa.pub` instead\n", cluster.Name)

		c.SSHPublicKey = utils.ExpandPath(c.SSHPublicKey)
		authorized, err := ioutil.ReadFile(c.SSHPublicKey)
		if err != nil {
			return fmt.Errorf("error reading SSH key file %q: %v", c.SSHPublicKey, err)
		}
		err = keyStore.AddSSHPublicKey(fi.SecretNameSSHPrimary, authorized)
		if err != nil {
			return fmt.Errorf("error addding SSH public key: %v", err)
		}
	}

	applyCmd := &cloudup.ApplyClusterCmd{
		Cluster:    cluster,
		Models:     strings.Split(c.Models, ","),
		Clientset:  clientset,
		TargetName: targetName,
		OutDir:     c.OutDir,
		DryRun:     isDryrun,
	}
	err = applyCmd.Run()
	if err != nil {
		return err
	}

	if isDryrun {
		target := applyCmd.Target.(*fi.DryRunTarget)
		if target.HasChanges() {
			fmt.Printf("Must specify --yes to apply changes\n")
		} else {
			fmt.Printf("No changes need to be applied\n")
		}
		return nil
	}

	// TODO: Only if not yet set?
	if !isDryrun {
		hasKubecfg, err := hasKubecfg(cluster.Name)
		if err != nil {
			glog.Warningf("error reading kubecfg: %v", err)
			hasKubecfg = true
		}

		kubecfgCert, err := keyStore.FindCert("kubecfg")
		if err != nil {
			// This is only a convenience; don't error because of it
			glog.Warningf("Ignoring error trying to fetch kubecfg cert - won't export kubecfg: %v", err)
			kubecfgCert = nil
		}
		if kubecfgCert != nil {
			glog.Infof("Exporting kubecfg for cluster")
			x := &kutil.CreateKubecfg{
				ContextName:  cluster.Name,
				KeyStore:     keyStore,
				SecretStore:  secretStore,
				KubeMasterIP: cluster.Spec.MasterPublicName,
			}

			err = x.WriteKubecfg()
			if err != nil {
				return err
			}
		} else {
			glog.Infof("kubecfg cert not found; won't export kubecfg")
		}

		if !hasKubecfg {
			// Assume initial creation
			fmt.Printf("\n")
			fmt.Printf("Cluster is starting.  It should be ready in a few minutes.\n")
			fmt.Printf("\n")
			fmt.Printf("Suggestions:\n")
			fmt.Printf(" * list nodes: kubectl get nodes --show-labels\n")
			if cluster.Spec.Topology.Masters == kops.TopologyPublic {
				fmt.Printf(" * ssh to the master: ssh -i ~/.ssh/id_rsa admin@%s\n", cluster.Spec.MasterPublicName)
			}else {
				fmt.Printf(" * ssh to the bastion: ssh -i ~/.ssh/id_rsa admin@%s\n", cluster.Spec.MasterPublicName)
			}
			fmt.Printf(" * read about installing addons: https://github.com/kubernetes/kops/blob/master/docs/addons.md\n")
			fmt.Printf("\n")
		}
	}

	return nil
}

func hasKubecfg(contextName string) (bool, error) {
	kubectl := &kutil.Kubectl{}

	config, err := kubectl.GetConfig(false)
	if err != nil {
		return false, fmt.Errorf("error getting config from kubectl: %v", err)
	}

	for _, context := range config.Contexts {
		if context.Name == contextName {
			return true, nil
		}
	}
	return false, nil
}
