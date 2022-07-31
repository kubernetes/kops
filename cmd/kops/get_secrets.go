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
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/tables"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	getSecretExample = templates.Examples(i18n.T(`
	# List the secrets
	kops get secrets

	# Get the admin static token for a cluster
	kops get secrets admin -oplaintext`))

	getSecretShort = i18n.T(`Get one or many secrets.`)
)

type GetSecretsOptions struct {
	*GetOptions
	Type        string
	SecretNames []string
}

func NewCmdGetSecrets(f *util.Factory, out io.Writer, getOptions *GetOptions) *cobra.Command {
	options := GetSecretsOptions{
		GetOptions: getOptions,
	}
	cmd := &cobra.Command{
		Use:     "secrets [SECRET_NAME]...",
		Aliases: []string{"secret"},
		Short:   getSecretShort,
		Example: getSecretExample,
		Args: func(cmd *cobra.Command, args []string) error {
			options.SecretNames = args
			options.ClusterName = rootCommand.ClusterName(true)

			if options.ClusterName == "" {
				return fmt.Errorf("--name is required")
			}

			return nil
		},
		ValidArgsFunction: completeSecretNames(f),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunGetSecrets(context.TODO(), f, out, &options)
		},
	}

	cmd.Flags().StringVarP(&options.Type, "type", "", "", "Filter by secret type")
	cmd.Flags().MarkHidden("type")
	return cmd
}

func listSecrets(secretStore fi.SecretStore, names []string) ([]string, error) {
	items, err := secretStore.ListSecrets()
	if err != nil {
		return nil, fmt.Errorf("listing secrets %v", err)
	}

	if len(names) != 0 {
		nameSet := sets.NewString(names...)
		var matches []string
		for _, item := range items {
			if nameSet.Has(item) {
				matches = append(matches, item)
			}
		}
		items = matches
	}

	return items, nil
}

func RunGetSecrets(ctx context.Context, f *util.Factory, out io.Writer, options *GetSecretsOptions) error {
	switch strings.ToLower(options.Type) {
	case "", "secret":
	// OK
	case "sshpublickey":
		return fmt.Errorf("use 'kops get sshpublickey' instead")
	case "keypair":
		return fmt.Errorf("use 'kops get keypairs' instead")
	default:
		return fmt.Errorf("unknown secret type %q", options.Type)
	}

	clientset, err := f.KopsClient()
	if err != nil {
		return err
	}

	cluster, err := GetCluster(ctx, f, options.ClusterName)
	if err != nil {
		return err
	}
	secretStore, err := clientset.SecretStore(cluster)
	if err != nil {
		return err
	}

	items, err := listSecrets(secretStore, options.SecretNames)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return fmt.Errorf("No secrets found")
	}
	switch options.Output {

	case OutputTable:
		t := &tables.Table{}
		t.AddColumn("NAME", func(i string) string {
			return i
		})
		return t.Render(items, out, "NAME")

	case OutputYaml:
		return fmt.Errorf("yaml output format is not (currently) supported for secrets")
	case OutputJSON:
		return fmt.Errorf("json output format is not (currently) supported for secrets")
	case "plaintext":
		for _, item := range items {
			var data string
			secret, err := secretStore.FindSecret(item)
			if err != nil {
				return fmt.Errorf("getting secret %q: %v", item, err)
			}
			if secret == nil {
				return fmt.Errorf("cannot find secret %q", item)
			}
			data = string(secret.Data)

			_, err = fmt.Fprintf(out, "%s\n", data)
			if err != nil {
				return fmt.Errorf("writing output: %v", err)
			}
		}
		return nil

	default:
		return fmt.Errorf("unknown output format: %q", options.Output)
	}
}
