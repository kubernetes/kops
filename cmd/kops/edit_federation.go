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
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"io"
	"k8s.io/kops/cmd/kops/util"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/editor"
)

type EditFederationOptions struct {
}

func NewCmdEditFederation(f *util.Factory, out io.Writer) *cobra.Command {
	options := &EditFederationOptions{}

	cmd := &cobra.Command{
		Use:     "federation",
		Aliases: []string{"federations"},
		Short:   "Edit federation",
		Long:    `Edit a federation configuration.`,
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

	old, err := clientset.Federations().Get(name)
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
	_, err = clientset.Federations().Update(newFed)
	if err != nil {
		return err
	}

	return nil
}
