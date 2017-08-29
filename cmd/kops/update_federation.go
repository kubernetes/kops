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

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/federation"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/util/i18n"
)

var (
	update_federation_long = templates.LongDesc(i18n.T(`
	Update federation cluster resources.
	`))

	update_federation_example = templates.Examples(i18n.T(`
	# After cluster has been edited or upgraded, configure it with:
	kops update federation k8s-cluster.example.com --yes --state=s3://kops-state-1234 --yes
	`))

	update_federation_short = i18n.T("Update federation cluster resources.")
)

type UpdateFederationOptions struct {
}

func NewCmdUpdateFederation(f *util.Factory, out io.Writer) *cobra.Command {
	options := &UpdateFederationOptions{}

	cmd := &cobra.Command{
		Use:     "federation",
		Short:   update_federation_short,
		Long:    update_federation_long,
		Example: update_federation_example,
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

	f, err := clientset.GetFederation(name)
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
