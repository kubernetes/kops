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

package cloudup

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/blang/semver/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	kopsbase "k8s.io/kops"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/awsmodel"
	"k8s.io/kops/pkg/model/azuremodel"
	"k8s.io/kops/pkg/model/components/etcdmanager"
	"k8s.io/kops/pkg/model/components/kubeapiserver"
	"k8s.io/kops/pkg/model/components/kubescheduler"
	"k8s.io/kops/pkg/model/domodel"
	"k8s.io/kops/pkg/model/gcemodel"
	"k8s.io/kops/pkg/model/hetznermodel"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/model/openstackmodel"
	"k8s.io/kops/pkg/model/scalewaymodel"
	"k8s.io/kops/pkg/nodemodel"
	"k8s.io/kops/pkg/templates"
	"k8s.io/kops/upup/models"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
	"k8s.io/kops/upup/pkg/fi/cloudup/bootstrapchannelbuilder"
	"k8s.io/kops/upup/pkg/fi/cloudup/do"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
	"k8s.io/kops/upup/pkg/fi/cloudup/metal"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
	"k8s.io/kops/util/pkg/vfs"
)

const (
	starline = "*********************************************************************************"

	// OldestSupportedKubernetesVersion is the oldest kubernetes version that is supported in kOps.
	OldestSupportedKubernetesVersion = "1.25.0"
	// OldestRecommendedKubernetesVersion is the oldest kubernetes version that is not deprecated in kOps.
	OldestRecommendedKubernetesVersion = "1.27.0"
)

// TerraformCloudProviders is the list of cloud providers with terraform target support
var TerraformCloudProviders = []kops.CloudProviderID{
	kops.CloudProviderAWS,
	kops.CloudProviderGCE,
	kops.CloudProviderHetzner,
	kops.CloudProviderScaleway,
	kops.CloudProviderDO,
}

type ApplyClusterCmd struct {
	Cloud   fi.Cloud
	Cluster *kops.Cluster

	InstanceGroups []*kops.InstanceGroup

	// TargetName specifies how we are operating e.g. direct to GCE, or AWS, or dry-run, or terraform
	TargetName string

	// Target is the fi.Target we will operate against
	Target fi.CloudupTarget

	// OutDir is a local directory in which we place output, can cache files etc
	OutDir string

	Clientset simple.Clientset

	// DryRun is true if this is only a dry run
	DryRun bool

	// AllowKopsDowngrade permits applying with a kops version older than what was last used to apply to the cluster.
	AllowKopsDowngrade bool

	// RunTasksOptions defines parameters for task execution, e.g. retry interval
	RunTasksOptions *fi.RunTasksOptions

	// The channel we are using
	channel *kops.Channel

	// Phase can be set to a Phase to run the specific subset of tasks, if we don't want to run everything
	Phase Phase

	// LifecycleOverrides is passed in to override the lifecycle for one of more tasks.
	// The key value is the task name such as InternetGateway and the value is the fi.Lifecycle
	// that is re-mapped.
	LifecycleOverrides map[string]fi.Lifecycle

	// GetAssets is whether this is called just to obtain the list of assets.
	GetAssets bool

	// TaskMap is the map of tasks that we built (output)
	TaskMap map[string]fi.CloudupTask

	// AdditionalObjects holds cluster-asssociated configuration objects, other than the Cluster and InstanceGroups.
	AdditionalObjects kubemanifest.ObjectList

	// DeletionProcessing controls whether we process deletions.
	DeletionProcessing fi.DeletionProcessingMode
}

// ApplyResults holds information about an ApplyClusterCmd operation.
type ApplyResults struct {
	// AssetBuilder holds the initialized AssetBuilder, listing all the image and file assets.
	AssetBuilder *assets.AssetBuilder
}

func (c *ApplyClusterCmd) Run(ctx context.Context) (*ApplyResults, error) {
	if c.TargetName == TargetTerraform {
		found := false
		for _, cp := range TerraformCloudProviders {
			if c.Cloud.ProviderID() == cp {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("cloud provider %v does not support the terraform target", c.Cloud.ProviderID())
		}
		if c.Cloud.ProviderID() == kops.CloudProviderDO && !featureflag.DOTerraform.Enabled() {
			return nil, fmt.Errorf("DO Terraform requires the DOTerraform feature flag to be enabled")
		}
	}
	if c.InstanceGroups == nil {
		list, err := c.Clientset.InstanceGroupsFor(c.Cluster).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		var instanceGroups []*kops.InstanceGroup
		for i := range list.Items {
			instanceGroups = append(instanceGroups, &list.Items[i])
		}
		c.InstanceGroups = instanceGroups
	}

	if c.AdditionalObjects == nil {
		additionalObjects, err := c.Clientset.AddonsFor(c.Cluster).List(ctx)
		if err != nil {
			return nil, err
		}
		// We use the nil object to mean "uninitialized"
		if additionalObjects == nil {
			additionalObjects = []*kubemanifest.Object{}
		}
		c.AdditionalObjects = additionalObjects
	}

	for _, ig := range c.InstanceGroups {
		// Try to guess the path for additional third party volume plugins in Flatcar
		image := strings.ToLower(ig.Spec.Image)
		if strings.Contains(image, "flatcar") {
			if c.Cluster.Spec.Kubelet == nil {
				c.Cluster.Spec.Kubelet = &kops.KubeletConfigSpec{}
			}
			if c.Cluster.Spec.Kubelet.VolumePluginDirectory == "" {
				c.Cluster.Spec.Kubelet.VolumePluginDirectory = "/var/lib/kubelet/volumeplugins/"
			}
		}
	}

	channel, err := ChannelForCluster(c.Clientset.VFSContext(), c.Cluster)
	if err != nil {
		klog.Warningf("%v", err)
	}
	c.channel = channel

	securityLifecycle := fi.LifecycleSync
	networkLifecycle := fi.LifecycleSync
	clusterLifecycle := fi.LifecycleSync

	switch c.Phase {
	case Phase(""):
		// Everything ... the default

	case PhaseNetwork:
		securityLifecycle = fi.LifecycleIgnore
		clusterLifecycle = fi.LifecycleIgnore

	case PhaseSecurity:
		networkLifecycle = fi.LifecycleExistsAndWarnIfChanges
		clusterLifecycle = fi.LifecycleIgnore

	case PhaseCluster:
		if c.TargetName == TargetDryRun {
			securityLifecycle = fi.LifecycleExistsAndWarnIfChanges
			networkLifecycle = fi.LifecycleExistsAndWarnIfChanges
		} else {
			networkLifecycle = fi.LifecycleExistsAndValidates
			securityLifecycle = fi.LifecycleExistsAndValidates
		}

	default:
		return nil, fmt.Errorf("unknown phase %q", c.Phase)
	}
	if c.GetAssets {
		networkLifecycle = fi.LifecycleIgnore
		securityLifecycle = fi.LifecycleIgnore
		clusterLifecycle = fi.LifecycleIgnore
	}

	assetBuilder := assets.NewAssetBuilder(c.Clientset.VFSContext(), c.Cluster.Spec.Assets, c.Cluster.Spec.KubernetesVersion, c.GetAssets)
	err = c.upgradeSpecs(ctx, assetBuilder)
	if err != nil {
		return nil, err
	}

	err = c.validateKopsVersion()
	if err != nil {
		return nil, err
	}

	err = c.validateKubernetesVersion()
	if err != nil {
		return nil, err
	}

	cluster := c.Cluster

	configBase, err := c.Clientset.VFSContext().BuildVfsPath(cluster.Spec.ConfigStore.Base)
	if err != nil {
		return nil, fmt.Errorf("error parsing configStore.base %q: %v", cluster.Spec.ConfigStore.Base, err)
	}

	if !c.AllowKopsDowngrade {
		kopsVersionUpdatedBytes, err := configBase.Join(registry.PathKopsVersionUpdated).ReadFile(ctx)
		if err == nil {
			kopsVersionUpdated := strings.TrimSpace(string(kopsVersionUpdatedBytes))
			version, err := semver.Parse(kopsVersionUpdated)
			if err != nil {
				return nil, fmt.Errorf("error parsing last kops version updated: %v", err)
			}
			if version.GT(semver.MustParse(kopsbase.Version)) {
				fmt.Printf("\n")
				fmt.Printf("%s\n", starline)
				fmt.Printf("\n")
				fmt.Printf("The cluster was last updated by kops version %s\n", kopsVersionUpdated)
				fmt.Printf("To permit updating by the older version %s, run with the --allow-kops-downgrade flag\n", kopsbase.Version)
				fmt.Printf("\n")
				fmt.Printf("%s\n", starline)
				fmt.Printf("\n")
				return nil, fmt.Errorf("kops version older than last used to update the cluster")
			}
		} else if err != os.ErrNotExist {
			return nil, fmt.Errorf("error reading last kops version used to update: %v", err)
		}
	}

	cloud := c.Cloud

	err = validation.DeepValidate(c.Cluster, c.InstanceGroups, true, c.Clientset.VFSContext(), cloud)
	if err != nil {
		return nil, err
	}

	if cluster.Spec.KubernetesVersion == "" {
		return nil, fmt.Errorf("KubernetesVersion not set")
	}
	if cluster.Spec.DNSZone == "" && cluster.PublishesDNSRecords() {
		return nil, fmt.Errorf("DNSZone not set")
	}

	l := &Loader{}
	l.Init()

	keyStore, err := c.Clientset.KeyStore(cluster)
	if err != nil {
		return nil, err
	}

	sshCredentialStore, err := c.Clientset.SSHCredentialStore(cluster)
	if err != nil {
		return nil, err
	}

	secretStore, err := c.Clientset.SecretStore(cluster)
	if err != nil {
		return nil, err
	}

	addonsClient := c.Clientset.AddonsFor(cluster)
	addons, err := addonsClient.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching addons: %v", err)
	}

	// Normalize k8s version
	versionWithoutV := strings.TrimSpace(cluster.Spec.KubernetesVersion)
	versionWithoutV = strings.TrimPrefix(versionWithoutV, "v")
	if cluster.Spec.KubernetesVersion != versionWithoutV {
		klog.Warningf("Normalizing kubernetes version: %q -> %q", cluster.Spec.KubernetesVersion, versionWithoutV)
		cluster.Spec.KubernetesVersion = versionWithoutV
	}

	// check if we should recommend turning off anonymousAuth
	{
		// we do a check here because setting modifying the kubelet object messes with the output
		warn := false
		if cluster.Spec.Kubelet == nil {
			warn = true
		} else if cluster.Spec.Kubelet.AnonymousAuth == nil {
			warn = true
		}

		if warn {
			fmt.Println("")
			fmt.Printf("%s\n", starline)
			fmt.Println("")
			fmt.Println("Kubelet anonymousAuth is currently turned on. This allows RBAC escalation and remote code execution possibilities.")
			fmt.Println("It is highly recommended you turn it off by setting 'spec.kubelet.anonymousAuth' to 'false' via 'kops edit cluster'")
			fmt.Println("")
			fmt.Println("See https://kops.sigs.k8s.io/security/#kubelet-api")
			fmt.Println("")
			fmt.Printf("%s\n", starline)
			fmt.Println("")
		}
	}

	encryptionConfigSecretHash := ""
	if fi.ValueOf(c.Cluster.Spec.EncryptionConfig) {
		secret, err := secretStore.FindSecret("encryptionconfig")
		if err != nil {
			return nil, fmt.Errorf("could not load encryptionconfig secret: %v", err)
		}
		if secret == nil {
			fmt.Println("")
			fmt.Println("You have encryptionConfig enabled, but no encryptionconfig secret has been set.")
			fmt.Println("See `kops create secret encryptionconfig -h` and https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/")
			return nil, fmt.Errorf("could not find encryptionconfig secret")
		}
		hashBytes := sha256.Sum256(secret.Data)
		encryptionConfigSecretHash = base64.URLEncoding.EncodeToString(hashBytes[:])
	}

	ciliumSpec := c.Cluster.Spec.Networking.Cilium
	if ciliumSpec != nil && ciliumSpec.EnableEncryption && ciliumSpec.EncryptionType == kops.CiliumEncryptionTypeIPSec {
		secret, err := secretStore.FindSecret("ciliumpassword")
		if err != nil {
			return nil, fmt.Errorf("could not load the ciliumpassword secret: %w", err)
		}
		if secret == nil {
			fmt.Println("")
			fmt.Println("You have cilium encryption enabled, but no ciliumpassword secret has been set.")
			fmt.Println("See `kops create secret ciliumpassword -h`")
			return nil, fmt.Errorf("could not find ciliumpassword secret")
		}
	}

	fileAssets := &nodemodel.FileAssets{Cluster: cluster}
	if err := fileAssets.AddFileAssets(assetBuilder); err != nil {
		return nil, err
	}

	project := ""
	scwZone := ""

	var sshPublicKeys [][]byte
	{
		keys, err := sshCredentialStore.FindSSHPublicKeys()
		if err != nil {
			return nil, fmt.Errorf("error retrieving SSH public key %q: %v", fi.SecretNameSSHPrimary, err)
		}

		for _, k := range keys {
			sshPublicKeys = append(sshPublicKeys, []byte(k.Spec.PublicKey))
		}
	}

	modelContext := &model.KopsModelContext{
		IAMModelContext:   iam.IAMModelContext{Cluster: cluster},
		InstanceGroups:    c.InstanceGroups,
		AdditionalObjects: c.AdditionalObjects,
	}

	switch cluster.GetCloudProvider() {
	case kops.CloudProviderGCE:
		{
			gceCloud := cloud.(gce.GCECloud)
			project = gceCloud.Project()
		}

	case kops.CloudProviderHetzner:
		{
			// Hetzner Cloud support is currently in beta
		}

	case kops.CloudProviderDO:
		{
			if len(sshPublicKeys) == 0 && (c.Cluster.Spec.SSHKeyName == nil || *c.Cluster.Spec.SSHKeyName == "") {
				return nil, fmt.Errorf("SSH public key must be specified when running with DigitalOcean (create with `kops create secret --name %s sshpublickey admin -i ~/.ssh/id_rsa.pub`)", cluster.ObjectMeta.Name)
			}
		}
	case kops.CloudProviderAWS:
		{
			awsCloud := cloud.(awsup.AWSCloud)

			accountID, partition, err := awsCloud.AccountInfo(ctx)
			if err != nil {
				return nil, err
			}
			modelContext.AWSAccountID = accountID
			modelContext.AWSPartition = partition

			if len(sshPublicKeys) > 1 {
				return nil, fmt.Errorf("exactly one 'admin' SSH public key can be specified when running with AWS; please delete a key using `kops delete secret`")
			}
		}

	case kops.CloudProviderAzure:
		{
			if !featureflag.Azure.Enabled() {
				return nil, fmt.Errorf("azure support is currently alpha, and is feature-gated. Please export KOPS_FEATURE_FLAGS=Azure")
			}

			if len(sshPublicKeys) == 0 {
				return nil, fmt.Errorf("SSH public key must be specified when running with AzureCloud (create with `kops create secret --name %s sshpublickey admin -i ~/.ssh/id_rsa.pub`)", cluster.ObjectMeta.Name)
			}

			if len(sshPublicKeys) != 1 {
				return nil, fmt.Errorf("exactly one 'admin' SSH public key can be specified when running with AzureCloud; please delete a key using `kops delete secret`")
			}
		}
	case kops.CloudProviderOpenstack:
		{
			if len(sshPublicKeys) == 0 {
				return nil, fmt.Errorf("SSH public key must be specified when running with Openstack (create with `kops create secret --name %s sshpublickey admin -i ~/.ssh/id_rsa.pub`)", cluster.ObjectMeta.Name)
			}

			if len(sshPublicKeys) != 1 {
				return nil, fmt.Errorf("exactly one 'admin' SSH public key can be specified when running with Openstack; please delete a key using `kops delete secret`")
			}
		}

	case kops.CloudProviderScaleway:
		{
			if !featureflag.Scaleway.Enabled() {
				return nil, fmt.Errorf("Scaleway support is currently alpha, and is feature-gated.  export KOPS_FEATURE_FLAGS=Scaleway")
			}

			if len(sshPublicKeys) == 0 {
				return nil, fmt.Errorf("SSH public key must be specified when running with Scaleway (create with `kops create secret --name %s sshpublickey admin -i ~/.ssh/id_rsa.pub`)", cluster.ObjectMeta.Name)
			}
			if len(sshPublicKeys) != 1 {
				return nil, fmt.Errorf("exactly one 'admin' SSH public key can be specified when running with Scaleway; please delete a key using `kops delete secret`")
			}

			scwCloud := cloud.(scaleway.ScwCloud)
			scwZone = scwCloud.Zone()
		}

	case kops.CloudProviderMetal:
		// Metal is a special case, we don't need to do anything here (yet)

	default:
		return nil, fmt.Errorf("unknown CloudProvider %q", cluster.GetCloudProvider())
	}

	modelContext.SSHPublicKeys = sshPublicKeys
	modelContext.Region = cloud.Region()

	if cluster.PublishesDNSRecords() {
		err = validateDNS(cluster, cloud)
		if err != nil {
			return nil, err
		}
	}

	tf := &TemplateFunctions{
		KopsModelContext: *modelContext,
		cloud:            cloud,
	}

	configBuilder, err := nodemodel.NewNodeUpConfigBuilder(cluster, assetBuilder, fileAssets.Assets, encryptionConfigSecretHash)
	if err != nil {
		return nil, err
	}
	bootstrapScriptBuilder := &model.BootstrapScriptBuilder{
		KopsModelContext:    modelContext,
		Lifecycle:           clusterLifecycle,
		NodeUpConfigBuilder: configBuilder,
		NodeUpAssets:        fileAssets.NodeUpAssets,
	}

	{
		templates, err := templates.LoadTemplates(ctx, cluster, models.NewAssetPath("cloudup/resources"))
		if err != nil {
			return nil, fmt.Errorf("error loading templates: %v", err)
		}

		err = tf.AddTo(templates.TemplateFunctions, secretStore)
		if err != nil {
			return nil, err
		}

		bcb := bootstrapchannelbuilder.NewBootstrapChannelBuilder(
			modelContext,
			clusterLifecycle,
			assetBuilder,
			templates,
			addons,
		)

		l.Builders = append(l.Builders,

			bcb,
			&model.PKIModelBuilder{
				KopsModelContext: modelContext,
				Lifecycle:        clusterLifecycle,
			},
			&model.IssuerDiscoveryModelBuilder{
				KopsModelContext: modelContext,
				Lifecycle:        clusterLifecycle,
				Cluster:          cluster,
			},
			&kubeapiserver.KubeApiserverBuilder{
				AssetBuilder:     assetBuilder,
				KopsModelContext: modelContext,
				Lifecycle:        clusterLifecycle,
			},
			&kubescheduler.KubeSchedulerBuilder{
				AssetBuilder:     assetBuilder,
				KopsModelContext: modelContext,
				Lifecycle:        clusterLifecycle,
			},
			&etcdmanager.EtcdManagerBuilder{
				AssetBuilder:     assetBuilder,
				KopsModelContext: modelContext,
				Lifecycle:        clusterLifecycle,
			},
			&model.MasterVolumeBuilder{KopsModelContext: modelContext, Lifecycle: clusterLifecycle},
			&model.ConfigBuilder{KopsModelContext: modelContext, Lifecycle: clusterLifecycle},
		)

		switch cluster.GetCloudProvider() {
		case kops.CloudProviderAWS:
			awsModelContext := &awsmodel.AWSModelContext{
				KopsModelContext: modelContext,
			}

			l.Builders = append(l.Builders,
				&awsmodel.APILoadBalancerBuilder{AWSModelContext: awsModelContext, Lifecycle: clusterLifecycle, SecurityLifecycle: securityLifecycle},
				&awsmodel.BastionModelBuilder{AWSModelContext: awsModelContext, Lifecycle: clusterLifecycle, SecurityLifecycle: securityLifecycle},
				&awsmodel.DNSModelBuilder{AWSModelContext: awsModelContext, Lifecycle: clusterLifecycle},
				&awsmodel.ExternalAccessModelBuilder{AWSModelContext: awsModelContext, Lifecycle: securityLifecycle},
				&awsmodel.FirewallModelBuilder{AWSModelContext: awsModelContext, Lifecycle: securityLifecycle},
				&awsmodel.SSHKeyModelBuilder{AWSModelContext: awsModelContext, Lifecycle: securityLifecycle},
				&awsmodel.NetworkModelBuilder{AWSModelContext: awsModelContext, Lifecycle: networkLifecycle},
				&awsmodel.IAMModelBuilder{AWSModelContext: awsModelContext, Lifecycle: securityLifecycle, Cluster: cluster},
				&awsmodel.OIDCProviderBuilder{AWSModelContext: awsModelContext, Lifecycle: securityLifecycle},
			)

			awsModelBuilder := &awsmodel.AutoscalingGroupModelBuilder{
				AWSModelContext:        awsModelContext,
				BootstrapScriptBuilder: bootstrapScriptBuilder,
				Lifecycle:              clusterLifecycle,
				SecurityLifecycle:      securityLifecycle,
				Cluster:                cluster,
			}

			if featureflag.Spotinst.Enabled() {
				l.Builders = append(l.Builders, &awsmodel.SpotInstanceGroupModelBuilder{
					AWSModelContext:        awsModelContext,
					BootstrapScriptBuilder: bootstrapScriptBuilder,
					Lifecycle:              clusterLifecycle,
					SecurityLifecycle:      securityLifecycle,
				})

				if featureflag.SpotinstHybrid.Enabled() {
					l.Builders = append(l.Builders, awsModelBuilder)
				}
			} else {
				l.Builders = append(l.Builders, awsModelBuilder)
			}

			nth := c.Cluster.Spec.CloudProvider.AWS.NodeTerminationHandler
			if nth.IsQueueMode() {
				l.Builders = append(l.Builders, &awsmodel.NodeTerminationHandlerBuilder{
					AWSModelContext: awsModelContext,
					Lifecycle:       clusterLifecycle,
				})
			}

		case kops.CloudProviderDO:
			doModelContext := &domodel.DOModelContext{
				KopsModelContext: modelContext,
			}
			l.Builders = append(l.Builders,
				&domodel.APILoadBalancerModelBuilder{DOModelContext: doModelContext, Lifecycle: securityLifecycle},
				&domodel.DropletBuilder{DOModelContext: doModelContext, BootstrapScriptBuilder: bootstrapScriptBuilder, Lifecycle: clusterLifecycle},
				&domodel.NetworkModelBuilder{DOModelContext: doModelContext, Lifecycle: networkLifecycle},
			)
		case kops.CloudProviderHetzner:
			hetznerModelContext := &hetznermodel.HetznerModelContext{
				KopsModelContext: modelContext,
			}
			l.Builders = append(l.Builders,
				&hetznermodel.NetworkModelBuilder{HetznerModelContext: hetznerModelContext, Lifecycle: networkLifecycle},
				&hetznermodel.ExternalAccessModelBuilder{HetznerModelContext: hetznerModelContext, Lifecycle: networkLifecycle},
				&hetznermodel.LoadBalancerModelBuilder{HetznerModelContext: hetznerModelContext, Lifecycle: networkLifecycle},
				&hetznermodel.ServerGroupModelBuilder{HetznerModelContext: hetznerModelContext, BootstrapScriptBuilder: bootstrapScriptBuilder, Lifecycle: clusterLifecycle},
			)
		case kops.CloudProviderGCE:
			gceModelContext := &gcemodel.GCEModelContext{
				ProjectID:        project,
				KopsModelContext: modelContext,
			}

			storageACLLifecycle := securityLifecycle
			if storageACLLifecycle != fi.LifecycleIgnore {
				// This is a best-effort permissions fix
				storageACLLifecycle = fi.LifecycleWarnIfInsufficientAccess
			}

			l.Builders = append(l.Builders,

				&gcemodel.APILoadBalancerBuilder{GCEModelContext: gceModelContext, Lifecycle: securityLifecycle},
				&gcemodel.ExternalAccessModelBuilder{GCEModelContext: gceModelContext, Lifecycle: securityLifecycle},
				&gcemodel.FirewallModelBuilder{GCEModelContext: gceModelContext, Lifecycle: securityLifecycle},
				&gcemodel.NetworkModelBuilder{GCEModelContext: gceModelContext, Lifecycle: networkLifecycle},
				&gcemodel.StorageAclBuilder{GCEModelContext: gceModelContext, Cloud: cloud.(gce.GCECloud), Lifecycle: storageACLLifecycle},
				&gcemodel.AutoscalingGroupModelBuilder{GCEModelContext: gceModelContext, BootstrapScriptBuilder: bootstrapScriptBuilder, Lifecycle: clusterLifecycle},
				&gcemodel.ServiceAccountsBuilder{GCEModelContext: gceModelContext, Lifecycle: clusterLifecycle},
			)
		case kops.CloudProviderAzure:
			azureModelContext := &azuremodel.AzureModelContext{
				KopsModelContext: modelContext,
			}
			l.Builders = append(l.Builders,
				&azuremodel.APILoadBalancerModelBuilder{AzureModelContext: azureModelContext, Lifecycle: clusterLifecycle},
				&azuremodel.NetworkModelBuilder{AzureModelContext: azureModelContext, Lifecycle: clusterLifecycle},
				&azuremodel.ResourceGroupModelBuilder{AzureModelContext: azureModelContext, Lifecycle: clusterLifecycle},

				&azuremodel.VMScaleSetModelBuilder{AzureModelContext: azureModelContext, BootstrapScriptBuilder: bootstrapScriptBuilder, Lifecycle: clusterLifecycle},
			)
		case kops.CloudProviderOpenstack:
			openstackModelContext := &openstackmodel.OpenstackModelContext{
				KopsModelContext: modelContext,
			}

			l.Builders = append(l.Builders,
				&openstackmodel.NetworkModelBuilder{OpenstackModelContext: openstackModelContext, Lifecycle: networkLifecycle},
				&openstackmodel.SSHKeyModelBuilder{OpenstackModelContext: openstackModelContext, Lifecycle: securityLifecycle},
				&openstackmodel.FirewallModelBuilder{OpenstackModelContext: openstackModelContext, Lifecycle: securityLifecycle},
				&openstackmodel.ServerGroupModelBuilder{OpenstackModelContext: openstackModelContext, BootstrapScriptBuilder: bootstrapScriptBuilder, Lifecycle: clusterLifecycle},
			)

		case kops.CloudProviderScaleway:
			scwModelContext := &scalewaymodel.ScwModelContext{
				KopsModelContext: modelContext,
			}
			l.Builders = append(l.Builders,
				&scalewaymodel.APILoadBalancerModelBuilder{ScwModelContext: scwModelContext, Lifecycle: networkLifecycle},
				&scalewaymodel.DNSModelBuilder{ScwModelContext: scwModelContext, Lifecycle: networkLifecycle},
				&scalewaymodel.InstanceModelBuilder{ScwModelContext: scwModelContext, BootstrapScriptBuilder: bootstrapScriptBuilder, Lifecycle: clusterLifecycle},
				&scalewaymodel.SSHKeyModelBuilder{ScwModelContext: scwModelContext, Lifecycle: securityLifecycle},
			)

		case kops.CloudProviderMetal:
			// No special builders for bare metal (yet)

		default:
			return nil, fmt.Errorf("unknown cloudprovider %q", cluster.GetCloudProvider())
		}
	}
	c.TaskMap, err = l.BuildTasks(ctx, c.LifecycleOverrides)
	if err != nil {
		return nil, fmt.Errorf("error building tasks: %v", err)
	}

	var target fi.CloudupTarget
	shouldPrecreateDNS := true

	deletionProcessingMode := c.DeletionProcessing
	switch c.TargetName {
	case TargetDirect:
		switch cluster.GetCloudProvider() {
		case kops.CloudProviderGCE:
			target = gce.NewGCEAPITarget(cloud.(gce.GCECloud))
		case kops.CloudProviderAWS:
			target = awsup.NewAWSAPITarget(cloud.(awsup.AWSCloud))
		case kops.CloudProviderDO:
			target = do.NewDOAPITarget(cloud.(do.DOCloud))
		case kops.CloudProviderHetzner:
			target = hetzner.NewHetznerAPITarget(cloud.(hetzner.HetznerCloud))
		case kops.CloudProviderOpenstack:
			target = openstack.NewOpenstackAPITarget(cloud.(openstack.OpenstackCloud))
		case kops.CloudProviderAzure:
			target = azure.NewAzureAPITarget(cloud.(azure.AzureCloud))
		case kops.CloudProviderScaleway:
			target = scaleway.NewScwAPITarget(cloud.(scaleway.ScwCloud))
		case kops.CloudProviderMetal:
			target = metal.NewAPITarget(cloud.(*metal.Cloud), nil)
		default:
			return nil, fmt.Errorf("direct configuration not supported with CloudProvider:%q", cluster.GetCloudProvider())
		}

	case TargetTerraform:
		outDir := c.OutDir
		tf := terraform.NewTerraformTarget(cloud, project, outDir, cluster.Spec.Target)

		// We include a few "util" variables in the TF output
		if err := tf.AddOutputVariable("region", terraformWriter.LiteralFromStringValue(cloud.Region())); err != nil {
			return nil, err
		}

		if project != "" {
			if err := tf.AddOutputVariable("project", terraformWriter.LiteralFromStringValue(project)); err != nil {
				return nil, err
			}
		}
		if scwZone != "" {
			if err := tf.AddOutputVariable("zone", terraformWriter.LiteralFromStringValue(scwZone)); err != nil {
				return nil, err
			}
		}

		if err := tf.AddOutputVariable("cluster_name", terraformWriter.LiteralFromStringValue(cluster.ObjectMeta.Name)); err != nil {
			return nil, err
		}

		target = tf

		// Can cause conflicts with terraform management
		shouldPrecreateDNS = false

		// Terraform tracks & performs deletions itself
		deletionProcessingMode = fi.DeletionProcessingModeIgnore

	case TargetDryRun:
		var out io.Writer = os.Stdout
		if c.GetAssets {
			out = io.Discard
		}
		target = fi.NewCloudupDryRunTarget(assetBuilder, out)

		// Avoid making changes on a dry-run
		shouldPrecreateDNS = false

	default:
		return nil, fmt.Errorf("unsupported target type %q", c.TargetName)
	}
	c.Target = target

	if target.DefaultCheckExisting() {
		c.TaskMap, err = l.FindDeletions(cloud, c.LifecycleOverrides)
		if err != nil {
			return nil, fmt.Errorf("error finding deletions: %w", err)
		}
	}

	context, err := fi.NewCloudupContext(ctx, deletionProcessingMode, target, cluster, cloud, keyStore, secretStore, configBase, c.TaskMap)
	if err != nil {
		return nil, fmt.Errorf("error building context: %v", err)
	}

	var options fi.RunTasksOptions
	if c.RunTasksOptions != nil {
		options = *c.RunTasksOptions
	} else {
		options.InitDefaults()
	}

	err = context.RunTasks(options)
	if err != nil {
		return nil, fmt.Errorf("error running tasks: %v", err)
	}

	if !cluster.PublishesDNSRecords() {
		shouldPrecreateDNS = false
	}

	if shouldPrecreateDNS && clusterLifecycle != fi.LifecycleIgnore {
		if err := precreateDNS(ctx, cluster, cloud); err != nil {
			klog.Warningf("unable to pre-create DNS records - cluster startup may be slower: %v", err)
		}
	}

	err = target.Finish(c.TaskMap) // This will finish the apply, and print the changes
	if err != nil {
		return nil, fmt.Errorf("error closing target: %v", err)
	}

	applyResults := &ApplyResults{
		AssetBuilder: assetBuilder,
	}

	return applyResults, nil
}

// upgradeSpecs ensures that fields are fully populated / defaulted
func (c *ApplyClusterCmd) upgradeSpecs(ctx context.Context, assetBuilder *assets.AssetBuilder) error {
	fullCluster, err := PopulateClusterSpec(ctx, c.Clientset, c.Cluster, c.InstanceGroups, c.Cloud, assetBuilder)
	if err != nil {
		return err
	}
	c.Cluster = fullCluster

	for i, g := range c.InstanceGroups {
		fullGroup, err := PopulateInstanceGroupSpec(fullCluster, g, c.Cloud, c.channel)
		if err != nil {
			return err
		}
		c.InstanceGroups[i] = fullGroup
	}

	return nil
}

// validateKopsVersion ensures that kops meet the version requirements / recommendations in the channel
func (c *ApplyClusterCmd) validateKopsVersion() error {
	kopsVersion, err := semver.ParseTolerant(kopsbase.Version)
	if err != nil {
		klog.Warningf("unable to parse kops version %q", kopsbase.Version)
		// Not a hard-error
		return nil
	}

	if c.channel == nil {
		klog.Warning("channel unavailable, skipping version validation")
		return nil
	}

	versionInfo := kops.FindKopsVersionSpec(c.channel.Spec.KopsVersions, kopsVersion)
	if versionInfo == nil {
		klog.Warningf("unable to find version information for kops version %q in channel", kopsVersion)
		// Not a hard-error
		return nil
	}

	recommended, err := versionInfo.FindRecommendedUpgrade(kopsVersion)
	if err != nil {
		klog.Warningf("unable to parse version recommendation for kops version %q in channel", kopsVersion)
	}

	required, err := versionInfo.IsUpgradeRequired(kopsVersion)
	if err != nil {
		klog.Warningf("unable to parse version requirement for kops version %q in channel", kopsVersion)
	}

	if recommended != nil && !required && !c.GetAssets {
		fmt.Printf("\n")
		fmt.Printf("%s\n", starline)
		fmt.Printf("\n")
		fmt.Printf("A new kops version is available: %s", recommended)
		fmt.Printf("\n")
		fmt.Printf("Upgrading is recommended\n")
		fmt.Printf("More information: %s\n", buildPermalink("upgrade_kops", recommended.String()))
		fmt.Printf("\n")
		fmt.Printf("%s\n", starline)
		fmt.Printf("\n")
	} else if required {
		fmt.Printf("\n")
		fmt.Printf("%s\n", starline)
		fmt.Printf("\n")
		if recommended != nil {
			fmt.Printf("a new kops version is available: %s\n", recommended)
		}
		fmt.Println("")
		fmt.Printf("This version of kops (%s) is no longer supported; upgrading is required\n", kopsbase.Version)
		fmt.Printf("(you can bypass this check by exporting KOPS_RUN_OBSOLETE_VERSION)\n")
		fmt.Println("")
		fmt.Printf("More information: %s\n", buildPermalink("upgrade_kops", recommended.String()))
		fmt.Printf("\n")
		fmt.Printf("%s\n", starline)
		fmt.Printf("\n")
	}

	if required {
		if os.Getenv("KOPS_RUN_OBSOLETE_VERSION") == "" {
			return fmt.Errorf("kops upgrade is required")
		}
	}

	return nil
}

// validateKubernetesVersion ensures that kubernetes meet the version requirements / recommendations in the channel
func (c *ApplyClusterCmd) validateKubernetesVersion() error {
	parsed, err := util.ParseKubernetesVersion(c.Cluster.Spec.KubernetesVersion)
	if err != nil {
		klog.Warningf("unable to parse kubernetes version %q", c.Cluster.Spec.KubernetesVersion)
		// Not a hard-error
		return nil
	}

	kopsVersion, err := semver.Parse(kopsbase.KOPS_RELEASE_VERSION)
	if err != nil {
		klog.Warningf("unable to parse kops version %q", kopsVersion)
	} else {
		tooNewVersion := kopsVersion
		tooNewVersion.Minor++
		tooNewVersion.Pre = nil
		tooNewVersion.Build = nil
		if util.IsKubernetesGTE(tooNewVersion.String(), *parsed) {
			bypassCheck := os.Getenv("KOPS_RUN_TOO_NEW_VERSION") != ""

			fmt.Printf("\n")
			fmt.Printf("%s\n", starline)
			fmt.Printf("\n")
			fmt.Printf("This version of kubernetes is not yet supported; upgrading kops is required\n")
			if bypassCheck {
				fmt.Printf("(this check has been bypassed by exporting KOPS_RUN_TOO_NEW_VERSION)\n")
			} else {
				fmt.Printf("(you can bypass this check by exporting KOPS_RUN_TOO_NEW_VERSION)\n")
			}
			fmt.Printf("\n")
			fmt.Printf("%s\n", starline)
			fmt.Printf("\n")
			if !bypassCheck {
				return fmt.Errorf("kops upgrade is required")
			}
		}
	}

	if !util.IsKubernetesGTE(OldestSupportedKubernetesVersion, *parsed) {
		fmt.Printf("This version of Kubernetes is no longer supported; upgrading Kubernetes is required\n")
		fmt.Printf("\n")
		fmt.Printf("More information: %s\n", buildPermalink("upgrade_k8s", OldestRecommendedKubernetesVersion))
		fmt.Printf("\n")
		fmt.Printf("%s\n", starline)
		fmt.Printf("\n")
		return fmt.Errorf("kubernetes upgrade is required")
	}
	if !util.IsKubernetesGTE(OldestRecommendedKubernetesVersion, *parsed) && !c.GetAssets {
		fmt.Printf("\n")
		fmt.Printf("%s\n", starline)
		fmt.Printf("\n")
		fmt.Printf("Kops support for this Kubernetes version is deprecated and will be removed in a future release.\n")
		fmt.Printf("\n")
		fmt.Printf("Upgrading Kubernetes is recommended\n")
		fmt.Printf("More information: %s\n", buildPermalink("upgrade_k8s", OldestRecommendedKubernetesVersion))
		fmt.Printf("\n")
		fmt.Printf("%s\n", starline)
		fmt.Printf("\n")

	}

	// TODO: make util.ParseKubernetesVersion not return a pointer
	kubernetesVersion := *parsed

	if c.channel == nil {
		klog.Warning("unable to load channel, skipping kubernetes version recommendation/requirements checks")
		return nil
	}

	versionInfo := kops.FindKubernetesVersionSpec(c.channel.Spec.KubernetesVersions, kubernetesVersion)
	if versionInfo == nil {
		klog.Warningf("unable to find version information for kubernetes version %q in channel", kubernetesVersion)
		// Not a hard-error
		return nil
	}

	recommended, err := versionInfo.FindRecommendedUpgrade(kubernetesVersion)
	if err != nil {
		klog.Warningf("unable to parse version recommendation for kubernetes version %q in channel", kubernetesVersion)
	}

	required, err := versionInfo.IsUpgradeRequired(kubernetesVersion)
	if err != nil {
		klog.Warningf("unable to parse version requirement for kubernetes version %q in channel", kubernetesVersion)
	}

	if recommended != nil && !required && !c.GetAssets {
		fmt.Printf("\n")
		fmt.Printf("%s\n", starline)
		fmt.Printf("\n")
		fmt.Printf("A new kubernetes version is available: %s\n", recommended)
		fmt.Printf("Upgrading is recommended (try kops upgrade cluster)\n")
		fmt.Printf("\n")
		fmt.Printf("More information: %s\n", buildPermalink("upgrade_k8s", recommended.String()))
		fmt.Printf("\n")
		fmt.Printf("%s\n", starline)
		fmt.Printf("\n")
	} else if required {
		fmt.Printf("\n")
		fmt.Printf("%s\n", starline)
		fmt.Printf("\n")
		if recommended != nil {
			fmt.Printf("A new kubernetes version is available: %s\n", recommended)
		}
		fmt.Printf("\n")
		fmt.Printf("This version of kubernetes is no longer supported; upgrading is required\n")
		fmt.Printf("(you can bypass this check by exporting KOPS_RUN_OBSOLETE_VERSION)\n")
		fmt.Printf("\n")
		fmt.Printf("More information: %s\n", buildPermalink("upgrade_k8s", recommended.String()))
		fmt.Printf("\n")
		fmt.Printf("%s\n", starline)
		fmt.Printf("\n")
	}

	if required {
		if os.Getenv("KOPS_RUN_OBSOLETE_VERSION") == "" {
			return fmt.Errorf("kubernetes upgrade is required")
		}
	}

	return nil
}

// buildPermalink returns a link to our "permalink docs", to further explain an error message
func buildPermalink(key, anchor string) string {
	url := "https://github.com/kubernetes/kops/blob/master/permalinks/" + key + ".md"
	if anchor != "" {
		url += "#" + anchor
	}
	return url
}

func ChannelForCluster(vfsContext *vfs.VFSContext, c *kops.Cluster) (*kops.Channel, error) {
	channelLocation := c.Spec.Channel
	if channelLocation == "" {
		channelLocation = kops.DefaultChannel
	}
	return kops.LoadChannel(vfsContext, channelLocation)
}
