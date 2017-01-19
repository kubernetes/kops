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
	"github.com/spf13/cobra"
	"io"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/federation"
)

type UpdateFederationOptions struct {
}

func NewCmdUpdateFederation(f *util.Factory, out io.Writer) *cobra.Command {
	options := &UpdateFederationOptions{}

	cmd := &cobra.Command{
		Use:   "federation",
		Short: "Update federation",
		Long:  `Updates a k8s federation.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunUpdateFederation(f, cmd, args, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}

func RunUpdateFederation(factory *util.Factory, cmd *cobra.Command, args []string, out io.Writer, options *UpdateFederationOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("Specify name of Federation to update")
	}
	if len(args) != 1 {
		return fmt.Errorf("Can only update one Federation at a time")
	}

	name := args[0]

	clientset, err := factory.Clientset()
	if err != nil {
		return err
	}

	f, err := clientset.Federations().Get(name)
	if err != nil {
		return fmt.Errorf("error reading federation %q: %v", name, err)
	}

	applyCmd := &federation.ApplyFederationOperation{
		Federation: f,
		KopsClient: clientset,
	}
	err = applyCmd.Run()
	if err != nil {
		return err
	}

	kubecfg, err := applyCmd.FindKubecfg()
	if err != nil {
		return err
	}

	if kubecfg == nil {
		return fmt.Errorf("cannot find configuration for federation")
	}

	err = kubecfg.WriteKubecfg()
	if err != nil {
		return err
	}

	return nil
}
