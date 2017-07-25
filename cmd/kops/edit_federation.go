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
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/pkg/pretty"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/editor"
	"k8s.io/kubernetes/pkg/util/i18n"
)

var (
	edit_federation_long = pretty.LongDesc(`
		Edit a cluster configuration.

		This command changes the federation cloud specification in the registry.

		To set your preferred editor, you can define the EDITOR environment variable.
		When you have done this, kops will use the editor that you have set.

		kops edit does not update the cloud resources, to apply the changes use ` + pretty.Bash("kops update cluster") + `.`)

	edit_federation_example = templates.Examples(i18n.T(`
		# Edit a cluster federation configuration.
		kops edit federation k8s-cluster.example.com --state=s3://kops-state-1234
		`))

	edit_federation_short = i18n.T(`Edit federation.`)
)

type EditFederationOptions struct {
}

func NewCmdEditFederation(f *util.Factory, out io.Writer) *cobra.Command {
	options := &EditFederationOptions{}

	cmd := &cobra.Command{
		Use:     "federation",
		Aliases: []string{"federations"},
		Short:   edit_federation_short,
		Long:    edit_federation_long,
		Example: edit_federation_example,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunEditFederation(f, cmd, args, os.Stdout, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}

func RunEditFederation(f *util.Factory, cmd *cobra.Command, args []string, out io.Writer, options *EditFederationOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("Specify name of Federation to edit")
	}
	if len(args) != 1 {
		return fmt.Errorf("Can only edit one Federation at a time")
	}

	name := args[0]

	clientset, err := f.Clientset()
	if err != nil {
		return err
	}

	if name == "" {
		return fmt.Errorf("name is required")
	}

	old, err := clientset.GetFederation(name)
	if err != nil {
		return fmt.Errorf("error reading Federation %q: %v", name, err)
	}
	if old == nil {
		return fmt.Errorf("Federation %q not found", name)
	}

	var (
		edit = editor.NewDefaultEditor(editorEnvs)
	)

	ext := "yaml"

	raw, err := kopsapi.ToVersionedYaml(old)
	if err != nil {
		return err
	}

	// launch the editor
	edited, file, err := edit.LaunchTempFile(fmt.Sprintf("%s-edit-", filepath.Base(os.Args[0])), ext, bytes.NewReader(raw))
	defer func() {
		if file != "" {
			os.Remove(file)
		}
	}()
	if err != nil {
		return fmt.Errorf("error launching editor: %v", err)
	}

	if bytes.Equal(edited, raw) {
		fmt.Fprintln(os.Stderr, "Edit cancelled, no changes made.")
		return nil
	}

	newObj, _, err := kopsapi.ParseVersionedYaml(edited)
	if err != nil {
		return fmt.Errorf("error parsing config: %v", err)
	}

	newFed, ok := newObj.(*kopsapi.Federation)
	if !ok {
		return fmt.Errorf("object was not of expected type: %T", newObj)
	}

	err = validation.ValidateFederation(newFed)
	if err != nil {
		return err
	}

	// Note we perform as much validation as we can, before writing a bad config
	_, err = clientset.FederationsFor(newFed).Update(newFed)
	if err != nil {
		return err
	}

	return nil
}
