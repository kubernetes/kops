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
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/upup/pkg/kutil"
	"os"
	"strings"
	"time"
)

type UpdateClusterOptions struct {
	Yes             bool
	Target          string
	Models          string
	OutDir          string
	SSHPublicKey    string
	MaxTaskDuration time.Duration
	CreateKubecfg   bool
}

func (o *UpdateClusterOptions) InitDefaults() {
	o.Yes = false
	o.Target = "direct"
	o.Models = strings.Join(cloudup.CloudupModels, ",")
	o.SSHPublicKey = ""
	o.OutDir = ""
	o.MaxTaskDuration = cloudup.DefaultMaxTaskDuration
	o.CreateKubecfg = true
}

func NewCmdUpdateCluster(f *util.Factory, out io.Writer) *cobra.Command {
	options := &UpdateClusterOptions{}
	options.InitDefaults()

	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Update cluster",
		Long:  `Updates a k8s cluster.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := rootCommand.ProcessArgs(args)
			if err != nil {
				exitWithError(err)
			}

			clusterName := rootCommand.ClusterName()

			err = RunUpdateCluster(f, clusterName, os.Stdout, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().BoolVar(&options.Yes, "yes", options.Yes, "Actually create cloud resources")
	cmd.Flags().StringVar(&options.Target, "target", options.Target, "Target - direct, terraform")
	cmd.Flags().StringVar(&options.Models, "model", options.Models, "Models to apply (separate multiple models with commas)")
	cmd.Flags().StringVar(&options.SSHPublicKey, "ssh-public-key", options.SSHPublicKey, "SSH public key to use (deprecated: use kops create secret instead)")
	cmd.Flags().StringVar(&options.OutDir, "out", options.OutDir, "Path to write any local output")

	return cmd
}

func RunUpdateCluster(f *util.Factory, clusterName string, out io.Writer, c *UpdateClusterOptions) error {
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

	cluster, err := GetCluster(f, clusterName)
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
		fmt.Fprintf(out, "--ssh-public-key on update is deprecated - please use `kops create secret --name %s sshpublickey admin -i ~/.ssh/id_rsa.pub` instead\n", cluster.ObjectMeta.Name)

		c.SSHPublicKey = utils.ExpandPath(c.SSHPublicKey)
		authorized, err := ioutil.ReadFile(c.SSHPublicKey)
		if err != nil {
			return fmt.Errorf("error reading SSH key file %q: %v", c.SSHPublicKey, err)
		}
		err = keyStore.AddSSHPublicKey(fi.SecretNameSSHPrimary, authorized)
		if err != nil {
			return fmt.Errorf("error addding SSH public key: %v", err)
		}

		glog.Infof("Using SSH public key: %v\n", c.SSHPublicKey)
	} else {
		items, err := keyStore.FindSSHPublicKeys(fi.SecretNameSSHPrimary)
		if err != nil {
			return fmt.Errorf("error finding SSH Key: %v", err)
		}
		if len(items) == 1 {
			c.SSHPublicKey = string(items[0].Data)
		} else if len(items) < 1 {
			glog.Infof("unable to find SSH key in key store")
		} else {
			var strItems []string
			for _, item := range items {
				strItems = append(strItems, string(item.Data))
			}
			c.SSHPublicKey = strings.Join(strItems, "|") //Join on pipe to insinuate choices
		}
	}

	applyCmd := &cloudup.ApplyClusterCmd{
		Cluster:         cluster,
		Models:          strings.Split(c.Models, ","),
		Clientset:       clientset,
		TargetName:      targetName,
		OutDir:          c.OutDir,
		DryRun:          isDryrun,
		MaxTaskDuration: c.MaxTaskDuration,
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
	if !isDryrun && c.CreateKubecfg {
		hasKubecfg, err := hasKubecfg(cluster.ObjectMeta.Name)
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
				ContextName:  cluster.ObjectMeta.Name,
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
			if c.Target == cloudup.TargetTerraform {
				fmt.Printf("\n")
				fmt.Printf("Terraform output has been placed into %s\n", c.OutDir)
				fmt.Printf("Run these commands to apply the configuration:\n")
				fmt.Printf("   cd %s\n", c.OutDir)
				fmt.Printf("   terraform plan\n")
				fmt.Printf("   terraform apply\n")
				fmt.Printf("\n")
			} else {
				fmt.Printf("\n")
				fmt.Printf("Cluster is starting.  It should be ready in a few minutes.\n")
				fmt.Printf("\n")
			}
			fmt.Printf("Suggestions:\n")
			fmt.Printf(" * list nodes: kubectl get nodes --show-labels\n")
			if cluster.Spec.Topology.Masters == kops.TopologyPublic {
				fmt.Printf(" * ssh to the master: ssh -i %s admin@%s\n", c.SSHPublicKey, cluster.Spec.MasterPublicName)
			} else {
				fmt.Printf(" * ssh to the bastion: ssh -i %s admin@%s\n", c.SSHPublicKey, cluster.Spec.MasterPublicName)
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
