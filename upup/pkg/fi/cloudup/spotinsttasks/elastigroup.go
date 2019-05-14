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

package spotinsttasks

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spotinst/spotinst-sdk-go/service/elastigroup/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/stringutil"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	"k8s.io/kops/pkg/resources/spotinst"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/utils"
)

//go:generate fitask -type=Elastigroup
type Elastigroup struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	ID                       *string
	MinSize                  *int64
	MaxSize                  *int64
	Risk                     *float64
	UtilizeReservedInstances *bool
	FallbackToOnDemand       *bool
	Product                  *string
	Orientation              *string
	Tags                     map[string]string
	UserData                 *fi.ResourceHolder
	ImageID                  *string
	OnDemandInstanceType     *string
	SpotInstanceTypes        []string
	IAMInstanceProfile       *awstasks.IAMInstanceProfile
	LoadBalancer             *awstasks.LoadBalancer
	SSHKey                   *awstasks.SSHKey
	Subnets                  []*awstasks.Subnet
	SecurityGroups           []*awstasks.SecurityGroup
	Monitoring               *bool
	AssociatePublicIP        *bool
	RootVolumeSize           *int64
	RootVolumeType           *string
	RootVolumeIOPS           *int64
	RootVolumeOptimization   *bool
	Tenancy                  *string
	AutoScalerEnabled        *bool
	AutoScalerClusterID      *string
	AutoScalerNodeLabels     map[string]string
}

var _ fi.CompareWithID = &Elastigroup{}

func (e *Elastigroup) CompareWithID() *string {
	return e.Name
}

func (e *Elastigroup) find(svc spotinst.Service, name string) (*aws.Group, error) {
	groups, err := svc.List(context.Background())
	if err != nil {
		return nil, fmt.Errorf("spotinst: failed to find elastigroup %s: %v", name, err)
	}

	var out *aws.Group
	for _, group := range groups {
		if group.Name() == name {
			out = group.Obj().(*aws.Group)
			break
		}
	}
	if out == nil {
		return nil, fmt.Errorf("spotinst: failed to find elastigroup %q", name)
	}

	return out, nil
}

var _ fi.HasCheckExisting = &Elastigroup{}

func (e *Elastigroup) Find(c *fi.Context) (*Elastigroup, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	group, err := e.find(cloud.Spotinst(), *e.Name)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, nil
	}

	actual := &Elastigroup{}
	actual.ID = group.ID
	actual.Name = group.Name
	actual.MinSize = fi.Int64(int64(fi.IntValue(group.Capacity.Minimum)))
	actual.MaxSize = fi.Int64(int64(fi.IntValue(group.Capacity.Maximum)))
	actual.Risk = group.Strategy.Risk
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
					actual.Subnets = append(actual.Subnets,
						&awstasks.Subnet{ID: zone.SubnetID})
				}
			}
			if subnetSlicesEqualIgnoreOrder(actual.Subnets, e.Subnets) {
				actual.Subnets = e.Subnets
			}
		}
	}

	// Launch specification.
	{
		lc := group.Compute.LaunchSpecification

		// Image.
		{
			actual.ImageID = lc.ImageID

			if e.ImageID != nil && actual.ImageID != nil &&
				fi.StringValue(actual.ImageID) != fi.StringValue(e.ImageID) {
				image, err := resolveImage(cloud, fi.StringValue(e.ImageID))
				if err != nil {
					return nil, err
				}
				if fi.StringValue(image.ImageId) == fi.StringValue(lc.ImageID) {
					actual.ImageID = e.ImageID
				}
			}
		}

		// Tags.
		{
			if len(lc.Tags) > 0 {
				actual.Tags = make(map[string]string)
				for _, tag := range lc.Tags {
					actual.Tags[fi.StringValue(tag.Key)] = fi.StringValue(tag.Value)
				}
			}
		}

		// Security groups.
		{
			for _, sgID := range lc.SecurityGroupIDs {
				actual.SecurityGroups = append(actual.SecurityGroups,
					&awstasks.SecurityGroup{ID: fi.String(sgID)})
			}
		}

		// Block device mappings.
		{
			for _, b := range lc.BlockDeviceMappings {
				if b.EBS == nil || b.EBS.SnapshotID != nil {
					continue // not the root
				}

				actual.RootVolumeType = b.EBS.VolumeType
				actual.RootVolumeSize = fi.Int64(int64(fi.IntValue(b.EBS.VolumeSize)))
				actual.RootVolumeIOPS = fi.Int64(int64(fi.IntValue(b.EBS.IOPS)))
			}
		}

		// EBS optimization.
		{
			if lc.EBSOptimized != nil {
				actual.RootVolumeOptimization = lc.EBSOptimized
			}
		}

		// User data.
		{
			if lc.UserData != nil {
				userData, err := base64.StdEncoding.DecodeString(fi.StringValue(lc.UserData))
				if err != nil {
					return nil, err
				}
				actual.UserData = fi.WrapResource(fi.NewStringResource(string(userData)))
			}
		}

		// Network interfaces.
		{
			associatePublicIP := false
			if len(lc.NetworkInterfaces) > 0 {
				for _, iface := range lc.NetworkInterfaces {
					if fi.BoolValue(iface.AssociatePublicIPAddress) {
						associatePublicIP = true
						break
					}
				}
			}
			actual.AssociatePublicIP = fi.Bool(associatePublicIP)
		}

		// Load balancer.
		{
			if lc.LoadBalancersConfig != nil && len(lc.LoadBalancersConfig.LoadBalancers) > 0 {
				lbs := lc.LoadBalancersConfig.LoadBalancers
				actual.LoadBalancer = &awstasks.LoadBalancer{Name: lbs[0].Name}

				if e.LoadBalancer != nil && actual.LoadBalancer != nil &&
					fi.StringValue(actual.LoadBalancer.Name) != fi.StringValue(e.LoadBalancer.Name) {
					elb, err := awstasks.FindLoadBalancerByNameTag(cloud, fi.StringValue(e.LoadBalancer.Name))
					if err != nil {
						return nil, err
					}
					if fi.StringValue(elb.LoadBalancerName) == fi.StringValue(lbs[0].Name) {
						actual.LoadBalancer = e.LoadBalancer
					}
				}
			}
		}

		// IAM instance profile.
		if lc.IAMInstanceProfile != nil {
			actual.IAMInstanceProfile = &awstasks.IAMInstanceProfile{Name: lc.IAMInstanceProfile.Name}
		}

		// SSH key.
		if lc.KeyPair != nil {
			actual.SSHKey = &awstasks.SSHKey{Name: lc.KeyPair}
		}

		// Tenancy.
		if lc.Tenancy != nil {
			actual.Tenancy = lc.Tenancy
		}

		// Monitoring.
		if lc.Monitoring != nil {
			actual.Monitoring = lc.Monitoring
		}
	}

	// Integration.
	{
		if group.Integration != nil && group.Integration.Kubernetes != nil {
			integration := group.Integration.Kubernetes

			// Cluster identifier.
			if integration.ClusterIdentifier != nil {
				actual.AutoScalerClusterID = integration.ClusterIdentifier
			}

			// Auto scaler.
			if integration.AutoScale != nil {
				if integration.AutoScale.IsEnabled != nil {
					actual.AutoScalerEnabled = integration.AutoScale.IsEnabled
				}

				// Labels.
				if integration.AutoScale.Labels != nil {
					labels := make(map[string]string)
					for _, label := range integration.AutoScale.Labels {
						labels[fi.StringValue(label.Key)] = fi.StringValue(label.Value)
					}
					if len(labels) > 0 {
						actual.AutoScalerNodeLabels = labels
					}
				}
			}
		}
	}

	// Avoid spurious changes
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *Elastigroup) CheckExisting(c *fi.Context) bool {
	cloud := c.Cloud.(awsup.AWSCloud)
	group, err := e.find(cloud.Spotinst(), *e.Name)
	return err == nil && group != nil
}

func (e *Elastigroup) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *Elastigroup) CheckChanges(a, e, changes *Elastigroup) error {
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

func (eg *Elastigroup) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *Elastigroup) error {
	return eg.createOrUpdate(t.Cloud.(awsup.AWSCloud), a, e, changes)
}

func (eg *Elastigroup) createOrUpdate(cloud awsup.AWSCloud, a, e, changes *Elastigroup) error {
	if a == nil {
		return eg.create(cloud, a, e, changes)
	} else {
		return eg.update(cloud, a, e, changes)
	}
}

func (_ *Elastigroup) create(cloud awsup.AWSCloud, a, e, changes *Elastigroup) error {
	klog.V(2).Infof("Creating elastigroup %q", *e.Name)
	e.applyDefaults()

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
		group.Capacity.SetTarget(fi.Int(int(*e.MinSize)))
		group.Capacity.SetMinimum(fi.Int(int(*e.MinSize)))
		group.Capacity.SetMaximum(fi.Int(int(*e.MaxSize)))
	}

	// Strategy.
	{
		group.Strategy.SetRisk(e.Risk)
		group.Strategy.SetAvailabilityVsCost(fi.String(string(normalizeOrientation(e.Orientation))))
		group.Strategy.SetFallbackToOnDemand(e.FallbackToOnDemand)
		group.Strategy.SetUtilizeReservedInstances(e.UtilizeReservedInstances)
	}

	// Compute.
	{
		group.Compute.SetProduct(e.Product)

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
			group.Compute.LaunchSpecification.SetMonitoring(e.Monitoring)
			group.Compute.LaunchSpecification.SetKeyPair(e.SSHKey.Name)

			if e.Tenancy != nil {
				group.Compute.LaunchSpecification.SetTenancy(e.Tenancy)
			}

			// Block device mappings.
			{
				rootDevices, err := e.buildRootDevice(cloud)
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
						mappings = append(mappings, e.buildBlockDeviceMapping(device, bdm))
					}
					for device, bdm := range ephemeralDevices {
						mappings = append(mappings, e.buildBlockDeviceMapping(device, bdm))
					}
					if len(mappings) > 0 {
						group.Compute.LaunchSpecification.SetBlockDeviceMappings(mappings)
					}
				}
			}

			// Image.
			{
				image, err := resolveImage(cloud, fi.StringValue(e.ImageID))
				if err != nil {
					return err
				}
				group.Compute.LaunchSpecification.SetImageId(image.ImageId)
			}

			// User data.
			{
				userData, err := e.UserData.AsString()
				if err != nil {
					return err
				}
				encoded := base64.StdEncoding.EncodeToString([]byte(userData))
				group.Compute.LaunchSpecification.SetUserData(fi.String(encoded))
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
					iface.SetDeviceIndex(fi.Int(0))
					iface.SetAssociatePublicIPAddress(fi.Bool(true))
					iface.SetDeleteOnTermination(fi.Bool(true))
					group.Compute.LaunchSpecification.SetNetworkInterfaces(
						[]*aws.NetworkInterface{iface},
					)
				}
			}

			// Load balancer.
			{
				if e.LoadBalancer != nil {
					elb, err := awstasks.FindLoadBalancerByNameTag(cloud, fi.StringValue(e.LoadBalancer.Name))
					if err != nil {
						return err
					}
					if elb != nil {
						lb := new(aws.LoadBalancer)
						lb.SetName(elb.LoadBalancerName)
						lb.SetType(fi.String("CLASSIC"))

						cfg := new(aws.LoadBalancersConfig)
						cfg.SetLoadBalancers([]*aws.LoadBalancer{lb})

						group.Compute.LaunchSpecification.SetLoadBalancersConfig(cfg)
					}
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
		if e.AutoScalerClusterID != nil {
			k8s := new(aws.KubernetesIntegration)
			k8s.SetClusterIdentifier(e.AutoScalerClusterID)
			k8s.SetIntegrationMode(fi.String("pod"))

			if e.AutoScalerEnabled != nil {
				autoScale := new(aws.AutoScaleKubernetes)
				autoScale.SetIsEnabled(e.AutoScalerEnabled)

				labelsMap := e.AutoScalerNodeLabels
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
		klog.V(2).Infof("(%d/%d) Attempting to create elastigroup: %s, config: %s",
			attempt, maxAttempts, *e.Name, stringutil.Stringify(group))

		// Wait for IAM instance profile to be ready.
		time.Sleep(10 * time.Second)

		// Wrap the raw object as an Elastigroup.
		eg, err := spotinst.NewElastigroup(cloud.ProviderID(), group)
		if err != nil {
			return err
		}

		// Create the Elastigroup.
		id, err := cloud.Spotinst().Create(context.Background(), eg)
		if err == nil {
			e.ID = fi.String(id)
			break
		}

		if errs, ok := err.(client.Errors); ok {
			for _, err := range errs {
				if strings.Contains(err.Message, "Invalid IAM Instance Profile name") {
					if attempt > maxAttempts {
						return fmt.Errorf("IAM instance profile not yet created/propagated (original error: %v)", err)
					}

					klog.V(4).Infof("Got an error indicating that the IAM instance profile %q is not ready %q", fi.StringValue(e.IAMInstanceProfile.Name), err)
					klog.Infof("Waiting for IAM instance profile %q to be ready", fi.StringValue(e.IAMInstanceProfile.Name))
					goto readyLoop
				}
			}

			return fmt.Errorf("spotinst: failed to create elastigroup: %v", err)
		}
	}

	return nil
}

func (_ *Elastigroup) update(cloud awsup.AWSCloud, a, e, changes *Elastigroup) error {
	klog.V(2).Infof("Updating elastigroup %q", *e.Name)

	actual, err := e.find(cloud.Spotinst(), *e.Name)
	if err != nil {
		klog.Errorf("Unable to resolve elastigroup %q, error: %s", *e.Name, err)
		return err
	}

	group := new(aws.Group)
	group.SetId(actual.ID)

	// Strategy.
	{
		// Risk.
		if changes.Risk != nil {
			if group.Strategy == nil {
				group.Strategy = new(aws.Strategy)
			}

			group.Strategy.SetRisk(e.Risk)
			changes.Risk = nil
		}

		// Orientation.
		if changes.Orientation != nil {
			if group.Strategy == nil {
				group.Strategy = new(aws.Strategy)
			}

			group.Strategy.SetAvailabilityVsCost(fi.String(string(normalizeOrientation(e.Orientation))))
			changes.Orientation = nil
		}

		// Fallback to on-demand.
		if changes.FallbackToOnDemand != nil {
			if group.Strategy == nil {
				group.Strategy = new(aws.Strategy)
			}

			group.Strategy.SetFallbackToOnDemand(e.FallbackToOnDemand)
			changes.FallbackToOnDemand = nil
		}

		// Utilize reserved instances.
		if changes.UtilizeReservedInstances != nil {
			if group.Strategy == nil {
				group.Strategy = new(aws.Strategy)
			}

			group.Strategy.SetUtilizeReservedInstances(e.UtilizeReservedInstances)
			changes.UtilizeReservedInstances = nil
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

					group.Compute.LaunchSpecification.SetUserData(fi.String(encoded))
					changes.UserData = nil
				}
			}

			// Network interfaces.
			{
				if changes.AssociatePublicIP != nil {
					if *changes.AssociatePublicIP {
						iface := new(aws.NetworkInterface)
						iface.SetDeviceIndex(fi.Int(0))
						iface.SetAssociatePublicIPAddress(fi.Bool(true))
						iface.SetDeleteOnTermination(fi.Bool(true))

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

			// Block device mappings.
			{
				if changes.RootVolumeType != nil || changes.RootVolumeSize != nil || changes.RootVolumeIOPS != nil {
					rootDevices, err := e.buildRootDevice(cloud)
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
							mappings = append(mappings, e.buildBlockDeviceMapping(device, bdm))
						}
						for device, bdm := range ephemeralDevices {
							mappings = append(mappings, e.buildBlockDeviceMapping(device, bdm))
						}
						if len(mappings) > 0 {
							if group.Compute == nil {
								group.Compute = new(aws.Compute)
							}
							if group.Compute.LaunchSpecification == nil {
								group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
							}

							group.Compute.LaunchSpecification.SetBlockDeviceMappings(mappings)
						}
					}

					changes.RootVolumeType = nil
					changes.RootVolumeSize = nil
					changes.RootVolumeIOPS = nil
				}
			}

			// Image.
			{
				if changes.ImageID != nil {
					image, err := resolveImage(cloud, fi.StringValue(e.ImageID))
					if err != nil {
						return err
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

			// Monitoring.
			{
				if changes.Monitoring != nil {
					if group.Compute == nil {
						group.Compute = new(aws.Compute)
					}
					if group.Compute.LaunchSpecification == nil {
						group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
					}

					group.Compute.LaunchSpecification.SetMonitoring(e.Monitoring)
					changes.Monitoring = nil
				}
			}

			// EBS optimization.
			{
				if changes.RootVolumeOptimization != nil {
					if group.Compute == nil {
						group.Compute = new(aws.Compute)
					}
					if group.Compute.LaunchSpecification == nil {
						group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
					}

					group.Compute.LaunchSpecification.SetEBSOptimized(e.RootVolumeOptimization)
					changes.RootVolumeOptimization = nil
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

			// Load balancer.
			{
				if changes.LoadBalancer != nil {
					elb, err := awstasks.FindLoadBalancerByNameTag(cloud, fi.StringValue(e.LoadBalancer.Name))
					if err != nil {
						return err
					}
					if elb != nil {
						lb := new(aws.LoadBalancer)
						lb.SetName(elb.LoadBalancerName)
						lb.SetType(fi.String("CLASSIC"))

						cfg := new(aws.LoadBalancersConfig)
						cfg.SetLoadBalancers([]*aws.LoadBalancer{lb})

						if group.Compute == nil {
							group.Compute = new(aws.Compute)
						}
						if group.Compute.LaunchSpecification == nil {
							group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
						}

						group.Compute.LaunchSpecification.SetLoadBalancersConfig(cfg)
						changes.LoadBalancer = nil
					}
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

			group.Capacity.SetMinimum(fi.Int(int(*e.MinSize)))
			changes.MinSize = nil

			// Scale up the target capacity, if needed.
			actual, err := e.find(cloud.Spotinst(), *e.Name)
			if err == nil && actual != nil {
				if int64(*actual.Capacity.Target) < *e.MinSize {
					group.Capacity.SetTarget(fi.Int(int(*e.MinSize)))
				}
			}
		}
		if changes.MaxSize != nil {
			if group.Capacity == nil {
				group.Capacity = new(aws.Capacity)
			}

			group.Capacity.SetMaximum(fi.Int(int(*e.MaxSize)))
			changes.MaxSize = nil
		}
	}

	// Integration.
	{
		if changes.AutoScalerClusterID != nil {
			if group.Integration == nil {
				group.Integration = new(aws.Integration)
			}
			if group.Integration.Kubernetes == nil {
				group.Integration.Kubernetes = new(aws.KubernetesIntegration)
			}

			group.Integration.Kubernetes.SetClusterIdentifier(e.AutoScalerClusterID)
			group.Integration.Kubernetes.SetIntegrationMode(fi.String("pod"))
			changes.AutoScalerClusterID = nil
		}

		if changes.AutoScalerEnabled != nil {
			if group.Integration == nil {
				group.Integration = new(aws.Integration)
			}
			if group.Integration.Kubernetes == nil {
				group.Integration.Kubernetes = new(aws.KubernetesIntegration)
			}
			if group.Integration.Kubernetes.AutoScale == nil {
				group.Integration.Kubernetes.AutoScale = new(aws.AutoScaleKubernetes)
			}

			group.Integration.Kubernetes.AutoScale.SetIsEnabled(e.AutoScalerEnabled)
			changes.AutoScalerEnabled = nil
		}

		if nodeLabels := changes.AutoScalerNodeLabels; nodeLabels != nil && len(nodeLabels) > 0 {
			if group.Integration == nil {
				group.Integration = new(aws.Integration)
			}
			if group.Integration.Kubernetes == nil {
				group.Integration.Kubernetes = new(aws.KubernetesIntegration)
			}
			if group.Integration.Kubernetes.AutoScale == nil {
				group.Integration.Kubernetes.AutoScale = new(aws.AutoScaleKubernetes)
			}

			group.Integration.Kubernetes.AutoScale.SetLabels(e.buildAutoScaleLabels(nodeLabels))
			changes.AutoScalerNodeLabels = nil
		}
	}

	empty := &Elastigroup{}
	if !reflect.DeepEqual(empty, changes) {
		klog.Warningf("Not all changes applied to elastigroup %q: %v", *group.ID, changes)
	}

	if group.Compute == nil &&
		group.Capacity == nil &&
		group.Strategy == nil &&
		group.Integration == nil {
		klog.V(2).Infof("No changes detected in elastigroup %q", *group.ID)
		return nil
	}

	klog.V(2).Infof("Updating elastigroup %q (config: %s)", *group.ID, stringutil.Stringify(group))

	// Wrap the raw object as an Elastigroup.
	eg, err := spotinst.NewElastigroup(cloud.ProviderID(), group)
	if err != nil {
		return err
	}

	// Update the Elastigroup.
	if err := cloud.Spotinst().Update(context.Background(), eg); err != nil {
		return fmt.Errorf("spotinst: failed to update elastigroup: %v", err)
	}

	return nil
}

type terraformElastigroup struct {
	Name                 *string                                 `json:"name,omitempty"`
	Description          *string                                 `json:"description,omitempty"`
	Product              *string                                 `json:"product,omitempty"`
	Region               *string                                 `json:"region,omitempty"`
	SubnetIDs            []*terraform.Literal                    `json:"subnet_ids,omitempty"`
	LoadBalancers        []*terraform.Literal                    `json:"elastic_load_balancers,omitempty"`
	NetworkInterfaces    []*terraformElastigroupNetworkInterface `json:"network_interface,omitempty"`
	RootBlockDevice      *terraformElastigroupBlockDevice        `json:"ebs_block_device,omitempty"`
	EphemeralBlockDevice []*terraformElastigroupBlockDevice      `json:"ephemeral_block_device,omitempty"`
	Integration          *terraformElastigroupIntegration        `json:"integration_kubernetes,omitempty"`
	Tags                 []*terraformElastigroupTag              `json:"tags,omitempty"`

	*terraformElastigroupCapacity
	*terraformElastigroupStrategy
	*terraformElastigroupInstanceTypes
	*terraformElastigroupLaunchSpec
}

type terraformElastigroupCapacity struct {
	MinSize         *int64  `json:"min_size,omitempty"`
	MaxSize         *int64  `json:"max_size,omitempty"`
	DesiredCapacity *int64  `json:"desired_capacity,omitempty"`
	CapacityUnit    *string `json:"capacity_unit,omitempty"`
}

type terraformElastigroupStrategy struct {
	SpotPercentage           *float64 `json:"spot_percentage,omitempty"`
	Orientation              *string  `json:"orientation,omitempty"`
	FallbackToOnDemand       *bool    `json:"fallback_to_ondemand,omitempty"`
	UtilizeReservedInstances *bool    `json:"utilize_reserved_instances,omitempty"`
}

type terraformElastigroupInstanceTypes struct {
	OnDemand *string  `json:"instance_types_ondemand,omitempty"`
	Spot     []string `json:"instance_types_spot,omitempty"`
}

type terraformElastigroupLaunchSpec struct {
	Monitoring         *bool                `json:"enable_monitoring,omitempty"`
	EBSOptimized       *bool                `json:"ebs_optimized,omitempty"`
	ImageID            *string              `json:"image_id,omitempty"`
	SecurityGroups     []*terraform.Literal `json:"security_groups,omitempty"`
	UserData           *terraform.Literal   `json:"user_data,omitempty"`
	IAMInstanceProfile *terraform.Literal   `json:"iam_instance_profile,omitempty"`
	KeyName            *terraform.Literal   `json:"key_name,omitempty"`
}

type terraformElastigroupBlockDevice struct {
	DeviceName          *string `json:"device_name,omitempty"`
	VirtualName         *string `json:"virtual_name,omitempty"`
	VolumeType          *string `json:"volume_type,omitempty"`
	VolumeSize          *int64  `json:"volume_size,omitempty"`
	DeleteOnTermination *bool   `json:"delete_on_termination,omitempty"`
}

type terraformElastigroupNetworkInterface struct {
	Description              *string `json:"description,omitempty"`
	DeviceIndex              *int    `json:"device_index,omitempty"`
	AssociatePublicIPAddress *bool   `json:"associate_public_ip_address,omitempty"`
	DeleteOnTermination      *bool   `json:"delete_on_termination,omitempty"`
}

type terraformElastigroupIntegration struct {
	IntegrationMode    *string `json:"integration_mode,omitempty"`
	ClusterIdentifier  *string `json:"cluster_identifier,omitempty"`
	AutoScaleIsEnabled *bool   `json:"autoscale_is_enabled,omitempty"`
}

type terraformElastigroupTag struct {
	Key   *string `json:"key"`
	Value *string `json:"value"`
}

func (_ *Elastigroup) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Elastigroup) error {
	cloud := t.Cloud.(awsup.AWSCloud)
	e.applyDefaults()

	tf := &terraformElastigroup{
		Name:        e.Name,
		Description: e.Name,
		Product:     e.Product,
		Region:      fi.String(cloud.Region()),
		terraformElastigroupCapacity: &terraformElastigroupCapacity{
			DesiredCapacity: e.MinSize,
			MinSize:         e.MinSize,
			MaxSize:         e.MaxSize,
			CapacityUnit:    fi.String("instance"),
		},
		terraformElastigroupStrategy: &terraformElastigroupStrategy{
			SpotPercentage:           e.Risk,
			Orientation:              fi.String(string(normalizeOrientation(e.Orientation))),
			FallbackToOnDemand:       e.FallbackToOnDemand,
			UtilizeReservedInstances: e.UtilizeReservedInstances,
		},
		terraformElastigroupInstanceTypes: &terraformElastigroupInstanceTypes{
			OnDemand: e.OnDemandInstanceType,
			Spot:     e.SpotInstanceTypes,
		},
		terraformElastigroupLaunchSpec: &terraformElastigroupLaunchSpec{},
	}

	// Image.
	{
		image, err := resolveImage(cloud, fi.StringValue(e.ImageID))
		if err != nil {
			return err
		}
		tf.ImageID = image.ImageId
	}

	var role string
	for key := range e.Tags {
		if strings.HasPrefix(key, awstasks.CloudTagInstanceGroupRolePrefix) {
			suffix := strings.TrimPrefix(key, awstasks.CloudTagInstanceGroupRolePrefix)
			if role != "" && role != suffix {
				return fmt.Errorf("spotinst: found multiple role tags %q vs %q", role, suffix)
			}
			role = suffix
		}
	}

	// Security groups.
	{
		for _, sg := range e.SecurityGroups {
			tf.SecurityGroups = append(tf.SecurityGroups, sg.TerraformLink())
			if role != "" {
				if err := t.AddOutputVariableArray(role+"_security_groups", sg.TerraformLink()); err != nil {
					return err
				}
			}
		}
	}

	// User data.
	{
		if e.UserData != nil {
			var err error
			tf.UserData, err = t.AddFile("spotinst_elastigroup_aws", *e.Name, "user_data", e.UserData)
			if err != nil {
				return err
			}
		}
	}

	// IAM instance profile.
	{
		if e.IAMInstanceProfile != nil {
			tf.IAMInstanceProfile = e.IAMInstanceProfile.TerraformLink()
		}
	}

	// Monitoring.
	{
		if e.Monitoring != nil {
			tf.Monitoring = e.Monitoring
		} else {
			tf.Monitoring = fi.Bool(false)
		}
	}

	// EBS optimization.
	{
		if e.RootVolumeOptimization != nil {
			tf.EBSOptimized = e.RootVolumeOptimization
		} else {
			tf.EBSOptimized = fi.Bool(false)
		}
	}

	// SSH key.
	{
		if e.SSHKey != nil {
			tf.KeyName = e.SSHKey.TerraformLink()
		}
	}

	// Subnets.
	{
		for _, subnet := range e.Subnets {
			tf.SubnetIDs = append(tf.SubnetIDs, subnet.TerraformLink())
			if role != "" {
				if err := t.AddOutputVariableArray(role+"_subnet_ids", subnet.TerraformLink()); err != nil {
					return err
				}
			}
		}
	}

	// Load balancer.
	{
		if e.LoadBalancer != nil {
			tf.LoadBalancers = append(tf.LoadBalancers, e.LoadBalancer.TerraformLink())
		}
	}

	// Public IP.
	{
		if e.AssociatePublicIP != nil && *e.AssociatePublicIP {
			tf.NetworkInterfaces = append(tf.NetworkInterfaces, &terraformElastigroupNetworkInterface{
				Description:              fi.String("eth0"),
				DeviceIndex:              fi.Int(0),
				AssociatePublicIPAddress: fi.Bool(true),
				DeleteOnTermination:      fi.Bool(true),
			})
		}
	}

	// Block Devices.
	{
		rootDevices, err := e.buildRootDevice(t.Cloud.(awsup.AWSCloud))
		if err != nil {
			return err
		}

		ephemeralDevices, err := e.buildEphemeralDevices(e.OnDemandInstanceType)
		if err != nil {
			return err
		}

		if len(rootDevices) != 0 {
			if len(rootDevices) != 1 {
				return fmt.Errorf("unexpectedly found multiple root devices")
			}

			for name, bdm := range rootDevices {
				tf.RootBlockDevice = &terraformElastigroupBlockDevice{
					DeviceName:          fi.String(name),
					VolumeType:          bdm.EbsVolumeType,
					VolumeSize:          bdm.EbsVolumeSize,
					DeleteOnTermination: fi.Bool(true),
				}
			}
		}

		if len(ephemeralDevices) != 0 {
			tf.EphemeralBlockDevice = []*terraformElastigroupBlockDevice{}
			for _, deviceName := range sets.StringKeySet(ephemeralDevices).List() {
				bdm := ephemeralDevices[deviceName]
				tf.EphemeralBlockDevice = append(tf.EphemeralBlockDevice, &terraformElastigroupBlockDevice{
					VirtualName: bdm.VirtualName,
					DeviceName:  fi.String(deviceName),
				})
			}
		}
	}

	// Integration.
	{
		if e.AutoScalerClusterID != nil {
			tf.Integration = &terraformElastigroupIntegration{
				IntegrationMode:   fi.String("pod"),
				ClusterIdentifier: e.AutoScalerClusterID,
			}
			if e.AutoScalerEnabled != nil {
				tf.Integration.AutoScaleIsEnabled = e.AutoScalerEnabled
			}
		}
	}

	// Tags.
	{
		tags := e.buildTags()
		for _, tag := range tags {
			tf.Tags = append(tf.Tags, &terraformElastigroupTag{
				Key:   tag.Key,
				Value: tag.Value,
			})
		}
	}

	return t.RenderResource("spotinst_elastigroup_aws", *e.Name, tf)
}

func (e *Elastigroup) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("spotinst_elastigroup_aws", *e.Name, "id")
}

func (e *Elastigroup) buildTags() []*aws.Tag {
	tags := make([]*aws.Tag, 0, len(e.Tags))

	for key, value := range e.Tags {
		tags = append(tags, &aws.Tag{
			Key:   fi.String(key),
			Value: fi.String(value),
		})
	}

	return tags
}

func (e *Elastigroup) buildAutoScaleLabels(labelsMap map[string]string) []*aws.AutoScaleLabel {
	labels := make([]*aws.AutoScaleLabel, 0, len(labelsMap))

	for key, value := range labelsMap {
		labels = append(labels, &aws.AutoScaleLabel{
			Key:   fi.String(key),
			Value: fi.String(value),
		})
	}

	return labels
}

func (e *Elastigroup) buildEphemeralDevices(instanceTypeName *string) (map[string]*awstasks.BlockDeviceMapping, error) {
	if instanceTypeName == nil {
		return nil, fi.RequiredField("InstanceType")
	}

	instanceType, err := awsup.GetMachineTypeInfo(*instanceTypeName)
	if err != nil {
		return nil, err
	}

	blockDeviceMappings := make(map[string]*awstasks.BlockDeviceMapping)
	for _, ed := range instanceType.EphemeralDevices() {
		m := &awstasks.BlockDeviceMapping{
			VirtualName: fi.String(ed.VirtualName),
		}
		blockDeviceMappings[ed.DeviceName] = m
	}

	return blockDeviceMappings, nil
}

func (e *Elastigroup) buildRootDevice(cloud awsup.AWSCloud) (map[string]*awstasks.BlockDeviceMapping, error) {
	image, err := resolveImage(cloud, fi.StringValue(e.ImageID))
	if err != nil {
		return nil, err
	}

	rootDeviceName := fi.StringValue(image.RootDeviceName)
	blockDeviceMappings := make(map[string]*awstasks.BlockDeviceMapping)

	rootDeviceMapping := &awstasks.BlockDeviceMapping{
		EbsDeleteOnTermination: fi.Bool(true),
		EbsVolumeSize:          e.RootVolumeSize,
		EbsVolumeType:          e.RootVolumeType,
		EbsVolumeIops:          e.RootVolumeIOPS,
	}
	blockDeviceMappings[rootDeviceName] = rootDeviceMapping

	return blockDeviceMappings, nil
}

func (e *Elastigroup) buildBlockDeviceMapping(deviceName string, i *awstasks.BlockDeviceMapping) *aws.BlockDeviceMapping {
	o := &aws.BlockDeviceMapping{}
	o.DeviceName = fi.String(deviceName)
	o.VirtualName = i.VirtualName

	if i.EbsDeleteOnTermination != nil || i.EbsVolumeSize != nil || i.EbsVolumeType != nil {
		o.EBS = &aws.EBS{}
		o.EBS.DeleteOnTermination = i.EbsDeleteOnTermination
		o.EBS.VolumeSize = fi.Int(int(fi.Int64Value(i.EbsVolumeSize)))
		o.EBS.VolumeType = i.EbsVolumeType

		// The parameter IOPS is not supported for gp2 volumes.
		if fi.StringValue(i.EbsVolumeType) != "gp2" {
			o.EBS.IOPS = fi.Int(int(fi.Int64Value(i.EbsVolumeIops)))
		}
	}

	return o
}

func (e *Elastigroup) applyDefaults() {
	if e.FallbackToOnDemand == nil {
		e.FallbackToOnDemand = fi.Bool(true)
	}

	if e.UtilizeReservedInstances == nil {
		e.UtilizeReservedInstances = fi.Bool(true)
	}

	if e.Product == nil || (e.Product != nil && fi.StringValue(e.Product) == "") {
		e.Product = fi.String("Linux/UNIX")
	}

	if e.Orientation == nil || (e.Orientation != nil && fi.StringValue(e.Orientation) == "") {
		e.Orientation = fi.String("balanced")
	}
}

func resolveImage(cloud awsup.AWSCloud, name string) (*ec2.Image, error) {
	image, err := cloud.ResolveImage(name)
	if err != nil {
		return nil, fmt.Errorf("spotinst: unable to resolve image %q: %v", name, err)
	} else if image == nil {
		return nil, fmt.Errorf("spotinst: unable to resolve image %q: not found", name)
	}

	return image, nil
}

func subnetSlicesEqualIgnoreOrder(l, r []*awstasks.Subnet) bool {
	var lIDs []string
	for _, s := range l {
		lIDs = append(lIDs, *s.ID)
	}

	var rIDs []string
	for _, s := range r {
		if s.ID == nil {
			klog.V(4).Infof("Subnet ID not set; returning not-equal: %v", s)
			return false
		}
		rIDs = append(rIDs, *s.ID)
	}

	return utils.StringSlicesEqualIgnoreOrder(lIDs, rIDs)
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
