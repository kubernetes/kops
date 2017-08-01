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

package cloudup

import (
	"fmt"
	"os"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kopsbase "k8s.io/kops"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/awsmodel"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/pkg/model/gcemodel"
	"k8s.io/kops/pkg/model/vspheremodel"
	"k8s.io/kops/pkg/templates"
	"k8s.io/kops/upup/models"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/vsphere"
	"k8s.io/kops/upup/pkg/fi/cloudup/vspheretasks"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/util/pkg/hashing"
	"k8s.io/kops/util/pkg/vfs"

	"github.com/blang/semver"
	"github.com/golang/glog"
)

const (
	DefaultMaxTaskDuration = 10 * time.Minute
	starline               = "*********************************************************************************\n"
)

var (
	// AlphaAllowDO is a feature flag that gates DigitalOcean support while it is alpha
	AlphaAllowDO = featureflag.New("AlphaAllowDO", featureflag.Bool(false))
	// AlphaAllowGCE is a feature flag that gates GCE support while it is alpha
	AlphaAllowGCE = featureflag.New("AlphaAllowGCE", featureflag.Bool(false))
	// AlphaAllowVsphere is a feature flag that gates vsphere support while it is alpha
	AlphaAllowVsphere = featureflag.New("AlphaAllowVsphere", featureflag.Bool(false))
	// CloudupModels a list of supported models
	CloudupModels = []string{"config", "proto", "cloudup"}
)

type ApplyClusterCmd struct {
	Cluster *kops.Cluster

	InstanceGroups []*kops.InstanceGroup

	// NodeUpSource is the location from which we download nodeup
	NodeUpSource string

	// Models is a list of cloudup models to apply
	Models []string

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
	Assets []string

	Clientset simple.Clientset

	// DryRun is true if this is only a dry run
	DryRun bool

	MaxTaskDuration time.Duration

	// The channel we are using
	channel *kops.Channel

	// Phase can be set to a Phase to run the specific subset of tasks, if we don't want to run everything
	Phase Phase
}

func (c *ApplyClusterCmd) Run() error {
	if c.MaxTaskDuration == 0 {
		c.MaxTaskDuration = DefaultMaxTaskDuration
	}

	if c.InstanceGroups == nil {
		list, err := c.Clientset.InstanceGroupsFor(c.Cluster).List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		var instanceGroups []*kops.InstanceGroup
		for i := range list.Items {
			instanceGroups = append(instanceGroups, &list.Items[i])
		}
		c.InstanceGroups = instanceGroups
	}

	if c.Models == nil {
		c.Models = CloudupModels
	}

	modelStore, err := findModelStore()
	if err != nil {
		return err
	}

	channel, err := ChannelForCluster(c.Cluster)
	if err != nil {
		return err
	}
	c.channel = channel

	assetBuilder := assets.NewAssetBuilder()
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

	err = validation.DeepValidate(c.Cluster, c.InstanceGroups, true)
	if err != nil {
		return err
	}

	cluster := c.Cluster

	if cluster.Spec.KubernetesVersion == "" {
		return fmt.Errorf("KubernetesVersion not set")
	}
	if cluster.Spec.DNSZone == "" && !dns.IsGossipHostname(cluster.ObjectMeta.Name) {
		return fmt.Errorf("DNSZone not set")
	}

	l := &Loader{}
	l.Init()
	l.Cluster = c.Cluster

	configBase, err := vfs.Context.BuildVfsPath(cluster.Spec.ConfigBase)
	if err != nil {
		return fmt.Errorf("error parsing config base %q: %v", cluster.Spec.ConfigBase, err)
	}

	keyStore, err := registry.KeyStore(cluster)
	if err != nil {
		return err
	}
	keyStore.(*fi.VFSCAStore).DryRun = c.DryRun

	secretStore, err := registry.SecretStore(cluster)
	if err != nil {
		return err
	}

	channels := []string{
		configBase.Join("addons", "bootstrap-channel.yaml").Path(),
	}

	// Normalize k8s version
	versionWithoutV := strings.TrimSpace(cluster.Spec.KubernetesVersion)
	if strings.HasPrefix(versionWithoutV, "v") {
		versionWithoutV = versionWithoutV[1:]
	}
	if cluster.Spec.KubernetesVersion != versionWithoutV {
		glog.Warningf("Normalizing kubernetes version: %q -> %q", cluster.Spec.KubernetesVersion, versionWithoutV)
		cluster.Spec.KubernetesVersion = versionWithoutV
	}

	if len(c.Assets) == 0 {
		var baseURL string
		if components.IsBaseURL(cluster.Spec.KubernetesVersion) {
			baseURL = cluster.Spec.KubernetesVersion
		} else {
			baseURL = "https://storage.googleapis.com/kubernetes-release/release/v" + cluster.Spec.KubernetesVersion
		}
		baseURL = strings.TrimSuffix(baseURL, "/")

		{
			defaultKubeletAsset := baseURL + "/bin/linux/amd64/kubelet"
			glog.V(2).Infof("Adding default kubelet release asset: %s", defaultKubeletAsset)

			hash, err := findHash(defaultKubeletAsset)
			if err != nil {
				return err
			}
			c.Assets = append(c.Assets, hash.Hex()+"@"+defaultKubeletAsset)
		}

		{
			defaultKubectlAsset := baseURL + "/bin/linux/amd64/kubectl"
			glog.V(2).Infof("Adding default kubectl release asset: %s", defaultKubectlAsset)

			hash, err := findHash(defaultKubectlAsset)
			if err != nil {
				return err
			}
			c.Assets = append(c.Assets, hash.Hex()+"@"+defaultKubectlAsset)
		}

		if usesCNI(cluster) {
			cniAsset, cniAssetHashString, err := findCNIAssets(cluster)

			if err != nil {
				return err
			}

			if cniAssetHashString == "" {
				glog.Warningf("cniAssetHashString is empty, using cniAsset directly: %s", cniAsset)
				c.Assets = append(c.Assets, cniAsset)
			} else {
				c.Assets = append(c.Assets, cniAssetHashString+"@"+cniAsset)
			}
		}

		if needsStaticUtils(cluster, c.InstanceGroups) {
			utilsLocation := BaseUrl() + "linux/amd64/utils.tar.gz"
			glog.V(4).Infof("Using default utils.tar.gz location: %q", utilsLocation)

			hash, err := findHash(utilsLocation)
			if err != nil {
				return err
			}
			c.Assets = append(c.Assets, hash.Hex()+"@"+utilsLocation)
		}
	}

	if c.NodeUpSource == "" {
		c.NodeUpSource = NodeUpLocation()
	}

	checkExisting := true

	l.AddTypes(map[string]interface{}{
		"keypair":     &fitasks.Keypair{},
		"secret":      &fitasks.Secret{},
		"managedFile": &fitasks.ManagedFile{},
	})

	cloud, err := BuildCloud(cluster)
	if err != nil {
		return err
	}

	region := ""
	project := ""

	var sshPublicKeys [][]byte
	{
		keys, err := keyStore.FindSSHPublicKeys(fi.SecretNameSSHPrimary)
		if err != nil {
			return fmt.Errorf("error retrieving SSH public key %q: %v", fi.SecretNameSSHPrimary, err)
		}

		for _, k := range keys {
			sshPublicKeys = append(sshPublicKeys, k.Data)
		}
	}

	modelContext := &model.KopsModelContext{
		Cluster:        cluster,
		InstanceGroups: c.InstanceGroups,
	}

	switch kops.CloudProviderID(cluster.Spec.CloudProvider) {
	case kops.CloudProviderGCE:
		{
			gceCloud := cloud.(*gce.GCECloud)
			region = gceCloud.Region
			project = gceCloud.Project

			if !AlphaAllowGCE.Enabled() {
				return fmt.Errorf("GCE support is currently alpha, and is feature-gated.  export KOPS_FEATURE_FLAGS=AlphaAllowGCE")
			}

			l.AddTypes(map[string]interface{}{
				"Disk":                 &gcetasks.Disk{},
				"Instance":             &gcetasks.Instance{},
				"InstanceTemplate":     &gcetasks.InstanceTemplate{},
				"Network":              &gcetasks.Network{},
				"InstanceGroupManager": &gcetasks.InstanceGroupManager{},
				"FirewallRule":         &gcetasks.FirewallRule{},
				"Address":              &gcetasks.Address{},
			})
		}

	case kops.CloudProviderDO:
		{
			if !AlphaAllowDO.Enabled() {
				return fmt.Errorf("DigitalOcean support is currently (very) alpha and is feature-gated. export KOPS_FEATURE_FLAGS=AlphaAllowDO to enable it")
			}

			// this is a no-op for now, add tasks to this list as more DO support is added
			l.AddTypes(map[string]interface{}{})
		}
	case kops.CloudProviderAWS:
		{
			awsCloud := cloud.(awsup.AWSCloud)
			region = awsCloud.Region()

			l.AddTypes(map[string]interface{}{
				// EC2
				"elasticIP":                   &awstasks.ElasticIP{},
				"instance":                    &awstasks.Instance{},
				"instanceElasticIPAttachment": &awstasks.InstanceElasticIPAttachment{},
				"instanceVolumeAttachment":    &awstasks.InstanceVolumeAttachment{},
				"ebsVolume":                   &awstasks.EBSVolume{},
				"sshKey":                      &awstasks.SSHKey{},

				// IAM
				"iamInstanceProfile":     &awstasks.IAMInstanceProfile{},
				"iamInstanceProfileRole": &awstasks.IAMInstanceProfileRole{},
				"iamRole":                &awstasks.IAMRole{},
				"iamRolePolicy":          &awstasks.IAMRolePolicy{},

				// VPC / Networking
				"dhcpOptions":           &awstasks.DHCPOptions{},
				"internetGateway":       &awstasks.InternetGateway{},
				"route":                 &awstasks.Route{},
				"routeTable":            &awstasks.RouteTable{},
				"routeTableAssociation": &awstasks.RouteTableAssociation{},
				"securityGroup":         &awstasks.SecurityGroup{},
				"securityGroupRule":     &awstasks.SecurityGroupRule{},
				"subnet":                &awstasks.Subnet{},
				"vpc":                   &awstasks.VPC{},
				"ngw":                   &awstasks.NatGateway{},
				"vpcDHDCPOptionsAssociation": &awstasks.VPCDHCPOptionsAssociation{},

				// ELB
				"loadBalancer":           &awstasks.LoadBalancer{},
				"loadBalancerAttachment": &awstasks.LoadBalancerAttachment{},

				// Autoscaling
				"autoscalingGroup":    &awstasks.AutoscalingGroup{},
				"launchConfiguration": &awstasks.LaunchConfiguration{},
			})

			if len(sshPublicKeys) == 0 {
				return fmt.Errorf("SSH public key must be specified when running with AWS (create with `kops create secret --name %s sshpublickey admin -i ~/.ssh/id_rsa.pub`)", cluster.ObjectMeta.Name)
			}

			modelContext.SSHPublicKeys = sshPublicKeys

			if len(sshPublicKeys) != 1 {
				return fmt.Errorf("Exactly one 'admin' SSH public key can be specified when running with AWS; please delete a key using `kops delete secret`")
			}

			l.TemplateFunctions["MachineTypeInfo"] = awsup.GetMachineTypeInfo
		}

	case kops.CloudProviderVSphere:
		{
			if !AlphaAllowVsphere.Enabled() {
				return fmt.Errorf("Vsphere support is currently alpha, and is feature-gated.  export KOPS_FEATURE_FLAGS=AlphaAllowVsphere")
			}

			vsphereCloud := cloud.(*vsphere.VSphereCloud)
			// TODO: map region with vCenter cluster, or datacenter, or datastore?
			region = vsphereCloud.Cluster

			l.AddTypes(map[string]interface{}{
				"instance": &vspheretasks.VirtualMachine{},
			})
		}

	default:
		return fmt.Errorf("unknown CloudProvider %q", cluster.Spec.CloudProvider)
	}

	modelContext.Region = region

	if dns.IsGossipHostname(cluster.ObjectMeta.Name) {
		glog.Infof("Gossip DNS: skipping DNS validation")
	} else {
		err = validateDNS(cluster, cloud)
		if err != nil {
			return err
		}
	}

	clusterTags, err := buildCloudupTags(cluster)
	if err != nil {
		return err
	}

	tf := &TemplateFunctions{
		cluster:        cluster,
		instanceGroups: c.InstanceGroups,
		tags:           clusterTags,
		region:         region,
		modelContext:   modelContext,
	}

	l.Tags = clusterTags
	l.WorkDir = c.OutDir
	l.ModelStore = modelStore

	iamLifecycle := lifecyclePointer(fi.LifecycleSync)
	networkLifecycle := lifecyclePointer(fi.LifecycleSync)
	clusterLifecycle := lifecyclePointer(fi.LifecycleSync)

	switch c.Phase {
	case Phase(""):
	// Everything ... the default

	case PhaseIAM:
		networkLifecycle = lifecyclePointer(fi.LifecycleIgnore)
		clusterLifecycle = lifecyclePointer(fi.LifecycleIgnore)

	case PhaseNetwork:
		iamLifecycle = lifecyclePointer(fi.LifecycleIgnore)
		clusterLifecycle = lifecyclePointer(fi.LifecycleIgnore)

	case PhaseCluster:
		if c.TargetName == TargetDryRun {
			iamLifecycle = lifecyclePointer(fi.LifecycleExistsAndWarnIfChanges)
			networkLifecycle = lifecyclePointer(fi.LifecycleExistsAndWarnIfChanges)
		} else {
			iamLifecycle = lifecyclePointer(fi.LifecycleExistsAndValidates)
			networkLifecycle = lifecyclePointer(fi.LifecycleExistsAndValidates)
		}
	default:
		return fmt.Errorf("unknown phase %q", c.Phase)
	}

	var fileModels []string
	for _, m := range c.Models {
		switch m {
		case "proto":
		// No proto code options; no file model

		case "cloudup":
			templates, err := templates.LoadTemplates(cluster, models.NewAssetPath("cloudup/resources"))
			if err != nil {
				return fmt.Errorf("error loading templates: %v", err)
			}
			tf.AddTo(templates.TemplateFunctions)

			l.Builders = append(l.Builders,
				&BootstrapChannelBuilder{
					cluster:      cluster,
					Lifecycle:    clusterLifecycle,
					templates:    templates,
					assetBuilder: assetBuilder,
				},
				&model.PKIModelBuilder{KopsModelContext: modelContext, Lifecycle: clusterLifecycle},
			)

			switch kops.CloudProviderID(cluster.Spec.CloudProvider) {
			case kops.CloudProviderAWS:
				awsModelContext := &awsmodel.AWSModelContext{
					KopsModelContext: modelContext,
				}

				l.Builders = append(l.Builders,
					&model.MasterVolumeBuilder{KopsModelContext: modelContext, Lifecycle: clusterLifecycle},

					&awsmodel.APILoadBalancerBuilder{AWSModelContext: awsModelContext, Lifecycle: networkLifecycle},
					&model.BastionModelBuilder{KopsModelContext: modelContext, Lifecycle: networkLifecycle},
					&model.DNSModelBuilder{KopsModelContext: modelContext, Lifecycle: networkLifecycle},
					&model.ExternalAccessModelBuilder{KopsModelContext: modelContext, Lifecycle: clusterLifecycle},
					&model.FirewallModelBuilder{KopsModelContext: modelContext, Lifecycle: clusterLifecycle},
					&model.SSHKeyModelBuilder{KopsModelContext: modelContext, Lifecycle: iamLifecycle},
				)

				l.Builders = append(l.Builders,
					&model.NetworkModelBuilder{KopsModelContext: modelContext, Lifecycle: networkLifecycle},
				)

				l.Builders = append(l.Builders,
					&model.IAMModelBuilder{KopsModelContext: modelContext, Lifecycle: iamLifecycle},
				)

			case kops.CloudProviderGCE:
				gceModelContext := &gcemodel.GCEModelContext{
					KopsModelContext: modelContext,
				}

				l.Builders = append(l.Builders,
					&model.MasterVolumeBuilder{KopsModelContext: modelContext, Lifecycle: clusterLifecycle},

					&gcemodel.APILoadBalancerBuilder{GCEModelContext: gceModelContext, Lifecycle: networkLifecycle},
					&gcemodel.ExternalAccessModelBuilder{GCEModelContext: gceModelContext, Lifecycle: networkLifecycle},
					&gcemodel.FirewallModelBuilder{GCEModelContext: gceModelContext, Lifecycle: networkLifecycle},
					&gcemodel.NetworkModelBuilder{GCEModelContext: gceModelContext, Lifecycle: networkLifecycle},
				)

			case kops.CloudProviderVSphere:
				// No special settings (yet!)

			default:
				return fmt.Errorf("unknown cloudprovider %q", cluster.Spec.CloudProvider)
			}

			fileModels = append(fileModels, m)

		default:
			fileModels = append(fileModels, m)
		}
	}

	l.TemplateFunctions["CA"] = func() fi.CAStore {
		return keyStore
	}
	l.TemplateFunctions["Secrets"] = func() fi.SecretStore {
		return secretStore
	}

	// RenderNodeUpConfig returns the NodeUp config, in YAML format
	// @@NOTE
	renderNodeUpConfig := func(ig *kops.InstanceGroup) (*nodeup.Config, error) {
		if ig == nil {
			return nil, fmt.Errorf("instanceGroup cannot be nil")
		}

		role := ig.Spec.Role
		if role == "" {
			return nil, fmt.Errorf("cannot determine role for instance group: %v", ig.ObjectMeta.Name)
		}

		nodeUpTags, err := buildNodeupTags(role, tf.cluster, tf.tags)
		if err != nil {
			return nil, err
		}

		config := &nodeup.Config{}
		for _, tag := range nodeUpTags.List() {
			config.Tags = append(config.Tags, tag)
		}

		config.Assets = c.Assets
		config.ClusterName = cluster.ObjectMeta.Name
		config.ConfigBase = fi.String(configBase.Path())
		config.InstanceGroupName = ig.ObjectMeta.Name

		var images []*nodeup.Image

		if components.IsBaseURL(cluster.Spec.KubernetesVersion) {
			baseURL := cluster.Spec.KubernetesVersion
			baseURL = strings.TrimSuffix(baseURL, "/")

			// TODO: pull kube-dns image
			// When using a custom version, we want to preload the images over http
			components := []string{"kube-proxy"}
			if role == kops.InstanceGroupRoleMaster {
				components = append(components, "kube-apiserver", "kube-controller-manager", "kube-scheduler")
			}
			for _, component := range components {
				imagePath := baseURL + "/bin/linux/amd64/" + component + ".tar"
				glog.Infof("Adding docker image: %s", imagePath)

				hash, err := findHash(imagePath)
				if err != nil {
					return nil, err
				}
				image := &nodeup.Image{
					Source: imagePath,
					Hash:   hash.Hex(),
				}
				images = append(images, image)
			}
		}

		{
			location := ProtokubeImageSource()

			hash, err := findHash(location)
			if err != nil {
				return nil, err
			}

			config.ProtokubeImage = &nodeup.Image{
				Name:   kopsbase.DefaultProtokubeImageName(),
				Source: location,
				Hash:   hash.Hex(),
			}
		}

		config.Images = images
		config.Channels = channels

		return config, nil
	}

	bootstrapScriptBuilder := &model.BootstrapScript{
		NodeUpConfigBuilder: renderNodeUpConfig,
		NodeUpSourceHash:    "",
		NodeUpSource:        c.NodeUpSource,
	}
	switch kops.CloudProviderID(cluster.Spec.CloudProvider) {
	case kops.CloudProviderAWS:
		awsModelContext := &awsmodel.AWSModelContext{
			KopsModelContext: modelContext,
		}

		l.Builders = append(l.Builders, &awsmodel.AutoscalingGroupModelBuilder{
			AWSModelContext: awsModelContext,
			BootstrapScript: bootstrapScriptBuilder,
			Lifecycle:       clusterLifecycle,
		})

	case kops.CloudProviderGCE:
		{
			gceModelContext := &gcemodel.GCEModelContext{
				KopsModelContext: modelContext,
			}

			l.Builders = append(l.Builders, &gcemodel.AutoscalingGroupModelBuilder{
				GCEModelContext: gceModelContext,
				BootstrapScript: bootstrapScriptBuilder,
				Lifecycle:       clusterLifecycle,
			})
		}
	case kops.CloudProviderVSphere:
		{
			vsphereModelContext := &vspheremodel.VSphereModelContext{
				KopsModelContext: modelContext,
			}

			l.Builders = append(l.Builders, &vspheremodel.AutoscalingGroupModelBuilder{
				VSphereModelContext: vsphereModelContext,
				BootstrapScript:     bootstrapScriptBuilder,
				Lifecycle:           clusterLifecycle,
			})
		}

	default:
		return fmt.Errorf("unknown cloudprovider %q", cluster.Spec.CloudProvider)
	}

	l.TemplateFunctions["Masters"] = tf.modelContext.MasterInstanceGroups

	tf.AddTo(l.TemplateFunctions)

	taskMap, err := l.BuildTasks(modelStore, fileModels, assetBuilder)
	if err != nil {
		return fmt.Errorf("error building tasks: %v", err)
	}

	var target fi.Target
	dryRun := false
	shouldPrecreateDNS := true

	switch c.TargetName {
	case TargetDirect:
		switch cluster.Spec.CloudProvider {
		case "gce":
			target = gce.NewGCEAPITarget(cloud.(*gce.GCECloud))
		case "aws":
			target = awsup.NewAWSAPITarget(cloud.(awsup.AWSCloud))
		case "vsphere":
			target = vsphere.NewVSphereAPITarget(cloud.(*vsphere.VSphereCloud))
		default:
			return fmt.Errorf("direct configuration not supported with CloudProvider:%q", cluster.Spec.CloudProvider)
		}

	case TargetTerraform:
		checkExisting = false
		outDir := c.OutDir
		tf := terraform.NewTerraformTarget(cloud, region, project, outDir)

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
		err = registry.WriteConfigDeprecated(configBase.Join(registry.PathClusterCompleted), c.Cluster)
		if err != nil {
			return fmt.Errorf("error writing completed cluster spec: %v", err)
		}

		for _, g := range c.InstanceGroups {
			_, err := c.Clientset.InstanceGroupsFor(c.Cluster).Update(g)
			if err != nil {
				return fmt.Errorf("error writing InstanceGroup %q to registry: %v", g.ObjectMeta.Name, err)
			}
		}
	}

	context, err := fi.NewContext(target, cloud, keyStore, secretStore, configBase, checkExisting, taskMap)
	if err != nil {
		return fmt.Errorf("error building context: %v", err)
	}
	defer context.Close()

	err = context.RunTasks(c.MaxTaskDuration)
	if err != nil {
		return fmt.Errorf("error running tasks: %v", err)
	}

	if dns.IsGossipHostname(cluster.Name) {
		shouldPrecreateDNS = false
	}

	if shouldPrecreateDNS {
		if err := precreateDNS(cluster, cloud); err != nil {
			glog.Warningf("unable to pre-create DNS records - cluster startup may be slower: %v", err)
		}
	}

	err = target.Finish(taskMap) //This will finish the apply, and print the changes
	if err != nil {
		return fmt.Errorf("error closing target: %v", err)
	}

	return nil
}

func findHash(url string) (*hashing.Hash, error) {
	for _, ext := range []string{".sha1"} {
		hashURL := url + ext
		b, err := vfs.Context.ReadFile(hashURL)
		if err != nil {
			glog.Infof("error reading hash file %q: %v", hashURL, err)
			continue
		}
		hashString := strings.TrimSpace(string(b))
		glog.V(2).Infof("Found hash %q for %q", hashString, url)

		return hashing.FromString(hashString)
	}
	return nil, fmt.Errorf("cannot determine hash for %v (have you specified a valid KubernetesVersion?)", url)
}

// upgradeSpecs ensures that fields are fully populated / defaulted
func (c *ApplyClusterCmd) upgradeSpecs(assetBuilder *assets.AssetBuilder) error {
	fullCluster, err := PopulateClusterSpec(c.Cluster, assetBuilder)
	if err != nil {
		return err
	}
	c.Cluster = fullCluster

	for i, g := range c.InstanceGroups {
		fullGroup, err := PopulateInstanceGroupSpec(fullCluster, g, c.channel)
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
		glog.Warningf("unable to parse kops version %q", kopsbase.Version)
		// Not a hard-error
		return nil
	}

	versionInfo := kops.FindKopsVersionSpec(c.channel.Spec.KopsVersions, kopsVersion)
	if versionInfo == nil {
		glog.Warningf("unable to find version information for kops version %q in channel", kopsVersion)
		// Not a hard-error
		return nil
	}

	recommended, err := versionInfo.FindRecommendedUpgrade(kopsVersion)
	if err != nil {
		glog.Warningf("unable to parse version recommendation for kops version %q in channel", kopsVersion)
	}

	required, err := versionInfo.IsUpgradeRequired(kopsVersion)
	if err != nil {
		glog.Warningf("unable to parse version requirement for kops version %q in channel", kopsVersion)
	}

	if recommended != nil && !required {
		fmt.Printf("\n")
		fmt.Printf(starline)
		fmt.Printf("\n")
		fmt.Printf("A new kops version is available: %s\n", recommended)
		fmt.Printf("\n")
		fmt.Printf("Upgrading is recommended\n")
		fmt.Printf("More information: %s\n", buildPermalink("upgrade_kops", recommended.String()))
		fmt.Printf("\n")
		fmt.Printf(starline)
		fmt.Printf("\n")
	} else if required {
		fmt.Printf("\n")
		fmt.Printf(starline)
		fmt.Printf("\n")
		if recommended != nil {
			fmt.Printf("A new kops version is available: %s\n", recommended)
		}
		fmt.Printf("\n")
		fmt.Printf("This version of kops is no longer supported; upgrading is required\n")
		fmt.Printf("(you can bypass this check by exporting KOPS_RUN_OBSOLETE_VERSION)\n")
		fmt.Printf("\n")
		fmt.Printf("More information: %s\n", buildPermalink("upgrade_kops", recommended.String()))
		fmt.Printf("\n")
		fmt.Printf(starline)
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
		glog.Warningf("unable to parse kubernetes version %q", c.Cluster.Spec.KubernetesVersion)
		// Not a hard-error
		return nil
	}

	// TODO: make util.ParseKubernetesVersion not return a pointer
	kubernetesVersion := *parsed

	versionInfo := kops.FindKubernetesVersionSpec(c.channel.Spec.KubernetesVersions, kubernetesVersion)
	if versionInfo == nil {
		glog.Warningf("unable to find version information for kubernetes version %q in channel", kubernetesVersion)
		// Not a hard-error
		return nil
	}

	recommended, err := versionInfo.FindRecommendedUpgrade(kubernetesVersion)
	if err != nil {
		glog.Warningf("unable to parse version recommendation for kubernetes version %q in channel", kubernetesVersion)
	}

	required, err := versionInfo.IsUpgradeRequired(kubernetesVersion)
	if err != nil {
		glog.Warningf("unable to parse version requirement for kubernetes version %q in channel", kubernetesVersion)
	}

	if recommended != nil && !required {
		fmt.Printf("\n")
		fmt.Printf(starline)
		fmt.Printf("\n")
		fmt.Printf("A new kubernetes version is available: %s\n", recommended)
		fmt.Printf("Upgrading is recommended (try kops upgrade cluster)\n")
		fmt.Printf("\n")
		fmt.Printf("More information: %s\n", buildPermalink("upgrade_k8s", recommended.String()))
		fmt.Printf("\n")
		fmt.Printf(starline)
		fmt.Printf("\n")
	} else if required {
		fmt.Printf("\n")
		fmt.Printf(starline)
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
		fmt.Printf(starline)
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

func ChannelForCluster(c *kops.Cluster) (*kops.Channel, error) {
	channelLocation := c.Spec.Channel
	if channelLocation == "" {
		channelLocation = kops.DefaultChannel
	}
	return kops.LoadChannel(channelLocation)
}

// needsStaticUtils checks if we need our static utils on this OS.
// This is only needed currently on CoreOS, but we don't have a nice way to detect it yet
func needsStaticUtils(c *kops.Cluster, instanceGroups []*kops.InstanceGroup) bool {
	// TODO: Do real detection of CoreOS (but this has to work with AMI names, and maybe even forked AMIs)
	return true
}

func lifecyclePointer(v fi.Lifecycle) *fi.Lifecycle {
	return &v
}
