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

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/upup/pkg/fi"
)

const (
	PRE_CREATE = "pre-create"
	CREATE     = "create"
	ASG_CREATE = "asg" // TODO what is a better more cloud generic term?
)

var StrategyTypes = sets.NewString(PRE_CREATE, CREATE, ASG_CREATE)

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

	Strategy string

	Cluster   *api.Cluster
	Clientset simple.Clientset
}

// RollingUpdate performs a rolling update on a K8s Cluster.
func (r *RollingUpdateCluster) RollingUpdate(groups map[string]*CloudInstanceGroup, instanceGroups *api.InstanceGroupList) error {
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

				err := group.RollingUpdate(r, instanceGroups, true, r.BastionInterval)

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
				err := group.RollingUpdate(r, instanceGroups, false, r.MasterInterval)

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
	if r.Strategy == PRE_CREATE && featureflag.DrainAndValidateRollingUpdate.Enabled() && featureflag.RollingUpdateStrategies.Enabled() {
		return r.RollingUpdateNodesPreCreate(nodeGroups)
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
				if r.Strategy == CREATE && featureflag.DrainAndValidateRollingUpdate.Enabled() && featureflag.RollingUpdateStrategies.Enabled() {
					err = group.RollingUpdateNodesCreate(r)
				} else {
					err = group.RollingUpdate(r, instanceGroups, false, r.NodeInterval)
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

// RollingUpdateNodesPreCreate create all new nodes instance group(s) then cordons all nodes.
// Old nodes are then drained and the old instance group(s) is deleted.
func (r *RollingUpdateCluster) RollingUpdateNodesPreCreate(nodeGroups map[string]*CloudInstanceGroup) error {

	nodeGroupsUpdate := make([]*CloudInstanceGroup, 0)

	// Figure out which CloudInstanceGroups need updating and create a new instance group for each
	{
		for _, group := range nodeGroups {
			update := group.NeedUpdate
			if r.Force {
				update = append(update, group.Ready...)
			}

			if len(update) == 0 {
				return nil
			}

			if _, ok := group.InstanceGroup.ObjectMeta.Annotations[KOPS_IG_CHILD]; !ok {
				suffix := getSuffix(group.InstanceGroup.ObjectMeta.Name)
				ig, err := group.Duplicate(r.Cluster, r.Clientset, suffix)
				if err != nil {
					return fmt.Errorf("unable to create instance group: %v", err)
				}
				glog.Infof("Creating Replacement Instance Group, %q, based on Instance Group %q.", ig.Name, group.InstanceGroup.Name)
			}

			nodeGroupsUpdate = append(nodeGroupsUpdate, group)
		}
	}

	if err := updateCluster(r.Cluster, r.Clientset); err != nil {
		return fmt.Errorf("unable to update cluster: %v", err)
	}

	glog.Info("Waiting for new Instance Group(s) to start")
	time.Sleep(r.NodeInterval)

	// get the new list of ig and validate cluster
	if err := validateCluster(r); err != nil {
		return fmt.Errorf("unable to validate cluster: %v", err)
	}

	if r.CloudOnly {
		glog.Warningf("not cordoning nodes as --cloud-only is set")
	} else {
		// cardon the nodes
		glog.Infof("Cordoning all nodes")
		for _, group := range nodeGroupsUpdate {
			if err := group.CordonNodes(r); err != nil {
				return fmt.Errorf("unable to cordon nodes: %v", err)
			}
		}
	}

	for _, group := range nodeGroupsUpdate {

		if err := group.DrainAndDelete(r); err != nil {
			return fmt.Errorf("unable to drain and delete nodes: %v", err)
		}

		// validate new nodes
		if err := validateCluster(r); err != nil {
			return fmt.Errorf("unable to validate cluster: %v", err)
		}

		glog.Infof("Deleted old Instance Group: %q", group.InstanceGroup.Name)
	}

	glog.Infof("Nodes rolling-update completed")

	return nil
}
