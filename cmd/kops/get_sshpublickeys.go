/*
Copyright 2021 The Kubernetes Authors.

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
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/sshcredentials"
	"k8s.io/kops/util/pkg/tables"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"sigs.k8s.io/yaml"
)

var (
	getSSHPublicKeysExample = templates.Examples(i18n.T(`
	# Get the SSH public key
	kops get sshpublickey`))

	getSSHPublicKeysShort = i18n.T(`Get one or many secrets.`)
)

type GetSSHPublicKeysOptions struct {
	*GetOptions
}

func NewCmdGetSSHPublicKeys(f *util.Factory, out io.Writer, getOptions *GetOptions) *cobra.Command {
	options := GetSSHPublicKeysOptions{
		GetOptions: getOptions,
	}
	cmd := &cobra.Command{
		Use:               "sshpublickeys [CLUSTER]",
		Aliases:           []string{"sshpublickey", "ssh"},
		Short:             getSSHPublicKeysShort,
		Example:           getSSHPublicKeysExample,
		Args:              rootCommand.clusterNameArgs(&options.ClusterName),
		ValidArgsFunction: commandutils.CompleteClusterName(f, true, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunGetSSHPublicKeys(context.TODO(), f, out, &options)
		},
	}

	return cmd
}

type SSHKeyItem struct {
	ID        string `json:"id"`
	PublicKey string `json:"publicKey"`
}

func RunGetSSHPublicKeys(ctx context.Context, f *util.Factory, out io.Writer, options *GetSSHPublicKeysOptions) error {
	clientset, err := f.KopsClient()
	if err != nil {
		return err
	}

	cluster, err := clientset.GetCluster(ctx, options.ClusterName)
	if err != nil {
		return err
	}

	sshCredentialStore, err := clientset.SSHCredentialStore(cluster)
	if err != nil {
		return err
	}

	var items []*SSHKeyItem

	l, err := sshCredentialStore.FindSSHPublicKeys()
	if err != nil {
		return fmt.Errorf("listing SSH credentials %v", err)
	}

	for _, key := range l {
		id, err := sshcredentials.Fingerprint(key.Spec.PublicKey)
		if err != nil {
			klog.Warningf("unable to compute fingerprint for public key")
		}
		item := &SSHKeyItem{
			ID:        id,
			PublicKey: key.Spec.PublicKey,
		}

		items = append(items, item)
	}

	switch options.Output {

	case OutputTable:
		if len(items) == 0 {
			return fmt.Errorf("no SSH public key found")
		}
		t := &tables.Table{}
		t.AddColumn("ID", func(i *SSHKeyItem) string {
			return i.ID
		})
		return t.Render(items, out, "ID")

	case OutputYaml:
		y, err := yaml.Marshal(items)
		if err != nil {
			return fmt.Errorf("unable to marshal YAML: %v", err)
		}
		if _, err := out.Write(y); err != nil {
			return fmt.Errorf("error writing to output: %v", err)
		}
	case OutputJSON:
		j, err := json.Marshal(items)
		if err != nil {
			return fmt.Errorf("unable to marshal JSON: %v", err)
		}
		if _, err := out.Write(j); err != nil {
			return fmt.Errorf("error writing to output: %v", err)
		}
	default:
		return fmt.Errorf("unknown output format: %q", options.Output)
	}

	return nil
}
