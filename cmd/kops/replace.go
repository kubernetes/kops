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

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/util/pkg/vfs"

	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/util/i18n"
)

var (
	replace_long = templates.LongDesc(i18n.T(`
		Replace a resource specification by filename or stdin.`))

	replace_example = templates.Examples(i18n.T(`
		# Replace a cluster specification using a file
		kops replace -f my-cluster.yaml
		`))

	replace_short = i18n.T(`Replace cluster resources.`)
)

type ReplaceOptions struct {
	resource.FilenameOptions
}

func NewCmdReplace(f *util.Factory, out io.Writer) *cobra.Command {
	options := &ReplaceOptions{}

	cmd := &cobra.Command{
		Use:     "replace -f FILENAME",
		Short:   replace_short,
		Long:    replace_long,
		Example: replace_example,
		Run: func(cmd *cobra.Command, args []string) {
			if cmdutil.IsFilenameEmpty(options.Filenames) {
				cmd.Help()
				return
			}

			cmdutil.CheckErr(RunReplace(f, cmd, out, options))
		},
	}

	cmd.Flags().StringSliceVarP(&options.Filenames, "filename", "f", options.Filenames, "A list of one or more files separated by a comma.")
	cmd.MarkFlagRequired("filename")

	return cmd
}

func RunReplace(f *util.Factory, cmd *cobra.Command, out io.Writer, c *ReplaceOptions) error {
	clientset, err := f.Clientset()
	if err != nil {
		return err
	}

	// Codecs provides access to encoding and decoding for the scheme
	codecs := kopsapi.Codecs //serializer.NewCodecFactory(scheme)

	codec := codecs.UniversalDecoder(kopsapi.SchemeGroupVersion)

	for _, f := range c.Filenames {
		contents, err := vfs.Context.ReadFile(f)
		if err != nil {
			return fmt.Errorf("error reading file %q: %v", f, err)
		}
		sections := bytes.Split(contents, []byte("\n---\n"))

		for _, section := range sections {

			o, gvk, err := codec.Decode(section, nil, nil)
			if err != nil {
				return fmt.Errorf("error parsing file %q: %v", f, err)
			}

			switch v := o.(type) {
			case *kopsapi.Federation:
				_, err = clientset.FederationsFor(v).Update(v)
				if err != nil {
					return fmt.Errorf("error replacing federation: %v", err)
				}

			case *kopsapi.Cluster:
				_, err = clientset.ClustersFor(v).Update(v)
				if err != nil {
					return fmt.Errorf("error replacing cluster: %v", err)
				}

			case *kopsapi.InstanceGroup:
				clusterName := v.ObjectMeta.Labels[kopsapi.LabelClusterName]
				if clusterName == "" {
					return fmt.Errorf("must specify %q label with cluster name to replace instanceGroup", kopsapi.LabelClusterName)
				}
				cluster, err := clientset.GetCluster(clusterName)
				if err != nil {
					return fmt.Errorf("error fetching cluster %q: %v", clusterName, err)
				}
				_, err = clientset.InstanceGroupsFor(cluster).Update(v)
				if err != nil {
					return fmt.Errorf("error replacing instanceGroup: %v", err)
				}

			default:
				glog.V(2).Infof("Type of object was %T", v)
				return fmt.Errorf("Unhandled kind %q in %q", gvk, f)
			}

		}
	}

	return nil
}
