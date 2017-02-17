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

// Based off of drain in kubectl:
// https://github.com/kubernetes/kubernetes/blob/master/pkg/kubectl/cmd/drain.go
// We have duplicated this code because the Drain options struct has private members that are setup
// via a cobra cmd object, and I would rather not hack that.
// Also, logging is a bit cleaner not using the kubectl code directly.

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/api"
	apierrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/meta"
	metav1 "k8s.io/kubernetes/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/apis/policy"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/kubectl"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/kubelet/types"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/wait"
)

// DrainOptions For Draining Node.
type DrainOptions struct {
	client             *internalclientset.Clientset
	restClient         *restclient.RESTClient
	factory            cmdutil.Factory
	Force              bool
	GracePeriodSeconds int
	IgnoreDaemonsets   bool
	Timeout            time.Duration
	backOff            clockwork.Clock
	DeleteLocalData    bool
	mapper             meta.RESTMapper
	nodeInfo           *resource.Info
	typer              runtime.ObjectTyper
}

// Allow tweaking default options for draining nodes.
type DrainCommand struct {
	Force              bool
	IgnoreDaemonsets   bool
	DeleteLocalData    bool
	GracePeriodSeconds int
	Timeout            int
}

// Takes a pod and returns a bool indicating whether or not to operate on the
// pod, an optional warning message, and an optional fatal error.
type podFilter func(api.Pod) (include bool, w *warning, f *fatal)
type warning struct {
	string
}
type fatal struct {
	string
}

const (
	EvictionKind        = "Eviction"
	EvictionSubresource = "pods/eviction"

	kDaemonsetFatal      = "DaemonSet-managed pods (use --ignore-daemonsets to ignore)"
	kDaemonsetWarning    = "Ignoring DaemonSet-managed pods"
	kLocalStorageFatal   = "pods with local storage (use --delete-local-data to override)"
	kLocalStorageWarning = "Deleting pods with local storage"
	kUnmanagedFatal      = "pods not managed by ReplicationController, ReplicaSet, Job, DaemonSet or StatefulSet (use --force to override)"
	kUnmanagedWarning    = "Deleting pods not managed by ReplicationController, ReplicaSet, Job, DaemonSet or StatefulSet"
	kMaxNodeUpdateRetry  = 10
)

// Create a NewDrainOptions.
func NewDrainOptions(command *DrainCommand, clusterName string) (*DrainOptions, error) {

	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: clusterName})
	f := cmdutil.NewFactory(config)

	if command != nil {
		duration, err := time.ParseDuration(fmt.Sprintf("%ds", command.GracePeriodSeconds))
		if err != nil {
			return nil, err
		}
		return &DrainOptions{
			factory:            f,
			backOff:            clockwork.NewRealClock(),
			Force:              command.Force,
			IgnoreDaemonsets:   command.IgnoreDaemonsets,
			DeleteLocalData:    command.DeleteLocalData,
			GracePeriodSeconds: command.GracePeriodSeconds,
			Timeout:            duration,
		}, nil
	}

	// return will defaults
	duration, err := time.ParseDuration("0s")
	if err != nil {
		return nil, err
	}
	return &DrainOptions{
		factory:            f,
		backOff:            clockwork.NewRealClock(),
		Force:              true,
		IgnoreDaemonsets:   true,
		DeleteLocalData:    true,
		GracePeriodSeconds: -1,
		Timeout:            duration,
	}, nil

}

// DrainTheNode drains a k8s node.
func (o *DrainOptions) DrainTheNode(nodeName string) (err error) {

	err = o.SetupDrain(nodeName)

	if err != nil {
		return fmt.Errorf("error setting up the drain: %v, node: %s", err, nodeName)
	}
	err = o.RunDrain()

	if err != nil {
		return fmt.Errorf("drain failed %v, %s", err, nodeName)
	}

	// Sleep a bit extra to let pods clear the node.
	// 90 seconds equals the default pod termination period plus 30 seconds.
	time.Sleep(time.Second * 90)

	return nil
}

// SetupDrain populates some fields from the factory, grabs command line
// arguments and looks up the node using Builder.
func (o *DrainOptions) SetupDrain(nodeName string) error {

	if nodeName == "" {
		return fmt.Errorf("nodeName cannot be empty")
	}

	var err error

	if o.client, err = o.factory.ClientSet(); err != nil {
		return fmt.Errorf("client or clientset nil %v", err)
	}

	o.restClient, err = o.factory.RESTClient()
	if err != nil {
		return fmt.Errorf("rest client problem %v", err)
	}

	o.mapper, o.typer = o.factory.Object()

	cmdNamespace, _, err := o.factory.DefaultNamespace()
	if err != nil {
		return fmt.Errorf("DefaultNamespace problem %v", err)
	}

	r := o.factory.NewBuilder().
		NamespaceParam(cmdNamespace).DefaultNamespace().
		ResourceNames("node", nodeName).
		Do()

	if err = r.Err(); err != nil {
		return fmt.Errorf("NewBuilder problem %v", err)
	}

	err = r.Visit(func(info *resource.Info, err error) error {
		if err != nil {
			return fmt.Errorf("internal vistor problem %v", err)
		}
		glog.V(5).Infof("info %v", info)
		o.nodeInfo = info
		return nil
	})

	if err != nil {
		glog.Fatalf("Error getting nodeInfo %v.", err)
		return fmt.Errorf("vistor problem %v", err)
	}

	if err = r.Err(); err != nil {
		return fmt.Errorf("vistor problem %v", err)
	}

	return nil
}

// RunDrain runs the 'drain' command
func (o *DrainOptions) RunDrain() error {
	if err := o.RunCordonOrUncordon(true); err != nil {
		return err
	}

	err := o.deleteOrEvictPodsSimple()
	if err == nil {
		glog.Infof("node %q drained", o.nodeInfo.Name)
	}
	return err
}

func (o *DrainOptions) deleteOrEvictPodsSimple() error {
	pods, err := o.getPodsForDeletion()
	if err != nil {
		return err
	}

	err = o.deleteOrEvictPods(pods)
	if err != nil {
		pendingPods, newErr := o.getPodsForDeletion()
		if newErr != nil {
			return newErr
		}
		glog.Warningf("There are pending pods when an error occurred: %v\n", err)
		for _, pendingPod := range pendingPods {
			glog.Warningf("%s/%s\n", "pod", pendingPod.Name)
		}
	}
	return err
}

func (o *DrainOptions) getController(sr *api.SerializedReference) (interface{}, error) {
	switch sr.Reference.Kind {
	case "ReplicationController":
		return o.client.Core().ReplicationControllers(sr.Reference.Namespace).Get(sr.Reference.Name, metav1.GetOptions{})
	case "DaemonSet":
		return o.client.Extensions().DaemonSets(sr.Reference.Namespace).Get(sr.Reference.Name, metav1.GetOptions{})
	case "Job":
		return o.client.Batch().Jobs(sr.Reference.Namespace).Get(sr.Reference.Name, metav1.GetOptions{})
	case "ReplicaSet":
		return o.client.Extensions().ReplicaSets(sr.Reference.Namespace).Get(sr.Reference.Name, metav1.GetOptions{})
	case "StatefulSet":
		return o.client.Apps().StatefulSets(sr.Reference.Namespace).Get(sr.Reference.Name, metav1.GetOptions{})
	}
	return nil, fmt.Errorf("Unknown controller kind %q", sr.Reference.Kind)
}

func (o *DrainOptions) getPodCreator(pod api.Pod) (*api.SerializedReference, error) {
	creatorRef, found := pod.ObjectMeta.Annotations[api.CreatedByAnnotation]
	if !found {
		return nil, nil
	}
	// Now verify that the specified creator actually exists.
	sr := &api.SerializedReference{}
	if err := runtime.DecodeInto(o.factory.Decoder(true), []byte(creatorRef), sr); err != nil {
		return nil, err
	}
	// We assume the only reason for an error is because the controller is
	// gone/missing, not for any other cause.  TODO(mml): something more
	// sophisticated than this
	_, err := o.getController(sr)
	if err != nil {
		return nil, err
	}
	return sr, nil
}

func (o *DrainOptions) unreplicatedFilter(pod api.Pod) (bool, *warning, *fatal) {
	// any finished pod can be removed
	if pod.Status.Phase == api.PodSucceeded || pod.Status.Phase == api.PodFailed {
		return true, nil, nil
	}

	sr, err := o.getPodCreator(pod)
	if err != nil {
		return false, nil, &fatal{err.Error()}
	}
	if sr != nil {
		return true, nil, nil
	}
	if !o.Force {
		return false, nil, &fatal{kUnmanagedFatal}
	}
	return true, &warning{kUnmanagedWarning}, nil
}

func (o *DrainOptions) daemonsetFilter(pod api.Pod) (bool, *warning, *fatal) {
	// Note that we return false in all cases where the pod is DaemonSet managed,
	// regardless of flags.  We never delete them, the only question is whether
	// their presence constitutes an error.
	sr, err := o.getPodCreator(pod)
	if err != nil {
		return false, nil, &fatal{err.Error()}
	}
	if sr == nil || sr.Reference.Kind != "DaemonSet" {
		return true, nil, nil
	}
	if _, err := o.client.Extensions().DaemonSets(sr.Reference.Namespace).Get(sr.Reference.Name, metav1.GetOptions{}); err != nil {
		return false, nil, &fatal{err.Error()}
	}
	if !o.IgnoreDaemonsets {
		return false, nil, &fatal{kDaemonsetFatal}
	}
	return false, &warning{kDaemonsetWarning}, nil
}

func mirrorPodFilter(pod api.Pod) (bool, *warning, *fatal) {
	if _, found := pod.ObjectMeta.Annotations[types.ConfigMirrorAnnotationKey]; found {
		return false, nil, nil
	}
	return true, nil, nil
}

func hasLocalStorage(pod api.Pod) bool {
	for _, volume := range pod.Spec.Volumes {
		if volume.EmptyDir != nil {
			return true
		}
	}

	return false
}

func (o *DrainOptions) localStorageFilter(pod api.Pod) (bool, *warning, *fatal) {
	if !hasLocalStorage(pod) {
		return true, nil, nil
	}
	if !o.DeleteLocalData {
		return false, nil, &fatal{kLocalStorageFatal}
	}
	return true, &warning{kLocalStorageWarning}, nil
}

// Map of status message to a list of pod names having that status.
type podStatuses map[string][]string

func (ps podStatuses) Message() string {
	msgs := []string{}

	for key, pods := range ps {
		msgs = append(msgs, fmt.Sprintf("%s: %s", key, strings.Join(pods, ", ")))
	}
	return strings.Join(msgs, "; ")
}

// getPodsForDeletion returns all the pods we're going to delete.  If there are
// any pods preventing us from deleting, we return that list in an error.
func (o *DrainOptions) getPodsForDeletion() (pods []api.Pod, err error) {
	podList, err := o.client.Core().Pods(api.NamespaceAll).List(api.ListOptions{
		FieldSelector: fields.SelectorFromSet(fields.Set{"spec.nodeName": o.nodeInfo.Name})})
	if err != nil {
		return pods, err
	}

	ws := podStatuses{}
	fs := podStatuses{}

	for _, pod := range podList.Items {
		podOk := true
		for _, filt := range []podFilter{mirrorPodFilter, o.localStorageFilter, o.unreplicatedFilter, o.daemonsetFilter} {
			filterOk, w, f := filt(pod)

			podOk = podOk && filterOk
			if w != nil {
				ws[w.string] = append(ws[w.string], pod.Name)
			}
			if f != nil {
				fs[f.string] = append(fs[f.string], pod.Name)
			}
		}
		if podOk {
			pods = append(pods, pod)
		}
	}

	if len(fs) > 0 {
		return []api.Pod{}, errors.New(fs.Message())
	}
	if len(ws) > 0 {
		glog.V(3).Infof("%s", ws.Message())
	}
	return pods, nil
}

func (o *DrainOptions) deletePod(pod api.Pod) error {
	deleteOptions := &api.DeleteOptions{}
	if o.GracePeriodSeconds >= 0 {
		gracePeriodSeconds := int64(o.GracePeriodSeconds)
		deleteOptions.GracePeriodSeconds = &gracePeriodSeconds
	}
	return o.client.Core().Pods(pod.Namespace).Delete(pod.Name, deleteOptions)
}

func (o *DrainOptions) evictPod(pod api.Pod, policyGroupVersion string) error {
	deleteOptions := &api.DeleteOptions{}
	if o.GracePeriodSeconds >= 0 {
		gracePeriodSeconds := int64(o.GracePeriodSeconds)
		deleteOptions.GracePeriodSeconds = &gracePeriodSeconds
	}
	eviction := &policy.Eviction{
		TypeMeta: metav1.TypeMeta{
			APIVersion: policyGroupVersion,
			Kind:       EvictionKind,
		},
		ObjectMeta: api.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
		DeleteOptions: deleteOptions,
	}
	// Remember to change change the URL manipulation func when Evction's version change
	return o.client.Policy().Evictions(eviction.Namespace).Evict(eviction)
}

// deleteOrEvictPods deletes or evicts the pods on the api server
func (o *DrainOptions) deleteOrEvictPods(pods []api.Pod) error {
	if len(pods) == 0 {
		return nil
	}

	policyGroupVersion, err := SupportEviction(o.client)
	if err != nil {
		return err
	}

	getPodFn := func(namespace, name string) (*api.Pod, error) {
		return o.client.Core().Pods(namespace).Get(name, metav1.GetOptions{})
	}

	if len(policyGroupVersion) > 0 {
		return o.evictPods(pods, policyGroupVersion, getPodFn)
	} else {
		return o.deletePods(pods, getPodFn)
	}
}

func (o *DrainOptions) evictPods(pods []api.Pod, policyGroupVersion string, getPodFn func(namespace, name string) (*api.Pod, error)) error {
	doneCh := make(chan bool, len(pods))
	errCh := make(chan error, 1)

	for _, pod := range pods {
		go func(pod api.Pod, doneCh chan bool, errCh chan error) {
			var err error
			for {
				err = o.evictPod(pod, policyGroupVersion)
				if err == nil {
					glog.V(3).Infof("evicted pod %q", pod.Name)
					break
				} else if apierrors.IsTooManyRequests(err) {
					time.Sleep(5 * time.Second)
				} else {

					// TODO this is the work around that I put in place for problems with evictions
					// TODO see https://github.com/kubernetes/kubernetes/issues/41656
					glog.Infof("trying to delete pod, because of error when evicting pod %q: %v", pod.Name, err)

					err2 := o.deletePod(pod)

					if err2 != nil {
						errCh <- fmt.Errorf("error when deleting, and evicting pod %q: %v, %v", pod.Name, err, err2)
						return
					}

					break
				}
			}
			podArray := []api.Pod{pod}
			_, err = o.waitForDelete(podArray, kubectl.Interval, time.Duration(math.MaxInt64), true, getPodFn)
			if err == nil {
				glog.V(3).Infof("finished evicting pod %q", pod.Name)
				doneCh <- true
			} else {
				errCh <- fmt.Errorf("error when waiting for pod %q terminating: %v", pod.Name, err)
			}
		}(pod, doneCh, errCh)
	}

	doneCount := 0
	// 0 timeout means infinite, we use MaxInt64 to represent it.
	var globalTimeout time.Duration
	if o.Timeout == 0 {
		globalTimeout = time.Duration(math.MaxInt64)
	} else {
		globalTimeout = o.Timeout
	}
	for {
		select {
		case err := <-errCh:
			return err
		case <-doneCh:
			doneCount++
			if doneCount == len(pods) {
				return nil
			}
		case <-time.After(globalTimeout):
			return fmt.Errorf("Drain did not complete within %v", globalTimeout)
		}
	}
}

func (o *DrainOptions) deletePods(pods []api.Pod, getPodFn func(namespace, name string) (*api.Pod, error)) error {
	// 0 timeout means infinite, we use MaxInt64 to represent it.
	var globalTimeout time.Duration
	if o.Timeout == 0 {
		globalTimeout = time.Duration(math.MaxInt64)
	} else {
		globalTimeout = o.Timeout
	}
	for _, pod := range pods {
		err := o.deletePod(pod)
		if err != nil {
			return err
		}
	}
	_, err := o.waitForDelete(pods, kubectl.Interval, globalTimeout, false, getPodFn)
	return err
}

func (o *DrainOptions) waitForDelete(pods []api.Pod, interval, timeout time.Duration, usingEviction bool, getPodFn func(string, string) (*api.Pod, error)) ([]api.Pod, error) {
	var verbStr string
	if usingEviction {
		verbStr = "evicted"
	} else {
		verbStr = "deleted"
	}
	err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		pendingPods := []api.Pod{}
		for i, pod := range pods {
			p, err := getPodFn(pod.Namespace, pod.Name)
			if apierrors.IsNotFound(err) || (p != nil && p.ObjectMeta.UID != pod.ObjectMeta.UID) {
				glog.V(3).Infof("pod %q %s", pod.Name, verbStr)
				continue
			} else if err != nil {
				return false, err
			} else {
				pendingPods = append(pendingPods, pods[i])
			}
		}
		pods = pendingPods
		if len(pendingPods) > 0 {
			return false, nil
		}
		return true, nil
	})
	return pods, err
}

// SupportEviction uses Discovery API to find out if the server support eviction subresource
// If support, it will return its groupVersion; Otherwise, it will return ""
func SupportEviction(clientset *internalclientset.Clientset) (string, error) {
	discoveryClient := clientset.Discovery()
	groupList, err := discoveryClient.ServerGroups()
	if err != nil {
		return "", err
	}
	foundPolicyGroup := false
	var policyGroupVersion string
	for _, group := range groupList.Groups {
		if group.Name == "policy" {
			foundPolicyGroup = true
			policyGroupVersion = group.PreferredVersion.GroupVersion
			break
		}
	}
	if !foundPolicyGroup {
		return "", nil
	}
	resourceList, err := discoveryClient.ServerResourcesForGroupVersion("v1")
	if err != nil {
		return "", err
	}
	for _, r := range resourceList.APIResources {
		if r.Name == EvictionSubresource && r.Kind == EvictionKind {
			return policyGroupVersion, nil
		}
	}
	return "", nil
}

// RunCordonOrUncordon runs either Cordon or Uncordon.  The desired value for
// "Unschedulable" is passed as the first arg.
func (o *DrainOptions) RunCordonOrUncordon(desired bool) error {
	cmdNamespace, _, err := o.factory.DefaultNamespace()
	if err != nil {
		return err
	}

	if o.nodeInfo.Mapping.GroupVersionKind.Kind == "Node" {
		unsched := reflect.ValueOf(o.nodeInfo.Object).Elem().FieldByName("Spec").FieldByName("Unschedulable")
		if unsched.Bool() == desired {
			glog.V(3).Infof("node cordon or uncordon %q %q %q", o.nodeInfo.Mapping.Resource, o.nodeInfo.Name, already(desired))
		} else {
			helper := resource.NewHelper(o.restClient, o.nodeInfo.Mapping)
			unsched.SetBool(desired)
			var err error
			for i := 0; i < kMaxNodeUpdateRetry; i++ {
				// We don't care about what previous versions may exist, we always want
				// to overwrite, and Replace always sets current ResourceVersion if version is "".
				helper.Versioner.SetResourceVersion(o.nodeInfo.Object, "")
				_, err = helper.Replace(cmdNamespace, o.nodeInfo.Name, true, o.nodeInfo.Object)
				if err != nil {
					if !apierrors.IsConflict(err) {
						return err
					}
				} else {
					break
				}
				// It's a race, no need to sleep
			}
			if err != nil {
				return err
			}
			glog.V(3).Infof("node cordon or uncordon %q %q %q", o.nodeInfo.Mapping.Resource, o.nodeInfo.Name, changed(desired))
		}
	} else {
		glog.V(3).Infof("node cordon or uncordon %q %q %q", o.nodeInfo.Mapping.Resource, o.nodeInfo.Name, "skipped")
	}

	return nil
}

// already() and changed() return suitable strings for {un,}cordoning

func already(desired bool) string {
	if desired {
		return "already cordoned"
	}
	return "already uncordoned"
}

func changed(desired bool) string {
	if desired {
		return "cordoned"
	}
	return "uncordoned"
}
