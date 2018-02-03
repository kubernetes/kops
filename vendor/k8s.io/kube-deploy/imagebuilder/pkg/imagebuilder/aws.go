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

// TODO: We should replace most of this code with a fast-install manifest
// This would also allow more customization, and get rid of half of this code
// BUT... there's a circular dependency in the PRs here... :-)

package imagebuilder

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"golang.org/x/crypto/ssh"
	"k8s.io/kube-deploy/imagebuilder/pkg/imagebuilder/executor"
)

const tagRoleKey = "k8s.io/role/imagebuilder"

// AWSInstance manages an AWS instance, used for building an image
type AWSInstance struct {
	instanceID string
	cloud      *AWSCloud
	instance   *ec2.Instance
}

var _ Instance = &AWSInstance{}

// Shutdown terminates the running instance
func (i *AWSInstance) Shutdown() error {
	glog.Infof("Terminating instance %q", i.instanceID)
	return i.cloud.TerminateInstance(i.instanceID)
}

// DialSSH establishes an SSH client connection to the instance
func (i *AWSInstance) DialSSH(config *ssh.ClientConfig) (executor.Executor, error) {
	publicIP, err := i.WaitPublicIP()
	if err != nil {
		return nil, err
	}

	for {
		// TODO: Timeout, check error code
		sshClient, err := ssh.Dial("tcp", publicIP+":22", config)
		if err != nil {
			glog.Warningf("error connecting to SSH on server %q: %v", publicIP, err)
			time.Sleep(5 * time.Second)
			continue
			//	return nil, fmt.Errorf("error connecting to SSH on server %q", publicIP)
		}

		return executor.NewSSH(sshClient), nil
	}
}

// WaitPublicIP waits for the instance to get a public IP, returning it
func (i *AWSInstance) WaitPublicIP() (string, error) {
	// TODO: Timeout
	for {
		instance, err := i.cloud.describeInstance(i.instanceID)
		if err != nil {
			return "", err
		}
		publicIP := aws.StringValue(instance.PublicIpAddress)
		if publicIP != "" {
			glog.Infof("Instance public IP is %q", publicIP)
			return publicIP, nil
		}
		glog.V(2).Infof("Sleeping before requerying instance for public IP: %q", i.instanceID)
		time.Sleep(5 * time.Second)
	}
}

type LocalhostInstance struct {
	cloud Cloud
}

// Shutdown terminates the running instance
func (i *LocalhostInstance) Shutdown() error {
	glog.Infof("Skipping termination of localhost")
	return nil
}

// DialSSH establishes an SSH client connection to the instance
func (i *LocalhostInstance) DialSSH(config *ssh.ClientConfig) (executor.Executor, error) {
	return &executor.LocalhostExecutor{}, nil
}

// AWSCloud is a helper type for talking to an AWS acccount
type AWSCloud struct {
	config *AWSConfig

	ec2 *ec2.EC2

	useLocalhost bool
}

var _ Cloud = &AWSCloud{}

func NewAWSCloud(ec2 *ec2.EC2, config *AWSConfig, useLocalhost bool) *AWSCloud {
	return &AWSCloud{
		ec2:          ec2,
		config:       config,
		useLocalhost: useLocalhost,
	}
}

func (a *AWSCloud) GetExtraEnv() (map[string]string, error) {
	env := make(map[string]string)

	if a.useLocalhost {
		return env, nil
	}

	credentials := a.ec2.Config.Credentials
	if credentials == nil {
		return nil, fmt.Errorf("unable to determine EC2 credentials")
	}

	creds, err := credentials.Get()
	if err != nil {
		return nil, fmt.Errorf("error fetching EC2 credentials: %v", err)
	}

	env["AWS_ACCESS_KEY_ID"] = creds.AccessKeyID
	env["AWS_SECRET_ACCESS_KEY"] = creds.SecretAccessKey
	env["AWS_SESSION_TOKEN"] = creds.SessionToken

	return env, nil
}

func (a *AWSCloud) describeInstance(instanceID string) (*ec2.Instance, error) {
	request := &ec2.DescribeInstancesInput{}
	request.InstanceIds = []*string{&instanceID}

	glog.V(2).Infof("AWS DescribeInstances InstanceId=%q", instanceID)
	response, err := a.ec2.DescribeInstances(request)
	if err != nil {
		return nil, fmt.Errorf("error making AWS DescribeInstances call: %v", err)
	}

	for _, reservation := range response.Reservations {
		for _, instance := range reservation.Instances {
			if aws.StringValue(instance.InstanceId) != instanceID {
				panic("Unexpected InstanceId found")
			}

			return instance, err
		}
	}
	return nil, nil
}

// TerminateInstance terminates the specified instance
func (a *AWSCloud) TerminateInstance(instanceID string) error {
	if a.useLocalhost {
		glog.Infof("Skipping termination as locahost")
		return nil
	}

	request := &ec2.TerminateInstancesInput{}
	request.InstanceIds = []*string{&instanceID}

	glog.V(2).Infof("AWS TerminateInstances instanceID=%q", instanceID)
	_, err := a.ec2.TerminateInstances(request)
	return err
}

// GetInstance returns the AWS instance matching our tags, or nil if not found
func (a *AWSCloud) GetInstance() (Instance, error) {
	if a.useLocalhost {
		return &LocalhostInstance{}, nil
	}

	request := &ec2.DescribeInstancesInput{}
	request.Filters = []*ec2.Filter{
		{
			Name:   aws.String("tag-key"),
			Values: aws.StringSlice([]string{tagRoleKey}),
		},
	}

	glog.V(2).Infof("AWS DescribeInstances Filter:tag-key=%s", tagRoleKey)
	response, err := a.ec2.DescribeInstances(request)
	if err != nil {
		return nil, fmt.Errorf("error making AWS DescribeInstances call: %v", err)
	}

	for _, reservation := range response.Reservations {
		for _, instance := range reservation.Instances {
			instanceID := aws.StringValue(instance.InstanceId)
			if instanceID == "" {
				panic("Found instance with empty instance ID")
			}

			if instance.State == nil {
				glog.Warningf("Ignoring instance with nil state: %q", instanceID)
			}

			state := aws.StringValue(instance.State.Name)
			switch state {
			case ec2.InstanceStateNameShuttingDown, ec2.InstanceStateNameTerminated, ec2.InstanceStateNameStopping, ec2.InstanceStateNameStopped:
				glog.Infof("Ignoring instance %q in state %q", instanceID, state)
				continue

			case ec2.InstanceStateNamePending, ec2.InstanceStateNameRunning:
				glog.V(2).Infof("Instance %q is in state %q", instanceID, state)

			default:
				glog.Warningf("Found instance %q in unknown state %q", instanceID, state)
			}

			glog.Infof("Found existing instance: %q", instanceID)
			return &AWSInstance{
				cloud:      a,
				instance:   instance,
				instanceID: instanceID,
			}, nil
		}
	}

	return nil, nil
}

// findSubnet returns a subnet tagged with our role tag, if one exists
func (c *AWSCloud) findSubnet() (*ec2.Subnet, error) {
	request := &ec2.DescribeSubnetsInput{}
	request.Filters = []*ec2.Filter{
		{
			Name:   aws.String("tag-key"),
			Values: aws.StringSlice([]string{tagRoleKey}),
		},
	}

	glog.V(2).Infof("AWS DescribeSubnets Filter:tag-key=%s", tagRoleKey)
	response, err := c.ec2.DescribeSubnets(request)
	if err != nil {
		return nil, fmt.Errorf("error making AWS DescribeSubnets call: %v", err)
	}

	for _, subnet := range response.Subnets {
		return subnet, nil
	}

	return nil, nil
}

// findSecurityGroup returns a security group tagged with our role tag, if one exists
func (c *AWSCloud) findSecurityGroup(vpcID string) (*ec2.SecurityGroup, error) {
	request := &ec2.DescribeSecurityGroupsInput{}
	request.Filters = []*ec2.Filter{
		{
			Name:   aws.String("tag-key"),
			Values: aws.StringSlice([]string{tagRoleKey}),
		},
		{
			Name:   aws.String("vpc-id"),
			Values: aws.StringSlice([]string{vpcID}),
		},
	}

	glog.V(2).Infof("AWS DescribeSecurityGroups Filter:tag-key=%s", tagRoleKey)
	response, err := c.ec2.DescribeSecurityGroups(request)
	if err != nil {
		return nil, fmt.Errorf("error making AWS DescribeSecurityGroups call: %v", err)
	}

	for _, sg := range response.SecurityGroups {
		return sg, nil
	}

	return nil, nil
}

// describeSubnet returns a subnet with the specified id, if it exists
func (c *AWSCloud) describeSubnet(subnetID string) (*ec2.Subnet, error) {
	request := &ec2.DescribeSubnetsInput{}
	request.SubnetIds = []*string{&subnetID}

	glog.V(2).Infof("AWS DescribeSubnetsInput ID:%q", subnetID)
	response, err := c.ec2.DescribeSubnets(request)
	if err != nil {
		return nil, fmt.Errorf("error making AWS DescribeSubnets call: %v", err)
	}

	for _, subnet := range response.Subnets {
		return subnet, nil
	}

	return nil, nil
}

// TagResource adds AWS tags to the specified resource
func (a *AWSCloud) TagResource(resourceId string, tags ...*ec2.Tag) error {
	request := &ec2.CreateTagsInput{}
	request.Resources = aws.StringSlice([]string{resourceId})
	request.Tags = tags

	glog.V(2).Infof("AWS CreateTags Resource=%q", resourceId)
	_, err := a.ec2.CreateTags(request)
	if err != nil {
		return fmt.Errorf("error making AWS CreateTag call: %v", err)
	}

	return err
}

func (c *AWSCloud) findSSHKey(name string) (*ec2.KeyPairInfo, error) {
	request := &ec2.DescribeKeyPairsInput{
		KeyNames: []*string{&name},
	}

	response, err := c.ec2.DescribeKeyPairs(request)
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == "InvalidKeyPair.NotFound" {
			return nil, nil
		}
	}
	if err != nil {
		return nil, fmt.Errorf("error listing AWS KeyPairs: %v", err)
	}

	if response == nil || len(response.KeyPairs) == 0 {
		return nil, nil
	}

	if len(response.KeyPairs) != 1 {
		return nil, fmt.Errorf("Found multiple AWS KeyPairs with Name %q", name)
	}

	k := response.KeyPairs[0]

	return k, nil
}

func (c *AWSCloud) ensureSSHKey() (string, error) {
	publicKey, err := ReadFile(c.config.SSHPublicKey)
	if err != nil {
		return "", err
	}

	// TODO: Use real OpenSSH or AWS fingerprint?
	hashBytes := md5.Sum([]byte(publicKey))
	hash := hex.EncodeToString(hashBytes[:])

	name := "imagebuilder-" + hash

	key, err := c.findSSHKey(name)
	if err != nil {
		return "", err
	}

	if key != nil {
		return *key.KeyName, nil
	}

	glog.V(2).Infof("Creating AWS KeyPair with Name:%q", name)

	request := &ec2.ImportKeyPairInput{}
	request.KeyName = &name
	request.PublicKeyMaterial = []byte(publicKey)

	response, err := c.ec2.ImportKeyPair(request)
	if err != nil {
		return "", fmt.Errorf("error creating AWS KeyPair: %v", err)
	}

	return *response.KeyName, nil
}

// CreateInstance creates an instance for building an image instance
func (c *AWSCloud) CreateInstance() (Instance, error) {
	if c.useLocalhost {
		return &LocalhostInstance{cloud: c}, nil
	}

	var err error
	sshKeyName := c.config.SSHKeyName
	if sshKeyName == "" {
		sshKeyName, err = c.ensureSSHKey()
		if err != nil {
			return nil, err
		}
	}

	subnetID := c.config.SubnetID
	if subnetID == "" {
		subnet, err := c.findSubnet()
		if err != nil {
			return nil, err
		}
		if subnet != nil {
			subnetID = aws.StringValue(subnet.SubnetId)
		}
		if subnetID == "" {
			return nil, fmt.Errorf("SubnetID must be specified, or a subnet must be tagged with %q", tagRoleKey)
		}
	}

	subnet, err := c.describeSubnet(subnetID)
	if err != nil {
		return nil, err
	}
	if subnet == nil {
		return nil, fmt.Errorf("could not find subnet %q", subnetID)
	}

	if c.config.ImageID == "" {
		return nil, fmt.Errorf("ImageID must be specified")
	}

	if c.config.InstanceType == "" {
		return nil, fmt.Errorf("InstanceType must be specified")
	}

	securityGroupID := c.config.SecurityGroupID
	if securityGroupID == "" {
		vpcID := *subnet.VpcId
		securityGroup, err := c.findSecurityGroup(vpcID)
		if err != nil {
			return nil, err
		}
		if securityGroup != nil {
			securityGroupID = aws.StringValue(securityGroup.GroupId)
		}
		if securityGroupID == "" {
			return nil, fmt.Errorf("SecurityGroupID must be specified, or a security group for VPC %q must be tagged with %q", vpcID, tagRoleKey)
		}
	}

	request := &ec2.RunInstancesInput{}
	request.ImageId = aws.String(c.config.ImageID)
	request.KeyName = aws.String(sshKeyName)
	request.InstanceType = aws.String(c.config.InstanceType)
	request.NetworkInterfaces = []*ec2.InstanceNetworkInterfaceSpecification{
		{
			DeviceIndex:              aws.Int64(0),
			AssociatePublicIpAddress: aws.Bool(true),
			SubnetId:                 aws.String(subnetID),
			Groups:                   aws.StringSlice([]string{securityGroupID}),
		},
	}
	request.MaxCount = aws.Int64(1)
	request.MinCount = aws.Int64(1)

	glog.V(2).Infof("AWS RunInstances InstanceType=%q ImageId=%q KeyName=%q", c.config.InstanceType, c.config.ImageID, sshKeyName)
	response, err := c.ec2.RunInstances(request)
	if err != nil {
		return nil, fmt.Errorf("error making AWS RunInstances call: %v", err)
	}

	for _, instance := range response.Instances {
		instanceID := aws.StringValue(instance.InstanceId)
		if instanceID == "" {
			return nil, fmt.Errorf("AWS RunInstances call returned empty InstanceId")
		}
		err := c.TagResource(instanceID, &ec2.Tag{
			Key: aws.String(tagRoleKey), Value: aws.String("'"),
		})
		if err != nil {
			glog.Warningf("Tagging instance %q failed; will terminate to prevent leaking", instanceID)
			e2 := c.TerminateInstance(instanceID)
			if e2 != nil {
				glog.Warningf("error terminating instance %q, will leak instance", instanceID)
			}
			return nil, err
		}

		return &AWSInstance{
			cloud:      c,
			instance:   instance,
			instanceID: instanceID,
		}, nil
	}
	return nil, fmt.Errorf("instance was not returned by AWS RunInstances")
}

// FindImage finds a registered image, matching by the name tag
func (a *AWSCloud) FindImage(imageName string) (Image, error) {
	image, err := findAWSImage(a.ec2, imageName)
	if err != nil {
		return nil, err
	}

	if image == nil {
		return nil, nil
	}

	imageID := aws.StringValue(image.ImageId)
	if imageID == "" {
		return nil, fmt.Errorf("found image with empty ImageId: %q", imageName)
	}

	if len(image.BlockDeviceMappings) == 0 {
		return nil, fmt.Errorf("found no matching snapshots for image: %q", imageName)
	}

	if len(image.BlockDeviceMappings) != 1 {
		// Image names are unique per user...
		return nil, fmt.Errorf("found multiple matching snapshots for image: %q", imageName)
	}

	return &AWSImage{
		ec2:     a.ec2,
		region:  a.config.Region,
		imageID: imageID,

		cachedImage: image,
	}, nil
}

func findAWSImage(client *ec2.EC2, imageName string) (*ec2.Image, error) {
	request := &ec2.DescribeImagesInput{}
	request.Filters = []*ec2.Filter{
		{
			Name:   aws.String("name"),
			Values: aws.StringSlice([]string{imageName}),
		},
	}
	request.Owners = aws.StringSlice([]string{"self"})

	glog.V(2).Infof("AWS DescribeImages Filter:Name=%q, Owner=self", imageName)
	response, err := client.DescribeImages(request)
	if err != nil {
		return nil, fmt.Errorf("error making AWS DescribeImages call: %v", err)
	}

	if len(response.Images) == 0 {
		return nil, nil
	}

	if len(response.Images) != 1 {
		// Image names are unique per user...
		return nil, fmt.Errorf("found multiple matching images for name: %q", imageName)
	}

	image := response.Images[0]
	return image, nil
}

// AWSImage represents an AMI on AWS
type AWSImage struct {
	ec2     *ec2.EC2
	region  string
	imageID string

	cachedImageMutex sync.Mutex
	cachedImage      *ec2.Image
}

// ID returns the AWS identifier for the image
func (i *AWSImage) ID() string {
	return i.imageID
}

// String returns a string representation of the image
func (i *AWSImage) String() string {
	return "AWSImage[id=" + i.imageID + "]"
}

// EnsurePublic makes the image accessible outside the current account
func (i *AWSImage) EnsurePublic() error {
	return i.ensurePublic()
}

// AddTags adds the specified tags on the image
func (i *AWSImage) AddTags(tags map[string]string) error {
	request := &ec2.CreateTagsInput{}
	request.Resources = append(request.Resources, aws.String(i.imageID))
	for k, v := range tags {
		request.Tags = append(request.Tags, &ec2.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}

	glog.V(2).Infof("AWS CreateTags on image %v", i.imageID)
	_, err := i.ec2.CreateTags(request)
	if err != nil {
		return fmt.Errorf("error tagging image %q: %v", i.imageID, err)
	}

	return err
}

func (i *AWSImage) waitStatusAvailable() error {
	imageID := i.imageID

	for {
		// TODO: Timeout
		request := &ec2.DescribeImagesInput{}
		request.ImageIds = aws.StringSlice([]string{imageID})

		glog.V(2).Infof("AWS DescribeImages ImageId=%q", imageID)
		response, err := i.ec2.DescribeImages(request)
		if err != nil {
			return fmt.Errorf("error making AWS DescribeImages call: %v", err)
		}

		if len(response.Images) == 0 {
			return fmt.Errorf("image not found %q", imageID)
		}

		if len(response.Images) != 1 {
			return fmt.Errorf("multiple images found with ID %q", imageID)
		}

		image := response.Images[0]

		state := aws.StringValue(image.State)
		glog.V(2).Infof("image state %q", state)
		if state == "available" {
			return nil
		}
		glog.Infof("Image %q not yet available (%s); waiting", imageID, state)
		time.Sleep(10 * time.Second)
	}
}

func waitSnapshotCompleted(client *ec2.EC2, snapshotID string) error {
	for {
		// TODO: Timeout
		request := &ec2.DescribeSnapshotsInput{}
		request.SnapshotIds = aws.StringSlice([]string{snapshotID})

		glog.V(2).Infof("AWS DescribeSnapshots SnapshotId=%q", snapshotID)
		response, err := client.DescribeSnapshots(request)
		if err != nil {
			return fmt.Errorf("error making AWS DescribeSnapshots call: %v", err)
		}

		if len(response.Snapshots) == 0 {
			return fmt.Errorf("snapshot not found %q", snapshotID)
		}

		if len(response.Snapshots) != 1 {
			return fmt.Errorf("multiple snapshots found with ID %q", snapshotID)
		}

		snapshot := response.Snapshots[0]

		state := aws.StringValue(snapshot.State)
		glog.V(2).Infof("snapshot state %q", state)
		if state == "completed" {
			return nil
		}
		glog.Infof("Snapshot %q not yet completed (%s); waiting", snapshotID, state)
		time.Sleep(10 * time.Second)
	}
}

func (i *AWSImage) image() (*ec2.Image, error) {
	i.cachedImageMutex.Lock()
	defer i.cachedImageMutex.Unlock()

	if i.cachedImage != nil {
		return i.cachedImage, nil
	}

	if i.imageID == "" {
		return nil, fmt.Errorf("imageID not set")
	}
	request := &ec2.DescribeImagesInput{}
	request.ImageIds = aws.StringSlice([]string{i.imageID})

	glog.V(2).Infof("AWS DescribeImages id=%q", i.imageID)
	response, err := i.ec2.DescribeImages(request)
	if err != nil {
		return nil, fmt.Errorf("error making AWS DescribeImages call: %v", err)
	}

	if len(response.Images) == 0 {
		return nil, nil
	}

	if len(response.Images) != 1 {
		// Image names are unique per user...
		return nil, fmt.Errorf("found multiple matching images for id: %q", i.imageID)
	}

	image := response.Images[0]
	i.cachedImage = image
	return image, nil
}

func (i *AWSImage) imageSnapshotId() (string, error) {
	image, err := i.image()
	if err != nil {
		return "", err
	}
	if image == nil {
		return "", fmt.Errorf("image not set")
	}
	if image.BlockDeviceMappings == nil {
		glog.Warningf("image did not have BlockDeviceMappings: %v", image)
		return "", nil
	}
	if len(image.BlockDeviceMappings) == 0 {
		glog.Warningf("image did not have BlockDeviceMappings: %v", image)
		return "", nil
	}
	if image.BlockDeviceMappings[0].Ebs == nil {
		glog.Warningf("image had nil BlockDeviceMappings[0].Ebs: %v", image)
		return "", nil
	}
	snapshotID := aws.StringValue(image.BlockDeviceMappings[0].Ebs.SnapshotId)
	if snapshotID == "" {
		glog.Warningf("image did not have snapshot: %v", image)
	}
	return snapshotID, nil
}

func (i *AWSImage) ensurePublic() error {
	err := i.waitStatusAvailable()
	if err != nil {
		return err
	}

	image, err := i.image()
	if err != nil {
		return err
	}

	// This is idempotent, so just always do it
	request := &ec2.ModifyImageAttributeInput{}
	request.ImageId = aws.String(i.imageID)
	request.LaunchPermission = &ec2.LaunchPermissionModifications{
		Add: []*ec2.LaunchPermission{
			{Group: aws.String("all")},
		},
	}

	glog.V(2).Infof("AWS ModifyImageAttribute Image=%q, LaunchPermission All", image)
	_, err = i.ec2.ModifyImageAttribute(request)
	if err != nil {
		return fmt.Errorf("error making image public (%q in region %q): %v", i.imageID, i.region, err)
	}

	snapshotID, err := i.imageSnapshotId()
	if err != nil {
		return err
	}
	if snapshotID == "" {
		return fmt.Errorf("found image with empty SnapshotId: %q", aws.StringValue(image.ImageId))
	}

	request2 := &ec2.ModifySnapshotAttributeInput{
		Attribute: aws.String("createVolumePermission"),
		GroupNames: []*string{
			aws.String("all"),
		},
		OperationType: aws.String("add"),
		SnapshotId:    aws.String(snapshotID),
	}

	glog.V(2).Infof("AWS ModifySnapshotAttribute Snapshot=%q, CreateVolumePermission All", snapshotID)
	_, err = i.ec2.ModifySnapshotAttribute(request2)
	if err != nil {
		return fmt.Errorf("error making snapshot public (%q in region %q): %v", snapshotID, i.region, err)
	}

	return err
}

// ReplicateImage copies the image to all accessable AWS regions
func (i *AWSImage) ReplicateImage(makePublic bool) (map[string]Image, error) {
	var results sync.Map

	glog.V(2).Infof("AWS DescribeRegions")
	request := &ec2.DescribeRegionsInput{}
	response, err := i.ec2.DescribeRegions(request)
	if err != nil {
		return nil, fmt.Errorf("error listing ec2 regions: %v", err)

	}
	results.Store(i.region, i)

	var wg sync.WaitGroup

	for _, region := range response.Regions {
		go func(regionName string) {
			var image *AWSImage
			if v, ok := results.Load(regionName); ok {
				image = v.(*AWSImage)
			}

			if image == nil {
				imageID, err := i.copyImageToRegion(regionName)
				if err != nil {
					results.Store(regionName, fmt.Errorf("error copying image to region %q: %v", regionName, err))
					wg.Done()
					return
				}
				targetEC2 := ec2.New(session.New(), &aws.Config{Region: &regionName})
				image = &AWSImage{
					ec2:     targetEC2,
					region:  regionName,
					imageID: imageID,
				}

				results.Store(regionName, image)
			}

			if makePublic {
				err := image.EnsurePublic()
				if err != nil {
					results.Store(regionName, fmt.Errorf("error making image public in region %q: %v", regionName, err))
					wg.Done()
					return
				}
			}

			wg.Done()
		}(aws.StringValue(region.RegionName))
		wg.Add(1)
	}

	wg.Wait()

	imagesByRegion := make(map[string]Image)
	var returnError error
	results.Range(func(k, v interface{}) bool {
		if err, ok := v.(error); ok {
			returnError = err
			return false
		}
		imagesByRegion[k.(string)] = v.(*AWSImage)
		return true
	})
	if returnError != nil {
		return nil, returnError
	}
	return imagesByRegion, nil
}

func (i *AWSImage) copyImageToRegion(regionName string) (string, error) {
	targetEC2 := ec2.New(session.New(), &aws.Config{Region: &regionName})

	image, err := i.image()
	if err != nil {
		return "", err
	}

	imageName := aws.StringValue(image.Name)
	description := aws.StringValue(image.Description)

	destImage, err := findAWSImage(targetEC2, imageName)
	if err != nil {
		return "", err
	}

	var imageID string

	// We've already copied the image
	if destImage != nil {
		imageID = aws.StringValue(destImage.ImageId)
	} else {
		var snapshotId string

		{
			sourceSnapshotID, err := i.imageSnapshotId()
			if err != nil {
				return "", err
			}
			if sourceSnapshotID == "" {
				return "", fmt.Errorf("found image with empty SnapshotId: %q", imageName)
			}

			request := &ec2.CopySnapshotInput{
				Description:       aws.String(description),
				SourceSnapshotId:  aws.String(sourceSnapshotID),
				SourceRegion:      aws.String(i.region),
				DestinationRegion: aws.String(regionName),
			}
			glog.V(2).Infof("AWS CopySnapshot SnapshotId=%q, Region=%q", sourceSnapshotID, regionName)
			response, err := targetEC2.CopySnapshot(request)
			if err != nil {
				return "", fmt.Errorf("error copying snapshot to region %q: %v", regionName, err)
			}
			snapshotId = aws.StringValue(response.SnapshotId)
		}

		if err := waitSnapshotCompleted(targetEC2, snapshotId); err != nil {
			return "", err
		}

		{
			request := &ec2.RegisterImageInput{
				Architecture: image.Architecture,
				//BlockDeviceMappings: image.BlockDeviceMappings,
				Description:        aws.String(description),
				EnaSupport:         image.EnaSupport,
				KernelId:           image.KernelId,
				Name:               aws.String(imageName),
				RamdiskId:          image.RamdiskId,
				RootDeviceName:     image.RootDeviceName,
				SriovNetSupport:    image.SriovNetSupport,
				VirtualizationType: image.VirtualizationType,
			}
			found := false
			for _, bdm := range image.BlockDeviceMappings {
				copy := *bdm
				if copy.Ebs != nil {
					if aws.StringValue(copy.Ebs.SnapshotId) != "" {
						found = true
						copy.Ebs.SnapshotId = aws.String(snapshotId)
						copy.Ebs.Encrypted = nil
					}
				}
				request.BlockDeviceMappings = append(request.BlockDeviceMappings, &copy)
			}
			if !found {
				return "", fmt.Errorf("unable to remap block device mappings for image %q", aws.StringValue(image.ImageId))
			}

			glog.V(2).Infof("AWS RegisterImage SnapshotId=%q, Region=%q", snapshotId, regionName)
			response, err := targetEC2.RegisterImage(request)
			if err != nil {
				return "", fmt.Errorf("error copying image to region %q: %v", regionName, err)
			}

			imageID = aws.StringValue(response.ImageId)
		}
	}

	return imageID, nil
}
