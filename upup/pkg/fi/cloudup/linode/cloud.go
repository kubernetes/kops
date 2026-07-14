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
	"os"
	"regexp"
	"strings"

	"github.com/linode/linodego/v2"
	v1 "k8s.io/api/core/v1"
	kopsv "k8s.io/kops"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

// LinodeClient is the subset of the Linode API client used by the Linode cloudup tasks.
type LinodeClient interface {
	ListVPCs(ctx context.Context, opts *linodego.ListOptions) ([]linodego.VPC, error)
	CreateVPC(ctx context.Context, opts linodego.VPCCreateOptions) (*linodego.VPC, error)
	UpdateVPC(ctx context.Context, vpcID int, opts linodego.VPCUpdateOptions) (*linodego.VPC, error)
	DeleteVPC(ctx context.Context, vpcID int) error
	ListSSHKeys(ctx context.Context, opts *linodego.ListOptions) ([]linodego.SSHKey, error)
	CreateSSHKey(ctx context.Context, opts linodego.SSHKeyCreateOptions) (*linodego.SSHKey, error)
	DeleteSSHKey(ctx context.Context, sshKeyID int) error
	ListVPCSubnets(ctx context.Context, vpcID int, opts *linodego.ListOptions) ([]linodego.VPCSubnet, error)
	CreateVPCSubnet(ctx context.Context, opts linodego.VPCSubnetCreateOptions, vpcID int) (*linodego.VPCSubnet, error)
	UpdateVPCSubnet(ctx context.Context, vpcID int, subnetID int, opts linodego.VPCSubnetUpdateOptions) (*linodego.VPCSubnet, error)
	DeleteVPCSubnet(ctx context.Context, vpcID int, subnetID int) error
}

// LinodeCloud exposes Linode (Akamai) cloud APIs used by kOps.
type LinodeCloud interface {
	fi.Cloud
	Client() LinodeClient
}

type Cloud struct {
	region string
	client LinodeClient
}

var _ LinodeCloud = &Cloud{}

var invalidLinodeLabelChars = regexp.MustCompile(`[^A-Za-z0-9_-]+`)

func NewCloud(region string) (LinodeCloud, error) {
	if region == "" {
		return nil, fmt.Errorf("region is required")
	}

	accessToken := os.Getenv("LINODE_TOKEN")
	if accessToken == "" {
		return nil, fmt.Errorf("%s is required", "LINODE_TOKEN")
	}

	client, err := linodego.NewClient(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Linode client: %w", err)
	}
	client.SetUserAgent("kops/" + kopsv.Version)
	client.SetToken(accessToken)

	return &Cloud{
		region: region,
		client: &client,
	}, nil
}

func (c *Cloud) Client() LinodeClient {
	return c.client
}

func (c *Cloud) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderLinode
}

func (c *Cloud) DNS() (dnsprovider.Interface, error) {
	return nil, fmt.Errorf("DNS is not yet implemented for Linode (Akamai)")
}

func (c *Cloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return nil, nil
}

func (c *Cloud) DeleteInstance(instance *cloudinstances.CloudInstance) error {
	return fmt.Errorf("instance deletion is not yet implemented for Linode (Akamai)")
}

func (c *Cloud) DeregisterInstance(instance *cloudinstances.CloudInstance) error {
	return nil
}

func (c *Cloud) DeleteGroup(group *cloudinstances.CloudInstanceGroup) error {
	return fmt.Errorf("instance group deletion is not yet implemented for Linode (Akamai)")
}

func (c *Cloud) DetachInstance(instance *cloudinstances.CloudInstance) error {
	return fmt.Errorf("instance detach is not yet implemented for Linode (Akamai)")
}

func (c *Cloud) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	return map[string]*cloudinstances.CloudInstanceGroup{}, nil
}

func (c *Cloud) Region() string {
	return c.region
}

func (c *Cloud) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	return &kops.ClusterStatus{}, nil
}

func (c *Cloud) GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	return nil, nil
}

// ListOptionsForLabel builds Linode (Akamai) list options that filter by an exact label match.
func ListOptionsForLabel(label string) (*linodego.ListOptions, error) {
	filter := &linodego.Filter{}
	filter.AddField(linodego.Eq, "label", label)
	filterJSON, err := filter.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error building Linode (Akamai) label filter: %w", err)
	}

	return &linodego.ListOptions{Filter: string(filterJSON)}, nil
}

// NormalizeLinodeLabel returns a normalized label for Linode (Akamai) resources
func NormalizeLinodeLabel(name string) string {
	name = invalidLinodeLabelChars.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-_")

	return name
}
