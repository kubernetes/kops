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

	"github.com/blang/semver"
	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/awsmodel"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/pkg/model/gcemodel"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/util/pkg/hashing"
	"k8s.io/kops/util/pkg/vfs"
)

const DefaultMaxTaskDuration = 10 * time.Minute

const starline = "*********************************************************************************\n"

// AlphaAllowGCE is a feature flag that gates GCE support while it is alpha
var AlphaAllowGCE = featureflag.New("AlphaAllowGCE", featureflag.Bool(false))

var CloudupModels = []string{"config", "proto", "cloudup"}

type ApplyClusterCmd struct {
	Cluster *api.Cluster

	InstanceGroups []*api.InstanceGroup

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
	channel *api.Channel
}

func (c *ApplyClusterCmd) Run() error {
	if c.MaxTaskDuration == 0 {
		c.MaxTaskDuration = DefaultMaxTaskDuration
	}

	if c.InstanceGroups == nil {
		list, err := c.Clientset.InstanceGroups(c.Cluster.ObjectMeta.Name).List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		var instanceGroups []*api.InstanceGroup
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

	err = c.upgradeSpecs()
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
	if cluster.Spec.DNSZone == "" {
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
			cniAsset, cniAssetHashString := findCNIAssets(cluster)
			c.Assets = append(c.Assets, cniAssetHashString+"@"+cniAsset)
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

		// DNS
		//"dnsZone": &dnstasks.DNSZone{},
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

	switch fi.CloudProviderID(cluster.Spec.CloudProvider) {
	case fi.CloudProviderGCE:
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

	case fi.CloudProviderAWS:
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

				//// Route53
				//"dnsName": &awstasks.DNSName{},
				//"dnsZone": &awstasks.DNSZone{},
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

	default:
		return fmt.Errorf("unknown CloudProvider %q", cluster.Spec.CloudProvider)
	}

	modelContext.Region = region

	err = validateDNS(cluster, cloud)
	if err != nil {
		return err
	}
	dnszone, err := findZone(cluster, cloud)
	if err != nil {
		return err
	}
	modelContext.HostedZoneID = dnszone.ID()

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

	var fileModels []string
	for _, m := range c.Models {
		switch m {
		case "proto":
		// No proto code options; no file model

		case "cloudup":
			l.Builders = append(l.Builders,
				&BootstrapChannelBuilder{cluster: cluster},
			)

			switch fi.CloudProviderID(cluster.Spec.CloudProvider) {
			case fi.CloudProviderAWS:
				awsModelContext := &awsmodel.AWSModelContext{
					KopsModelContext: modelContext,
				}

				l.Builders = append(l.Builders,
					&model.PKIModelBuilder{KopsModelContext: modelContext},
					&model.MasterVolumeBuilder{KopsModelContext: modelContext},

					&awsmodel.APILoadBalancerBuilder{AWSModelContext: awsModelContext},
					&model.BastionModelBuilder{KopsModelContext: modelContext},
					&model.DNSModelBuilder{KopsModelContext: modelContext},
					&model.ExternalAccessModelBuilder{KopsModelContext: modelContext},
					&model.FirewallModelBuilder{KopsModelContext: modelContext},
					&model.IAMModelBuilder{KopsModelContext: modelContext},
					&model.NetworkModelBuilder{KopsModelContext: modelContext},
					&model.SSHKeyModelBuilder{KopsModelContext: modelContext},
				)

			case fi.CloudProviderGCE:
				gceModelContext := &gcemodel.GCEModelContext{
					KopsModelContext: modelContext,
				}

				l.Builders = append(l.Builders,
					&model.PKIModelBuilder{KopsModelContext: modelContext},
					&model.MasterVolumeBuilder{KopsModelContext: modelContext},

					&gcemodel.APILoadBalancerBuilder{GCEModelContext: gceModelContext},
					//&model.BastionModelBuilder{KopsModelContext: modelContext},
					//&model.DNSModelBuilder{KopsModelContext: modelContext},
					&gcemodel.ExternalAccessModelBuilder{GCEModelContext: gceModelContext},
					&gcemodel.FirewallModelBuilder{GCEModelContext: gceModelContext},
					//&model.IAMModelBuilder{KopsModelContext: modelContext},
					&gcemodel.NetworkModelBuilder{GCEModelContext: gceModelContext},
					//&model.SSHKeyModelBuilder{KopsModelContext: modelContext},
				)

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
	renderNodeUpConfig := func(ig *api.InstanceGroup) (*nodeup.NodeUpConfig, error) {
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

		config := &nodeup.NodeUpConfig{}
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
			if role == api.InstanceGroupRoleMaster {
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
				Name:   kops.DefaultProtokubeImageName(),
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
	switch fi.CloudProviderID(cluster.Spec.CloudProvider) {
	case fi.CloudProviderAWS:
		awsModelContext := &awsmodel.AWSModelContext{
			KopsModelContext: modelContext,
		}

		l.Builders = append(l.Builders, &awsmodel.AutoscalingGroupModelBuilder{
			AWSModelContext: awsModelContext,
			BootstrapScript: bootstrapScriptBuilder,
		})

	case fi.CloudProviderGCE:
		{
			gceModelContext := &gcemodel.GCEModelContext{
				KopsModelContext: modelContext,
			}

			l.Builders = append(l.Builders, &gcemodel.AutoscalingGroupModelBuilder{
				GCEModelContext: gceModelContext,
				BootstrapScript: bootstrapScriptBuilder,
			})
		}

	default:
		return fmt.Errorf("unknown cloudprovider %q", cluster.Spec.CloudProvider)
	}

	//// TotalNodeCount computes the total count of nodes
	//l.TemplateFunctions["TotalNodeCount"] = func() (int, error) {
	//	count := 0
	//	for _, group := range c.InstanceGroups {
	//		if group.IsMaster() {
	//			continue
	//		}
	//		if group.Spec.MaxSize != nil {
	//			count += *group.Spec.MaxSize
	//		} else if group.Spec.MinSize != nil {
	//			count += *group.Spec.MinSize
	//		} else {
	//			// Guestimate
	//			count += 5
	//		}
	//	}
	//	return count, nil
	//}
	l.TemplateFunctions["Region"] = func() string {
		return region
	}
	l.TemplateFunctions["Masters"] = tf.modelContext.MasterInstanceGroups

	tf.AddTo(l.TemplateFunctions)

	taskMap, err := l.BuildTasks(modelStore, fileModels)
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
		target = fi.NewDryRunTarget(os.Stdout)
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
			_, err := c.Clientset.InstanceGroups(c.Cluster.ObjectMeta.Name).Update(g)
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
func (c *ApplyClusterCmd) upgradeSpecs() error {
	//err := c.Cluster.PerformAssignments()
	//if err != nil {
	//	return fmt.Errorf("error populating configuration: %v", err)
	//}

	fullCluster, err := PopulateClusterSpec(c.Cluster)
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
	kopsVersion, err := semver.ParseTolerant(kops.Version)
	if err != nil {
		glog.Warningf("unable to parse kops version %q", kops.Version)
		// Not a hard-error
		return nil
	}

	versionInfo := api.FindKopsVersionSpec(c.channel.Spec.KopsVersions, kopsVersion)
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

	versionInfo := api.FindKubernetesVersionSpec(c.channel.Spec.KubernetesVersions, kubernetesVersion)
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

func ChannelForCluster(c *api.Cluster) (*api.Channel, error) {
	channelLocation := c.Spec.Channel
	if channelLocation == "" {
		channelLocation = api.DefaultChannel
	}
	return api.LoadChannel(channelLocation)
}

// needsStaticUtils checks if we need our static utils on this OS.
// This is only needed currently on CoreOS, but we don't have a nice way to detect it yet
func needsStaticUtils(c *api.Cluster, instanceGroups []*api.InstanceGroup) bool {
	// TODO: Do real detection of CoreOS (but this has to work with AMI names, and maybe even forked AMIs)
	return true
}
