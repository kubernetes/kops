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
	"k8s.io/kops/pkg/apis/kops/v1alpha1"
	k8sapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/editor"
	"k8s.io/kubernetes/pkg/runtime/serializer/json"
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

	var b bytes.Buffer

	yamlSerde := json.NewYAMLSerializer(json.DefaultMetaFactory, k8sapi.Scheme, k8sapi.Scheme)
	encoder := k8sapi.Codecs.EncoderForVersion(yamlSerde, v1alpha1.SchemeGroupVersion)

	if err := encoder.Encode(old, &b); err != nil {
		return fmt.Errorf("error parsing Federation: %v", err)
	}

	raw := b.Bytes()

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

	codec := k8sapi.Codecs.UniversalDecoder(kopsapi.SchemeGroupVersion)

	newObj, _, err := codec.Decode(edited, nil, nil)
	if err != nil {
		return fmt.Errorf("error parsing: %v", err)
	}

	newFed := newObj.(*kopsapi.Federation)
	err = newFed.Validate()
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
