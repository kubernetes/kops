package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
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
	"os"
	"path"
	"strings"
)

func main() {
	dryrun := false
	flag.BoolVar(&dryrun, "dryrun", false, "Don't create cloud resources; just show what would be done")
	target := "direct"
	flag.StringVar(&target, "target", target, "Target - direct, terraform")
	configFile := ""
	flag.StringVar(&configFile, "conf", configFile, "Configuration file to load")
	modelDir := "models/cloudup"
	flag.StringVar(&modelDir, "model", modelDir, "Source directory to use as model")
	stateDir := "./state"
	flag.StringVar(&stateDir, "state", stateDir, "Directory to use to store local state")
	nodeModelDir := "models/nodeup"
	flag.StringVar(&nodeModelDir, "nodemodel", nodeModelDir, "Source directory to use as model for node configuration")

	// TODO: Replace all these with a direct binding to the CloudConfig
	// (we have plenty of reflection helpers if one isn't already available!)
	config := &cloudup.CloudConfig{}
	flag.StringVar(&config.CloudProvider, "cloud", config.CloudProvider, "Cloud provider to use - gce, aws")
	flag.StringVar(&config.Zone, "zone", config.Zone, "Cloud zone to target (warning - will be replaced by region)")
	flag.StringVar(&config.Project, "project", config.Project, "Project to use (must be set on GCE)")
	flag.StringVar(&config.ClusterName, "name", config.ClusterName, "Name for cluster")
	flag.StringVar(&config.KubernetesVersion, "kubernetes-version", config.KubernetesVersion, "Version of kubernetes to run")
	//flag.StringVar(&config.Region, "region", config.Region, "Cloud region to target")

	sshPublicKey := path.Join(os.Getenv("HOME"), ".ssh", "id_rsa.pub")
	flag.StringVar(&sshPublicKey, "ssh-public-key", sshPublicKey, "SSH public key to use")

	flag.Parse()

	if dryrun {
		target = "dryrun"
	}

	cmd := &CreateClusterCmd{
		Config:       config,
		ModelDir:     modelDir,
		StateDir:     stateDir,
		Target:       target,
		NodeModelDir: nodeModelDir,
		SSHPublicKey: sshPublicKey,
	}

	if configFile != "" {
		//confFile := path.Join(cmd.StateDir, "kubernetes.yaml")
		err := cmd.LoadConfig(configFile)
		if err != nil {
			glog.Errorf("error loading config: %v", err)
			os.Exit(1)
		}
	}

	err := cmd.Run()
	if err != nil {
		glog.Errorf("error running command: %v", err)
		os.Exit(1)
	}

	glog.Infof("Completed successfully")
}

type CreateClusterCmd struct {
	// Config is the cluster configuration
	Config *cloudup.CloudConfig
	// ModelDir is the directory in which the cloudup model is found
	ModelDir string
	// StateDir is a directory in which we store state (such as the PKI tree)
	StateDir string
	// Target specifies how we are operating e.g. direct to GCE, or AWS, or dry-run, or terraform
	Target string
	// The directory in which the node model is found
	NodeModelDir string
	// The SSH public key (file) to use
	SSHPublicKey string
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
	if c.StateDir == "" {
		return fmt.Errorf("state dir is required")
	}

	if c.Config.CloudProvider == "" {
		return fmt.Errorf("must specify CloudProvider.  Specify with -cloud")
	}

	tags := make(map[string]struct{})

	l := &cloudup.Loader{}
	l.Init()

	caStore, err := fi.NewFilesystemCAStore(path.Join(c.StateDir, "pki"))
	if err != nil {
		return fmt.Errorf("error building CA store: %v", err)
	}
	secretStore, err := fi.NewFilesystemSecretStore(path.Join(c.StateDir, "secrets"))
	if err != nil {
		return fmt.Errorf("error building secret store: %v", err)
	}

	if len(c.Config.Assets) == 0 {
		if c.Config.KubernetesVersion == "" {
			return fmt.Errorf("Must either specify a KubernetesVersion (-kubernetes-version) or provide an asset with the release bundle")
		}
		defaultReleaseAsset := fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/v%s/kubernetes-server-linux-amd64.tar.gz", c.Config.KubernetesVersion)
		glog.Infof("Adding default kubernetes release asset: %s", defaultReleaseAsset)
		// TODO: Verify it exists, get the hash (that will check that KubernetesVersion is valid)
		c.Config.Assets = append(c.Config.Assets, defaultReleaseAsset)
	}

	if c.Config.NodeUp.Location == "" {
		location := "https://kubeupv2.s3.amazonaws.com/nodeup/nodeup.tar.gz"
		glog.Infof("Using default nodeup location: %q", location)
		c.Config.NodeUp.Location = location
	}

	var cloud fi.Cloud

	var project string
	var region string

	checkExisting := true

	c.Config.NodeUpTags = append(c.Config.NodeUpTags, "_jessie", "_debian_family", "_systemd")

	l.AddTypes(map[string]interface{}{
		"keypair": &fitasks.Keypair{},
	})

	switch c.Config.CloudProvider {
	case "gce":
		{
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
			zone := c.Config.Zone
			if zone == "" {
				return fmt.Errorf("Must specify a zone (use -zone)")
			}
			tokens := strings.Split(zone, "-")
			if len(tokens) <= 2 {
				return fmt.Errorf("Invalid Zone: %v", zone)
			}
			region = tokens[0] + "-" + tokens[1]

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
				"autoscalingGroup":            &awstasks.AutoscalingGroup{},
				"dhcpOptions":                 &awstasks.DHCPOptions{},
				"elasticIP":                   &awstasks.ElasticIP{},
				"iamInstanceProfile":          &awstasks.IAMInstanceProfile{},
				"iamInstanceProfileRole":      &awstasks.IAMInstanceProfileRole{},
				"iamRole":                     &awstasks.IAMRole{},
				"iamRolePolicy":               &awstasks.IAMRolePolicy{},
				"instance":                    &awstasks.Instance{},
				"instanceElasticIPAttachment": &awstasks.InstanceElasticIPAttachment{},
				"instanceVolumeAttachment":    &awstasks.InstanceVolumeAttachment{},
				"internetGateway":             &awstasks.InternetGateway{},
				"internetGatewayAttachment":   &awstasks.InternetGatewayAttachment{},
				"ebsVolume":                   &awstasks.EBSVolume{},
				"route":                       &awstasks.Route{},
				"routeTable":                  &awstasks.RouteTable{},
				"routeTableAssociation":       &awstasks.RouteTableAssociation{},
				"securityGroup":               &awstasks.SecurityGroup{},
				"securityGroupIngress":        &awstasks.SecurityGroupIngress{},
				"sshKey":                      &awstasks.SSHKey{},
				"subnet":                      &awstasks.Subnet{},
				"vpc":                         &awstasks.VPC{},
				"vpcDHDCPOptionsAssociation": &awstasks.VPCDHCPOptionsAssociation{},
			})

			// For now a zone to be specified...
			// This will be replace with a region when we go full HA
			zone := c.Config.Zone
			if zone == "" {
				return fmt.Errorf("Must specify a zone (use -zone)")
			}
			if len(zone) <= 2 {
				return fmt.Errorf("Invalid AWS zone: %v", zone)
			}

			if c.SSHPublicKey == "" {
				return fmt.Errorf("SSH public key must be specified when running with AWS")
			}

			region := zone[:len(zone)-1]
			c.Config.Region = region

			if c.Config.ClusterName == "" {
				return fmt.Errorf("ClusterName is required for AWS")
			}

			cloudTags := map[string]string{"KubernetesCluster": c.Config.ClusterName}

			awsCloud, err := awsup.NewAWSCloud(region, cloudTags)
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
	l.StateDir = c.StateDir
	l.NodeModelDir = c.NodeModelDir
	l.OptionsLoader = loader.NewOptionsLoader(c.Config)

	l.TemplateFunctions["CA"] = func() fi.CAStore {
		return caStore
	}
	l.TemplateFunctions["Secrets"] = func() fi.SecretStore {
		return secretStore
	}
	l.TemplateFunctions["GetOrCreateSecret"] = func(id string) (string, error) {
		secret, err := secretStore.FindSecret(id)
		if err != nil {
			return "", fmt.Errorf("error finding secret %q: %v", id, err)
		}
		if secret == nil {
			secret, err = secretStore.CreateSecret(id)
			if err != nil {
				return "", fmt.Errorf("error creating secret %q: %v", id, err)
			}
		}
		return secret.AsString()
	}

	if c.SSHPublicKey != "" {
		authorized, err := ioutil.ReadFile(c.SSHPublicKey)
		if err != nil {
			return fmt.Errorf("error reading SSH key file %q: %v", c.SSHPublicKey, err)
		}

		l.Resources["ssh-public-key"] = fi.NewStringResource(string(authorized))
	}

	taskMap, err := l.Build(c.ModelDir)
	if err != nil {
		glog.Exitf("error building: %v", err)
	}

	if c.Config.ClusterName == "" {
		return fmt.Errorf("ClusterName is required")
	}

	if c.Config.Zone == "" {
		return fmt.Errorf("Zone is required")
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
		target = terraform.NewTerraformTarget(region, project, os.Stdout)

	case "dryrun":
		target = fi.NewDryRunTarget(os.Stdout)
	default:
		return fmt.Errorf("unsupported target type %q", c.Target)
	}

	context, err := fi.NewContext(target, cloud, caStore, checkExisting)
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
