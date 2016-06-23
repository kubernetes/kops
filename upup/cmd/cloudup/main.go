package main

import (
	goflag "flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/spf13/pflag"
	"io/ioutil"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup"
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
	"os/exec"
	"path"
	"strings"
)

func main() {
	executableLocation, err := exec.LookPath(os.Args[0])
	if err != nil {
		glog.Fatalf("Cannot determine location of cloudup tool: %q.  Please report this problem!", os.Args[0])
	}

	modelsBaseDirDefault := path.Join(path.Dir(executableLocation), "models")

	dryrun := pflag.Bool("dryrun", false, "Don't create cloud resources; just show what would be done")
	target := pflag.String("target", "direct", "Target - direct, terraform")
	//configFile := pflag.String("conf", "", "Configuration file to load")
	modelsBaseDir := pflag.String("modelstore", modelsBaseDirDefault, "Source directory where models are stored")
	models := pflag.String("model", "proto,cloudup", "Models to apply (separate multiple models with commas)")
	nodeModel := pflag.String("nodemodel", "nodeup", "Model to use for node configuration")
	stateLocation := pflag.String("state", "", "Location to use to store configuration state")

	cloudProvider := pflag.String("cloud", "", "Cloud provider to use - gce, aws")

	zones := pflag.String("zones", "", "Zones in which to run nodes")
	masterZones := pflag.String("master-zones", "", "Zones in which to run masters (must be an odd number)")

	project := pflag.String("project", "", "Project to use (must be set on GCE)")
	clusterName := pflag.String("name", "", "Name for cluster")
	kubernetesVersion := pflag.String("kubernetes-version", "", "Version of kubernetes to run (defaults to latest)")

	sshPublicKey := pflag.String("ssh-public-key", "~/.ssh/id_rsa.pub", "SSH public key to use")

	nodeSize := pflag.String("node-size", "", "Set instance size for nodes")

	masterSize := pflag.String("master-size", "", "Set instance size for masters")

	nodeCount := pflag.Int("node-count", 0, "Set the number of nodes")

	dnsZone := pflag.String("dns-zone", "", "DNS hosted zone to use (defaults to last two components of cluster name)")
	outDir := pflag.String("out", "", "Path to write any local output")

	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	pflag.Parse()
	goflag.CommandLine.Parse([]string{})

	isDryrun := false
	if *dryrun {
		isDryrun = true
		*target = "dryrun"
	}

	if *stateLocation == "" {
		glog.Errorf("--state is required")
		os.Exit(1)
	}

	statePath, err := vfs.Context.BuildVfsPath(*stateLocation)
	if err != nil {
		glog.Errorf("error building state location: %v", err)
		os.Exit(1)
	}

	if *outDir == "" {
		*outDir = "out"
	}

	stateStore, err := fi.NewVFSStateStore(statePath, isDryrun)
	if err != nil {
		glog.Errorf("error building state store: %v", err)
		os.Exit(1)
	}

	// TODO: Replace all these with a direct binding to the CloudConfig
	// (we have plenty of reflection helpers if one isn't already available!)
	config := &cloudup.CloudConfig{}
	err = stateStore.ReadConfig(config)
	if err != nil {
		glog.Errorf("error loading configuration: %v", err)
		os.Exit(1)
	}

	if *zones != "" {
		existingZones := make(map[string]*cloudup.ZoneConfig)
		for _, zone := range config.NodeZones {
			existingZones[zone.Name] = zone
		}

		for _, zone := range parseZoneList(*zones) {
			if existingZones[zone] == nil {
				config.NodeZones = append(config.NodeZones, &cloudup.ZoneConfig{
					Name: zone,
				})
			}
		}
	}

	if *masterZones == "" {
		if len(config.MasterZones) == 0 {
			for _, nodeZone := range config.NodeZones {
				config.MasterZones = append(config.MasterZones, nodeZone.Name)
			}
		}
	} else {
		config.MasterZones = parseZoneList(*masterZones)
	}

	if *nodeSize != "" {
		config.NodeMachineType = *nodeSize
	}
	if *nodeCount != 0 {
		config.NodeCount = *nodeCount
	}

	if *masterSize != "" {
		config.MasterMachineType = *masterSize
	}

	if *dnsZone != "" {
		config.DNSZone = *dnsZone
	}

	if *cloudProvider != "" {
		config.CloudProvider = *cloudProvider
	}

	if *project != "" {
		config.Project = *project
	}

	if *clusterName != "" {
		config.ClusterName = *clusterName
	}

	if *kubernetesVersion != "" {
		config.KubernetesVersion = *kubernetesVersion
	}

	err = config.PerformAssignments()
	if err != nil {
		glog.Errorf("error populating configuration: %v", err)
		os.Exit(1)
	}

	err = stateStore.WriteConfig(config)
	if err != nil {
		glog.Errorf("error writing updated configuration: %v", err)
		os.Exit(1)
	}

	if *sshPublicKey != "" {
		*sshPublicKey = utils.ExpandPath(*sshPublicKey)
	}

	cmd := &CreateClusterCmd{
		Config:       config,
		ModelStore:   *modelsBaseDir,
		Models:       strings.Split(*models, ","),
		StateStore:   stateStore,
		Target:       *target,
		NodeModel:    *nodeModel,
		SSHPublicKey: *sshPublicKey,
		OutDir:       *outDir,
	}

	//if *configFile != "" {
	//	//confFile := path.Join(cmd.StateDir, "kubernetes.yaml")
	//	err := cmd.LoadConfig(configFile)
	//	if err != nil {
	//		glog.Errorf("error loading config: %v", err)
	//		os.Exit(1)
	//	}
	//}

	err = cmd.Run()
	if err != nil {
		glog.Errorf("error running command: %v", err)
		os.Exit(1)
	}

	glog.Infof("Completed successfully")
}

func parseZoneList(s string) []string {
	var filtered []string
	for _, v := range strings.Split(s, ",") {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		v = strings.ToLower(v)
		filtered = append(filtered, v)
	}
	return filtered
}

type CreateClusterCmd struct {
	// Config is the cluster configuration
	Config *cloudup.CloudConfig
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
}

func (c *CreateClusterCmd) LoadConfig(configFile string) error {
	conf, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("error loading configuration file %q: %v", configFile, err)
	}
	err = utils.YamlUnmarshal(conf, c.Config)
	if err != nil {
		return fmt.Errorf("error parsing configuration file %q: %v", configFile, err)
	}
	return nil
}

func (c *CreateClusterCmd) Run() error {
	// TODO: Make these configurable?
	useMasterASG := true
	useMasterLB := false

	// We (currently) have to use protokube with ASGs
	useProtokube := useMasterASG

	if c.Config.NodeUp == nil {
		c.Config.NodeUp = &cloudup.NodeUpConfig{}
	}

	if c.Config.ClusterName == "" {
		return fmt.Errorf("--name is required (e.g. mycluster.myzone.com)")
	}

	if c.Config.MasterPublicName == "" {
		c.Config.MasterPublicName = "api." + c.Config.ClusterName
	}
	if c.Config.DNSZone == "" {
		tokens := strings.Split(c.Config.MasterPublicName, ".")
		c.Config.DNSZone = strings.Join(tokens[len(tokens)-2:], ".")
		glog.Infof("Defaulting DNS zone to: %s", c.Config.DNSZone)
	}

	if len(c.Config.NodeZones) == 0 {
		return fmt.Errorf("must specify at least one NodeZone")
	}

	if len(c.Config.MasterZones) == 0 {
		return fmt.Errorf("must specify at least one MasterZone")
	}

	// Check for master zone duplicates
	{
		masterZones := make(map[string]bool)
		for _, z := range c.Config.MasterZones {
			if masterZones[z] {
				return fmt.Errorf("MasterZones contained a duplicate value:  %v", z)
			}
			masterZones[z] = true
		}
	}

	// Check for node zone duplicates
	{
		nodeZones := make(map[string]bool)
		for _, z := range c.Config.NodeZones {
			if nodeZones[z.Name] {
				return fmt.Errorf("NodeZones contained a duplicate value:  %v", z)
			}
			nodeZones[z.Name] = true
		}
	}

	if (len(c.Config.MasterZones) % 2) == 0 {
		// Not technically a requirement, but doesn't really make sense to allow
		return fmt.Errorf("There should be an odd number of master-zones, for etcd's quorum.  Hint: Use -zone and -master-zone to declare node zones and master zones separately.")
	}

	if c.StateStore == nil {
		return fmt.Errorf("StateStore is required")
	}

	if c.Config.CloudProvider == "" {
		return fmt.Errorf("--cloud is required (e.g. aws, gce)")
	}

	tags := make(map[string]struct{})

	l := &cloudup.Loader{}
	l.Init()

	keyStore := c.StateStore.CA()
	secretStore := c.StateStore.Secrets()

	if vfs.IsClusterReadable(secretStore.VFSPath()) {
		vfsPath := secretStore.VFSPath()
		c.Config.SecretStore = vfsPath.Path()
		if s3Path, ok := vfsPath.(*vfs.S3Path); ok {
			if c.Config.MasterPermissions == nil {
				c.Config.MasterPermissions = &cloudup.CloudPermissions{}
			}
			c.Config.MasterPermissions.AddS3Bucket(s3Path.Bucket())
			if c.Config.NodePermissions == nil {
				c.Config.NodePermissions = &cloudup.CloudPermissions{}
			}
			c.Config.NodePermissions.AddS3Bucket(s3Path.Bucket())
		}
	} else {
		// We could implement this approach, but it seems better to get all clouds using cluster-readable storage
		return fmt.Errorf("secrets path is not cluster readable: %v", secretStore.VFSPath())
	}

	if vfs.IsClusterReadable(keyStore.VFSPath()) {
		vfsPath := keyStore.VFSPath()
		c.Config.KeyStore = vfsPath.Path()
		if s3Path, ok := vfsPath.(*vfs.S3Path); ok {
			if c.Config.MasterPermissions == nil {
				c.Config.MasterPermissions = &cloudup.CloudPermissions{}
			}
			c.Config.MasterPermissions.AddS3Bucket(s3Path.Bucket())
			if c.Config.NodePermissions == nil {
				c.Config.NodePermissions = &cloudup.CloudPermissions{}
			}
			c.Config.NodePermissions.AddS3Bucket(s3Path.Bucket())
		}
	} else {
		// We could implement this approach, but it seems better to get all clouds using cluster-readable storage
		return fmt.Errorf("keyStore path is not cluster readable: %v", keyStore.VFSPath())
	}

	if vfs.IsClusterReadable(c.StateStore.VFSPath()) {
		c.Config.ConfigStore = c.StateStore.VFSPath().Path()
	} else {
		// We do support this...
	}

	if c.Config.KubernetesVersion == "" {
		stableURL := "https://storage.googleapis.com/kubernetes-release/release/stable.txt"
		b, err := vfs.Context.ReadFile(stableURL)
		if err != nil {
			return fmt.Errorf("--kubernetes-version not specified, and unable to download latest version from %q: %v", stableURL, err)
		}
		latestVersion := strings.TrimSpace(string(b))
		glog.Infof("Using kubernetes latest stable version: %s", latestVersion)

		c.Config.KubernetesVersion = latestVersion
		//return fmt.Errorf("Must either specify a KubernetesVersion (-kubernetes-version) or provide an asset with the release bundle")
	}

	// Normalize k8s version
	versionWithoutV := strings.TrimSpace(c.Config.KubernetesVersion)
	if strings.HasPrefix(versionWithoutV, "v") {
		versionWithoutV = versionWithoutV[1:]
	}
	if c.Config.KubernetesVersion != versionWithoutV {
		glog.Warningf("Normalizing kubernetes version: %q -> %q", c.Config.KubernetesVersion, versionWithoutV)
		c.Config.KubernetesVersion = versionWithoutV
	}

	if len(c.Config.Assets) == 0 {
		//defaultReleaseAsset := fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/v%s/kubernetes-server-linux-amd64.tar.gz", c.Config.KubernetesVersion)
		//glog.Infof("Adding default kubernetes release asset: %s", defaultReleaseAsset)

		defaultKubeletAsset := fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/v%s/bin/linux/amd64/kubelet", c.Config.KubernetesVersion)
		glog.Infof("Adding default kubelet release asset: %s", defaultKubeletAsset)

		defaultKubectlAsset := fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/v%s/bin/linux/amd64/kubectl", c.Config.KubernetesVersion)
		glog.Infof("Adding default kubelet release asset: %s", defaultKubectlAsset)

		// TODO: Verify assets exist, get the hash (that will check that KubernetesVersion is valid)

		c.Config.Assets = append(c.Config.Assets, defaultKubeletAsset, defaultKubectlAsset)
	}

	if c.Config.NodeUp.Location == "" {
		location := "https://kubeupv2.s3.amazonaws.com/nodeup/nodeup.tar.gz"
		glog.Infof("Using default nodeup location: %q", location)
		c.Config.NodeUp.Location = location
	}

	var cloud fi.Cloud

	var project string

	checkExisting := true

	c.Config.NodeUpTags = append(c.Config.NodeUpTags, "_jessie", "_debian_family", "_systemd")

	if useProtokube {
		tags["_protokube"] = struct{}{}
		c.Config.NodeUpTags = append(c.Config.NodeUpTags, "_protokube")
	} else {
		tags["_not_protokube"] = struct{}{}
		c.Config.NodeUpTags = append(c.Config.NodeUpTags, "_not_protokube")
	}

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

	if c.Config.MasterPublicName != "" {
		tags["_master_dns"] = struct{}{}
	}

	l.AddTypes(map[string]interface{}{
		"keypair": &fitasks.Keypair{},
		"secret":  &fitasks.Secret{},
	})

	switch c.Config.CloudProvider {
	case "gce":
		{
			glog.Fatalf("GCE is (probably) not working currently - please ping @justinsb for cleanup")
			tags["_gce"] = struct{}{}
			c.Config.NodeUpTags = append(c.Config.NodeUpTags, "_gce")

			l.AddTypes(map[string]interface{}{
				"persistentDisk":       &gcetasks.PersistentDisk{},
				"instance":             &gcetasks.Instance{},
				"instanceTemplate":     &gcetasks.InstanceTemplate{},
				"network":              &gcetasks.Network{},
				"managedInstanceGroup": &gcetasks.ManagedInstanceGroup{},
				"firewallRule":         &gcetasks.FirewallRule{},
				"ipAddress":            &gcetasks.IPAddress{},
			})

			// For now a zone to be specified...
			// This will be replace with a region when we go full HA
			zone := c.Config.NodeZones[0]
			if zone.Name == "" {
				return fmt.Errorf("Must specify a zone (use -zone)")
			}
			tokens := strings.Split(zone.Name, "-")
			if len(tokens) <= 2 {
				return fmt.Errorf("Invalid Zone: %v", zone.Name)
			}
			region := tokens[0] + "-" + tokens[1]

			if c.Config.Region != "" && region != c.Config.Region {
				return fmt.Errorf("zone %q is not in region %q", zone, c.Config.Region)
			}
			c.Config.Region = region

			project = c.Config.Project
			if project == "" {
				return fmt.Errorf("project is required for GCE")
			}
			gceCloud, err := gce.NewGCECloud(region, project)
			if err != nil {
				return err
			}
			cloud = gceCloud
		}

	case "aws":
		{
			tags["_aws"] = struct{}{}
			c.Config.NodeUpTags = append(c.Config.NodeUpTags, "_aws")

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

			if len(c.Config.NodeZones) == 0 {
				// TODO: Auto choose zones from region?
				return fmt.Errorf("Must specify a zone (use -zone)")
			}
			if len(c.Config.MasterZones) == 0 {
				return fmt.Errorf("Must specify a master zones")
			}

			nodeZones := make(map[string]bool)
			for _, zone := range c.Config.NodeZones {
				if len(zone.Name) <= 2 {
					return fmt.Errorf("Invalid AWS zone: %q", zone.Name)
				}

				nodeZones[zone.Name] = true

				zoneRegion := zone.Name[:len(zone.Name)-1]
				if c.Config.Region != "" && zoneRegion != c.Config.Region {
					return fmt.Errorf("Clusters cannot span multiple regions")
				}

				c.Config.Region = zoneRegion
			}

			for _, zone := range c.Config.MasterZones {
				if !nodeZones[zone] {
					// We could relax this, but this seems like a reasonable constraint
					return fmt.Errorf("All MasterZones must (currently) also be NodeZones")
				}
			}

			err := awsup.ValidateRegion(c.Config.Region)
			if err != nil {
				return err
			}

			if c.SSHPublicKey == "" {
				return fmt.Errorf("SSH public key must be specified when running with AWS")
			}

			cloudTags := map[string]string{awsup.TagClusterName: c.Config.ClusterName}

			awsCloud, err := awsup.NewAWSCloud(c.Config.Region, cloudTags)
			if err != nil {
				return err
			}

			var nodeZoneNames []string
			for _, z := range c.Config.NodeZones {
				nodeZoneNames = append(nodeZoneNames, z.Name)
			}
			err = awsCloud.ValidateZones(nodeZoneNames)
			if err != nil {
				return err
			}
			cloud = awsCloud

			l.TemplateFunctions["MachineTypeInfo"] = awsup.GetMachineTypeInfo
		}

	default:
		return fmt.Errorf("unknown CloudProvider %q", c.Config.CloudProvider)
	}

	l.Tags = tags
	l.WorkDir = c.OutDir
	l.ModelStore = c.ModelStore
	l.NodeModel = c.NodeModel
	l.OptionsLoader = loader.NewOptionsLoader(c.Config)

	l.TemplateFunctions["HasTag"] = func(tag string) bool {
		_, found := l.Tags[tag]
		return found
	}

	// TODO: Sort this out...
	l.OptionsLoader.TemplateFunctions["HasTag"] = l.TemplateFunctions["HasTag"]

	l.TemplateFunctions["CA"] = func() fi.CAStore {
		return keyStore
	}
	l.TemplateFunctions["Secrets"] = func() fi.SecretStore {
		return secretStore
	}

	if c.SSHPublicKey != "" {
		authorized, err := ioutil.ReadFile(c.SSHPublicKey)
		if err != nil {
			return fmt.Errorf("error reading SSH key file %q: %v", c.SSHPublicKey, err)
		}

		l.Resources["ssh-public-key"] = fi.NewStringResource(string(authorized))
	}

	taskMap, err := l.Build(c.ModelStore, c.Models)
	if err != nil {
		glog.Exitf("error building: %v", err)
	}

	var target fi.Target

	switch c.Target {
	case "direct":
		switch c.Config.CloudProvider {
		case "gce":
			target = gce.NewGCEAPITarget(cloud.(*gce.GCECloud))
		case "aws":
			target = awsup.NewAWSAPITarget(cloud.(*awsup.AWSCloud))
		default:
			return fmt.Errorf("direct configuration not supported with CloudProvider:%q", c.Config.CloudProvider)
		}

	case "terraform":
		checkExisting = false
		outDir := path.Join(c.OutDir, "terraform")
		target = terraform.NewTerraformTarget(cloud, c.Config.Region, project, outDir)

	case "dryrun":
		target = fi.NewDryRunTarget(os.Stdout)
	default:
		return fmt.Errorf("unsupported target type %q", c.Target)
	}

	context, err := fi.NewContext(target, cloud, keyStore, secretStore, checkExisting)
	if err != nil {
		glog.Exitf("error building context: %v", err)
	}
	defer context.Close()

	err = context.RunTasks(taskMap)
	if err != nil {
		glog.Exitf("error running tasks: %v", err)
	}

	err = target.Finish(taskMap)
	if err != nil {
		glog.Exitf("error closing target: %v", err)
	}

	return nil
}
