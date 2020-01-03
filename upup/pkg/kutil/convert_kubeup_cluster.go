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

package kutil

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"
	"k8s.io/kops"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/client/simple"
	awsresources "k8s.io/kops/pkg/resources/aws"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

// ConvertKubeupCluster performs a conversion of a cluster that was imported from kube-up
type ConvertKubeupCluster struct {
	OldClusterName string
	NewClusterName string
	Cloud          fi.Cloud

	Clientset simple.Clientset

	ClusterConfig  *kopsapi.Cluster
	InstanceGroups []*kopsapi.InstanceGroup

	// Channel is the channel that we are upgrading to
	Channel *kopsapi.Channel
}

func (x *ConvertKubeupCluster) Upgrade() error {
	awsCloud := x.Cloud.(awsup.AWSCloud)

	cluster := x.ClusterConfig

	newClusterName := x.NewClusterName
	if newClusterName == "" {
		return fmt.Errorf("NewClusterName must be specified")
	}
	oldClusterName := x.OldClusterName
	if oldClusterName == "" {
		return fmt.Errorf("OldClusterName must be specified")
	}

	oldKeyStore, err := x.Clientset.KeyStore(cluster)
	if err != nil {
		return err
	}

	oldTags := awsCloud.Tags()

	newTags := awsCloud.Tags()
	newTags["KubernetesCluster"] = newClusterName

	// Build completed cluster (force errors asap)
	cluster.ObjectMeta.Name = newClusterName

	newConfigBase, err := x.Clientset.ConfigBaseFor(cluster)
	if err != nil {
		return fmt.Errorf("error building ConfigBase for cluster: %v", err)
	}
	cluster.Spec.ConfigBase = newConfigBase.Path()

	newKeyStore, err := x.Clientset.KeyStore(cluster)
	if err != nil {
		return err
	}

	// Set KubernetesVersion from channel
	if x.Channel != nil {
		kubernetesVersion := kopsapi.RecommendedKubernetesVersion(x.Channel, kops.Version)
		if kubernetesVersion != nil {
			cluster.Spec.KubernetesVersion = kubernetesVersion.String()
		}
	}

	err = cloudup.PerformAssignments(cluster)
	if err != nil {
		return fmt.Errorf("error populating cluster defaults: %v", err)
	}

	if cluster.ObjectMeta.Annotations != nil {
		// Remove the management annotation for the new cluster
		delete(cluster.ObjectMeta.Annotations, kopsapi.AnnotationNameManagement)
	}

	assetBuilder := assets.NewAssetBuilder(cluster, "")
	fullCluster, err := cloudup.PopulateClusterSpec(x.Clientset, cluster, assetBuilder)
	if err != nil {
		return err
	}

	// Try to pre-query as much as possible before doing anything destructive
	instances, err := findInstances(awsCloud)
	if err != nil {
		return fmt.Errorf("error finding instances: %v", err)
	}

	subnets, err := awsresources.DescribeSubnets(x.Cloud)
	if err != nil {
		return fmt.Errorf("error finding subnets: %v", err)
	}

	securityGroups, err := awsresources.DescribeSecurityGroups(x.Cloud, x.OldClusterName)
	if err != nil {
		return fmt.Errorf("error finding security groups: %v", err)
	}

	volumes, err := awsresources.DescribeVolumes(x.Cloud)
	if err != nil {
		return err
	}

	dhcpOptions, err := awsresources.DescribeDhcpOptions(x.Cloud)
	if err != nil {
		return err
	}

	routeTables, err := awsresources.DescribeRouteTables(x.Cloud, oldClusterName)
	if err != nil {
		return err
	}

	autoscalingGroups, err := awsup.FindAutoscalingGroups(awsCloud, oldTags)
	if err != nil {
		return err
	}

	elbs, _, err := awsresources.DescribeELBs(x.Cloud)
	if err != nil {
		return err
	}

	// Find masters
	var masters []*ec2.Instance
	for _, instance := range instances {
		role, _ := awsup.FindEC2Tag(instance.Tags, "Role")
		if role == oldClusterName+"-master" {
			masters = append(masters, instance)
		}
	}
	if len(masters) == 0 {
		return fmt.Errorf("could not find masters")
	}

	// Stop autoscalingGroups
	for _, group := range autoscalingGroups {
		name := aws.StringValue(group.AutoScalingGroupName)
		klog.Infof("Stopping instances in autoscaling group %q", name)

		request := &autoscaling.UpdateAutoScalingGroupInput{
			AutoScalingGroupName: group.AutoScalingGroupName,
			DesiredCapacity:      aws.Int64(0),
			MinSize:              aws.Int64(0),
			MaxSize:              aws.Int64(0),
		}

		_, err := awsCloud.Autoscaling().UpdateAutoScalingGroup(request)
		if err != nil {
			return fmt.Errorf("error updating autoscaling group %q: %v", name, err)
		}
	}

	var waitStopped []string

	// Stop masters
	for _, master := range masters {
		masterInstanceID := aws.StringValue(master.InstanceId)

		masterState := aws.StringValue(master.State.Name)
		if masterState == "terminated" {
			klog.Infof("master already terminated: %q", masterInstanceID)
			continue
		}

		klog.Infof("Stopping master: %q", masterInstanceID)

		request := &ec2.StopInstancesInput{
			InstanceIds: []*string{master.InstanceId},
		}

		_, err := awsCloud.EC2().StopInstances(request)
		if err != nil {
			return fmt.Errorf("error stopping master instance: %v", err)
		}
		waitStopped = append(waitStopped, aws.StringValue(master.InstanceId))
	}

	if len(waitStopped) != 0 {
		for {
			instances, err := findInstances(awsCloud)
			if err != nil {
				return fmt.Errorf("error finding instances: %v", err)
			}

			instanceMap := make(map[string]*ec2.Instance)
			for _, i := range instances {
				instanceMap[aws.StringValue(i.InstanceId)] = i
			}

			allStopped := true
			for _, id := range waitStopped {
				instance := instanceMap[id]
				if instance != nil {
					state := aws.StringValue(instance.State.Name)
					switch state {
					case "terminated", "stopped":
						klog.Infof("instance %v no longer running (%v)", id, state)
					default:
						klog.Infof("waiting for instance %v to stop (currently %v)", id, state)
						allStopped = false
					}
				}
			}

			if allStopped {
				break
			}
			time.Sleep(5 * time.Second)
		}
	}

	// Detach volumes from masters
	for _, master := range masters {
		for _, bdm := range master.BlockDeviceMappings {
			if bdm.Ebs == nil || bdm.Ebs.VolumeId == nil {
				continue
			}
			volumeID := aws.StringValue(bdm.Ebs.VolumeId)
			masterInstanceID := aws.StringValue(master.InstanceId)
			klog.Infof("Detaching volume %q from instance %q", volumeID, masterInstanceID)

			request := &ec2.DetachVolumeInput{
				VolumeId:   bdm.Ebs.VolumeId,
				InstanceId: master.InstanceId,
			}

			for {
				_, err := awsCloud.EC2().DetachVolume(request)
				if err != nil {
					if awsup.AWSErrorCode(err) == "IncorrectState" {
						klog.Infof("will retry volume detach (master has probably not stopped yet): %q", err)
						time.Sleep(5 * time.Second)
						continue
					}
					return fmt.Errorf("error detaching volume %q from master instance %q: %v", volumeID, masterInstanceID, err)
				}
				break
			}
		}
	}

	// Retag VPC
	// We have to be careful because VPCs can be shared
	{
		vpcID := cluster.Spec.NetworkID
		retagGateway := false

		if vpcID != "" {
			tags, err := awsCloud.GetTags(vpcID)
			if err != nil {
				return fmt.Errorf("error getting VPC tags: %v", err)
			}

			clusterTag := tags[awsup.TagClusterName]
			if clusterTag != "" {
				if clusterTag != oldClusterName {
					return fmt.Errorf("VPC is tagged with a different cluster: %v", clusterTag)
				}
				replaceTags := make(map[string]string)
				replaceTags[awsup.TagClusterName] = newClusterName

				klog.Infof("Retagging VPC %q", vpcID)

				err := awsCloud.CreateTags(vpcID, replaceTags)
				if err != nil {
					return fmt.Errorf("error re-tagging VPC: %v", err)
				}

				// The VPC was tagged as ours, so make sure the gateway is consistently retagged
				retagGateway = true
			}
		}

		if retagGateway {
			gateways, err := awsresources.DescribeInternetGatewaysIgnoreTags(x.Cloud)
			if err != nil {
				return fmt.Errorf("error listing gateways: %v", err)
			}
			for _, igw := range gateways {
				match := false
				for _, a := range igw.Attachments {
					if vpcID == aws.StringValue(a.VpcId) {
						match = true
					}
				}
				if !match {
					continue
				}

				id := aws.StringValue(igw.InternetGatewayId)

				clusterTag, _ := awsup.FindEC2Tag(igw.Tags, awsup.TagClusterName)
				if clusterTag == "" || clusterTag == oldClusterName {
					replaceTags := make(map[string]string)
					replaceTags[awsup.TagClusterName] = newClusterName

					klog.Infof("Retagging InternetGateway %q", id)

					err := awsCloud.CreateTags(id, replaceTags)
					if err != nil {
						return fmt.Errorf("error re-tagging InternetGateway: %v", err)
					}
				}
			}
		}
	}

	// Retag subnets
	for _, s := range subnets {
		id := aws.StringValue(s.SubnetId)

		klog.Infof("Retagging Subnet %q", id)

		err := awsCloud.AddAWSTags(id, newTags)
		if err != nil {
			return fmt.Errorf("error re-tagging Subnet %q: %v", id, err)
		}
	}

	// Retag route tables
	for _, routeTable := range routeTables {
		id := aws.StringValue(routeTable.RouteTableId)

		clusterTag, _ := awsup.FindEC2Tag(routeTable.Tags, awsup.TagClusterName)
		if clusterTag != "" {
			if clusterTag != oldClusterName {
				return fmt.Errorf("RouteTable is tagged with a different cluster: %v", clusterTag)
			}
			replaceTags := make(map[string]string)
			replaceTags[awsup.TagClusterName] = newClusterName
			// Set the same name so we use the same route table
			// As otherwise we don't attach the route table because the subnet is considered shared
			replaceTags["Name"] = newClusterName

			klog.Infof("Retagging RouteTable %q", id)

			err := awsCloud.CreateTags(id, replaceTags)
			if err != nil {
				return fmt.Errorf("error re-tagging RouteTable: %v", err)
			}
		}
	}

	// Retag security groups
	for _, s := range securityGroups {
		id := aws.StringValue(s.GroupId)

		klog.Infof("Retagging SecurityGroup %q", id)

		err := awsCloud.AddAWSTags(id, newTags)
		if err != nil {
			return fmt.Errorf("error re-tagging SecurityGroup %q: %v", id, err)
		}
	}

	// Retag DHCP options
	// We have to be careful because DHCP options can be shared
	for _, dhcpOption := range dhcpOptions {
		id := aws.StringValue(dhcpOption.DhcpOptionsId)

		clusterTag, _ := awsup.FindEC2Tag(dhcpOption.Tags, awsup.TagClusterName)
		if clusterTag != "" {
			if clusterTag != oldClusterName {
				return fmt.Errorf("DHCP options are tagged with a different cluster: %v", clusterTag)
			}
			replaceTags := make(map[string]string)
			replaceTags[awsup.TagClusterName] = newClusterName

			klog.Infof("Retagging DHCPOptions %q", id)

			err := awsCloud.CreateTags(id, replaceTags)
			if err != nil {
				return fmt.Errorf("error re-tagging DHCP options: %v", err)
			}
		}

	}

	// Adopt LoadBalancers & LoadBalancer Security Groups
	for _, elb := range elbs {
		id := aws.StringValue(elb.LoadBalancerName)

		// TODO: Batch re-tag?
		replaceTags := make(map[string]string)
		replaceTags[awsup.TagClusterName] = newClusterName

		klog.Infof("Retagging ELB %q", id)
		err := awsCloud.CreateELBTags(id, replaceTags)
		if err != nil {
			return fmt.Errorf("error re-tagging ELB %q: %v", id, err)
		}

	}

	for _, elb := range elbs {
		for _, sg := range elb.SecurityGroups {
			id := aws.StringValue(sg)

			// TODO: Batch re-tag?
			replaceTags := make(map[string]string)
			replaceTags[awsup.TagClusterName] = newClusterName

			klog.Infof("Retagging ELB security group %q", id)
			err := awsCloud.CreateTags(id, replaceTags)
			if err != nil {
				return fmt.Errorf("error re-tagging ELB security group %q: %v", id, err)
			}
		}

	}

	// Adopt Volumes
	for _, volume := range volumes {
		id := aws.StringValue(volume.VolumeId)

		// TODO: Batch re-tag?
		replaceTags := make(map[string]string)
		replaceTags[awsup.TagClusterName] = newClusterName

		name, _ := awsup.FindEC2Tag(volume.Tags, "Name")
		if name == oldClusterName+"-master-pd" {
			klog.Infof("Found master volume %q: %s", id, name)

			az := aws.StringValue(volume.AvailabilityZone)
			replaceTags["Name"] = az + ".etcd-main." + newClusterName
		}
		klog.Infof("Retagging volume %q", id)
		err := awsCloud.CreateTags(id, replaceTags)
		if err != nil {
			return fmt.Errorf("error re-tagging volume %q: %v", id, err)
		}
	}

	err = registry.CreateClusterConfig(x.Clientset, cluster, x.InstanceGroups)
	if err != nil {
		return fmt.Errorf("error writing updated configuration: %v", err)
	}

	// TODO: No longer needed?
	err = registry.WriteConfigDeprecated(cluster, newConfigBase.Join(registry.PathClusterCompleted), fullCluster)
	if err != nil {
		return fmt.Errorf("error writing completed cluster spec: %v", err)
	}

	if err != nil {
		return fmt.Errorf("error writing completed cluster spec: %v", err)
	}

	oldCACertPool, err := oldKeyStore.CertificatePool(fi.CertificateId_CA, true)
	if err != nil {
		return fmt.Errorf("error reading old CA certs: %v", err)
	}
	for _, ca := range oldCACertPool.Secondary {
		err := newKeyStore.AddCert(fi.CertificateId_CA, ca)
		if err != nil {
			return fmt.Errorf("error importing old CA certs: %v", err)
		}
	}
	if oldCACertPool.Primary != nil {
		err := newKeyStore.AddCert(fi.CertificateId_CA, oldCACertPool.Primary)
		if err != nil {
			return fmt.Errorf("error importing old CA certs: %v", err)
		}
	}

	return nil
}
