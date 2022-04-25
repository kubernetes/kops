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

package gcetasks

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

const (
	// terraform 0.12 with google cloud provider 3.2 will complain if the length of the name_prefix is more than 32
	InstanceTemplateNamePrefixMaxLength = 32

	accessConfigOneToOneNAT = "ONE_TO_ONE_NAT"
)

// InstanceTemplate represents a GCE InstanceTemplate
// +kops:fitask
type InstanceTemplate struct {
	Name *string

	// NamePrefix is used as the prefix for the names; we add a timestamp.  Max = InstanceTemplateNamePrefixMaxLength
	NamePrefix *string

	Lifecycle fi.Lifecycle

	Network     *Network
	Tags        []string
	Labels      map[string]string
	Preemptible *bool

	BootDiskImage  *string
	BootDiskSizeGB *int64
	BootDiskType   *string

	CanIPForward  *bool
	Subnet        *Subnet
	AliasIPRanges map[string]string

	Scopes          []string
	ServiceAccounts []*ServiceAccount

	Metadata    map[string]fi.Resource
	MachineType *string

	// HasExternalIP is set to true when an external IP is allocated to an instance.
	HasExternalIP *bool

	// ID is the actual name
	ID *string
}

var (
	_ fi.Task          = &InstanceTemplate{}
	_ fi.CompareWithID = &InstanceTemplate{}
)

func (e *InstanceTemplate) CompareWithID() *string {
	return e.ID
}

func (e *InstanceTemplate) Find(c *fi.Context) (*InstanceTemplate, error) {
	cloud := c.Cloud.(gce.GCECloud)

	templates, err := cloud.Compute().InstanceTemplates().List(context.Background(), cloud.Project())
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing InstanceTemplates: %v", err)
	}

	expected, err := e.mapToGCE(cloud.Project(), cloud.Region())
	if err != nil {
		return nil, err
	}

	for _, r := range templates {
		if !strings.HasPrefix(r.Name, fi.StringValue(e.NamePrefix)+"-") {
			continue
		}

		if !matches(expected, r) {
			continue
		}

		actual := &InstanceTemplate{}

		p := r.Properties

		actual.Tags = append(actual.Tags, p.Tags.Items...)
		actual.Labels = p.Labels
		actual.MachineType = fi.String(lastComponent(p.MachineType))
		actual.CanIPForward = &p.CanIpForward

		bootDiskImage, err := ShortenImageURL(cloud.Project(), p.Disks[0].InitializeParams.SourceImage)
		if err != nil {
			return nil, fmt.Errorf("error parsing source image URL: %v", err)
		}
		actual.BootDiskImage = fi.String(bootDiskImage)
		actual.BootDiskType = &p.Disks[0].InitializeParams.DiskType
		actual.BootDiskSizeGB = &p.Disks[0].InitializeParams.DiskSizeGb

		if p.Scheduling != nil {
			actual.Preemptible = &p.Scheduling.Preemptible
		}
		if len(p.NetworkInterfaces) != 0 {
			ni := p.NetworkInterfaces[0]
			actual.Network = &Network{Name: fi.String(lastComponent(ni.Network))}

			if len(ni.AliasIpRanges) != 0 {
				actual.AliasIPRanges = make(map[string]string)
				for _, aliasIPRange := range ni.AliasIpRanges {
					actual.AliasIPRanges[aliasIPRange.SubnetworkRangeName] = aliasIPRange.IpCidrRange
				}
			}

			if ni.Subnetwork != "" {
				actual.Subnet = &Subnet{Name: fi.String(lastComponent(ni.Subnetwork))}
			}

			acs := ni.AccessConfigs
			if len(acs) > 0 {
				if len(acs) != 1 {
					return nil, fmt.Errorf("unexpected number of access configs in template %q: %d", *actual.Name, len(acs))
				}
				if acs[0].Type != accessConfigOneToOneNAT {
					return nil, fmt.Errorf("unexpected access type in template %q: %s", *actual.Name, acs[0].Type)
				}
				actual.HasExternalIP = fi.Bool(true)
			} else {
				actual.HasExternalIP = fi.Bool(false)
			}
		}

		for _, serviceAccount := range p.ServiceAccounts {
			for _, scope := range serviceAccount.Scopes {
				actual.Scopes = append(actual.Scopes, scopeToShortForm(scope))
			}
			actual.ServiceAccounts = append(actual.ServiceAccounts, &ServiceAccount{
				Email: &serviceAccount.Email,
			})
		}

		// When we deal with additional disks (local disks), we'll need to map them like this...
		//for i, disk := range p.Disks {
		//	if i == 0 {
		//		source := disk.Source
		//
		//		// TODO: Parse source URL instead of assuming same project/zone?
		//		name := lastComponent(source)
		//		d, err := cloud.Compute.Disks.Get(cloud.Project, *e.Zone, name).Do()
		//		if err != nil {
		//			if gce.IsNotFound(err) {
		//				return nil, fmt.Errorf("disk not found %q: %v", source, err)
		//			}
		//			return nil, fmt.Errorf("error querying for disk %q: %v", source, err)
		//		} else {
		//			imageURL, err := gce.ParseGoogleCloudURL(d.SourceImage)
		//			if err != nil {
		//				return nil, fmt.Errorf("unable to parse image URL: %q", d.SourceImage)
		//			}
		//			actual.Image = fi.String(imageURL.Project + "/" + imageURL.Name)
		//		}
		//	}
		//}

		if p.Metadata != nil {
			actual.Metadata = make(map[string]fi.Resource)
			for _, meta := range p.Metadata.Items {
				actual.Metadata[meta.Key] = fi.NewStringResource(fi.StringValue(meta.Value))
			}
		}

		// Prevent spurious changes
		actual.Name = e.Name
		actual.NamePrefix = e.NamePrefix

		actual.ID = &r.Name
		if e.ID == nil {
			e.ID = actual.ID
		}

		// System fields
		actual.Lifecycle = e.Lifecycle

		return actual, nil
	}

	return nil, nil
}

func (e *InstanceTemplate) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *InstanceTemplate) CheckChanges(a, e, changes *InstanceTemplate) error {
	if fi.StringValue(e.BootDiskImage) == "" {
		return fi.RequiredField("BootDiskImage")
	}
	if fi.StringValue(e.MachineType) == "" {
		return fi.RequiredField("MachineType")
	}
	return nil
}

func (e *InstanceTemplate) mapToGCE(project string, region string) (*compute.InstanceTemplate, error) {
	// TODO: This is similar to Instance...
	var scheduling *compute.Scheduling

	if fi.BoolValue(e.Preemptible) {
		scheduling = &compute.Scheduling{
			AutomaticRestart:  fi.Bool(false),
			OnHostMaintenance: "TERMINATE",
			Preemptible:       true,
		}
	} else {
		scheduling = &compute.Scheduling{
			AutomaticRestart: fi.Bool(true),
			// TODO: Migrate or terminate?
			OnHostMaintenance: "MIGRATE",
			Preemptible:       false,
		}
	}

	var disks []*compute.AttachedDisk
	disks = append(disks, &compute.AttachedDisk{
		Kind: "compute#attachedDisk",
		InitializeParams: &compute.AttachedDiskInitializeParams{
			SourceImage: BuildImageURL(project, *e.BootDiskImage),
			DiskSizeGb:  *e.BootDiskSizeGB,
			DiskType:    *e.BootDiskType,
		},
		Boot:       true,
		DeviceName: "persistent-disks-0",
		Index:      0,
		AutoDelete: true,
		Mode:       "READ_WRITE",
		Type:       "PERSISTENT",
	})

	var tags *compute.Tags
	if e.Tags != nil {
		tags = &compute.Tags{
			Items: e.Tags,
		}
	}

	var networkInterfaces []*compute.NetworkInterface

	networkProject := project
	if e.Network.Project != nil {
		networkProject = *e.Network.Project
	}

	ni := &compute.NetworkInterface{
		Kind:    "compute#networkInterface",
		Network: e.Network.URL(networkProject),
	}
	if fi.BoolValue(e.HasExternalIP) {
		ni.AccessConfigs = []*compute.AccessConfig{
			{
				Kind:        "compute#accessConfig",
				Type:        accessConfigOneToOneNAT,
				NetworkTier: "PREMIUM",
			},
		}
	}

	if e.Subnet != nil {
		ni.Subnetwork = e.Subnet.URL(networkProject, region)
	}
	if e.AliasIPRanges != nil {
		for k, v := range e.AliasIPRanges {
			ni.AliasIpRanges = append(ni.AliasIpRanges, &compute.AliasIpRange{
				SubnetworkRangeName: k,
				IpCidrRange:         v,
			})
		}
	}
	networkInterfaces = append(networkInterfaces, ni)

	scopes := make([]string, 0)
	if e.Scopes != nil {
		for _, s := range e.Scopes {
			s = scopeToLongForm(s)
			scopes = append(scopes, s)
		}
	}

	var serviceAccounts []*compute.ServiceAccount
	for _, sa := range e.ServiceAccounts {
		serviceAccounts = append(serviceAccounts, &compute.ServiceAccount{
			Email:  fi.StringValue(sa.Email),
			Scopes: scopes,
		})
	}

	var metadataItems []*compute.MetadataItems
	for key, r := range e.Metadata {
		v, err := fi.ResourceAsString(r)
		if err != nil {
			return nil, fmt.Errorf("error rendering InstanceTemplate metadata %q: %v", key, err)
		}
		metadataItems = append(metadataItems, &compute.MetadataItems{
			Key:   key,
			Value: fi.String(v),
		})
	}

	i := &compute.InstanceTemplate{
		Kind: "compute#instanceTemplate",
		Properties: &compute.InstanceProperties{
			CanIpForward: *e.CanIPForward,

			Disks: disks,

			MachineType: *e.MachineType,

			Metadata: &compute.Metadata{
				Kind:  "compute#metadata",
				Items: metadataItems,
			},

			NetworkInterfaces: networkInterfaces,

			Scheduling: scheduling,

			ServiceAccounts: serviceAccounts,

			Labels: e.Labels,
			Tags:   tags,
		},
	}

	return i, nil
}

type ByKey []*compute.MetadataItems

func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

func matches(l, r *compute.InstanceTemplate) bool {
	normalizeInstanceProperties := func(v *compute.InstanceProperties) *compute.InstanceProperties {
		c := *v
		if c.Metadata != nil {
			cm := *c.Metadata
			c.Metadata = &cm
			c.Metadata.Fingerprint = ""
			sort.Sort(ByKey(c.Metadata.Items))
		}
		// Ignore output fields
		for _, ni := range c.NetworkInterfaces {
			ni.Name = ""
		}
		return &c
	}
	normalize := func(v *compute.InstanceTemplate) *compute.InstanceTemplate {
		c := *v
		c.SelfLink = ""
		c.CreationTimestamp = ""
		c.Id = 0
		c.Name = ""
		c.Properties = normalizeInstanceProperties(c.Properties)
		return &c
	}
	normalizedL := normalize(l)
	normalizedR := normalize(r)

	if !reflect.DeepEqual(normalizedL, normalizedR) {
		if klog.V(10).Enabled() {
			ls := fi.DebugAsJsonStringIndent(normalizedL)
			rs := fi.DebugAsJsonStringIndent(normalizedR)
			klog.V(10).Infof("Not equal")
			klog.V(10).Infof(diff.FormatDiff(ls, rs))
		}
		return false
	}

	return true
}

func (e *InstanceTemplate) URL(project string) (string, error) {
	if e.ID == nil {
		return "", fmt.Errorf("InstanceTemplate not yet built; ID is not yet known")
	}
	return fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/instanceTemplates/%s", project, *e.ID), nil
}

func (_ *InstanceTemplate) RenderGCE(t *gce.GCEAPITarget, a, e, changes *InstanceTemplate) error {
	project := t.Cloud.Project()
	region := t.Cloud.Region()

	i, err := e.mapToGCE(project, region)
	if err != nil {
		return err
	}

	if a == nil {
		klog.V(4).Infof("Creating InstanceTemplate %v", i)

		name := fi.StringValue(e.NamePrefix) + "-" + strconv.FormatInt(time.Now().Unix(), 10)
		e.ID = &name
		i.Name = name

		op, err := t.Cloud.Compute().InstanceTemplates().Insert(t.Cloud.Project(), i)
		if err != nil {
			return fmt.Errorf("error creating InstanceTemplate: %v", err)
		}

		if err := t.Cloud.WaitForOp(op); err != nil {
			return fmt.Errorf("error creating InstanceTemplate: %v", err)
		}
	} else {
		return fmt.Errorf("Cannot apply changes to InstanceTemplate: %v", changes)
	}

	return nil
}

type terraformInstanceTemplate struct {
	NamePrefix            string                                   `cty:"name_prefix"`
	CanIPForward          bool                                     `cty:"can_ip_forward"`
	MachineType           string                                   `cty:"machine_type"`
	ServiceAccounts       []*terraformTemplateServiceAccount       `cty:"service_account"`
	Scheduling            *terraformScheduling                     `cty:"scheduling"`
	Disks                 []*terraformInstanceTemplateAttachedDisk `cty:"disk"`
	Labels                map[string]string                        `cty:"labels"`
	NetworkInterfaces     []*terraformNetworkInterface             `cty:"network_interface"`
	Metadata              map[string]*terraformWriter.Literal      `cty:"metadata"`
	MetadataStartupScript *terraformWriter.Literal                 `cty:"metadata_startup_script"`
	Tags                  []string                                 `cty:"tags"`
}

type terraformTemplateServiceAccount struct {
	Email  *terraformWriter.Literal `cty:"email"`
	Scopes []string                 `cty:"scopes"`
}

type terraformScheduling struct {
	AutomaticRestart  bool   `cty:"automatic_restart"`
	OnHostMaintenance string `cty:"on_host_maintenance"`
	Preemptible       bool   `cty:"preemptible"`
}

type terraformInstanceTemplateAttachedDisk struct {
	AutoDelete bool   `cty:"auto_delete"`
	DeviceName string `cty:"device_name"`

	// scratch vs persistent
	Type        string `cty:"type"`
	Boot        bool   `cty:"boot"`
	DiskName    string `cty:"disk_name"`
	SourceImage string `cty:"source_image"`
	Source      string `cty:"source"`
	Interface   string `cty:"interface"`
	Mode        string `cty:"mode"`
	DiskType    string `cty:"disk_type"`
	DiskSizeGB  int64  `cty:"disk_size_gb"`
}

type terraformNetworkInterface struct {
	Network      *terraformWriter.Literal `cty:"network"`
	Subnetwork   *terraformWriter.Literal `cty:"subnetwork"`
	AccessConfig []*terraformAccessConfig `cty:"access_config"`
}

type terraformAccessConfig struct {
	NatIP *terraformWriter.Literal `cty:"nat_ip"`
}

func addNetworks(network *Network, subnet *Subnet, networkInterfaces []*compute.NetworkInterface) []*terraformNetworkInterface {
	ni := make([]*terraformNetworkInterface, 0)
	for _, g := range networkInterfaces {
		tf := &terraformNetworkInterface{}
		if network != nil {
			tf.Network = network.TerraformLink()
		}
		if subnet != nil {
			tf.Subnetwork = subnet.TerraformLink()
		}
		for _, gac := range g.AccessConfigs {
			tac := &terraformAccessConfig{}
			natIP := gac.NatIP
			if natIP != "" {
				tac.NatIP = terraformWriter.LiteralFromStringValue(natIP)
			}

			tf.AccessConfig = append(tf.AccessConfig, tac)
		}

		ni = append(ni, tf)
	}
	return ni
}

func addMetadata(target *terraform.TerraformTarget, name string, metadata *compute.Metadata) (map[string]*terraformWriter.Literal, error) {
	if metadata == nil {
		return nil, nil
	}
	m := make(map[string]*terraformWriter.Literal)
	for _, g := range metadata.Items {
		val := fi.StringValue(g.Value)
		if strings.Contains(val, "\n") {
			tfResource, err := target.AddFileBytes("google_compute_instance_template", name, "metadata_"+g.Key, []byte(val), false)
			if err != nil {
				return nil, err
			}
			m[g.Key] = tfResource
		} else {
			m[g.Key] = terraformWriter.LiteralFromStringValue(val)
		}
	}
	return m, nil
}

func mapServiceAccountsToTerraform(serviceAccounts []*ServiceAccount, saScopes []string) []*terraformTemplateServiceAccount {
	var scopes []string
	for _, s := range saScopes {
		s = scopeToLongForm(s)
		scopes = append(scopes, s)
	}
	// Note that GCE currently only allows one service account per VM,
	// but the model in both the API and terraform allows more.
	var out []*terraformTemplateServiceAccount
	for _, serviceAccount := range serviceAccounts {
		tsa := &terraformTemplateServiceAccount{
			Email:  serviceAccount.TerraformLink(),
			Scopes: scopes,
		}
		out = append(out, tsa)
	}
	return out
}

func (_ *InstanceTemplate) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *InstanceTemplate) error {
	project := t.Project

	i, err := e.mapToGCE(project, t.Cloud.Region())
	if err != nil {
		return err
	}

	name := fi.StringValue(e.Name)

	tf := &terraformInstanceTemplate{
		NamePrefix: fi.StringValue(e.NamePrefix) + "-",
	}

	tf.CanIPForward = i.Properties.CanIpForward
	tf.MachineType = lastComponent(i.Properties.MachineType)
	tf.Labels = i.Properties.Labels
	tf.Tags = i.Properties.Tags.Items

	tf.ServiceAccounts = mapServiceAccountsToTerraform(e.ServiceAccounts, e.Scopes)

	for _, d := range i.Properties.Disks {
		tfd := &terraformInstanceTemplateAttachedDisk{
			AutoDelete:  d.AutoDelete,
			Boot:        d.Boot,
			DeviceName:  d.DeviceName,
			DiskName:    d.InitializeParams.DiskName,
			SourceImage: d.InitializeParams.SourceImage,
			Source:      d.Source,
			Interface:   d.Interface,
			Mode:        d.Mode,
			DiskType:    d.InitializeParams.DiskType,
			DiskSizeGB:  d.InitializeParams.DiskSizeGb,
			Type:        d.Type,
		}
		tf.Disks = append(tf.Disks, tfd)
	}

	tf.NetworkInterfaces = addNetworks(e.Network, e.Subnet, i.Properties.NetworkInterfaces)

	metadata, err := addMetadata(t, name, i.Properties.Metadata)
	if err != nil {
		return err
	}
	tf.Metadata = metadata

	if i.Properties.Scheduling != nil {
		tf.Scheduling = &terraformScheduling{
			AutomaticRestart:  fi.BoolValue(i.Properties.Scheduling.AutomaticRestart),
			OnHostMaintenance: i.Properties.Scheduling.OnHostMaintenance,
			Preemptible:       i.Properties.Scheduling.Preemptible,
		}
	}

	return t.RenderResource("google_compute_instance_template", name, tf)
}

func (i *InstanceTemplate) TerraformLink() *terraformWriter.Literal {
	return terraformWriter.LiteralSelfLink("google_compute_instance_template", *i.Name)
}
