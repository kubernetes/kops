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
	"fmt"
	"os"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/klog"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/drain"
	"k8s.io/kops/upup/pkg/fi"
)

const rollingUpdateTaintKey = "kops.k8s.io/scheduled-for-update"

// RollingUpdateInstanceGroup is the AWS ASG backing an InstanceGroup.
type RollingUpdateInstanceGroup struct {
	// Cloud is the kops cloud provider
	Cloud fi.Cloud
	// CloudGroup is the kops cloud provider groups
	CloudGroup *cloudinstances.CloudInstanceGroup

	// TODO should remove the need to have rollingupdate struct and add:
	// TODO - the kubernetes client
	// TODO - the cluster name
	// TODO - the client config
	// TODO - fail on validate
	// TODO - fail on drain
	// TODO - cloudonly
}

// NewRollingUpdateInstanceGroup creates a new struct
func NewRollingUpdateInstanceGroup(cloud fi.Cloud, cloudGroup *cloudinstances.CloudInstanceGroup) (*RollingUpdateInstanceGroup, error) {
	if cloud == nil {
		return nil, fmt.Errorf("cloud provider is required")
	}
	if cloudGroup == nil {
		return nil, fmt.Errorf("cloud group is required")
	}

	// TODO check more values in cloudGroup that they are set properly

	return &RollingUpdateInstanceGroup{
		Cloud:      cloud,
		CloudGroup: cloudGroup,
	}, nil
}

// promptInteractive asks the user to continue, mostly copied from vendor/google.golang.org/api/examples/gmail.go.
func promptInteractive(upgradedHostId, upgradedHostName string) (stopPrompting bool, err error) {
	stopPrompting = false
	scanner := bufio.NewScanner(os.Stdin)
	if upgradedHostName != "" {
		klog.Infof("Pausing after finished %q, node %q", upgradedHostId, upgradedHostName)
	} else {
		klog.Infof("Pausing after finished %q", upgradedHostId)
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
func (r *RollingUpdateInstanceGroup) RollingUpdate(rollingUpdateData *RollingUpdateCluster, cluster *api.Cluster, isBastion bool, sleepAfterTerminate time.Duration, validationTimeout time.Duration) (err error) {

	// we should not get here, but hey I am going to check.
	if rollingUpdateData == nil {
		return fmt.Errorf("rollingUpdate cannot be nil")
	}

	// Do not need a k8s client if you are doing cloudonly.
	if rollingUpdateData.K8sClient == nil && !rollingUpdateData.CloudOnly {
		return fmt.Errorf("rollingUpdate is missing a k8s client")
	}

	noneReady := len(r.CloudGroup.Ready) == 0
	numInstances := len(r.CloudGroup.Ready) + len(r.CloudGroup.NeedUpdate)
	update := r.CloudGroup.NeedUpdate
	if rollingUpdateData.Force {
		update = append(update, r.CloudGroup.Ready...)
	}

	if len(update) == 0 {
		return nil
	}

	if isBastion {
		klog.V(3).Info("Not validating the cluster as instance is a bastion.")
	} else if rollingUpdateData.CloudOnly {
		klog.V(3).Info("Not validating cluster as validation is turned off via the cloud-only flag.")
	} else {
		if err = r.validateCluster(rollingUpdateData, cluster); err != nil {
			if rollingUpdateData.FailOnValidate {
				return err
			}
			klog.V(2).Infof("Ignoring cluster validation error: %v", err)
			klog.Info("Cluster validation failed, but proceeding since fail-on-validate-error is set to false")
		}
	}

	if !rollingUpdateData.CloudOnly {
		err = r.taintAllNeedUpdate(update, rollingUpdateData)
		if err != nil {
			return err
		}
	}

	settings := resolveSettings(cluster, r.CloudGroup.InstanceGroup, numInstances)

	runningDrains := 0
	maxSurge := settings.MaxSurge.IntValue()
	if maxSurge > len(update) {
		maxSurge = len(update)
	}
	maxConcurrency := maxSurge + settings.MaxUnavailable.IntValue()

	if maxConcurrency == 0 {
		klog.Infof("Rolling updates for InstanceGroup %s are disabled", r.CloudGroup.InstanceGroup.Name)
		return nil
	}

	if r.CloudGroup.InstanceGroup.Spec.Role == api.InstanceGroupRoleMaster && maxSurge != 0 {
		// Masters are incapable of surging because they rely on registering themselves through
		// the local apiserver. That apiserver depends on the local etcd, which relies on being
		// joined to the etcd cluster.
		maxSurge = 0
		maxConcurrency = settings.MaxUnavailable.IntValue()
		if maxConcurrency == 0 {
			maxConcurrency = 1
		}
	}

	if rollingUpdateData.Interactive {
		if maxSurge > 1 {
			maxSurge = 1
		}
		maxConcurrency = 1
	}

	update = prioritizeUpdate(update)

	if maxSurge > 0 && !rollingUpdateData.CloudOnly {
		for numSurge := 1; numSurge <= maxSurge; numSurge++ {
			u := update[len(update)-numSurge]
			if !u.Detached {
				if err := r.detachInstance(u); err != nil {
					return err
				}

				// If noneReady, wait until after one node is detached and its replacement validates
				// before detaching more in case the current spec does not result in usable nodes.
				if numSurge == maxSurge || noneReady {
					// Wait for the minimum interval
					klog.Infof("waiting for %v after detaching instance", sleepAfterTerminate)
					time.Sleep(sleepAfterTerminate)

					if err := r.maybeValidate(rollingUpdateData, validationTimeout, "detaching"); err != nil {
						return err
					}
					noneReady = false
				}
			}
		}
	}

	terminateChan := make(chan error, maxConcurrency)

	for uIdx, u := range update {
		go func(m *cloudinstances.CloudInstanceGroupMember) {
			terminateChan <- r.drainTerminateAndWait(m, rollingUpdateData, isBastion, sleepAfterTerminate)
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

		err = r.maybeValidate(rollingUpdateData, validationTimeout, "removing")
		if err != nil {
			return waitForPendingBeforeReturningError(runningDrains, terminateChan, err)
		}

		if rollingUpdateData.Interactive {
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
				rollingUpdateData.Interactive = false
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

		err = r.maybeValidate(rollingUpdateData, validationTimeout, "removing")
		if err != nil {
			return err
		}
	}

	return nil
}

func prioritizeUpdate(update []*cloudinstances.CloudInstanceGroupMember) []*cloudinstances.CloudInstanceGroupMember {
	// The priorities are, in order:
	//   attached before detached
	//   TODO unhealthy before healthy
	//   NeedUpdate before Ready (preserve original order)
	result := make([]*cloudinstances.CloudInstanceGroupMember, 0, len(update))
	var detached []*cloudinstances.CloudInstanceGroupMember
	for _, u := range update {
		if u.Detached {
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

func (r *RollingUpdateInstanceGroup) taintAllNeedUpdate(update []*cloudinstances.CloudInstanceGroupMember, rollingUpdateData *RollingUpdateCluster) error {
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
		klog.Infof("Tainting %d %s in %q instancegroup.", len(toTaint), noun, r.CloudGroup.InstanceGroup.Name)
		for _, n := range toTaint {
			if err := r.patchTaint(rollingUpdateData, n); err != nil {
				if rollingUpdateData.FailOnDrainError {
					return fmt.Errorf("failed to taint node %q: %v", n, err)
				}
				klog.Infof("Ignoring error tainting node %q: %v", n, err)
			}
		}
	}
	return nil
}

func (r *RollingUpdateInstanceGroup) patchTaint(rollingUpdateData *RollingUpdateCluster, node *corev1.Node) error {
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

	_, err = rollingUpdateData.K8sClient.CoreV1().Nodes().Patch(node.Name, types.StrategicMergePatchType, patchBytes)
	return err
}

func (r *RollingUpdateInstanceGroup) drainTerminateAndWait(u *cloudinstances.CloudInstanceGroupMember, rollingUpdateData *RollingUpdateCluster, isBastion bool, sleepAfterTerminate time.Duration) error {
	instanceId := u.ID

	nodeName := ""
	if u.Node != nil {
		nodeName = u.Node.Name
	}

	if isBastion {
		// We don't want to validate for bastions - they aren't part of the cluster
	} else if rollingUpdateData.CloudOnly {

		klog.Warning("Not draining cluster nodes as 'cloudonly' flag is set.")

	} else {

		if u.Node != nil {
			klog.Infof("Draining the node: %q.", nodeName)

			if err := r.DrainNode(u, rollingUpdateData); err != nil {
				if rollingUpdateData.FailOnDrainError {
					return fmt.Errorf("failed to drain node %q: %v", nodeName, err)
				}
				klog.Infof("Ignoring error draining node %q: %v", nodeName, err)
			}
		} else {
			klog.Warningf("Skipping drain of instance %q, because it is not registered in kubernetes", instanceId)
		}
	}

	// We unregister the node before deleting it; if the replacement comes up with the same name it would otherwise still be cordoned
	// (It often seems like GCE tries to re-use names)
	if !isBastion && !rollingUpdateData.CloudOnly {
		if u.Node == nil {
			klog.Warningf("no kubernetes Node associated with %s, skipping node deletion", instanceId)
		} else {
			klog.Infof("deleting node %q from kubernetes", nodeName)
			if err := r.deleteNode(u.Node, rollingUpdateData); err != nil {
				return fmt.Errorf("error deleting node %q: %v", nodeName, err)
			}
		}
	}

	if err := r.DeleteInstance(u); err != nil {
		klog.Errorf("error deleting instance %q, node %q: %v", instanceId, nodeName, err)
		return err
	}

	// Wait for the minimum interval
	klog.Infof("waiting for %v after terminating instance", sleepAfterTerminate)
	time.Sleep(sleepAfterTerminate)

	return nil
}

func (r *RollingUpdateInstanceGroup) maybeValidate(rollingUpdateData *RollingUpdateCluster, validationTimeout time.Duration, operation string) error {
	if rollingUpdateData.CloudOnly {
		klog.Warningf("Not validating cluster as cloudonly flag is set.")

	} else {
		klog.Info("Validating the cluster.")

		if err := r.validateClusterWithDuration(rollingUpdateData, validationTimeout); err != nil {

			if rollingUpdateData.FailOnValidate {
				klog.Errorf("Cluster did not validate within %s", validationTimeout)
				return fmt.Errorf("error validating cluster after %s a node: %v", operation, err)
			}

			klog.Warningf("Cluster validation failed after %s instance, proceeding since fail-on-validate is set to false: %v", operation, err)
		}
	}
	return nil
}

// validateClusterWithDuration runs validation.ValidateCluster until either we get positive result or the timeout expires
func (r *RollingUpdateInstanceGroup) validateClusterWithDuration(rollingUpdateData *RollingUpdateCluster, duration time.Duration) error {
	// Try to validate cluster at least once, this will handle durations that are lower
	// than our tick time
	if r.tryValidateCluster(rollingUpdateData, duration, rollingUpdateData.ValidateTickDuration) {
		return nil
	}

	timeout := time.After(duration)
	ticker := time.NewTicker(rollingUpdateData.ValidateTickDuration)
	defer ticker.Stop()
	// Keep trying until we're timed out or got a result or got an error
	for {
		select {
		case <-timeout:
			// Got a timeout fail with a timeout error
			return fmt.Errorf("cluster did not validate within a duration of %q", duration)
		case <-ticker.C:
			// Got a tick, validate cluster
			if r.tryValidateCluster(rollingUpdateData, duration, rollingUpdateData.ValidateTickDuration) {
				return nil
			}
			// ValidateCluster didn't work yet, so let's try again
			// this will exit up to the for loop
		}
	}
}

func (r *RollingUpdateInstanceGroup) tryValidateCluster(rollingUpdateData *RollingUpdateCluster, duration time.Duration, tickDuration time.Duration) bool {
	result, err := rollingUpdateData.ClusterValidator.Validate()

	if err == nil && len(result.Failures) == 0 && rollingUpdateData.ValidateSuccessDuration > 0 {
		klog.Infof("Cluster validated; revalidating in %s to make sure it does not flap.", rollingUpdateData.ValidateSuccessDuration)
		time.Sleep(rollingUpdateData.ValidateSuccessDuration)
		result, err = rollingUpdateData.ClusterValidator.Validate()
	}

	if err != nil {
		klog.Infof("Cluster did not validate, will try again in %q until duration %q expires: %v.", tickDuration, duration, err)
		return false
	} else if len(result.Failures) > 0 {
		messages := []string{}
		for _, failure := range result.Failures {
			messages = append(messages, failure.Message)
		}
		klog.Infof("Cluster did not pass validation, will try again in %q until duration %q expires: %s.", tickDuration, duration, strings.Join(messages, ", "))
		return false
	} else {
		klog.Info("Cluster validated.")
		return true
	}
}

// validateCluster runs our validation methods on the K8s Cluster.
func (r *RollingUpdateInstanceGroup) validateCluster(rollingUpdateData *RollingUpdateCluster, cluster *api.Cluster) error {
	result, err := rollingUpdateData.ClusterValidator.Validate()
	if err != nil {
		return fmt.Errorf("cluster %q did not validate: %v", cluster.Name, err)
	}
	if len(result.Failures) > 0 {
		messages := []string{}
		for _, failure := range result.Failures {
			messages = append(messages, failure.Message)
		}
		return fmt.Errorf("cluster %q did not pass validation: %s", cluster.Name, strings.Join(messages, ", "))
	}

	return nil

}

// detachInstance detaches a Cloud Instance
func (r *RollingUpdateInstanceGroup) detachInstance(u *cloudinstances.CloudInstanceGroupMember) error {
	id := u.ID
	nodeName := ""
	if u.Node != nil {
		nodeName = u.Node.Name
	}
	if nodeName != "" {
		klog.Infof("Detaching instance %q, node %q, in group %q.", id, nodeName, r.CloudGroup.HumanName)
	} else {
		klog.Infof("Detaching instance %q, in group %q.", id, r.CloudGroup.HumanName)
	}

	if err := r.Cloud.DetachInstance(u); err != nil {
		if nodeName != "" {
			return fmt.Errorf("error detaching instance %q, node %q: %v", id, nodeName, err)
		} else {
			return fmt.Errorf("error detaching instance %q: %v", id, err)
		}
	}

	return nil
}

// DeleteInstance deletes an Cloud Instance.
func (r *RollingUpdateInstanceGroup) DeleteInstance(u *cloudinstances.CloudInstanceGroupMember) error {
	id := u.ID
	nodeName := ""
	if u.Node != nil {
		nodeName = u.Node.Name
	}
	if nodeName != "" {
		klog.Infof("Stopping instance %q, node %q, in group %q (this may take a while).", id, nodeName, r.CloudGroup.HumanName)
	} else {
		klog.Infof("Stopping instance %q, in group %q (this may take a while).", id, r.CloudGroup.HumanName)
	}

	if err := r.Cloud.DeleteInstance(u); err != nil {
		if nodeName != "" {
			return fmt.Errorf("error deleting instance %q, node %q: %v", id, nodeName, err)
		} else {
			return fmt.Errorf("error deleting instance %q: %v", id, err)
		}
	}

	return nil

}

// DrainNode drains a K8s node.
func (r *RollingUpdateInstanceGroup) DrainNode(u *cloudinstances.CloudInstanceGroupMember, rollingUpdateData *RollingUpdateCluster) error {
	if rollingUpdateData.K8sClient == nil {
		return fmt.Errorf("K8sClient not set")
	}

	if u.Node == nil {
		return fmt.Errorf("node not set")
	}

	if u.Node.Name == "" {
		return fmt.Errorf("node name not set")
	}

	helper := &drain.Helper{
		Client:              rollingUpdateData.K8sClient,
		Force:               true,
		GracePeriodSeconds:  -1,
		IgnoreAllDaemonSets: true,
		Out:                 os.Stdout,
		ErrOut:              os.Stderr,

		// We want to proceed even when pods are using local data (emptyDir)
		DeleteLocalData: true,

		// Other options we might want to set:
		// Timeout?
	}

	if err := drain.RunCordonOrUncordon(helper, u.Node, true); err != nil {
		return fmt.Errorf("error cordoning node: %v", err)
	}

	if err := drain.RunNodeDrain(helper, u.Node.Name); err != nil {
		return fmt.Errorf("error draining node: %v", err)
	}

	if rollingUpdateData.PostDrainDelay > 0 {
		klog.Infof("Waiting for %s for pods to stabilize after draining.", rollingUpdateData.PostDrainDelay)
		time.Sleep(rollingUpdateData.PostDrainDelay)
	}

	return nil
}

// DeleteNode deletes a node from the k8s API.  It does not delete the underlying instance.
func (r *RollingUpdateInstanceGroup) deleteNode(node *corev1.Node, rollingUpdateData *RollingUpdateCluster) error {
	k8sclient := rollingUpdateData.K8sClient
	var options metav1.DeleteOptions
	err := k8sclient.CoreV1().Nodes().Delete(node.Name, &options)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}

		return fmt.Errorf("error deleting node: %v", err)
	}

	return nil
}
