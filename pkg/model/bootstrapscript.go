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

package model

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sort"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/model/resources"
	"k8s.io/kops/pkg/wellknownservices"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/architectures"
)

type NodeUpConfigBuilder interface {
	BuildConfig(ig *kops.InstanceGroup, wellKnownAddresses WellKnownAddresses, keysets map[string]*fi.Keyset) (*nodeup.Config, *nodeup.BootConfig, error)
}

// WellKnownAddresses holds known addresses for well-known services
type WellKnownAddresses map[wellknownservices.WellKnownService][]string

// BootstrapScriptBuilder creates the bootstrap script
type BootstrapScriptBuilder struct {
	*KopsModelContext
	Lifecycle           fi.Lifecycle
	NodeUpAssets        map[architectures.Architecture]*assets.MirroredAsset
	NodeUpConfigBuilder NodeUpConfigBuilder
}

type BootstrapScript struct {
	Name      string
	Lifecycle fi.Lifecycle
	cluster   *kops.Cluster
	ig        *kops.InstanceGroup
	builder   *BootstrapScriptBuilder
	resource  fi.CloudupTaskDependentResource

	// hasAddressTasks holds fi.HasAddress tasks, that contribute well-known services.
	hasAddressTasks []fi.HasAddress

	// caTasks hold the CA tasks, for dependency analysis.
	caTasks map[string]*fitasks.Keypair

	// nodeupConfig contains the nodeup config.
	nodeupConfig fi.CloudupTaskDependentResource
}

var (
	_ fi.CloudupTask            = &BootstrapScript{}
	_ fi.HasName                = &BootstrapScript{}
	_ fi.CloudupHasDependencies = &BootstrapScript{}
)

// kubeEnv returns the boot config for the instance group
func (b *BootstrapScript) kubeEnv(ig *kops.InstanceGroup, c *fi.CloudupContext) (*nodeup.BootConfig, error) {
	wellKnownAddresses := make(WellKnownAddresses)

	for _, hasAddress := range b.hasAddressTasks {
		addresses, err := hasAddress.FindAddresses(c)
		if err != nil {
			return nil, fmt.Errorf("error finding address for %v: %v", hasAddress, err)
		}
		if len(addresses) == 0 {
			// Such tasks won't have an address in dry-run mode, until the resource is created
			klog.V(2).Infof("Task did not have an address: %v", hasAddress)
			continue
		}

		klog.V(8).Infof("Resolved alternateNames %q for %q", addresses, hasAddress)

		for _, wellKnownService := range hasAddress.GetWellKnownServices() {
			wellKnownAddresses[wellKnownService] = append(wellKnownAddresses[wellKnownService], addresses...)
		}
	}

	for k := range wellKnownAddresses {
		sort.Strings(wellKnownAddresses[k])
	}

	keysets := make(map[string]*fi.Keyset)
	for _, caTask := range b.caTasks {
		name := *caTask.Name
		keyset := caTask.Keyset()
		if keyset == nil {
			return nil, fmt.Errorf("failed to get keyset from %q", name)
		}
		keysets[name] = keyset
	}
	config, bootConfig, err := b.builder.NodeUpConfigBuilder.BuildConfig(ig, wellKnownAddresses, keysets)
	if err != nil {
		return nil, err
	}

	configData, err := utils.YamlMarshal(config)
	if err != nil {
		return nil, fmt.Errorf("error converting nodeup config to yaml: %v", err)
	}
	sum256 := sha256.Sum256(configData)
	bootConfig.NodeupConfigHash = base64.StdEncoding.EncodeToString(sum256[:])
	b.nodeupConfig.Resource = fi.NewBytesResource(configData)

	return bootConfig, nil
}

func KeypairNamesForInstanceGroup(cluster *kops.Cluster, ig *kops.InstanceGroup) []string {
	keypairs := []string{"kubernetes-ca"}

	// Add keypairs for default etcd clusters (main and events, not cilium)
	if ig.IsControlPlane() {
		for _, etcdCluster := range cluster.Spec.EtcdClusters {
			k := etcdCluster.Name
			if k != "events" && k != "main" {
				// Likely cilium
				continue
			}
			keypairs = append(keypairs, "etcd-manager-ca-"+k, "etcd-peers-ca-"+k)
			// The client ca certificate is shared between events and main etcd clusters
			keypairs = append(keypairs, "etcd-clients-ca")
		}
	}

	if ig.HasAPIServer() {
		keypairs = append(keypairs, "apiserver-aggregator-ca", "service-account", "etcd-clients-ca")
	}

	// Add keypairs for cilium etcd clusters (not the default etcd clusters)
	for _, etcdCluster := range cluster.Spec.EtcdClusters {
		k := etcdCluster.Name
		if k == "events" || k == "main" {
			// Not cilium
			continue
		}

		keypairs = append(keypairs, "etcd-manager-ca-"+k, "etcd-peers-ca-"+k, "etcd-clients-ca-"+k)
	}

	if ig.IsBastion() {
		keypairs = nil
	}

	return keypairs
}

// ResourceNodeUp generates and returns a nodeup (bootstrap) script from a
// template file, substituting in specific env vars & cluster spec configuration
func (b *BootstrapScriptBuilder) ResourceNodeUp(c *fi.CloudupModelBuilderContext, ig *kops.InstanceGroup) (fi.Resource, error) {
	keypairNames := KeypairNamesForInstanceGroup(b.Cluster, ig)

	if ig.IsBastion() {
		// Bastions can have AdditionalUserData, but if there isn't any skip this part
		if len(ig.Spec.AdditionalUserData) == 0 {
			return nil, nil
		}
	}

	caTasks := map[string]*fitasks.Keypair{}
	for _, keypair := range keypairNames {
		caTaskObject, found := c.Tasks["Keypair/"+keypair]
		if !found {
			return nil, fmt.Errorf("keypair/%s task not found", keypair)
		}
		caTasks[keypair] = caTaskObject.(*fitasks.Keypair)
	}

	task := &BootstrapScript{
		Name:      ig.Name,
		Lifecycle: b.Lifecycle,
		cluster:   b.Cluster,
		ig:        ig,
		builder:   b,
		caTasks:   caTasks,
	}
	task.resource.Task = task
	task.nodeupConfig.Task = task
	c.AddTask(task)

	c.AddTask(&fitasks.ManagedFile{
		Name:      fi.PtrTo("nodeupconfig-" + ig.Name),
		Lifecycle: b.Lifecycle,
		Location:  fi.PtrTo("igconfig/" + ig.Spec.Role.ToLowerString() + "/" + ig.Name + "/nodeupconfig.yaml"),
		Contents:  &task.nodeupConfig,
	})
	return &task.resource, nil
}

func (b *BootstrapScript) GetName() *string {
	return &b.Name
}

func (b *BootstrapScript) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	var deps []fi.CloudupTask

	for _, task := range tasks {
		if hasAddress, ok := task.(fi.HasAddress); ok && len(hasAddress.GetWellKnownServices()) > 0 {
			deps = append(deps, task)
			b.hasAddressTasks = append(b.hasAddressTasks, hasAddress)
		}
	}

	for _, task := range b.caTasks {
		deps = append(deps, task)
	}

	return deps
}

func (b *BootstrapScript) Run(c *fi.CloudupContext) error {
	if b.Lifecycle == fi.LifecycleIgnore {
		return nil
	}

	bootConfig, err := b.kubeEnv(b.ig, c)
	if err != nil {
		return err
	}

	var nodeupScript resources.NodeUpScript
	nodeupScript.NodeUpAssets = b.builder.NodeUpAssets
	nodeupScript.BootConfig = bootConfig

	nodeupScript.WithEnvironmentVariables(b.cluster, b.ig)
	nodeupScript.WithProxyEnv(b.cluster)
	nodeupScript.WithSysctls()

	nodeupScript.CompressUserData = fi.ValueOf(b.ig.Spec.CompressUserData)

	nodeupScript.CloudProvider = string(c.T.Cluster.GetCloudProvider())

	nodeupScriptResource, err := nodeupScript.Build()
	if err != nil {
		return err
	}

	b.resource.Resource = fi.FunctionToResource(func() ([]byte, error) {
		nodeupScript, err := fi.ResourceAsString(nodeupScriptResource)
		if err != nil {
			return nil, err
		}

		awsUserData, err := resources.AWSMultipartMIME(nodeupScript, b.ig)
		if err != nil {
			return nil, err
		}

		return []byte(awsUserData), nil
	})
	return nil
}
