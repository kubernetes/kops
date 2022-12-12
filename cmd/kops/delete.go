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
	"k8s.io/klog/v2"
	"k8s.io/kops/cmd/kops/util"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/util/pkg/text"
	"k8s.io/kops/util/pkg/vfs"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

type DeleteOptions struct {
	Filenames []string
	Yes       bool
}

var (
	deleteExample = templates.Examples(i18n.T(`
		# Delete a cluster using a manifest file
		kops delete -f my-cluster.yaml

		# Delete a cluster using a pasted manifest file from stdin.
		pbpaste | kops delete -f -
	`))

	deleteShort = i18n.T("Delete clusters, instancegroups, instances, and secrets.")
)

func NewCmdDelete(f *util.Factory, out io.Writer) *cobra.Command {
	options := &DeleteOptions{}

	cmd := &cobra.Command{
		Use:        "delete {-f FILENAME}...",
		Short:      deleteShort,
		Example:    deleteExample,
		SuggestFor: []string{"rm"},
		Args:       cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunDelete(context.TODO(), f, out, options)
		},
	}

	cmd.Flags().StringSliceVarP(&options.Filenames, "filename", "f", options.Filenames, "Filename to use to delete the resource")
	cmd.Flags().BoolVarP(&options.Yes, "yes", "y", options.Yes, "Specify --yes to immediately delete the resource")
	cmd.MarkFlagRequired("filename")

	// create subcommands
	cmd.AddCommand(NewCmdDeleteCluster(f, out))
	cmd.AddCommand(NewCmdDeleteInstance(f, out))
	cmd.AddCommand(NewCmdDeleteInstanceGroup(f, out))
	cmd.AddCommand(NewCmdDeleteSecret(f, out))
	cmd.AddCommand(NewCmdDeleteSSHPublicKey(f, out))

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
				return fmt.Errorf("reading from stdin: %v", err)
			}
		} else {
			contents, err = vfs.FromContext(ctx).ReadFile(f)
			if err != nil {
				return fmt.Errorf("reading file %q: %v", f, err)
			}
		}

		sections := text.SplitContentToSections(contents)
		for _, section := range sections {
			o, gvk, err := kopscodecs.Decode(section, nil)
			if err != nil {
				return fmt.Errorf("parsing file %q: %v", f, err)
			}

			switch v := o.(type) {
			case *kopsapi.Cluster:
				options := &DeleteClusterOptions{
					ClusterName: v.ObjectMeta.Name,
					Yes:         d.Yes,
				}
				err = RunDeleteCluster(ctx, factory, out, options)
				if err != nil {
					return err
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
					return err
				}
			case *kopsapi.SSHCredential:
				options := &DeleteSSHPublicKeyOptions{
					ClusterName: v.ObjectMeta.Labels[kopsapi.LabelClusterName],
				}

				err = RunDeleteSSHPublicKey(ctx, factory, out, options)
				if err != nil {
					return err
				}
			default:
				klog.V(2).Infof("Type of object was %T", v)
				return fmt.Errorf("unhandled kind %q in %s", gvk, f)
			}
		}
	}

	return nil
}
