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
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	compute "google.golang.org/api/compute/v0.beta"
	"k8s.io/klog"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

// terraform 0.12 with google cloud provider 3.2 will complain if the length of the name_prefix is more than 32
const InstanceTemplateNamePrefixMaxLength = 32

// InstanceTemplate represents a GCE InstanceTemplate
//go:generate fitask -type=InstanceTemplate
type InstanceTemplate struct {
	Name *string

	// NamePrefix is used as the prefix for the names; we add a timestamp.  Max = InstanceTemplateNamePrefixMaxLength
	NamePrefix *string

	Lifecycle *fi.Lifecycle

	Network *Network
	Tags    []string
	//Labels      map[string]string
	Preemptible *bool

	BootDiskImage  *string
	BootDiskSizeGB *int64
	BootDiskType   *string

	CanIPForward  *bool
	Subnet        *Subnet
	AliasIPRanges map[string]string

	Scopes          []string
	ServiceAccounts []string

	Metadata    map[string]*fi.ResourceHolder
	MachineType *string

	// ID is the actual name
	ID *string
}

var _ fi.CompareWithID = &InstanceTemplate{}

func (e *InstanceTemplate) CompareWithID() *string {
	return e.ID
}

func (e *InstanceTemplate) Find(c *fi.Context) (*InstanceTemplate, error) {
	cloud := c.Cloud.(gce.GCECloud)

	response, err := cloud.Compute().InstanceTemplates.List(cloud.Project()).Do()
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

	for _, r := range response.Items {
		if !strings.HasPrefix(r.Name, fi.StringValue(e.NamePrefix)+"-") {
			continue
		}

		if !matches(expected, r) {
			continue
		}

		actual := &InstanceTemplate{}

		p := r.Properties

		actual.Tags = append(actual.Tags, p.Tags.Items...)
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
		}

		for _, serviceAccount := range p.ServiceAccounts {
			for _, scope := range serviceAccount.Scopes {
				actual.Scopes = append(actual.Scopes, scopeToShortForm(scope))
			}
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
			actual.Metadata = make(map[string]*fi.ResourceHolder)
			for _, meta := range p.Metadata.Items {
				actual.Metadata[meta.Key] = fi.WrapResource(fi.NewStringResource(fi.StringValue(meta.Value)))
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

	klog.Infof("We should be using NVME for GCE")

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
	ni := &compute.NetworkInterface{
		Kind: "compute#networkInterface",
		AccessConfigs: []*compute.AccessConfig{{
			Kind: "compute#accessConfig",
			//NatIP: *e.IPAddress.Address,
			Type:        "ONE_TO_ONE_NAT",
			NetworkTier: "PREMIUM",
		}},
		Network: e.Network.URL(project),
	}
	if e.Subnet != nil {
		ni.Subnetwork = e.Subnet.URL(project, region)
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
	serviceAccounts := []*compute.ServiceAccount{
		{
			Email:  e.ServiceAccounts[0],
			Scopes: scopes,
		},
	}
	// if e.ServiceAccounts != nil {
	// 	for _, s := range e.ServiceAccounts {
	// 		serviceAccounts = append(serviceAccounts, &compute.ServiceAccount{
	// 			Email:  s,
	// 			Scopes: scopes,
	// 		})
	// 	}
	// } else {
	// 	serviceAccounts = append(serviceAccounts, &compute.ServiceAccount{
	// 		Email:  "default",
	// 		Scopes: scopes,
	// 	})
	// }

	var metadataItems []*compute.MetadataItems
	for key, r := range e.Metadata {
		v, err := r.AsString()
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

			Tags: tags,
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
		if klog.V(10) {
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

		op, err := t.Cloud.Compute().InstanceTemplates.Insert(t.Cloud.Project(), i).Do()
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
	terraformInstanceCommon
	NamePrefix string `json:"name_prefix"`
}

type terraformInstanceCommon struct {
	CanIPForward          bool                          `json:"can_ip_forward"`
	MachineType           string                        `json:"machine_type,omitempty"`
	ServiceAccount        *terraformServiceAccount      `json:"service_account,omitempty"`
	Scheduling            *terraformScheduling          `json:"scheduling,omitempty"`
	Disks                 []*terraformAttachedDisk      `json:"disk,omitempty"`
	NetworkInterfaces     []*terraformNetworkInterface  `json:"network_interface,omitempty"`
	Metadata              map[string]*terraform.Literal `json:"metadata,omitempty"`
	MetadataStartupScript *terraform.Literal            `json:"metadata_startup_script,omitempty"`
	Tags                  []string                      `json:"tags,omitempty"`

	// Only for instances:
	Zone string `json:"zone,omitempty"`
}

type terraformServiceAccount struct {
	Email  string   `json:"email"`
	Scopes []string `json:"scopes"`
}

type terraformScheduling struct {
	AutomaticRestart  bool   `json:"automatic_restart"`
	OnHostMaintenance string `json:"on_host_maintenance,omitempty"`
	Preemptible       bool   `json:"preemptible"`
}

type terraformAttachedDisk struct {
	// These values are common
	AutoDelete bool   `json:"auto_delete,omitempty"`
	DeviceName string `json:"device_name,omitempty"`

	// DANGER - common but different meaning:
	//   for an instance template this is scratch vs persistent
	//   for an instance this is 'pd-standard', 'pd-ssd', 'local-ssd' etc
	Type string `json:"type,omitempty"`

	// These values are only for instance templates:
	Boot        bool   `json:"boot,omitempty"`
	DiskName    string `json:"disk_name,omitempty"`
	SourceImage string `json:"source_image,omitempty"`
	Source      string `json:"source,omitempty"`
	Interface   string `json:"interface,omitempty"`
	Mode        string `json:"mode,omitempty"`
	DiskType    string `json:"disk_type,omitempty"`
	DiskSizeGB  int64  `json:"disk_size_gb,omitempty"`

	// These values are only for instances:
	Disk    string `json:"disk,omitempty"`
	Image   string `json:"image,omitempty"`
	Scratch bool   `json:"scratch,omitempty"`
	Size    int64  `json:"size,omitempty"`
}

type terraformNetworkInterface struct {
	Network      *terraform.Literal       `json:"network,omitempty"`
	Subnetwork   *terraform.Literal       `json:"subnetwork,omitempty"`
	AccessConfig []*terraformAccessConfig `json:"access_config"`
}

type terraformAccessConfig struct {
	NatIP *terraform.Literal `json:"nat_ip,omitempty"`
}

func (t *terraformInstanceCommon) AddNetworks(network *Network, subnet *Subnet, networkInterfacs []*compute.NetworkInterface) {
	for _, g := range networkInterfacs {
		tf := &terraformNetworkInterface{}
		if network != nil {
			tf.Network = network.TerraformName()
		}
		if subnet != nil {
			tf.Subnetwork = subnet.TerraformName()
		}
		for _, gac := range g.AccessConfigs {
			tac := &terraformAccessConfig{}
			natIP := gac.NatIP
			if strings.HasPrefix(natIP, "${") {
				tac.NatIP = terraform.LiteralExpression(natIP)
			} else if natIP != "" {
				tac.NatIP = terraform.LiteralFromStringValue(natIP)
			}

			tf.AccessConfig = append(tf.AccessConfig, tac)
		}

		t.NetworkInterfaces = append(t.NetworkInterfaces, tf)
	}
}

func (t *terraformInstanceCommon) AddMetadata(target *terraform.TerraformTarget, name string, metadata *compute.Metadata) error {
	if metadata != nil {
		if t.Metadata == nil {
			t.Metadata = make(map[string]*terraform.Literal)
		}
		for _, g := range metadata.Items {
			v := fi.NewStringResource(fi.StringValue(g.Value))
			tfResource, err := target.AddFile("google_compute_instance_template", name, "metadata_"+g.Key, v)
			if err != nil {
				return err
			}

			t.Metadata[g.Key] = tfResource
		}
	}

	return nil
}

func (t *terraformInstanceCommon) AddServiceAccounts(serviceAccounts []*compute.ServiceAccount) {
	// there's an inconsistency here- GCP only lets you have one service account per VM
	// terraform gets it right, but the golang api doesn't. womp womp :(
	if len(serviceAccounts) != 1 {
		klog.Fatal("Instances can only have 1 service account assigned.")
	} else {
		klog.Infof("adding csa: %v", serviceAccounts[0].Email)
		csa := serviceAccounts[0]
		tsa := &terraformServiceAccount{
			Email:  csa.Email,
			Scopes: csa.Scopes,
		}
		// for _, scope := range csa.Scopes {
		// 	tsa.Scopes = append(tsa.Scopes, scope)
		// }
		t.ServiceAccount = tsa
	}
}
func (_ *InstanceTemplate) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *InstanceTemplate) error {
	project := t.Project

	i, err := e.mapToGCE(project, t.Region)
	if err != nil {
		return err
	}

	name := fi.StringValue(e.Name)

	tf := &terraformInstanceTemplate{
		NamePrefix: fi.StringValue(e.NamePrefix) + "-",
	}

	tf.CanIPForward = i.Properties.CanIpForward
	tf.MachineType = lastComponent(i.Properties.MachineType)
	//tf.Zone = i.Properties.Zone
	tf.Tags = i.Properties.Tags.Items

	tf.AddServiceAccounts(i.Properties.ServiceAccounts)

	for _, d := range i.Properties.Disks {
		tfd := &terraformAttachedDisk{
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

	tf.AddNetworks(e.Network, e.Subnet, i.Properties.NetworkInterfaces)

	tf.AddMetadata(t, name, i.Properties.Metadata)

	if i.Properties.Scheduling != nil {
		tf.Scheduling = &terraformScheduling{
			AutomaticRestart:  fi.BoolValue(i.Properties.Scheduling.AutomaticRestart),
			OnHostMaintenance: i.Properties.Scheduling.OnHostMaintenance,
			Preemptible:       i.Properties.Scheduling.Preemptible,
		}
	}

	return t.RenderResource("google_compute_instance_template", name, tf)
}

func (i *InstanceTemplate) TerraformLink() *terraform.Literal {
	return terraform.LiteralSelfLink("google_compute_instance_template", *i.Name)
}
