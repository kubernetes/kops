/*
Copyright 2019 The Kubernetes Authors.

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
	corev1 "k8s.io/api/core/v1"
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
	SpotPercentage           *float64
	UtilizeReservedInstances *bool
	FallbackToOnDemand       *bool
	HealthCheckType          *string
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
	Tenancy                  *string
	RootVolumeOpts           *RootVolumeOpts
	AutoScalerOpts           *AutoScalerOpts
}

type RootVolumeOpts struct {
	Type         *string
	Size         *int32
	IOPS         *int32
	Optimization *bool
}

type AutoScalerOpts struct {
	Enabled   *bool
	ClusterID *string
	Cooldown  *int
	Labels    map[string]string
	Taints    []*corev1.Taint
	Headroom  *AutoScalerHeadroomOpts
	Down      *AutoScalerDownOpts
}

type AutoScalerHeadroomOpts struct {
	CPUPerUnit *int
	GPUPerUnit *int
	MemPerUnit *int
	NumOfUnits *int
}

type AutoScalerDownOpts struct {
	MaxPercentage     *int
	EvaluationPeriods *int
}

var _ fi.CompareWithID = &Elastigroup{}

func (e *Elastigroup) CompareWithID() *string {
	return e.Name
}

var _ fi.HasDependencies = &Elastigroup{}

func (e *Elastigroup) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task

	if e.IAMInstanceProfile != nil {
		deps = append(deps, e.IAMInstanceProfile)
	}

	if e.LoadBalancer != nil {
		deps = append(deps, e.LoadBalancer)
	}

	if e.SSHKey != nil {
		deps = append(deps, e.SSHKey)
	}

	if e.Subnets != nil {
		for _, subnet := range e.Subnets {
			deps = append(deps, subnet)
		}
	}

	if e.SecurityGroups != nil {
		for _, sg := range e.SecurityGroups {
			deps = append(deps, sg)
		}
	}

	return deps
}

func (e *Elastigroup) find(svc spotinst.InstanceGroupService, name string) (*aws.Group, error) {
	klog.V(4).Infof("Attempting to find Elastigroup: %q", name)

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

	klog.V(4).Infof("Elastigroup/%s: %s", name, stringutil.Stringify(out))
	return out, nil
}

var _ fi.HasCheckExisting = &Elastigroup{}

func (e *Elastigroup) Find(c *fi.Context) (*Elastigroup, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	group, err := e.find(cloud.Spotinst().Elastigroup(), *e.Name)
	if err != nil {
		return nil, err
	}

	actual := &Elastigroup{}
	actual.ID = group.ID
	actual.Name = group.Name

	// Capacity.
	{
		actual.MinSize = fi.Int64(int64(fi.IntValue(group.Capacity.Minimum)))
		actual.MaxSize = fi.Int64(int64(fi.IntValue(group.Capacity.Maximum)))
	}

	// Strategy.
	{
		actual.SpotPercentage = group.Strategy.Risk
		actual.Orientation = group.Strategy.AvailabilityVsCost
		actual.FallbackToOnDemand = group.Strategy.FallbackToOnDemand
		actual.UtilizeReservedInstances = group.Strategy.UtilizeReservedInstances
	}

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
			if lc.Tags != nil && len(lc.Tags) > 0 {
				actual.Tags = make(map[string]string)
				for _, tag := range lc.Tags {
					actual.Tags[fi.StringValue(tag.Key)] = fi.StringValue(tag.Value)
				}
			}
		}

		// Security groups.
		{
			if lc.SecurityGroupIDs != nil {
				for _, sgID := range lc.SecurityGroupIDs {
					actual.SecurityGroups = append(actual.SecurityGroups,
						&awstasks.SecurityGroup{ID: fi.String(sgID)})
				}
			}
		}

		// Root volume options.
		{
			// Block device mappings.
			{
				if lc.BlockDeviceMappings != nil {
					for _, b := range lc.BlockDeviceMappings {
						if b.EBS == nil || b.EBS.SnapshotID != nil {
							continue // not the root
						}
						if actual.RootVolumeOpts == nil {
							actual.RootVolumeOpts = new(RootVolumeOpts)
						}
						if b.EBS.IOPS != nil {
							actual.RootVolumeOpts.IOPS = fi.Int32(int32(fi.IntValue(b.EBS.IOPS)))
						}

						actual.RootVolumeOpts.Type = b.EBS.VolumeType
						actual.RootVolumeOpts.Size = fi.Int32(int32(fi.IntValue(b.EBS.VolumeSize)))
					}
				}
			}

			// EBS optimization.
			{
				if fi.BoolValue(lc.EBSOptimized) {
					if actual.RootVolumeOpts == nil {
						actual.RootVolumeOpts = new(RootVolumeOpts)
					}

					actual.RootVolumeOpts.Optimization = lc.EBSOptimized
				}
			}
		}

		// User data.
		{
			var userData []byte

			if lc.UserData != nil {
				userData, err = base64.StdEncoding.DecodeString(fi.StringValue(lc.UserData))
				if err != nil {
					return nil, err
				}
			}

			actual.UserData = fi.WrapResource(fi.NewStringResource(string(userData)))
		}

		// Network interfaces.
		{
			associatePublicIP := false

			if lc.NetworkInterfaces != nil && len(lc.NetworkInterfaces) > 0 {
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

		// Health check.
		if lc.HealthCheckType != nil {
			actual.HealthCheckType = lc.HealthCheckType
		}
	}

	// Auto Scaler.
	{
		if group.Integration != nil && group.Integration.Kubernetes != nil {
			integration := group.Integration.Kubernetes

			actual.AutoScalerOpts = new(AutoScalerOpts)
			actual.AutoScalerOpts.ClusterID = integration.ClusterIdentifier

			if integration.AutoScale != nil {
				actual.AutoScalerOpts.Enabled = integration.AutoScale.IsEnabled
				actual.AutoScalerOpts.Cooldown = integration.AutoScale.Cooldown

				// Headroom.
				if headroom := integration.AutoScale.Headroom; headroom != nil {
					actual.AutoScalerOpts.Headroom = new(AutoScalerHeadroomOpts)

					if v := fi.IntValue(headroom.CPUPerUnit); v > 0 {
						actual.AutoScalerOpts.Headroom.CPUPerUnit = headroom.CPUPerUnit
					}
					if v := fi.IntValue(headroom.GPUPerUnit); v > 0 {
						actual.AutoScalerOpts.Headroom.GPUPerUnit = headroom.GPUPerUnit
					}
					if v := fi.IntValue(headroom.MemoryPerUnit); v > 0 {
						actual.AutoScalerOpts.Headroom.MemPerUnit = headroom.MemoryPerUnit
					}
					if v := fi.IntValue(headroom.NumOfUnits); v > 0 {
						actual.AutoScalerOpts.Headroom.NumOfUnits = headroom.NumOfUnits
					}
				}

				// Scale down.
				if down := integration.AutoScale.Down; down != nil {
					actual.AutoScalerOpts.Down = &AutoScalerDownOpts{
						MaxPercentage:     down.MaxScaleDownPercentage,
						EvaluationPeriods: down.EvaluationPeriods,
					}
				}

				// Labels.
				if labels := integration.AutoScale.Labels; labels != nil {
					actual.AutoScalerOpts.Labels = make(map[string]string)

					for _, label := range labels {
						actual.AutoScalerOpts.Labels[fi.StringValue(label.Key)] = fi.StringValue(label.Value)
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
	group, err := e.find(cloud.Spotinst().Elastigroup(), *e.Name)
	return err == nil && group != nil
}

func (e *Elastigroup) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *Elastigroup) CheckChanges(a, e, changes *Elastigroup) error {
	if e.Name == nil {
		return fi.RequiredField("Name")
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
	klog.V(2).Infof("Creating Elastigroup %q", *e.Name)
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
		group.Strategy.SetRisk(e.SpotPercentage)
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
				if e.UserData != nil {
					userData, err := e.UserData.AsString()
					if err != nil {
						return err
					}

					if len(userData) > 0 {
						encoded := base64.StdEncoding.EncodeToString([]byte(userData))
						group.Compute.LaunchSpecification.SetUserData(fi.String(encoded))
					}
				}
			}

			// IAM instance profile.
			{
				if e.IAMInstanceProfile != nil {
					iprof := new(aws.IAMInstanceProfile)
					iprof.SetName(e.IAMInstanceProfile.GetName())
					group.Compute.LaunchSpecification.SetIAMInstanceProfile(iprof)
				}
			}

			// Security groups.
			{
				if e.SecurityGroups != nil {
					securityGroupIDs := make([]string, len(e.SecurityGroups))
					for i, sg := range e.SecurityGroups {
						securityGroupIDs[i] = *sg.ID
					}
					group.Compute.LaunchSpecification.SetSecurityGroupIDs(securityGroupIDs)
				}
			}

			// Public IP.
			{
				if e.AssociatePublicIP != nil {
					iface := &aws.NetworkInterface{
						Description:              fi.String("eth0"),
						DeviceIndex:              fi.Int(0),
						DeleteOnTermination:      fi.Bool(true),
						AssociatePublicIPAddress: e.AssociatePublicIP,
					}

					group.Compute.LaunchSpecification.SetNetworkInterfaces([]*aws.NetworkInterface{iface})
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
				if e.Tags != nil {
					group.Compute.LaunchSpecification.SetTags(e.buildTags())
				}
			}

			// Health check.
			{
				if e.HealthCheckType != nil {
					group.Compute.LaunchSpecification.SetHealthCheckType(e.HealthCheckType)
				}
			}
		}
	}

	// Auto Scaler.
	{
		if opts := e.AutoScalerOpts; opts != nil {
			k8s := new(aws.KubernetesIntegration)
			k8s.SetIntegrationMode(fi.String("pod"))
			k8s.SetClusterIdentifier(opts.ClusterID)

			if opts.Enabled != nil {
				autoScaler := new(aws.AutoScaleKubernetes)
				autoScaler.IsEnabled = opts.Enabled
				autoScaler.IsAutoConfig = fi.Bool(true)
				autoScaler.Cooldown = opts.Cooldown

				// Headroom.
				if headroom := opts.Headroom; headroom != nil {
					autoScaler.IsAutoConfig = fi.Bool(false)
					autoScaler.Headroom = &aws.AutoScaleHeadroom{
						CPUPerUnit:    headroom.CPUPerUnit,
						GPUPerUnit:    headroom.GPUPerUnit,
						MemoryPerUnit: headroom.MemPerUnit,
						NumOfUnits:    headroom.NumOfUnits,
					}
				}

				// Scale down.
				if down := opts.Down; down != nil {
					autoScaler.Down = &aws.AutoScaleDown{
						MaxScaleDownPercentage: down.MaxPercentage,
						EvaluationPeriods:      down.EvaluationPeriods,
					}
				}

				// Labels.
				if labels := opts.Labels; labels != nil {
					autoScaler.Labels = e.buildAutoScaleLabels(labels)
				}

				k8s.SetAutoScale(autoScaler)
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
		klog.V(2).Infof("(%d/%d) Attempting to create Elastigroup: %q, config: %s",
			attempt, maxAttempts, *e.Name, stringutil.Stringify(group))

		// Wait for IAM instance profile to be ready.
		time.Sleep(10 * time.Second)

		// Wrap the raw object as an Elastigroup.
		eg, err := spotinst.NewElastigroup(cloud.ProviderID(), group)
		if err != nil {
			return err
		}

		// Create the Elastigroup.
		id, err := cloud.Spotinst().Elastigroup().Create(context.Background(), eg)
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
	klog.V(2).Infof("Updating Elastigroup %q", *e.Name)

	actual, err := e.find(cloud.Spotinst().Elastigroup(), *e.Name)
	if err != nil {
		klog.Errorf("Unable to resolve Elastigroup %q, error: %v", *e.Name, err)
		return err
	}

	var changed bool
	group := new(aws.Group)
	group.SetId(actual.ID)

	// Strategy.
	{
		// Spot percentage.
		if changes.SpotPercentage != nil {
			if group.Strategy == nil {
				group.Strategy = new(aws.Strategy)
			}

			group.Strategy.SetRisk(e.SpotPercentage)
			changes.SpotPercentage = nil
			changed = true
		}

		// Orientation.
		if changes.Orientation != nil {
			if group.Strategy == nil {
				group.Strategy = new(aws.Strategy)
			}

			group.Strategy.SetAvailabilityVsCost(fi.String(string(normalizeOrientation(e.Orientation))))
			changes.Orientation = nil
			changed = true
		}

		// Fallback to on-demand.
		if changes.FallbackToOnDemand != nil {
			if group.Strategy == nil {
				group.Strategy = new(aws.Strategy)
			}

			group.Strategy.SetFallbackToOnDemand(e.FallbackToOnDemand)
			changes.FallbackToOnDemand = nil
			changed = true
		}

		// Utilize reserved instances.
		if changes.UtilizeReservedInstances != nil {
			if group.Strategy == nil {
				group.Strategy = new(aws.Strategy)
			}

			group.Strategy.SetUtilizeReservedInstances(e.UtilizeReservedInstances)
			changes.UtilizeReservedInstances = nil
			changed = true
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
			changed = true
		}

		// On-demand instance type.
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
				changed = true
			}
		}

		// Spot instance types.
		{
			if changes.SpotInstanceTypes != nil {
				if group.Compute == nil {
					group.Compute = new(aws.Compute)
				}
				if group.Compute.InstanceTypes == nil {
					group.Compute.InstanceTypes = new(aws.InstanceTypes)
				}

				types := make([]string, len(e.SpotInstanceTypes))
				copy(types, e.SpotInstanceTypes)

				group.Compute.InstanceTypes.SetSpot(types)
				changes.SpotInstanceTypes = nil
				changed = true
			}
		}

		// Availability zones.
		{
			if changes.Subnets != nil {
				if group.Compute == nil {
					group.Compute = new(aws.Compute)
				}

				zones := make([]*aws.AvailabilityZone, len(e.Subnets))
				for i, subnet := range e.Subnets {
					zone := new(aws.AvailabilityZone)
					zone.SetName(subnet.AvailabilityZone)
					zone.SetSubnetId(subnet.ID)
					zones[i] = zone
				}

				group.Compute.SetAvailabilityZones(zones)
				changes.Subnets = nil
				changed = true
			}
		}

		// Launch specification.
		{
			// Security groups.
			{
				if changes.SecurityGroups != nil {
					if group.Compute == nil {
						group.Compute = new(aws.Compute)
					}
					if group.Compute.LaunchSpecification == nil {
						group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
					}

					securityGroupIDs := make([]string, len(e.SecurityGroups))
					for i, sg := range e.SecurityGroups {
						securityGroupIDs[i] = *sg.ID
					}

					group.Compute.LaunchSpecification.SetSecurityGroupIDs(securityGroupIDs)
					changes.SecurityGroups = nil
					changed = true
				}
			}

			// User data.
			{
				if changes.UserData != nil {
					userData, err := e.UserData.AsString()
					if err != nil {
						return err
					}

					if len(userData) > 0 {
						if group.Compute == nil {
							group.Compute = new(aws.Compute)
						}
						if group.Compute.LaunchSpecification == nil {
							group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
						}

						encoded := base64.StdEncoding.EncodeToString([]byte(userData))
						group.Compute.LaunchSpecification.SetUserData(fi.String(encoded))
						changed = true
					}

					changes.UserData = nil
				}
			}

			// Network interfaces.
			{
				if changes.AssociatePublicIP != nil {
					if group.Compute == nil {
						group.Compute = new(aws.Compute)
					}
					if group.Compute.LaunchSpecification == nil {
						group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
					}

					iface := &aws.NetworkInterface{
						Description:              fi.String("eth0"),
						DeviceIndex:              fi.Int(0),
						DeleteOnTermination:      fi.Bool(true),
						AssociatePublicIPAddress: changes.AssociatePublicIP,
					}

					group.Compute.LaunchSpecification.SetNetworkInterfaces([]*aws.NetworkInterface{iface})
					changes.AssociatePublicIP = nil
					changed = true
				}
			}

			// Root volume options.
			{
				if opts := changes.RootVolumeOpts; opts != nil {

					// Block device mappings.
					{
						if opts.Type != nil || opts.Size != nil || opts.IOPS != nil {
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
									changed = true
								}
							}
						}
					}

					// EBS optimization.
					{
						if opts.Optimization != nil {
							if group.Compute == nil {
								group.Compute = new(aws.Compute)
							}
							if group.Compute.LaunchSpecification == nil {
								group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
							}

							group.Compute.LaunchSpecification.SetEBSOptimized(e.RootVolumeOpts.Optimization)
							changed = true
						}
					}

					changes.RootVolumeOpts = nil
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
						changed = true
					}

					changes.ImageID = nil
				}
			}

			// Tags.
			{
				if changes.Tags != nil {
					if group.Compute == nil {
						group.Compute = new(aws.Compute)
					}
					if group.Compute.LaunchSpecification == nil {
						group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
					}

					group.Compute.LaunchSpecification.SetTags(e.buildTags())
					changes.Tags = nil
					changed = true
				}
			}

			// IAM instance profile.
			{
				if changes.IAMInstanceProfile != nil {
					if group.Compute == nil {
						group.Compute = new(aws.Compute)
					}
					if group.Compute.LaunchSpecification == nil {
						group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
					}

					iprof := new(aws.IAMInstanceProfile)
					iprof.SetName(e.IAMInstanceProfile.GetName())

					group.Compute.LaunchSpecification.SetIAMInstanceProfile(iprof)
					changes.IAMInstanceProfile = nil
					changed = true
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
					changed = true
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
					changed = true
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
						changed = true
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
					changed = true
				}
			}

			// Health check.
			{
				if changes.HealthCheckType != nil {
					if group.Compute == nil {
						group.Compute = new(aws.Compute)
					}
					if group.Compute.LaunchSpecification == nil {
						group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
					}

					group.Compute.LaunchSpecification.SetHealthCheckType(e.HealthCheckType)
					changes.HealthCheckType = nil
					changed = true
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
			changed = true

			// Scale up the target capacity, if needed.
			if int64(*actual.Capacity.Target) < *e.MinSize {
				group.Capacity.SetTarget(fi.Int(int(*e.MinSize)))
			}
		}
		if changes.MaxSize != nil {
			if group.Capacity == nil {
				group.Capacity = new(aws.Capacity)
			}

			group.Capacity.SetMaximum(fi.Int(int(*e.MaxSize)))
			changes.MaxSize = nil
			changed = true
		}
	}

	// Auto Scaler.
	{
		if opts := changes.AutoScalerOpts; opts != nil {
			if opts.Enabled != nil {
				autoScaler := new(aws.AutoScaleKubernetes)
				autoScaler.IsEnabled = e.AutoScalerOpts.Enabled
				autoScaler.Cooldown = e.AutoScalerOpts.Cooldown

				// Headroom.
				if headroom := opts.Headroom; headroom != nil {
					autoScaler.IsAutoConfig = fi.Bool(false)
					autoScaler.Headroom = &aws.AutoScaleHeadroom{
						CPUPerUnit:    e.AutoScalerOpts.Headroom.CPUPerUnit,
						GPUPerUnit:    e.AutoScalerOpts.Headroom.GPUPerUnit,
						MemoryPerUnit: e.AutoScalerOpts.Headroom.MemPerUnit,
						NumOfUnits:    e.AutoScalerOpts.Headroom.NumOfUnits,
					}
				} else if a.AutoScalerOpts != nil && a.AutoScalerOpts.Headroom != nil {
					autoScaler.IsAutoConfig = fi.Bool(true)
					autoScaler.SetHeadroom(nil)
				}

				// Scale down.
				if down := opts.Down; down != nil {
					autoScaler.Down = &aws.AutoScaleDown{
						MaxScaleDownPercentage: down.MaxPercentage,
						EvaluationPeriods:      down.EvaluationPeriods,
					}
				} else if a.AutoScalerOpts.Down != nil {
					autoScaler.SetDown(nil)
				}

				// Labels.
				if labels := opts.Labels; labels != nil {
					autoScaler.Labels = e.buildAutoScaleLabels(e.AutoScalerOpts.Labels)
				} else if a.AutoScalerOpts.Labels != nil {
					autoScaler.SetLabels(nil)
				}

				k8s := new(aws.KubernetesIntegration)
				k8s.SetAutoScale(autoScaler)

				integration := new(aws.Integration)
				integration.SetKubernetes(k8s)

				group.SetIntegration(integration)
				changed = true
			}

			changes.AutoScalerOpts = nil
		}
	}

	empty := &Elastigroup{}
	if !reflect.DeepEqual(empty, changes) {
		klog.Warningf("Not all changes applied to Elastigroup %q: %v", *group.ID, changes)
	}

	if !changed {
		klog.V(2).Infof("No changes detected in Elastigroup %q", *group.ID)
		return nil
	}

	klog.V(2).Infof("Updating Elastigroup %q (config: %s)", *group.ID, stringutil.Stringify(group))

	// Wrap the raw object as an Elastigroup.
	eg, err := spotinst.NewElastigroup(cloud.ProviderID(), group)
	if err != nil {
		return err
	}

	// Update the Elastigroup.
	if err := cloud.Spotinst().Elastigroup().Update(context.Background(), eg); err != nil {
		return fmt.Errorf("spotinst: failed to update elastigroup: %v", err)
	}

	return nil
}

type terraformElastigroup struct {
	Name                 *string                                 `json:"name,omitempty" cty:"name"`
	Description          *string                                 `json:"description,omitempty" cty:"description"`
	Product              *string                                 `json:"product,omitempty" cty:"product"`
	Region               *string                                 `json:"region,omitempty" cty:"region"`
	SubnetIDs            []*terraform.Literal                    `json:"subnet_ids,omitempty" cty:"subnet_ids"`
	LoadBalancers        []*terraform.Literal                    `json:"elastic_load_balancers,omitempty" cty:"elastic_load_balancers"`
	NetworkInterfaces    []*terraformElastigroupNetworkInterface `json:"network_interface,omitempty" cty:"network_interface"`
	RootBlockDevice      *terraformElastigroupBlockDevice        `json:"ebs_block_device,omitempty" cty:"ebs_block_device"`
	EphemeralBlockDevice []*terraformElastigroupBlockDevice      `json:"ephemeral_block_device,omitempty" cty:"ephemeral_block_device"`
	Integration          *terraformElastigroupIntegration        `json:"integration_kubernetes,omitempty" cty:"integration_kubernetes"`
	Tags                 []*terraformKV                          `json:"tags,omitempty" cty:"tags"`
	Lifecycle            *terraformLifecycle                     `json:"lifecycle,omitempty" cty:"lifecycle"`

	*terraformElastigroupCapacity
	*terraformElastigroupStrategy
	*terraformElastigroupInstanceTypes
	*terraformElastigroupLaunchSpec
}

type terraformElastigroupCapacity struct {
	MinSize         *int64  `json:"min_size,omitempty" cty:"min_size"`
	MaxSize         *int64  `json:"max_size,omitempty" cty:"max_size"`
	DesiredCapacity *int64  `json:"desired_capacity,omitempty" cty:"desired_capacity"`
	CapacityUnit    *string `json:"capacity_unit,omitempty" cty:"capacity_unit"`
}

type terraformElastigroupStrategy struct {
	SpotPercentage           *float64 `json:"spot_percentage,omitempty" cty:"spot_percentage"`
	Orientation              *string  `json:"orientation,omitempty" cty:"orientation"`
	FallbackToOnDemand       *bool    `json:"fallback_to_ondemand,omitempty" cty:"fallback_to_ondemand"`
	UtilizeReservedInstances *bool    `json:"utilize_reserved_instances,omitempty" cty:"utilize_reserved_instances"`
}

type terraformElastigroupInstanceTypes struct {
	OnDemand *string  `json:"instance_types_ondemand,omitempty" cty:"instance_types_ondemand"`
	Spot     []string `json:"instance_types_spot,omitempty" cty:"instance_types_spot"`
}

type terraformElastigroupLaunchSpec struct {
	Monitoring         *bool                `json:"enable_monitoring,omitempty" cty:"enable_monitoring"`
	EBSOptimized       *bool                `json:"ebs_optimized,omitempty" cty:"ebs_optimized"`
	ImageID            *string              `json:"image_id,omitempty" cty:"image_id"`
	HealthCheckType    *string              `json:"health_check_type,omitempty" cty:"health_check_type"`
	SecurityGroups     []*terraform.Literal `json:"security_groups,omitempty" cty:"security_groups"`
	UserData           *terraform.Literal   `json:"user_data,omitempty" cty:"user_data"`
	IAMInstanceProfile *terraform.Literal   `json:"iam_instance_profile,omitempty" cty:"iam_instance_profile"`
	KeyName            *terraform.Literal   `json:"key_name,omitempty" cty:"key_name"`
}

type terraformElastigroupBlockDevice struct {
	DeviceName          *string `json:"device_name,omitempty" cty:"device_name"`
	VirtualName         *string `json:"virtual_name,omitempty" cty:"virtual_name"`
	VolumeType          *string `json:"volume_type,omitempty" cty:"volume_type"`
	VolumeSize          *int64  `json:"volume_size,omitempty" cty:"volume_size"`
	DeleteOnTermination *bool   `json:"delete_on_termination,omitempty" cty:"delete_on_termination"`
}

type terraformElastigroupNetworkInterface struct {
	Description              *string `json:"description,omitempty" cty:"description"`
	DeviceIndex              *int    `json:"device_index,omitempty" cty:"device_index"`
	AssociatePublicIPAddress *bool   `json:"associate_public_ip_address,omitempty" cty:"associate_public_ip_address"`
	DeleteOnTermination      *bool   `json:"delete_on_termination,omitempty" cty:"delete_on_termination"`
}

type terraformElastigroupIntegration struct {
	IntegrationMode   *string `json:"integration_mode,omitempty" cty:"integration_mode"`
	ClusterIdentifier *string `json:"cluster_identifier,omitempty" cty:"cluster_identifier"`

	*terraformAutoScaler
}

type terraformAutoScaler struct {
	Enabled    *bool                        `json:"autoscale_is_enabled,omitempty" cty:"autoscale_is_enabled"`
	AutoConfig *bool                        `json:"autoscale_is_auto_config,omitempty" cty:"autoscale_is_auto_config"`
	Cooldown   *int                         `json:"autoscale_cooldown,omitempty" cty:"autoscale_cooldown"`
	Headroom   *terraformAutoScalerHeadroom `json:"autoscale_headroom,omitempty" cty:"autoscale_headroom"`
	Down       *terraformAutoScalerDown     `json:"autoscale_down,omitempty" cty:"autoscale_down"`
	Labels     []*terraformKV               `json:"autoscale_labels,omitempty" cty:"autoscale_labels"`
}

type terraformAutoScalerHeadroom struct {
	CPUPerUnit *int `json:"cpu_per_unit,omitempty" cty:"cpu_per_unit"`
	GPUPerUnit *int `json:"gpu_per_unit,omitempty" cty:"gpu_per_unit"`
	MemPerUnit *int `json:"memory_per_unit,omitempty" cty:"memory_per_unit"`
	NumOfUnits *int `json:"num_of_units,omitempty" cty:"num_of_units"`
}

type terraformAutoScalerDown struct {
	MaxPercentage     *int `json:"max_scale_down_percentage,omitempty" cty:"max_scale_down_percentage"`
	EvaluationPeriods *int `json:"evaluation_periods,omitempty" cty:"evaluation_periods"`
}

type terraformKV struct {
	Key   *string `json:"key" cty:"key"`
	Value *string `json:"value" cty:"value"`
}

type terraformLifecycle struct {
	IgnoreChanges []string `json:"ignore_changes,omitempty" cty:"ignore_changes"`
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
			SpotPercentage:           e.SpotPercentage,
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
	if e.ImageID != nil {
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
	if e.SecurityGroups != nil {
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
	if e.UserData != nil {
		var err error
		tf.UserData, err = t.AddFile("spotinst_elastigroup_aws", *e.Name, "user_data", e.UserData)
		if err != nil {
			return err
		}
	}

	// IAM instance profile.
	if e.IAMInstanceProfile != nil {
		tf.IAMInstanceProfile = e.IAMInstanceProfile.TerraformLink()
	}

	// Monitoring.
	if e.Monitoring != nil {
		tf.Monitoring = e.Monitoring
	}

	// Health check.
	if e.HealthCheckType != nil {
		tf.HealthCheckType = e.HealthCheckType
	}

	// SSH key.
	if e.SSHKey != nil {
		tf.KeyName = e.SSHKey.TerraformLink()
	}

	// Subnets.
	if e.Subnets != nil {
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
	if e.LoadBalancer != nil {
		tf.LoadBalancers = append(tf.LoadBalancers, e.LoadBalancer.TerraformLink())
	}

	// Public IP.
	if e.AssociatePublicIP != nil {
		tf.NetworkInterfaces = append(tf.NetworkInterfaces, &terraformElastigroupNetworkInterface{
			Description:              fi.String("eth0"),
			DeviceIndex:              fi.Int(0),
			DeleteOnTermination:      fi.Bool(true),
			AssociatePublicIPAddress: e.AssociatePublicIP,
		})
	}

	// Root volume options.
	{
		if opts := e.RootVolumeOpts; opts != nil {

			// Block device mappings.
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

			// EBS optimization.
			{
				if opts.Optimization != nil {
					tf.EBSOptimized = opts.Optimization
				}
			}
		}
	}

	// Auto Scaler.
	{
		if opts := e.AutoScalerOpts; opts != nil {
			tf.Integration = &terraformElastigroupIntegration{
				IntegrationMode:   fi.String("pod"),
				ClusterIdentifier: opts.ClusterID,
			}

			if opts.Enabled != nil {
				tf.Integration.terraformAutoScaler = &terraformAutoScaler{
					Enabled:    opts.Enabled,
					AutoConfig: fi.Bool(true),
					Cooldown:   opts.Cooldown,
				}

				// Headroom.
				if headroom := opts.Headroom; headroom != nil {
					tf.Integration.AutoConfig = fi.Bool(false)
					tf.Integration.Headroom = &terraformAutoScalerHeadroom{
						CPUPerUnit: headroom.CPUPerUnit,
						GPUPerUnit: headroom.GPUPerUnit,
						MemPerUnit: headroom.MemPerUnit,
						NumOfUnits: headroom.NumOfUnits,
					}
				}

				// Scale down.
				if down := opts.Down; down != nil {
					tf.Integration.Down = &terraformAutoScalerDown{
						MaxPercentage:     down.MaxPercentage,
						EvaluationPeriods: down.EvaluationPeriods,
					}
				}

				// Labels.
				if labels := opts.Labels; labels != nil {
					tf.Integration.Labels = make([]*terraformKV, 0, len(labels))
					for k, v := range labels {
						tf.Integration.Labels = append(tf.Integration.Labels, &terraformKV{
							Key:   fi.String(k),
							Value: fi.String(v),
						})
					}
				}

				// Ignore capacity changes because the auto scaler updates the
				// desired capacity overtime.
				if fi.BoolValue(tf.Integration.Enabled) {
					tf.Lifecycle = &terraformLifecycle{
						IgnoreChanges: []string{
							"desired_capacity",
						},
					}
				}
			}
		}
	}

	// Tags.
	{
		if e.Tags != nil {
			tags := e.buildTags()
			for _, tag := range tags {
				tf.Tags = append(tf.Tags, &terraformKV{
					Key:   tag.Key,
					Value: tag.Value,
				})
			}
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
		EbsVolumeSize:          fi.Int64(int64(fi.Int32Value(e.RootVolumeOpts.Size))),
		EbsVolumeType:          e.RootVolumeOpts.Type,
	}

	// The parameter IOPS is not supported for gp2 volumes.
	if e.RootVolumeOpts.IOPS != nil && fi.StringValue(e.RootVolumeOpts.Type) != "gp2" {
		rootDeviceMapping.EbsVolumeIops = fi.Int64(int64(fi.Int32Value(e.RootVolumeOpts.IOPS)))
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
		if i.EbsVolumeIops != nil && fi.StringValue(i.EbsVolumeType) != "gp2" {
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

	if e.Monitoring == nil {
		e.Monitoring = fi.Bool(false)
	}

	if e.HealthCheckType == nil {
		e.HealthCheckType = fi.String("K8S_NODE")
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
