package cloudup

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/upup/pkg/fi/hashing"
	"k8s.io/kops/upup/pkg/fi/nodeup"
	"k8s.io/kops/upup/pkg/fi/vfs"
	"os"
	"path"
	"strings"
)

const MaxAttemptsWithNoProgress = 3

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

	// ClusterRegistry manages the cluster configuration storage
	ClusterRegistry *api.ClusterRegistry

	// DryRun is true if this is only a dry run
	DryRun bool
}

func (c *ApplyClusterCmd) Run() error {
	modelStore, err := findModelStore()
	if err != nil {
		return err
	}

	err = api.DeepValidate(c.Cluster, c.InstanceGroups, true)
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

	if c.ClusterRegistry == nil {
		return fmt.Errorf("ClusterRegistry is required")
	}

	l := &Loader{}
	l.Init()
	l.Cluster = c.Cluster

	keyStore := c.ClusterRegistry.KeyStore(cluster.Name)
	keyStore.(*fi.VFSCAStore).DryRun = c.DryRun
	secretStore := c.ClusterRegistry.SecretStore(cluster.Name)

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
		if isBaseURL(cluster.Spec.KubernetesVersion) {
			baseURL = cluster.Spec.KubernetesVersion
		} else {
			baseURL = "https://storage.googleapis.com/kubernetes-release/release/v" + cluster.Spec.KubernetesVersion
		}
		baseURL = strings.TrimSuffix(baseURL, "/")

		{
			defaultKubeletAsset := baseURL + "/bin/linux/amd64/kubelet"
			glog.Infof("Adding default kubelet release asset: %s", defaultKubeletAsset)

			hash, err := findHash(defaultKubeletAsset)
			if err != nil {
				return err
			}
			c.Assets = append(c.Assets, hash.Hex()+"@"+defaultKubeletAsset)
		}

		{
			defaultKubectlAsset := baseURL + "/bin/linux/amd64/kubectl"
			glog.Infof("Adding default kubectl release asset: %s", defaultKubectlAsset)

			hash, err := findHash(defaultKubectlAsset)
			if err != nil {
				return err
			}
			c.Assets = append(c.Assets, hash.Hex()+"@"+defaultKubectlAsset)
		}
	}

	if c.NodeUpSource == "" {
		location := "https://kubeupv2.s3.amazonaws.com/nodeup/nodeup-1.3.tar.gz"
		glog.Infof("Using default nodeup location: %q", location)
		c.NodeUpSource = location
	}

	checkExisting := true

	l.AddTypes(map[string]interface{}{
		"keypair": &fitasks.Keypair{},
		"secret":  &fitasks.Secret{},
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

	switch cluster.Spec.CloudProvider {
	case "gce":
		{
			gceCloud := cloud.(*gce.GCECloud)
			region = gceCloud.Region
			project = gceCloud.Project

			glog.Fatalf("GCE is (probably) not working currently - please ping @justinsb for cleanup")

			l.AddTypes(map[string]interface{}{
				"persistentDisk":       &gcetasks.PersistentDisk{},
				"instance":             &gcetasks.Instance{},
				"instanceTemplate":     &gcetasks.InstanceTemplate{},
				"network":              &gcetasks.Network{},
				"managedInstanceGroup": &gcetasks.ManagedInstanceGroup{},
				"firewallRule":         &gcetasks.FirewallRule{},
				"ipAddress":            &gcetasks.IPAddress{},
			})
		}

	case "aws":
		{
			awsCloud := cloud.(*awsup.AWSCloud)
			region = awsCloud.Region

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
				"vpcDHDCPOptionsAssociation": &awstasks.VPCDHCPOptionsAssociation{},

				// ELB
				"loadBalancer":             &awstasks.LoadBalancer{},
				"loadBalancerAttachment":   &awstasks.LoadBalancerAttachment{},
				"loadBalancerHealthChecks": &awstasks.LoadBalancerHealthChecks{},

				// Autoscaling
				"autoscalingGroup":    &awstasks.AutoscalingGroup{},
				"launchConfiguration": &awstasks.LaunchConfiguration{},

				// Route53
				"dnsName": &awstasks.DNSName{},
				"dnsZone": &awstasks.DNSZone{},
			})

			if len(sshPublicKeys) == 0 {
				return fmt.Errorf("SSH public key must be specified when running with AWS (create with `kops create secret --name %s sshpublickey admin -i ~/.ssh/id_rsa.pub`)", cluster.Name)
			}

			if len(sshPublicKeys) != 1 {
				return fmt.Errorf("Exactly one 'admin' SSH public key can be specified when running with AWS; please delete a key using `kops delete secret`")
			} else {
				l.Resources["ssh-public-key"] = fi.NewStringResource(string(sshPublicKeys[0]))

				// SSHKeyName computes a unique SSH key name, combining the cluster name and the SSH public key fingerprint
				l.TemplateFunctions["SSHKeyName"] = func() (string, error) {
					fingerprint, err := awstasks.ComputeOpenSSHKeyFingerprint(string(sshPublicKeys[0]))
					if err != nil {
						return "", err
					}

					name := "kubernetes." + cluster.Name + "-" + fingerprint
					return name, nil
				}
			}

			l.TemplateFunctions["MachineTypeInfo"] = awsup.GetMachineTypeInfo
		}

	default:
		return fmt.Errorf("unknown CloudProvider %q", cluster.Spec.CloudProvider)
	}

	clusterTags, err := buildClusterTags(cluster)
	if err != nil {
		return err
	}

	tf := &TemplateFunctions{
		cluster: cluster,
		tags:    clusterTags,
		region:  region,
	}

	l.Tags = clusterTags
	l.WorkDir = c.OutDir
	l.ModelStore = modelStore

	l.TemplateFunctions["CA"] = func() fi.CAStore {
		return keyStore
	}
	l.TemplateFunctions["Secrets"] = func() fi.SecretStore {
		return secretStore
	}

	// RenderNodeUpConfig returns the NodeUp config, in YAML format
	l.TemplateFunctions["RenderNodeUpConfig"] = func(args []string) (string, error) {
		var role api.InstanceGroupRole
		for _, arg := range args {
			if arg == "_kubernetes_master" {
				if role != "" {
					return "", fmt.Errorf("found duplicate role tags in args: %v", args)
				}
				role = api.InstanceGroupRoleMaster
			}
			if arg == "_kubernetes_pool" {
				if role != "" {
					return "", fmt.Errorf("found duplicate role tags in args: %v", args)
				}
				role = api.InstanceGroupRoleNode
			}
		}
		if role == "" {
			return "", fmt.Errorf("cannot determine role from args: %v", args)
		}

		nodeUpTags, err := buildNodeupTags(role, tf.cluster, tf.tags)
		if err != nil {
			return "", err
		}

		config := &nodeup.NodeUpConfig{}
		for _, tag := range args {
			config.Tags = append(config.Tags, tag)
		}
		for _, tag := range nodeUpTags {
			config.Tags = append(config.Tags, tag)
		}

		config.Assets = c.Assets

		config.ClusterName = cluster.Name

		configPath, err := c.ClusterRegistry.ConfigurationPath(cluster.Name)
		if err != nil {
			return "", err
		}
		config.ClusterLocation = configPath.Path()

		var images []*nodeup.Image

		if isBaseURL(cluster.Spec.KubernetesVersion) {
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
					return "", err
				}
				image := &nodeup.Image{
					Source: imagePath,
					Hash:   hash.Hex(),
				}
				images = append(images, image)
			}
		}

		config.Images = images
		yaml, err := api.ToYaml(config)
		if err != nil {
			return "", err
		}

		return string(yaml), nil
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
	l.TemplateFunctions["NodeSets"] = func() []*api.InstanceGroup {
		var groups []*api.InstanceGroup
		for _, ig := range c.InstanceGroups {
			if ig.IsMaster() {
				continue
			}
			groups = append(groups, ig)
		}
		return groups
	}
	l.TemplateFunctions["Masters"] = func() []*api.InstanceGroup {
		var groups []*api.InstanceGroup
		for _, ig := range c.InstanceGroups {
			if !ig.IsMaster() {
				continue
			}
			groups = append(groups, ig)
		}
		return groups
	}
	//l.TemplateFunctions["NodeUp"] = c.populateNodeUpConfig
	l.TemplateFunctions["NodeUpSource"] = func() string {
		return c.NodeUpSource
	}
	l.TemplateFunctions["NodeUpSourceHash"] = func() string {
		return ""
	}

	tf.AddTo(l.TemplateFunctions)

	taskMap, err := l.BuildTasks(modelStore, c.Models)
	if err != nil {
		return fmt.Errorf("error building tasks: %v", err)
	}

	var target fi.Target

	switch c.TargetName {
	case TargetDirect:
		switch cluster.Spec.CloudProvider {
		case "gce":
			target = gce.NewGCEAPITarget(cloud.(*gce.GCECloud))
		case "aws":
			target = awsup.NewAWSAPITarget(cloud.(*awsup.AWSCloud))
		default:
			return fmt.Errorf("direct configuration not supported with CloudProvider:%q", cluster.Spec.CloudProvider)
		}

	case TargetTerraform:
		checkExisting = false
		outDir := path.Join(c.OutDir, "terraform")
		target = terraform.NewTerraformTarget(cloud, region, project, outDir)

	case TargetDryRun:
		target = fi.NewDryRunTarget(os.Stdout)
	default:
		return fmt.Errorf("unsupported target type %q", c.TargetName)
	}
	c.Target = target

	context, err := fi.NewContext(target, cloud, keyStore, secretStore, checkExisting)
	if err != nil {
		return fmt.Errorf("error building context: %v", err)
	}
	defer context.Close()

	err = context.RunTasks(taskMap, MaxAttemptsWithNoProgress)
	if err != nil {
		return fmt.Errorf("error running tasks: %v", err)
	}

	err = target.Finish(taskMap)
	if err != nil {
		return fmt.Errorf("error closing target: %v", err)
	}

	return nil
}

func isBaseURL(kubernetesVersion string) bool {
	return strings.HasPrefix(kubernetesVersion, "http:") || strings.HasPrefix(kubernetesVersion, "https:")
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
		glog.Infof("Found hash %q for %q", hashString, url)

		return hashing.FromString(hashString)
	}
	return nil, fmt.Errorf("cannot determine hash for %v (have you specified a valid KubernetesVersion?)", url)
}
