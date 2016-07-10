package cloudup

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/upup/pkg/fi/loader"
	"k8s.io/kops/upup/pkg/fi/vfs"
	"net"
	"os"
	"path"
	"strings"
)

const DefaultNodeTypeAWS = "t2.medium"
const DefaultNodeTypeGCE = "n1-standard-2"

type CreateClusterCmd struct {
	// Cluster is the api object representing the whole cluster
	Cluster *api.Cluster

	// InstanceGroups is the configuration for each InstanceGroup
	// These are the groups for both master & nodes
	// We normally expect most Master NodeSets to be MinSize=MaxSize=1,
	// but they don't have to be.
	InstanceGroups []*api.InstanceGroup

	// nodes is the set of InstanceGroups for the nodes
	nodes []*api.InstanceGroup
	// masters is the set of InstanceGroups for the masters
	masters []*api.InstanceGroup

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

	// ClusterRegistry manages the cluster configuration storage
	ClusterRegistry *api.ClusterRegistry

	// DryRun is true if this is only a dry run
	DryRun bool
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

	clusterName := c.Cluster.Name
	if clusterName == "" {
		return fmt.Errorf("ClusterName is required (e.g. --name=mycluster.myzone.com)")
	}

	if c.Cluster.Spec.MasterPublicName == "" {
		c.Cluster.Spec.MasterPublicName = "api." + c.Cluster.Name
	}
	if c.Cluster.Spec.DNSZone == "" {
		tokens := strings.Split(c.Cluster.Spec.MasterPublicName, ".")
		c.Cluster.Spec.DNSZone = strings.Join(tokens[len(tokens)-2:], ".")
		glog.Infof("Defaulting DNS zone to: %s", c.Cluster.Spec.DNSZone)
	}

	if len(c.Cluster.Spec.Zones) == 0 {
		// TODO: Auto choose zones from region?
		return fmt.Errorf("must configure at least one Zone (use --zones)")
	}

	if len(c.InstanceGroups) == 0 {
		return fmt.Errorf("must configure at least one InstanceGroup")
	}

	for i, g := range c.InstanceGroups {
		if g.Name == "" {
			return fmt.Errorf("InstanceGroup #%d Name not set", i)
		}
		if g.Spec.Role == "" {
			return fmt.Errorf("InstanceGroup %q Role not set", g.Name)
		}
	}

	masters, err := c.populateMasters()
	if err != nil {
		return err
	}
	c.masters = masters
	if len(c.masters) == 0 {
		return fmt.Errorf("must configure at least one Master InstanceGroup")
	}

	nodes, err := c.populateNodeSets()
	if err != nil {
		return err
	}
	c.nodes = nodes
	if len(c.nodes) == 0 {
		return fmt.Errorf("must configure at least one Node InstanceGroup")
	}

	err = c.assignSubnets()
	if err != nil {
		return err
	}

	// Check that instance groups are defined in valid zones
	{
		clusterZones := make(map[string]*api.ClusterZoneSpec)
		for _, z := range c.Cluster.Spec.Zones {
			if clusterZones[z.Name] != nil {
				return fmt.Errorf("Zones contained a duplicate value: %v", z.Name)
			}
			clusterZones[z.Name] = z
		}

		for _, group := range c.InstanceGroups {
			for _, z := range group.Spec.Zones {
				if clusterZones[z] == nil {
					return fmt.Errorf("InstanceGroup %q is configured in %q, but this is not configured as a Zone in the cluster", group.Name, z)
				}
			}
		}

		// Check etcd configuration
		{
			for i, etcd := range c.Cluster.Spec.EtcdClusters {
				if etcd.Name == "" {
					return fmt.Errorf("EtcdClusters #%d did not specify a Name", i)
				}

				for i, m := range etcd.Members {
					if m.Name == "" {
						return fmt.Errorf("EtcdMember #%d of etcd-cluster %s did not specify a Name", i, etcd.Name)
					}

					z := m.Zone
					if z == "" {
						return fmt.Errorf("EtcdMember %s:%s did not specify a Zone", etcd.Name, m.Name)
					}
				}

				etcdZones := make(map[string]*api.EtcdMemberSpec)
				etcdNames := make(map[string]*api.EtcdMemberSpec)

				for _, m := range etcd.Members {
					if etcdNames[m.Name] != nil {
						return fmt.Errorf("EtcdMembers found with same name %q in etcd-cluster %q", m.Name, etcd.Name)
					}

					if etcdZones[m.Zone] != nil {
						// Maybe this should just be a warning
						return fmt.Errorf("EtcdMembers are in the same zone %q in etcd-cluster %q", m.Zone, etcd.Name)
					}

					if clusterZones[m.Zone] == nil {
						return fmt.Errorf("EtcdMembers for %q is configured in zone %q, but that is not configured at the k8s-cluster level", etcd.Name, m.Zone)
					}
					etcdZones[m.Zone] = m
				}

				if (len(etcdZones) % 2) == 0 {
					// Not technically a requirement, but doesn't really make sense to allow
					return fmt.Errorf("There should be an odd number of master-zones, for etcd's quorum.  Hint: Use --zones and --master-zones to declare node zones and master zones separately.")
				}
			}
		}
	}

	if c.ClusterRegistry == nil {
		return fmt.Errorf("ClusterRegistry is required")
	}

	if c.Cluster.Spec.CloudProvider == "" {
		return fmt.Errorf("--cloud is required (e.g. aws, gce)")
	}

	tags := make(map[string]struct{})

	l := &Loader{}
	l.Init()

	keyStore := c.ClusterRegistry.KeyStore(clusterName)
	if c.DryRun {
		keyStore.(*fi.VFSCAStore).DryRun = true
	}
	secretStore := c.ClusterRegistry.SecretStore(clusterName)

	if vfs.IsClusterReadable(secretStore.VFSPath()) {
		vfsPath := secretStore.VFSPath()
		c.Cluster.Spec.SecretStore = vfsPath.Path()
		if s3Path, ok := vfsPath.(*vfs.S3Path); ok {
			if c.Cluster.Spec.MasterPermissions == nil {
				c.Cluster.Spec.MasterPermissions = &api.CloudPermissions{}
			}
			c.Cluster.Spec.MasterPermissions.AddS3Bucket(s3Path.Bucket())
			if c.Cluster.Spec.NodePermissions == nil {
				c.Cluster.Spec.NodePermissions = &api.CloudPermissions{}
			}
			c.Cluster.Spec.NodePermissions.AddS3Bucket(s3Path.Bucket())
		}
	} else {
		// We could implement this approach, but it seems better to get all clouds using cluster-readable storage
		return fmt.Errorf("secrets path is not cluster readable: %v", secretStore.VFSPath())
	}

	if vfs.IsClusterReadable(keyStore.VFSPath()) {
		vfsPath := keyStore.VFSPath()
		c.Cluster.Spec.KeyStore = vfsPath.Path()
		if s3Path, ok := vfsPath.(*vfs.S3Path); ok {
			if c.Cluster.Spec.MasterPermissions == nil {
				c.Cluster.Spec.MasterPermissions = &api.CloudPermissions{}
			}
			c.Cluster.Spec.MasterPermissions.AddS3Bucket(s3Path.Bucket())
			if c.Cluster.Spec.NodePermissions == nil {
				c.Cluster.Spec.NodePermissions = &api.CloudPermissions{}
			}
			c.Cluster.Spec.NodePermissions.AddS3Bucket(s3Path.Bucket())
		}
	} else {
		// We could implement this approach, but it seems better to get all clouds using cluster-readable storage
		return fmt.Errorf("keyStore path is not cluster readable: %v", keyStore.VFSPath())
	}

	configPath, err := c.ClusterRegistry.ConfigurationPath(clusterName)
	if err != nil {
		return err
	}
	if vfs.IsClusterReadable(configPath) {
		c.Cluster.Spec.ConfigStore = configPath.Path()
	} else {
		// We do support this...
	}

	if c.Cluster.Spec.KubernetesVersion == "" {
		stableURL := "https://storage.googleapis.com/kubernetes-release/release/stable.txt"
		b, err := vfs.Context.ReadFile(stableURL)
		if err != nil {
			return fmt.Errorf("--kubernetes-version not specified, and unable to download latest version from %q: %v", stableURL, err)
		}
		latestVersion := strings.TrimSpace(string(b))
		glog.Infof("Using kubernetes latest stable version: %s", latestVersion)

		c.Cluster.Spec.KubernetesVersion = latestVersion
		//return fmt.Errorf("Must either specify a KubernetesVersion (-kubernetes-version) or provide an asset with the release bundle")
	}

	// Normalize k8s version
	versionWithoutV := strings.TrimSpace(c.Cluster.Spec.KubernetesVersion)
	if strings.HasPrefix(versionWithoutV, "v") {
		versionWithoutV = versionWithoutV[1:]
	}
	if c.Cluster.Spec.KubernetesVersion != versionWithoutV {
		glog.Warningf("Normalizing kubernetes version: %q -> %q", c.Cluster.Spec.KubernetesVersion, versionWithoutV)
		c.Cluster.Spec.KubernetesVersion = versionWithoutV
	}

	if len(c.Assets) == 0 {
		//defaultReleaseAsset := fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/v%s/kubernetes-server-linux-amd64.tar.gz", c.Config.KubernetesVersion)
		//glog.Infof("Adding default kubernetes release asset: %s", defaultReleaseAsset)

		defaultKubeletAsset := fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/v%s/bin/linux/amd64/kubelet", c.Cluster.Spec.KubernetesVersion)
		glog.Infof("Adding default kubelet release asset: %s", defaultKubeletAsset)

		defaultKubectlAsset := fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/v%s/bin/linux/amd64/kubectl", c.Cluster.Spec.KubernetesVersion)
		glog.Infof("Adding default kubelet release asset: %s", defaultKubectlAsset)

		// TODO: Verify assets exist, get the hash (that will check that KubernetesVersion is valid)

		c.Assets = append(c.Assets, defaultKubeletAsset, defaultKubectlAsset)
	}

	if c.NodeUpSource == "" {
		location := "https://kubeupv2.s3.amazonaws.com/nodeup/nodeup-1.3.tar.gz"
		glog.Infof("Using default nodeup location: %q", location)
		c.NodeUpSource = location
	}

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

	if c.Cluster.Spec.MasterPublicName != "" {
		tags["_master_dns"] = struct{}{}
	}

	l.AddTypes(map[string]interface{}{
		"keypair": &fitasks.Keypair{},
		"secret":  &fitasks.Secret{},
	})

	cloud, err := BuildCloud(c.Cluster)
	if err != nil {
		return err
	}

	region := ""
	project := ""

	switch c.Cluster.Spec.CloudProvider {
	case "gce":
		{
			gceCloud := cloud.(*gce.GCECloud)
			region = gceCloud.Region
			project = gceCloud.Project

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
		}

	case "aws":
		{
			awsCloud := cloud.(*awsup.AWSCloud)
			region = awsCloud.Region

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

			if c.SSHPublicKey == "" {
				return fmt.Errorf("SSH public key must be specified when running with AWS")
			}

			l.TemplateFunctions["MachineTypeInfo"] = awsup.GetMachineTypeInfo
		}

	default:
		return fmt.Errorf("unknown CloudProvider %q", c.Cluster.Spec.CloudProvider)
	}

	tf := &TemplateFunctions{
		cluster: c.Cluster,
	}

	l.Tags = tags
	l.WorkDir = c.OutDir
	l.ModelStore = c.ModelStore
	l.NodeModel = c.NodeModel

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
		for _, group := range c.nodes {
			if group.Spec.MaxSize != nil {
				count += *group.Spec.MaxSize
			} else if group.Spec.MinSize != nil {
				count += *group.Spec.MinSize
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
	l.TemplateFunctions["NodeUpSource"] = func() string {
		return c.NodeUpSource
	}
	l.TemplateFunctions["NodeUpSourceHash"] = func() string {
		return ""
	}
	l.TemplateFunctions["ClusterLocation"] = func() (string, error) {
		configPath, err := c.ClusterRegistry.ConfigurationPath(clusterName)
		if err != nil {
			return "", err
		}
		return configPath.Path(), nil
	}
	l.TemplateFunctions["Assets"] = func() []string {
		return c.Assets
	}

	l.TemplateFunctions["Base64Encode"] = func(s string) string {
		return base64.StdEncoding.EncodeToString([]byte(s))
	}
	l.TemplateFunctions["ClusterName"] = func() string {
		return clusterName
	}
	l.TemplateFunctions["replace"] = func(s, find, replace string) string {
		return strings.Replace(s, find, replace, -1)
	}
	l.TemplateFunctions["join"] = func(a []string, sep string) string {
		return strings.Join(a, sep)
	}

	tf.AddTo(l.TemplateFunctions)

	l.OptionsLoader = loader.NewOptionsLoader(l.TemplateFunctions)

	if c.SSHPublicKey != "" {
		authorized, err := ioutil.ReadFile(c.SSHPublicKey)
		if err != nil {
			return fmt.Errorf("error reading SSH key file %q: %v", c.SSHPublicKey, err)
		}

		l.Resources["ssh-public-key"] = fi.NewStringResource(string(authorized))
	}

	completed, err := l.BuildCompleteSpec(&c.Cluster.Spec, c.ModelStore, c.Models)
	if err != nil {
		return fmt.Errorf("error building complete spec: %v", err)
	}
	l.cluster = &api.Cluster{}
	*l.cluster = *c.Cluster
	l.cluster.Spec = *completed
	tf.cluster = l.cluster

	err = l.cluster.Validate()
	if err != nil {
		return fmt.Errorf("Completed cluster failed validation: %v", err)
	}

	taskMap, err := l.BuildTasks(c.ModelStore, c.Models)
	if err != nil {
		return fmt.Errorf("error building tasks: %v", err)
	}

	err = c.ClusterRegistry.WriteCompletedConfig(l.cluster)
	if err != nil {
		return fmt.Errorf("error writing completed cluster spec: %v", err)
	}

	var target fi.Target

	switch c.Target {
	case "direct":
		switch c.Cluster.Spec.CloudProvider {
		case "gce":
			target = gce.NewGCEAPITarget(cloud.(*gce.GCECloud))
		case "aws":
			target = awsup.NewAWSAPITarget(cloud.(*awsup.AWSCloud))
		default:
			return fmt.Errorf("direct configuration not supported with CloudProvider:%q", c.Cluster.Spec.CloudProvider)
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
func (c *CreateClusterCmd) populateNodeSets() ([]*api.InstanceGroup, error) {
	var results []*api.InstanceGroup
	for _, src := range c.InstanceGroups {
		if src.IsMaster() {
			continue
		}
		n := &api.InstanceGroup{}
		*n = *src

		if n.Spec.MachineType == "" {
			n.Spec.MachineType = c.defaultMachineType()
		}

		if n.Spec.Image == "" {
			n.Spec.Image = c.defaultImage()
		}

		if len(n.Spec.Zones) == 0 {
			for _, z := range c.Cluster.Spec.Zones {
				n.Spec.Zones = append(n.Spec.Zones, z.Name)
			}
		}
		results = append(results, n)
	}
	return results, nil
}

// populateMasters returns the Masters with values populated from defaults or top-level config
func (c *CreateClusterCmd) populateMasters() ([]*api.InstanceGroup, error) {
	var results []*api.InstanceGroup
	for _, src := range c.InstanceGroups {
		if !src.IsMaster() {
			continue
		}

		m := &api.InstanceGroup{}
		*m = *src

		if len(src.Spec.Zones) == 0 {
			return nil, fmt.Errorf("Master InstanceGroup %s did not specify any Zones", src.Name)
		}

		if m.Spec.MachineType == "" {
			m.Spec.MachineType = c.defaultMachineType()
		}

		if m.Spec.Image == "" {
			m.Spec.Image = c.defaultImage()
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
	cluster := c.Cluster
	switch cluster.Spec.CloudProvider {
	case "aws":
		return DefaultNodeTypeAWS
	case "gce":
		return DefaultNodeTypeGCE
	default:
		glog.V(2).Infof("Cannot set default MachineType for CloudProvider=%q", cluster.Spec.CloudProvider)
		return ""
	}
}

// defaultImage returns the default Image, based on the cloudprovider
func (c *CreateClusterCmd) defaultImage() string {
	// TODO: Use spec
	cluster := c.Cluster
	switch cluster.Spec.CloudProvider {
	case "aws":
		return "282335181503/k8s-1.3-debian-jessie-amd64-hvm-ebs-2016-06-18"
	default:
		glog.V(2).Infof("Cannot set default Image for CloudProvider=%q", cluster.Spec.CloudProvider)
		return ""
	}
}

func (c *CreateClusterCmd) assignSubnets() error {
	cluster := c.Cluster
	if cluster.Spec.NonMasqueradeCIDR == "" {
		glog.Warningf("NonMasqueradeCIDR not set; can't auto-assign dependent subnets")
		return nil
	}

	_, nonMasqueradeCIDR, err := net.ParseCIDR(cluster.Spec.NonMasqueradeCIDR)
	if err != nil {
		return fmt.Errorf("error parsing NonMasqueradeCIDR %q: %v", cluster.Spec.NonMasqueradeCIDR, err)
	}
	nmOnes, nmBits := nonMasqueradeCIDR.Mask.Size()

	if cluster.Spec.KubeControllerManager == nil {
		cluster.Spec.KubeControllerManager = &api.KubeControllerManagerConfig{}
	}

	if cluster.Spec.KubeControllerManager.ClusterCIDR == "" {
		// Allocate as big a range as possible: the NonMasqueradeCIDR mask + 1, with a '1' in the extra bit
		ip := nonMasqueradeCIDR.IP.Mask(nonMasqueradeCIDR.Mask)

		ip4 := ip.To4()
		if ip4 != nil {
			n := binary.BigEndian.Uint32(ip4)
			n += uint32(1 << uint(nmBits-nmOnes-1))
			ip = make(net.IP, len(ip4))
			binary.BigEndian.PutUint32(ip, n)
		} else {
			return fmt.Errorf("IPV6 subnet computations not yet implements")
		}

		cidr := net.IPNet{IP: ip, Mask: net.CIDRMask(nmOnes+1, nmBits)}
		cluster.Spec.KubeControllerManager.ClusterCIDR = cidr.String()
		glog.V(2).Infof("Defaulted KubeControllerManager.ClusterCIDR to %v", cluster.Spec.KubeControllerManager.ClusterCIDR)
	}

	if cluster.Spec.ServiceClusterIPRange == "" {
		// Allocate from the '0' subnet; but only carve off 1/4 of that (i.e. add 1 + 2 bits to the netmask)
		cidr := net.IPNet{IP: nonMasqueradeCIDR.IP.Mask(nonMasqueradeCIDR.Mask), Mask: net.CIDRMask(nmOnes+3, nmBits)}
		cluster.Spec.ServiceClusterIPRange = cidr.String()
		glog.V(2).Infof("Defaulted ServiceClusterIPRange to %v", cluster.Spec.ServiceClusterIPRange)
	}

	return nil
}
