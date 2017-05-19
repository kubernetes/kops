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

import (
	"fmt"
	"sync"
	"time"

	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/validation"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kubernetes/pkg/kubectl/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

const (
	PRE_CREATE = "pre-create"
	CREATE     = "create"
	ASG_CREATE = "asg" // TODO what is a better more cloud generic term?
)

var AlgorithmTypes = sets.NewString(PRE_CREATE, CREATE, ASG_CREATE)

// RollingUpdateCluster is a struct containing cluster information for a rolling update.
type RollingUpdateCluster struct {
	Cloud fi.Cloud

	MasterInterval  time.Duration
	NodeInterval    time.Duration
	BastionInterval time.Duration

	Force bool

	K8sClient        kubernetes.Interface
	ClientConfig     clientcmd.ClientConfig
	FailOnDrainError bool
	FailOnValidate   bool
	Algorithm        string
	CloudOnly        bool
	ClusterName      string
	ValidateRetries  int
	DrainInterval    time.Duration

	Cluster   *api.Cluster
	Clientset simple.Clientset
}

// FindCloudInstanceGroups joins data from the cloud and the instance groups into a map that can be used for updates.
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
				glog.Warningf("Found ASG with no corresponding instance group %q", name)
			}
			continue
		}
		group := buildCloudInstanceGroup(instancegroup, asg, nodeMap)
		groups[instancegroup.ObjectMeta.Name] = group
	}

	return groups, nil
}

// RollingUpdate performs a rolling update on a K8s Cluster.
func (c *RollingUpdateCluster) RollingUpdate(groups map[string]*CloudInstanceGroup, instanceGroups *api.InstanceGroupList) error {
	if len(groups) == 0 {
		glog.Infof("Cloud Instance Group length is zero. Not doing a rolling-update.")
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
				results[k] = fmt.Errorf("function panic bastions")
				resultsMutex.Unlock()

				defer wg.Done()

				err := group.RollingUpdate(c, instanceGroups, true, c.BastionInterval)

				resultsMutex.Lock()
				results[k] = err
				resultsMutex.Unlock()
			}(k, bastionGroup)
		}

		wg.Wait()
	}

	// TODO do we need a go func() here?
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
				results[k] = fmt.Errorf("function panic masters")
				resultsMutex.Unlock()
			}

			defer wg.Done()

			for k, group := range masterGroups {
				err := group.RollingUpdate(c, instanceGroups, false, c.MasterInterval)

				resultsMutex.Lock()
				results[k] = err
				resultsMutex.Unlock()

				// TODO: Bail on error?
			}
		}()

		wg.Wait()
	}

	// TODO - Do we want a WaitGroup on this?  I am not sure why we have wait groups and
	// TODO - go func() here?
	if c.Algorithm == PRE_CREATE && featureflag.DrainAndValidateRollingUpdate.Enabled() {
		return c.RollingUpdateNodesPreCreate(nodeGroups)
	} else {
		var wg sync.WaitGroup

		// We run nodes in series, even if they are in separate instance groups
		// typically they will not being separate instance groups. If you roll the nodes in parallel
		// you can get into a scenario where you can evict multiple statefulset pods from the same
		// statefulset at the same time. Further improvements needs to be made to protect from this as
		// well.

		wg.Add(1)

		go func() {
			for k := range nodeGroups {
				resultsMutex.Lock()
				results[k] = fmt.Errorf("function panic nodes")
				resultsMutex.Unlock()
			}

			defer wg.Done()
			for k, group := range nodeGroups {
				var err error
				if c.Algorithm == CREATE && featureflag.DrainAndValidateRollingUpdate.Enabled() {
					err = c.RollingUpdateNodesCreate(group)
				} else {
					err = group.RollingUpdate(c, instanceGroups, false, c.NodeInterval)
				}

				resultsMutex.Lock()
				results[k] = err
				resultsMutex.Unlock()

				// TODO: Bail on error?
			}
		}()

		wg.Wait()

		for _, err := range results {
			if err != nil {
				return err
			}
		}
	}

	glog.Infof("Rolling update completed!")
	return nil
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

// CloudInstanceGroupInstance describes an instance in an autoscaling group.
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

func (c *RollingUpdateCluster) updateCluster() error {
	applyCmd := &cloudup.ApplyClusterCmd{
		Cluster:         c.Cluster,
		Models:          nil,
		Clientset:       c.Clientset,
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

// RollingUpdateNodesPreCreate iterates throw each instance group.  First a new instance groups is created
// and the old nodes are cardoned.  Next the old nodes are drained and the old instance groups is deleted.
func (c *RollingUpdateCluster) RollingUpdateNodesCreate(group *CloudInstanceGroup) error {

	// Figure out which CloudInstanceGroups need updating and create a new instancegroup for each
	update := group.NeedUpdate
	if c.Force {
		update = append(update, group.Ready...)
	}

	if len(update) == 0 {
		return nil
	}

	if _, err := c.createInstanceGroup(group.InstanceGroup, getSuffix()); err != nil {
		return fmt.Errorf("unable to create new instance group: %v", err)
	}

	if err := c.updateCluster(); err != nil {
		return fmt.Errorf("unable to update cluster: %v", err)
	}

	time.Sleep(c.NodeInterval)

	// get the new list of ig and validate cluster
	if err := c.getIGAndValidateCluster(); err != nil {
		return fmt.Errorf("unable to validate cluster: %v", err)
	}

	if !c.CloudOnly {
		// cardon the nodes
		for _, u := range update {
			if err := group.CardonNode(u, c); err != nil {
				return fmt.Errorf("unable to cardon node: %v", err)
			}
		}
	}

	// Drain the nodes
	// TODO move Drain code into delete code
	if !c.CloudOnly {
		for _, u := range update {
			// TODO handle pod disruption budgets
			if err := group.DrainNode(u, c, false); err != nil {
				if c.FailOnDrainError {
					return fmt.Errorf("Failed to drain node %q: %v", u.Node.Name, err)
				}
				glog.Infof("Ignoring error draining node %q: %v", u.Node.Name, err)
			}

		}
	}

	return c.deleteInstanceGroups(group)
}

// RollingUpdateNodesPreCreate create all new nodes instance group(s) then cardons all nodes.
// Old nodes are then drained and the old instance group(s) is deleted.
func (c *RollingUpdateCluster) RollingUpdateNodesPreCreate(nodeGroups map[string]*CloudInstanceGroup) error {

	nodeGroupsUpdate := make([]*CloudInstanceGroup, 0)
	instanceGroupsNew := make([]*api.InstanceGroup, 0)

	// Figure out which CloudInstanceGroups need updating and create a new instancegroup for each
	{
		suffix := getSuffix()
		for _, group := range nodeGroups {
			update := group.NeedUpdate
			if c.Force {
				update = append(update, group.Ready...)
			}

			if len(update) == 0 {
				return nil
			}

			ig, err := c.createInstanceGroup(group.InstanceGroup, suffix)
			if err != nil {
				return fmt.Errorf("unable to create instance group: %v", err)
			}

			nodeGroupsUpdate = append(nodeGroupsUpdate, group)
			instanceGroupsNew = append(instanceGroupsNew, ig)
			glog.Infof("Creating Replacement Instance Group, %q, based on Instance Group %q.", ig.Name, group.InstanceGroup.Name)
		}
	}

	if err := c.updateCluster(); err != nil {
		return fmt.Errorf("unable to update cluster: %v", err)
	}

	glog.Info("Waiting for new Instance Group(s) to start")
	time.Sleep(c.NodeInterval)

	// get the new list of ig and validate cluster
	if err := c.getIGAndValidateCluster(); err != nil {
		return fmt.Errorf("unable to validate cluster: %v", err)
	}

	if !c.CloudOnly {
		// cardon the nodes
		glog.Infof("Cordoning all nodes")
		for _, group := range nodeGroupsUpdate {
			for _, u := range group.NeedUpdate {
				if err := group.CardonNode(u, c); err != nil {
					return fmt.Errorf("unable to cardon node: %v", err)
				}
				glog.Infof("Cordoned %q", u.Node.Name)
			}
		}
	}

	for _, group := range nodeGroupsUpdate {

		// Drain the nodes
		// TODO move Drain code into delete code
		if !c.CloudOnly {
			for _, u := range group.NeedUpdate {
				// TODO handle pod disruption budgets
				if err := group.DrainNode(u, c, false); err != nil {
					if c.FailOnDrainError {
						return fmt.Errorf("Failed to drain node %q: %v", u.Node.Name, err)
					}
					glog.Infof("Ignoring error draining node %q: %v", u.Node.Name, err)
					continue
				}

				glog.Infof("Drained node %q", u.Node.Name)
			}
		}

		if err := c.deleteInstanceGroups(group); err != nil {
			return err
		}

		glog.Infof("Deleted old Instance Group: %q", group.InstanceGroup.Name)
	}

	glog.Infof("Nodes rolling-update completed")

	return nil
}

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

	c := rollingUpdateData.Cloud.(awsup.AWSCloud)

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
			}
			glog.V(2).Infof("Ignoring cluster validation error: %v", err)
			glog.Infof("Cluster validation failed, but proceeding since fail-on-validate-error is set to false")
		}
	}

	for _, u := range update {

		instanceId := aws.StringValue(u.ASGInstance.InstanceId)

		nodeName := ""
		if u.Node != nil {
			nodeName = u.Node.Name
		}

		if isBastion {

			if err = n.DeleteAWSInstance(u, instanceId, nodeName, c); err != nil {
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

				if err = n.DrainNode(u, rollingUpdateData, false); err != nil {
					if rollingUpdateData.FailOnDrainError {
						return fmt.Errorf("Failed to drain node %q: %v", nodeName, err)
					}
					glog.Infof("Ignoring error draining node %q: %v", nodeName, err)
				}
			} else {
				glog.Warningf("Skipping drain of instance %q, because it is not registered in kubernetes", instanceId)
			}
		}

		if err = n.DeleteAWSInstance(u, instanceId, nodeName, c); err != nil {
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

			if err = rollingUpdateData.validateClusterWithRetries(instanceGroupList, t); err != nil {
				return fmt.Errorf("error validating cluster after removing a node: %v", err)
			}
		}
	}

	return nil
}

func (c *RollingUpdateCluster) getIGAndValidateCluster() error {
	// get the new list of ig
	list, err := c.Clientset.InstanceGroups(c.ClusterName).List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("unable to get instance groups: %v", err)
	}

	// validate the cluster
	if err := c.validateClusterWithRetries(list, c.NodeInterval); err != nil {
		return fmt.Errorf("unable to validate cluster: %v", err)
	}
	return nil
}

func (c *RollingUpdateCluster) deleteInstanceGroups(group *CloudInstanceGroup) error {
	// Delete the cloud group
	if err := group.Delete(c.Cloud); err != nil {
		return fmt.Errorf("Failed to delete cloud group %q: %v", group.ASGName, err)
	}

	// Delete the instance group
	if err := c.Clientset.InstanceGroups(c.ClusterName).Delete(group.InstanceGroup.Name, &metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("Failed to delete instance group %q: %v", group.InstanceGroup.Name, err)
	}

	return nil
}

func getSuffix() string {
	t := time.Now()
	return " rolling-update " + t.Format("2006-01-02-15:04:05")
}

func (c *RollingUpdateCluster) createInstanceGroup(orig *api.InstanceGroup, suffix string) (*api.InstanceGroup, error) {
	obj, err := conversion.NewCloner().DeepCopy(orig)

	if err != nil {
		return nil, fmt.Errorf("unable to clone instance group: %v", err)
	}
	ig, ok := obj.(*api.InstanceGroup)

	if !ok {
		return nil, fmt.Errorf("unexpected object type: %T", obj)
	}

	// DeepCopy does not get the maps need to copy those through
	ig.ObjectMeta.Labels = orig.ObjectMeta.Labels
	ig.ObjectMeta.Annotations = orig.ObjectMeta.Annotations
	ig.Spec.CloudLabels = orig.Spec.CloudLabels
	ig.Spec.NodeLabels = orig.Spec.NodeLabels
	ig.ObjectMeta.Name = orig.ObjectMeta.Name + suffix

	if err := cloudup.CreateInstanceGroup(ig, c.ClusterName, c.Clientset); err != nil {
		return nil, fmt.Errorf("unable to create new instance group: %v", err)
	}

	glog.V(4).Infof("adding instance group: %+v", ig)
	glog.V(4).Infof("based on instance group: %+v", orig)

	return ig, nil
}

// ValidateClusterWithRetries runs our validation methods on the K8s Cluster x times and then fails.
func (c *RollingUpdateCluster) validateClusterWithRetries(instanceGroupList *api.InstanceGroupList, t time.Duration) (err error) {

	for i := 0; i <= c.ValidateRetries; i++ {

		if _, err = validation.ValidateCluster(c.ClusterName, instanceGroupList, c.K8sClient); err != nil {
			glog.Infof("Cluster did not validate, and waiting longer: %v.", err)
			time.Sleep(t / 2)
		} else {
			glog.Infof("Cluster validated.")
			return nil
		}

	}

	if c.FailOnValidate {
		return fmt.Errorf("cluster validation failed: %v", err)
	}

	glog.Warningf("Cluster validation failed after removing instance, proceeding since fail-on-validate is set to false: %v", err)

	return nil
}

// ValidateCluster runs our validation methods on the K8s Cluster.
func (n *CloudInstanceGroup) ValidateCluster(rollingUpdateData *RollingUpdateCluster, instanceGroupList *api.InstanceGroupList) error {

	if _, err := validation.ValidateCluster(rollingUpdateData.ClusterName, instanceGroupList, rollingUpdateData.K8sClient); err != nil {
		return fmt.Errorf("cluster %q did not pass validation: %v", rollingUpdateData.ClusterName, err)
	}

	return nil

}

// DeleteAWSInstance deletes an EC2 AWS Instance.
func (n *CloudInstanceGroup) DeleteAWSInstance(u *CloudInstanceGroupInstance, instanceId string, nodeName string, c awsup.AWSCloud) error {

	if nodeName != "" {
		glog.Infof("Stopping instance %q, node %q, in AWS ASG %q.", instanceId, nodeName, n.ASGName)
	} else {
		glog.Infof("Stopping instance %q, in AWS ASG %q.", instanceId, n.ASGName)
	}

	request := &autoscaling.TerminateInstanceInAutoScalingGroupInput{
		InstanceId:                     u.ASGInstance.InstanceId,
		ShouldDecrementDesiredCapacity: aws.Bool(false),
	}

	if _, err := c.Autoscaling().TerminateInstanceInAutoScalingGroup(request); err != nil {
		if nodeName != "" {
			return fmt.Errorf("error deleting instance %q, node %q: %v", instanceId, nodeName, err)
		}
		return fmt.Errorf("error deleting instance %q: %v", instanceId, err)
	}

	return nil

}

// DrainNode drains a K8s node.
func (n *CloudInstanceGroup) DrainNode(u *CloudInstanceGroupInstance, rollingUpdateData *RollingUpdateCluster, cardon bool) error {
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
	if err := options.SetupDrain(cmd, args); err != nil {
		return fmt.Errorf("error setting up drain: %v", err)
	}

	if cardon {
		if err := options.RunCordonOrUncordon(true); err != nil {
			return fmt.Errorf("error cordoning node node: %v", err)
		}
	}

	if err := options.RunDrain(); err != nil {
		return fmt.Errorf("error draining node: %v", err)
	}

	if rollingUpdateData.DrainInterval > time.Second*0 {
		glog.V(3).Infof("Waiting for %s for pods to stabilize after draining.", rollingUpdateData.DrainInterval)
		time.Sleep(rollingUpdateData.DrainInterval)
	}

	return nil
}

// CardonNode cardons a K8s node.
func (n *CloudInstanceGroup) CardonNode(u *CloudInstanceGroupInstance, rollingUpdateData *RollingUpdateCluster) error {
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
	if err := options.SetupDrain(cmd, args); err != nil {
		return fmt.Errorf("error setting up drain: %v", err)
	}

	if err := options.RunCordonOrUncordon(true); err != nil {
		return fmt.Errorf("error cordoning node node: %v", err)
	}
	return nil
}

func (g *CloudInstanceGroup) Delete(cloud fi.Cloud) error {
	// TODO add code for aws, gce, vsphere
	c := cloud.(awsup.AWSCloud)

	// TODO: Graceful? Use cordon and drain?

	// Delete ASG
	{
		asgName := aws.StringValue(g.asg.AutoScalingGroupName)
		glog.V(2).Infof("Deleting autoscaling group %q", asgName)
		request := &autoscaling.DeleteAutoScalingGroupInput{
			AutoScalingGroupName: g.asg.AutoScalingGroupName,
			ForceDelete:          aws.Bool(true),
		}
		if _, err := c.Autoscaling().DeleteAutoScalingGroup(request); err != nil {
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
		if _, err := c.Autoscaling().DeleteLaunchConfiguration(request); err != nil {
			return fmt.Errorf("error deleting autoscaling launch configuration %q: %v", lcName, err)
		}
	}

	return nil
}

func (n *CloudInstanceGroup) String() string {
	return "CloudInstanceGroup:" + n.ASGName
}
