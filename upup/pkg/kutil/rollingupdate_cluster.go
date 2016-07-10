package kutil

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"sync"
	"time"
)

// RollingUpdateCluster restarts cluster nodes
type RollingUpdateCluster struct {
	Cloud fi.Cloud
}

func FindCloudInstanceGroups(cloud fi.Cloud, cluster *api.Cluster, instancegroups []*api.InstanceGroup, warnUnmatched bool) (map[string]*CloudInstanceGroup, error) {
	awsCloud := cloud.(*awsup.AWSCloud)

	groups := make(map[string]*CloudInstanceGroup)

	tags := awsCloud.Tags()

	asgs, err := findAutoscalingGroups(awsCloud, tags)
	if err != nil {
		return nil, err
	}

	for _, asg := range asgs {
		name := aws.StringValue(asg.AutoScalingGroupName)
		var instancegroup *api.InstanceGroup
		for _, g := range instancegroups {
			var asgName string
			switch g.Spec.Role {
			case api.InstanceGroupRoleMaster:
				asgName = g.Name + ".masters." + cluster.Name
			case api.InstanceGroupRoleNode:
				asgName = g.Name + "." + cluster.Name
			default:
				glog.Warningf("Ignoring InstanceGroup of unknown role %q", g.Spec.Role)
				continue
			}

			if name == asgName {
				if instancegroup != nil {
					return nil, fmt.Errorf("Found multiple instance groups matching ASG %q", asgName)
				}
				instancegroup = g
			}
		}
		if instancegroup == nil {
			if warnUnmatched {
				glog.Warningf("Found ASG with no corresponding instance group: %q", name)
			}
			continue
		}
		group := buildCloudInstanceGroup(instancegroup, asg)
		groups[instancegroup.Name] = group
	}

	return groups, nil
}

func (c *RollingUpdateCluster) RollingUpdate(groups map[string]*CloudInstanceGroup) error {
	if len(groups) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	var resultsMutex sync.Mutex
	results := make(map[string]error)

	for k, group := range groups {
		wg.Add(1)
		go func(k string, group *CloudInstanceGroup) {
			resultsMutex.Lock()
			results[k] = fmt.Errorf("function panic")
			resultsMutex.Unlock()

			defer wg.Done()
			err := group.RollingUpdate(c.Cloud)

			resultsMutex.Lock()
			results[k] = err
			resultsMutex.Unlock()
		}(k, group)
	}

	wg.Wait()

	for _, err := range results {
		if err != nil {
			return err
		}
	}

	return nil
}

// CloudInstanceGroup is the AWS ASG backing an InstanceGroup
type CloudInstanceGroup struct {
	InstanceGroup *api.InstanceGroup
	ASGName       string
	Status        string
	Ready         []*autoscaling.Instance
	NeedUpdate    []*autoscaling.Instance

	asg *autoscaling.Group
}

func (c *CloudInstanceGroup) MinSize() int {
	return int(aws.Int64Value(c.asg.MinSize))
}

func (c *CloudInstanceGroup) MaxSize() int {
	return int(aws.Int64Value(c.asg.MaxSize))
}

func buildCloudInstanceGroup(ig *api.InstanceGroup, g *autoscaling.Group) *CloudInstanceGroup {
	n := &CloudInstanceGroup{
		ASGName:       aws.StringValue(g.AutoScalingGroupName),
		InstanceGroup: ig,
		asg:           g,
	}

	findLaunchConfigurationName := aws.StringValue(g.LaunchConfigurationName)

	for _, i := range g.Instances {
		if findLaunchConfigurationName == aws.StringValue(i.LaunchConfigurationName) {
			n.Ready = append(n.Ready, i)
		} else {
			n.NeedUpdate = append(n.NeedUpdate, i)
		}
	}

	if len(n.NeedUpdate) == 0 {
		n.Status = "Ready"
	} else {
		n.Status = "NeedsUpdate"
	}

	return n
}

func (n *CloudInstanceGroup) RollingUpdate(cloud fi.Cloud) error {
	c := cloud.(*awsup.AWSCloud)

	for _, i := range n.NeedUpdate {
		glog.Infof("Stopping instance %q in AWS ASG %q", *i.InstanceId, n.ASGName)

		// TODO: Evacuate through k8s first?

		// TODO: Temporarily increase size of ASG?

		// TODO: Remove from ASG first so status is immediately updated?

		// TODO: Batch termination, like a rolling-update

		request := &ec2.TerminateInstancesInput{
			InstanceIds: []*string{i.InstanceId},
		}
		_, err := c.EC2.TerminateInstances(request)
		if err != nil {
			return fmt.Errorf("error deleting instance %q: %v", i.InstanceId, err)
		}

		// TODO: Wait for node to appear back in k8s
		time.Sleep(time.Minute)
	}

	return nil
}

func (g *CloudInstanceGroup) Delete(cloud fi.Cloud) error {
	c := cloud.(*awsup.AWSCloud)

	// TODO: Graceful?

	// Delete ASG
	{
		asgName := aws.StringValue(g.asg.AutoScalingGroupName)
		glog.V(2).Infof("Deleting autoscaling group %q", asgName)
		request := &autoscaling.DeleteAutoScalingGroupInput{
			AutoScalingGroupName: g.asg.AutoScalingGroupName,
			ForceDelete:          aws.Bool(true),
		}
		_, err := c.Autoscaling.DeleteAutoScalingGroup(request)
		if err != nil {
			return fmt.Errorf("error deleting autoscaling group %q: %v", asgName, err)
		}
	}

	// Delete LaunchConfig
	{
		lcName := aws.StringValue(g.asg.LaunchConfigurationName)
		glog.V(2).Infof("Deleting autoscaling launch configuration %q", lcName)
		request := &autoscaling.DeleteLaunchConfigurationInput{
			LaunchConfigurationName: g.asg.LaunchConfigurationName,
		}
		_, err := c.Autoscaling.DeleteLaunchConfiguration(request)
		if err != nil {
			return fmt.Errorf("error deleting autoscaling launch configuration %q: %v", lcName, err)
		}
	}

	return nil
}

func (n *CloudInstanceGroup) String() string {
	return "CloudInstanceGroup:" + n.ASGName
}
