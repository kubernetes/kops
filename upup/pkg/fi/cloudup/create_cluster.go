package cloudup

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/gce"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/gcetasks"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kube-deploy/upup/pkg/fi/fitasks"
	"k8s.io/kube-deploy/upup/pkg/fi/loader"
	"k8s.io/kube-deploy/upup/pkg/fi/utils"
	"k8s.io/kube-deploy/upup/pkg/fi/vfs"
	"os"
	"path"
	"strings"
)

const DefaultNodeTypeAWS = "t2.medium"
const DefaultNodeTypeGCE = "n1-standard-2"

// Path for completed cluster spec in the state store
const PathClusterCompleted = "cluster.spec"

type CreateClusterCmd struct {
	// ClusterConfig is the cluster configuration
	ClusterConfig *ClusterConfig

	// NodeSets is the configuration for each NodeSet (group of nodes)
	NodeSets []*NodeSetConfig

	//// NodeUp stores the configuration we are going to pass to nodeup
	//NodeUpConfig  *nodeup.NodeConfig

	// NodeUpSource is the location from which we download nodeup
	NodeUpSource string

	// Tags to pass to NodeUp
	NodeUpTags []string

	// ModelStore is the location where models are found
	ModelStore string
	// Models is a list of cloudup models to apply
	Models []string
	// StateStore is a StateStore in which we store state (such as the PKI tree)
	StateStore fi.StateStore
	// Target specifies how we are operating e.g. direct to GCE, or AWS, or dry-run, or terraform
	Target string
	// The node model to use
	NodeModel string
	// The SSH public key (file) to use
	SSHPublicKey string
	// OutDir is a local directory in which we place output, can cache files etc
	OutDir string

	// Assets is a list of sources for files (primarily when not using everything containerized)
	Assets []string
}

func (c *CreateClusterCmd) LoadConfig(configFile string) error {
	conf, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("error loading configuration file %q: %v", configFile, err)
	}
	err = utils.YamlUnmarshal(conf, c.ClusterConfig)
	if err != nil {
		return fmt.Errorf("error parsing configuration file %q: %v", configFile, err)
	}
	return nil
}

func (c *CreateClusterCmd) Run() error {
	// TODO: Make these configurable?
	useMasterASG := true
	useMasterLB := false

	//// We (currently) have to use protokube with ASGs
	//useProtokube := useMasterASG

	//if c.NodeUpConfig == nil {
	//	c.NodeUpConfig = &nodeup.NodeConfig{}
	//}

	if c.ClusterConfig.ClusterName == "" {
		return fmt.Errorf("--name is required (e.g. mycluster.myzone.com)")
	}

	if c.ClusterConfig.MasterPublicName == "" {
		c.ClusterConfig.MasterPublicName = "api." + c.ClusterConfig.ClusterName
	}
	if c.ClusterConfig.DNSZone == "" {
		tokens := strings.Split(c.ClusterConfig.MasterPublicName, ".")
		c.ClusterConfig.DNSZone = strings.Join(tokens[len(tokens)-2:], ".")
		glog.Infof("Defaulting DNS zone to: %s", c.ClusterConfig.DNSZone)
	}

	if len(c.ClusterConfig.Zones) == 0 {
		// TODO: Auto choose zones from region?
		return fmt.Errorf("must configuration at least one Zone (use --zones)")
	}

	if len(c.NodeSets) == 0 {
		return fmt.Errorf("must configure at least one NodeSet")
	}

	if len(c.ClusterConfig.Masters) == 0 {
		return fmt.Errorf("must configure at least one Master")
	}

	// Check basic master configuration
	{
		masterZones := make(map[string]string)
		for i, m := range c.ClusterConfig.Masters {
			k := m.Name
			if k == "" {
				return fmt.Errorf("Master #%d did not have a key specified", i)
			}

			z := m.Zone
			if z == "" {
				return fmt.Errorf("Master %s did not specify a zone", k)
			}
			if masterZones[z] != "" {
				return fmt.Errorf("Masters %s and %s are in the same zone", k, masterZones[z])
			}
			masterZones[z] = k
		}
	}

	{
		zones := make(map[string]bool)
		for _, z := range c.ClusterConfig.Zones {
			if zones[z.Name] {
				return fmt.Errorf("Zones contained a duplicate value: %v", z.Name)
			}
			zones[z.Name] = true
		}

		for _, m := range c.ClusterConfig.Masters {
			if !zones[m.Zone] {
				// We could relax this, but this seems like a reasonable constraint
				return fmt.Errorf("Master %q is configured in %q, but this is not configured as a Zone", m.Name, m.Zone)
			}
		}

	}

	if (len(c.ClusterConfig.Masters) % 2) == 0 {
		// Not technically a requirement, but doesn't really make sense to allow
		return fmt.Errorf("There should be an odd number of master-zones, for etcd's quorum.  Hint: Use --zone and --master-zone to declare node zones and master zones separately.")
	}

	if c.StateStore == nil {
		return fmt.Errorf("StateStore is required")
	}

	if c.ClusterConfig.CloudProvider == "" {
		return fmt.Errorf("--cloud is required (e.g. aws, gce)")
	}

	tags := make(map[string]struct{})

	l := &Loader{}
	l.Init()

	keyStore := c.StateStore.CA()
	secretStore := c.StateStore.Secrets()

	if vfs.IsClusterReadable(secretStore.VFSPath()) {
		vfsPath := secretStore.VFSPath()
		c.ClusterConfig.SecretStore = vfsPath.Path()
		if s3Path, ok := vfsPath.(*vfs.S3Path); ok {
			if c.ClusterConfig.MasterPermissions == nil {
				c.ClusterConfig.MasterPermissions = &CloudPermissions{}
			}
			c.ClusterConfig.MasterPermissions.AddS3Bucket(s3Path.Bucket())
			if c.ClusterConfig.NodePermissions == nil {
				c.ClusterConfig.NodePermissions = &CloudPermissions{}
			}
			c.ClusterConfig.NodePermissions.AddS3Bucket(s3Path.Bucket())
		}
	} else {
		// We could implement this approach, but it seems better to get all clouds using cluster-readable storage
		return fmt.Errorf("secrets path is not cluster readable: %v", secretStore.VFSPath())
	}

	if vfs.IsClusterReadable(keyStore.VFSPath()) {
		vfsPath := keyStore.VFSPath()
		c.ClusterConfig.KeyStore = vfsPath.Path()
		if s3Path, ok := vfsPath.(*vfs.S3Path); ok {
			if c.ClusterConfig.MasterPermissions == nil {
				c.ClusterConfig.MasterPermissions = &CloudPermissions{}
			}
			c.ClusterConfig.MasterPermissions.AddS3Bucket(s3Path.Bucket())
			if c.ClusterConfig.NodePermissions == nil {
				c.ClusterConfig.NodePermissions = &CloudPermissions{}
			}
			c.ClusterConfig.NodePermissions.AddS3Bucket(s3Path.Bucket())
		}
	} else {
		// We could implement this approach, but it seems better to get all clouds using cluster-readable storage
		return fmt.Errorf("keyStore path is not cluster readable: %v", keyStore.VFSPath())
	}

	if vfs.IsClusterReadable(c.StateStore.VFSPath()) {
		c.ClusterConfig.ConfigStore = c.StateStore.VFSPath().Path()
	} else {
		// We do support this...
	}

	if c.ClusterConfig.KubernetesVersion == "" {
		stableURL := "https://storage.googleapis.com/kubernetes-release/release/stable.txt"
		b, err := vfs.Context.ReadFile(stableURL)
		if err != nil {
			return fmt.Errorf("--kubernetes-version not specified, and unable to download latest version from %q: %v", stableURL, err)
		}
		latestVersion := strings.TrimSpace(string(b))
		glog.Infof("Using kubernetes latest stable version: %s", latestVersion)

		c.ClusterConfig.KubernetesVersion = latestVersion
		//return fmt.Errorf("Must either specify a KubernetesVersion (-kubernetes-version) or provide an asset with the release bundle")
	}

	// Normalize k8s version
	versionWithoutV := strings.TrimSpace(c.ClusterConfig.KubernetesVersion)
	if strings.HasPrefix(versionWithoutV, "v") {
		versionWithoutV = versionWithoutV[1:]
	}
	if c.ClusterConfig.KubernetesVersion != versionWithoutV {
		glog.Warningf("Normalizing kubernetes version: %q -> %q", c.ClusterConfig.KubernetesVersion, versionWithoutV)
		c.ClusterConfig.KubernetesVersion = versionWithoutV
	}

	if len(c.Assets) == 0 {
		//defaultReleaseAsset := fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/v%s/kubernetes-server-linux-amd64.tar.gz", c.Config.KubernetesVersion)
		//glog.Infof("Adding default kubernetes release asset: %s", defaultReleaseAsset)

		defaultKubeletAsset := fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/v%s/bin/linux/amd64/kubelet", c.ClusterConfig.KubernetesVersion)
		glog.Infof("Adding default kubelet release asset: %s", defaultKubeletAsset)

		defaultKubectlAsset := fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/v%s/bin/linux/amd64/kubectl", c.ClusterConfig.KubernetesVersion)
		glog.Infof("Adding default kubelet release asset: %s", defaultKubectlAsset)

		// TODO: Verify assets exist, get the hash (that will check that KubernetesVersion is valid)

		c.Assets = append(c.Assets, defaultKubeletAsset, defaultKubectlAsset)
	}

	if c.NodeUpSource == "" {
		location := "https://kubeupv2.s3.amazonaws.com/nodeup/nodeup-1.3.tar.gz"
		glog.Infof("Using default nodeup location: %q", location)
		c.NodeUpSource = location
	}

	var cloud fi.Cloud

	var project string

	checkExisting := true

	//c.NodeUpConfig.Tags = append(c.NodeUpConfig.Tags, "_jessie", "_debian_family", "_systemd")
	//
	//if useProtokube {
	//	tags["_protokube"] = struct{}{}
	//	c.NodeUpConfig.Tags = append(c.NodeUpConfig.Tags, "_protokube")
	//} else {
	//	tags["_not_protokube"] = struct{}{}
	//	c.NodeUpConfig.Tags = append(c.NodeUpConfig.Tags, "_not_protokube")
	//}

	c.NodeUpTags = append(c.NodeUpTags, "_protokube")

	if useMasterASG {
		tags["_master_asg"] = struct{}{}
	} else {
		tags["_master_single"] = struct{}{}
	}

	if useMasterLB {
		tags["_master_lb"] = struct{}{}
	} else {
		tags["_not_master_lb"] = struct{}{}
	}

	if c.ClusterConfig.MasterPublicName != "" {
		tags["_master_dns"] = struct{}{}
	}

	l.AddTypes(map[string]interface{}{
		"keypair": &fitasks.Keypair{},
		"secret":  &fitasks.Secret{},
	})

	region := ""

	switch c.ClusterConfig.CloudProvider {
	case "gce":
		{
			glog.Fatalf("GCE is (probably) not working currently - please ping @justinsb for cleanup")
			tags["_gce"] = struct{}{}
			c.NodeUpTags = append(c.NodeUpTags, "_gce")

			l.AddTypes(map[string]interface{}{
				"persistentDisk":       &gcetasks.PersistentDisk{},
				"instance":             &gcetasks.Instance{},
				"instanceTemplate":     &gcetasks.InstanceTemplate{},
				"network":              &gcetasks.Network{},
				"managedInstanceGroup": &gcetasks.ManagedInstanceGroup{},
				"firewallRule":         &gcetasks.FirewallRule{},
				"ipAddress":            &gcetasks.IPAddress{},
			})

			nodeZones := make(map[string]bool)
			for _, zone := range c.ClusterConfig.Zones {
				nodeZones[zone.Name] = true

				tokens := strings.Split(zone.Name, "-")
				if len(tokens) <= 2 {
					return fmt.Errorf("Invalid GCE Zone: %v", zone.Name)
				}
				zoneRegion := tokens[0] + "-" + tokens[1]
				if region != "" && zoneRegion != region {
					return fmt.Errorf("Clusters cannot span multiple regions")
				}

				region = zoneRegion
			}

			//err := awsup.ValidateRegion(region)
			//if err != nil {
			//	return err
			//}

			project = c.ClusterConfig.Project
			if project == "" {
				return fmt.Errorf("project is required for GCE")
			}
			gceCloud, err := gce.NewGCECloud(region, project)
			if err != nil {
				return err
			}

			//var zoneNames []string
			//for _, z := range c.Config.Zones {
			//	zoneNames = append(zoneNames, z.Name)
			//}
			//err = gceCloud.ValidateZones(zoneNames)
			//if err != nil {
			//	return err
			//}

			cloud = gceCloud
		}

	case "aws":
		{
			tags["_aws"] = struct{}{}
			c.NodeUpTags = append(c.NodeUpTags, "_aws")

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

			nodeZones := make(map[string]bool)
			for _, zone := range c.ClusterConfig.Zones {
				if len(zone.Name) <= 2 {
					return fmt.Errorf("Invalid AWS zone: %q", zone.Name)
				}

				nodeZones[zone.Name] = true

				zoneRegion := zone.Name[:len(zone.Name)-1]
				if region != "" && zoneRegion != region {
					return fmt.Errorf("Clusters cannot span multiple regions")
				}

				region = zoneRegion
			}

			err := awsup.ValidateRegion(region)
			if err != nil {
				return err
			}

			if c.SSHPublicKey == "" {
				return fmt.Errorf("SSH public key must be specified when running with AWS")
			}

			cloudTags := map[string]string{awsup.TagClusterName: c.ClusterConfig.ClusterName}

			awsCloud, err := awsup.NewAWSCloud(region, cloudTags)
			if err != nil {
				return err
			}

			var zoneNames []string
			for _, z := range c.ClusterConfig.Zones {
				zoneNames = append(zoneNames, z.Name)
			}
			err = awsCloud.ValidateZones(zoneNames)
			if err != nil {
				return err
			}
			cloud = awsCloud

			l.TemplateFunctions["MachineTypeInfo"] = awsup.GetMachineTypeInfo
		}

	default:
		return fmt.Errorf("unknown CloudProvider %q", c.ClusterConfig.CloudProvider)
	}

	l.Tags = tags
	l.WorkDir = c.OutDir
	l.ModelStore = c.ModelStore
	l.NodeModel = c.NodeModel
	l.OptionsLoader = loader.NewOptionsLoader(c.ClusterConfig)

	l.TemplateFunctions["HasTag"] = func(tag string) bool {
		_, found := l.Tags[tag]
		return found
	}

	l.TemplateFunctions["CA"] = func() fi.CAStore {
		return keyStore
	}
	l.TemplateFunctions["Secrets"] = func() fi.SecretStore {
		return secretStore
	}

	l.TemplateFunctions["NodeUpTags"] = func() []string {
		return c.NodeUpTags
	}

	// TotalNodeCount computes the total count of nodes
	l.TemplateFunctions["TotalNodeCount"] = func() (int, error) {
		count := 0
		for _, nodeset := range c.NodeSets {
			if nodeset.MaxSize != nil {
				count += *nodeset.MaxSize
			} else if nodeset.MinSize != nil {
				count += *nodeset.MinSize
			} else {
				// Guestimate
				count += 5
			}
		}
		return count, nil
	}
	l.TemplateFunctions["Region"] = func() string {
		return region
	}
	l.TemplateFunctions["NodeSets"] = c.populateNodeSets
	l.TemplateFunctions["Masters"] = c.populateMasters
	//l.TemplateFunctions["NodeUp"] = c.populateNodeUpConfig
	l.TemplateFunctions["NodeUpSource"] = func() string { return c.NodeUpSource }
	l.TemplateFunctions["NodeUpSourceHash"] = func() string { return "" }
	l.TemplateFunctions["ClusterLocation"] = func() string { return c.StateStore.VFSPath().Join(PathClusterCompleted).Path() }
	l.TemplateFunctions["Assets"] = func() []string { return c.Assets }

	// TODO: Fix this duplication
	l.OptionsLoader.TemplateFunctions["HasTag"] = l.TemplateFunctions["HasTag"]
	l.OptionsLoader.TemplateFunctions["TotalNodeCount"] = l.TemplateFunctions["TotalNodeCount"]
	l.OptionsLoader.TemplateFunctions["Assets"] = l.TemplateFunctions["Assets"]

	if c.SSHPublicKey != "" {
		authorized, err := ioutil.ReadFile(c.SSHPublicKey)
		if err != nil {
			return fmt.Errorf("error reading SSH key file %q: %v", c.SSHPublicKey, err)
		}

		l.Resources["ssh-public-key"] = fi.NewStringResource(string(authorized))
	}

	completed, err := l.BuildCompleteSpec(c.ModelStore, c.Models)
	if err != nil {
		return fmt.Errorf("error building complete spec: %v", err)
	}

	taskMap, err := l.BuildTasks(c.ModelStore, c.Models)
	if err != nil {
		return fmt.Errorf("error building tasks: %v", err)
	}

	err = c.StateStore.WriteConfig(PathClusterCompleted, completed)
	if err != nil {
		return fmt.Errorf("error writing cluster spec: %v", err)
	}

	var target fi.Target

	switch c.Target {
	case "direct":
		switch c.ClusterConfig.CloudProvider {
		case "gce":
			target = gce.NewGCEAPITarget(cloud.(*gce.GCECloud))
		case "aws":
			target = awsup.NewAWSAPITarget(cloud.(*awsup.AWSCloud))
		default:
			return fmt.Errorf("direct configuration not supported with CloudProvider:%q", c.ClusterConfig.CloudProvider)
		}

	case "terraform":
		checkExisting = false
		outDir := path.Join(c.OutDir, "terraform")
		target = terraform.NewTerraformTarget(cloud, region, project, outDir)

	case "dryrun":
		target = fi.NewDryRunTarget(os.Stdout)
	default:
		return fmt.Errorf("unsupported target type %q", c.Target)
	}

	context, err := fi.NewContext(target, cloud, keyStore, secretStore, checkExisting)
	if err != nil {
		return fmt.Errorf("error building context: %v", err)
	}
	defer context.Close()

	err = context.RunTasks(taskMap)
	if err != nil {
		return fmt.Errorf("error running tasks: %v", err)
	}

	err = target.Finish(taskMap)
	if err != nil {
		return fmt.Errorf("error closing target: %v", err)
	}

	return nil
}

// populateNodeSets returns the NodeSets with values populated from defaults or top-level config
func (c *CreateClusterCmd) populateNodeSets() ([]*NodeSetConfig, error) {
	var results []*NodeSetConfig
	for _, src := range c.NodeSets {
		n := &NodeSetConfig{}
		*n = *src

		if n.MachineType == "" {
			n.MachineType = c.defaultMachineType()
		}

		if n.Image == "" {
			n.Image = c.defaultImage()
		}

		results = append(results, n)
	}
	return results, nil
}

// populateMasters returns the Masters with values populated from defaults or top-level config
func (c *CreateClusterCmd) populateMasters() ([]*MasterConfig, error) {
	cluster := c.ClusterConfig

	var results []*MasterConfig
	for _, src := range cluster.Masters {
		m := &MasterConfig{}
		*m = *src

		if m.MachineType == "" {
			m.MachineType = c.defaultMachineType()
		}

		if m.Image == "" {
			m.Image = c.defaultImage()
		}

		results = append(results, m)
	}
	return results, nil
}

//// populateNodeUpConfig returns the NodeUpConfig with values populated from defaults or top-level config
//func (c*CreateClusterCmd) populateNodeUpConfig() (*nodeup.NodeConfig, error) {
//	conf := &nodeup.NodeConfig{}
//	*conf = *c.NodeUpConfig
//
//	return conf, nil
//}

// defaultMachineType returns the default MachineType, based on the cloudprovider
func (c *CreateClusterCmd) defaultMachineType() string {
	cluster := c.ClusterConfig
	switch cluster.CloudProvider {
	case "aws":
		return DefaultNodeTypeAWS
	case "gce":
		return DefaultNodeTypeGCE
	default:
		glog.V(2).Infof("Cannot set default MachineType for CloudProvider=%q", cluster.CloudProvider)
		return ""
	}
}

// defaultImage returns the default Image, based on the cloudprovider
func (c *CreateClusterCmd) defaultImage() string {
	// TODO: Use spec
	cluster := c.ClusterConfig
	switch cluster.CloudProvider {
	case "aws":
		return "282335181503/k8s-1.3-debian-jessie-amd64-hvm-ebs-2016-06-18"
	default:
		glog.V(2).Infof("Cannot set default Image for CloudProvider=%q", cluster.CloudProvider)
		return ""
	}
}
