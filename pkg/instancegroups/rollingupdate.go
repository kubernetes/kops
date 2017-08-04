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
	"sync"
	"time"

	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/validation"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/pkg/kubectl/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

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
	CloudOnly        bool
	ClusterName      string
	ValidateRetries  int
	DrainInterval    time.Duration
}

// RollingUpdate performs a rolling update on a K8s Cluster.
func (c *RollingUpdateCluster) RollingUpdate(groups map[string]*fi.CloudGroup, instanceGroups *api.InstanceGroupList) error {
	if len(groups) == 0 {
		glog.Infof("Cloud Instance Group length is zero. Not doing a rolling-update.")
		return nil
	}

	var resultsMutex sync.Mutex
	results := make(map[string]error)

	masterGroups := make(map[string]*fi.CloudGroup)
	nodeGroups := make(map[string]*fi.CloudGroup)
	bastionGroups := make(map[string]*fi.CloudGroup)
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
			go func(k string, group *fi.CloudGroup) {
				resultsMutex.Lock()
				results[k] = fmt.Errorf("function panic bastions")
				resultsMutex.Unlock()

				defer wg.Done()

				err := c.RollingUpdateDefault(group, instanceGroups, true, c.BastionInterval)

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
				results[k] = fmt.Errorf("function panic masters")
				resultsMutex.Unlock()
			}

			defer wg.Done()

			for k, group := range masterGroups {
				err := c.RollingUpdateDefault(group, instanceGroups, false, c.MasterInterval)

				resultsMutex.Lock()
				results[k] = err
				resultsMutex.Unlock()

				// TODO: Bail on error?
			}
		}()

		wg.Wait()
	}

	// Upgrade nodes, with greater parallelism
	{
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
				err := c.RollingUpdateDefault(group, instanceGroups, false, c.NodeInterval)

				resultsMutex.Lock()
				results[k] = err
				resultsMutex.Unlock()

				// TODO: Bail on error?
			}
		}()

		wg.Wait()
	}

	for _, err := range results {
		if err != nil {
			return err
		}
	}

	glog.Infof("Rolling update completed!")
	return nil
}

// TODO: Temporarily increase size of ASG?
// TODO: Remove from ASG first so status is immediately updated?
// TODO: Batch termination, like a rolling-update

// RollingUpdateDefault performs a rolling update on a list of ec2 instances.
func (c *RollingUpdateCluster) RollingUpdateDefault(n *fi.CloudGroup, instanceGroupList *api.InstanceGroupList, isBastion bool, t time.Duration) (err error) {

	// we should not get here, but hey I am going to check.
	if c == nil {
		return fmt.Errorf("rollingUpdate cannot be nil")
	}

	// Do not need a k8s client if you are doing cloudonly.
	if c.K8sClient == nil && !c.CloudOnly {
		return fmt.Errorf("rollingUpdate is missing a k8s client")
	}

	if instanceGroupList == nil {
		return fmt.Errorf("rollingUpdate is missing the InstanceGroupList")
	}

	update := n.NeedUpdate
	if c.Force {
		update = append(update, n.Ready...)
	}

	if len(update) == 0 {
		return nil
	}

	if isBastion {
		glog.V(3).Info("Not validating the cluster as instance is a bastion.")
	} else if c.CloudOnly {
		glog.V(3).Info("Not validating cluster as validation is turned off via the cloud-only flag.")
	} else if featureflag.DrainAndValidateRollingUpdate.Enabled() {
		if err = c.ValidateCluster(instanceGroupList); err != nil {
			if c.FailOnValidate {
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

			if nodeName != "" {
				glog.Infof("Stopping instance %q, node %q, in AWS ASG %q.", *u.ID, nodeName, n.GroupName)
			} else {
				glog.Infof("Stopping instance %q, in AWS ASG %q.", *u.ID, n.GroupName)
			}

			if err := c.Cloud.DeleteInstance(u.ID); err != nil {
				return fmt.Errorf("error deleting instance: %v", err)
			}

			glog.Infof("Deleted a bastion instance, %s, and continuing with rolling-update.", *instanceId)

			continue

		} else if c.CloudOnly {

			glog.Warningf("Not draining cluster nodes as 'cloudonly' flag is set.")

		} else if featureflag.DrainAndValidateRollingUpdate.Enabled() {

			if u.Node != nil {
				glog.Infof("Draining the node: %q.", nodeName)

				if err = c.DrainNode(u); err != nil {
					if c.FailOnDrainError {
						return fmt.Errorf("Failed to drain node %q: %v", nodeName, err)
					} else {
						glog.Infof("Ignoring error draining node %q: %v", nodeName, err)
					}
				}
			} else {
				glog.Warningf("Skipping drain of instance %q, because it is not registered in kubernetes", instanceId)
			}
		}

		if nodeName != "" {
			glog.Infof("Stopping instance %q, node %q, in AWS ASG %q.", *u.ID, nodeName, n.GroupName)
		} else {
			glog.Infof("Stopping instance %q, in AWS ASG %q.", *u.ID, n.GroupName)
		}

		if err := c.Cloud.DeleteInstance(u.ID); err != nil {
			return fmt.Errorf("error deleting instance: %v", err)
		}

		// Wait for new instances to be created
		time.Sleep(t)

		if c.CloudOnly {

			glog.Warningf("Not validating cluster as cloudonly flag is set.")
			continue

		} else if featureflag.DrainAndValidateRollingUpdate.Enabled() {

			glog.Infof("Validating the cluster.")

			if err = c.ValidateClusterWithRetries(instanceGroupList, t); err != nil {

				if c.FailOnValidate {
					return fmt.Errorf("error validating cluster after removing a node: %v", err)
				}

				glog.Warningf("Cluster validation failed after removing instance, proceeding since fail-on-validate is set to false: %v", err)
			}
		}
	}

	return nil
}

// ValidateClusterWithRetries runs our validation methods on the K8s Cluster x times and then fails.
func (c *RollingUpdateCluster) ValidateClusterWithRetries(instanceGroupList *api.InstanceGroupList, t time.Duration) (err error) {

	// TODO - We are going to need to improve Validate to allow for more than one node, not master
	// TODO - going down at a time.
	for i := 0; i <= c.ValidateRetries; i++ {

		if _, err = validation.ValidateCluster(c.ClusterName, instanceGroupList, c.K8sClient); err != nil {
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
func (c *RollingUpdateCluster) ValidateCluster(instanceGroupList *api.InstanceGroupList) error {

	if _, err := validation.ValidateCluster(c.ClusterName, instanceGroupList, c.K8sClient); err != nil {
		return fmt.Errorf("cluster %q did not pass validation: %v", c.ClusterName, err)
	}

	return nil

}

// DrainNode drains a K8s node.
func (c *RollingUpdateCluster) DrainNode(u *fi.CloudGroupInstance) error {
	if c.ClientConfig == nil {
		return fmt.Errorf("ClientConfig not set")
	}
	f := cmdutil.NewFactory(c.ClientConfig)

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

	if c.DrainInterval > time.Second*0 {
		glog.V(3).Infof("Waiting for %s for pods to stabilize after draining.", c.DrainInterval)
		time.Sleep(c.DrainInterval)
	}

	return nil
}
