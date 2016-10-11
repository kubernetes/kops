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
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"io"
	"k8s.io/kops/cmd/kops/util"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/util/pkg/vfs"
	k8sapi "k8s.io/kubernetes/pkg/api"
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
		Short: "Create a resource by filename or stdin",
		Run: func(cmd *cobra.Command, args []string) {
			if cmdutil.IsFilenameEmpty(options.Filenames) {
				cmd.Help()
				return
			}
			//cmdutil.CheckErr(ValidateArgs(cmd, args))
			//cmdutil.CheckErr(cmdutil.ValidateOutputArgs(cmd))
			cmdutil.CheckErr(RunCreate(f, cmd, out, options))
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

func RunCreate(f *util.Factory, cmd *cobra.Command, out io.Writer, c *CreateOptions) error {
	// Codecs provides access to encoding and decoding for the scheme
	codecs := k8sapi.Codecs //serializer.NewCodecFactory(scheme)

	codec := codecs.UniversalDecoder(kopsapi.SchemeGroupVersion)

	for _, f := range c.Filenames {
		contents, err := vfs.Context.ReadFile(f)
		if err != nil {
			return fmt.Errorf("error reading file %q: %v", f, err)
		}

		o, gvk, err := codec.Decode(contents, nil, nil)
		if err != nil {
			return fmt.Errorf("error parsing file %q: %v", f, err)
		}

		switch v := o.(type) {
		default:
			glog.V(2).Infof("Type of object was %T", v)
			return fmt.Errorf("Unhandled kind %q in %q", gvk, f)
		}

	}

	return nil
}
