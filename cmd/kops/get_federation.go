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
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/util/pkg/tables"
	k8sapi "k8s.io/kubernetes/pkg/api"
	"os"
	"strings"
)

type GetFederationOptions struct {
}

func init() {
	var options GetFederationOptions

	cmd := &cobra.Command{
		Use:     "federations",
		Aliases: []string{"federation"},
		Short:   "get federations",
		Long:    `List or get federations.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunGetFederations(&rootCommand, os.Stdout, &options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	getCmd.cobraCommand.AddCommand(cmd)
}

func RunGetFederations(context Factory, out io.Writer, options *GetFederationOptions) error {
	client, err := context.Clientset()
	if err != nil {
		return err
	}

	list, err := client.Federations().List(k8sapi.ListOptions{})
	if err != nil {
		return err
	}

	var federations []*api.Federation
	for i := range list.Items {
		federations = append(federations, &list.Items[i])
	}
	if len(federations) == 0 {
		fmt.Fprintf(out, "No federations found\n")
		return nil
	}

	output := getCmd.output
	if output == OutputTable {
		t := &tables.Table{}
		t.AddColumn("NAME", func(f *api.Federation) string {
			return f.ObjectMeta.Name
		})
		t.AddColumn("CONTROLLERS", func(f *api.Federation) string {
			return strings.Join(f.Spec.Controllers, ",")
		})
		t.AddColumn("MEMBERS", func(f *api.Federation) string {
			return strings.Join(f.Spec.Members, ",")
		})
		return t.Render(federations, out, "NAME", "CONTROLLERS", "MEMBERS")
	} else if output == OutputYaml {
		for _, f := range federations {
			y, err := api.ToYaml(f)
			if err != nil {
				return fmt.Errorf("error marshaling yaml for %q: %v", f.ObjectMeta.Name, err)
			}
			_, err = out.Write(y)
			if err != nil {
				return fmt.Errorf("error writing to output: %v", err)
			}
		}
		return nil
	} else {
		return fmt.Errorf("Unknown output format: %q", output)
	}
}
