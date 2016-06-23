package kutil

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
)

// UpgradeCluster performs an upgrade of a k8s cluster
type UpgradeCluster struct {
	OldClusterName string
	NewClusterName string
	Cloud          fi.Cloud

	StateStore fi.StateStore

	Config *cloudup.CloudConfig
}

func (x *UpgradeCluster) Upgrade() error {
	awsCloud := x.Cloud.(*awsup.AWSCloud)

	config := x.Config

	newClusterName := x.NewClusterName
	if newClusterName == "" {
		return fmt.Errorf("NewClusterName must be specified")
	}
	oldClusterName := x.OldClusterName
	if oldClusterName == "" {
		return fmt.Errorf("OldClusterName must be specified")
	}

	newTags := awsCloud.Tags()
	newTags["KubernetesCluster"] = newClusterName

	// Try to pre-query as much as possible before doing anything destructive
	instances, err := findInstances(awsCloud)
	if err != nil {
		return fmt.Errorf("error finding instances: %v", err)
	}

	volumes, err := DescribeVolumes(x.Cloud)
	if err != nil {
		return err
	}

	dhcpOptions, err := DescribeDhcpOptions(x.Cloud)
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

	// Stop masters
	for _, master := range masters {
		masterInstanceID := aws.StringValue(master.InstanceId)
		glog.Infof("Stopping master: %q", masterInstanceID)

		request := &ec2.StopInstancesInput{
			InstanceIds: []*string{master.InstanceId},
		}

		_, err := awsCloud.EC2.StopInstances(request)
		if err != nil {
			return fmt.Errorf("error stopping master instance: %v", err)
		}
	}

	// TODO: Wait for master to stop & detach volumes?

	//subnets, err := DescribeSubnets(x.Cloud)
	//if err != nil {
	//	return fmt.Errorf("error finding subnets: %v", err)
	//}
	//for _, s := range subnets {
	//	id := aws.StringValue(s.SubnetId)
	//	err := awsCloud.AddAWSTags(id, newTags)
	//	if err != nil {
	//		return fmt.Errorf("error re-tagging subnet %q: %v", id, err)
	//	}
	//}

	// Retag VPC
	// We have to be careful because VPCs can be shared
	{
		vpcID := config.NetworkID
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
				err := awsCloud.CreateTags(vpcID, replaceTags)
				if err != nil {
					return fmt.Errorf("error re-tagging VPC: %v", err)
				}
			}

		}
	}

	// Retag DHCP options
	// We have to be careful because DHCP options can be shared
	for _, dhcpOption := range dhcpOptions {
		clusterTag, _ := awsup.FindEC2Tag(dhcpOption.Tags, awsup.TagClusterName)
		if clusterTag != "" {
			if clusterTag != oldClusterName {
				return fmt.Errorf("DHCP options are tagged with a different cluster: %v", clusterTag)
			}
			replaceTags := make(map[string]string)
			replaceTags[awsup.TagClusterName] = newClusterName
			err := awsCloud.CreateTags(*dhcpOption.DhcpOptionsId, replaceTags)
			if err != nil {
				return fmt.Errorf("error re-tagging DHCP options: %v", err)
			}
		}

	}

	// TODO: Retag internet gateway (may not be tagged at all though...)
	// TODO: Share more code with cluste deletion?

	// TODO: Adopt LoadBalancers & LoadBalancer Security Groups

	// Adopt Volumes
	for _, volume := range volumes {
		id := aws.StringValue(volume.VolumeId)

		// TODO: Batch re-tag?
		replaceTags := make(map[string]string)
		replaceTags[awsup.TagClusterName] = newClusterName

		name, _ := awsup.FindEC2Tag(volume.Tags, "Name")
		if name == oldClusterName+"-master-pd" {
			glog.Infof("Found master volume %q: %s", id, name)
			replaceTags["Name"] = "kubernetes.master." + aws.StringValue(volume.AvailabilityZone) + "." + newClusterName
		}
		err := awsCloud.CreateTags(id, replaceTags)
		if err != nil {
			return fmt.Errorf("error re-tagging volume %q: %v", id, err)
		}
	}

	config.ClusterName = newClusterName
	err = x.StateStore.WriteConfig(config)
	if err != nil {
		return fmt.Errorf("error writing updated configuration: %v", err)
	}

	return nil
}
