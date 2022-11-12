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

package main

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kubectl/pkg/util/i18n"
)

type GetOptions struct {
	ClusterName string
	Output      string
}

const (
	OutputYaml  = "yaml"
	OutputTable = "table"
	OutputJSON  = "json"
)

func NewCmdGet(f *util.Factory, out io.Writer) *cobra.Command {
	options := &GetOptions{
		Output: OutputTable,
	}

	cmd := &cobra.Command{
		Use:        "get",
		SuggestFor: []string{"list"},
		Short:      i18n.T(`Get one or many resources.`),
		Args:       rootCommand.clusterNameArgs(&options.ClusterName),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunGet(context.TODO(), f, out, options)
		},
	}

	cmd.PersistentFlags().StringVarP(&options.Output, "output", "o", options.Output, "output format. One of: table, yaml, json")
	cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{OutputTable, OutputJSON, OutputYaml}, cobra.ShellCompDirectiveNoFileComp
	})

	// create subcommands
	cmd.AddCommand(NewCmdGetAll(f, out, options))
	cmd.AddCommand(NewCmdGetAssets(f, out, options))
	cmd.AddCommand(NewCmdGetCluster(f, out, options))
	cmd.AddCommand(NewCmdGetInstanceGroups(f, out, options))
	cmd.AddCommand(NewCmdGetInstances(f, out, options))
	cmd.AddCommand(NewCmdGetKeypairs(f, out, options))
	cmd.AddCommand(NewCmdGetSecrets(f, out, options))
	cmd.AddCommand(NewCmdGetSSHPublicKeys(f, out, options))

	return cmd
}

func RunGet(ctx context.Context, f commandutils.Factory, out io.Writer, options *GetOptions) error {
	klog.Warning("`kops get [CLUSTER]` is deprecated: use `kops get all [CLUSTER]`")
	return RunGetAll(ctx, f, out, &GetAllOptions{options})
}

func writeYAMLSep(out io.Writer) error {
	_, err := out.Write([]byte("\n---\n\n"))
	if err != nil {
		return fmt.Errorf("error writing to stdout: %v", err)
	}
	return nil
}

type marshalFunc func(obj runtime.Object) ([]byte, error)

func marshalToWriter(obj runtime.Object, marshal marshalFunc, w io.Writer) error {
	b, err := marshal(obj)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	if err != nil {
		return fmt.Errorf("error writing to stdout: %v", err)
	}
	return nil
}

// obj must be a pointer to a marshalable object
func marshalYaml(obj runtime.Object) ([]byte, error) {
	y, err := kopscodecs.ToVersionedYaml(obj)
	if err != nil {
		return nil, fmt.Errorf("error marshaling yaml: %v", err)
	}
	return y, nil
}

// obj must be a pointer to a marshalable object
func marshalJSON(obj runtime.Object) ([]byte, error) {
	j, err := kopscodecs.ToVersionedJSON(obj)
	if err != nil {
		return nil, fmt.Errorf("error marshaling json: %v", err)
	}
	return j, nil
}
