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
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/blang/semver/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	kopsbase "k8s.io/kops"
	"k8s.io/kops/pkg/acls"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/client/simple/vfsclientset"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/alimodel"
	"k8s.io/kops/pkg/model/awsmodel"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/pkg/model/components/etcdmanager"
	"k8s.io/kops/pkg/model/components/kubeapiserver"
	"k8s.io/kops/pkg/model/domodel"
	"k8s.io/kops/pkg/model/gcemodel"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/model/openstackmodel"
	"k8s.io/kops/pkg/model/spotinstmodel"
	"k8s.io/kops/pkg/resources/digitalocean"
	"k8s.io/kops/pkg/templates"
	"k8s.io/kops/upup/models"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/aliup"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/do"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/hashing"
	"k8s.io/kops/util/pkg/vfs"
)

const (
	starline = "*********************************************************************************"
)

var (
	// AlphaAllowDO is a feature flag that gates DigitalOcean support while it is alpha
	AlphaAllowDO = featureflag.New("AlphaAllowDO", featureflag.Bool(false))
	// AlphaAllowGCE is a feature flag that gates GCE support while it is alpha
	AlphaAllowGCE = featureflag.New("AlphaAllowGCE", featureflag.Bool(false))
	// AlphaAllowALI is a feature flag that gates aliyun support while it is alpha
	AlphaAllowALI = featureflag.New("AlphaAllowALI", featureflag.Bool(false))
	// OldestSupportedKubernetesVersion is the oldest kubernetes version that is supported in Kops
	OldestSupportedKubernetesVersion = "1.11.0"
	// OldestRecommendedKubernetesVersion is the oldest kubernetes version that is not deprecated in Kops
	OldestRecommendedKubernetesVersion = "1.13.0"
)

type ApplyClusterCmd struct {
	Cloud   fi.Cloud
	Cluster *kops.Cluster

	InstanceGroups []*kops.InstanceGroup

	// NodeUpSource is the location from which we download nodeup
	NodeUpSource map[architectures.Architecture]string

	// NodeUpHash is the sha hash
	NodeUpHash map[architectures.Architecture]string

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
	Assets map[architectures.Architecture][]*MirroredAsset

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

	// TaskMap is the map of tasks that we built (output)
	TaskMap map[string]fi.Task
}

func (c *ApplyClusterCmd) Run(ctx context.Context) error {
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

	stageAssetsLifecycle := fi.LifecycleSync
	securityLifecycle := fi.LifecycleSync
	networkLifecycle := fi.LifecycleSync
	clusterLifecycle := fi.LifecycleSync

	switch c.Phase {
	case Phase(""):
		// Everything ... the default

		// until we implement finding assets we need to Ignore them
		stageAssetsLifecycle = fi.LifecycleIgnore
	case PhaseStageAssets:
		networkLifecycle = fi.LifecycleIgnore
		securityLifecycle = fi.LifecycleIgnore
		clusterLifecycle = fi.LifecycleIgnore

	case PhaseNetwork:
		stageAssetsLifecycle = fi.LifecycleIgnore
		securityLifecycle = fi.LifecycleIgnore
		clusterLifecycle = fi.LifecycleIgnore

	case PhaseSecurity:
		stageAssetsLifecycle = fi.LifecycleIgnore
		networkLifecycle = fi.LifecycleExistsAndWarnIfChanges
		clusterLifecycle = fi.LifecycleIgnore

	case PhaseCluster:
		if c.TargetName == TargetDryRun {
			stageAssetsLifecycle = fi.LifecycleIgnore
			securityLifecycle = fi.LifecycleExistsAndWarnIfChanges
			networkLifecycle = fi.LifecycleExistsAndWarnIfChanges
		} else {
			stageAssetsLifecycle = fi.LifecycleIgnore
			networkLifecycle = fi.LifecycleExistsAndValidates
			securityLifecycle = fi.LifecycleExistsAndValidates
		}

	default:
		return fmt.Errorf("unknown phase %q", c.Phase)
	}

	// This is kinda a hack.  Need to move phases out of fi.  If we use Phase here we introduce a circular
	// go dependency.
	phase := string(c.Phase)
	assetBuilder := assets.NewAssetBuilder(c.Cluster, phase)
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
	}

	if err := c.addFileAssets(assetBuilder); err != nil {
		return err
	}

	// Only setup transfer of kops assets if using a FileRepository
	if c.Cluster.Spec.Assets != nil && c.Cluster.Spec.Assets.FileRepository != nil {
		if err := SetKopsAssetsLocations(assetBuilder); err != nil {
			return err
		}
	}

	checkExisting := true

	region := ""
	project := ""

	var sshPublicKeys [][]byte
	{
		keys, err := sshCredentialStore.FindSSHPublicKeys(fi.SecretNameSSHPrimary)
		if err != nil {
			return fmt.Errorf("error retrieving SSH public key %q: %v", fi.SecretNameSSHPrimary, err)
		}

		for _, k := range keys {
			sshPublicKeys = append(sshPublicKeys, []byte(k.Spec.PublicKey))
		}
	}

	modelContext := &model.KopsModelContext{
		IAMModelContext: iam.IAMModelContext{Cluster: cluster},
		InstanceGroups:  c.InstanceGroups,
	}

	switch kops.CloudProviderID(cluster.Spec.CloudProvider) {
	case kops.CloudProviderGCE:
		{
			gceCloud := cloud.(gce.GCECloud)
			region = gceCloud.Region()
			project = gceCloud.Project()

			if !AlphaAllowGCE.Enabled() {
				return fmt.Errorf("GCE support is currently alpha, and is feature-gated.  export KOPS_FEATURE_FLAGS=AlphaAllowGCE")
			}

			modelContext.SSHPublicKeys = sshPublicKeys
		}

	case kops.CloudProviderDO:
		{
			if !AlphaAllowDO.Enabled() {
				return fmt.Errorf("DigitalOcean support is currently (very) alpha and is feature-gated. export KOPS_FEATURE_FLAGS=AlphaAllowDO to enable it")
			}

			if len(sshPublicKeys) == 0 && (c.Cluster.Spec.SSHKeyName == nil || *c.Cluster.Spec.SSHKeyName == "") {
				return fmt.Errorf("SSH public key must be specified when running with DigitalOcean (create with `kops create secret --name %s sshpublickey admin -i ~/.ssh/id_rsa.pub`)", cluster.ObjectMeta.Name)
			}

			modelContext.SSHPublicKeys = sshPublicKeys
		}
	case kops.CloudProviderAWS:
		{
			awsCloud := cloud.(awsup.AWSCloud)
			region = awsCloud.Region()

			accountID, partition, err := awsCloud.AccountInfo()
			if err != nil {
				return err
			}
			modelContext.AWSAccountID = accountID
			modelContext.AWSPartition = partition

			if len(sshPublicKeys) == 0 && c.Cluster.Spec.SSHKeyName == nil {
				return fmt.Errorf("SSH public key must be specified when running with AWS (create with `kops create secret --name %s sshpublickey admin -i ~/.ssh/id_rsa.pub`)", cluster.ObjectMeta.Name)
			}

			modelContext.SSHPublicKeys = sshPublicKeys

			if len(sshPublicKeys) > 1 {
				return fmt.Errorf("exactly one 'admin' SSH public key can be specified when running with AWS; please delete a key using `kops delete secret`")
			}
		}

	case kops.CloudProviderALI:
		{
			if !AlphaAllowALI.Enabled() {
				return fmt.Errorf("aliyun support is currently alpha, and is feature-gated.  export KOPS_FEATURE_FLAGS=AlphaAllowALI")
			}

			aliCloud := cloud.(aliup.ALICloud)
			region = aliCloud.Region()

			if len(sshPublicKeys) == 0 {
				return fmt.Errorf("SSH public key must be specified when running with ALICloud (create with `kops create secret --name %s sshpublickey admin -i ~/.ssh/id_rsa.pub`)", cluster.ObjectMeta.Name)
			}

			modelContext.SSHPublicKeys = sshPublicKeys

			if len(sshPublicKeys) != 1 {
				return fmt.Errorf("exactly one 'admin' SSH public key can be specified when running with ALICloud; please delete a key using `kops delete secret`")
			}
		}

	case kops.CloudProviderOpenstack:
		{

			osCloud := cloud.(openstack.OpenstackCloud)
			region = osCloud.Region()

			if len(sshPublicKeys) == 0 {
				return fmt.Errorf("SSH public key must be specified when running with Openstack (create with `kops create secret --name %s sshpublickey admin -i ~/.ssh/id_rsa.pub`)", cluster.ObjectMeta.Name)
			}

			modelContext.SSHPublicKeys = sshPublicKeys

			if len(sshPublicKeys) != 1 {
				return fmt.Errorf("exactly one 'admin' SSH public key can be specified when running with Openstack; please delete a key using `kops delete secret`")
			}
		}
	default:
		return fmt.Errorf("unknown CloudProvider %q", cluster.Spec.CloudProvider)
	}

	modelContext.Region = region

	if dns.IsGossipHostname(cluster.ObjectMeta.Name) {
		klog.Infof("Gossip DNS: skipping DNS validation")
	} else {
		err = validateDNS(cluster, cloud)
		if err != nil {
			return err
		}
	}

	tf := &TemplateFunctions{
		KopsModelContext: *modelContext,
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

		l.Builders = append(l.Builders,
			&BootstrapChannelBuilder{
				KopsModelContext: modelContext,
				Lifecycle:        &clusterLifecycle,
				assetBuilder:     assetBuilder,
				templates:        templates,
				ClusterAddons:    addons,
			},
			&model.PKIModelBuilder{
				KopsModelContext: modelContext,
				Lifecycle:        &clusterLifecycle,
			},
			&kubeapiserver.KubeApiserverBuilder{
				AssetBuilder:     assetBuilder,
				KopsModelContext: modelContext,
				Lifecycle:        &clusterLifecycle,
			},
			&etcdmanager.EtcdManagerBuilder{
				AssetBuilder:     assetBuilder,
				KopsModelContext: modelContext,
				Lifecycle:        &clusterLifecycle,
			},
		)

		switch kops.CloudProviderID(cluster.Spec.CloudProvider) {
		case kops.CloudProviderAWS:
			awsModelContext := &awsmodel.AWSModelContext{
				KopsModelContext: modelContext,
			}

			l.Builders = append(l.Builders,
				&model.MasterVolumeBuilder{KopsModelContext: modelContext, Lifecycle: &clusterLifecycle},
				&awsmodel.APILoadBalancerBuilder{AWSModelContext: awsModelContext, Lifecycle: &clusterLifecycle, SecurityLifecycle: &securityLifecycle},
				&model.BastionModelBuilder{KopsModelContext: modelContext, Lifecycle: &clusterLifecycle, SecurityLifecycle: &securityLifecycle},
				&model.DNSModelBuilder{KopsModelContext: modelContext, Lifecycle: &clusterLifecycle},
				&model.ExternalAccessModelBuilder{KopsModelContext: modelContext, Lifecycle: &securityLifecycle},
				&model.FirewallModelBuilder{KopsModelContext: modelContext, Lifecycle: &securityLifecycle},
				&model.SSHKeyModelBuilder{KopsModelContext: modelContext, Lifecycle: &securityLifecycle},
			)

			l.Builders = append(l.Builders,
				&model.NetworkModelBuilder{KopsModelContext: modelContext, Lifecycle: &networkLifecycle},
			)

			l.Builders = append(l.Builders,
				&model.IAMModelBuilder{KopsModelContext: modelContext, Lifecycle: &securityLifecycle},
				&awsmodel.OIDCProviderBuilder{KopsModelContext: modelContext, Lifecycle: &securityLifecycle, KeyStore: keyStore},
			)
		case kops.CloudProviderDO:
			doModelContext := &domodel.DOModelContext{
				KopsModelContext: modelContext,
			}
			l.Builders = append(l.Builders,
				&model.MasterVolumeBuilder{KopsModelContext: modelContext, Lifecycle: &clusterLifecycle},
				&domodel.APILoadBalancerModelBuilder{DOModelContext: doModelContext, Lifecycle: &securityLifecycle},
			)

		case kops.CloudProviderGCE:
			gceModelContext := &gcemodel.GCEModelContext{
				KopsModelContext: modelContext,
			}

			storageACLLifecycle := securityLifecycle
			if storageACLLifecycle != fi.LifecycleIgnore {
				// This is a best-effort permissions fix
				storageACLLifecycle = fi.LifecycleWarnIfInsufficientAccess
			}

			l.Builders = append(l.Builders,
				&model.MasterVolumeBuilder{KopsModelContext: modelContext, Lifecycle: &clusterLifecycle},

				&gcemodel.APILoadBalancerBuilder{GCEModelContext: gceModelContext, Lifecycle: &securityLifecycle},
				&gcemodel.ExternalAccessModelBuilder{GCEModelContext: gceModelContext, Lifecycle: &securityLifecycle},
				&gcemodel.FirewallModelBuilder{GCEModelContext: gceModelContext, Lifecycle: &securityLifecycle},
				&gcemodel.NetworkModelBuilder{GCEModelContext: gceModelContext, Lifecycle: &networkLifecycle},
			)

			l.Builders = append(l.Builders,
				&gcemodel.StorageAclBuilder{GCEModelContext: gceModelContext, Cloud: cloud.(gce.GCECloud), Lifecycle: &storageACLLifecycle},
			)

		case kops.CloudProviderALI:
			aliModelContext := &alimodel.ALIModelContext{
				KopsModelContext: modelContext,
			}
			l.Builders = append(l.Builders,
				&model.MasterVolumeBuilder{KopsModelContext: modelContext, Lifecycle: &clusterLifecycle},
				&alimodel.APILoadBalancerModelBuilder{ALIModelContext: aliModelContext, Lifecycle: &clusterLifecycle},
				&alimodel.NetworkModelBuilder{ALIModelContext: aliModelContext, Lifecycle: &clusterLifecycle},
				&alimodel.RAMModelBuilder{ALIModelContext: aliModelContext, Lifecycle: &clusterLifecycle},
				&alimodel.SSHKeyModelBuilder{ALIModelContext: aliModelContext, Lifecycle: &clusterLifecycle},
				&alimodel.FirewallModelBuilder{ALIModelContext: aliModelContext, Lifecycle: &clusterLifecycle},
				&alimodel.ExternalAccessModelBuilder{ALIModelContext: aliModelContext, Lifecycle: &clusterLifecycle},
			)

		case kops.CloudProviderOpenstack:
			openstackModelContext := &openstackmodel.OpenstackModelContext{
				KopsModelContext: modelContext,
			}

			l.Builders = append(l.Builders,
				&model.MasterVolumeBuilder{KopsModelContext: modelContext, Lifecycle: &clusterLifecycle},
				// &openstackmodel.APILBModelBuilder{OpenstackModelContext: openstackModelContext, Lifecycle: &clusterLifecycle},
				&openstackmodel.NetworkModelBuilder{OpenstackModelContext: openstackModelContext, Lifecycle: &networkLifecycle},
				&openstackmodel.SSHKeyModelBuilder{OpenstackModelContext: openstackModelContext, Lifecycle: &securityLifecycle},
				&openstackmodel.FirewallModelBuilder{OpenstackModelContext: openstackModelContext, Lifecycle: &securityLifecycle},
			)

		default:
			return fmt.Errorf("unknown cloudprovider %q", cluster.Spec.CloudProvider)
		}
	}

	configBuilder, err := c.newNodeUpConfigBuilder(assetBuilder)
	if err != nil {
		return err
	}
	bootstrapScriptBuilder := &model.BootstrapScriptBuilder{
		NodeUpConfigBuilder: configBuilder,
		NodeUpSource:        c.NodeUpSource,
		NodeUpSourceHash:    c.NodeUpHash,
	}
	switch kops.CloudProviderID(cluster.Spec.CloudProvider) {
	case kops.CloudProviderAWS:
		{
			awsModelContext := &awsmodel.AWSModelContext{
				KopsModelContext: modelContext,
			}

			awsModelBuilder := &awsmodel.AutoscalingGroupModelBuilder{
				AWSModelContext:        awsModelContext,
				BootstrapScriptBuilder: bootstrapScriptBuilder,
				Lifecycle:              &clusterLifecycle,
				SecurityLifecycle:      &securityLifecycle,
			}

			if featureflag.Spotinst.Enabled() {
				l.Builders = append(l.Builders, &spotinstmodel.InstanceGroupModelBuilder{
					KopsModelContext:       modelContext,
					BootstrapScriptBuilder: bootstrapScriptBuilder,
					Lifecycle:              &clusterLifecycle,
					SecurityLifecycle:      &securityLifecycle,
				})

				if featureflag.SpotinstHybrid.Enabled() {
					l.Builders = append(l.Builders, awsModelBuilder)
				}
			} else {
				l.Builders = append(l.Builders, awsModelBuilder)
			}
		}
	case kops.CloudProviderDO:
		doModelContext := &domodel.DOModelContext{
			KopsModelContext: modelContext,
		}

		l.Builders = append(l.Builders, &domodel.DropletBuilder{
			DOModelContext:         doModelContext,
			BootstrapScriptBuilder: bootstrapScriptBuilder,
			Lifecycle:              &clusterLifecycle,
		})
	case kops.CloudProviderGCE:
		{
			gceModelContext := &gcemodel.GCEModelContext{
				KopsModelContext: modelContext,
			}

			l.Builders = append(l.Builders, &gcemodel.AutoscalingGroupModelBuilder{
				GCEModelContext:        gceModelContext,
				BootstrapScriptBuilder: bootstrapScriptBuilder,
				Lifecycle:              &clusterLifecycle,
			})
		}

	case kops.CloudProviderALI:
		{
			aliModelContext := &alimodel.ALIModelContext{
				KopsModelContext: modelContext,
			}

			l.Builders = append(l.Builders, &alimodel.ScalingGroupModelBuilder{
				ALIModelContext:        aliModelContext,
				BootstrapScriptBuilder: bootstrapScriptBuilder,
				Lifecycle:              &clusterLifecycle,
			})
		}

	case kops.CloudProviderOpenstack:
		openstackModelContext := &openstackmodel.OpenstackModelContext{
			KopsModelContext: modelContext,
		}

		l.Builders = append(l.Builders, &openstackmodel.ServerGroupModelBuilder{
			OpenstackModelContext:  openstackModelContext,
			BootstrapScriptBuilder: bootstrapScriptBuilder,
			Lifecycle:              &clusterLifecycle,
		})

	default:
		return fmt.Errorf("unknown cloudprovider %q", cluster.Spec.CloudProvider)
	}

	taskMap, err := l.BuildTasks(assetBuilder, &stageAssetsLifecycle, c.LifecycleOverrides)
	if err != nil {
		return fmt.Errorf("error building tasks: %v", err)
	}

	c.TaskMap = taskMap

	var target fi.Target
	dryRun := false
	shouldPrecreateDNS := true

	switch c.TargetName {
	case TargetDirect:
		switch kops.CloudProviderID(cluster.Spec.CloudProvider) {
		case kops.CloudProviderGCE:
			target = gce.NewGCEAPITarget(cloud.(gce.GCECloud))
		case kops.CloudProviderAWS:
			target = awsup.NewAWSAPITarget(cloud.(awsup.AWSCloud))
		case kops.CloudProviderDO:
			target = do.NewDOAPITarget(cloud.(*digitalocean.Cloud))
		case kops.CloudProviderOpenstack:
			target = openstack.NewOpenstackAPITarget(cloud.(openstack.OpenstackCloud))
		case kops.CloudProviderALI:
			target = aliup.NewALIAPITarget(cloud.(aliup.ALICloud))
		default:
			return fmt.Errorf("direct configuration not supported with CloudProvider:%q", cluster.Spec.CloudProvider)
		}

	case TargetTerraform:
		checkExisting = false
		outDir := c.OutDir
		tf := terraform.NewTerraformTarget(cloud, region, project, outDir, cluster.Spec.Target)

		// We include a few "util" variables in the TF output
		if err := tf.AddOutputVariable("region", terraform.LiteralFromStringValue(region)); err != nil {
			return err
		}

		if project != "" {
			if err := tf.AddOutputVariable("project", terraform.LiteralFromStringValue(project)); err != nil {
				return err
			}
		}

		if err := tf.AddOutputVariable("cluster_name", terraform.LiteralFromStringValue(cluster.ObjectMeta.Name)); err != nil {
			return err
		}

		target = tf

		// Can cause conflicts with terraform management
		shouldPrecreateDNS = false

	case TargetCloudformation:
		checkExisting = false
		outDir := c.OutDir
		target = cloudformation.NewCloudformationTarget(cloud, region, project, outDir)

		// Can cause conflicts with cloudformation management
		shouldPrecreateDNS = false

	case TargetDryRun:
		target = fi.NewDryRunTarget(assetBuilder, os.Stdout)
		dryRun = true

		// Avoid making changes on a dry-run
		shouldPrecreateDNS = false

	default:
		return fmt.Errorf("unsupported target type %q", c.TargetName)
	}
	c.Target = target

	if !dryRun {
		acl, err := acls.GetACL(configBase, cluster)
		if err != nil {
			return err
		}
		err = configBase.Join(registry.PathKopsVersionUpdated).WriteFile(bytes.NewReader([]byte(kopsbase.Version)), acl)
		if err != nil {
			return fmt.Errorf("error writing kops version: %v", err)
		}

		err = registry.WriteConfigDeprecated(cluster, configBase.Join(registry.PathClusterCompleted), c.Cluster)
		if err != nil {
			return fmt.Errorf("error writing completed cluster spec: %v", err)
		}

		vfsMirror := vfsclientset.NewInstanceGroupMirror(cluster, configBase)

		for _, g := range c.InstanceGroups {
			// TODO: We need to update the mirror (below), but do we need to update the primary?
			_, err := c.Clientset.InstanceGroupsFor(c.Cluster).Update(ctx, g, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("error writing InstanceGroup %q to registry: %v", g.ObjectMeta.Name, err)
			}

			// TODO: Don't write if vfsMirror == c.ClientSet
			if err := vfsMirror.WriteMirror(g); err != nil {
				return fmt.Errorf("error writing instance group spec to mirror: %v", err)
			}
		}
	}

	context, err := fi.NewContext(target, cluster, cloud, keyStore, secretStore, configBase, checkExisting, taskMap)
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

	if shouldPrecreateDNS {
		if err := precreateDNS(ctx, cluster, cloud); err != nil {
			klog.Warningf("unable to pre-create DNS records - cluster startup may be slower: %v", err)
		}
	}

	err = target.Finish(taskMap) //This will finish the apply, and print the changes
	if err != nil {
		return fmt.Errorf("error closing target: %v", err)
	}

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

	if recommended != nil && !required {
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
	if !util.IsKubernetesGTE(OldestRecommendedKubernetesVersion, *parsed) {
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

	if recommended != nil && !required {
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

	c.Assets = make(map[architectures.Architecture][]*MirroredAsset)
	c.NodeUpSource = make(map[architectures.Architecture]string)
	c.NodeUpHash = make(map[architectures.Architecture]string)
	for _, arch := range architectures.GetSupported() {
		c.Assets[arch] = []*MirroredAsset{}
		c.NodeUpSource[arch] = ""
		c.NodeUpHash[arch] = ""

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
			c.Assets[arch] = append(c.Assets[arch], BuildMirroredAsset(u, hash))
		}

		cniAsset, cniAssetHash, err := findCNIAssets(c.Cluster, assetBuilder, arch)
		if err != nil {
			return err
		}
		c.Assets[arch] = append(c.Assets[arch], BuildMirroredAsset(cniAsset, cniAssetHash))

		if c.Cluster.Spec.Networking.LyftVPC != nil {
			lyftAsset, lyftAssetHash, err := findLyftVPCAssets(c.Cluster, assetBuilder, arch)
			if err != nil {
				return err
			}
			c.Assets[arch] = append(c.Assets[arch], BuildMirroredAsset(lyftAsset, lyftAssetHash))
		}

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
		c.Assets[arch] = append(c.Assets[arch], BuildMirroredAsset(containerRuntimeAssetUrl, containerRuntimeAssetHash))

		asset, err := NodeUpAsset(assetBuilder, arch)
		if err != nil {
			return err
		}
		c.NodeUpSource[arch] = strings.Join(asset.Locations, ",")
		c.NodeUpHash[arch] = asset.Hash.Hex()

		// Explicitly add the protokube image,
		// otherwise when the Target is DryRun this asset is not added
		// Is there a better way to call this?
		_, _, err = ProtokubeImageSource(assetBuilder, arch)
		if err != nil {
			return err
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
	switch kops.CloudProviderID(c.Spec.CloudProvider) {
	case kops.CloudProviderGCE:
		return true
	default:
		return false
	}
}

type nodeUpConfigBuilder struct {
	*ApplyClusterCmd
	assetBuilder   *assets.AssetBuilder
	channels       []string
	configBase     vfs.Path
	cluster        *kops.Cluster
	etcdManifests  map[kops.InstanceGroupRole][]string
	images         map[kops.InstanceGroupRole]map[architectures.Architecture][]*nodeup.Image
	protokubeImage map[kops.InstanceGroupRole]map[architectures.Architecture]*nodeup.Image
}

func (c *ApplyClusterCmd) newNodeUpConfigBuilder(assetBuilder *assets.AssetBuilder) (model.NodeUpConfigBuilder, error) {
	cluster := c.Cluster

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

	useGossip := dns.IsGossipHostname(cluster.Spec.MasterInternalName)

	etcdManifests := map[kops.InstanceGroupRole][]string{}
	images := map[kops.InstanceGroupRole]map[architectures.Architecture][]*nodeup.Image{}
	protokubeImage := map[kops.InstanceGroupRole]map[architectures.Architecture]*nodeup.Image{}

	for _, role := range kops.AllInstanceGroupRoles {
		isMaster := role == kops.InstanceGroupRoleMaster

		images[role] = make(map[architectures.Architecture][]*nodeup.Image)
		if components.IsBaseURL(cluster.Spec.KubernetesVersion) {
			// When using a custom version, we want to preload the images over http
			components := []string{"kube-proxy"}
			if isMaster {
				components = append(components, "kube-apiserver", "kube-controller-manager", "kube-scheduler")
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

		if isMaster || useGossip {
			protokubeImage[role] = make(map[architectures.Architecture]*nodeup.Image)
			for _, arch := range architectures.GetSupported() {
				u, hash, err := ProtokubeImageSource(assetBuilder, arch)
				if err != nil {
					return nil, err
				}

				asset := BuildMirroredAsset(u, hash)

				protokubeImage[role][arch] = &nodeup.Image{
					Name:    kopsbase.DefaultProtokubeImageName(),
					Sources: asset.Locations,
					Hash:    asset.Hash.Hex(),
				}
			}
		}

		if role == kops.InstanceGroupRoleMaster {
			for _, etcdCluster := range cluster.Spec.EtcdClusters {
				if etcdCluster.Provider == kops.EtcdProviderTypeManager {
					p := configBase.Join("manifests/etcd/" + etcdCluster.Name + ".yaml").Path()
					etcdManifests[role] = append(etcdManifests[role], p)
				}
			}
		}
	}

	configBuilder := nodeUpConfigBuilder{
		ApplyClusterCmd: c,
		assetBuilder:    assetBuilder,
		channels:        channels,
		configBase:      configBase,
		cluster:         cluster,
		etcdManifests:   etcdManifests,
		images:          images,
		protokubeImage:  protokubeImage,
	}
	return &configBuilder, nil
}

// BuildNodeUpConfig returns the NodeUp config, in YAML format
func (n *nodeUpConfigBuilder) BuildConfig(ig *kops.InstanceGroup, apiserverAdditionalIPs []string) (*nodeup.Config, error) {
	cluster := n.cluster

	if ig == nil {
		return nil, fmt.Errorf("instanceGroup cannot be nil")
	}

	role := ig.Spec.Role
	if role == "" {
		return nil, fmt.Errorf("cannot determine role for instance group: %v", ig.ObjectMeta.Name)
	}

	config := nodeup.NewConfig(cluster, ig)
	config.Assets = make(map[architectures.Architecture][]string)
	for _, arch := range architectures.GetSupported() {
		config.Assets[arch] = []string{}
		for _, a := range n.Assets[arch] {
			config.Assets[arch] = append(config.Assets[arch], a.CompactString())
		}
	}
	config.ClusterName = cluster.ObjectMeta.Name
	config.ConfigBase = fi.String(n.configBase.Path())
	config.InstanceGroupName = ig.ObjectMeta.Name

	if role == kops.InstanceGroupRoleMaster {
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

	config.Images = n.images[role]
	config.Channels = n.channels
	config.EtcdManifests = n.etcdManifests[role]
	config.ProtokubeImage = n.protokubeImage[role]

	return config, nil
}
