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
	"time"

	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/client-go/pkg/api/v1"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/pkg/validation"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

// TODO add code comments

// FindCloudInstanceGroups joins data from the cloud and the instance groups into a map that can be used for updates.
func FindCloudInstanceGroups(cloud fi.Cloud, cluster *api.Cluster, igs []*api.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*CloudInstanceGroup, error) {
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
		var ig *api.InstanceGroup
		for _, g := range igs {
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
				if ig != nil {
					return nil, fmt.Errorf("Found multiple instance groups matching ASG %q", asgName)
				}
				ig = g
			}
		}
		if ig == nil {
			if warnUnmatched {
				glog.Warningf("Found ASG with no corresponding instance group %q", name)
			}
			continue
		}
		group := buildCloudInstanceGroup(ig, asg, nodeMap)
		groups[ig.ObjectMeta.Name] = group
	}

	return groups, nil
}

// DeleteInstanceGroup removes the cloud resources for an InstanceGroup
type DeleteInstanceGroup struct {
	Cluster   *api.Cluster
	Cloud     fi.Cloud
	Clientset simple.Clientset
}

func (d *DeleteInstanceGroup) DeleteInstanceGroup(group *api.InstanceGroup) error {

	if group == nil {
		return fmt.Errorf("unable to delete instance as function call with nil value")
	}

	if d.Cluster == nil {
		return fmt.Errorf("unable to delete instance as cluster is not defined")
	}
	if d.Cloud == nil {
		return fmt.Errorf("unable to delete instance as cloud is not defined")
	}

	groups, err := FindCloudInstanceGroups(d.Cloud, d.Cluster, []*api.InstanceGroup{group}, false, nil)
	cig := groups[group.ObjectMeta.Name]
	if cig == nil {
		glog.Warningf("AutoScalingGroup %q not found in cloud - skipping delete", group.ObjectMeta.Name)
	} else {
		if len(groups) != 1 {
			return fmt.Errorf("Multiple InstanceGroup resources found in cloud")
		}

		glog.Infof("Deleting AutoScalingGroup %q", group.ObjectMeta.Name)

		err = cig.Delete(d.Cloud)
		if err != nil {
			return fmt.Errorf("error deleting cloud resources for InstanceGroup: %v", err)
		}
	}

	if err = d.Clientset.InstanceGroupsFor(d.Cluster).Delete(group.ObjectMeta.Name, nil); err != nil {
		return fmt.Errorf("unable to delete instance group: %q, %v", group.ObjectMeta.Name, err)
	}

	return nil
}

const (
	KOPS_IG_PARENT = "kops.alpha.kubernetes.io/instancegroup/parent"
	KOPS_IG_CHILD  = "kops.alpha.kubernetes.io/instancegroup/child"
)

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

func (c *CloudInstanceGroup) String() string {
	return "CloudInstanceGroup:" + c.ASGName
}

func (c *CloudInstanceGroup) MinSize() int {
	return int(aws.Int64Value(c.asg.MinSize))
}

func (c *CloudInstanceGroup) MaxSize() int {
	return int(aws.Int64Value(c.asg.MaxSize))
}

// TODO: Remove from ASG first so status is immediately updated?

// RollingUpdate performs a rolling update on a list of ec2 instances.
func (c *CloudInstanceGroup) RollingUpdate(rollingUpdateData *RollingUpdateCluster, instanceGroupList *api.InstanceGroupList, isBastion bool, t time.Duration) (err error) {

	// we should not get here, but hey I am going to check.
	if rollingUpdateData == nil {
		return fmt.Errorf("rollingUpdate cannot be nil")
	}

	// Do not need a k8s client if you are doing cloud only.
	if rollingUpdateData.K8sClient == nil && !rollingUpdateData.CloudOnly {
		return fmt.Errorf("rollingUpdate is missing a k8s client")
	}

	if instanceGroupList == nil {
		return fmt.Errorf("rollingUpdate is missing the InstanceGroupList")
	}

	cloud := rollingUpdateData.Cloud.(awsup.AWSCloud)

	update := c.NeedUpdate
	if rollingUpdateData.Force {
		update = append(update, c.Ready...)
	}

	if len(update) == 0 {
		return nil
	}

	v := &validation.ValidateClusterRetries{
		Cluster:         rollingUpdateData.Cluster,
		Clientset:       rollingUpdateData.Clientset,
		Interval:        rollingUpdateData.NodeInterval,
		K8sClient:       rollingUpdateData.K8sClient,
		ValidateRetries: rollingUpdateData.ValidateRetries,
	}

	if isBastion {
		glog.V(3).Info("Not validating the cluster as instance is a bastion.")
	} else if rollingUpdateData.CloudOnly {
		glog.V(3).Info("Not validating cluster as validation is turned off via the cloud-only flag.")
	} else if featureflag.DrainAndValidateRollingUpdate.Enabled() {
		if err = v.ValidateCluster(instanceGroupList); err != nil {
			if rollingUpdateData.FailOnValidate {
				return fmt.Errorf("error validating cluster: %v", err)
			} else {
				glog.V(2).Infof("Ignoring cluster validation error: %v", err)
				glog.Infof("Cluster validation failed, but proceeding since fail-on-validate-error is set to false")
			}
		}
	}

	nodeAdapter, err := validation.NewNodeAPIAdapter(rollingUpdateData.K8sClient, rollingUpdateData.NodeInterval, rollingUpdateData.ClientConfig)

	for _, u := range update {

		instanceId := aws.StringValue(u.ASGInstance.InstanceId)

		nodeName := ""
		if u.Node != nil {
			nodeName = u.Node.Name
		}

		if isBastion {

			if err = c.DeleteAWSInstance(u, instanceId, nodeName, cloud); err != nil {
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

				if err = nodeAdapter.DrainNode(u.Node.Name, false, rollingUpdateData.DrainInterval); err != nil {
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

		if err = c.DeleteAWSInstance(u, instanceId, nodeName, cloud); err != nil {
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

			if err = v.ValidateClusterWithRetries(instanceGroupList); err != nil {

				if rollingUpdateData.FailOnValidate {
					return fmt.Errorf("error validating cluster after removing a node: %v", err)
				}

				glog.Warningf("Cluster validation failed after removing instance, proceeding since fail-on-validate is set to false: %v", err)
			}
		}
	}

	return nil
}

// RollingUpdateNodesCreate iterates throw each instance group.  First a new instance groups is created
// and the old nodes are cordoned.  Next the old nodes are drained and the old instance groups is deleted.
func (c *CloudInstanceGroup) RollingUpdateNodesCreate(r *RollingUpdateCluster) error {

	// Figure out which CloudInstanceGroups need updating and create a new instance group for each
	if r.Force {
		c.NeedUpdate = append(c.NeedUpdate, c.Ready...)
	}

	if len(c.NeedUpdate) == 0 {
		return nil
	}

	if _, ok := c.InstanceGroup.ObjectMeta.Annotations[KOPS_IG_CHILD]; !ok {

		suffix := getSuffix(c.InstanceGroup.ObjectMeta.Name)

		if _, err := c.Duplicate(r.Cluster, r.Clientset, suffix); err != nil {
			return fmt.Errorf("unable to create new instance group: %v", err)
		}
	}

	if err := updateCluster(r.Cluster, r.Clientset); err != nil {
		return fmt.Errorf("unable to update cluster: %v", err)
	}

	glog.Info("Waiting for new Instance Group to start")
	time.Sleep(r.NodeInterval)

	// get the new list of ig and validate cluster
	if err := validateCluster(r); err != nil {
		return fmt.Errorf("unable to validate cluster: %v", err)
	}

	if err := c.CordonNodes(r); err != nil {
		return fmt.Errorf("unable to cordon nodes: %v", err)
	}

	if err := c.DrainAndDelete(r); err != nil {
		return fmt.Errorf("unable to drain and delete nodes: %v", err)
	}

	// validate new nodes
	if err := validateCluster(r); err != nil {
		return fmt.Errorf("unable to validate cluster: %v", err)
	}

	return nil
}

func (c *CloudInstanceGroup) CordonNodes(r *RollingUpdateCluster) error {

	if r.CloudOnly {
		glog.Warningf("not cordoning nodes as --cloud-only is set")
	}

	nodeAdapter, _ := validation.NewNodeAPIAdapter(r.K8sClient, r.NodeInterval, r.ClientConfig)
	for _, u := range c.NeedUpdate {
		if err := nodeAdapter.CordonNode(u.Node.Name); err != nil {
			if r.FailOnDrainError {
				return fmt.Errorf("unable to cardon node: %v", err)
			}
		}
		glog.Infof("Cordoned %q", u.Node.Name)
	}

	return nil
}

func (c *CloudInstanceGroup) DrainAndDelete(r *RollingUpdateCluster) error {

	if r.CloudOnly {
		glog.Warningf("not draining nodes as --cloud-only is set")
	} else {
		nodeAdapter, _ := validation.NewNodeAPIAdapter(r.K8sClient, r.NodeInterval, r.ClientConfig)
		// Drain the nodes
		// TODO move Drain code into delete code
		for _, u := range c.NeedUpdate {
			// TODO handle pod disruption budgets
			if err := nodeAdapter.DrainNode(u.Node.Name, false, r.DrainInterval); err != nil {
				if r.FailOnDrainError {
					return fmt.Errorf("Failed to drain node %q: %v", u.Node.Name, err)
				}
				glog.Infof("Ignoring error draining node %q: %v", u.Node.Name, err)
			}

			glog.Infof("Drained node %q", u.Node.Name)
		}

	}

	d := &DeleteInstanceGroup{
		Cluster:   r.Cluster,
		Clientset: r.Clientset,
		Cloud:     r.Cloud,
	}
	return d.DeleteInstanceGroup(c.InstanceGroup)

}

// DeleteAWSInstance deletes an EC2 AWS Instance.
func (c *CloudInstanceGroup) DeleteAWSInstance(u *CloudInstanceGroupInstance, instanceId string, nodeName string, cloud awsup.AWSCloud) error {

	if nodeName != "" {
		glog.Infof("Stopping instance %q, node %q, in AWS ASG %q.", instanceId, nodeName, c.ASGName)
	} else {
		glog.Infof("Stopping instance %q, in AWS ASG %q.", instanceId, c.ASGName)
	}

	request := &autoscaling.TerminateInstanceInAutoScalingGroupInput{
		InstanceId:                     u.ASGInstance.InstanceId,
		ShouldDecrementDesiredCapacity: aws.Bool(false),
	}

	if _, err := cloud.Autoscaling().TerminateInstanceInAutoScalingGroup(request); err != nil {
		if nodeName != "" {
			return fmt.Errorf("error deleting instance %q, node %q: %v", instanceId, nodeName, err)
		}
		return fmt.Errorf("error deleting instance %q: %v", instanceId, err)
	}

	return nil

}

func (c *CloudInstanceGroup) Delete(cloud fi.Cloud) error {
	cloudInterface := cloud.(awsup.AWSCloud)

	// TODO: Graceful?

	// Delete ASG
	{
		asgName := aws.StringValue(c.asg.AutoScalingGroupName)
		glog.V(2).Infof("Deleting autoscaling group %q", asgName)
		request := &autoscaling.DeleteAutoScalingGroupInput{
			AutoScalingGroupName: c.asg.AutoScalingGroupName,
			ForceDelete:          aws.Bool(true),
		}
		_, err := cloudInterface.Autoscaling().DeleteAutoScalingGroup(request)
		if err != nil {
			return fmt.Errorf("error deleting autoscaling group %q: %v", asgName, err)
		}
	}

	// Delete LaunchConfig
	{
		lcName := aws.StringValue(c.asg.LaunchConfigurationName)
		glog.V(2).Infof("Deleting autoscaling launch configuration %q", lcName)
		request := &autoscaling.DeleteLaunchConfigurationInput{
			LaunchConfigurationName: c.asg.LaunchConfigurationName,
		}
		_, err := cloudInterface.Autoscaling().DeleteLaunchConfiguration(request)
		if err != nil {
			return fmt.Errorf("error deleting autoscaling launch configuration %q: %v", lcName, err)
		}
	}

	return nil
}

func (c *CloudInstanceGroup) Duplicate(cluster *api.Cluster, clientSet simple.Clientset, suffix string) (*api.InstanceGroup, error) {
	obj, err := conversion.NewCloner().DeepCopy(c.InstanceGroup)

	if err != nil {
		return nil, fmt.Errorf("unable to clone instance group: %v", err)
	}

	ig, ok := obj.(*api.InstanceGroup)

	if !ok {
		return nil, fmt.Errorf("unexpected object type: %T", obj)
	}

	{
		// DeepCopy does not get the maps need to copy those through
		ig.ObjectMeta.Labels = c.InstanceGroup.ObjectMeta.Labels
		ig.ObjectMeta.Annotations = c.InstanceGroup.ObjectMeta.Annotations
		ig.ObjectMeta.Name = c.InstanceGroup.ObjectMeta.Name + suffix
		ig.Spec.Role = c.InstanceGroup.Spec.Role

		if ig.ObjectMeta.Annotations == nil {
			ig.ObjectMeta.Annotations = make(map[string]string)
		}
		if c.InstanceGroup.ObjectMeta.Annotations == nil {
			c.InstanceGroup.ObjectMeta.Annotations = make(map[string]string)
		}

		ig.ObjectMeta.Annotations[KOPS_IG_PARENT] = c.InstanceGroup.ObjectMeta.Name
		c.InstanceGroup.ObjectMeta.Annotations[KOPS_IG_CHILD] = ig.ObjectMeta.Name
		ig.Spec.CloudLabels = c.InstanceGroup.Spec.CloudLabels
		ig.Spec.NodeLabels = c.InstanceGroup.Spec.NodeLabels
	}

	{
		if err := cloudup.CreateInstanceGroup(ig, cluster, clientSet); err != nil {
			return nil, fmt.Errorf("unable to create new instance group: %v", err)
		}

		if _, err := clientSet.InstanceGroupsFor(cluster).Update(ig); err != nil {
			return nil, fmt.Errorf("unable to create update instance group: %v", err)
		}
	}

	glog.V(4).Infof("adding instance group: %+v", ig)
	glog.V(4).Infof("based on instance group: %+v", c.InstanceGroup)

	return ig, nil
}

func validateCluster(r *RollingUpdateCluster) error {
	if r.CloudOnly {
		glog.Warningf("Not validating cluster as --cloud-only is set")
		return nil
	}
	v := &validation.ValidateClusterRetries{
		Cluster:         r.Cluster,
		Clientset:       r.Clientset,
		Interval:        r.NodeInterval,
		K8sClient:       r.K8sClient,
		ValidateRetries: r.ValidateRetries,
	}

	// get the new list of ig and validate cluster
	if err := v.GetInstanceGroupsAndValidateCluster(); err != nil {
		if r.FailOnValidate {
			return fmt.Errorf("unable to validate cluster: %v", err)
		}
	}

	return nil
}

func updateCluster(cluster *api.Cluster, clientset simple.Clientset) error {

	if clientset == nil {
		return fmt.Errorf("client must be set, it is nil")
	}
	if cluster == nil {
		return fmt.Errorf("cluster must be set, it is nil")
	}

	glog.V(4).Infof("cluster: %+v", cluster)
	glog.V(4).Infof("client set: %+v", clientset)

	applyCmd := &cloudup.ApplyClusterCmd{
		Cluster:         cluster,
		Models:          nil,
		Clientset:       clientset,
		TargetName:      cloudup.TargetDirect,
		OutDir:          ".",
		DryRun:          false,
		MaxTaskDuration: cloudup.DefaultMaxTaskDuration,
		InstanceGroups:  nil,
	}

	if err := applyCmd.Run(); err != nil {
		return err
	}

	return nil
}

const (
	ig_prefix    = "-rolling-update-"
	ig_ts_layout = "2006-01-02-15-04-05"
)

var igRegex = regexp.MustCompile(ig_prefix + "(\\d{4}-\\d{2}-\\d{2}-\\d{2}-\\d{2}-\\d{2}$)")

func getSuffix(oldName string) string {
	t := time.Now()
	return getSuffixWithTime(oldName, t)
}

func getSuffixWithTime(oldName string, t time.Time) string {

	timeStamp := ig_prefix + t.Format(ig_ts_layout)

	if igRegex.MatchString(oldName) {
		return igRegex.ReplaceAllString(oldName, timeStamp)
	}

	return oldName + timeStamp

}
