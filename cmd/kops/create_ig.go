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
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/editor"
	"k8s.io/kubernetes/pkg/util/i18n"
)

type CreateInstanceGroupOptions struct {
	Role    string
	Subnets []string
}

var (
	create_ig_long = templates.LongDesc(i18n.T(`
		Create an instancegroup configuration.  kops has the concept of "instance groups",
		which are a group of similar virutal machines. On AWS, they map to an
		AutoScalingGroup. An ig work either as a Kubernetes master or a node.`))

	create_ig_example = templates.Examples(i18n.T(`

		# Create an instancegroup for the k8s-cluster.example.com cluster.
		kops create ig --name=k8s-cluster.example.com node-example \
		  --role node --subnet my-subnet-name
		`))

	create_ig_short = i18n.T(`Create an instancegroup.`)
)

//  NewCmdCreateInstanceGroup create a new cobra command object for creating a instancegroup.
func NewCmdCreateInstanceGroup(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateInstanceGroupOptions{
		Role: string(api.InstanceGroupRoleNode),
	}

	cmd := &cobra.Command{
		Use:     "instancegroup",
		Aliases: []string{"instancegroups", "ig"},
		Short:   create_ig_short,
		Long:    create_ig_long,
		Example: create_ig_example,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunCreateInstanceGroup(f, cmd, args, os.Stdout, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	// TODO: Create Enum helper - or is there one in k8s already?
	var allRoles []string
	for _, r := range api.AllInstanceGroupRoles {
		allRoles = append(allRoles, string(r))
	}

	cmd.Flags().StringVar(&options.Role, "role", options.Role, "Type of instance group to create ("+strings.Join(allRoles, ",")+")")
	cmd.Flags().StringSliceVar(&options.Subnets, "subnet", options.Subnets, "Subnets in which to create instance group")

	return cmd
}

func RunCreateInstanceGroup(f *util.Factory, cmd *cobra.Command, args []string, out io.Writer, options *CreateInstanceGroupOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("Specify name of instance group to create")
	}
	if len(args) != 1 {
		return fmt.Errorf("Can only create one instance group at a time!")
	}
	groupName := args[0]

	cluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	clientset, err := rootCommand.Clientset()
	if err != nil {
		return err
	}

	channel, err := cloudup.ChannelForCluster(cluster)
	if err != nil {
		return err
	}

	existing, err := clientset.InstanceGroupsFor(cluster).Get(groupName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if existing != nil {
		return fmt.Errorf("instance group %q already exists", groupName)
	}

	// Populate some defaults
	ig := &api.InstanceGroup{}
	ig.ObjectMeta.Name = groupName

	role, ok := api.ParseInstanceGroupRole(options.Role, true)
	if !ok {
		return fmt.Errorf("unknown role %q", options.Role)
	}
	ig.Spec.Role = role

	if len(options.Subnets) == 0 {
		return fmt.Errorf("cannot create instance group without subnets; specify --subnet flag(s)")
	}
	ig.Spec.Subnets = options.Subnets

	ig, err = cloudup.PopulateInstanceGroupSpec(cluster, ig, channel)
	if err != nil {
		return err
	}

	var (
		edit = editor.NewDefaultEditor(editorEnvs)
	)

	raw, err := api.ToVersionedYaml(ig)
	if err != nil {
		return err
	}
	ext := "yaml"

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

	obj, _, err := api.ParseVersionedYaml(edited)
	if err != nil {
		return fmt.Errorf("error parsing yaml: %v", err)
	}
	group, ok := obj.(*api.InstanceGroup)
	if !ok {
		return fmt.Errorf("unexpected object type: %T", obj)
	}

	err = validation.ValidateInstanceGroup(group)
	if err != nil {
		return err
	}

	_, err = clientset.InstanceGroupsFor(cluster).Create(group)
	if err != nil {
		return fmt.Errorf("error storing InstanceGroup: %v", err)
	}

	return nil
}
