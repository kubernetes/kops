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
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"k8s.io/kops/cmd/kops/util"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/commands"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/util/pkg/text"
	"k8s.io/kops/util/pkg/vfs"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
	"k8s.io/kubernetes/pkg/kubectl/util/templates"
)

var (
	replaceLong = templates.LongDesc(i18n.T(`
		Replace a resource desired configuration by filename or stdin.`))

	replaceExample = templates.Examples(i18n.T(`
		# Replace a cluster desired configuration using a YAML file
		kops replace -f my-cluster.yaml

		# Replace an instancegroup using YAML passed into stdin.
		cat instancegroup.yaml | kops replace -f -

		# Note, if the resource does not exist the command will error, use --force to provision resource
		kops replace -f my-cluster.yaml --force
		`))

	replaceShort = i18n.T(`Replace cluster resources.`)
)

// replaceOptions is the options for the command
type replaceOptions struct {
	// Filenames is a list of files containing resources
	Filenames []string
	// create any resources not found - we limit to instance groups only for now
	force bool
}

// NewCmdReplace returns a new replace command
func NewCmdReplace(f *util.Factory, out io.Writer) *cobra.Command {
	options := &replaceOptions{}

	cmd := &cobra.Command{
		Use:     "replace -f FILENAME",
		Short:   replaceShort,
		Long:    replaceLong,
		Example: replaceExample,
		Run: func(cmd *cobra.Command, args []string) {
			if len(options.Filenames) == 0 {
				cmd.Help()
				return
			}

			cmdutil.CheckErr(RunReplace(f, cmd, out, options))
		},
	}
	cmd.Flags().StringSliceVarP(&options.Filenames, "filename", "f", options.Filenames, "A list of one or more files separated by a comma.")
	cmd.Flags().BoolVarP(&options.force, "force", "", false, "Force any changes, which will also create any non-existing resource")
	cmd.MarkFlagRequired("filename")

	return cmd
}

// RunReplace processes the replace command
func RunReplace(f *util.Factory, cmd *cobra.Command, out io.Writer, c *replaceOptions) error {
	clientset, err := f.Clientset()
	if err != nil {
		return err
	}

	for _, f := range c.Filenames {
		var contents []byte
		if f == "-" {
			contents, err = ConsumeStdin()
			if err != nil {
				return err
			}
		} else {
			contents, err = vfs.Context.ReadFile(f)
			if err != nil {
				return fmt.Errorf("error reading file %q: %v", f, err)
			}
		}
		sections := text.SplitContentToSections(contents)

		for _, section := range sections {
			o, gvk, err := kopscodecs.Decode(section, nil)
			if err != nil {
				return fmt.Errorf("error parsing file %q: %v", f, err)
			}

			switch v := o.(type) {
			case *kopsapi.Cluster:
				{
					if v.Spec.ExternalCloudControllerManager != nil && !featureflag.EnableExternalCloudController.Enabled() {
						klog.Warningf("Without setting the feature flag `+EnableExternalCloudController` the external cloud controller manager configuration will be discarded")
					}
					// Retrieve the current status of the cluster.  This will eventually be part of the cluster object.
					statusDiscovery := &commands.CloudDiscoveryStatusStore{}
					status, err := statusDiscovery.FindClusterStatus(v)
					if err != nil {
						return err
					}

					// Check if the cluster exists already
					clusterName := v.Name
					cluster, err := clientset.GetCluster(clusterName)
					if err != nil {
						if errors.IsNotFound(err) {
							cluster = nil
						} else {
							return fmt.Errorf("error fetching cluster %q: %v", clusterName, err)
						}
					}
					if cluster == nil {
						if !c.force {
							return fmt.Errorf("cluster %v does not exist (try adding --force flag)", clusterName)
						}
						_, err = clientset.CreateCluster(v)
						if err != nil {
							return fmt.Errorf("error creating cluster: %v", err)
						}
					} else {
						_, err = clientset.UpdateCluster(v, status)
						if err != nil {
							return fmt.Errorf("error replacing cluster: %v", err)
						}
					}
				}

			case *kopsapi.InstanceGroup:
				clusterName := v.ObjectMeta.Labels[kopsapi.LabelClusterName]
				if clusterName == "" {
					return fmt.Errorf("must specify %q label with cluster name to replace instanceGroup", kopsapi.LabelClusterName)
				}
				cluster, err := clientset.GetCluster(clusterName)
				if err != nil {
					if errors.IsNotFound(err) {
						return fmt.Errorf("cluster %q not found", clusterName)
					}
					return fmt.Errorf("error fetching cluster %q: %v", clusterName, err)
				}
				// check if the instancegroup exists already
				igName := v.ObjectMeta.Name
				ig, err := clientset.InstanceGroupsFor(cluster).Get(igName, metav1.GetOptions{})
				if err != nil {
					if errors.IsNotFound(err) {
						if !c.force {
							return fmt.Errorf("instanceGroup: %v does not exist (try adding --force flag)", igName)
						}
					} else {
						return fmt.Errorf("unable to check for instanceGroup: %v", err)
					}
				}
				switch ig {
				case nil:
					klog.Infof("instanceGroup: %v was not found, creating resource now", igName)
					_, err = clientset.InstanceGroupsFor(cluster).Create(v)
					if err != nil {
						return fmt.Errorf("error creating instanceGroup: %v", err)
					}
				default:
					_, err = clientset.InstanceGroupsFor(cluster).Update(v)
					if err != nil {
						return fmt.Errorf("error replacing instanceGroup: %v", err)
					}
				}
			case *kopsapi.SSHCredential:
				clusterName := v.ObjectMeta.Labels[kopsapi.LabelClusterName]
				if clusterName == "" {
					return fmt.Errorf("must specify %q label with cluster name to replace SSHCredential", kopsapi.LabelClusterName)
				}
				if v.Spec.PublicKey == "" {
					return fmt.Errorf("spec.PublicKey is required")
				}

				cluster, err := clientset.GetCluster(clusterName)
				if err != nil {
					return err
				}

				sshCredentialStore, err := clientset.SSHCredentialStore(cluster)
				if err != nil {
					return err
				}

				sshKeyArr := []byte(v.Spec.PublicKey)
				err = sshCredentialStore.AddSSHPublicKey("admin", sshKeyArr)
				if err != nil {
					return fmt.Errorf("error replacing SSHCredential: %v", err)
				}
			default:
				klog.V(2).Infof("Type of object was %T", v)
				return fmt.Errorf("Unhandled kind %q in %q", gvk, f)
			}
		}
	}

	return nil
}
