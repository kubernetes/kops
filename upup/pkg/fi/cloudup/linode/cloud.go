/*
Copyright 2026 The Kubernetes Authors.

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

package linode

import (
	"context"
	"fmt"
	"net"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/linode/linodego"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/truncate"
	"k8s.io/kops/upup/pkg/fi"
)

// LinodeClient is the Linode (Akamai) API surface used by cloudup tasks.
type LinodeClient interface {
	ListSSHKeys(ctx context.Context, opts *linodego.ListOptions) ([]linodego.SSHKey, error)
	CreateSSHKey(ctx context.Context, opts linodego.SSHKeyCreateOptions) (*linodego.SSHKey, error)
	DeleteSSHKey(ctx context.Context, keyID int) error
	ListInstances(ctx context.Context, opts *linodego.ListOptions) ([]linodego.Instance, error)
	CreateInstance(ctx context.Context, opts linodego.InstanceCreateOptions) (*linodego.Instance, error)
	DeleteInstance(ctx context.Context, linodeID int) error
	UpdateInstance(ctx context.Context, linodeID int, opts linodego.InstanceUpdateOptions) (*linodego.Instance, error)
	ListVolumes(ctx context.Context, opts *linodego.ListOptions) ([]linodego.Volume, error)
	CreateVolume(ctx context.Context, opts linodego.VolumeCreateOptions) (*linodego.Volume, error)
	DeleteVolume(ctx context.Context, volumeID int) error
	ListNodeBalancers(ctx context.Context, opts *linodego.ListOptions) ([]linodego.NodeBalancer, error)
	GetNodeBalancer(ctx context.Context, nodebalancerID int) (*linodego.NodeBalancer, error)
	CreateNodeBalancer(ctx context.Context, opts linodego.NodeBalancerCreateOptions) (*linodego.NodeBalancer, error)
	DeleteNodeBalancer(ctx context.Context, nodebalancerID int) error
	ListNodeBalancerConfigs(ctx context.Context, nodebalancerID int, opts *linodego.ListOptions) ([]linodego.NodeBalancerConfig, error)
	CreateNodeBalancerConfig(ctx context.Context, nodebalancerID int, opts linodego.NodeBalancerConfigCreateOptions) (*linodego.NodeBalancerConfig, error)
	RebuildNodeBalancerConfig(ctx context.Context, nodebalancerID int, configID int, opts linodego.NodeBalancerConfigRebuildOptions) (*linodego.NodeBalancerConfig, error)
	ListNodeBalancerNodes(ctx context.Context, nodebalancerID int, configID int, opts *linodego.ListOptions) ([]linodego.NodeBalancerNode, error)
	CreateNodeBalancerNode(ctx context.Context, nodebalancerID int, configID int, opts linodego.NodeBalancerNodeCreateOptions) (*linodego.NodeBalancerNode, error)
	UpdateNodeBalancerNode(ctx context.Context, nodebalancerID int, configID int, nodeID int, opts linodego.NodeBalancerNodeUpdateOptions) (*linodego.NodeBalancerNode, error)
}

// LinodeCloud exposes cloud behavior required by cloudup for Linode (Akamai).
type LinodeCloud interface {
	fi.Cloud
	AccessToken() string
	Client() LinodeClient
}

var _ fi.Cloud = &Cloud{}

// Cloud holds cloud-level behavior and credentials for Linode (Akamai) operations.
type Cloud struct {
	accessToken string
	region      string
	client      LinodeClient
}

const (
	TagKubernetesInstanceGroup = "kops.k8s.io/instance-group"
	TagKubernetesInstanceRole  = "kops.k8s.io/instance-role"
	TagEtcdClusterName         = "kops.k8s.io/etcd"
)

// NewCloud builds a Linode (Akamai) cloud wrapper using LINODE_TOKEN.
func NewCloud(region string) (*Cloud, error) {
	accessToken := os.Getenv("LINODE_TOKEN")
	if accessToken == "" {
		return nil, fmt.Errorf("LINODE_TOKEN is required")
	}
	if region == "" {
		return nil, fmt.Errorf("region is required")
	}

	linodeClient := linodego.NewClient(nil)
	linodeClient.SetUserAgent("kops")
	linodeClient.SetToken(accessToken)

	return &Cloud{
		accessToken: accessToken,
		region:      region,
		client:      &linodeClient,
	}, nil
}

// AccessToken returns the configured Linode (Akamai) API token.
func (c *Cloud) AccessToken() string {
	return c.accessToken
}

// Client returns the Linode (Akamai) API client.
func (c *Cloud) Client() LinodeClient {
	return c.client
}

// ProviderID returns the cloud provider ID for Linode (Akamai).
func (c *Cloud) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderLinode
}

// DNS returns nil as DNS support has not yet been wired for Linode (Akamai).
func (c *Cloud) DNS() (dnsprovider.Interface, error) {
	return nil, nil
}

// FindVPCInfo returns VPC information for the given ID. This is currently not implemented for Linode (Akamai).
func (c *Cloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return nil, nil
}

// DeleteInstance deletes the given Linode (Akamai) instance.
func (c *Cloud) DeleteInstance(instance *cloudinstances.CloudInstance) error {
	if instance == nil {
		return fmt.Errorf("instance is required")
	}
	if c.client == nil {
		return fmt.Errorf("linode client is not configured")
	}

	instanceID, err := strconv.Atoi(instance.ID)
	if err != nil {
		return fmt.Errorf("invalid Linode (Akamai) instance ID %q: %w", instance.ID, err)
	}

	if err := c.client.DeleteInstance(context.Background(), instanceID); err != nil {
		if linodego.IsNotFound(err) {
			klog.V(4).Infof("Linode (Akamai) instance %q was already deleted", instance.ID)
			return nil
		}
		return fmt.Errorf("error deleting Linode (Akamai) instance %q: %w", instance.ID, err)
	}

	return nil
}

// DeregisterInstance deregisters the given Linode (Akamai) instance from the cloud provider.
func (c *Cloud) DeregisterInstance(instance *cloudinstances.CloudInstance) error {
	return nil
}

// DeleteGroup deletes all Linode (Akamai) instances that belong to the given instance group.
func (c *Cloud) DeleteGroup(group *cloudinstances.CloudInstanceGroup) error {
	if group == nil || group.InstanceGroup == nil {
		return fmt.Errorf("instance group is required")
	}
	if c.client == nil {
		return fmt.Errorf("linode client is not configured")
	}

	instances, err := c.client.ListInstances(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("error listing Linode (Akamai) instances for group %q: %w", group.InstanceGroup.Name, err)
	}

	instanceGroupTag := BuildLinodeTag(TagKubernetesInstanceGroup, group.InstanceGroup.Name)
	clusterTag := ""
	if group.InstanceGroup.Labels != nil {
		if clusterName := group.InstanceGroup.Labels[kops.LabelClusterName]; clusterName != "" {
			clusterTag = BuildLinodeTag(kops.LabelClusterName, clusterName)
		}
	}

	for _, instance := range instances {
		if !slices.Contains(instance.Tags, instanceGroupTag) {
			continue
		}
		if clusterTag != "" && !slices.Contains(instance.Tags, clusterTag) {
			continue
		}

		instanceID := strconv.Itoa(instance.ID)
		if err := c.DeleteInstance(&cloudinstances.CloudInstance{ID: instanceID, CloudInstanceGroup: group}); err != nil {
			return fmt.Errorf("error deleting Linode (Akamai) instance %q in group %q: %w", instanceID, group.InstanceGroup.Name, err)
		}
	}

	return nil
}

// DetachInstance removes the role tag from the given Linode (Akamai) instance so that it is no longer
// considered part of a particular role in the cluster.
func (c *Cloud) DetachInstance(instance *cloudinstances.CloudInstance) error {
	if instance == nil {
		return fmt.Errorf("instance is required")
	}
	if c.client == nil {
		return fmt.Errorf("Linode (Akamai) client is not configured")
	}

	instanceID, err := strconv.Atoi(instance.ID)
	if err != nil {
		return fmt.Errorf("invalid Linode (Akamai) instance ID %q: %w", instance.ID, err)
	}

	instances, err := c.client.ListInstances(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("error listing Linode (Akamai) instances for detach %q: %w", instance.ID, err)
	}

	var current *linodego.Instance
	for i := range instances {
		if instances[i].ID == instanceID {
			current = &instances[i]
			break
		}
	}

	if current == nil {
		klog.V(4).Infof("Linode (Akamai) instance %q was already deleted before detach", instance.ID)
		return nil
	}

	updatedTags, removedRoleTag := removeTagsWithPrefix(current.Tags, BuildLinodeTag(TagKubernetesInstanceRole, ""))
	if !removedRoleTag {
		return nil
	}

	if _, err := c.client.UpdateInstance(context.Background(), instanceID, linodego.InstanceUpdateOptions{Tags: &updatedTags}); err != nil {
		if linodego.IsNotFound(err) {
			klog.V(4).Infof("Linode (Akamai) instance %q was already deleted during detach", instance.ID)
			return nil
		}
		return fmt.Errorf("error detaching Linode (Akamai) instance %q: %w", instance.ID, err)
	}

	return nil
}

// GetCloudGroups returns a map of cloud instance groups for the given cluster and instance groups,
// populated with the nodes that belong to each group.
func (c *Cloud) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	nodeMap := cloudinstances.GetNodeMap(nodes, cluster)

	groups := make(map[string]*cloudinstances.CloudInstanceGroup, len(instancegroups))
	for _, ig := range instancegroups {
		groups[ig.Name] = &cloudinstances.CloudInstanceGroup{
			HumanName:     ig.Name,
			InstanceGroup: ig,
			MinSize:       int(fi.ValueOf(ig.Spec.MinSize)),
			TargetSize:    int(fi.ValueOf(ig.Spec.MinSize)),
			MaxSize:       int(fi.ValueOf(ig.Spec.MaxSize)),
		}
	}

	for instanceID, node := range nodeMap {
		igName := ""
		if node.Labels != nil {
			igName = node.Labels[kops.NodeLabelInstanceGroup]
		}
		if igName == "" {
			if warnUnmatched {
				klog.Warningf("node %q does not have the %q label", node.Name, kops.NodeLabelInstanceGroup)
			}
			continue
		}

		group := groups[igName]
		if group == nil {
			if warnUnmatched {
				klog.Warningf("node %q references unknown instance group %q", node.Name, igName)
			}
			continue
		}

		if _, err := group.NewCloudInstance(instanceID, cloudinstances.CloudInstanceStatusUpToDate, node); err != nil {
			return nil, err
		}
	}

	for _, group := range groups {
		group.AdjustNeedUpdate()
	}

	return groups, nil
}

// Region returns the Linode (Akamai) region that this cloud is configured for.
func (c *Cloud) Region() string {
	return c.region
}

// FindClusterStatus returns the current status of the given cluster in Linode (Akamai).
func (c *Cloud) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	return &kops.ClusterStatus{}, nil
}

// GetApiIngressStatus returns the API ingress status for the given cluster in Linode (Akamai).
func (c *Cloud) GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	var ingresses []fi.ApiIngressStatus

	publicName := cluster.Spec.API.PublicName
	if publicName != "" {
		if net.ParseIP(publicName) == nil {
			ingresses = append(ingresses, fi.ApiIngressStatus{Hostname: publicName})
		} else {
			ingresses = append(ingresses, fi.ApiIngressStatus{IP: publicName})
		}

		return ingresses, nil
	}

	if cluster.Spec.API.LoadBalancer != nil && c.client != nil {
		lbName := "api." + cluster.Name
		label := NormalizedLoadBalancerLabel(lbName)

		nbs, err := c.client.ListNodeBalancers(context.Background(), nil)
		if err != nil {
			return nil, fmt.Errorf("error listing Linode (Akamai) load balancers: %w", err)
		}

		for _, nb := range nbs {
			if nb.Label != nil && *nb.Label == label {
				if nb.IPv4 == nil || *nb.IPv4 == "" {
					continue
				}
				ingresses = append(ingresses, fi.ApiIngressStatus{IP: *nb.IPv4})
				return ingresses, nil
			}
		}

		return ingresses, nil
	}

	// Fallback: return control plane instance IPs
	if c.client == nil {
		return ingresses, nil
	}

	instances, err := c.client.ListInstances(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("error listing Linode (Akamai) instances: %w", err)
	}

	clusterTag := BuildLinodeTag(kops.LabelClusterName, cluster.Name)
	controlPlaneTag := BuildLinodeTag(TagKubernetesInstanceRole, string(kops.InstanceGroupRoleControlPlane))
	apiServerTag := BuildLinodeTag(TagKubernetesInstanceRole, string(kops.InstanceGroupRoleAPIServer))

	ipSet := make(map[string]struct{})
	for _, instance := range instances {
		if !slices.Contains(instance.Tags, clusterTag) {
			continue
		}
		if !slices.Contains(instance.Tags, controlPlaneTag) && !slices.Contains(instance.Tags, apiServerTag) {
			continue
		}

		ip := selectPublicIPv4(instance.IPv4)
		if ip == "" {
			continue
		}

		ipSet[ip] = struct{}{}
	}

	if len(ipSet) == 0 {
		return ingresses, nil
	}

	ips := make([]string, 0, len(ipSet))
	for ip := range ipSet {
		ips = append(ips, ip)
	}
	slices.Sort(ips)

	for _, ip := range ips {
		ingresses = append(ingresses, fi.ApiIngressStatus{IP: ip})
	}

	return ingresses, nil
}

// BuildLinodeTag returns a Linode (Akamai) tag in the form "key:value".
func BuildLinodeTag(key, value string) string {
	return key + ":" + value
}

func removeTagsWithPrefix(tags []string, prefix string) ([]string, bool) {
	filtered := make([]string, 0, len(tags))
	removed := false

	for _, tag := range tags {
		if strings.HasPrefix(tag, prefix) {
			removed = true
			continue
		}
		filtered = append(filtered, tag)
	}

	return filtered, removed
}

// SelectIPv4 returns the first preferred IPv4 from ips. When preferPrivate is true
// the first private IPv4 is returned; otherwise the first public (non-private,
// non-loopback, non-link-local) IPv4. Falls back to the first any-IPv4.
func SelectIPv4(ips []*net.IP, preferPrivate bool) string {
	var fallback string

	for _, ipPtr := range ips {
		if ipPtr == nil {
			continue
		}

		ip := *ipPtr
		if ip.To4() == nil {
			continue
		}

		if fallback == "" {
			fallback = ip.String()
		}

		if preferPrivate {
			if ip.IsPrivate() {
				return ip.String()
			}
		} else {
			if !ip.IsPrivate() && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() {
				return ip.String()
			}
		}
	}

	return fallback
}

func selectPublicIPv4(ips []*net.IP) string {
	return SelectIPv4(ips, false)
}

// NormalizedLoadBalancerLabel returns a normalized label for a NodeBalancer name.
// NodeBalancer labels must match [A-Za-z0-9_-]+ and be 32 chars or fewer.
func NormalizedLoadBalancerLabel(name string) string {
	label := invalidLinodeLabelChars.ReplaceAllString(name, "-")
	label = strings.Trim(label, "-_")
	if label == "" {
		return "kops-api"
	}

	if len(label) > 32 {
		label = label[:32]
		label = strings.Trim(label, "-_")
		if label == "" {
			return "kops-api"
		}
	}

	return label
}

var invalidLinodeLabelChars = regexp.MustCompile(`[^A-Za-z0-9_-]+`)

// NormalizeLinodeSSHKeyLabel returns a normalized label for a Linode (Akamai) SSH key.
// SSH key labels must match [A-Za-z0-9_-]+ and be 64 chars or fewer.
func NormalizeLinodeSSHKeyLabel(name string) string {
	name = invalidLinodeLabelChars.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-_")
	if name == "" {
		return "kubernetes-ssh-key"
	}

	name = truncate.TruncateString(name, truncate.TruncateStringOptions{MaxLength: 64, AlwaysAddHash: false})
	name = strings.Trim(name, "-_")
	if name == "" {
		return "kubernetes-ssh-key"
	}

	return name
}
