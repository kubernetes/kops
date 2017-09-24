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

package instancegroups

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/client-go/pkg/api/v1"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/pkg/validation"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kubernetes/pkg/kubectl/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

// FindCloudInstanceGroups joins data from the cloud and the instance groups into a map that can be used for updates.
func FindCloudInstanceGroups(cloud fi.Cloud, cluster *api.Cluster, instancegroups []*api.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*CloudInstanceGroup, error) {
	awsCloud := cloud.(awsup.AWSCloud)

	groups := make(map[string]*CloudInstanceGroup)

	tags := awsCloud.Tags()

	asgs, err := resources.FindAutoscalingGroups(awsCloud, tags)
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
				glog.Warningf("Found ASG with no corresponding instance group %q", name)
			}
			continue
		}
		group := buildCloudInstanceGroup(instancegroup, asg, nodeMap)
		groups[instancegroup.ObjectMeta.Name] = group
	}

	return groups, nil
}



// CloudInstanceGroup is the AWS ASG backing an InstanceGroup.
type CloudInstanceGroup struct {
	InstanceGroup *api.InstanceGroup
	ASGName       string
	Status        string
	Ready         []*CloudInstanceGroupInstance
	NeedUpdate    []*CloudInstanceGroupInstance

	asg *autoscaling.Group
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

// CloudInstanceGroupInstance describes an instance in an autoscaling group.
type CloudInstanceGroupInstance struct {
	ASGInstance *autoscaling.Instance
	Node        *v1.Node
}

func (n *CloudInstanceGroup) String() string {
	return "CloudInstanceGroup:" + n.ASGName
}

func (c *CloudInstanceGroup) MinSize() int {
	return int(aws.Int64Value(c.asg.MinSize))
}

func (c *CloudInstanceGroup) MaxSize() int {
	return int(aws.Int64Value(c.asg.MaxSize))
}

// TODO: Temporarily increase size of ASG?
// TODO: Remove from ASG first so status is immediately updated?
// TODO: Batch termination, like a rolling-update

// RollingUpdate performs a rolling update on a list of ec2 instances.
func (n *CloudInstanceGroup) RollingUpdate(rollingUpdateData *RollingUpdateCluster, instanceGroupList *api.InstanceGroupList, isBastion bool, t time.Duration) (err error) {

	// we should not get here, but hey I am going to check.
	if rollingUpdateData == nil {
		return fmt.Errorf("rollingUpdate cannot be nil")
	}

	// Do not need a k8s client if you are doing cloudonly.
	if rollingUpdateData.K8sClient == nil && !rollingUpdateData.CloudOnly {
		return fmt.Errorf("rollingUpdate is missing a k8s client")
	}

	if instanceGroupList == nil {
		return fmt.Errorf("rollingUpdate is missing the InstanceGroupList")
	}

	c := rollingUpdateData.Cloud

	update := n.NeedUpdate
	if rollingUpdateData.Force {
		update = append(update, n.Ready...)
	}

	if len(update) == 0 {
		return nil
	}

	if isBastion {
		glog.V(3).Info("Not validating the cluster as instance is a bastion.")
	} else if rollingUpdateData.CloudOnly {
		glog.V(3).Info("Not validating cluster as validation is turned off via the cloud-only flag.")
	} else if featureflag.DrainAndValidateRollingUpdate.Enabled() {
		if err = n.ValidateCluster(rollingUpdateData, instanceGroupList); err != nil {
			if rollingUpdateData.FailOnValidate {
				return fmt.Errorf("error validating cluster: %v", err)
			} else {
				glog.V(2).Infof("Ignoring cluster validation error: %v", err)
				glog.Infof("Cluster validation failed, but proceeding since fail-on-validate-error is set to false")
			}
		}
	}

	for _, u := range update {

		instanceId := aws.StringValue(u.ASGInstance.InstanceId)

		nodeName := ""
		if u.Node != nil {
			nodeName = u.Node.Name
		}

		if isBastion {
			if err = n.DeleteInstance(u, instanceId, nodeName, c); err != nil {
				glog.Errorf("Error deleting aws instance %q: %v", instanceId, err)
				return err
			}

			glog.Infof("Deleted a bastion instance, %s, and continuing with rolling-update.", instanceId)

			continue

		} else if rollingUpdateData.CloudOnly {

			glog.Warningf("Not draining cluster nodes as 'cloudonly' flag is set.")

		} else if featureflag.DrainAndValidateRollingUpdate.Enabled() {

			if u.Node != nil {
				glog.Infof("Draining the node: %q.", nodeName)

				if err = n.DrainNode(u, rollingUpdateData); err != nil {
					if rollingUpdateData.FailOnDrainError {
						return fmt.Errorf("Failed to drain node %q: %v", nodeName, err)
					} else {
						glog.Infof("Ignoring error draining node %q: %v", nodeName, err)
					}
				}
			} else {
				glog.Warningf("Skipping drain of instance %q, because it is not registered in kubernetes", instanceId)
			}
		}

		if err = n.DeleteInstance(u, instanceId, nodeName, c); err != nil {
			glog.Errorf("Error deleting aws instance %q, node %q: %v", instanceId, nodeName, err)
			return err
		}

		// Wait for new EC2 instances to be created
		time.Sleep(t)

		if rollingUpdateData.CloudOnly {

			glog.Warningf("Not validating cluster as cloudonly flag is set.")
			continue

		} else if featureflag.DrainAndValidateRollingUpdate.Enabled() {

			glog.Infof("Validating the cluster.")

			if err = n.ValidateClusterWithDuration(rollingUpdateData, instanceGroupList, t); err != nil {

				if rollingUpdateData.FailOnValidate {
					glog.Errorf("Cluster did not validate within the set duration of %q, you can retry, and maybe extend the duration", t)
					return fmt.Errorf("error validating cluster after removing a node: %v", err)
				}

				glog.Warningf("Cluster validation failed after removing instance, proceeding since fail-on-validate is set to false: %v", err)
			}
		}
	}

	return nil
}

// ValidateClusterWithDuration runs validation.ValidateCluster until either we get positive result or the timeout expires
func (n *CloudInstanceGroup) ValidateClusterWithDuration(rollingUpdateData *RollingUpdateCluster, instanceGroupList *api.InstanceGroupList, duration time.Duration) error {
	// TODO should we expose this to the UI?
	tickDuration := 30 * time.Second
	// Try to validate cluster at least once, this will handle durations that are lower
	// than our tick time
	if n.tryValidateCluster(rollingUpdateData, instanceGroupList, duration, tickDuration) {
		return nil
	}

	timeout := time.After(duration)
	tick := time.Tick(tickDuration)
	// Keep trying until we're timed out or got a result or got an error
	for {
		select {
		case <-timeout:
			// Got a timeout fail with a timeout error
			return fmt.Errorf("cluster did not validate within a duation of %q", duration)
		case <-tick:
			// Got a tick, validate cluster
			if n.tryValidateCluster(rollingUpdateData, instanceGroupList, duration, tickDuration) {
				return nil
			}
			// ValidateCluster didn't work yet, so let's try again
			// this will exit up to the for loop
		}
	}
}

func (n *CloudInstanceGroup) tryValidateCluster(rollingUpdateData *RollingUpdateCluster, instanceGroupList *api.InstanceGroupList, duration time.Duration, tickDuration time.Duration) bool {
	if _, err := validation.ValidateCluster(rollingUpdateData.ClusterName, instanceGroupList, rollingUpdateData.K8sClient); err != nil {
		glog.Infof("Cluster did not validate, will try again in %q util duration %q expires: %v.", tickDuration, duration, err)
		return false
	} else {
		glog.Infof("Cluster validated.")
		return true
	}
}

// ValidateCluster runs our validation methods on the K8s Cluster.
func (n *CloudInstanceGroup) ValidateCluster(rollingUpdateData *RollingUpdateCluster, instanceGroupList *api.InstanceGroupList) error {

	if _, err := validation.ValidateCluster(rollingUpdateData.ClusterName, instanceGroupList, rollingUpdateData.K8sClient); err != nil {
		return fmt.Errorf("cluster %q did not pass validation: %v", rollingUpdateData.ClusterName, err)
	}

	return nil

}

// DeleteInstance deletes an Cloud Instance.
func (n *CloudInstanceGroup) DeleteInstance(u *CloudInstanceGroupInstance, instanceId string, nodeName string, c fi.Cloud) error {

	if nodeName != "" {
		glog.Infof("Stopping instance %q, node %q, in AWS ASG %q.", instanceId, nodeName, n.ASGName)
	} else {
		glog.Infof("Stopping instance %q, in AWS ASG %q.", instanceId, n.asg.AutoScalingGroupName)
	}

	if err := c.DeleteInstance(u.ASGInstance.InstanceId); err != nil {
		if nodeName != "" {
			return fmt.Errorf("error deleting instance %q, node %q: %v", instanceId, nodeName, err)
		}
		return fmt.Errorf("error deleting instance %q: %v", instanceId, err)
	}

	return nil

}

// DrainNode drains a K8s node.
func (n *CloudInstanceGroup) DrainNode(u *CloudInstanceGroupInstance, rollingUpdateData *RollingUpdateCluster) error {
	if rollingUpdateData.ClientConfig == nil {
		return fmt.Errorf("ClientConfig not set")
	}
	f := cmdutil.NewFactory(rollingUpdateData.ClientConfig)

	// TODO: Send out somewhere else, also DrainOptions has errout
	out := os.Stdout
	errOut := os.Stderr

	options := &cmd.DrainOptions{
		Factory:          f,
		Out:              out,
		IgnoreDaemonsets: true,
		Force:            true,
		DeleteLocalData:  true,
		ErrOut:           errOut,
	}

	cmd := &cobra.Command{
		Use: "cordon NODE",
	}
	args := []string{u.Node.Name}
	err := options.SetupDrain(cmd, args)
	if err != nil {
		return fmt.Errorf("error setting up drain: %v", err)
	}

	err = options.RunCordonOrUncordon(true)
	if err != nil {
		return fmt.Errorf("error cordoning node node: %v", err)
	}

	err = options.RunDrain()
	if err != nil {
		return fmt.Errorf("error draining node: %v", err)
	}

	if rollingUpdateData.DrainInterval > time.Second*0 {
		glog.V(3).Infof("Waiting for %s for pods to stabilize after draining.", rollingUpdateData.DrainInterval)
		time.Sleep(rollingUpdateData.DrainInterval)
	}

	return nil
}

// Delete and CloudInstanceGroups
func (g *CloudInstanceGroup) Delete(cloud fi.Cloud) error {

	// TODO: Leaving func in place in order to cordon nd drain nodes
	return cloud.DeleteGroup(*g.asg.AutoScalingGroupName, *g.asg.LaunchConfigurationName)
}
