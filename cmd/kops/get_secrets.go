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
	"os"

	"strings"

	"github.com/spf13/cobra"
	"io"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/tables"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/util/i18n"
)

var (
	get_secret_long = templates.LongDesc(i18n.T(`
	Display one or many secrets.`))

	get_secret_example = templates.Examples(i18n.T(`
	# Get a secret
	kops get secrets kube -oplaintext

	# Get the admin password for a cluster
	kops get secrets admin -oplaintext`))

	get_secret_short = i18n.T(`Get one or many secrets.`)
)

type GetSecretsOptions struct {
	*GetOptions
	Type string
}

func NewCmdGetSecrets(f *util.Factory, out io.Writer, getOptions *GetOptions) *cobra.Command {
	options := GetSecretsOptions{
		GetOptions: getOptions,
	}
	cmd := &cobra.Command{
		Use:     "secrets",
		Aliases: []string{"secret"},
		Short:   get_secret_short,
		Long:    get_secret_long,
		Example: get_secret_example,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunGetSecrets(&options, args)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().StringVarP(&options.Type, "type", "", "", "Filter by secret type")
	return cmd
}

func listSecrets(keyStore fi.CAStore, secretStore fi.SecretStore, secretType string, names []string) ([]*fi.KeystoreItem, error) {
	var items []*fi.KeystoreItem

	findType := strings.ToLower(secretType)
	switch findType {
	case "":
	// OK
	case "sshpublickey", "keypair", "secret":
	// OK
	default:
		return nil, fmt.Errorf("unknown secret type %q", secretType)
	}

	{
		l, err := keyStore.List()
		if err != nil {
			return nil, fmt.Errorf("error listing CA store items %v", err)
		}

		for _, i := range l {
			if findType != "" && findType != strings.ToLower(i.Type) {
				continue
			}
			items = append(items, i)
		}
	}

	if findType == "" || findType == strings.ToLower(fi.SecretTypeSecret) {
		l, err := secretStore.ListSecrets()
		if err != nil {
			return nil, fmt.Errorf("error listing secrets %v", err)
		}

		for _, id := range l {
			i := &fi.KeystoreItem{
				Name: id,
				Type: fi.SecretTypeSecret,
			}
			if findType != "" && findType != strings.ToLower(i.Type) {
				continue
			}

			items = append(items, i)
		}
	}

	if len(names) != 0 {
		var matches []*fi.KeystoreItem
		for _, arg := range names {
			var found []*fi.KeystoreItem
			for _, i := range items {
				// There may be multiple secrets with the same name (of different type)
				if i.Name == arg {
					found = append(found, i)
				}
			}

			if len(found) == 0 {
				return nil, fmt.Errorf("Secret not found: %q", arg)
			}

			matches = append(matches, found...)
		}
		items = matches
	}

	return items, nil
}

func RunGetSecrets(options *GetSecretsOptions, args []string) error {
	cluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	keyStore, err := registry.KeyStore(cluster)
	if err != nil {
		return err
	}

	secretStore, err := registry.SecretStore(cluster)
	if err != nil {
		return err
	}

	items, err := listSecrets(keyStore, secretStore, options.Type, args)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Fprintf(os.Stderr, "No secrets found\n")

		return nil
	}
	switch options.output {

	case OutputTable:

		t := &tables.Table{}
		t.AddColumn("NAME", func(i *fi.KeystoreItem) string {
			return i.Name
		})
		t.AddColumn("ID", func(i *fi.KeystoreItem) string {
			return i.Id
		})
		t.AddColumn("TYPE", func(i *fi.KeystoreItem) string {
			return i.Type
		})
		return t.Render(items, os.Stdout, "TYPE", "NAME", "ID")

	case OutputYaml:
		return fmt.Errorf("yaml output format is not (currently) supported for secrets")
	case OutputJSON:
		return fmt.Errorf("json output format is not (currently) supported for secrets")
	case "plaintext":
		for _, i := range items {
			var data string
			switch i.Type {
			case fi.SecretTypeSecret:
				secret, err := secretStore.FindSecret(i.Name)
				if err != nil {
					return fmt.Errorf("error getting secret %q: %v", i.Name, err)
				}
				if secret == nil {
					return fmt.Errorf("cannot find secret %q", i.Name)
				}
				data = string(secret.Data)

			default:
				return fmt.Errorf("secret type %v cannot (currently) be exported as plaintext", i.Type)
			}

			_, err := fmt.Fprintf(os.Stdout, "%s\n", data)
			if err != nil {
				return fmt.Errorf("error writing output: %v", err)
			}
		}
		return nil

	default:
		return fmt.Errorf("Unknown output format: %q", options.output)
	}
}
