/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/storage/v1"
	"io/ioutil"
	"k8s.io/kube-deploy/imagebuilder/pkg/imagebuilder"
	"k8s.io/kube-deploy/imagebuilder/pkg/imagebuilder/executor"
	"net/url"
	"os"
	"path"
	"strings"
)

var flagConfig = flag.String("config", "", "Config file to load")

//var flagRegion = flag.String("region", "", "Cloud region to use")
//var flagImage = flag.String("image", "", "Image to use as builder")
//var flagSSHKey = flag.String("sshkey", "", "Name of SSH key to use")
//var flagInstanceType = flag.String("instancetype", "m3.medium", "Instance type to launch")
//var flagSubnet = flag.String("subnet", "", "Subnet in which to launch")
//var flagSecurityGroup = flag.String("securitygroup", "", "Security group to use for launch")
//var flagTemplatePath = flag.String("template", "", "Path to image template")

var flagUp = flag.Bool("up", true, "Set to create instance (if not found)")
var flagBuild = flag.Bool("build", true, "Set to build image")
var flagTag = flag.Bool("tag", true, "Set to tag image")
var flagPublish = flag.Bool("publish", true, "Set to publish image")
var flagReplicate = flag.Bool("replicate", true, "Set to copy the image to all regions")
var flagDown = flag.Bool("down", true, "Set to shut down instance (if found)")

var flagLocalhost = flag.Bool("localhost", false, "Set to use local machine for execution")

func loadConfig(dest interface{}, src string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return fmt.Errorf("error reading file %q: %v", src, err)
	}

	err = yaml.Unmarshal(data, dest)
	if err != nil {
		return fmt.Errorf("error parsing file %q: %v", src, err)
	}

	return nil
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	flag.Set("alsologtostderr", "true")
	flag.Parse()

	if *flagConfig == "" {
		glog.Exitf("--config must be specified")
	}

	var templateContext interface{}

	config := &imagebuilder.Config{}
	config.InitDefaults()
	err := loadConfig(config, *flagConfig)
	if err != nil {
		glog.Exitf("Error loading config: %v", err)
	}

	var cloud imagebuilder.Cloud
	switch config.Cloud {
	case "aws":
		awsConfig, awsCloud, err := initAWS(*flagLocalhost)
		if err != nil {
			glog.Exitf("%v", err)
		}
		templateContext = awsConfig
		cloud = awsCloud

	case "gce":
		if *flagPublish {
			glog.Exitf("Publishing images is not supported on gce (pass --publish=false)")
		}

		gceConfig, gceCloud, err := initGCE()
		if err != nil {
			glog.Exitf("%v", err)
		}
		templateContext = gceConfig
		cloud = gceCloud

	case "":
		glog.Exitf("Cloud not set")
	default:
		glog.Exitf("Unknown cloud: %q", config.Cloud)
	}

	if *flagBuild && config.TemplatePath == "" {
		glog.Fatalf("TemplatePath must be provided")
	}

	var bvzTemplate *imagebuilder.BootstrapVzTemplate
	var imageName string
	if config.TemplatePath != "" {
		templateResolved := path.Join(path.Dir(*flagConfig), config.TemplatePath)

		templateRaw, err := imagebuilder.ReadFile(templateResolved)
		if err != nil {
			glog.Fatalf("error reading template: %v", err)
		}

		templateString, err := imagebuilder.ExpandTemplate(templateResolved, string(templateRaw), templateContext)
		if err != nil {
			glog.Fatalf("error executing template: %v", err)
		}

		bvzTemplate, err = imagebuilder.NewBootstrapVzTemplate(templateString)
		if err != nil {
			glog.Fatalf("error parsing template: %v", err)
		}

		imageName, err = bvzTemplate.BuildImageName()
		if err != nil {
			glog.Fatalf("error inferring image name: %v", err)
		}

		glog.Infof("Parsed template %q; will build image with name %s", config.TemplatePath, imageName)
	}

	instance, err := cloud.GetInstance()
	if err != nil {
		glog.Fatalf("error getting instance: %v", err)
	}

	if instance == nil && *flagUp {
		instance, err = cloud.CreateInstance()
		if err != nil {
			glog.Fatalf("error creating instance: %v", err)
		}
	}

	image, err := cloud.FindImage(imageName)
	if err != nil {
		glog.Fatalf("error finding image %q: %v", imageName, err)
	}

	if image != nil {
		glog.Infof("found existing image %q", image)
	}

	if *flagBuild && image == nil {
		if instance == nil {
			glog.Fatalf("Instance was not found (specify --up?)")
		}

		sshConfig := &ssh.ClientConfig{
			User: config.SSHUsername,
		}

		if !*flagLocalhost {
			if config.SSHPrivateKey == "" {
				glog.Fatalf("SSHPublicKey is required")
				// We used to allow the SSH agent, but probably more trouble than it is worth?
				//sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
				//if err != nil {
				//	glog.Fatalf("error connecting to SSH agent: %v", err)
				//}
				//
				//sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers))
			} else {
				keyBytes, err := imagebuilder.ReadFile(config.SSHPrivateKey)
				if err != nil {
					glog.Exitf("error loading SSH private key: %v", err)
				}
				key, err := ssh.ParsePrivateKey(keyBytes)
				if err != nil {
					glog.Exitf("error parsing SSH private key: %v", err)
				}

				sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(key))
			}
		}
		x, err := instance.DialSSH(sshConfig)
		if err != nil {
			glog.Fatalf("error SSHing to instance: %v", err)
		}
		defer x.Close()

		sshHelper := executor.NewTarget(x)

		builder := imagebuilder.NewBuilder(config, sshHelper)
		err = builder.RunSetupCommands()
		if err != nil {
			glog.Fatalf("error setting up instance: %v", err)
		}

		extraEnv, err := cloud.GetExtraEnv()
		if err != nil {
			glog.Fatalf("error building environment: %v", err)
		}

		err = builder.BuildImage(bvzTemplate.Bytes(), extraEnv)
		if err != nil {
			glog.Fatalf("error building image: %v", err)
		}

		image, err = cloud.FindImage(imageName)
		if err != nil {
			glog.Fatalf("error finding image %q: %v", imageName, err)
		}

		if image == nil {
			glog.Fatalf("image not found after build: %q", imageName)
		}
	}

	if *flagTag {
		if image == nil {
			glog.Fatalf("image not found: %q", imageName)
		}

		glog.Infof("Tagging image %q", image)

		tags := make(map[string]string)
		for k, v := range config.Tags {
			tags[k] = v
		}

		{
			t := time.Now().UTC().Format("20060102150405")
			tags["k8s.io/build"] = t
		}

		err = image.AddTags(tags)
		if err != nil {
			glog.Fatalf("error tagging image %q: %v", imageName, err)
		}

		glog.Infof("Tagged image %q", image)
	}

	if *flagPublish {
		if image == nil {
			glog.Fatalf("image not found: %q", imageName)
		}

		glog.Infof("Making image public: %v", image)

		err = image.EnsurePublic()
		if err != nil {
			glog.Fatalf("error making image public %q: %v", imageName, err)
		}

		glog.Infof("Made image public: %v", image)
	}

	if *flagReplicate {
		if image == nil {
			glog.Fatalf("image not found: %q", imageName)
		}

		glog.Infof("Copying image to all regions: %v", image)

		images, err := image.ReplicateImage(*flagPublish)
		if err != nil {
			glog.Fatalf("error replicating image %q: %v", imageName, err)
		}

		for region, imageID := range images {
			glog.Infof("Image in region %q: %q", region, imageID)
		}
	}

	if *flagDown {
		if instance == nil {
			glog.Infof("Instance not found / already shutdown")
		} else {
			err := instance.Shutdown()
			if err != nil {
				glog.Fatalf("error terminating instance: %v", err)
			}
		}
	}
}

func initAWS(useLocalhost bool) (*imagebuilder.AWSConfig, *imagebuilder.AWSCloud, error) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}
	awsConfig := &imagebuilder.AWSConfig{}
	awsConfig.InitDefaults(region)
	err := loadConfig(awsConfig, *flagConfig)
	if err != nil {
		glog.Exitf("Error loading AWS config: %v", err)
	}

	if awsConfig.Region == "" {
		glog.Exitf("Region must be set")
	}

	ec2Client := ec2.New(session.New(), &aws.Config{Region: &awsConfig.Region})
	awsCloud := imagebuilder.NewAWSCloud(ec2Client, awsConfig, useLocalhost)

	return awsConfig, awsCloud, nil
}

func initGCE() (*imagebuilder.GCEConfig, *imagebuilder.GCECloud, error) {
	config := &imagebuilder.GCEConfig{}
	config.InitDefaults()
	err := loadConfig(config, *flagConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("Error loading GCE config: %v", err)
	}

	if config.Project == "" {
		return nil, nil, fmt.Errorf("Project must be set")
	}

	if config.MachineName == "" {
		return nil, nil, fmt.Errorf("Name must be set")
	}
	if config.Zone == "" {
		return nil, nil, fmt.Errorf("Zone must be set")
	}
	if config.MachineType == "" {
		return nil, nil, fmt.Errorf("MachineType must be set")
	}

	if config.Image == "" {
		return nil, nil, fmt.Errorf("Image must be set")
	}

	if config.GCSDestination == "" {
		return nil, nil, fmt.Errorf("GCSDestination must be set")
	}

	// Avoid common mistake...
	if !strings.HasSuffix(config.GCSDestination, "/") {
		return nil, nil, fmt.Errorf("GCSDestination should end in /")
	}
	if !strings.HasPrefix(config.GCSDestination, "gs://") {
		return nil, nil, fmt.Errorf("GCSDestination should start with gs://")
	}

	ctx := context.Background()

	client, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		return nil, nil, fmt.Errorf("error building google API client: %v", err)
	}
	computeService, err := compute.New(client)
	if err != nil {
		return nil, nil, fmt.Errorf("error building compute API client: %v", err)
	}

	storageService, err := storage.New(client)
	if err != nil {
		return nil, nil, fmt.Errorf("error building compute API client: %v", err)
	}
	u, err := url.Parse(config.GCSDestination)
	if err != nil {
		return nil, nil, fmt.Errorf("GCSDestination %q is not a well-formed URL: %v", config.GCSDestination, err)
	}
	glog.Infof("Checking for bucket %q", u.Host)
	_, err = storageService.Buckets.Get(u.Host).Do()
	if err != nil {
		if imagebuilder.IsGCENotFound(err) {
			return nil, nil, fmt.Errorf("GCS bucket does not exist: %v", config.GCSDestination)
		}
		return nil, nil, fmt.Errorf("Error checking that bucket exists: %v", err)
	}

	cloud := imagebuilder.NewGCECloud(computeService, config)

	return config, cloud, nil
}
