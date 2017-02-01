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

package kutil

// TODO move this business logic into a service than can be called via the api

import (
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	api "k8s.io/kops/pkg/apis/kops"
	validate "k8s.io/kops/pkg/validation"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kubernetes/pkg/api/v1"
	k8s_clientset "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
)

// RollingUpdateCluster restarts cluster nodes
type RollingUpdateCluster struct {
	Cloud fi.Cloud

	MasterInterval  time.Duration
	NodeInterval    time.Duration
	BastionInterval time.Duration
	K8sClient       *k8s_clientset.Clientset

	ForceDrain     bool
	FailOnValidate bool

	Force bool

	CloudOnly   bool
	ClusterName string
}

// RollingUpdateData is used to pass information to perform a rolling update
type RollingUpdateData struct {
	Cloud             fi.Cloud
	Force             bool
	Interval          time.Duration
	InstanceGroupList *api.InstanceGroupList
	IsBastion         bool

	K8sClient *k8s_clientset.Clientset

	ForceDrain     bool
	FailOnValidate bool

	CloudOnly   bool
	ClusterName string
}

// TODO move retries to RollingUpdateCluster
const retries = 8

// Find CloudInstanceGroups
func FindCloudInstanceGroups(cloud fi.Cloud, cluster *api.Cluster, instancegroups []*api.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*CloudInstanceGroup, error) {
	awsCloud := cloud.(awsup.AWSCloud)

	groups := make(map[string]*CloudInstanceGroup)

	tags := awsCloud.Tags()

	asgs, err := findAutoscalingGroups(awsCloud, tags)
	if err != nil {
		return nil, err
	}

	nodeMap := make(map[string]*v1.Node)
	for i := range nodes {
		node := &nodes[i]
		awsID := node.Spec.ExternalID
		nodeMap[awsID] = node
	}

	for _, asg := range asgs {
		name := aws.StringValue(asg.AutoScalingGroupName)
		var instancegroup *api.InstanceGroup
		for _, g := range instancegroups {
			var asgName string
			switch g.Spec.Role {
			case api.InstanceGroupRoleMaster:
				asgName = g.ObjectMeta.Name + ".masters." + cluster.ObjectMeta.Name
			case api.InstanceGroupRoleNode:
				asgName = g.ObjectMeta.Name + "." + cluster.ObjectMeta.Name
			case api.InstanceGroupRoleBastion:
				asgName = g.ObjectMeta.Name + "." + cluster.ObjectMeta.Name
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
		group := buildCloudInstanceGroup(instancegroup, asg, nodeMap)
		groups[instancegroup.ObjectMeta.Name] = group
	}

	return groups, nil
}

// TODO: should we check to see if api updates exist in the cluster
// TODO: for instance should we check if Petsets exist when upgrading 1.4.x -> 1.5.x

// Perform a rolling update on a K8s Cluster
func (c *RollingUpdateCluster) RollingUpdate(groups map[string]*CloudInstanceGroup, instanceGroups *api.InstanceGroupList) error {
	if len(groups) == 0 {
		return nil
	}

	var resultsMutex sync.Mutex
	results := make(map[string]error)

	masterGroups := make(map[string]*CloudInstanceGroup)
	nodeGroups := make(map[string]*CloudInstanceGroup)
	bastionGroups := make(map[string]*CloudInstanceGroup)
	for k, group := range groups {
		switch group.InstanceGroup.Spec.Role {
		case api.InstanceGroupRoleNode:
			nodeGroups[k] = group
		case api.InstanceGroupRoleMaster:
			masterGroups[k] = group
		case api.InstanceGroupRoleBastion:
			bastionGroups[k] = group
		default:
			return fmt.Errorf("unknown group type for group %q", group.InstanceGroup.ObjectMeta.Name)
		}
	}

	// Upgrade bastions first; if these go down we can't see anything
	{
		var wg sync.WaitGroup

		for k, bastionGroup := range bastionGroups {
			wg.Add(1)
			go func(k string, group *CloudInstanceGroup) {
				resultsMutex.Lock()
				results[k] = fmt.Errorf("function panic")
				resultsMutex.Unlock()

				defer wg.Done()

				rollingUpdateData := c.CreateRollingUpdateData(instanceGroups, true)

				err := group.RollingUpdate(rollingUpdateData)

				resultsMutex.Lock()
				results[k] = err
				resultsMutex.Unlock()
			}(k, bastionGroup)
		}

		wg.Wait()
	}

	// Upgrade master next
	{
		var wg sync.WaitGroup

		// We run master nodes in series, even if they are in separate instance groups
		// typically they will be in separate instance groups, so we can force the zones,
		// and we don't want to roll all the masters at the same time.  See issue #284
		wg.Add(1)

		go func() {
			for k := range masterGroups {
				resultsMutex.Lock()
				results[k] = fmt.Errorf("function panic")
				resultsMutex.Unlock()
			}

			defer wg.Done()

			for k, group := range masterGroups {
				rollingUpdateData := &RollingUpdateData{
					Cloud:             c.Cloud,
					Force:             c.Force,
					Interval:          c.MasterInterval,
					InstanceGroupList: instanceGroups,
					IsBastion:         false,
					K8sClient:         c.K8sClient,
					FailOnValidate:    c.FailOnValidate,
					ForceDrain:        c.ForceDrain,
					CloudOnly:         c.CloudOnly,
					ClusterName:       c.ClusterName,
				}

				err := group.RollingUpdate(rollingUpdateData)
				resultsMutex.Lock()
				results[k] = err
				resultsMutex.Unlock()

				// TODO: Bail on error?
			}
		}()

		wg.Wait()
	}

	// Upgrade nodes, with greater parallelism
	// TODO increase each instancegroups nodes by one
	{
		var wg sync.WaitGroup

		for k, nodeGroup := range nodeGroups {
			wg.Add(1)
			go func(k string, group *CloudInstanceGroup) {
				resultsMutex.Lock()
				results[k] = fmt.Errorf("function panic")
				resultsMutex.Unlock()

				defer wg.Done()

				rollingUpdateData := c.CreateRollingUpdateData(instanceGroups, false)

				err := group.RollingUpdate(rollingUpdateData)

				resultsMutex.Lock()
				results[k] = err
				resultsMutex.Unlock()
			}(k, nodeGroup)
		}

		wg.Wait()
	}

	for _, err := range results {
		if err != nil {
			return err
		}
	}

	glog.Info("\nRolling update completed!\n")
	return nil
}

func (c *RollingUpdateCluster) CreateRollingUpdateData(instanceGroups *api.InstanceGroupList, isBastion bool) *RollingUpdateData {
	return &RollingUpdateData{
		Cloud:             c.Cloud,
		Force:             c.Force,
		Interval:          c.NodeInterval,
		InstanceGroupList: instanceGroups,
		IsBastion:         isBastion,
		K8sClient:         c.K8sClient,
		FailOnValidate:    c.FailOnValidate,
		ForceDrain:        c.ForceDrain,
		CloudOnly:         c.CloudOnly,
		ClusterName:       c.ClusterName,
	}
}

// CloudInstanceGroup is the AWS ASG backing an InstanceGroup
type CloudInstanceGroup struct {
	InstanceGroup *api.InstanceGroup
	ASGName       string
	Status        string
	Ready         []*CloudInstanceGroupInstance
	NeedUpdate    []*CloudInstanceGroupInstance

	asg *autoscaling.Group
}

// CloudInstanceGroupInstance describes an instance in an autoscaling group
type CloudInstanceGroupInstance struct {
	ASGInstance *autoscaling.Instance
	Node        *v1.Node
}

func (c *CloudInstanceGroup) MinSize() int {
	return int(aws.Int64Value(c.asg.MinSize))
}

func (c *CloudInstanceGroup) MaxSize() int {
	return int(aws.Int64Value(c.asg.MaxSize))
}

func buildCloudInstanceGroup(ig *api.InstanceGroup, g *autoscaling.Group, nodeMap map[string]*v1.Node) *CloudInstanceGroup {
	n := &CloudInstanceGroup{
		ASGName:       aws.StringValue(g.AutoScalingGroupName),
		InstanceGroup: ig,
		asg:           g,
	}

	readyLaunchConfigurationName := aws.StringValue(g.LaunchConfigurationName)

	for _, i := range g.Instances {
		c := &CloudInstanceGroupInstance{ASGInstance: i}

		node := nodeMap[aws.StringValue(i.InstanceId)]
		if node != nil {
			c.Node = node
		}

		if readyLaunchConfigurationName == aws.StringValue(i.LaunchConfigurationName) {
			n.Ready = append(n.Ready, c)
		} else {
			n.NeedUpdate = append(n.NeedUpdate, c)
		}
	}

	if len(n.NeedUpdate) == 0 {
		n.Status = "Ready"
	} else {
		n.Status = "NeedsUpdate"
	}

	return n
}

// RollingUpdate performs a rolling update on a list of ec2 instances.
func (n *CloudInstanceGroup) RollingUpdate(rollingUpdateData *RollingUpdateData) error {

	// we should not get here, but hey I am going to check
	if rollingUpdateData == nil || rollingUpdateData.InstanceGroupList == nil || rollingUpdateData.K8sClient == nil {
		return fmt.Errorf("RollingUpdate is missing a data element: %v", rollingUpdateData)
	}

	c := rollingUpdateData.Cloud.(awsup.AWSCloud)

	update := n.NeedUpdate
	if rollingUpdateData.Force {
		update = append(update, n.Ready...)
	}

	if !rollingUpdateData.IsBastion && rollingUpdateData.FailOnValidate && !rollingUpdateData.CloudOnly {
		_, err := validate.ValidateCluster(rollingUpdateData.ClusterName, rollingUpdateData.InstanceGroupList, rollingUpdateData.K8sClient)
		if err != nil {
			return fmt.Errorf("Cluster %s does not pass validateion", rollingUpdateData.ClusterName)
		}
	}

	for _, u := range update {

		if !rollingUpdateData.IsBastion {
			drain, err := NewDrainOptions(nil, u.Node.ClusterName)

			if err != nil {
				glog.Warningf("Error creating drain: %v", err)
				if rollingUpdateData.ForceDrain == false {
					return err
				}
			} else {
				err = drain.DrainTheNode(u.Node.Name)
				if err != nil {
					glog.Warningf("setupErr: %v", err)
				}
				if rollingUpdateData.ForceDrain == false {
					return err
				}
			}
		}

		// TODO: Temporarily increase size of ASG?
		// TODO: Remove from ASG first so status is immediately updated?
		// TODO: Batch termination, like a rolling-update
		// TODO: check if an asg is running the correct number of instances

		instanceID := aws.StringValue(u.ASGInstance.InstanceId)
		glog.Infof("Stopping instance %q in AWS ASG %q", instanceID, n.ASGName)

		request := &ec2.TerminateInstancesInput{
			InstanceIds: []*string{u.ASGInstance.InstanceId},
		}
		_, err := c.EC2().TerminateInstances(request)
		if err != nil {
			return fmt.Errorf("error deleting instance %q: %v", instanceID, err)
		}

		if !rollingUpdateData.IsBastion {
			// Wait for new EC2 instances to be created
			time.Sleep(rollingUpdateData.Interval)

			// Wait until the cluster is happy
			// TODO: do we need to respect cloud only??
			for i := 0; i <= retries; i++ {

				if rollingUpdateData.CloudOnly {
					time.Sleep(rollingUpdateData.Interval)
				} else {
					_, err = validate.ValidateCluster(rollingUpdateData.ClusterName, rollingUpdateData.InstanceGroupList, rollingUpdateData.K8sClient)
					if err != nil {
						glog.Infof("Unable to validate k8s cluster: %s.", err)
						time.Sleep(rollingUpdateData.Interval / 2)
					} else {
						glog.Info("Cluster validated proceeding with next step in rolling update")
						break
					}
				}
			}

			if err != nil && rollingUpdateData.FailOnValidate && !rollingUpdateData.CloudOnly {
				return fmt.Errorf("validation timed out while performing rolling update: %v", err)
			}
		}

	}

	return nil
}

// Delete a ASG
func (g *CloudInstanceGroup) Delete(cloud fi.Cloud) error {
	c := cloud.(awsup.AWSCloud)

	// TODO: Graceful?

	// Delete ASG
	{
		asgName := aws.StringValue(g.asg.AutoScalingGroupName)
		glog.V(2).Infof("Deleting autoscaling group %q", asgName)
		request := &autoscaling.DeleteAutoScalingGroupInput{
			AutoScalingGroupName: g.asg.AutoScalingGroupName,
			ForceDelete:          aws.Bool(true),
		}
		_, err := c.Autoscaling().DeleteAutoScalingGroup(request)
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
		_, err := c.Autoscaling().DeleteLaunchConfiguration(request)
		if err != nil {
			return fmt.Errorf("error deleting autoscaling launch configuration %q: %v", lcName, err)
		}
	}

	return nil
}

func (n *CloudInstanceGroup) String() string {
	return "CloudInstanceGroup:" + n.ASGName
}
