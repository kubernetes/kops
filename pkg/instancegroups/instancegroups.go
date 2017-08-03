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

	"golang.org/x/net/context"
	compute "google.golang.org/api/compute/v0.beta"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/client-go/pkg/api/v1"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/validation"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/pkg/kubectl/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/aws"
	"k8s.io/kops/pkg/resources"
)

// TODO how do we refactor this?  I cannot move the code into awsup.AWSCloud as we
// TODO get a import cycle not allowed

// FindCloudInstanceGroups joins data from the cloud and the instance groups into a map that can be used for updates.
func FindCloudInstanceGroups(cloud fi.Cloud, cluster *api.Cluster, instancegroups []*api.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*CloudInstanceGroup, error) {

	groups := make(map[string]*CloudInstanceGroup)
	nodeMap := make(map[string]*v1.Node)
	for i := range nodes {
		node := &nodes[i]
		awsID := node.Spec.ExternalID
		nodeMap[awsID] = node
	}

	switch c := cloud.(type) {
	case awsup.AWSCloud:
		tags := c.Tags()

		asgs, err := resources.FindAutoscalingGroups(c, tags)
		if err != nil {
			return nil, fmt.Errorf("unable to find autoscale groups: %v", err)
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
			group := awsBuildCloudInstanceGroup(instancegroup, asg, nodeMap)
			groups[instancegroup.ObjectMeta.Name] = group
		}

	case gce.GCECloud:
		ctx := context.Background()

		instanceTemplates := make(map[string]*compute.InstanceTemplate)
		{
			templates, err := resources.FindInstanceTemplates(c,cluster.ObjectMeta.Name)
			if err != nil {
				return nil, err
			}
			for _, t := range templates {
				instanceTemplates[t.SelfLink] = t
			}
		}

		var migs []*compute.InstanceGroupManager

		// TODO we need to iterate through the instance groups rather can mig
		zones, err := c.Zones()
		if err != nil {
			return nil, err
		}
		for _, zoneName := range zones {
			err := c.Compute().InstanceGroupManagers.List(c.Project(), zoneName).Pages(ctx, func(page *compute.InstanceGroupManagerList) error {
				for _, mig := range page.Items {
					instanceTemplate := instanceTemplates[mig.InstanceTemplate]
					if instanceTemplate == nil {
						glog.V(2).Infof("Ignoring MIG with unmanaged InstanceTemplate: %s", mig.InstanceTemplate)
						continue
					}

					migs = append(migs, mig)
				}
				return nil
			})

			if err != nil {
				return nil, fmt.Errorf("error listing InstanceGroupManagers: %v", err)
			}
		}

		for _, mig := range migs {
			name := mig.Name
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
			group := gceBuildCloudInstanceGroup(instancegroup, mig, nodeMap, c)
			groups[instancegroup.ObjectMeta.Name] = group
		}

	default:
		return nil, fmt.Errorf("Cloud is not implmemented as of yet: %v", cloud)
	}

	return groups, nil
}

func gceBuildCloudInstanceGroup(ig *api.InstanceGroup, g *compute.InstanceGroupManager, nodeMap map[string]*v1.Node, cloud gce.GCECloud) *CloudInstanceGroup {
	n := &CloudInstanceGroup{
		GroupName:     g.Name,
		InstanceGroup: ig,
		GroupTemplateName: g.InstanceTemplate,
	}

	// TODO FIXME
	//readyLaunchConfigurationName := g.InstanceTemplate
	// This call is not paginated
	instances, _ := cloud.Compute().InstanceGroupManagers.ListManagedInstances(cloud.Project(), cloud.Region(), g.Name).Do()
	/*
	if err != nil {
		return fmt.Errorf("error listing ManagedInstances in %s: %v", igm.Name, err)
	}*/

	for _, i := range instances.ManagedInstances {
		name := gce.LastComponent(i.Instance)

		c := &CloudInstanceGroupInstance{
			// FIXME
			//ID: c.Zone() + "/" + name,
			ID: aws.String(name),
		}

		// FIXME not sure if this will work :)
		node := nodeMap[i.Instance]
		if node != nil {
			c.Node = node
		}

		n.NeedUpdate = append(n.NeedUpdate, c)

		/* not certain how to do this
		if readyLaunchConfigurationName == aws.StringValue(i.InstanceTemplate) {
			n.Ready = append(n.Ready, c)
		} else {
			n.NeedUpdate = append(n.NeedUpdate, c)
		}*/
	}


	if len(n.NeedUpdate) == 0 {
		n.Status = "Ready"
	} else {
		n.Status = "NeedsUpdate"
	}

	return n
}

func awsBuildCloudInstanceGroup(ig *api.InstanceGroup, g *autoscaling.Group, nodeMap map[string]*v1.Node) *CloudInstanceGroup {
	n := &CloudInstanceGroup{
		GroupName:     aws.StringValue(g.AutoScalingGroupName),
		InstanceGroup: ig,
		GroupTemplateName: aws.StringValue(g.LaunchConfigurationName),
	}

	readyLaunchConfigurationName := aws.StringValue(g.LaunchConfigurationName)

	for _, i := range g.Instances {
		c := &CloudInstanceGroupInstance{
			ID: i.InstanceId,
		}

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

// DeleteInstanceGroup removes the cloud resources for an InstanceGroup
type DeleteInstanceGroup struct {
	Cluster   *api.Cluster
	Cloud     fi.Cloud
	Clientset simple.Clientset
}

func (c *DeleteInstanceGroup) DeleteInstanceGroup(group *api.InstanceGroup) error {
	groups, err := FindCloudInstanceGroups(c.Cloud, c.Cluster, []*api.InstanceGroup{group}, false, nil)
	if err != nil {
		return fmt.Errorf("error finding CloudInstanceGroups: %v", err)
	}
	cig := groups[group.ObjectMeta.Name]
	if cig == nil {
		glog.Warningf("Group %q not found in cloud - skipping delete", group.ObjectMeta.Name)
	} else {
		if len(groups) != 1 {
			return fmt.Errorf("Multiple InstanceGroup resources found in cloud")
		}

		glog.Infof("Deleting Group %q", group.ObjectMeta.Name)

		err = cig.Delete(c.Cloud)
		if err != nil {
			return fmt.Errorf("error deleting cloud resources for InstanceGroup: %v", err)
		}
	}

	err = c.Clientset.InstanceGroupsFor(c.Cluster).Delete(group.ObjectMeta.Name, nil)
	if err != nil {
		return err
	}

	return nil
}

// CloudInstanceGroup is the AWS ASG backing an InstanceGroup.
type CloudInstanceGroup struct {
	InstanceGroup     *api.InstanceGroup
	GroupName         string
	GroupTemplateName string
	Status            string
	Ready             []*CloudInstanceGroupInstance
	NeedUpdate        []*CloudInstanceGroupInstance
	MinSize           int
	MaxSize           int
}

// CloudInstanceGroupInstance describes an instance in an autoscaling group.
type CloudInstanceGroupInstance struct {
	ID   *string
	Node *v1.Node
}

func (n *CloudInstanceGroup) String() string {
	return "CloudInstanceGroup:" + n.GroupName
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

		instanceId := u.ID

		nodeName := ""
		if u.Node != nil {
			nodeName = u.Node.Name
		}

		if isBastion {

			if err = n.DeleteInstance(u, nodeName, c); err != nil {
				glog.Errorf("Error deleting instance %q: %v", *instanceId, err)
				return err
			}

			glog.Infof("Deleted a bastion instance, %s, and continuing with rolling-update.", *instanceId)

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

		if err = n.DeleteInstance(u, nodeName, c); err != nil {
			glog.Errorf("Error deleting instance %q, node %q: %v", instanceId, nodeName, err)
			return err
		}

		// Wait for new instances to be created
		time.Sleep(t)

		if rollingUpdateData.CloudOnly {

			glog.Warningf("Not validating cluster as cloudonly flag is set.")
			continue

		} else if featureflag.DrainAndValidateRollingUpdate.Enabled() {

			glog.Infof("Validating the cluster.")

			if err = n.ValidateClusterWithRetries(rollingUpdateData, instanceGroupList, t); err != nil {

				if rollingUpdateData.FailOnValidate {
					return fmt.Errorf("error validating cluster after removing a node: %v", err)
				}

				glog.Warningf("Cluster validation failed after removing instance, proceeding since fail-on-validate is set to false: %v", err)
			}
		}
	}

	return nil
}

// ValidateClusterWithRetries runs our validation methods on the K8s Cluster x times and then fails.
func (n *CloudInstanceGroup) ValidateClusterWithRetries(rollingUpdateData *RollingUpdateCluster, instanceGroupList *api.InstanceGroupList, t time.Duration) (err error) {

	// TODO - We are going to need to improve Validate to allow for more than one node, not master
	// TODO - going down at a time.
	for i := 0; i <= rollingUpdateData.ValidateRetries; i++ {

		if _, err = validation.ValidateCluster(rollingUpdateData.ClusterName, instanceGroupList, rollingUpdateData.K8sClient); err != nil {
			glog.Infof("Cluster did not validate, and waiting longer: %v.", err)
			time.Sleep(t / 2)
		} else {
			glog.Infof("Cluster validated.")
			return nil
		}

	}

	// for loop is done, and did not end when the cluster validated
	return fmt.Errorf("cluster validation failed: %v", err)
}

// ValidateCluster runs our validation methods on the K8s Cluster.
func (n *CloudInstanceGroup) ValidateCluster(rollingUpdateData *RollingUpdateCluster, instanceGroupList *api.InstanceGroupList) error {

	if _, err := validation.ValidateCluster(rollingUpdateData.ClusterName, instanceGroupList, rollingUpdateData.K8sClient); err != nil {
		return fmt.Errorf("cluster %q did not pass validation: %v", rollingUpdateData.ClusterName, err)
	}

	return nil

}

// DeleteInstance deletes an Instance.
func (n *CloudInstanceGroup) DeleteInstance(u *CloudInstanceGroupInstance, nodeName string, c fi.Cloud) error {

	if nodeName != "" {
		glog.Infof("Stopping instance %q, node %q, in AWS ASG %q.", *u.ID, nodeName, n.GroupName)
	} else {
		glog.Infof("Stopping instance %q, in AWS ASG %q.", *u.ID, n.GroupName)
	}

	if err := c.DeleteInstance(u.ID); err != nil {
		return fmt.Errorf("error deleting instance: %v", err)
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

func (g *CloudInstanceGroup) Delete(cloud fi.Cloud) error {

	if err := cloud.DeleteGroup(g.GroupName, g.GroupTemplateName); err != nil {
		return fmt.Errorf("unable to delete cloud group: %v", err)
	}

	return nil
}

// StringValue returns the value of the string pointer passed in or
// "" if the pointer is nil.
func StringValue(v *string) string {
	if v != nil {
		return *v
	}
	return ""
}

// Int64Value returns the value of the int64 pointer passed in or
// 0 if the pointer is nil.
func Int64Value(v *int64) int64 {
	if v != nil {
		return *v
	}
	return 0
}
