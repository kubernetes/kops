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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/util/i18n"
)

var (
	replaceLong = templates.LongDesc(i18n.T(`
		Replace a resource specification by filename or stdin.`))

	replaceExample = templates.Examples(i18n.T(`
		# Replace a cluster specification using a file
		kops replace -f my-cluster.yaml

		# Note, if the resource does not exist the command will error, use --force to provision resource
		kops replace -f my-cluster.yaml --force
		`))

	replaceShort = i18n.T(`Replace cluster resources.`)
)

// replaceOptions is the options for the command
type replaceOptions struct {
	// a list of files containing resources
	resource.FilenameOptions
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
			if cmdutil.IsFilenameEmpty(options.Filenames) {
				cmd.Help()
				return
			}

			cmdutil.CheckErr(RunReplace(f, cmd, out, options))
		},
	}
	cmd.Flags().StringSliceVarP(&options.Filenames, "filename", "f", options.Filenames, "A list of one or more files separated by a comma.")
	cmd.Flags().BoolVarP(&options.force, "force", "", false, "Force any changes, which will also create any non-existing respurce (defaults to instancegroups only)")
	cmd.MarkFlagRequired("filename")

	return cmd
}

// RunReplace processes the replace command
func RunReplace(f *util.Factory, cmd *cobra.Command, out io.Writer, c *replaceOptions) error {
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
				if cluster == nil {
					return fmt.Errorf("cluster %q not found", clusterName)
				}
				// check if the instancegroup exists already
				igName := v.ObjectMeta.Name
				ig, err := clientset.InstanceGroupsFor(cluster).Get(igName, metav1.GetOptions{})
				if err != nil {
					return fmt.Errorf("unable to check for instanceGroup: %v", err)
				}
				switch ig {
				case nil:
					if !c.force {
						return fmt.Errorf("instanceGroup: %v does not exist (try adding --force flag)", igName)
					}
					glog.Infof("instanceGroup: %v was not found, creating resource now", igName)
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
			default:
				glog.V(2).Infof("Type of object was %T", v)
				return fmt.Errorf("Unhandled kind %q in %q", gvk, f)
			}
		}
	}

	return nil
}
