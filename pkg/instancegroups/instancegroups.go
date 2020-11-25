/*
Copyright 2019 The Kubernetes Authors.

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
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/klog/v2"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/validation"
	"k8s.io/kubectl/pkg/drain"
)

const rollingUpdateTaintKey = "kops.k8s.io/scheduled-for-update"

// promptInteractive asks the user to continue, mostly copied from vendor/google.golang.org/api/examples/gmail.go.
func promptInteractive(upgradedHostID, upgradedHostName string) (stopPrompting bool, err error) {
	stopPrompting = false
	scanner := bufio.NewScanner(os.Stdin)
	if upgradedHostName != "" {
		klog.Infof("Pausing after finished %q, node %q", upgradedHostID, upgradedHostName)
	} else {
		klog.Infof("Pausing after finished %q", upgradedHostID)
	}
	fmt.Print("Continue? (Y)es, (N)o, (A)lwaysYes: [Y] ")
	scanner.Scan()
	err = scanner.Err()
	if err != nil {
		klog.Infof("unable to interpret input: %v", err)
		return stopPrompting, err
	}
	val := scanner.Text()
	val = strings.TrimSpace(val)
	val = strings.ToLower(val)
	switch val {
	case "n":
		klog.Info("User signaled to stop")
		os.Exit(3)
	case "a":
		klog.Info("Always Yes, stop prompting for rest of hosts")
		stopPrompting = true
	}
	return stopPrompting, err
}

// RollingUpdate performs a rolling update on a list of instances.
func (c *RollingUpdateCluster) rollingUpdateInstanceGroup(group *cloudinstances.CloudInstanceGroup, sleepAfterTerminate time.Duration) (err error) {
	isBastion := group.InstanceGroup.IsBastion()
	// Do not need a k8s client if you are doing cloudonly.
	if c.K8sClient == nil && !c.CloudOnly {
		return fmt.Errorf("rollingUpdate is missing a k8s client")
	}

	noneReady := len(group.Ready) == 0
	numInstances := len(group.Ready) + len(group.NeedUpdate)
	update := group.NeedUpdate
	if c.Force {
		update = append(update, group.Ready...)
	}

	if len(update) == 0 {
		return nil
	}

	if isBastion {
		klog.V(3).Info("Not validating the cluster as instance is a bastion.")
	} else if err = c.maybeValidate("", 1, group); err != nil {
		return err
	}

	if !c.CloudOnly {
		err = c.taintAllNeedUpdate(group, update)
		if err != nil {
			return err
		}
	}

	settings := resolveSettings(c.Cluster, group.InstanceGroup, numInstances)

	runningDrains := 0
	maxSurge := settings.MaxSurge.IntValue()
	if maxSurge > len(update) {
		maxSurge = len(update)
	}
	maxConcurrency := maxSurge + settings.MaxUnavailable.IntValue()

	if group.InstanceGroup.Spec.Role == api.InstanceGroupRoleMaster && maxSurge != 0 {
		// Masters are incapable of surging because they rely on registering themselves through
		// the local apiserver. That apiserver depends on the local etcd, which relies on being
		// joined to the etcd cluster.
		maxSurge = 0
		maxConcurrency = settings.MaxUnavailable.IntValue()
		if maxConcurrency == 0 {
			maxConcurrency = 1
		}
	}

	if c.Interactive {
		if maxSurge > 1 {
			maxSurge = 1
		}
		maxConcurrency = 1
	}

	update = prioritizeUpdate(update)

	if maxSurge > 0 && !c.CloudOnly {
		for numSurge := 1; numSurge <= maxSurge; numSurge++ {
			u := update[len(update)-numSurge]
			if u.Status != cloudinstances.CloudInstanceStatusDetached {
				if err := c.detachInstance(u); err != nil {
					return err
				}

				// If noneReady, wait until after one node is detached and its replacement validates
				// before detaching more in case the current spec does not result in usable nodes.
				if numSurge == maxSurge || noneReady {
					// Wait for the minimum interval
					klog.Infof("waiting for %v after detaching instance", sleepAfterTerminate)
					time.Sleep(sleepAfterTerminate)

					if err := c.maybeValidate(" after detaching instance", c.ValidateCount, group); err != nil {
						return err
					}
					noneReady = false
				}
			}
		}
	}

	if !*settings.DrainAndTerminate {
		klog.Infof("Rolling updates for InstanceGroup %s are disabled", group.InstanceGroup.Name)
		return nil
	}

	terminateChan := make(chan error, maxConcurrency)

	for uIdx, u := range update {
		go func(m *cloudinstances.CloudInstance) {
			terminateChan <- c.drainTerminateAndWait(m, sleepAfterTerminate)
		}(u)
		runningDrains++

		// Wait until after one node is deleted and its replacement validates before the concurrent draining
		// in case the current spec does not result in usable nodes.
		if runningDrains < maxConcurrency && (!noneReady || uIdx > 0) {
			continue
		}

		err = <-terminateChan
		runningDrains--
		if err != nil {
			return waitForPendingBeforeReturningError(runningDrains, terminateChan, err)
		}

		err = c.maybeValidate(" after terminating instance", c.ValidateCount, group)
		if err != nil {
			return waitForPendingBeforeReturningError(runningDrains, terminateChan, err)
		}

		if c.Interactive {
			nodeName := ""
			if u.Node != nil {
				nodeName = u.Node.Name
			}

			stopPrompting, err := promptInteractive(u.ID, nodeName)
			if err != nil {
				return err
			}
			if stopPrompting {
				// Is a pointer to a struct, changes here push back into the original
				c.Interactive = false
			}
		}

		// Validation tends to return failures from the start of drain until the replacement is
		// fully ready, so sweep up as many completions as we can before starting the next drain.
	sweep:
		for runningDrains > 0 {
			select {
			case err = <-terminateChan:
				runningDrains--
				if err != nil {
					return waitForPendingBeforeReturningError(runningDrains, terminateChan, err)
				}
			default:
				break sweep
			}
		}
	}

	if runningDrains > 0 {
		for runningDrains > 0 {
			err = <-terminateChan
			runningDrains--
			if err != nil {
				return waitForPendingBeforeReturningError(runningDrains, terminateChan, err)
			}
		}

		err = c.maybeValidate(" after terminating instance", c.ValidateCount, group)
		if err != nil {
			return err
		}
	}

	return nil
}

func prioritizeUpdate(update []*cloudinstances.CloudInstance) []*cloudinstances.CloudInstance {
	// The priorities are, in order:
	//   attached before detached
	//   TODO unhealthy before healthy
	//   NeedUpdate before Ready (preserve original order)
	result := make([]*cloudinstances.CloudInstance, 0, len(update))
	var detached []*cloudinstances.CloudInstance
	for _, u := range update {
		if u.Status == cloudinstances.CloudInstanceStatusDetached {
			detached = append(detached, u)
		} else {
			result = append(result, u)
		}
	}

	result = append(result, detached...)
	return result
}

func waitForPendingBeforeReturningError(runningDrains int, terminateChan chan error, err error) error {
	for runningDrains > 0 {
		<-terminateChan
		runningDrains--
	}
	return err
}

func (c *RollingUpdateCluster) taintAllNeedUpdate(group *cloudinstances.CloudInstanceGroup, update []*cloudinstances.CloudInstance) error {
	var toTaint []*corev1.Node
	for _, u := range update {
		if u.Node != nil && !u.Node.Spec.Unschedulable {
			foundTaint := false
			for _, taint := range u.Node.Spec.Taints {
				if taint.Key == rollingUpdateTaintKey {
					foundTaint = true
				}
			}
			if !foundTaint {
				toTaint = append(toTaint, u.Node)
			}
		}
	}
	if len(toTaint) > 0 {
		noun := "nodes"
		if len(toTaint) == 1 {
			noun = "node"
		}
		klog.Infof("Tainting %d %s in %q instancegroup.", len(toTaint), noun, group.InstanceGroup.Name)
		for _, n := range toTaint {
			if err := c.patchTaint(n); err != nil {
				if c.FailOnDrainError {
					return fmt.Errorf("failed to taint node %q: %v", n, err)
				}
				klog.Infof("Ignoring error tainting node %q: %v", n, err)
			}
		}
	}
	return nil
}

func (c *RollingUpdateCluster) patchTaint(node *corev1.Node) error {
	oldData, err := json.Marshal(node)
	if err != nil {
		return err
	}

	node.Spec.Taints = append(node.Spec.Taints, corev1.Taint{
		Key:    rollingUpdateTaintKey,
		Effect: corev1.TaintEffectPreferNoSchedule,
	})

	newData, err := json.Marshal(node)
	if err != nil {
		return err
	}

	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, node)
	if err != nil {
		return err
	}

	_, err = c.K8sClient.CoreV1().Nodes().Patch(c.Ctx, node.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

func (c *RollingUpdateCluster) drainTerminateAndWait(u *cloudinstances.CloudInstance, sleepAfterTerminate time.Duration) error {
	instanceID := u.ID

	nodeName := ""
	if u.Node != nil {
		nodeName = u.Node.Name
	}

	isBastion := u.CloudInstanceGroup.InstanceGroup.IsBastion()

	if isBastion {
		// We don't want to validate for bastions - they aren't part of the cluster
	} else if c.CloudOnly {

		klog.Warning("Not draining cluster nodes as 'cloudonly' flag is set.")

	} else {

		if u.Node != nil {
			klog.Infof("Draining the node: %q.", nodeName)

			if err := c.drainNode(u); err != nil {
				if c.FailOnDrainError {
					return fmt.Errorf("failed to drain node %q: %v", nodeName, err)
				}
				klog.Infof("Ignoring error draining node %q: %v", nodeName, err)
			}
		} else {
			klog.Warningf("Skipping drain of instance %q, because it is not registered in kubernetes", instanceID)
		}
	}

	// We unregister the node before deleting it; if the replacement comes up with the same name it would otherwise still be cordoned
	// (It often seems like GCE tries to re-use names)
	if !isBastion && !c.CloudOnly {
		if u.Node == nil {
			klog.Warningf("no kubernetes Node associated with %s, skipping node deletion", instanceID)
		} else {
			klog.Infof("deleting node %q from kubernetes", nodeName)
			if err := c.deleteNode(u.Node); err != nil {
				return fmt.Errorf("error deleting node %q: %v", nodeName, err)
			}
		}
	}

	if err := c.deleteInstance(u); err != nil {
		klog.Errorf("error deleting instance %q, node %q: %v", instanceID, nodeName, err)
		return err
	}

	if err := c.reconcileInstanceGroup(); err != nil {
		klog.Errorf("error reconciling instance group %q: %v", u.CloudInstanceGroup.HumanName, err)
		return err
	}

	// Wait for the minimum interval
	klog.Infof("waiting for %v after terminating instance", sleepAfterTerminate)
	time.Sleep(sleepAfterTerminate)

	return nil
}

func (c *RollingUpdateCluster) reconcileInstanceGroup() error {
	if api.CloudProviderID(c.Cluster.Spec.CloudProvider) != api.CloudProviderOpenstack &&
		api.CloudProviderID(c.Cluster.Spec.CloudProvider) != api.CloudProviderDO {
		return nil
	}
	rto := fi.RunTasksOptions{}
	rto.InitDefaults()
	applyCmd := &cloudup.ApplyClusterCmd{
		Cloud:              c.Cloud,
		Clientset:          c.Clientset,
		Cluster:            c.Cluster,
		DryRun:             false,
		AllowKopsDowngrade: true,
		RunTasksOptions:    &rto,
		OutDir:             "",
		Phase:              "",
		TargetName:         "direct",
		LifecycleOverrides: map[string]fi.Lifecycle{},
	}

	return applyCmd.Run(c.Ctx)

}

func (c *RollingUpdateCluster) maybeValidate(operation string, validateCount int, group *cloudinstances.CloudInstanceGroup) error {
	if c.CloudOnly {
		klog.Warningf("Not validating cluster as cloudonly flag is set.")

	} else {
		klog.Info("Validating the cluster.")

		if err := c.validateClusterWithTimeout(validateCount, group); err != nil {

			if c.FailOnValidate {
				klog.Errorf("Cluster did not validate within %s", c.ValidationTimeout)
				return fmt.Errorf("error validating cluster%s: %v", operation, err)
			}

			klog.Warningf("Cluster validation failed%s, proceeding since fail-on-validate is set to false: %v", operation, err)
		}
	}
	return nil
}

// validateClusterWithTimeout runs validation.ValidateCluster until either we get positive result or the timeout expires
func (c *RollingUpdateCluster) validateClusterWithTimeout(validateCount int, group *cloudinstances.CloudInstanceGroup) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.ValidationTimeout)
	defer cancel()

	if validateCount == 0 {
		klog.Warningf("skipping cluster validation because validate-count was 0")
		return nil
	}

	successCount := 0

	for {
		// Note that we validate at least once before checking the timeout, in case the cluster is healthy with a short timeout
		result, err := c.ClusterValidator.Validate()
		if err == nil && !hasFailureRelevantToGroup(result.Failures, group) {
			successCount++
			if successCount >= validateCount {
				klog.Info("Cluster validated.")
				return nil
			}
			klog.Infof("Cluster validated; revalidating in %s to make sure it does not flap.", c.ValidateSuccessDuration)
			time.Sleep(c.ValidateSuccessDuration)
			continue
		}

		if err != nil {
			if ctx.Err() != nil {
				klog.Infof("Cluster did not validate within deadline: %v.", err)
				break
			}
			klog.Infof("Cluster did not validate, will retry in %q: %v.", c.ValidateTickDuration, err)
		} else if len(result.Failures) > 0 {
			messages := []string{}
			for _, failure := range result.Failures {
				messages = append(messages, failure.Message)
			}
			if ctx.Err() != nil {
				klog.Infof("Cluster did not pass validation within deadline: %s.", strings.Join(messages, ", "))
				break
			}
			klog.Infof("Cluster did not pass validation, will retry in %q: %s.", c.ValidateTickDuration, strings.Join(messages, ", "))
		}

		// Reset the success count; we want N consecutive successful validations
		successCount = 0

		// Wait before retrying in some cases
		// TODO: Should we check if we have enough time left before the deadline?
		time.Sleep(c.ValidateTickDuration)
	}

	return fmt.Errorf("cluster did not validate within a duration of %q", c.ValidationTimeout)
}

// checks if the validation failures returned after cluster validation are relevant to the current
// instance group whose rolling update is occurring
func hasFailureRelevantToGroup(failures []*validation.ValidationError, group *cloudinstances.CloudInstanceGroup) bool {
	// Ignore non critical validation errors in other instance groups like below target size errors
	for _, failure := range failures {
		// Certain failures like a system-critical-pod failure and dns server related failures
		// set their InstanceGroup to nil, since we cannot associate the failure to any one group
		if failure.InstanceGroup == nil {
			return true
		}

		// if there is a failure in the same instance group or a failure which has cluster wide impact
		if (failure.InstanceGroup.IsMaster()) || (failure.InstanceGroup == group.InstanceGroup) {
			return true
		}
	}

	return false
}

// detachInstance detaches a Cloud Instance
func (c *RollingUpdateCluster) detachInstance(u *cloudinstances.CloudInstance) error {
	id := u.ID
	nodeName := ""
	if u.Node != nil {
		nodeName = u.Node.Name
	}
	if nodeName != "" {
		klog.Infof("Detaching instance %q, node %q, in group %q.", id, nodeName, u.CloudInstanceGroup.HumanName)
	} else {
		klog.Infof("Detaching instance %q, in group %q.", id, u.CloudInstanceGroup.HumanName)
	}

	if err := c.Cloud.DetachInstance(u); err != nil {
		if nodeName != "" {
			return fmt.Errorf("error detaching instance %q, node %q: %v", id, nodeName, err)
		}
		return fmt.Errorf("error detaching instance %q: %v", id, err)
	}

	return nil
}

// deleteInstance deletes an Cloud Instance.
func (c *RollingUpdateCluster) deleteInstance(u *cloudinstances.CloudInstance) error {
	id := u.ID
	nodeName := ""
	if u.Node != nil {
		nodeName = u.Node.Name
	}
	if nodeName != "" {
		klog.Infof("Stopping instance %q, node %q, in group %q (this may take a while).", id, nodeName, u.CloudInstanceGroup.HumanName)
	} else {
		klog.Infof("Stopping instance %q, in group %q (this may take a while).", id, u.CloudInstanceGroup.HumanName)
	}

	if err := c.Cloud.DeleteInstance(u); err != nil {
		if nodeName != "" {
			return fmt.Errorf("error deleting instance %q, node %q: %v", id, nodeName, err)
		}
		return fmt.Errorf("error deleting instance %q: %v", id, err)
	}

	return nil
}

// drainNode drains a K8s node.
func (c *RollingUpdateCluster) drainNode(u *cloudinstances.CloudInstance) error {
	if c.K8sClient == nil {
		return fmt.Errorf("K8sClient not set")
	}

	if u.Node == nil {
		return fmt.Errorf("node not set")
	}

	if u.Node.Name == "" {
		return fmt.Errorf("node name not set")
	}

	helper := &drain.Helper{
		Client:              c.K8sClient,
		Force:               true,
		GracePeriodSeconds:  -1,
		IgnoreAllDaemonSets: true,
		Out:                 os.Stdout,
		ErrOut:              os.Stderr,

		// We want to proceed even when pods are using emptyDir volumes
		DeleteEmptyDirData: true,

		// Other options we might want to set:
		// Timeout?
	}

	if err := drain.RunCordonOrUncordon(helper, u.Node, true); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error cordoning node: %v", err)
	}

	if err := drain.RunNodeDrain(helper, u.Node.Name); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error draining node: %v", err)
	}

	if c.PostDrainDelay > 0 {
		klog.Infof("Waiting for %s for pods to stabilize after draining.", c.PostDrainDelay)
		time.Sleep(c.PostDrainDelay)
	}

	return nil
}

// deleteNode deletes a node from the k8s API.  It does not delete the underlying instance.
func (c *RollingUpdateCluster) deleteNode(node *corev1.Node) error {
	var options metav1.DeleteOptions
	err := c.K8sClient.CoreV1().Nodes().Delete(c.Ctx, node.Name, options)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}

		return fmt.Errorf("error deleting node: %v", err)
	}

	return nil
}

// UpdateSingleInstance performs a rolling update on a single instance
func (c *RollingUpdateCluster) UpdateSingleInstance(cloudMember *cloudinstances.CloudInstance, detach bool) error {
	if detach {
		if cloudMember.CloudInstanceGroup.InstanceGroup.IsMaster() {
			klog.Warning("cannot detach master instances. Assuming --surge=false")

		} else {
			err := c.detachInstance(cloudMember)
			if err != nil {
				return fmt.Errorf("failed to detach instance: %v", err)
			}
			if err := c.maybeValidate(" after detaching instance", c.ValidateCount, cloudMember.CloudInstanceGroup); err != nil {
				return err
			}
		}
	}

	return c.drainTerminateAndWait(cloudMember, 0)
}
