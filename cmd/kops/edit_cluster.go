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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/commands"
	"k8s.io/kops/pkg/edit"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/pkg/try"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	util_editor "k8s.io/kubernetes/pkg/kubectl/cmd/util/editor"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
	"k8s.io/kubernetes/pkg/kubectl/util/templates"
)

type EditClusterOptions struct {
}

var (
	editClusterLong = templates.LongDesc(i18n.T(`Edit a cluster configuration.

	This command changes the desired cluster configuration in the registry.

    	To set your preferred editor, you can define the EDITOR environment variable.
    	When you have done this, kops will use the editor that you have set.

	kops edit does not update the cloud resources, to apply the changes use "kops update cluster".`))

	editClusterExample = templates.Examples(i18n.T(`
		# Edit a cluster configuration in AWS.
		kops edit cluster k8s.cluster.site --state=s3://kops-state-1234
	`))
)

func NewCmdEditCluster(f *util.Factory, out io.Writer) *cobra.Command {
	options := &EditClusterOptions{}

	cmd := &cobra.Command{
		Use:     "cluster",
		Short:   i18n.T("Edit cluster."),
		Long:    editClusterLong,
		Example: editClusterExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunEditCluster(f, cmd, args, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}

func RunEditCluster(f *util.Factory, cmd *cobra.Command, args []string, out io.Writer, options *EditClusterOptions) error {
	err := rootCommand.ProcessArgs(args)
	if err != nil {
		return err
	}

	oldCluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	err = oldCluster.FillDefaults()
	if err != nil {
		return err
	}

	clientset, err := f.Clientset()
	if err != nil {
		return err
	}

	instanceGroups, err := commands.ReadAllInstanceGroups(clientset, oldCluster)
	if err != nil {
		return err
	}

	var (
		editor = util_editor.NewDefaultEditor(editorEnvs)
	)

	ext := "yaml"
	raw, err := kopscodecs.ToVersionedYaml(oldCluster)
	if err != nil {
		return err
	}

	var (
		results = editResults{}
		edited  = []byte{}
		file    string
	)

	containsError := false

	for {
		buf := &bytes.Buffer{}
		results.header.writeTo(buf)
		results.header.flush()

		if !containsError {
			buf.Write(raw)
		} else {
			buf.Write(stripComments(edited))
		}

		// launch the editor
		editedDiff := edited
		edited, file, err = editor.LaunchTempFile(fmt.Sprintf("%s-edit-", filepath.Base(os.Args[0])), ext, buf)
		if err != nil {
			return preservedFile(fmt.Errorf("error launching editor: %v", err), results.file, out)
		}

		if containsError {
			if bytes.Equal(stripComments(editedDiff), stripComments(edited)) {
				return preservedFile(fmt.Errorf("%s", "Edit cancelled, no valid changes were saved."), file, out)
			}
		}

		if len(results.file) > 0 {
			try.RemoveFile(results.file)
		}

		if bytes.Equal(stripComments(raw), stripComments(edited)) {
			try.RemoveFile(file)
			fmt.Fprintln(out, "Edit cancelled, no changes made.")
			return nil
		}

		lines, err := hasLines(bytes.NewBuffer(edited))
		if err != nil {
			return preservedFile(err, file, out)
		}
		if !lines {
			try.RemoveFile(file)
			fmt.Fprintln(out, "Edit cancelled, saved file was empty.")
			return nil
		}

		newObj, _, err := kopscodecs.Decode(edited, nil)
		if err != nil {
			return preservedFile(fmt.Errorf("error parsing config: %s", err), file, out)
		}

		newCluster, ok := newObj.(*kopsapi.Cluster)
		if !ok {
			results = editResults{
				file: file,
			}
			results.header.addError(fmt.Sprintf("object was not of expected type: %T", newObj))
			containsError = true
			continue
		}

		extraFields, err := edit.HasExtraFields(string(edited), newObj)
		if err != nil {
			results = editResults{
				file: file,
			}
			results.header.addError(fmt.Sprintf("error checking for extra fields: %v", err))
			containsError = true
			continue
		}
		if extraFields != "" {
			results = editResults{
				file: file,
			}
			lines := strings.Split(extraFields, "\n")
			for _, line := range lines {
				results.header.addExtraFields(line)
			}
			containsError = true
			continue
		}

		err = cloudup.PerformAssignments(newCluster)
		if err != nil {
			return preservedFile(fmt.Errorf("error populating configuration: %v", err), file, out)
		}

		assetBuilder := assets.NewAssetBuilder(newCluster, "")
		fullCluster, err := cloudup.PopulateClusterSpec(clientset, newCluster, assetBuilder)
		if err != nil {
			results = editResults{
				file: file,
			}
			results.header.addError(fmt.Sprintf("error populating cluster spec: %s", err))
			containsError = true
			continue
		}

		err = validation.DeepValidate(fullCluster, instanceGroups, true)
		if err != nil {
			results = editResults{
				file: file,
			}
			results.header.addError(fmt.Sprintf("validation failed: %s", err))
			containsError = true
			continue
		}

		configBase, err := registry.ConfigBase(newCluster)
		if err != nil {
			return preservedFile(err, file, out)
		}

		// Retrieve the current status of the cluster.  This will eventually be part of the cluster object.
		statusDiscovery := &commands.CloudDiscoveryStatusStore{}
		status, err := statusDiscovery.FindClusterStatus(oldCluster)
		if err != nil {
			return err
		}

		// Note we perform as much validation as we can, before writing a bad config
		_, err = clientset.UpdateCluster(newCluster, status)
		if err != nil {
			return preservedFile(err, file, out)
		}

		err = registry.WriteConfigDeprecated(newCluster, configBase.Join(registry.PathClusterCompleted), fullCluster)
		if err != nil {
			return preservedFile(fmt.Errorf("error writing completed cluster spec: %v", err), file, out)
		}

		return nil
	}
}

type editResults struct {
	header editHeader
	file   string
}

type editHeader struct {
	errors      []string
	extraFields []string
}

func (h *editHeader) addError(err string) {
	h.errors = append(h.errors, err)
}

func (h *editHeader) addExtraFields(line string) {
	h.extraFields = append(h.extraFields, line)
}

func (h *editHeader) flush() {
	h.errors = []string{}
	h.extraFields = []string{}
}

func (h *editHeader) writeTo(w io.Writer) error {
	fmt.Fprint(w, `# Please edit the object below. Lines beginning with a '#' will be ignored,
# and an empty file will abort the edit. If an error occurs while saving this file will be
# reopened with the relevant failures.
#
`)
	for _, error := range h.errors {
		fmt.Fprintf(w, "# %s\n", error)
		fmt.Fprintln(w, "#")
	}
	if len(h.extraFields) != 0 {
		fmt.Fprintf(w, "# Found fields that are not recognized\n")
		for _, l := range h.extraFields {
			fmt.Fprintf(w, "# %s\n", l)
		}
		fmt.Fprintln(w, "#")
	}
	return nil
}

// stripComments is used for dropping comments from a YAML file
func stripComments(file []byte) []byte {
	stripped := []byte{}
	lines := bytes.Split(file, []byte("\n"))
	for i, line := range lines {
		if bytes.HasPrefix(bytes.TrimSpace(line), []byte("#")) {
			continue
		}
		stripped = append(stripped, line...)
		if i < len(lines)-1 {
			stripped = append(stripped, '\n')
		}
	}
	return stripped
}

// hasLines returns true if any line in the provided stream is non empty - has non-whitespace
// characters, or the first non-whitespace character is a '#' indicating a comment. Returns
// any errors encountered reading the stream.
func hasLines(r io.Reader) (bool, error) {
	// TODO: if any files we read have > 64KB lines, we'll need to switch to bytes.ReadLine
	// TODO: probably going to be secrets
	s := bufio.NewScanner(r)
	for s.Scan() {
		if line := strings.TrimSpace(s.Text()); len(line) > 0 && line[0] != '#' {
			return true, nil
		}
	}
	if err := s.Err(); err != nil && err != io.EOF {
		return false, err
	}
	return false, nil
}

// preservedFile writes out a message about the provided file if it exists to the
// provided output stream when an error happens. Used to notify the user where
// their updates were preserved.
func preservedFile(err error, path string, out io.Writer) error {
	if len(path) > 0 {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			fmt.Fprintf(out, "A copy of your changes has been stored to %q\n", path)
		}
	}
	return err
}
