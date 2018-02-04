package imagebuilder

import (
	"strings"

	"github.com/golang/glog"
)

type Config struct {
	Cloud         string
	TemplatePath  string
	SetupCommands [][]string

	BootstrapVZRepo   string
	BootstrapVZBranch string

	SSHUsername   string
	SSHPublicKey  string
	SSHPrivateKey string

	// Tags to add to the image
	Tags map[string]string
}

func (c *Config) InitDefaults() {
	c.BootstrapVZRepo = "https://github.com/andsens/bootstrap-vz.git"
	c.BootstrapVZBranch = "master"

	c.SSHUsername = "admin"
	c.SSHPublicKey = "~/.ssh/id_rsa.pub"
	c.SSHPrivateKey = "~/.ssh/id_rsa"

	setupCommands := []string{
		"sudo apt-get update",
		"sudo apt-get install --yes git python debootstrap python-pip kpartx parted",
		"sudo pip install termcolor jsonschema fysom docopt pyyaml boto boto3",
	}
	for _, cmd := range setupCommands {
		c.SetupCommands = append(c.SetupCommands, strings.Split(cmd, " "))
	}
}

type AWSConfig struct {
	Config

	Region          string
	ImageID         string
	InstanceType    string
	SSHKeyName      string
	SubnetID        string
	SecurityGroupID string
}

func (c *AWSConfig) InitDefaults(region string) {
	c.Config.InitDefaults()
	c.InstanceType = "m3.medium"

	if region == "" {
		region = "us-east-1"
	}

	c.Region = region
	switch c.Region {
	case "cn-north-1":
		glog.Infof("Detected cn-north-1 region")
		// A slightly older image, but the newest one we have
		c.ImageID = "ami-da69a1b7"

	// Debian 8.7 images from https://wiki.debian.org/Cloud/AmazonEC2Image/Jessie
	case "ap-northeast-1":
		c.ImageID = "ami-dbc0bcbc"
	case "ap-northeast-2":
		c.ImageID = "ami-6d8b5a03"
	case "ap-south-1":
		c.ImageID = "ami-9a83f5f5"
	case "ap-southeast-1":
		c.ImageID = "ami-0842e96b"
	case "ap-southeast-2":
		c.ImageID = "ami-881317eb"
	case "ca-central-1":
		c.ImageID = "ami-a1fe43c5"
	case "eu-central-1":
		c.ImageID = "ami-5900cc36"
	case "eu-west-1":
		c.ImageID = "ami-402f1a33"
	case "eu-west-2":
		c.ImageID = "ami-87848ee3"
	case "sa-east-1":
		c.ImageID = "ami-b256ccde"
	case "us-east-1":
		c.ImageID = "ami-b14ba7a7"
	case "us-east-2":
		c.ImageID = "ami-b2795cd7"
	case "us-west-1":
		c.ImageID = "ami-94bdeef4"
	case "us-west-2":
		c.ImageID = "ami-221ea342"

	default:
		glog.Warningf("Building in unknown region %q - will require specifying an image, may not work correctly")
	}

	// Not all regions support m3.medium
	switch c.Region {
	case "us-east-2":
		c.InstanceType = "m4.large"
	}
}

type GCEConfig struct {
	Config

	// To create an image on GCE, we have to upload it to a bucket first
	GCSDestination string

	Project     string
	Zone        string
	MachineName string

	MachineType string
	Image       string
}

func (c *GCEConfig) InitDefaults() {
	c.Config.InitDefaults()
	c.MachineName = "k8s-imagebuilder"
	c.Zone = "us-central1-f"
	c.MachineType = "n1-standard-2"
	c.Image = "https://www.googleapis.com/compute/v1/projects/debian-cloud/global/images/debian-8-jessie-v20160329"
}
