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

package do

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	dns "k8s.io/kops/dnsprovider/pkg/dnsprovider/providers/do"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/protokube/pkg/etcd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

const (
	TagKubernetesClusterIndex        = "k8s-index"
	TagNameEtcdClusterPrefix         = "etcdCluster-"
	TagNameRolePrefix                = "k8s.io/role/"
	TagKubernetesClusterNamePrefix   = "KubernetesCluster"
	TagKubernetesClusterMasterPrefix = "KubernetesCluster-Master"
	TagKubernetesInstanceGroup       = "kops-instancegroup"
)

type DOInstanceGroup struct {
	ClusterName       string
	InstanceGroupName string
	GroupType         string   // will be either "master" or "worker"
	Members           []string // will store the droplet names that matches.
}

// TokenSource implements oauth2.TokenSource
type TokenSource struct {
	AccessToken string
}

// DOCloud exposes all the interfaces required to operate on DigitalOcean resources
type DOCloud interface {
	fi.Cloud
	DropletsService() godo.DropletsService
	DropletActionService() godo.DropletActionsService
	VolumeService() godo.StorageService
	VolumeActionService() godo.StorageActionsService
	LoadBalancersService() godo.LoadBalancersService
	DomainService() godo.DomainsService
	ActionsService() godo.ActionsService
	VPCsService() godo.VPCsService
	FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error)
	GetAllLoadBalancers() ([]godo.LoadBalancer, error)
	GetAllDropletsByTag(tag string) ([]godo.Droplet, error)
	GetAllVolumesByRegion() ([]godo.Volume, error)
	GetVPCUUID(networkCIDR string, vpcName string) (string, error)
	GetAllVPCs() ([]*godo.VPC, error)
}

var readBackoff = wait.Backoff{
	Duration: time.Second,
	Factor:   2,
	Jitter:   0.1,
	Steps:    10,
}

// static compile time check to validate DOCloud's fi.Cloud Interface.
var _ fi.Cloud = &doCloudImplementation{}

// doCloudImplementation holds the godo client object to interact with DO resources.
type doCloudImplementation struct {
	Client *godo.Client

	dns dnsprovider.Interface

	// region holds the DO region.
	region string
}

// Token() returns oauth2.Token
func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

// NewCloud returns a Cloud, expecting the env var DIGITALOCEAN_ACCESS_TOKEN
// NewCloud will return an err if DIGITALOCEAN_ACCESS_TOKEN is not defined
func NewDOCloud(region string) (DOCloud, error) {
	accessToken := os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
	if accessToken == "" {
		return nil, errors.New("DIGITALOCEAN_ACCESS_TOKEN is required")
	}

	tokenSource := &TokenSource{
		AccessToken: accessToken,
	}

	oauthClient := oauth2.NewClient(context.TODO(), tokenSource)
	client := godo.NewClient(oauthClient)

	return &doCloudImplementation{
		Client: client,
		dns:    dns.NewProvider(client),
		region: region,
	}, nil
}

func (c *doCloudImplementation) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	return getCloudGroups(c, cluster, instancegroups, warnUnmatched, nodes)
}

// DeleteGroup is not implemented yet, is a func that needs to delete a DO instance group.
func (c *doCloudImplementation) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	klog.V(8).Info("digitalocean cloud provider DeleteGroup not implemented yet")
	return fmt.Errorf("digital ocean cloud provider does not support deleting cloud groups at this time")
}

// DeregisterInstance drains a cloud instance and loadbalancers.
func (c *doCloudImplementation) DeregisterInstance(i *cloudinstances.CloudInstance) error {
	klog.V(8).Info("DO DeregisterInstance not implemented")
	return nil
}

func (c *doCloudImplementation) DeleteInstance(i *cloudinstances.CloudInstance) error {
	dropletID, err := strconv.Atoi(i.ID)
	if err != nil {
		return fmt.Errorf("failed to convert droplet ID to int: %s", err)
	}

	_, _, err = c.Client.DropletActions.Shutdown(context.TODO(), dropletID)
	if err != nil {
		return fmt.Errorf("error stopping instance %q: %v", dropletID, err)
	}

	// Wait for 5 min to stop the instance
	for i := 0; i < 5; i++ {
		droplet, _, err := c.Client.Droplets.Get(context.TODO(), dropletID)
		if err != nil {
			return fmt.Errorf("error describing instance %q: %v", dropletID, err)
		}

		klog.V(8).Infof("stopping DO instance %q, current Status: %q", droplet, droplet.Status)

		if droplet.Status == "off" {
			break
		}

		if i == 5 {
			return fmt.Errorf("fail to stop DO instance %v in 5 mins", dropletID)
		}

		time.Sleep(time.Minute * 1)
	}

	_, err = c.Client.Droplets.Delete(context.TODO(), dropletID)
	if err != nil {
		return fmt.Errorf("error stopping instance %q: %v", dropletID, err)
	}

	klog.V(8).Infof("deleted droplet instance %q", dropletID)

	return nil
}

// DetachInstance is not implemented yet. It needs to cause a cloud instance to no longer be counted against the group's size limits.
func (c *doCloudImplementation) DetachInstance(i *cloudinstances.CloudInstance) error {
	klog.V(8).Info("digitalocean cloud provider DetachInstance not implemented yet")
	return fmt.Errorf("digital ocean cloud provider does not support surging")
}

// ProviderID returns the kops api identifier for DigitalOcean cloud provider
func (c *doCloudImplementation) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderDO
}

// Region returns the DO region we will target
func (c *doCloudImplementation) Region() string {
	return c.region
}

func (c *doCloudImplementation) DNS() (dnsprovider.Interface, error) {
	provider, err := dnsprovider.GetDnsProvider(dns.ProviderName, nil)
	if err != nil {
		return nil, fmt.Errorf("error building DNS provider: %v", err)
	}
	return provider, nil
}

// Volumes returns an implementation of godo.StorageService
func (c *doCloudImplementation) VolumeService() godo.StorageService {
	return c.Client.Storage
}

// VolumeActions returns an implementation of godo.StorageActionsService
func (c *doCloudImplementation) VolumeActionService() godo.StorageActionsService {
	return c.Client.StorageActions
}

// DropletsService returns the droplets client interface.
func (c *doCloudImplementation) DropletsService() godo.DropletsService {
	return c.Client.Droplets
}

func (c *doCloudImplementation) DropletActionService() godo.DropletActionsService {
	return c.Client.DropletActions
}

func (c *doCloudImplementation) LoadBalancersService() godo.LoadBalancersService {
	return c.Client.LoadBalancers
}

func (c *doCloudImplementation) DomainService() godo.DomainsService {
	return c.Client.Domains
}

func (c *doCloudImplementation) ActionsService() godo.ActionsService {
	return c.Client.Actions
}

func (c *doCloudImplementation) VPCsService() godo.VPCsService {
	return c.Client.VPCs
}

// FindVPCInfo is not implemented, it's only here to satisfy the fi.Cloud interface
func (c *doCloudImplementation) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return nil, errors.New("not implemented")
}

func (c *doCloudImplementation) GetVPCUUID(networkCIDR string, vpcName string) (string, error) {
	vpcUUID := ""
	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		vpcs, err := c.GetAllVPCs()
		if err != nil {
			return false, err
		}

		for _, vpc := range vpcs {
			if vpc.IPRange == networkCIDR && vpc.Name == vpcName {
				vpcUUID = vpc.ID
				return true, nil
			}
		}

		return false, fmt.Errorf("vpc not yet created..")
	})

	if err != nil {
		return "", err
	}

	if done {
		return vpcUUID, nil
	} else {
		return "", wait.ErrWaitTimeout
	}
}

func (c *doCloudImplementation) GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	var ingresses []fi.ApiIngressStatus
	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		if cluster.Spec.MasterPublicName != "" {
			// Note that this must match Digital Ocean's lb name
			klog.V(2).Infof("Querying DO to find Loadbalancers for API (%q)", cluster.Name)

			loadBalancers, err := c.GetAllLoadBalancers()
			if err != nil {
				return false, fmt.Errorf("LoadBalancers.List returned error: %v", err)
			}

			lbName := "api-" + strings.Replace(cluster.Name, ".", "-", -1)

			for _, lb := range loadBalancers {
				if lb.Name == lbName {
					klog.V(10).Infof("Matching LB name found for API (%q)", cluster.Name)

					if lb.Status != "active" {
						return false, fmt.Errorf("load-balancer is not yet active (current status: %s)", lb.Status)
					}

					address := lb.IP
					ingresses = append(ingresses, fi.ApiIngressStatus{IP: address})
				}
			}
		}
		return true, nil
	})
	if done {
		return ingresses, nil
	} else {
		if err == nil {
			err = wait.ErrWaitTimeout
		}
		return ingresses, err
	}
}

// FindClusterStatus discovers the status of the cluster, by looking for the tagged etcd volumes
func (c *doCloudImplementation) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	etcdStatus, err := findEtcdStatus(c, cluster)
	if err != nil {
		return nil, err
	}
	status := &kops.ClusterStatus{
		EtcdClusters: etcdStatus,
	}
	klog.V(2).Infof("Cluster status (from cloud): %v", fi.DebugAsJsonString(status))
	return status, nil
}

// findEtcdStatus discovers the status of etcd, by looking for the tagged etcd volumes
func findEtcdStatus(c *doCloudImplementation, cluster *kops.Cluster) ([]kops.EtcdClusterStatus, error) {
	statusMap := make(map[string]*kops.EtcdClusterStatus)
	volumes, err := c.GetAllVolumesByRegion()
	if err != nil {
		return nil, fmt.Errorf("failed to get all volumes by region from %s: %v", c.Region(), err)
	}

	for _, volume := range volumes {
		volumeID := volume.ID

		etcdClusterName := ""
		var etcdClusterSpec *etcd.EtcdClusterSpec

		for _, myTag := range volume.Tags {
			klog.V(8).Infof("findEtcdStatus status (from cloud): checking if volume with tag %q belongs to cluster", myTag)
			// check if volume belongs to this cluster.
			// tag will be in the format "KubernetesCluster:dev5-k8s-local" (where clusterName is dev5.k8s.local)
			clusterName := strings.Replace(cluster.Name, ".", "-", -1)
			if strings.Contains(myTag, fmt.Sprintf("%s:%s", TagKubernetesClusterNamePrefix, clusterName)) {
				klog.V(10).Infof("findEtcdStatus cluster comparison matched for tag: %v", myTag)
				// this volume belongs to our cluster, add this to our etcdClusterSpec.
				// loop through the tags again and
				for _, volumeTag := range volume.Tags {
					if strings.Contains(volumeTag, TagKubernetesClusterIndex) {
						volumeTagParts := strings.Split(volumeTag, ":")
						if len(volumeTagParts) < 2 {
							return nil, fmt.Errorf("volume tag split failed, too few components for tag %q on volume %q", volumeTag, volume)
						}
						dropletIndex := volumeTagParts[1]
						etcdClusterSpec, err = c.getEtcdClusterSpec(volume.Name, dropletIndex)
						if err != nil {
							return nil, fmt.Errorf("error parsing etcd cluster tag %q on volume %q: %v", volumeTag, volumeID, err)
						}

						klog.V(10).Infof("findEtcdStatus etcdClusterSpec: %v", fi.DebugAsJsonString(etcdClusterSpec))
						etcdClusterName = etcdClusterSpec.ClusterKey
						status := statusMap[etcdClusterName]
						if status == nil {
							status = &kops.EtcdClusterStatus{
								Name: etcdClusterName,
							}
							statusMap[etcdClusterName] = status
						}

						memberName := etcdClusterSpec.NodeName
						status.Members = append(status.Members, &kops.EtcdMemberStatus{
							Name:     memberName,
							VolumeID: volume.ID,
						})
					}
				}
			}
		}
	}

	status := make([]kops.EtcdClusterStatus, 0, len(statusMap))
	for _, v := range statusMap {
		status = append(status, *v)
	}

	return status, nil
}

func (c *doCloudImplementation) getEtcdClusterSpec(volumeName string, dropletName string) (*etcd.EtcdClusterSpec, error) {
	var clusterKey string
	if strings.Contains(volumeName, "etcd-main") {
		clusterKey = "main"
	} else if strings.Contains(volumeName, "etcd-events") {
		clusterKey = "events"
	} else {
		return nil, fmt.Errorf("could not determine etcd cluster type for volume: %s", volumeName)
	}

	return &etcd.EtcdClusterSpec{
		ClusterKey: clusterKey,
		NodeName:   dropletName,
		NodeNames:  []string{dropletName},
	}, nil
}

func getCloudGroups(c *doCloudImplementation, cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	nodeMap := cloudinstances.GetNodeMap(nodes, cluster)

	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	instanceGroups, err := findInstanceGroups(c, cluster.ObjectMeta.Name)
	if err != nil {
		return nil, fmt.Errorf("unable to find autoscale groups: %v", err)
	}

	for _, doGroup := range instanceGroups {
		name := doGroup.InstanceGroupName

		instancegroup, err := matchInstanceGroup(name, cluster.ObjectMeta.Name, instancegroups)
		if err != nil {
			return nil, fmt.Errorf("error getting instance group for doGroup %q", name)
		}
		if instancegroup == nil {
			if warnUnmatched {
				klog.Warningf("Found doGroup with no corresponding instance group %q", name)
			}
			continue
		}

		groups[instancegroup.ObjectMeta.Name], err = buildCloudInstanceGroup(c, instancegroup, doGroup, nodeMap)
		if err != nil {
			return nil, fmt.Errorf("error getting cloud instance group %q: %v", instancegroup.ObjectMeta.Name, err)
		}
	}

	klog.V(8).Infof("Cloud Instance Group Info = %v", groups)
	return groups, nil
}

// findInstanceGroups finds instance groups matching the specified tags
func findInstanceGroups(c *doCloudImplementation, clusterName string) ([]DOInstanceGroup, error) {
	var result []DOInstanceGroup
	instanceGroupMap := make(map[string][]string) // map of instance group name with droplet ids

	clusterTag := "KubernetesCluster:" + strings.Replace(clusterName, ".", "-", -1)
	droplets, err := c.GetAllDropletsByTag(clusterTag)
	if err != nil {
		return nil, fmt.Errorf("get all droplets for tag %s returned error. Error=%v", clusterTag, err)
	}

	instanceGroupName := ""
	for _, droplet := range droplets {
		doInstanceGroup, err := getDropletInstanceGroup(droplet.Tags)
		if err != nil {
			return nil, fmt.Errorf("get droplets Instance group for tags %v returned error. Error=%v", droplet.Tags, err)
		}

		instanceGroupName = fmt.Sprintf("%s-%s", clusterName, doInstanceGroup)
		instanceGroupMap[instanceGroupName] = append(instanceGroupMap[instanceGroupName], strconv.Itoa(droplet.ID))

		result = append(result, DOInstanceGroup{
			InstanceGroupName: instanceGroupName,
			GroupType:         instanceGroupName,
			ClusterName:       clusterName,
			Members:           instanceGroupMap[instanceGroupName],
		})
	}

	klog.V(8).Infof("InstanceGroup Info = %v", result)

	return result, nil
}

func getDropletInstanceGroup(tags []string) (string, error) {
	for _, tag := range tags {
		klog.V(8).Infof("Check tag = %s", tag)
		if strings.Contains(strings.ToLower(tag), TagKubernetesInstanceGroup) {
			tagParts := strings.Split(tag, ":")
			if len(tagParts) < 2 {
				return "", fmt.Errorf("tag split failed, too few components for tag %q", tag)
			}
			return tagParts[1], nil
		}
	}

	return "", fmt.Errorf("Didn't find k8s-instancegroup for tag %v", tags)
}

// matchInstanceGroup filters a list of instancegroups for recognized cloud groups
func matchInstanceGroup(name string, clusterName string, instancegroups []*kops.InstanceGroup) (*kops.InstanceGroup, error) {
	var instancegroup *kops.InstanceGroup
	for _, g := range instancegroups {
		var groupName string

		switch g.Spec.Role {
		case kops.InstanceGroupRoleMaster, kops.InstanceGroupRoleNode:
			groupName = clusterName + "-" + g.ObjectMeta.Name
		default:
			klog.Warningf("Ignoring InstanceGroup of unknown role %q", g.Spec.Role)
			continue
		}

		if name == groupName {
			if instancegroup != nil {
				return nil, fmt.Errorf("found multiple instance groups matching servergrp %q", groupName)
			}
			instancegroup = g
		}
	}

	return instancegroup, nil
}

func buildCloudInstanceGroup(c *doCloudImplementation, ig *kops.InstanceGroup, g DOInstanceGroup, nodeMap map[string]*v1.Node) (*cloudinstances.CloudInstanceGroup, error) {
	cg := &cloudinstances.CloudInstanceGroup{
		HumanName:     g.InstanceGroupName,
		InstanceGroup: ig,
		Raw:           g,
		MinSize:       int(fi.Int32Value(ig.Spec.MinSize)),
		TargetSize:    int(fi.Int32Value(ig.Spec.MinSize)),
		MaxSize:       int(fi.Int32Value(ig.Spec.MaxSize)),
	}

	for _, member := range g.Members {

		// TODO use a hash of the godo.DropletCreateRequest fields to calculate the second parameter.
		_, err := cg.NewCloudInstance(member, cloudinstances.CloudInstanceStatusUpToDate, nodeMap[member])
		if err != nil {
			return nil, fmt.Errorf("error creating cloud instance group member: %v", err)
		}
	}

	return cg, nil
}

func (c *doCloudImplementation) GetAllLoadBalancers() ([]godo.LoadBalancer, error) {
	allLoadBalancers := []godo.LoadBalancer{}

	opt := &godo.ListOptions{}
	for {
		lbs, resp, err := c.LoadBalancersService().List(context.TODO(), opt)
		if err != nil {
			return nil, err
		}

		allLoadBalancers = append(allLoadBalancers, lbs...)

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		opt.Page = page + 1
	}

	return allLoadBalancers, nil
}

func (c *doCloudImplementation) GetAllVPCs() ([]*godo.VPC, error) {
	allVPCs := []*godo.VPC{}

	opt := &godo.ListOptions{}
	for {
		vpcs, resp, err := c.VPCsService().List(context.TODO(), opt)
		if err != nil {
			return nil, err
		}

		allVPCs = append(allVPCs, vpcs...)

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		opt.Page = page + 1
	}

	return allVPCs, nil
}

func (c *doCloudImplementation) GetAllDropletsByTag(tag string) ([]godo.Droplet, error) {
	allDroplets := []godo.Droplet{}

	opt := &godo.ListOptions{}
	for {
		droplets, resp, err := c.DropletsService().ListByTag(context.TODO(), tag, opt)
		if err != nil {
			return nil, err
		}

		allDroplets = append(allDroplets, droplets...)

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		opt.Page = page + 1
	}

	return allDroplets, nil
}

func (c *doCloudImplementation) GetAllVolumesByRegion() ([]godo.Volume, error) {
	allVolumes := []godo.Volume{}

	opt := &godo.ListOptions{}
	for {
		volumes, resp, err := c.VolumeService().ListVolumes(context.TODO(), &godo.ListVolumeParams{
			Region:      c.Region(),
			ListOptions: opt,
		})
		if err != nil {
			return nil, err
		}

		allVolumes = append(allVolumes, volumes...)

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		opt.Page = page + 1
	}

	return allVolumes, nil
}
