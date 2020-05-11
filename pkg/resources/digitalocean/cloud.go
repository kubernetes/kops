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

package digitalocean

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
	"k8s.io/klog"

	v1 "k8s.io/api/core/v1"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/resources/digitalocean/dns"
	"k8s.io/kops/protokube/pkg/etcd"
	"k8s.io/kops/upup/pkg/fi"
)

const TagKubernetesClusterIndex = "k8s-index"
const TagKubernetesClusterNamePrefix = "KubernetesCluster"

// TokenSource implements oauth2.TokenSource
type TokenSource struct {
	AccessToken string
}

// Token() returns oauth2.Token
func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

// Cloud exposes all the interfaces required to operate on DigitalOcean resources
type Cloud struct {
	Client *godo.Client

	dns dnsprovider.Interface

	// RegionName holds the region, renamed to avoid conflict with Region()
	RegionName string
}

var _ fi.Cloud = &Cloud{}

// NewCloud returns a Cloud, expecting the env var DIGITALOCEAN_ACCESS_TOKEN
// NewCloud will return an err if DIGITALOCEAN_ACCESS_TOKEN is not defined
func NewCloud(region string) (*Cloud, error) {
	accessToken := os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
	if accessToken == "" {
		return nil, errors.New("DIGITALOCEAN_ACCESS_TOKEN is required")
	}

	tokenSource := &TokenSource{
		AccessToken: accessToken,
	}

	oauthClient := oauth2.NewClient(context.TODO(), tokenSource)
	client := godo.NewClient(oauthClient)

	return &Cloud{
		Client:     client,
		dns:        dns.NewProvider(client),
		RegionName: region,
	}, nil
}

// GetCloudGroups is not implemented yet, that needs to return the instances and groups that back a kops cluster.
func (c *Cloud) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	klog.V(8).Info("digitalocean cloud provider GetCloudGroups not implemented yet")
	return nil, fmt.Errorf("digital ocean cloud provider does not support getting cloud groups at this time")
}

// DeleteGroup is not implemented yet, is a func that needs to delete a DO instance group.
func (c *Cloud) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	klog.V(8).Info("digitalocean cloud provider DeleteGroup not implemented yet")
	return fmt.Errorf("digital ocean cloud provider does not support deleting cloud groups at this time")
}

// DeleteInstance is not implemented yet, is func needs to delete a DO instance.
func (c *Cloud) DeleteInstance(i *cloudinstances.CloudInstanceGroupMember) error {
	klog.V(8).Info("digitalocean cloud provider DeleteInstance not implemented yet")
	return fmt.Errorf("digital ocean cloud provider does not support deleting cloud instances at this time")
}

// DetachInstance is not implemented yet. It needs to cause a cloud instance to no longer be counted against the group's size limits.
func (c *Cloud) DetachInstance(i *cloudinstances.CloudInstanceGroupMember) error {
	klog.V(8).Info("digitalocean cloud provider DetachInstance not implemented yet")
	return fmt.Errorf("digital ocean cloud provider does not support surging")
}

// ProviderID returns the kops api identifier for DigitalOcean cloud provider
func (c *Cloud) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderDO
}

// Region returns the DO region we will target
func (c *Cloud) Region() string {
	return c.RegionName
}

// DNS returns a DO implementation for dnsprovider.Interface
func (c *Cloud) DNS() (dnsprovider.Interface, error) {
	return c.dns, nil
}

// Volumes returns an implementation of godo.StorageService
func (c *Cloud) Volumes() godo.StorageService {
	return c.Client.Storage
}

// VolumeActions returns an implementation of godo.StorageActionsService
func (c *Cloud) VolumeActions() godo.StorageActionsService {
	return c.Client.StorageActions
}

func (c *Cloud) Droplets() godo.DropletsService {
	return c.Client.Droplets
}

func (c *Cloud) LoadBalancers() godo.LoadBalancersService {
	return c.Client.LoadBalancers
}

func (c *Cloud) GetAllLoadBalancers() ([]godo.LoadBalancer, error) {
	return getAllLoadBalancers(c)
}

// FindVPCInfo is not implemented, it's only here to satisfy the fi.Cloud interface
func (c *Cloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return nil, errors.New("not implemented")
}

func (c *Cloud) GetApiIngressStatus(cluster *kops.Cluster) ([]kops.ApiIngressStatus, error) {
	var ingresses []kops.ApiIngressStatus
	if cluster.Spec.MasterPublicName != "" {
		// Note that this must match Digital Ocean's lb name
		klog.V(2).Infof("Querying DO to find Loadbalancers for API (%q)", cluster.Name)

		loadBalancers, err := getAllLoadBalancers(c)
		if err != nil {
			return nil, fmt.Errorf("LoadBalancers.List returned error: %v", err)
		}

		lbName := "api-" + strings.Replace(cluster.Name, ".", "-", -1)

		for _, lb := range loadBalancers {
			if lb.Name == lbName {
				klog.V(10).Infof("Matching LB name found for API (%q)", cluster.Name)

				if lb.Status != "active" {
					return nil, fmt.Errorf("load-balancer is not yet active (current status: %s)", lb.Status)
				}

				address := lb.IP
				ingresses = append(ingresses, kops.ApiIngressStatus{IP: address})

				return ingresses, nil
			}
		}
	}

	return nil, nil
}

// FindClusterStatus discovers the status of the cluster, by looking for the tagged etcd volumes
func (c *Cloud) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
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
func findEtcdStatus(c *Cloud, cluster *kops.Cluster) ([]kops.EtcdClusterStatus, error) {
	statusMap := make(map[string]*kops.EtcdClusterStatus)
	klog.V(2).Infof("Querying DO for etcd volumes")

	volumes, err := getAllVolumesByRegion(c, c.RegionName)

	if err != nil {
		return nil, fmt.Errorf("error describing volumes: %v", err)
	}

	for _, volume := range volumes {
		volumeID := volume.ID

		etcdClusterName := ""
		var etcdClusterSpec *etcd.EtcdClusterSpec

		for _, myTag := range volume.Tags {
			klog.V(2).Infof("findEtcdStatus status (from cloud): %v", myTag)
			// check if volume belongs to this cluster.
			// tag will be in the format "KubernetesCluster:dev5-k8s-local" (where clusterName is dev5.k8s.local)
			clusterName := strings.Replace(cluster.Name, ".", "-", -1)
			if strings.Contains(myTag, fmt.Sprintf("%s, %s", TagKubernetesClusterNamePrefix, ":", clusterName)) {
				klog.V(2).Infof("findEtcdStatus cluster comparison matched for tag: %v", myTag)
				// this volume belongs to our cluster, add this to our etcdClusterSpec.
				// loop through the tags again and
				for _, volumeTag := range volume.Tags {
					if strings.Contains(volumeTag, TagKubernetesClusterIndex) {
						if len(strings.Split(volumeTag, ":")) < 2 {
							return nil, fmt.Errorf("error splitting the volume tag %q on volume %q", volumeTag, volume)
						}
						dropletIndex := strings.Split(volumeTag, ":")[1]
						etcdClusterSpec, err = c.getEtcdClusterSpec(volume.Name, dropletIndex)
						klog.V(2).Infof("findEtcdStatus etcdClusterSpec: %v", fi.DebugAsJsonString(etcdClusterSpec))
						if err != nil {
							return nil, fmt.Errorf("error parsing etcd cluster tag %q on volume %q: %v", volumeTag, volumeID, err)
						}

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
							VolumeId: volume.ID,
						})
					}
				}
			}
		}
	}

	var status []kops.EtcdClusterStatus
	for _, v := range statusMap {
		status = append(status, *v)
	}

	return status, nil
}

func (c *Cloud) getEtcdClusterSpec(volumeName string, dropletName string) (*etcd.EtcdClusterSpec, error) {
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
