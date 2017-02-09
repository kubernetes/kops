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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/validation"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/pkg/featureflag"
	k8s_clientset "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
)

// RollingUpdateClusterDrainValidate restarts cluster nodes.
type RollingUpdateClusterDrainValidate struct {
	Cloud fi.Cloud

	MasterInterval  time.Duration
	NodeInterval    time.Duration
	BastionInterval time.Duration
	K8sClient       *k8s_clientset.Clientset

	FailOnDrainError bool
	FailOnValidate   bool

	Force bool

	CloudOnly   bool
	ClusterName string

}
const retries = 8

// RollingUpdateDrainValidate performs a rolling update on a K8s Cluster.
func (c *RollingUpdateClusterDrainValidate) RollingUpdateDrainValidate(groups map[string]*CloudInstanceGroup, instanceGroups *api.InstanceGroupList) error {
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
			return fmt.Errorf("unknown group type for group: %q", group.InstanceGroup.ObjectMeta.Name)
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

				err := group.RollingUpdateDrainValidate(c, instanceGroups,true, c.BastionInterval)

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
				// FIXME refactor
				err := group.RollingUpdateDrainValidate(c, instanceGroups,false, c.MasterInterval)
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

		for k, nodeGroup := range nodeGroups {
			wg.Add(1)
			go func(k string, group *CloudInstanceGroup) {
				resultsMutex.Lock()
				results[k] = fmt.Errorf("function panic")
				resultsMutex.Unlock()

				defer wg.Done()

				err := group.RollingUpdateDrainValidate(c, instanceGroups,false, c.NodeInterval)

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

	glog.Infof("\nRolling update completed!\n")
	return nil
}

// ValidateClusterWithRetries runs our validation methods on the K8s Cluster x times and then fails.
func (n *CloudInstanceGroup) ValidateClusterWithRetries(rollingUpdateData *RollingUpdateClusterDrainValidate, instanceGroupList *api.InstanceGroupList,t time.Duration) (err error) {

	for i := 0; i <= retries; i++ {
		_, err = validation.ValidateCluster(rollingUpdateData.ClusterName, instanceGroupList, rollingUpdateData.K8sClient)
		if err != nil {
			glog.Infof("Waiting longer for kops validate to pass: %s.", err)
			time.Sleep(t / 2)
		} else {
			glog.Infof("Cluster validated proceeding with next step in rolling update.")
			break
		}
	}

	return nil
}

// ValidateCluster runs our validation methods on the K8s Cluster.
func (n *CloudInstanceGroup) ValidateCluster(rollingUpdateData *RollingUpdateClusterDrainValidate, instanceGroupList *api.InstanceGroupList) error {

	_, err := validation.ValidateCluster(rollingUpdateData.ClusterName, instanceGroupList, rollingUpdateData.K8sClient)
	if err != nil {
		return fmt.Errorf("cluster %s does not pass validation", rollingUpdateData.ClusterName)
	}

	return nil

}

// DeleteAWSInstance deletes an EC2 AWS Instance.
func (n *CloudInstanceGroup) DeleteAWSInstance(u *CloudInstanceGroupInstance, c awsup.AWSCloud) error {

	instanceID := aws.StringValue(u.ASGInstance.InstanceId)

	glog.Infof("Stopping instance %q in AWS ASG %q.", instanceID, n.ASGName)

	request := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{u.ASGInstance.InstanceId},
	}
	_, err := c.EC2().TerminateInstances(request)
	if err != nil {
		return fmt.Errorf("error deleting instance %q: %v", instanceID, err)
	}

	return nil

}

// DrainNode drains a K8s node.
func (n *CloudInstanceGroup) DrainNode(u *CloudInstanceGroupInstance, rollingUpdateData *RollingUpdateClusterDrainValidate) error {
	drain, err := NewDrainOptions(nil, rollingUpdateData.ClusterName)

	if err != nil {

		glog.Warningf("Error creating drain: %v.", err)
		if rollingUpdateData.FailOnDrainError {
			return fmt.Errorf("error creating drain: %v.", err)
		} else {
			glog.Infof("Proceeding with rolling-update since fail-on-drain-error is set to false.")
		}

	} else {

		err = drain.DrainTheNode(u.Node.Name)
		if err != nil {
			glog.Warningf("Error draining node: %v.", err)
			if rollingUpdateData.FailOnDrainError {
				return fmt.Errorf("error draining node: %v", err)
			} else {
				glog.Infof("Proceeding with rolling-update since fail-on-drain-error is set to false.")
			}
		}

	}

	return nil
}

// TODO: Temporarily increase size of ASG?
// TODO: Remove from ASG first so status is immediately updated?
// TODO: Batch termination, like a rolling-update

// RollingUpdateDrainValidate performs a rolling update on a list of ec2 instances.
func (n *CloudInstanceGroup) RollingUpdateDrainValidate(rollingUpdateData *RollingUpdateClusterDrainValidate, instanceGroupList *api.InstanceGroupList, isBastion bool, t time.Duration) (err error) {

	// we should not get here, but hey I am going to check
	if rollingUpdateData == nil {
		return fmt.Errorf("rollingUpdate cannot be nil")
	}

	// Do not need a k8s client if you are doing cloud only
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


	if isBastion || rollingUpdateData.CloudOnly {
		glog.Infof("Not validating the cluster")
	} else if rollingUpdateData.FailOnValidate && featureflag.DrainAndValidateRollingUpdate.Enabled(){
		if err = n.ValidateCluster(rollingUpdateData, instanceGroupList); err != nil {
			glog.Errorf("Error validating cluster: %s.", err)
			return err
		}
	}

	for _, u := range update {


		if isBastion {

			if err = n.DeleteAWSInstance(u, c); err != nil {
				return err
			}
			glog.Infof("Updated a bastion instance and continuing.")

			continue
		}

		if rollingUpdateData.CloudOnly {
			glog.Warningf("Not draining nodes - cloud only is set.")
		} else if featureflag.DrainAndValidateRollingUpdate.Enabled() {

			glog.Infof("Draining the node.")

			if err = n.DrainNode(u,rollingUpdateData); err == nil {
				glog.Errorf("Error draining node: %s.", err)
				return err
			}
		}

		if err = n.DeleteAWSInstance(u, c); err != nil {
			glog.Errorf("Error deleting aws instance: %s.", err)
			return err
		}

		// Wait for new EC2 instances to be created
		time.Sleep(t)

		if rollingUpdateData.CloudOnly {
			glog.Warningf("Not validating nodes as cloudonly flag is set.")
			return nil
		} else if featureflag.DrainAndValidateRollingUpdate.Enabled() {

			glog.Infof("Validating the cluster.")
			if err = n.ValidateClusterWithRetries(rollingUpdateData, instanceGroupList,t); err != nil {
				glog.Errorf("Error validating cluster: %s.", err)
				return err
			}

			if err != nil && rollingUpdateData.FailOnValidate {
				return fmt.Errorf("validation timed out while performing rolling update: %v", err)
			}
		}

	}

	return nil
}
