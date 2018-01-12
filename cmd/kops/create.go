/*
Copyright 2014 The Kubernetes Authors.

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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kops/cmd/kops/util"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/v1alpha1"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/util/pkg/vfs"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
)

type CreateOptions struct {
	resource.FilenameOptions
}

func NewCmdCreate(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateOptions{}

	cmd := &cobra.Command{
		Use:   "create -f FILENAME",
		Short: "Create a resource by filename or stdin.",
		Run: func(cmd *cobra.Command, args []string) {
			if cmdutil.IsFilenameEmpty(options.Filenames) {
				cmd.Help()
				return
			}
			//cmdutil.CheckErr(ValidateArgs(cmd, args))
			//cmdutil.CheckErr(cmdutil.ValidateOutputArgs(cmd))
			cmdutil.CheckErr(RunCreate(f, out, options))
		},
	}

	cmd.Flags().StringSliceVarP(&options.Filenames, "filename", "f", options.Filenames, "Filename to use to create the resource")
	//usage := "to use to create the resource"
	//cmdutil.AddFilenameOptionFlags(cmd, options, usage)
	cmd.MarkFlagRequired("filename")
	//cmdutil.AddValidateFlags(cmd)
	//cmdutil.AddOutputFlagsForMutation(cmd)
	//cmdutil.AddApplyAnnotationFlags(cmd)
	//cmdutil.AddRecordFlag(cmd)
	//cmdutil.AddInclude3rdPartyFlags(cmd)

	// create subcommands
	cmd.AddCommand(NewCmdCreateCluster(f, out))
	cmd.AddCommand(NewCmdCreateInstanceGroup(f, out))
	cmd.AddCommand(NewCmdCreateSecret(f, out))
	return cmd
}

func RunCreate(f *util.Factory, out io.Writer, c *CreateOptions) error {
	clientset, err := f.Clientset()
	if err != nil {
		return err
	}

	// Codecs provides access to encoding and decoding for the scheme
	codecs := kopsapi.Codecs //serializer.NewCodecFactory(scheme)

	codec := codecs.UniversalDecoder(kopsapi.SchemeGroupVersion)

	var clusterName = ""
	//var cSpec = false
	var sb bytes.Buffer
	fmt.Fprintf(&sb, "\n")
	for _, f := range c.Filenames {
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
			case *kopsapi.Federation:
				_, err = clientset.Federations().Create(v)
				if err != nil {
					if apierrors.IsAlreadyExists(err) {
						return fmt.Errorf("federation %q already exists", v.ObjectMeta.Name)
					}
					return fmt.Errorf("error creating federation: %v", err)
				}
				fmt.Fprintf(&sb, "Created federation/%q\n", v.ObjectMeta.Name)

			case *kopsapi.Cluster:
				// Adding a PerformAssignments() call here as the user might be trying to use
				// the new `-f` feature, with an old cluster definition.
				err = cloudup.PerformAssignments(v)
				if err != nil {
					return fmt.Errorf("error populating configuration: %v", err)
				}
				_, err = clientset.Clusters().Create(v)
				if err != nil {
					if apierrors.IsAlreadyExists(err) {
						return fmt.Errorf("cluster %q already exists", v.ObjectMeta.Name)
					}
					return fmt.Errorf("error creating cluster: %v", err)
				} else {
					fmt.Fprintf(&sb, "Created cluster/%s\n", v.ObjectMeta.Name)
					//cSpec = true
				}

			case *kopsapi.InstanceGroup:
				clusterName = v.ObjectMeta.Labels[kopsapi.LabelClusterName]
				if clusterName == "" {
					return fmt.Errorf("must specify %q label with cluster name to create instanceGroup", kopsapi.LabelClusterName)
				}
				_, err = clientset.InstanceGroups(clusterName).Create(v)
				if err != nil {
					if apierrors.IsAlreadyExists(err) {
						return fmt.Errorf("instanceGroup %q already exists", v.ObjectMeta.Name)
					}
					return fmt.Errorf("error creating instanceGroup: %v", err)
				} else {
					fmt.Fprintf(&sb, "Created instancegroup/%s\n", v.ObjectMeta.Name)
				}

			default:
				glog.V(2).Infof("Type of object was %T", v)
				return fmt.Errorf("Unhandled kind %q in %s", gvk, f)
			}
		}

	}
	{
		// If there is a value in this sb, this should mean that we have something to deploy
		// so let's advise the user how to engage the cloud provider and deploy
		if sb.String() != "" {
			fmt.Fprintf(&sb, "\n")
			fmt.Fprintf(&sb, "To deploy these resources, run: kops update cluster %s --yes\n", clusterName)
			fmt.Fprintf(&sb, "\n")
		}
		_, err := out.Write(sb.Bytes())
		if err != nil {
			return fmt.Errorf("error writing to output: %v", err)
		}
	}
	return nil
}
