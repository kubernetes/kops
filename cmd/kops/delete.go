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
	"io"

	"bytes"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kops/cmd/kops/util"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/v1alpha1"
	"k8s.io/kops/util/pkg/vfs"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
)

type DeleteOptions struct {
	resource.FilenameOptions
	Yes bool
}

func NewCmdDelete(f *util.Factory, out io.Writer) *cobra.Command {
	options := &DeleteOptions{}

	cmd := &cobra.Command{
		Use:        "delete -f FILENAME [--yes]",
		Short:      "Delete clusters and instancegroups",
		Long:       `Delete clusters and instancegroups`,
		SuggestFor: []string{"rm"},
		Run: func(cmd *cobra.Command, args []string) {
			if cmdutil.IsFilenameEmpty(options.Filenames) {
				cmd.Help()
				return
			}
			cmdutil.CheckErr(RunDelete(f, out, options))
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

func RunDelete(factory *util.Factory, out io.Writer, d *DeleteOptions) error {
	// Codecs provides access to encoding and decoding for the scheme
	codecs := kopsapi.Codecs //serializer.NewCodecFactory(scheme)

	codec := codecs.UniversalDecoder(kopsapi.SchemeGroupVersion)

	var sb bytes.Buffer
	fmt.Fprintf(&sb, "\n")
	for _, f := range d.Filenames {
		contents, err := vfs.Context.ReadFile(f)
		if err != nil {
			return fmt.Errorf("error reading file %q: %v", f, err)
		}

		sections := bytes.Split(contents, []byte("\n---\n"))
		for _, section := range sections {
			defaults := &schema.GroupVersionKind{
				Group:   v1alpha1.SchemeGroupVersion.Group,
				Version: v1alpha1.SchemeGroupVersion.Version,
			}
			o, gvk, err := codec.Decode(section, defaults, nil)
			if err != nil {
				return fmt.Errorf("error parsing file %q: %v", f, err)
			}

			switch v := o.(type) {
			case *kopsapi.Cluster:
				options := &DeleteClusterOptions{}
				options.ClusterName = v.ObjectMeta.Name
				options.Yes = d.Yes
				err = RunDeleteCluster(factory, out, options)
				if err != nil {
					exitWithError(err)
				}
				if d.Yes {
					fmt.Fprintf(&sb, "Deleted cluster/%s\n", v.ObjectMeta.Name)
				}
			case *kopsapi.InstanceGroup:
				options := &DeleteInstanceGroupOptions{}
				options.GroupName = v.ObjectMeta.Name
				options.ClusterName = v.ObjectMeta.Labels[kopsapi.LabelClusterName]
				options.Yes = d.Yes
				err := RunDeleteInstanceGroup(factory, out, options)
				if err != nil {
					exitWithError(err)
				}
				if d.Yes {
					fmt.Fprintf(&sb, "Deleted instancegroup/%s\n", v.ObjectMeta.Name)
				}
			default:
				glog.V(2).Infof("Type of object was %T", v)
				return fmt.Errorf("Unhandled kind %q in %s", gvk, f)
			}
		}
	}
	{
		_, err := out.Write(sb.Bytes())
		if err != nil {
			return fmt.Errorf("error writing to output: %v", err)
		}
	}

	return nil
}
