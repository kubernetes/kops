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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/upup/pkg/kutil"
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

			err = RunUpdateCluster(f, clusterName, out, options)
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
	cmd.Flags().BoolVar(&options.CreateKubecfg, "create-kube-config", options.CreateKubecfg, "Will control automatically creating the kube config file on your local filesystem")
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
		} else if c.Target == cloudup.TargetCloudformation {
			c.OutDir = "out/cloudformation"
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
	}

	var instanceGroups []*kops.InstanceGroup
	{
		list, err := clientset.InstanceGroups(cluster.ObjectMeta.Name).List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		for i := range list.Items {
			instanceGroups = append(instanceGroups, &list.Items[i])
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
		InstanceGroups:  instanceGroups,
	}

	err = applyCmd.Run()
	if err != nil {
		return err
	}

	if isDryrun {
		target := applyCmd.Target.(*fi.DryRunTarget)
		if target.HasChanges() {
			fmt.Fprintf(out, "Must specify --yes to apply changes\n")
		} else {
			fmt.Fprintf(out, "No changes need to be applied\n")
		}
		return nil
	}

	firstRun := false

	if !isDryrun && c.CreateKubecfg {
		hasKubecfg, err := hasKubecfg(cluster.ObjectMeta.Name)
		if err != nil {
			glog.Warningf("error reading kubecfg: %v", err)
			hasKubecfg = true
		}
		firstRun = !hasKubecfg

		kubecfgCert, err := keyStore.FindCert("kubecfg")
		if err != nil {
			// This is only a convenience; don't error because of it
			glog.Warningf("Ignoring error trying to fetch kubecfg cert - won't export kubecfg: %v", err)
			kubecfgCert = nil
		}
		if kubecfgCert != nil {
			glog.Infof("Exporting kubecfg for cluster")
			conf, err := kubeconfig.BuildKubecfg(cluster, keyStore, secretStore)
			if err != nil {
				return err
			}
			err = conf.WriteKubecfg()
			if err != nil {
				return err
			}
		} else {
			glog.Infof("kubecfg cert not found; won't export kubecfg")
		}
	}

	if !isDryrun {
		sb := new(bytes.Buffer)

		if c.Target == cloudup.TargetTerraform {
			fmt.Fprintf(sb, "\n")
			fmt.Fprintf(sb, "Terraform output has been placed into %s\n", c.OutDir)

			if firstRun {
				fmt.Fprintf(sb, "Run these commands to apply the configuration:\n")
				fmt.Fprintf(sb, "   cd %s\n", c.OutDir)
				fmt.Fprintf(sb, "   terraform plan\n")
				fmt.Fprintf(sb, "   terraform apply\n")
				fmt.Fprintf(sb, "\n")
			}
		} else if c.Target == cloudup.TargetCloudformation {
			fmt.Fprintf(sb, "\n")
			fmt.Fprintf(sb, "Cloudformation output has been placed into %s\n", c.OutDir)

			if firstRun {
				cfName := "kubernetes-" + strings.Replace(clusterName, ".", "-", -1)
				cfPath := filepath.Join(c.OutDir, "kubernetes.json")
				fmt.Fprintf(sb, "Run this command to apply the configuration:\n")
				fmt.Fprintf(sb, "   aws cloudformation create-stack --capabilities CAPABILITY_NAMED_IAM --stack-name %s --template-body file://%s\n", cfName, cfPath)
				fmt.Fprintf(sb, "\n")
			}
		} else if firstRun {
			fmt.Fprintf(sb, "\n")
			fmt.Fprintf(sb, "Cluster is starting.  It should be ready in a few minutes.\n")
			fmt.Fprintf(sb, "\n")
		} else {
			// TODO: Different message if no changes were needed
			fmt.Fprintf(sb, "\n")
			fmt.Fprintf(sb, "Cluster changes have been applied to the cloud.\n")
			fmt.Fprintf(sb, "\n")
		}

		// More suggestions on first run
		if firstRun {
			fmt.Fprintf(sb, "Suggestions:\n")
			fmt.Fprintf(sb, " * validate cluster: kops validate cluster\n")
			fmt.Fprintf(sb, " * list nodes: kubectl get nodes --show-labels\n")
			if !usesBastion(instanceGroups) {
				fmt.Fprintf(sb, " * ssh to the master: ssh -i ~/.ssh/id_rsa admin@%s\n", cluster.Spec.MasterPublicName)
			} else {
				bastionPublicName := findBastionPublicName(cluster)
				if bastionPublicName != "" {
					fmt.Fprintf(sb, " * ssh to the bastion: ssh -i ~/.ssh/id_rsa admin@%s\n", bastionPublicName)
				} else {
					fmt.Fprintf(sb, " * to ssh to the bastion, you probably want to configure a bastionPublicName")
				}
			}
			fmt.Fprintf(sb, " * read about installing addons: https://github.com/kubernetes/kops/blob/master/docs/addons.md\n")
			fmt.Fprintf(sb, "\n")
		}

		if !firstRun {
			// TODO: Detect if rolling-update is needed
			fmt.Fprintf(sb, "\n")
			fmt.Fprintf(sb, "Changes may require instances to restart: kops rolling-update cluster\n")
			fmt.Fprintf(sb, "\n")
		}

		_, err := out.Write(sb.Bytes())
		if err != nil {
			return fmt.Errorf("error writing to output: %v", err)
		}
	}

	return nil
}

func usesBastion(instanceGroups []*kops.InstanceGroup) bool {
	for _, ig := range instanceGroups {
		if ig.Spec.Role == kops.InstanceGroupRoleBastion {
			return true
		}
	}

	return false
}

func findBastionPublicName(c *kops.Cluster) string {
	topology := c.Spec.Topology
	if topology == nil {
		return ""
	}
	bastion := topology.Bastion
	if bastion == nil {
		return ""
	}
	return bastion.BastionPublicName
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
