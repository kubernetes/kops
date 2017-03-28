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

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/util/pkg/vfs"

	"bytes"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
)

type ReplaceOptions struct {
	resource.FilenameOptions
}

func NewCmdReplace(f *util.Factory, out io.Writer) *cobra.Command {
	options := &ReplaceOptions{}

	cmd := &cobra.Command{
		Use:   "replace -f FILENAME",
		Short: "Replace a resource by filename or stdin.",
		Run: func(cmd *cobra.Command, args []string) {
			if cmdutil.IsFilenameEmpty(options.Filenames) {
				cmd.Help()
				return
			}

			cmdutil.CheckErr(RunReplace(f, cmd, out, options))
		},
	}

	cmd.Flags().StringSliceVarP(&options.Filenames, "filename", "f", options.Filenames, "Filename to use to replace the resource")
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
				_, err = clientset.Federations().Update(v)
				if err != nil {
					return fmt.Errorf("error replacing federation: %v", err)
				}

			case *kopsapi.Cluster:
				_, err = clientset.Clusters().Update(v)
				if err != nil {
					return fmt.Errorf("error replacing cluster: %v", err)
				}

			case *kopsapi.InstanceGroup:
				clusterName := v.ObjectMeta.Labels[kopsapi.LabelClusterName]
				if clusterName == "" {
					return fmt.Errorf("must specify %q label with cluster name to replace instanceGroup", kopsapi.LabelClusterName)
				}
				_, err = clientset.InstanceGroups(clusterName).Update(v)
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
