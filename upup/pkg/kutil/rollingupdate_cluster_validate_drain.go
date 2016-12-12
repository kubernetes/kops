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
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	api "k8s.io/kops/pkg/apis/kops"
	validate "k8s.io/kops/pkg/validation"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	k8s_clientset "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
)

// RollingUpdateCluster restarts cluster nodes
type RollingUpdateClusterDV struct {
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
type RollingUpdateDataDV struct {
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

// TODO: should we check to see if api updates exist in the cluster
// TODO: for instance should we check if Petsets exist when upgrading 1.4.x -> 1.5.x

// Perform a rolling update on a K8s Cluster
func (c *RollingUpdateClusterDV) RollingUpdateDrainValidate(groups map[string]*CloudInstanceGroup, instanceGroups *api.InstanceGroupList) error {
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

				err := group.RollingUpdateDV(rollingUpdateData)

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
				rollingUpdateData := c.CreateRollingUpdateData(instanceGroups, false)

				err := group.RollingUpdateDV(rollingUpdateData)
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

				err := group.RollingUpdateDV(rollingUpdateData)

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

func (c *RollingUpdateClusterDV) CreateRollingUpdateData(instanceGroups *api.InstanceGroupList, isBastion bool) *RollingUpdateDataDV {
	return &RollingUpdateDataDV{
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

// RollingUpdate performs a rolling update on a list of ec2 instances.
func (n *CloudInstanceGroup) RollingUpdateDV(rollingUpdateData *RollingUpdateDataDV) error {

	// we should not get here, but hey I am going to check
	if rollingUpdateData == nil {
		return fmt.Errorf("RollingUpdate cannot be nil")
	}

	// Do not need a k8s client if you are doing cloud only
	if rollingUpdateData.K8sClient == nil && !rollingUpdateData.CloudOnly {
		return fmt.Errorf("RollingUpdate is missing a k8s client")
	}

	if rollingUpdateData.InstanceGroupList == nil {
		return fmt.Errorf("RollingUpdate is missing a the InstanceGroupList")
	}

	c := rollingUpdateData.Cloud.(awsup.AWSCloud)

	update := n.NeedUpdate
	if rollingUpdateData.Force {
		update = append(update, n.Ready...)
	}

	// TODO is this logic correct
	if !rollingUpdateData.IsBastion && rollingUpdateData.FailOnValidate && !rollingUpdateData.CloudOnly {
		_, err := validate.ValidateCluster(rollingUpdateData.ClusterName, rollingUpdateData.InstanceGroupList, rollingUpdateData.K8sClient)
		if err != nil {
			return fmt.Errorf("Cluster %s does not pass validation", rollingUpdateData.ClusterName)
		}
	}

	for _, u := range update {

		if !rollingUpdateData.IsBastion {
			if rollingUpdateData.CloudOnly {
				glog.Warningf("not draining nodes - cloud only is set")
			} else {
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
					glog.Warningf("sleeping only - not validating nodes as cloudonly flag is set")
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

			if rollingUpdateData.CloudOnly {
				glog.Warningf("not validating nodes as cloudonly flag is set")
			} else if err != nil && rollingUpdateData.FailOnValidate {
				return fmt.Errorf("validation timed out while performing rolling update: %v", err)
			}
		}

	}

	return nil
}
