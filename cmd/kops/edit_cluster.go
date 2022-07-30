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
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/commands"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/edit"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/pkg/pretty"
	"k8s.io/kops/pkg/try"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	util_editor "k8s.io/kubectl/pkg/cmd/util/editor"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

type EditClusterOptions struct {
	ClusterName string

	// Sets allows setting values directly in the spec.
	Sets []string
	// Unsets allows unsetting values directly in the spec.
	Unsets []string
}

var (
	editClusterLong = pretty.LongDesc(i18n.T(`Edit a cluster configuration.

	This command changes the desired cluster configuration in the registry.

    To set your preferred editor, you can define the EDITOR environment variable.
    When you have done this, kOps will use the editor that you have set.

	kops edit does not update the cloud resources; to apply the changes use ` + pretty.Bash("kops update cluster") + `.`))

	editClusterExample = templates.Examples(i18n.T(`
	# Edit a cluster configuration in AWS.
	kops edit cluster k8s.cluster.site --state=s3://my-state-store
	`))
)

func NewCmdEditCluster(f *util.Factory, out io.Writer) *cobra.Command {
	options := &EditClusterOptions{}

	cmd := &cobra.Command{
		Use:               "cluster [CLUSTER]",
		Short:             i18n.T("Edit cluster."),
		Long:              editClusterLong,
		Example:           editClusterExample,
		Args:              rootCommand.clusterNameArgs(&options.ClusterName),
		ValidArgsFunction: commandutils.CompleteClusterName(f, true, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunEditCluster(context.TODO(), f, out, options)
		},
	}

	if featureflag.SpecOverrideFlag.Enabled() {
		cmd.Flags().StringSliceVar(&options.Sets, "set", options.Sets, "Directly set values in the spec")
		cmd.RegisterFlagCompletionFunc("set", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		})
		cmd.Flags().StringSliceVar(&options.Unsets, "unset", options.Unsets, "Directly unset values in the spec")
		cmd.RegisterFlagCompletionFunc("unset", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		})
	}

	return cmd
}

func RunEditCluster(ctx context.Context, f *util.Factory, out io.Writer, options *EditClusterOptions) error {
	oldCluster, err := GetCluster(ctx, f, options.ClusterName)
	if err != nil {
		return err
	}

	err = oldCluster.FillDefaults()
	if err != nil {
		return err
	}

	clientset, err := f.KopsClient()
	if err != nil {
		return err
	}

	instanceGroups, err := commands.ReadAllInstanceGroups(ctx, clientset, oldCluster)
	if err != nil {
		return err
	}

	if len(options.Unsets)+len(options.Sets) > 0 {
		newCluster := oldCluster.DeepCopy()
		if err := commands.UnsetClusterFields(options.Unsets, newCluster); err != nil {
			return err
		}
		if err := commands.SetClusterFields(options.Sets, newCluster); err != nil {
			return err
		}

		failure, err := updateCluster(ctx, clientset, oldCluster, newCluster, instanceGroups)
		if err != nil {
			return err
		}
		if failure != "" {
			return fmt.Errorf("%s", failure)
		}
		return nil
	}

	editor := util_editor.NewDefaultEditor(commandutils.EditorEnvs)

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
				return preservedFile(fmt.Errorf("%s", "Edit cancelled: no valid changes were saved."), file, out)
			}
		}

		if len(results.file) > 0 {
			try.RemoveFile(results.file)
		}

		if bytes.Equal(stripComments(raw), stripComments(edited)) {
			try.RemoveFile(file)
			fmt.Fprintln(out, "Edit cancelled: no changes made.")
			return nil
		}

		lines, err := hasLines(bytes.NewBuffer(edited))
		if err != nil {
			return preservedFile(err, file, out)
		}
		if !lines {
			try.RemoveFile(file)
			fmt.Fprintln(out, "Edit cancelled: saved file was empty.")
			return nil
		}

		newObj, _, err := kopscodecs.Decode(edited, nil)
		if err != nil {
			return preservedFile(fmt.Errorf("error parsing config: %s", err), file, out)
		}

		newCluster, ok := newObj.(*api.Cluster)
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

		failure, err := updateCluster(ctx, clientset, oldCluster, newCluster, instanceGroups)
		if err != nil {
			return preservedFile(err, file, out)
		}
		if failure != "" {
			results = editResults{
				file: file,
			}
			results.header.addError(failure)
			containsError = true
			continue
		}

		return nil
	}
}

func updateCluster(ctx context.Context, clientset simple.Clientset, oldCluster, newCluster *api.Cluster, instanceGroups []*api.InstanceGroup) (string, error) {
	cloud, err := cloudup.BuildCloud(newCluster)
	if err != nil {
		return "", err
	}

	err = cloudup.PerformAssignments(newCluster, cloud)
	if err != nil {
		return "", fmt.Errorf("error populating configuration: %v", err)
	}

	assetBuilder := assets.NewAssetBuilder(newCluster, false)
	fullCluster, err := cloudup.PopulateClusterSpec(clientset, newCluster, cloud, assetBuilder)
	if err != nil {
		return fmt.Sprintf("error populating cluster spec: %s", err), nil
	}

	err = validation.DeepValidate(fullCluster, instanceGroups, true, cloud)
	if err != nil {
		return fmt.Sprintf("validation failed: %s", err), nil
	}

	// Retrieve the current status of the cluster.  This will eventually be part of the cluster object.
	status, err := cloud.FindClusterStatus(oldCluster)
	if err != nil {
		return "", err
	}

	// Note we perform as much validation as we can, before writing a bad config
	_, err = clientset.UpdateCluster(ctx, newCluster, status)
	return "", err
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
