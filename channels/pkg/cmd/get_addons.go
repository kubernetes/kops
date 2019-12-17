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

/*
Copyright 2017 The Kubernetes Authors.

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

package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/channels/pkg/channels"
	"k8s.io/kops/util/pkg/tables"
)

type GetAddonsOptions struct {
}

func NewCmdGetAddons(f Factory, out io.Writer) *cobra.Command {
	var options GetAddonsOptions

	cmd := &cobra.Command{
		Use:     "addons",
		Aliases: []string{"addon"},
		Short:   "get addons",
		Long:    `List or get addons.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunGetAddons(f, out, &options)
		},
	}

	return cmd
}

type addonInfo struct {
	Name      string
	Version   *channels.ChannelVersion
	Namespace *v1.Namespace
}

func RunGetAddons(f Factory, out io.Writer, options *GetAddonsOptions) error {
	k8sClient, err := f.KubernetesClient()
	if err != nil {
		return err
	}

	namespaces, err := k8sClient.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error listing namespaces: %v", err)
	}

	var info []*addonInfo

	for i := range namespaces.Items {
		ns := &namespaces.Items[i]
		addons := channels.FindAddons(ns)
		for name, version := range addons {
			i := &addonInfo{
				Name:      name,
				Version:   version,
				Namespace: ns,
			}
			info = append(info, i)
		}
	}

	if len(info) == 0 {
		fmt.Printf("\nNo managed addons found\n")
		return nil
	}

	{
		t := &tables.Table{}
		t.AddColumn("NAME", func(r *addonInfo) string {
			return r.Name
		})
		t.AddColumn("NAMESPACE", func(r *addonInfo) string {
			return r.Namespace.Name
		})
		t.AddColumn("VERSION", func(r *addonInfo) string {
			if r.Version == nil {
				return "-"
			}
			if r.Version.Version != nil {
				return *r.Version.Version
			}
			return "?"
		})
		t.AddColumn("CHANNEL", func(r *addonInfo) string {
			if r.Version == nil {
				return "-"
			}
			if r.Version.Channel != nil {
				return *r.Version.Channel
			}
			return "?"
		})

		columns := []string{"NAMESPACE", "NAME", "VERSION", "CHANNEL"}
		err := t.Render(info, os.Stdout, columns...)
		if err != nil {
			return err
		}
	}

	fmt.Printf("\n")

	return nil
}
