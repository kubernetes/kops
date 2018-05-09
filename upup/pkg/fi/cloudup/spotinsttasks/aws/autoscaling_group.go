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

package aws

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/spotinst/spotinst-sdk-go/service/elastigroup/providers/aws"
	spotinstsdk "github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/stringutil"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/spotinst"
)

//go:generate fitask -type=AutoscalingGroup
type AutoscalingGroup struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	ID                     *string
	MinSize                *int64
	MaxSize                *int64
	Risk                   *float64
	Product                *string
	Orientation            *string
	Tags                   map[string]string
	UserData               *fi.ResourceHolder
	ImageID                *string
	OnDemandInstanceType   *string
	SpotInstanceTypes      []string
	Subnets                []*Subnet
	SSHKey                 *SSHKey
	SecurityGroups         []*SecurityGroup
	IAMInstanceProfile     *IAMInstanceProfile
	AssociatePublicIP      *bool
	RootVolumeSize         *int64
	RootVolumeType         *string
	RootVolumeIOPS         *int64
	RootVolumeOptimization *bool
	Tenancy                *string

	IntegrationAutoScaleEnabled    *bool
	IntegrationClusterIdentifier   *string
	IntegrationAutoScaleNodeLabels map[string]string
}

var _ fi.CompareWithID = &AutoscalingGroup{}

func (e *AutoscalingGroup) CompareWithID() *string {
	return e.Name
}

func findAutoscalingGroup(cloud spotinst.Cloud, name string) (*aws.Group, error) {
	output, err := cloud.Service().CloudProviderAWS().List(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("spotinst: failed to find group %s: %v", name, err)
	}
	if output == nil || (output != nil && len(output.Groups) == 0) {
		return nil, fmt.Errorf("spotinst: failed to find group %s: got an empty response", name)
	}

	var out *aws.Group
	for _, group := range output.Groups {
		if spotinstsdk.StringValue(group.Name) == name {
			out = group
			break
		}
	}
	if out == nil {
		return nil, fmt.Errorf("spotinst: failed to find group %s: group does not exist", name)
	}

	return out, nil
}

var _ fi.HasCheckExisting = &AutoscalingGroup{}

func (e *AutoscalingGroup) Find(c *fi.Context) (*AutoscalingGroup, error) {
	cloud := c.Cloud.(spotinst.Cloud)

	group, err := findAutoscalingGroup(cloud, *e.Name)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, nil
	}

	actual := &AutoscalingGroup{}
	actual.ID = group.ID
	actual.Name = group.Name
	actual.MinSize = spotinstsdk.Int64(int64(spotinstsdk.IntValue(group.Capacity.Minimum)))
	actual.MaxSize = spotinstsdk.Int64(int64(spotinstsdk.IntValue(group.Capacity.Maximum)))
	actual.Orientation = group.Strategy.AvailabilityVsCost

	// Compute.
	{
		compute := group.Compute
		actual.Product = compute.Product

		// Instance types.
		{
			actual.OnDemandInstanceType = compute.InstanceTypes.OnDemand
			actual.SpotInstanceTypes = compute.InstanceTypes.Spot
		}

		// Subnets.
		{
			for _, zone := range compute.AvailabilityZones {
				if zone.SubnetID != nil {
					actual.Subnets = append(actual.Subnets, &Subnet{ID: zone.SubnetID})
				}
			}
			if subnetSlicesEqualIgnoreOrder(actual.Subnets, e.Subnets) {
				actual.Subnets = e.Subnets
			}
		}
	}

	// Launch Specification.
	{
		lc := group.Compute.LaunchSpecification

		// Image.
		{
			image, err := cloud.Cloud().(awsup.AWSCloud).ResolveImage(*lc.ImageID)
			if err != nil {
				glog.Warningf("spotinst: unable to resolve image: %q: %v", *e.ImageID, err)
			} else if image == nil {
				glog.Warningf("spotinst: unable to resolve image: %q: not found", *e.ImageID)
			} else {
				actual.ImageID = image.Name
			}
		}

		// Tags.
		{
			if len(lc.Tags) > 0 {
				actual.Tags = make(map[string]string)
				for _, tag := range lc.Tags {
					actual.Tags[*tag.Key] = *tag.Value
				}
			}
		}

		// Security groups.
		{
			for _, sg := range lc.SecurityGroupIDs {
				actual.SecurityGroups = append(actual.SecurityGroups, &SecurityGroup{ID: spotinstsdk.String(sg)})
			}
		}

		// Block device mappings.
		{
			for _, b := range lc.BlockDeviceMappings {
				if b.EBS == nil || b.EBS.SnapshotID != nil {
					// Not the root.
					continue
				}
				actual.RootVolumeType = b.EBS.VolumeType
				actual.RootVolumeSize = spotinstsdk.Int64(int64(spotinstsdk.IntValue(b.EBS.VolumeSize)))
				actual.RootVolumeIOPS = spotinstsdk.Int64(int64(spotinstsdk.IntValue(b.EBS.IOPS)))
			}
		}

		// User data.
		{
			if lc.UserData != nil {
				userData, err := base64.StdEncoding.DecodeString(*lc.UserData)
				if err != nil {
					return nil, fmt.Errorf("spotinst: error decoding user data: %v", err)
				}
				actual.UserData = fi.WrapResource(fi.NewStringResource(string(userData)))
			}
		}

		// Network interfaces.
		{
			associatePublicIP := false
			if len(lc.NetworkInterfaces) > 0 {
				for _, iface := range lc.NetworkInterfaces {
					if spotinstsdk.BoolValue(iface.AssociatePublicIPAddress) {
						associatePublicIP = true
						break
					}
				}
			}
			actual.AssociatePublicIP = spotinstsdk.Bool(associatePublicIP)
		}

		if lc.IAMInstanceProfile != nil {
			actual.IAMInstanceProfile = &IAMInstanceProfile{Name: lc.IAMInstanceProfile.Name}
		}

		if lc.KeyPair != nil {
			actual.SSHKey = &SSHKey{Name: lc.KeyPair}
		}

		if lc.Tenancy != nil {
			actual.Tenancy = lc.Tenancy
		}
	}

	// Avoid spurious changes
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *AutoscalingGroup) CheckExisting(c *fi.Context) bool {
	cloud := c.Cloud.(spotinst.Cloud)
	group, err := findAutoscalingGroup(cloud, *e.Name)
	return err == nil && group != nil
}

func (e *AutoscalingGroup) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *AutoscalingGroup) CheckChanges(a, e, changes *AutoscalingGroup) error {
	if e.ImageID == nil {
		return fi.RequiredField("ImageID")
	}
	if e.OnDemandInstanceType == nil {
		return fi.RequiredField("OnDemandInstanceType")
	}
	if a != nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	}
	return nil
}

func (_ *AutoscalingGroup) Render(t *spotinst.Target, a, e, changes *AutoscalingGroup) error {
	if a == nil {
		glog.V(2).Infof("Creating Spotinst group: %q", *e.Name)

		group := &aws.Group{
			Capacity: new(aws.Capacity),
			Strategy: new(aws.Strategy),
			Compute: &aws.Compute{
				LaunchSpecification: new(aws.LaunchSpecification),
				InstanceTypes:       new(aws.InstanceTypes),
			},
		}

		// General.
		{
			group.SetName(e.Name)
			group.SetDescription(e.Name)
		}

		// Capacity.
		{
			group.Capacity.SetTarget(spotinstsdk.Int(int(*e.MinSize)))
			group.Capacity.SetMinimum(spotinstsdk.Int(int(*e.MinSize)))
			group.Capacity.SetMaximum(spotinstsdk.Int(int(*e.MaxSize)))
		}

		// Strategy.
		{
			group.Strategy.SetRisk(e.Risk)
			group.Strategy.SetAvailabilityVsCost(spotinstsdk.String(string(normalizeOrientation(e.Orientation))))
			group.Strategy.SetFallbackToOnDemand(spotinstsdk.Bool(true))
			group.Strategy.SetUtilizeReservedInstances(spotinstsdk.Bool(true))
		}

		// Compute.
		{
			group.Compute.SetProduct(spotinstsdk.String(string(normalizeProduct(e.Product))))

			// Instance types.
			{
				group.Compute.InstanceTypes.SetOnDemand(e.OnDemandInstanceType)
				group.Compute.InstanceTypes.SetSpot(e.SpotInstanceTypes)
			}

			// Availability zones.
			{
				zones := make([]*aws.AvailabilityZone, len(e.Subnets))
				for i, subnet := range e.Subnets {
					zone := new(aws.AvailabilityZone)
					zone.SetName(subnet.AvailabilityZone)
					zone.SetSubnetId(subnet.ID)
					zones[i] = zone
				}
				group.Compute.SetAvailabilityZones(zones)
			}

			// Launch Specification.
			{
				group.Compute.LaunchSpecification.SetMonitoring(spotinstsdk.Bool(false))
				group.Compute.LaunchSpecification.SetKeyPair(e.SSHKey.Name)

				if e.Tenancy != nil {
					group.Compute.LaunchSpecification.SetTenancy(e.Tenancy)
				}

				// Block device mappings.
				{
					rootDevices, err := e.buildRootDevice(t.Cloud.(spotinst.Cloud))
					if err != nil {
						return err
					}

					ephemeralDevices, err := e.buildEphemeralDevices(e.OnDemandInstanceType)
					if err != nil {
						return err
					}

					if len(rootDevices) != 0 || len(ephemeralDevices) != 0 {
						var mappings []*aws.BlockDeviceMapping
						for device, bdm := range rootDevices {
							mappings = append(mappings, bdm.ToGroup(device))
						}
						for device, bdm := range ephemeralDevices {
							mappings = append(mappings, bdm.ToGroup(device))
						}
						if len(mappings) > 0 {
							group.Compute.LaunchSpecification.SetBlockDeviceMappings(mappings)
						}
					}
				}

				// Image ID.
				{
					image, err := t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).ResolveImage(*e.ImageID)
					if err != nil {
						return fmt.Errorf("spotinst: unable to resolve image: %q: %v", *e.ImageID, err)
					} else if image == nil {
						return fmt.Errorf("spotinst: unable to resolve image: %q: not found", *e.ImageID)
					}
					group.Compute.LaunchSpecification.SetImageId(image.ImageId)
				}

				// User data.
				{
					if e.UserData != nil {
						userData, err := e.UserData.AsString()
						if err != nil {
							return err
						}
						encoded := base64.StdEncoding.EncodeToString([]byte(userData))
						group.Compute.LaunchSpecification.SetUserData(spotinstsdk.String(encoded))
					}
				}

				// IAM instance profile.
				{
					iprof := new(aws.IAMInstanceProfile)
					iprof.SetName(e.IAMInstanceProfile.GetName())
					group.Compute.LaunchSpecification.SetIAMInstanceProfile(iprof)
				}

				// Security groups.
				{
					securityGroupIDs := make([]string, len(e.SecurityGroups))
					for i, sg := range e.SecurityGroups {
						securityGroupIDs[i] = *sg.ID
					}
					group.Compute.LaunchSpecification.SetSecurityGroupIDs(securityGroupIDs)
				}

				// Public IP.
				{
					if *e.AssociatePublicIP {
						iface := new(aws.NetworkInterface)
						iface.SetDeviceIndex(spotinstsdk.Int(0))
						iface.SetAssociatePublicIPAddress(spotinstsdk.Bool(true))
						iface.SetDeleteOnTermination(spotinstsdk.Bool(true))
						group.Compute.LaunchSpecification.SetNetworkInterfaces(
							[]*aws.NetworkInterface{iface},
						)
					}
				}

				// Tags.
				{
					tags := e.buildTags()
					group.Compute.LaunchSpecification.SetTags(tags)
				}
			}
		}

		// Integration.
		{
			if e.IntegrationClusterIdentifier != nil {
				k8s := new(aws.KubernetesIntegration)
				k8s.SetClusterIdentifier(e.IntegrationClusterIdentifier)
				k8s.SetIntegrationMode(spotinstsdk.String("pod"))

				if e.IntegrationAutoScaleEnabled != nil {
					autoScale := new(aws.AutoScale)
					autoScale.SetIsEnabled(e.IntegrationAutoScaleEnabled)

					labelsMap := e.IntegrationAutoScaleNodeLabels
					if labelsMap != nil && len(labelsMap) > 0 {
						labels := e.buildAutoScaleLabels(labelsMap)
						autoScale.SetLabels(labels)
					}

					k8s.SetAutoScale(autoScale)
				}

				integration := new(aws.Integration)
				integration.SetKubernetes(k8s)

				group.SetIntegration(integration)
			}
		}

		attempt := 0
		maxAttempts := 10
	readyLoop:
		for {
			attempt++
			glog.V(2).Infof("(%d/%d) Attempting to create group: %s, group: %s",
				attempt, maxAttempts, *e.Name, stringutil.Stringify(group))

			// Wait for IAM instance profile to be ready.
			time.Sleep(7 * time.Second) // lucky 7

			input := &aws.CreateGroupInput{Group: group}
			out, err := t.Cloud.(spotinst.Cloud).Service().CloudProviderAWS().Create(context.Background(), input)
			if err == nil {
				e.ID = out.Group.ID
				break
			}

			if errs, ok := err.(client.Errors); ok {
				for _, err := range errs {
					if strings.Contains(err.Message, "Invalid IAM Instance Profile name") {
						if attempt > maxAttempts {
							return fmt.Errorf("IAM instance profile not yet created/propagated (original error: %v)", err)
						}
						glog.V(4).Infof("Got an error indicating that the IAM instance profile %q is not ready: %q", fi.StringValue(e.IAMInstanceProfile.Name), err)
						glog.Infof("Waiting for IAM instance profile %q to be ready", fi.StringValue(e.IAMInstanceProfile.Name))
						goto readyLoop
					}
				}
				return fmt.Errorf("spotinst: failed to create group: %v", err)
			}
		}
	} else {
		group := new(aws.Group)

		glog.V(2).Infof("Resolving group name: %s", *e.Name)
		actual, err := findAutoscalingGroup(t.Cloud.(spotinst.Cloud), *e.Name)
		if err != nil {
			glog.Errorf("Unable to resolve group %q, error: %s", *e.Name, err)
		}
		glog.V(2).Infof("Group name %q resolved: %s", *e.Name, *actual.ID)
		group.SetId(actual.ID)

		// Strategy.
		{
			// Orientation.
			if changes.Orientation != nil {
				if group.Strategy == nil {
					group.Strategy = new(aws.Strategy)
				}
				group.Strategy.SetAvailabilityVsCost(spotinstsdk.String(string(normalizeOrientation(e.Orientation))))
				changes.Orientation = nil
			}
		}

		// Compute.
		{
			// Product.
			if changes.Product != nil {
				if group.Compute == nil {
					group.Compute = new(aws.Compute)
				}
				group.Compute.SetProduct(e.Product)
				changes.Product = nil
			}

			// OnDemand instance type.
			{
				if changes.OnDemandInstanceType != nil {
					if group.Compute == nil {
						group.Compute = new(aws.Compute)
					}
					if group.Compute.InstanceTypes == nil {
						group.Compute.InstanceTypes = new(aws.InstanceTypes)
					}
					group.Compute.InstanceTypes.SetOnDemand(e.OnDemandInstanceType)
					changes.OnDemandInstanceType = nil
				}
			}

			// Spot instance types.
			{
				if changes.SpotInstanceTypes != nil {
					types := make([]string, len(e.SpotInstanceTypes))
					for i, typ := range e.SpotInstanceTypes {
						types[i] = typ
					}
					if group.Compute == nil {
						group.Compute = new(aws.Compute)
					}
					if group.Compute.InstanceTypes == nil {
						group.Compute.InstanceTypes = new(aws.InstanceTypes)
					}
					group.Compute.InstanceTypes.SetSpot(types)
					changes.SpotInstanceTypes = nil
				}
			}

			// Availability zones.
			{
				if changes.Subnets != nil {
					zones := make([]*aws.AvailabilityZone, len(e.Subnets))
					for i, subnet := range e.Subnets {
						zone := new(aws.AvailabilityZone)
						zone.SetName(subnet.AvailabilityZone)
						zone.SetSubnetId(subnet.ID)
						zones[i] = zone
					}
					if group.Compute == nil {
						group.Compute = new(aws.Compute)
					}
					group.Compute.SetAvailabilityZones(zones)
					changes.Subnets = nil
				}
			}

			// Launch specification.
			{
				// Security groups.
				{
					if changes.SecurityGroups != nil {
						securityGroupIDs := make([]string, len(e.SecurityGroups))
						for i, sg := range e.SecurityGroups {
							securityGroupIDs[i] = *sg.ID
						}
						if group.Compute == nil {
							group.Compute = new(aws.Compute)
						}
						if group.Compute.LaunchSpecification == nil {
							group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
						}
						group.Compute.LaunchSpecification.SetSecurityGroupIDs(securityGroupIDs)
						changes.SecurityGroups = nil
					}
				}

				// User data.
				{
					if changes.UserData != nil {
						userData, err := e.UserData.AsString()
						if err != nil {
							return err
						}
						encoded := base64.StdEncoding.EncodeToString([]byte(userData))
						if group.Compute == nil {
							group.Compute = new(aws.Compute)
						}
						if group.Compute.LaunchSpecification == nil {
							group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
						}
						group.Compute.LaunchSpecification.SetUserData(spotinstsdk.String(encoded))
						changes.UserData = nil
					}
				}

				// Network interfaces.
				{
					if changes.AssociatePublicIP != nil {
						if *changes.AssociatePublicIP {
							iface := new(aws.NetworkInterface)
							iface.SetDeviceIndex(spotinstsdk.Int(0))
							iface.SetAssociatePublicIPAddress(spotinstsdk.Bool(true))
							iface.SetDeleteOnTermination(spotinstsdk.Bool(true))
							if group.Compute == nil {
								group.Compute = new(aws.Compute)
							}
							if group.Compute.LaunchSpecification == nil {
								group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
							}
							group.Compute.LaunchSpecification.SetNetworkInterfaces(
								[]*aws.NetworkInterface{iface},
							)
						}
						changes.AssociatePublicIP = nil
					}
				}

				// Image ID.
				{
					if changes.ImageID != nil {
						image, err := t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).ResolveImage(*e.ImageID)
						if err != nil {
							return fmt.Errorf("spotinst: unable to resolve image: %q: %v", *e.ImageID, err)
						} else if image == nil {
							return fmt.Errorf("spotinst: unable to resolve image: %q: not found", *e.ImageID)
						}
						if *actual.Compute.LaunchSpecification.ImageID != *image.ImageId {
							if group.Compute == nil {
								group.Compute = new(aws.Compute)
							}
							if group.Compute.LaunchSpecification == nil {
								group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
							}
							group.Compute.LaunchSpecification.SetImageId(image.ImageId)
						}
						changes.ImageID = nil
					}
				}

				// Tags.
				{
					if changes.Tags != nil {
						tags := e.buildTags()
						if group.Compute == nil {
							group.Compute = new(aws.Compute)
						}
						if group.Compute.LaunchSpecification == nil {
							group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
						}
						group.Compute.LaunchSpecification.SetTags(tags)
						changes.Tags = nil
					}
				}

				// IAM instance profile.
				{
					if changes.IAMInstanceProfile != nil {
						iprof := new(aws.IAMInstanceProfile)
						iprof.SetName(e.IAMInstanceProfile.GetName())
						if group.Compute == nil {
							group.Compute = new(aws.Compute)
						}
						if group.Compute.LaunchSpecification == nil {
							group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
						}
						group.Compute.LaunchSpecification.SetIAMInstanceProfile(iprof)
						changes.IAMInstanceProfile = nil
					}
				}

				// SSH key.
				{
					if changes.SSHKey != nil {
						if group.Compute == nil {
							group.Compute = new(aws.Compute)
						}
						if group.Compute.LaunchSpecification == nil {
							group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
						}
						group.Compute.LaunchSpecification.SetKeyPair(e.SSHKey.Name)
						changes.SSHKey = nil
					}
				}

				// Tenancy.
				{
					if changes.Tenancy != nil {
						if group.Compute == nil {
							group.Compute = new(aws.Compute)
						}
						if group.Compute.LaunchSpecification == nil {
							group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
						}
						group.Compute.LaunchSpecification.SetTenancy(e.Tenancy)
						changes.Tenancy = nil
					}
				}
			}
		}

		// Capacity.
		{
			if changes.MinSize != nil {
				if group.Capacity == nil {
					group.Capacity = new(aws.Capacity)
				}
				group.Capacity.SetMinimum(spotinstsdk.Int(int(*e.MinSize)))
				changes.MinSize = nil

				// Scale up the target capacity, if needed.
				actual, err := findAutoscalingGroup(t.Cloud.(spotinst.Cloud), *e.Name)
				if err == nil && actual != nil {
					if int64(*actual.Capacity.Target) < *e.MinSize {
						group.Capacity.SetTarget(spotinstsdk.Int(int(*e.MinSize)))
					}
				}
			}
			if changes.MaxSize != nil {
				if group.Capacity == nil {
					group.Capacity = new(aws.Capacity)
				}
				group.Capacity.SetMaximum(spotinstsdk.Int(int(*e.MaxSize)))
				changes.MaxSize = nil
			}
		}

		// Integration.
		{
			if v := changes.IntegrationAutoScaleNodeLabels; v != nil && len(v) > 0 {
				if group.Integration == nil {
					group.Integration = new(aws.Integration)
				}
				if group.Integration.Kubernetes == nil {
					group.Integration.Kubernetes = new(aws.KubernetesIntegration)
				}
				if group.Integration.Kubernetes.AutoScale == nil {
					group.Integration.Kubernetes.AutoScale = new(aws.AutoScale)
				}

				labels := e.buildAutoScaleLabels(v)
				group.Integration.Kubernetes.AutoScale.SetLabels(labels)
				changes.IntegrationAutoScaleNodeLabels = nil
			}
		}

		empty := &AutoscalingGroup{}
		if !reflect.DeepEqual(empty, changes) {
			glog.Warningf("Not all changes applied to group: %v", changes)
		}

		if group.Compute == nil &&
			group.Capacity == nil &&
			group.Strategy == nil &&
			group.Integration == nil {
			glog.V(2).Infof("No changes detected in group: %s", *group.ID)
			return nil
		}

		glog.V(2).Infof("Updating group %q group: %s", *group.ID, stringutil.Stringify(group))
		input := &aws.UpdateGroupInput{Group: group}
		_, err = t.Cloud.(spotinst.Cloud).Service().CloudProviderAWS().Update(context.Background(), input)
		if err != nil {
			return fmt.Errorf("spotinst: failed to update group: %v", err)
		}
	}

	return nil
}

func (e *AutoscalingGroup) buildTags() []*aws.Tag {
	tags := make([]*aws.Tag, 0, len(e.Tags))
	for key, value := range e.Tags {
		tags = append(tags, &aws.Tag{
			Key:   spotinstsdk.String(key),
			Value: spotinstsdk.String(value),
		})
	}
	return tags
}

func (e *AutoscalingGroup) buildAutoScaleLabels(labelsMap map[string]string) []*aws.AutoScaleLabel {
	labels := make([]*aws.AutoScaleLabel, 0, len(labelsMap))
	for key, value := range labelsMap {
		labels = append(labels, &aws.AutoScaleLabel{
			Key:   spotinstsdk.String(key),
			Value: spotinstsdk.String(value),
		})
	}
	return labels
}

func (e *AutoscalingGroup) buildEphemeralDevices(instanceTypeName *string) (map[string]*BlockDeviceMapping, error) {
	if instanceTypeName == nil {
		return nil, fi.RequiredField("InstanceType")
	}
	instanceType, err := awsup.GetMachineTypeInfo(*instanceTypeName)
	if err != nil {
		return nil, err
	}
	blockDeviceMappings := make(map[string]*BlockDeviceMapping)
	for _, ed := range instanceType.EphemeralDevices() {
		m := &BlockDeviceMapping{VirtualName: fi.String(ed.VirtualName)}
		blockDeviceMappings[ed.DeviceName] = m
	}
	return blockDeviceMappings, nil
}

func (e *AutoscalingGroup) buildRootDevice(cloud spotinst.Cloud) (map[string]*BlockDeviceMapping, error) {
	imageID := fi.StringValue(e.ImageID)
	image, err := cloud.Cloud().(awsup.AWSCloud).ResolveImage(imageID)
	if err != nil {
		return nil, fmt.Errorf("spotinst: unable to resolve image: %q: %v", imageID, err)
	} else if image == nil {
		return nil, fmt.Errorf("spotinst: unable to resolve image: %q: not found", imageID)
	}

	rootDeviceName := spotinstsdk.StringValue(image.RootDeviceName)
	blockDeviceMappings := make(map[string]*BlockDeviceMapping)
	rootDeviceMapping := &BlockDeviceMapping{
		EbsDeleteOnTermination: spotinstsdk.Bool(true),
		EbsVolumeSize:          e.RootVolumeSize,
		EbsVolumeType:          e.RootVolumeType,
		EbsVolumeIOPS:          e.RootVolumeIOPS,
	}
	blockDeviceMappings[rootDeviceName] = rootDeviceMapping

	return blockDeviceMappings, nil
}

type Orientation string

const (
	OrientationBalanced              Orientation = "balanced"
	OrientationCost                  Orientation = "costOriented"
	OrientationAvailability          Orientation = "availabilityOriented"
	OrientationEqualZoneDistribution Orientation = "equalAzDistribution"
)

func normalizeOrientation(orientation *string) Orientation {
	out := OrientationBalanced

	// Fast path.
	if orientation == nil {
		return out
	}

	switch *orientation {
	case "cost":
		out = OrientationCost
	case "availability":
		out = OrientationAvailability
	case "equal-distribution":
		out = OrientationEqualZoneDistribution
	}

	return out
}

type Product string

const (
	ProductLinuxUnix    Product = "Linux/UNIX"
	ProductLinuxUnixVPC Product = "Linux/UNIX (Amazon VPC)"
)

func normalizeProduct(product *string) Product {
	out := ProductLinuxUnix

	// Fast path.
	if product == nil {
		return out
	}

	switch *product {
	case "Linux/UNIX":
		out = ProductLinuxUnix
	case "Linux/UNIX (Amazon VPC)":
		out = ProductLinuxUnixVPC
	}

	return out
}
