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
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	"k8s.io/kops/cmd/kops/util"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/pkg/sshcredentials"
	"k8s.io/kops/util/pkg/text"
	"k8s.io/kops/util/pkg/vfs"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

type DeleteOptions struct {
	Filenames []string
	Yes       bool
}

var (
	deleteLong = templates.LongDesc(i18n.T(`
	Delete Kubernetes clusters, instancegroups, and secrets, or a combination of the before mentioned.
	`))

	deleteExample = templates.Examples(i18n.T(`
		# Delete a cluster using a manifest file
		kops delete -f my-cluster.yaml

		# Delete a cluster using a pasted manifest file from stdin.
		pbpaste | kops delete -f -

		# Delete a cluster in AWS.
		kops delete cluster --name=k8s.example.com --state=s3://kops-state-1234

		# Delete an instancegroup for the k8s-cluster.example.com cluster.
		# The --yes option runs the command immediately.
		kops delete ig --name=k8s-cluster.example.com node-example --yes
	`))

	deleteShort = i18n.T("Delete clusters,instancegroups, or secrets.")
)

func NewCmdDelete(f *util.Factory, out io.Writer) *cobra.Command {
	options := &DeleteOptions{}

	cmd := &cobra.Command{
		Use:        "delete -f FILENAME [--yes]",
		Short:      deleteShort,
		Long:       deleteLong,
		Example:    deleteExample,
		SuggestFor: []string{"rm"},
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.TODO()
			if len(options.Filenames) == 0 {
				cmd.Help()
				return
			}
			cmdutil.CheckErr(RunDelete(ctx, f, out, options))
		},
	}

	cmd.Flags().StringSliceVarP(&options.Filenames, "filename", "f", options.Filenames, "Filename to use to delete the resource")
	cmd.Flags().BoolVarP(&options.Yes, "yes", "y", options.Yes, "Specify --yes to delete the resource")
	cmd.MarkFlagRequired("filename")

	// create subcommands
	cmd.AddCommand(NewCmdDeleteCluster(f, out))
	cmd.AddCommand(NewCmdDeleteInstanceGroup(f, out))
	cmd.AddCommand(NewCmdDeleteSecret(f, out))

	return cmd
}

func RunDelete(ctx context.Context, factory *util.Factory, out io.Writer, d *DeleteOptions) error {
	// We could have more than one cluster in a manifest so we are using a set
	deletedClusters := sets.NewString()

	for _, f := range d.Filenames {
		var contents []byte
		var err error
		if f == "-" {
			contents, err = ConsumeStdin()
			if err != nil {
				return fmt.Errorf("error reading from stdin: %v", err)
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
				options := &DeleteClusterOptions{
					ClusterName: v.ObjectMeta.Name,
					Yes:         d.Yes,
				}
				err = RunDeleteCluster(ctx, factory, out, options)
				if err != nil {
					exitWithError(err)
				}
				deletedClusters.Insert(v.ObjectMeta.Name)
			case *kopsapi.InstanceGroup:
				options := &DeleteInstanceGroupOptions{
					GroupName:   v.ObjectMeta.Name,
					ClusterName: v.ObjectMeta.Labels[kopsapi.LabelClusterName],
					Yes:         d.Yes,
				}

				// If the cluster has been already deleted we cannot delete the ig
				if deletedClusters.Has(options.ClusterName) {
					klog.V(4).Infof("Skipping instance group %q because cluster %q has been deleted", v.ObjectMeta.Name, options.ClusterName)
					continue
				}

				err := RunDeleteInstanceGroup(ctx, factory, out, options)
				if err != nil {
					exitWithError(err)
				}
			case *kopsapi.SSHCredential:
				fingerprint, err := sshcredentials.Fingerprint(v.Spec.PublicKey)
				if err != nil {
					klog.Error("unable to compute fingerprint for public key")
				}

				options := &DeleteSecretOptions{
					ClusterName: v.ObjectMeta.Labels[kopsapi.LabelClusterName],
					SecretType:  "SSHPublicKey",
					SecretName:  "admin",
					SecretID:    fingerprint,
				}

				err = RunDeleteSecret(ctx, factory, out, options)
				if err != nil {
					exitWithError(err)
				}
			default:
				klog.V(2).Infof("Type of object was %T", v)
				return fmt.Errorf("Unhandled kind %q in %s", gvk, f)
			}
		}
	}

	return nil
}
