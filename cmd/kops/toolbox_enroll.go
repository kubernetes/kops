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
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/model/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/mirrors"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

type ToolboxEnrollOptions struct {
	ClusterName   string
	InstanceGroup string
}

func (o *ToolboxEnrollOptions) InitDefaults() {
}

func NewCmdToolboxEnroll(f commandutils.Factory, out io.Writer) *cobra.Command {
	options := &ToolboxEnrollOptions{}
	options.InitDefaults()

	cmd := &cobra.Command{
		Use:   "enroll [CLUSTER]",
		Short: i18n.T(`Add machine to cluster`),
		Long: templates.LongDesc(i18n.T(`
			Adds an individual machine to the cluster.`)),
		Example: templates.Examples(i18n.T(`
			kops toolbox enroll --name k8s-cluster.example.com
		`)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunToolboxEnroll(cmd.Context(), f, out, options)
		},
	}

	cmd.Flags().StringVar(&options.ClusterName, "cluster", options.ClusterName, "Name of cluster to join")
	cmd.Flags().StringVar(&options.InstanceGroup, "instance-group", options.InstanceGroup, "Name of instance-group to join")

	return cmd
}

func RunToolboxEnroll(ctx context.Context, f commandutils.Factory, out io.Writer, options *ToolboxEnrollOptions) error {
	if options.ClusterName == "" {
		return fmt.Errorf("cluster is required")
	}
	if options.InstanceGroup == "" {
		return fmt.Errorf("instance-group is required")
	}
	clientset, err := f.KopsClient()
	if err != nil {
		return err
	}

	cluster, err := clientset.GetCluster(ctx, options.ClusterName)
	if err != nil {
		return err
	}

	if cluster == nil {
		return fmt.Errorf("cluster not found %q", options.ClusterName)
	}

	ig, err := clientset.InstanceGroupsFor(cluster).Get(ctx, options.InstanceGroup, v1.GetOptions{})
	if err != nil {
		return err
	}

	keyStore, err := clientset.KeyStore(cluster)
	if err != nil {
		return err
	}

	getAssets := false
	assetBuilder := assets.NewAssetBuilder(cluster, getAssets)

	assets := make(map[architectures.Architecture][]*mirrors.MirroredAsset)

	nodeupAssets := make(map[architectures.Architecture]*mirrors.MirroredAsset)
	for _, arch := range architectures.GetSupported() {
		asset, err := cloudup.NodeUpAsset(assetBuilder, arch)
		if err != nil {
			return err
		}
		nodeupAssets[arch] = asset
	}

	encryptionConfigSecretHash := ""

	nodeupConfigBuilder, err := cloudup.NewNodeUpConfigBuilder(cluster, assetBuilder, assets, encryptionConfigSecretHash)
	if err != nil {
		return err
	}

	var apiserverAdditionalIPs []string
	keysets := make(map[string]*fi.Keyset)

	// {
	// 	defaultCA := &fitasks.Keypair{
	// 		Name:      fi.String(fi.CertificateIDCA),
	// 		Lifecycle: fi.LifecycleExistsAndValidates,
	// 		Subject:   "cn=kubernetes-ca",
	// 		Type:      "ca",
	// 	}

	// 	keys[*defaultCA.Name] = defaultCA
	// }
	// 	modelBuilderContext.AddTask(defaultCA)
	// }

	{
		name := "kubernetes-ca"
		keyset, err := keyStore.FindKeyset(name)
		if err != nil {
			return fmt.Errorf("error finding key %q: %w", name, err)
		}
		keysets[name] = keyset
	}
	_, bootConfig, err := nodeupConfigBuilder.BuildConfig(ig, apiserverAdditionalIPs, keysets)
	if err != nil {
		return err
	}

	// configData, err := utils.YamlMarshal(nodeupConfig)
	// if err != nil {
	// 	return fmt.Errorf("error converting nodeup config to yaml: %w", err)
	// }
	// sum256 := sha256.Sum256(configData)

	// fmt.Printf("configData: %s\n", string(configData))

	bootConfigData, err := utils.YamlMarshal(bootConfig)
	if err != nil {
		return fmt.Errorf("error converting boot config to yaml: %w", err)
	}

	var script resources.NodeUpScript
	script.NodeUpAssets = nodeupAssets
	script.KubeEnv = string(bootConfigData)

	resource, err := script.Build()
	if err != nil {
		return fmt.Errorf("error building script: %w", err)
	}

	scriptBytes, err := fi.ResourceAsBytes(resource)
	if err != nil {
		return fmt.Errorf("error generating script: %w", err)
	}

	if _, err := os.Stdout.Write(scriptBytes); err != nil {
		return err
	}

	// config, bootConfig := nodeup.NewConfig(cluster, ig)

	// b := &model.BootstrapScript{
	// 	Name: ig.Name,
	// 	// Lifecycle: b.Lifecycle,
	// 	ig:      ig,
	// 	builder: b,
	// 	caTasks: caTasks,
	// }
	// task.resource.Task = task
	// task.nodeupConfig.Task = task
	// c.AddTask(task)

	// target := fi.NewDryRunTarget(assetBuilder, out)

	// var cloud fi.Cloud
	// var keyStore fi.Keystore
	// var secretStore fi.SecretStore
	// var clusterConfigBase vfs.Path
	// var checkExisting bool

	// bootstrapScriptBuilder := &model.BootstrapScriptBuilder{
	// 	Cluster: cluster,
	// }

	// 	var loader cloudup.Loader
	// 	loader.Init()
	// loader.Builders = append(loader.Builders, bootstrapScriptBuilder)

	// 	tasks, err := loader.BuildTasks(nil)
	// 	if err != nil {
	// 		return fmt.Errorf("error building tasks: %w", err)
	// 	}

	// modelBuilderContext := fi.ModelBuilderContext{
	// 	Tasks: make(map[string]fi.Task),
	// }

	// context, err := fi.NewContext(target, cluster, cloud, keyStore, secretStore, clusterConfigBase, checkExisting, tasks)
	// if err != nil {
	// 	return fmt.Errorf("error building context: %v", err)
	// }
	// defer context.Close()

	// {
	// 	defaultCA := &fitasks.Keypair{
	// 		Name:      fi.String(fi.CertificateIDCA),
	// 		Lifecycle: fi.LifecycleExistsAndValidates,
	// 		Subject:   "cn=kubernetes-ca",
	// 		Type:      "ca",
	// 	}
	// 	modelBuilderContext.AddTask(defaultCA)
	// }

	// resource, err := bootstrapScriptBuilder.ResourceNodeUp(&modelBuilderContext, ig)
	// if err != nil {
	// 	return fmt.Errorf("error building nodeup script: %w", err)
	// }

	// r, err := resource.Open()
	// if err != nil {
	// 	return fmt.Errorf("error opening nodeup script: %w", err)
	// }
	// script, err := io.ReadAll(r)
	// if err != nil {
	// 	return fmt.Errorf("error reading nodeup script: %w", err)
	// }
	// fmt.Printf("%s\n", string(script))

	return nil
}
