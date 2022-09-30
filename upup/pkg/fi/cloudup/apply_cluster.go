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
	"net"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"k8s.io/kops/pkg/model/yandexmodel"
	"k8s.io/kops/upup/pkg/fi/cloudup/yandex"

	"github.com/blang/semver/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	kopsbase "k8s.io/kops"
	"k8s.io/kops/pkg/apis/kops"
	apiModel "k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/awsmodel"
	"k8s.io/kops/pkg/model/azuremodel"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/pkg/model/components/etcdmanager"
	"k8s.io/kops/pkg/model/components/kubeapiserver"
	"k8s.io/kops/pkg/model/components/kubescheduler"
	"k8s.io/kops/pkg/model/domodel"
	"k8s.io/kops/pkg/model/gcemodel"
	"k8s.io/kops/pkg/model/hetznermodel"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/model/openstackmodel"
	"k8s.io/kops/pkg/templates"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/models"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
	"k8s.io/kops/upup/pkg/fi/cloudup/bootstrapchannelbuilder"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/do"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/hashing"
	"k8s.io/kops/util/pkg/mirrors"
	"k8s.io/kops/util/pkg/reflectutils"
	"k8s.io/kops/util/pkg/vfs"
)

const (
	starline = "*********************************************************************************"

	// OldestSupportedKubernetesVersion is the oldest kubernetes version that is supported in kOps.
	OldestSupportedKubernetesVersion = "1.21.0"
	// OldestRecommendedKubernetesVersion is the oldest kubernetes version that is not deprecated in kOps.
	OldestRecommendedKubernetesVersion = "1.23.0"
)

// TerraformCloudProviders is the list of cloud providers with terraform target support
var TerraformCloudProviders = []kops.CloudProviderID{
	kops.CloudProviderAWS,
	kops.CloudProviderGCE,
	kops.CloudProviderHetzner,
}

type ApplyClusterCmd struct {
	Cloud   fi.Cloud
	Cluster *kops.Cluster

	InstanceGroups []*kops.InstanceGroup

	// NodeUpAssets are the assets for downloading nodeup
	NodeUpAssets map[architectures.Architecture]*mirrors.MirroredAsset

	// TargetName specifies how we are operating e.g. direct to GCE, or AWS, or dry-run, or terraform
	TargetName string

	// Target is the fi.Target we will operate against
	Target fi.Target

	// OutDir is a local directory in which we place output, can cache files etc
	OutDir string

	// Assets is a list of sources for files (primarily when not using everything containerized)
	// Formats:
	//  raw url: http://... or https://...
	//  url with hash: <hex>@http://... or <hex>@https://...
	Assets map[architectures.Architecture][]*mirrors.MirroredAsset

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
	TaskMap map[string]fi.Task

	// ImageAssets are the image assets we use (output).
	ImageAssets []*assets.ImageAsset
	// FileAssets are the file assets we use (output).
	FileAssets []*assets.FileAsset

	// AdditionalObjects holds cluster-asssociated configuration objects, other than the Cluster and InstanceGroups.
	AdditionalObjects kubemanifest.ObjectList
}

func (c *ApplyClusterCmd) Run(ctx context.Context) error {
	if c.TargetName == TargetTerraform {
		found := false
		for _, cp := range TerraformCloudProviders {
			if c.Cloud.ProviderID() == cp {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("cloud provider %v does not support the terraform target", c.Cloud.ProviderID())
		}
	}
	if c.InstanceGroups == nil {
		list, err := c.Clientset.InstanceGroupsFor(c.Cluster).List(ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}
		var instanceGroups []*kops.InstanceGroup
		for i := range list.Items {
			instanceGroups = append(instanceGroups, &list.Items[i])
		}
		c.InstanceGroups = instanceGroups
	}

	if c.AdditionalObjects == nil {
		additionalObjects, err := c.Clientset.AddonsFor(c.Cluster).List()
		if err != nil {
			return err
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

	channel, err := ChannelForCluster(c.Cluster)
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
		return fmt.Errorf("unknown phase %q", c.Phase)
	}
	if c.GetAssets {
		networkLifecycle = fi.LifecycleIgnore
		securityLifecycle = fi.LifecycleIgnore
		clusterLifecycle = fi.LifecycleIgnore
	}

	assetBuilder := assets.NewAssetBuilder(c.Cluster, c.GetAssets)
	err = c.upgradeSpecs(assetBuilder)
	if err != nil {
		return err
	}

	err = c.validateKopsVersion()
	if err != nil {
		return err
	}

	err = c.validateKubernetesVersion()
	if err != nil {
		return err
	}

	cluster := c.Cluster

	configBase, err := vfs.Context.BuildVfsPath(cluster.Spec.ConfigBase)
	if err != nil {
		return fmt.Errorf("error parsing config base %q: %v", cluster.Spec.ConfigBase, err)
	}

	if !c.AllowKopsDowngrade {
		kopsVersionUpdatedBytes, err := configBase.Join(registry.PathKopsVersionUpdated).ReadFile()
		if err == nil {
			kopsVersionUpdated := strings.TrimSpace(string(kopsVersionUpdatedBytes))
			version, err := semver.Parse(kopsVersionUpdated)
			if err != nil {
				return fmt.Errorf("error parsing last kops version updated: %v", err)
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
				return fmt.Errorf("kops version older than last used to update the cluster")
			}
		} else if err != os.ErrNotExist {
			return fmt.Errorf("error reading last kops version used to update: %v", err)
		}
	}

	cloud := c.Cloud

	err = validation.DeepValidate(c.Cluster, c.InstanceGroups, true, cloud)
	if err != nil {
		return err
	}

	if cluster.Spec.KubernetesVersion == "" {
		return fmt.Errorf("KubernetesVersion not set")
	}
	if cluster.Spec.DNSZone == "" && !dns.IsGossipHostname(cluster.ObjectMeta.Name) {
		return fmt.Errorf("DNSZone not set")
	}

	l := &Loader{}
	l.Init()

	keyStore, err := c.Clientset.KeyStore(cluster)
	if err != nil {
		return err
	}

	sshCredentialStore, err := c.Clientset.SSHCredentialStore(cluster)
	if err != nil {
		return err
	}

	secretStore, err := c.Clientset.SecretStore(cluster)
	if err != nil {
		return err
	}

	addonsClient := c.Clientset.AddonsFor(cluster)
	addons, err := addonsClient.List()
	if err != nil {
		return fmt.Errorf("error fetching addons: %v", err)
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
	if fi.BoolValue(c.Cluster.Spec.EncryptionConfig) {
		secret, err := secretStore.FindSecret("encryptionconfig")
		if err != nil {
			return fmt.Errorf("could not load encryptionconfig secret: %v", err)
		}
		if secret == nil {
			fmt.Println("")
			fmt.Println("You have encryptionConfig enabled, but no encryptionconfig secret has been set.")
			fmt.Println("See `kops create secret encryptionconfig -h` and https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/")
			return fmt.Errorf("could not find encryptionconfig secret")
		}
		hashBytes := sha256.Sum256(secret.Data)
		encryptionConfigSecretHash = base64.URLEncoding.EncodeToString(hashBytes[:])
	}

	ciliumSpec := c.Cluster.Spec.Networking.Cilium
	if ciliumSpec != nil && ciliumSpec.EnableEncryption && ciliumSpec.EncryptionType == kops.CiliumEncryptionTypeIPSec {
		secret, err := secretStore.FindSecret("ciliumpassword")
		if err != nil {
			return fmt.Errorf("could not load the ciliumpassword secret: %w", err)
		}
		if secret == nil {
			fmt.Println("")
			fmt.Println("You have cilium encryption enabled, but no ciliumpassword secret has been set.")
			fmt.Println("See `kops create secret ciliumpassword -h`")
			return fmt.Errorf("could not find ciliumpassword secret")
		}
	}

	if err := c.addFileAssets(assetBuilder); err != nil {
		return err
	}

	checkExisting := true

	project := ""

	var sshPublicKeys [][]byte
	{
		keys, err := sshCredentialStore.FindSSHPublicKeys()
		if err != nil {
			return fmt.Errorf("error retrieving SSH public key %q: %v", fi.SecretNameSSHPrimary, err)
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

	switch cluster.Spec.GetCloudProvider() {
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
				return fmt.Errorf("SSH public key must be specified when running with DigitalOcean (create with `kops create secret --name %s sshpublickey admin -i ~/.ssh/id_rsa.pub`)", cluster.ObjectMeta.Name)
			}
		}
	case kops.CloudProviderAWS:
		{
			awsCloud := cloud.(awsup.AWSCloud)

			accountID, partition, err := awsCloud.AccountInfo()
			if err != nil {
				return err
			}
			modelContext.AWSAccountID = accountID
			modelContext.AWSPartition = partition

			if len(sshPublicKeys) > 1 {
				return fmt.Errorf("exactly one 'admin' SSH public key can be specified when running with AWS; please delete a key using `kops delete secret`")
			}
		}

	case kops.CloudProviderAzure:
		{
			if !featureflag.Azure.Enabled() {
				return fmt.Errorf("azure support is currently alpha, and is feature-gated. Please export KOPS_FEATURE_FLAGS=Azure")
			}

			if len(sshPublicKeys) == 0 {
				return fmt.Errorf("SSH public key must be specified when running with AzureCloud (create with `kops create secret --name %s sshpublickey admin -i ~/.ssh/id_rsa.pub`)", cluster.ObjectMeta.Name)
			}

			if len(sshPublicKeys) != 1 {
				return fmt.Errorf("exactly one 'admin' SSH public key can be specified when running with AzureCloud; please delete a key using `kops delete secret`")
			}
		}
	case kops.CloudProviderOpenstack:
		{
			if len(sshPublicKeys) == 0 {
				return fmt.Errorf("SSH public key must be specified when running with Openstack (create with `kops create secret --name %s sshpublickey admin -i ~/.ssh/id_rsa.pub`)", cluster.ObjectMeta.Name)
			}

			if len(sshPublicKeys) != 1 {
				return fmt.Errorf("exactly one 'admin' SSH public key can be specified when running with Openstack; please delete a key using `kops delete secret`")
			}
		}
	case kops.CloudProviderYandex:
		{
			if !featureflag.Yandex.Enabled() {
				return fmt.Errorf("yandex support is currently alpha, and is feature-gated. Please export KOPS_FEATURE_FLAGS=Yandex")
			}
		}
	default:
		return fmt.Errorf("unknown CloudProvider %q", cluster.Spec.GetCloudProvider())
	}

	modelContext.SSHPublicKeys = sshPublicKeys
	modelContext.Region = cloud.Region()

	if dns.IsGossipHostname(cluster.ObjectMeta.Name) {
		klog.V(2).Infof("Gossip DNS: skipping DNS validation")
	} else {
		err = validateDNS(cluster, cloud)
		if err != nil {
			return err
		}
	}

	tf := &TemplateFunctions{
		KopsModelContext: *modelContext,
		cloud:            cloud,
	}

	configBuilder, err := newNodeUpConfigBuilder(cluster, assetBuilder, c.Assets, encryptionConfigSecretHash)
	if err != nil {
		return err
	}
	bootstrapScriptBuilder := &model.BootstrapScriptBuilder{
		KopsModelContext:    modelContext,
		Lifecycle:           clusterLifecycle,
		NodeUpConfigBuilder: configBuilder,
		NodeUpAssets:        c.NodeUpAssets,
		Cluster:             cluster,
	}

	{
		templates, err := templates.LoadTemplates(cluster, models.NewAssetPath("cloudup/resources"))
		if err != nil {
			return fmt.Errorf("error loading templates: %v", err)
		}

		err = tf.AddTo(templates.TemplateFunctions, secretStore)
		if err != nil {
			return err
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

		switch cluster.Spec.GetCloudProvider() {
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
				&awsmodel.OIDCProviderBuilder{AWSModelContext: awsModelContext, Lifecycle: securityLifecycle, KeyStore: keyStore},
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

			nth := c.Cluster.Spec.NodeTerminationHandler
			if nth != nil && fi.BoolValue(nth.Enabled) && fi.BoolValue(nth.EnableSQSTerminationDraining) {
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
		case kops.CloudProviderYandex:
			yandexModelContext := &yandexmodel.YandexModelContext{
				KopsModelContext: modelContext,
			}
			l.Builders = append(l.Builders,
				&yandexmodel.NetworkModelBuilder{YandexModelContext: yandexModelContext, Lifecycle: networkLifecycle},
				&yandexmodel.SubnetModelBuilder{YandexModelContext: yandexModelContext, Lifecycle: networkLifecycle},
				&yandexmodel.InstanceModelBuilder{YandexModelContext: yandexModelContext, BootstrapScriptBuilder: bootstrapScriptBuilder, Lifecycle: clusterLifecycle},
				&yandexmodel.APILoadBalancerModelBuilder{YandexModelContext: yandexModelContext, Lifecycle: networkLifecycle},
			)
		default:
			return fmt.Errorf("unknown cloudprovider %q", cluster.Spec.GetCloudProvider())
		}
	}
	c.TaskMap, err = l.BuildTasks(c.LifecycleOverrides)
	if err != nil {
		return fmt.Errorf("error building tasks: %v", err)
	}

	var target fi.Target
	shouldPrecreateDNS := true

	switch c.TargetName {
	case TargetDirect:
		switch cluster.Spec.GetCloudProvider() {
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
		case kops.CloudProviderYandex:
			target = yandex.NewYandexAPITarget(cloud.(yandex.YandexCloud))
		default:
			return fmt.Errorf("direct configuration not supported with CloudProvider:%q", cluster.Spec.GetCloudProvider())
		}

	case TargetTerraform:
		checkExisting = false
		outDir := c.OutDir
		var vfsProvider *vfs.TerraformProvider
		if tfPath, ok := configBase.(vfs.TerraformPath); ok && featureflag.TerraformManagedFiles.Enabled() {
			var err error
			vfsProvider, err = tfPath.TerraformProvider()
			if err != nil {
				return err
			}
		}
		tf := terraform.NewTerraformTarget(cloud, project, vfsProvider, outDir, cluster.Spec.Target)

		// We include a few "util" variables in the TF output
		if err := tf.AddOutputVariable("region", terraformWriter.LiteralFromStringValue(cloud.Region())); err != nil {
			return err
		}

		if project != "" {
			if err := tf.AddOutputVariable("project", terraformWriter.LiteralFromStringValue(project)); err != nil {
				return err
			}
		}

		if err := tf.AddOutputVariable("cluster_name", terraformWriter.LiteralFromStringValue(cluster.ObjectMeta.Name)); err != nil {
			return err
		}

		target = tf

		// Can cause conflicts with terraform management
		shouldPrecreateDNS = false

	case TargetCloudformation:
		checkExisting = false
		outDir := c.OutDir
		target = cloudformation.NewCloudformationTarget(cloud, project, outDir)

		// Can cause conflicts with cloudformation management
		shouldPrecreateDNS = false

		fmt.Printf("\n")
		fmt.Printf("%s\n", starline)
		fmt.Printf("\n")
		fmt.Printf("Kops support for CloudFormation is deprecated and will be removed in a future release.\n")
		fmt.Printf("\n")
		fmt.Printf("%s\n", starline)
		fmt.Printf("\n")

	case TargetDryRun:
		var out io.Writer = os.Stdout
		if c.GetAssets {
			out = io.Discard
		}
		target = fi.NewDryRunTarget(assetBuilder, out)

		// Avoid making changes on a dry-run
		shouldPrecreateDNS = false

	default:
		return fmt.Errorf("unsupported target type %q", c.TargetName)
	}
	c.Target = target

	if checkExisting {
		c.TaskMap, err = l.FindDeletions(cloud, c.LifecycleOverrides)
		if err != nil {
			return fmt.Errorf("error finding deletions: %w", err)
		}
	}

	context, err := fi.NewContext(target, cluster, cloud, keyStore, secretStore, configBase, checkExisting, c.TaskMap)
	if err != nil {
		return fmt.Errorf("error building context: %v", err)
	}
	defer context.Close()

	var options fi.RunTasksOptions
	if c.RunTasksOptions != nil {
		options = *c.RunTasksOptions
	} else {
		options.InitDefaults()
	}

	err = context.RunTasks(options)
	if err != nil {
		return fmt.Errorf("error running tasks: %v", err)
	}

	if dns.IsGossipHostname(cluster.Name) {
		shouldPrecreateDNS = false
	}

	if shouldPrecreateDNS && clusterLifecycle != fi.LifecycleIgnore {
		if err := precreateDNS(ctx, cluster, cloud); err != nil {
			klog.Warningf("unable to pre-create DNS records - cluster startup may be slower: %v", err)
		}
	}

	err = target.Finish(c.TaskMap) // This will finish the apply, and print the changes
	if err != nil {
		return fmt.Errorf("error closing target: %v", err)
	}

	c.ImageAssets = assetBuilder.ImageAssets
	c.FileAssets = assetBuilder.FileAssets

	return nil
}

// upgradeSpecs ensures that fields are fully populated / defaulted
func (c *ApplyClusterCmd) upgradeSpecs(assetBuilder *assets.AssetBuilder) error {
	fullCluster, err := PopulateClusterSpec(c.Clientset, c.Cluster, c.Cloud, assetBuilder)
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
			fmt.Printf("\n")
			fmt.Printf("%s\n", starline)
			fmt.Printf("\n")
			fmt.Printf("This version of kubernetes is not yet supported; upgrading kops is required\n")
			fmt.Printf("(you can bypass this check by exporting KOPS_RUN_TOO_NEW_VERSION)\n")
			fmt.Printf("\n")
			fmt.Printf("%s\n", starline)
			fmt.Printf("\n")
			if os.Getenv("KOPS_RUN_TOO_NEW_VERSION") == "" {
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

// addFileAssets adds the file assets within the assetBuilder
func (c *ApplyClusterCmd) addFileAssets(assetBuilder *assets.AssetBuilder) error {
	var baseURL string
	if components.IsBaseURL(c.Cluster.Spec.KubernetesVersion) {
		baseURL = c.Cluster.Spec.KubernetesVersion
	} else {
		baseURL = "https://storage.googleapis.com/kubernetes-release/release/v" + c.Cluster.Spec.KubernetesVersion
	}

	c.Assets = make(map[architectures.Architecture][]*mirrors.MirroredAsset)
	c.NodeUpAssets = make(map[architectures.Architecture]*mirrors.MirroredAsset)
	for _, arch := range architectures.GetSupported() {
		c.Assets[arch] = []*mirrors.MirroredAsset{}

		k8sAssetsNames := []string{
			fmt.Sprintf("/bin/linux/%s/kubelet", arch),
			fmt.Sprintf("/bin/linux/%s/kubectl", arch),
		}

		if needsMounterAsset(c.Cluster, c.InstanceGroups) {
			k8sAssetsNames = append(k8sAssetsNames, fmt.Sprintf("/bin/linux/%s/mounter", arch))
		}

		for _, an := range k8sAssetsNames {
			k, err := url.Parse(baseURL)
			if err != nil {
				return err
			}
			k.Path = path.Join(k.Path, an)

			u, hash, err := assetBuilder.RemapFileAndSHA(k)
			if err != nil {
				return err
			}
			c.Assets[arch] = append(c.Assets[arch], mirrors.BuildMirroredAsset(u, hash))
		}

		cniAsset, cniAssetHash, err := findCNIAssets(c.Cluster, assetBuilder, arch)
		if err != nil {
			return err
		}
		c.Assets[arch] = append(c.Assets[arch], mirrors.BuildMirroredAsset(cniAsset, cniAssetHash))

		var containerRuntimeAssetUrl *url.URL
		var containerRuntimeAssetHash *hashing.Hash
		switch c.Cluster.Spec.ContainerRuntime {
		case "docker":
			containerRuntimeAssetUrl, containerRuntimeAssetHash, err = findDockerAsset(c.Cluster, assetBuilder, arch)
		case "containerd":
			containerRuntimeAssetUrl, containerRuntimeAssetHash, err = findContainerdAsset(c.Cluster, assetBuilder, arch)
		default:
			err = fmt.Errorf("unknown container runtime: %q", c.Cluster.Spec.ContainerRuntime)
		}
		if err != nil {
			return err
		}
		c.Assets[arch] = append(c.Assets[arch], mirrors.BuildMirroredAsset(containerRuntimeAssetUrl, containerRuntimeAssetHash))

		if c.Cluster.Spec.ContainerRuntime == "containerd" {
			var runcAssetUrl *url.URL
			var runcAssetHash *hashing.Hash
			runcAssetUrl, runcAssetHash, err = findRuncAsset(c.Cluster, assetBuilder, arch)
			if err != nil {
				return err
			}
			if runcAssetUrl != nil && runcAssetHash != nil {
				c.Assets[arch] = append(c.Assets[arch], mirrors.BuildMirroredAsset(runcAssetUrl, runcAssetHash))
			}
		}

		asset, err := NodeUpAsset(assetBuilder, arch)
		if err != nil {
			return err
		}
		c.NodeUpAssets[arch] = asset
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

func ChannelForCluster(c *kops.Cluster) (*kops.Channel, error) {
	channelLocation := c.Spec.Channel
	if channelLocation == "" {
		channelLocation = kops.DefaultChannel
	}
	return kops.LoadChannel(channelLocation)
}

// needsMounterAsset checks if we need the mounter program
// This is only needed currently on ContainerOS i.e. GCE, but we don't have a nice way to detect it yet
func needsMounterAsset(c *kops.Cluster, instanceGroups []*kops.InstanceGroup) bool {
	// TODO: Do real detection of ContainerOS (but this has to work with image names, and maybe even forked images)
	switch c.Spec.GetCloudProvider() {
	case kops.CloudProviderGCE:
		return true
	default:
		return false
	}
}

type nodeUpConfigBuilder struct {
	// Assets is a list of sources for files (primarily when not using everything containerized)
	// Formats:
	//  raw url: http://... or https://...
	//  url with hash: <hex>@http://... or <hex>@https://...
	assets map[architectures.Architecture][]*mirrors.MirroredAsset

	assetBuilder               *assets.AssetBuilder
	channels                   []string
	configBase                 vfs.Path
	cluster                    *kops.Cluster
	etcdManifests              map[string][]string
	images                     map[kops.InstanceGroupRole]map[architectures.Architecture][]*nodeup.Image
	protokubeAsset             map[architectures.Architecture][]*mirrors.MirroredAsset
	channelsAsset              map[architectures.Architecture][]*mirrors.MirroredAsset
	encryptionConfigSecretHash string
}

func newNodeUpConfigBuilder(cluster *kops.Cluster, assetBuilder *assets.AssetBuilder, assets map[architectures.Architecture][]*mirrors.MirroredAsset, encryptionConfigSecretHash string) (model.NodeUpConfigBuilder, error) {
	configBase, err := vfs.Context.BuildVfsPath(cluster.Spec.ConfigBase)
	if err != nil {
		return nil, fmt.Errorf("error parsing config base %q: %v", cluster.Spec.ConfigBase, err)
	}

	channels := []string{
		configBase.Join("addons", "bootstrap-channel.yaml").Path(),
	}

	for i := range cluster.Spec.Addons {
		channels = append(channels, cluster.Spec.Addons[i].Manifest)
	}

	etcdManifests := map[string][]string{}
	images := map[kops.InstanceGroupRole]map[architectures.Architecture][]*nodeup.Image{}
	protokubeAsset := map[architectures.Architecture][]*mirrors.MirroredAsset{}
	channelsAsset := map[architectures.Architecture][]*mirrors.MirroredAsset{}

	for _, arch := range architectures.GetSupported() {
		asset, err := ProtokubeAsset(assetBuilder, arch)
		if err != nil {
			return nil, err
		}
		protokubeAsset[arch] = append(protokubeAsset[arch], asset)
	}

	for _, arch := range architectures.GetSupported() {
		asset, err := ChannelsAsset(assetBuilder, arch)
		if err != nil {
			return nil, err
		}
		channelsAsset[arch] = append(channelsAsset[arch], asset)
	}

	for _, role := range kops.AllInstanceGroupRoles {
		isMaster := role == kops.InstanceGroupRoleMaster
		isAPIServer := role == kops.InstanceGroupRoleAPIServer

		images[role] = make(map[architectures.Architecture][]*nodeup.Image)
		if components.IsBaseURL(cluster.Spec.KubernetesVersion) {
			// When using a custom version, we want to preload the images over http
			components := []string{"kube-proxy"}
			if isMaster {
				components = append(components, "kube-apiserver", "kube-controller-manager", "kube-scheduler")
			}
			if isAPIServer {
				components = append(components, "kube-apiserver")
			}

			for _, arch := range architectures.GetSupported() {
				for _, component := range components {
					baseURL, err := url.Parse(cluster.Spec.KubernetesVersion)
					if err != nil {
						return nil, err
					}

					baseURL.Path = path.Join(baseURL.Path, "/bin/linux", string(arch), component+".tar")

					u, hash, err := assetBuilder.RemapFileAndSHA(baseURL)
					if err != nil {
						return nil, err
					}

					image := &nodeup.Image{
						Sources: []string{u.String()},
						Hash:    hash.Hex(),
					}
					images[role][arch] = append(images[role][arch], image)
				}
			}
		}

		// `docker load` our images when using a KOPS_BASE_URL, so we
		// don't need to push/pull from a registry
		if os.Getenv("KOPS_BASE_URL") != "" && isMaster {
			for _, arch := range architectures.GetSupported() {
				for _, name := range []string{"kops-controller", "dns-controller", "kube-apiserver-healthcheck"} {
					baseURL, err := url.Parse(os.Getenv("KOPS_BASE_URL"))
					if err != nil {
						return nil, err
					}

					baseURL.Path = path.Join(baseURL.Path, "/images/"+name+"-"+string(arch)+".tar.gz")

					u, hash, err := assetBuilder.RemapFileAndSHA(baseURL)
					if err != nil {
						return nil, err
					}

					image := &nodeup.Image{
						Sources: []string{u.String()},
						Hash:    hash.Hex(),
					}
					images[role][arch] = append(images[role][arch], image)
				}
			}
		}
		if os.Getenv("KOPS_BASE_URL") != "" && isAPIServer {
			for _, arch := range architectures.GetSupported() {
				for _, name := range []string{"kube-apiserver-healthcheck"} {
					baseURL, err := url.Parse(os.Getenv("KOPS_BASE_URL"))
					if err != nil {
						return nil, err
					}

					baseURL.Path = path.Join(baseURL.Path, "/images/"+name+"-"+string(arch)+".tar.gz")

					u, hash, err := assetBuilder.RemapFileAndSHA(baseURL)
					if err != nil {
						return nil, err
					}

					image := &nodeup.Image{
						Sources: []string{u.String()},
						Hash:    hash.Hex(),
					}
					images[role][arch] = append(images[role][arch], image)
				}
			}
		}

		if isMaster {
			for _, etcdCluster := range cluster.Spec.EtcdClusters {
				for _, member := range etcdCluster.Members {
					instanceGroup := fi.StringValue(member.InstanceGroup)
					etcdManifest := fmt.Sprintf("manifests/etcd/%s-%s.yaml", etcdCluster.Name, instanceGroup)
					etcdManifests[instanceGroup] = append(etcdManifests[instanceGroup], configBase.Join(etcdManifest).Path())
				}
			}
		}
	}

	configBuilder := nodeUpConfigBuilder{
		assetBuilder:               assetBuilder,
		assets:                     assets,
		channels:                   channels,
		configBase:                 configBase,
		cluster:                    cluster,
		etcdManifests:              etcdManifests,
		images:                     images,
		protokubeAsset:             protokubeAsset,
		channelsAsset:              channelsAsset,
		encryptionConfigSecretHash: encryptionConfigSecretHash,
	}

	return &configBuilder, nil
}

// BuildConfig returns the NodeUp config and auxiliary config.
func (n *nodeUpConfigBuilder) BuildConfig(ig *kops.InstanceGroup, apiserverAdditionalIPs []string, keysets map[string]*fi.Keyset) (*nodeup.Config, *nodeup.BootConfig, error) {
	cluster := n.cluster

	if ig == nil {
		return nil, nil, fmt.Errorf("instanceGroup cannot be nil")
	}

	role := ig.Spec.Role
	if role == "" {
		return nil, nil, fmt.Errorf("cannot determine role for instance group: %v", ig.ObjectMeta.Name)
	}

	useGossip := dns.IsGossipHostname(cluster.Spec.MasterInternalName)
	isMaster := role == kops.InstanceGroupRoleMaster
	hasAPIServer := isMaster || role == kops.InstanceGroupRoleAPIServer

	config, bootConfig := nodeup.NewConfig(cluster, ig)

	config.Assets = make(map[architectures.Architecture][]string)
	for _, arch := range architectures.GetSupported() {
		config.Assets[arch] = []string{}
		for _, a := range n.assets[arch] {
			config.Assets[arch] = append(config.Assets[arch], a.CompactString())
		}
	}

	if role != kops.InstanceGroupRoleBastion {
		if err := loadCertificates(keysets, fi.CertificateIDCA, config, true); err != nil {
			return nil, nil, err
		}
		if keysets["etcd-clients-ca-cilium"] != nil {
			if err := loadCertificates(keysets, "etcd-clients-ca-cilium", config, hasAPIServer || apiModel.UseKopsControllerForNodeBootstrap(n.cluster)); err != nil {
				return nil, nil, err
			}
		}

		if isMaster {
			if err := loadCertificates(keysets, "etcd-clients-ca", config, true); err != nil {
				return nil, nil, err
			}
			for _, etcdCluster := range cluster.Spec.EtcdClusters {
				k := etcdCluster.Name
				if err := loadCertificates(keysets, "etcd-manager-ca-"+k, config, true); err != nil {
					return nil, nil, err
				}
				if err := loadCertificates(keysets, "etcd-peers-ca-"+k, config, true); err != nil {
					return nil, nil, err
				}
				if k != "events" && k != "main" {
					if err := loadCertificates(keysets, "etcd-clients-ca-"+k, config, true); err != nil {
						return nil, nil, err
					}
				}
			}
			config.KeypairIDs["service-account"] = keysets["service-account"].Primary.Id
		} else {
			if keysets["etcd-client-cilium"] != nil {
				config.KeypairIDs["etcd-client-cilium"] = keysets["etcd-client-cilium"].Primary.Id
			}
		}

		if hasAPIServer {
			if err := loadCertificates(keysets, "apiserver-aggregator-ca", config, true); err != nil {
				return nil, nil, err
			}
			if keysets["etcd-clients-ca"] != nil {
				if err := loadCertificates(keysets, "etcd-clients-ca", config, true); err != nil {
					return nil, nil, err
				}
			}
			config.KeypairIDs["service-account"] = keysets["service-account"].Primary.Id

			config.APIServerConfig.EncryptionConfigSecretHash = n.encryptionConfigSecretHash
			serviceAccountPublicKeys, err := keysets["service-account"].ToPublicKeys()
			if err != nil {
				return nil, nil, fmt.Errorf("encoding service-account keys: %w", err)
			}
			config.APIServerConfig.ServiceAccountPublicKeys = serviceAccountPublicKeys
		} else {
			for _, key := range []string{"kubelet", "kube-proxy", "kube-router"} {
				if keysets[key] != nil {
					config.KeypairIDs[key] = keysets[key].Primary.Id
				}
			}
		}

		if isMaster || useGossip {
			for _, arch := range architectures.GetSupported() {
				for _, a := range n.protokubeAsset[arch] {
					config.Assets[arch] = append(config.Assets[arch], a.CompactString())
				}
			}

			for _, arch := range architectures.GetSupported() {
				for _, a := range n.channelsAsset[arch] {
					config.Assets[arch] = append(config.Assets[arch], a.CompactString())
				}
			}
		}
	}

	useConfigServer := featureflag.KopsControllerStateStore.Enabled() && (role != kops.InstanceGroupRoleMaster)
	if useConfigServer {
		baseURL := url.URL{
			Scheme: "https",
			Host:   net.JoinHostPort("kops-controller.internal."+cluster.ObjectMeta.Name, strconv.Itoa(wellknownports.KopsControllerPort)),
			Path:   "/",
		}

		configServer := &nodeup.ConfigServerOptions{
			Server:         baseURL.String(),
			CACertificates: config.CAs[fi.CertificateIDCA],
		}

		bootConfig.ConfigServer = configServer
		delete(config.CAs, fi.CertificateIDCA)
	} else {
		bootConfig.ConfigBase = fi.String(n.configBase.Path())
	}

	if isMaster {
		config.ApiserverAdditionalIPs = apiserverAdditionalIPs
	}

	for _, manifest := range n.assetBuilder.StaticManifests {
		match := false
		for _, r := range manifest.Roles {
			if r == role {
				match = true
			}
		}

		if !match {
			continue
		}

		config.StaticManifests = append(config.StaticManifests, &nodeup.StaticManifest{
			Key:  manifest.Key,
			Path: manifest.Path,
		})
	}

	for _, staticFile := range n.assetBuilder.StaticFiles {
		match := false
		for _, r := range staticFile.Roles {
			if r == role {
				match = true
			}
		}

		if !match {
			continue
		}

		config.FileAssets = append(config.FileAssets, kops.FileAssetSpec{
			Content: staticFile.Content,
			Path:    staticFile.Path,
		})
	}

	config.Images = n.images[role]
	config.Channels = n.channels
	config.EtcdManifests = n.etcdManifests[ig.Name]

	if ig.Spec.Containerd != nil || cluster.Spec.ContainerRuntime == "containerd" {
		config.ContainerdConfig = n.buildContainerdConfig(ig)
	}

	if (cluster.Spec.Containerd != nil && cluster.Spec.Containerd.NvidiaGPU != nil) || (ig.Spec.Containerd != nil && ig.Spec.Containerd.NvidiaGPU != nil) {
		config.NvidiaGPU = n.buildNvidiaConfig(ig)
	}

	if ig.Spec.WarmPool != nil || cluster.Spec.WarmPool != nil {
		config.WarmPoolImages = n.buildWarmPoolImages(ig)
	}

	if ig.Spec.Packages != nil {
		config.Packages = ig.Spec.Packages
	}

	return config, bootConfig, nil
}

func loadCertificates(keysets map[string]*fi.Keyset, name string, config *nodeup.Config, includeKeypairID bool) error {
	keyset := keysets[name]
	if keyset == nil {
		return fmt.Errorf("key %q not found", name)
	}
	certificates, err := keyset.ToCertificateBytes()
	if err != nil {
		return fmt.Errorf("failed to read %q certificates: %w", name, err)
	}
	config.CAs[name] = string(certificates)
	if includeKeypairID {
		if keyset.Primary == nil {
			return fmt.Errorf("key %q did not have primary set", name)
		}
		config.KeypairIDs[name] = keyset.Primary.Id
	}
	return nil
}

// buildNvidiaConfig builds nvidia configuration for instance group
func (n *nodeUpConfigBuilder) buildNvidiaConfig(ig *kops.InstanceGroup) *kops.NvidiaGPUConfig {
	config := &kops.NvidiaGPUConfig{}
	if n.cluster.Spec.Containerd != nil && n.cluster.Spec.Containerd.NvidiaGPU != nil {
		config = n.cluster.Spec.Containerd.NvidiaGPU
	}

	if ig.Spec.Containerd != nil && ig.Spec.Containerd.NvidiaGPU != nil {
		reflectutils.JSONMergeStruct(&config, ig.Spec.Containerd.NvidiaGPU)
	}

	if config.DriverPackage == "" {
		config.DriverPackage = kops.NvidiaDefaultDriverPackage
	}

	return config
}

// buildContainerdConfig builds containerd configuration for instance. Instance group configuration will override cluster configuration
func (n *nodeUpConfigBuilder) buildContainerdConfig(ig *kops.InstanceGroup) *kops.ContainerdConfig {
	config := n.cluster.Spec.Containerd.DeepCopy()
	if ig.Spec.Containerd != nil {
		reflectutils.JSONMergeStruct(&config, ig.Spec.Containerd)
	}
	return config
}

// buildWarmPoolImages returns a list of container images that should be pre-pulled during instance pre-initialization
func (n *nodeUpConfigBuilder) buildWarmPoolImages(ig *kops.InstanceGroup) []string {
	if ig == nil || ig.Spec.Role == kops.InstanceGroupRoleMaster {
		return nil
	}

	images := map[string]bool{}

	// Add component and addon images that impact startup time
	// TODO: Exclude images that only run on control-plane nodes in a generic way
	desiredImagePrefixes := []string{
		// Ignore images hosted in private ECR repositories as containerd cannot actually pull these
		//"602401143452.dkr.ecr.us-west-2.amazonaws.com/", // Amazon VPC CNI
		// Ignore images hosted on docker.io until a solution for rate limiting is implemented
		//"docker.io/calico/",
		//"docker.io/cilium/",
		//"docker.io/cloudnativelabs/kube-router:",
		//"docker.io/weaveworks/",
		"registry.k8s.io/kube-proxy:",
		"registry.k8s.io/provider-aws/",
		"registry.k8s.io/sig-storage/csi-node-driver-registrar:",
		"registry.k8s.io/sig-storage/livenessprobe:",
		"quay.io/calico/",
		"quay.io/cilium/",
		"quay.io/coreos/flannel:",
		"quay.io/weaveworks/",
	}
	assetBuilder := n.assetBuilder
	if assetBuilder != nil {
		for _, image := range assetBuilder.ImageAssets {
			for _, prefix := range desiredImagePrefixes {
				if strings.HasPrefix(image.DownloadLocation, prefix) {
					images[image.DownloadLocation] = true
				}
			}
		}
	}

	var unique []string
	for image := range images {
		unique = append(unique, image)
	}
	sort.Strings(unique)

	return unique
}
