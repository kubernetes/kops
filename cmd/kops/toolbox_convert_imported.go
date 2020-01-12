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
	"fmt"
	"io"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/kutil"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
	"k8s.io/kubernetes/pkg/kubectl/util/templates"
)

var (
	toolboxConvertImportedLong = templates.LongDesc(i18n.T(`
	Convert an imported cluster into a kops cluster.`))

	toolboxConvertImportedExample = templates.Examples(i18n.T(`

	# Import and convert a cluster
	kops import cluster --name k8s-cluster.example.com --region us-east-1 \
	  --state=s3://k8s-cluster.example.com

	kops toolbox convert-imported k8s-cluster.example.com  \
	  --newname k8s-cluster.example.com
	`))

	toolboxConvertImportedShort = i18n.T(`Convert an imported cluster into a kops cluster.`)
)

type ToolboxConvertImportedOptions struct {
	NewClusterName string

	// Channel is the location of the api.Channel to use for our defaults
	Channel string

	ClusterName string
}

func (o *ToolboxConvertImportedOptions) InitDefaults() {
	o.Channel = api.DefaultChannel
}

func NewCmdToolboxConvertImported(f *util.Factory, out io.Writer) *cobra.Command {
	options := &ToolboxConvertImportedOptions{}
	options.InitDefaults()

	cmd := &cobra.Command{
		Use:     "convert-imported",
		Short:   toolboxConvertImportedShort,
		Long:    toolboxConvertImportedLong,
		Example: toolboxConvertImportedExample,
		Run: func(cmd *cobra.Command, args []string) {
			if err := rootCommand.ProcessArgs(args); err != nil {
				exitWithError(err)
			}

			options.ClusterName = rootCommand.ClusterName()

			err := RunToolboxConvertImported(f, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().StringVar(&options.NewClusterName, "newname", options.NewClusterName, "new cluster name")
	cmd.Flags().StringVar(&options.Channel, "channel", options.Channel, "Channel to use for upgrade")

	return cmd
}

func RunToolboxConvertImported(f *util.Factory, out io.Writer, options *ToolboxConvertImportedOptions) error {
	clientset, err := f.Clientset()
	if err != nil {
		return err
	}

	if options.ClusterName == "" {
		return fmt.Errorf("ClusterName is required")
	}

	cluster, err := clientset.GetCluster(options.ClusterName)
	if err != nil {
		return err
	}

	if cluster == nil {
		return fmt.Errorf("cluster %q not found", options.ClusterName)
	}

	list, err := clientset.InstanceGroupsFor(cluster).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	var instanceGroups []*api.InstanceGroup
	for i := range list.Items {
		instanceGroups = append(instanceGroups, &list.Items[i])
	}

	if cluster.ObjectMeta.Annotations[api.AnnotationNameManagement] != api.AnnotationValueManagementImported {
		return fmt.Errorf("cluster %q does not appear to be a cluster imported using kops import", cluster.ObjectMeta.Name)
	}

	if options.NewClusterName == "" {
		return fmt.Errorf("--newname is required for converting an imported cluster")
	}

	oldClusterName := cluster.ObjectMeta.Name
	if oldClusterName == "" {
		return fmt.Errorf("(Old) ClusterName must be set in configuration")
	}

	// TODO: Switch to cloudup.BuildCloud
	if len(cluster.Spec.Subnets) == 0 {
		return fmt.Errorf("configuration must include Subnets")
	}

	region := ""
	for _, subnet := range cluster.Spec.Subnets {
		if len(subnet.Name) <= 2 {
			return fmt.Errorf("invalid AWS zone: %q", subnet.Zone)
		}

		zoneRegion := subnet.Zone[:len(subnet.Zone)-1]
		if region != "" && zoneRegion != region {
			return fmt.Errorf("clusters cannot span multiple regions")
		}

		region = zoneRegion
	}

	tags := map[string]string{"KubernetesCluster": oldClusterName}
	cloud, err := awsup.NewAWSCloud(region, tags)
	if err != nil {
		return fmt.Errorf("error initializing AWS client: %v", err)
	}

	channel, err := api.LoadChannel(options.Channel)
	if err != nil {
		return err
	}

	d := &kutil.ConvertKubeupCluster{
		NewClusterName: options.NewClusterName,
		OldClusterName: oldClusterName,
		Cloud:          cloud,
		ClusterConfig:  cluster,
		InstanceGroups: instanceGroups,
		Clientset:      clientset,
		Channel:        channel,
	}

	err = d.Upgrade()
	if err != nil {
		return err
	}

	return nil
}
