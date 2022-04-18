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

package validation

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/pager"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/dns"
)

// ValidationCluster uses a cluster to validate.
type ValidationCluster struct {
	Failures []*ValidationError `json:"failures,omitempty"`

	Nodes []*ValidationNode `json:"nodes,omitempty"`
}

// ValidationError holds a validation failure
type ValidationError struct {
	Kind    string `json:"type,omitempty"`
	Name    string `json:"name,omitempty"`
	Message string `json:"message,omitempty"`
	// The InstanceGroup field is used to indicate which instance group this validation error is coming from
	InstanceGroup *kops.InstanceGroup `json:"instanceGroup,omitempty"`
}

type ClusterValidator interface {
	// Validate validates a k8s cluster
	Validate() (*ValidationCluster, error)
}

type clusterValidatorImpl struct {
	cluster        *kops.Cluster
	cloud          fi.Cloud
	instanceGroups []*kops.InstanceGroup
	host           string
	k8sClient      kubernetes.Interface
}

func (v *ValidationCluster) addError(failure *ValidationError) {
	v.Failures = append(v.Failures, failure)
}

// ValidationNode represents the validation status for a node
type ValidationNode struct {
	Name     string             `json:"name,omitempty"`
	Zone     string             `json:"zone,omitempty"`
	Role     string             `json:"role,omitempty"`
	Hostname string             `json:"hostname,omitempty"`
	Status   v1.ConditionStatus `json:"status,omitempty"`
}

// hasPlaceHolderIP checks if the API DNS has been updated.
func hasPlaceHolderIP(host string) (string, error) {
	apiAddr, err := url.Parse(host)
	if err != nil {
		return "", fmt.Errorf("unable to parse Kubernetes cluster API URL: %v", err)
	}
	hostAddrs, err := net.LookupHost(apiAddr.Hostname())
	if err != nil {
		return "", fmt.Errorf("unable to resolve Kubernetes cluster API URL dns: %v", err)
	}

	sort.Strings(hostAddrs)
	for _, h := range hostAddrs {
		if h == cloudup.PlaceholderIP || h == cloudup.PlaceholderIPv6 {
			return h, nil
		}
	}

	return "", nil
}

func NewClusterValidator(cluster *kops.Cluster, cloud fi.Cloud, instanceGroupList *kops.InstanceGroupList, host string, k8sClient kubernetes.Interface) (ClusterValidator, error) {
	var instanceGroups []*kops.InstanceGroup

	for i := range instanceGroupList.Items {
		ig := &instanceGroupList.Items[i]
		instanceGroups = append(instanceGroups, ig)
	}

	if len(instanceGroups) == 0 {
		return nil, fmt.Errorf("no InstanceGroup objects found")
	}

	return &clusterValidatorImpl{
		cluster:        cluster,
		cloud:          cloud,
		instanceGroups: instanceGroups,
		host:           host,
		k8sClient:      k8sClient,
	}, nil
}

func (v *clusterValidatorImpl) Validate() (*ValidationCluster, error) {
	ctx := context.TODO()

	clusterName := v.cluster.Name
	dnsProvider := kops.ExternalDNSProviderDNSController
	if v.cluster.Spec.ExternalDNS != nil && v.cluster.Spec.ExternalDNS.Provider == kops.ExternalDNSProviderExternalDNS {
		dnsProvider = kops.ExternalDNSProviderExternalDNS
	}

	validation := &ValidationCluster{}

	// Do not use if we are running gossip
	if !dns.IsGossipHostname(clusterName) {
		hasPlaceHolderIPAddress, err := hasPlaceHolderIP(v.host)
		if err != nil {
			return nil, err
		}

		if hasPlaceHolderIPAddress != "" {
			message := fmt.Sprintf("Validation Failed\n\n"+
				"The %[1]v Kubernetes deployment has not updated the Kubernetes cluster's API DNS entry to the correct IP address."+
				"  The API DNS IP address is the placeholder address that kops creates: %[2]v."+
				"  Please wait about 5-10 minutes for a master to start, %[1]v to launch, and DNS to propagate."+
				"  The protokube container and %[1]v deployment logs may contain more diagnostic information."+
				"  Etcd and the API DNS entries must be updated for a kops Kubernetes cluster to start.", dnsProvider, hasPlaceHolderIPAddress)
			validation.addError(&ValidationError{
				Kind:    "dns",
				Name:    "apiserver",
				Message: message,
			})
			return validation, nil
		}
	}

	nodeList, err := v.k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing nodes: %v", err)
	}

	warnUnmatched := false
	cloudGroups, err := v.cloud.GetCloudGroups(v.cluster, v.instanceGroups, warnUnmatched, nodeList.Items)
	if err != nil {
		return nil, err
	}
	readyNodes, nodeInstanceGroupMapping := validation.validateNodes(cloudGroups, v.instanceGroups)

	if err := validation.collectPodFailures(ctx, v.k8sClient, readyNodes, nodeInstanceGroupMapping); err != nil {
		return nil, fmt.Errorf("cannot get pod health for %q: %v", clusterName, err)
	}

	return validation, nil
}

var masterStaticPods = []string{
	"kube-apiserver",
	"kube-controller-manager",
	"kube-scheduler",
}

func (v *ValidationCluster) collectPodFailures(ctx context.Context, client kubernetes.Interface, nodes []v1.Node,
	nodeInstanceGroupMapping map[string]*kops.InstanceGroup,
) error {
	masterWithoutPod := map[string]map[string]bool{}
	nodeByAddress := map[string]string{}

	for _, node := range nodes {
		labels := node.GetLabels()
		if _, found := labels["node-role.kubernetes.io/control-plane"]; found {
			masterWithoutPod[node.Name] = map[string]bool{}
			for _, pod := range masterStaticPods {
				masterWithoutPod[node.Name][pod] = true
			}
		}

		for _, nodeAddress := range node.Status.Addresses {
			nodeByAddress[nodeAddress.Address] = node.Name
		}
	}

	err := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
		return client.CoreV1().Pods(metav1.NamespaceAll).List(ctx, opts)
	})).EachListItem(context.TODO(), metav1.ListOptions{}, func(obj runtime.Object) error {
		pod := obj.(*v1.Pod)

		app := pod.GetLabels()["k8s-app"]
		if pod.Namespace == "kube-system" && masterWithoutPod[nodeByAddress[pod.Status.HostIP]][app] {
			delete(masterWithoutPod[nodeByAddress[pod.Status.HostIP]], app)
		}

		priority := pod.Spec.PriorityClassName
		if priority != "system-cluster-critical" && priority != "system-node-critical" {
			return nil
		}
		if pod.Status.Phase == v1.PodSucceeded {
			return nil
		}

		var podNode *kops.InstanceGroup
		if priority == "system-node-critical" {
			podNode = nodeInstanceGroupMapping[nodeByAddress[pod.Status.HostIP]]
		}

		if pod.Status.Phase == v1.PodPending {
			v.addError(&ValidationError{
				Kind:          "Pod",
				Name:          pod.Namespace + "/" + pod.Name,
				Message:       fmt.Sprintf("%s pod %q is pending", priority, pod.Name),
				InstanceGroup: podNode,
			})
			return nil
		}
		if pod.Status.Phase == v1.PodUnknown {
			v.addError(&ValidationError{
				Kind:          "Pod",
				Name:          pod.Namespace + "/" + pod.Name,
				Message:       fmt.Sprintf("%s pod %q is unknown phase", priority, pod.Name),
				InstanceGroup: podNode,
			})
			return nil
		}
		var notready []string
		for _, container := range pod.Status.ContainerStatuses {
			if !container.Ready {
				notready = append(notready, container.Name)
			}
		}
		if len(notready) != 0 {
			v.addError(&ValidationError{
				Kind:          "Pod",
				Name:          pod.Namespace + "/" + pod.Name,
				Message:       fmt.Sprintf("%s pod %q is not ready (%s)", priority, pod.Name, strings.Join(notready, ",")),
				InstanceGroup: podNode,
			})
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error listing Pods: %v", err)
	}

	for node, nodeMap := range masterWithoutPod {
		for app := range nodeMap {
			v.addError(&ValidationError{
				Kind:          "Node",
				Name:          node,
				Message:       fmt.Sprintf("master %q is missing %s pod", node, app),
				InstanceGroup: nodeInstanceGroupMapping[node],
			})
		}
	}

	return nil
}

func (v *ValidationCluster) validateNodes(cloudGroups map[string]*cloudinstances.CloudInstanceGroup, groups []*kops.InstanceGroup) ([]v1.Node, map[string]*kops.InstanceGroup) {
	var readyNodes []v1.Node
	groupsSeen := map[string]bool{}
	nodeInstanceGroupMapping := map[string]*kops.InstanceGroup{}

	for _, cloudGroup := range cloudGroups {
		var allMembers []*cloudinstances.CloudInstance
		allMembers = append(allMembers, cloudGroup.Ready...)
		allMembers = append(allMembers, cloudGroup.NeedUpdate...)

		groupsSeen[cloudGroup.InstanceGroup.Name] = true
		numNodes := 0
		for _, m := range allMembers {
			if m.Status != cloudinstances.CloudInstanceStatusDetached {
				numNodes++
			}
		}
		if numNodes < cloudGroup.TargetSize {
			v.addError(&ValidationError{
				Kind: "InstanceGroup",
				Name: cloudGroup.InstanceGroup.Name,
				Message: fmt.Sprintf("InstanceGroup %q did not have enough nodes %d vs %d",
					cloudGroup.InstanceGroup.Name,
					numNodes,
					cloudGroup.TargetSize),
				InstanceGroup: cloudGroup.InstanceGroup,
			})
		}

		for _, member := range allMembers {
			node := member.Node

			if node == nil {
				nodeExpectedToJoin := true
				if cloudGroup.InstanceGroup.Spec.Role == kops.InstanceGroupRoleBastion {
					// bastion nodes don't join the cluster
					nodeExpectedToJoin = false
				}
				if member.State == cloudinstances.WarmPool {
					nodeExpectedToJoin = false
				}

				if member.Status == cloudinstances.CloudInstanceStatusDetached {
					nodeExpectedToJoin = false
				}

				if nodeExpectedToJoin {
					v.addError(&ValidationError{
						Kind:          "Machine",
						Name:          member.ID,
						Message:       fmt.Sprintf("machine %q has not yet joined cluster", member.ID),
						InstanceGroup: cloudGroup.InstanceGroup,
					})
				}
				continue
			}

			nodeInstanceGroupMapping[node.Name] = cloudGroup.InstanceGroup

			role := strings.ToLower(string(cloudGroup.InstanceGroup.Spec.Role))
			if role == "" {
				role = "node"
			}

			n := &ValidationNode{
				Name:     node.Name,
				Zone:     node.ObjectMeta.Labels["topology.kubernetes.io/zone"],
				Hostname: node.ObjectMeta.Labels["kubernetes.io/hostname"],
				Role:     role,
				Status:   getNodeReadyStatus(node),
			}
			if n.Zone == "" {
				n.Zone = node.ObjectMeta.Labels["failure-domain.beta.kubernetes.io/zone"]
			}

			ready := isNodeReady(node)
			if ready {
				readyNodes = append(readyNodes, *node)
			}

			switch n.Role {
			case "master", "apiserver", "node":
				if !ready {
					v.addError(&ValidationError{
						Kind:          "Node",
						Name:          node.Name,
						Message:       fmt.Sprintf("node %q of role %q is not ready", node.Name, n.Role),
						InstanceGroup: cloudGroup.InstanceGroup,
					})
				}

				v.Nodes = append(v.Nodes, n)
			default:
				klog.Warningf("ignoring node with role %q", n.Role)

			}
		}
	}

	for _, ig := range groups {
		if !groupsSeen[ig.Name] {
			v.addError(&ValidationError{
				Kind:          "InstanceGroup",
				Name:          ig.Name,
				Message:       fmt.Sprintf("InstanceGroup %q is missing from the cloud provider", ig.Name),
				InstanceGroup: ig,
			})
		}
	}

	return readyNodes, nodeInstanceGroupMapping
}
