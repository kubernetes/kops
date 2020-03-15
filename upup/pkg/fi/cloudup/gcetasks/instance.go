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
	"strings"

	compute "google.golang.org/api/compute/v0.beta"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

var scopeAliases map[string]string

//go:generate fitask -type=Instance
type Instance struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	Network        *Network
	Tags           []string
	Preemptible    *bool
	Image          *string
	Disks          map[string]*Disk
	ServiceAccount *string

	CanIPForward *bool
	IPAddress    *Address
	Subnet       *Subnet

	Scopes []string

	Metadata    map[string]fi.Resource
	Zone        *string
	MachineType *string

	metadataFingerprint string
}

var _ fi.CompareWithID = &Instance{}

func (e *Instance) CompareWithID() *string {
	return e.Name
}

func (e *Instance) Find(c *fi.Context) (*Instance, error) {
	cloud := c.Cloud.(gce.GCECloud)

	r, err := cloud.Compute().Instances.Get(cloud.Project(), *e.Zone, *e.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing Instances: %v", err)
	}

	actual := &Instance{}
	actual.Name = &r.Name
	actual.Tags = append(actual.Tags, r.Tags.Items...)
	actual.Zone = fi.String(lastComponent(r.Zone))
	actual.MachineType = fi.String(lastComponent(r.MachineType))
	actual.CanIPForward = &r.CanIpForward

	if r.Scheduling != nil {
		actual.Preemptible = &r.Scheduling.Preemptible
	}
	if len(r.NetworkInterfaces) != 0 {
		ni := r.NetworkInterfaces[0]
		actual.Network = &Network{Name: fi.String(lastComponent(ni.Network))}
		if len(ni.AccessConfigs) != 0 {
			ac := ni.AccessConfigs[0]
			if ac.NatIP != "" {
				addr, err := cloud.Compute().Addresses.List(cloud.Project(), cloud.Region()).Filter("address eq " + ac.NatIP).Do()
				if err != nil {
					return nil, fmt.Errorf("error querying for address %q: %v", ac.NatIP, err)
				} else if len(addr.Items) != 0 {
					actual.IPAddress = &Address{Name: &addr.Items[0].Name}
				} else {
					return nil, fmt.Errorf("address not found %q: %v", ac.NatIP, err)
				}
			}
		}
	}

	for _, serviceAccount := range r.ServiceAccounts {
		for _, scope := range serviceAccount.Scopes {
			actual.Scopes = append(actual.Scopes, scopeToShortForm(scope))
		}
	}

	actual.Disks = make(map[string]*Disk)
	for i, disk := range r.Disks {
		if i == 0 {
			source := disk.Source

			// TODO: Parse source URL instead of assuming same project/zone?
			name := lastComponent(source)
			d, err := cloud.Compute().Disks.Get(cloud.Project(), *e.Zone, name).Do()
			if err != nil {
				if gce.IsNotFound(err) {
					return nil, fmt.Errorf("disk not found %q: %v", source, err)
				}
				return nil, fmt.Errorf("error querying for disk %q: %v", source, err)
			}

			image, err := ShortenImageURL(cloud.Project(), d.SourceImage)
			if err != nil {
				return nil, fmt.Errorf("error parsing source image URL: %v", err)
			}
			actual.Image = fi.String(image)
		} else {
			url, err := gce.ParseGoogleCloudURL(disk.Source)
			if err != nil {
				return nil, fmt.Errorf("unable to parse disk source URL: %q", disk.Source)
			}

			actual.Disks[disk.DeviceName] = &Disk{Name: &url.Name}
		}
	}

	if r.Metadata != nil {
		actual.Metadata = make(map[string]fi.Resource)
		for _, i := range r.Metadata.Items {
			actual.Metadata[i.Key] = fi.NewStringResource(fi.StringValue(i.Value))
		}
		actual.metadataFingerprint = r.Metadata.Fingerprint
	}

	return actual, nil
}

func (e *Instance) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *Instance) CheckChanges(a, e, changes *Instance) error {
	return nil
}

func init() {
	scopeAliases = map[string]string{
		"storage-ro":       "https://www.googleapis.com/auth/devstorage.read_only",
		"storage-rw":       "https://www.googleapis.com/auth/devstorage.read_write",
		"compute-ro":       "https://www.googleapis.com/auth/compute.read_only",
		"compute-rw":       "https://www.googleapis.com/auth/compute",
		"monitoring":       "https://www.googleapis.com/auth/monitoring",
		"monitoring-write": "https://www.googleapis.com/auth/monitoring.write",
		"logging-write":    "https://www.googleapis.com/auth/logging.write",
	}
}

func scopeToLongForm(s string) string {
	e, found := scopeAliases[s]
	if found {
		return e
	}
	return s
}

func scopeToShortForm(s string) string {
	for k, v := range scopeAliases {
		if v == s {
			return k
		}
	}
	return s
}

func (e *Instance) mapToGCE(project string, ipAddressResolver func(*Address) (*string, error)) (*compute.Instance, error) {
	zone := *e.Zone

	var scheduling *compute.Scheduling
	if fi.BoolValue(e.Preemptible) {
		scheduling = &compute.Scheduling{
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
		InitializeParams: &compute.AttachedDiskInitializeParams{
			SourceImage: BuildImageURL(project, *e.Image),
		},
		Boot:       true,
		DeviceName: "persistent-disks-0",
		Index:      0,
		AutoDelete: true,
		Mode:       "READ_WRITE",
		Type:       "PERSISTENT",
	})

	for name, disk := range e.Disks {
		disks = append(disks, &compute.AttachedDisk{
			Source:     disk.URL(project),
			AutoDelete: false,
			Mode:       "READ_WRITE",
			DeviceName: name,
		})
	}

	var tags *compute.Tags
	if e.Tags != nil {
		tags = &compute.Tags{
			Items: e.Tags,
		}
	}

	var networkInterfaces []*compute.NetworkInterface
	if e.IPAddress != nil {
		addr, err := ipAddressResolver(e.IPAddress)
		if err != nil {
			return nil, fmt.Errorf("unable to resolve IP for instance: %v", err)
		}
		if addr == nil {
			return nil, fmt.Errorf("instance IP address has not yet been created")
		}
		networkInterface := &compute.NetworkInterface{
			AccessConfigs: []*compute.AccessConfig{{
				NatIP: *addr,
				Type:  "ONE_TO_ONE_NAT",
			}},
			Network: e.Network.URL(project),
		}
		if e.Subnet != nil {
			networkInterface.Subnetwork = *e.Subnet.Name
		}
		networkInterfaces = append(networkInterfaces, networkInterface)
	}

	var serviceAccounts []*compute.ServiceAccount
	if e.ServiceAccount != nil {
		if e.Scopes != nil {
			var scopes []string
			for _, s := range e.Scopes {
				s = scopeToLongForm(s)
				scopes = append(scopes, s)
			}
			serviceAccounts = append(serviceAccounts, &compute.ServiceAccount{
				Email:  fi.StringValue(e.ServiceAccount),
				Scopes: scopes,
			})
		}
	}

	var metadataItems []*compute.MetadataItems
	for key, r := range e.Metadata {
		v, err := fi.ResourceAsString(r)
		if err != nil {
			return nil, fmt.Errorf("error rendering Instance metadata %q: %v", key, err)
		}
		metadataItems = append(metadataItems, &compute.MetadataItems{
			Key:   key,
			Value: fi.String(v),
		})
	}

	i := &compute.Instance{
		CanIpForward: *e.CanIPForward,

		Disks: disks,

		MachineType: BuildMachineTypeURL(project, zone, *e.MachineType),

		Metadata: &compute.Metadata{
			Items: metadataItems,
		},

		Name: *e.Name,

		NetworkInterfaces: networkInterfaces,

		Scheduling: scheduling,

		ServiceAccounts: serviceAccounts,

		Tags: tags,
	}

	return i, nil
}

func (i *Instance) isZero() bool {
	zero := &Instance{}
	return reflect.DeepEqual(zero, i)
}

func (_ *Instance) RenderGCE(t *gce.GCEAPITarget, a, e, changes *Instance) error {
	cloud := t.Cloud
	project := cloud.Project()
	zone := *e.Zone

	ipAddressResolver := func(ip *Address) (*string, error) {
		return ip.IPAddress, nil
	}

	i, err := e.mapToGCE(project, ipAddressResolver)
	if err != nil {
		return err
	}

	if a == nil {
		klog.V(2).Infof("Creating instance %q", i.Name)
		_, err := cloud.Compute().Instances.Insert(project, zone, i).Do()
		if err != nil {
			return fmt.Errorf("error creating Instance: %v", err)
		}
	} else {
		if changes.Metadata != nil {
			klog.V(2).Infof("Updating instance metadata on %q", i.Name)

			i.Metadata.Fingerprint = a.metadataFingerprint

			op, err := cloud.Compute().Instances.SetMetadata(project, zone, i.Name, i.Metadata).Do()
			if err != nil {
				return fmt.Errorf("error setting metadata on instance: %v", err)
			}

			if err := cloud.WaitForOp(op); err != nil {
				return fmt.Errorf("error setting metadata on instance: %v", err)
			}

			changes.Metadata = nil
		}

		if !changes.isZero() {
			klog.Errorf("Cannot apply changes to Instance: %v", changes)
			return fmt.Errorf("cannot apply changes to Instance: %v", changes)
		}
	}

	return nil
}

func BuildMachineTypeURL(project, zone, name string) string {
	return fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/machineTypes/%s", project, zone, name)
}

func BuildImageURL(defaultProject, nameSpec string) string {
	tokens := strings.Split(nameSpec, "/")
	var project, name string
	if len(tokens) == 2 {
		project = tokens[0]
		name = tokens[1]
	} else if len(tokens) == 1 {
		project = defaultProject
		name = tokens[0]
	} else {
		klog.Exitf("Cannot parse image spec: %q", nameSpec)
	}

	u := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/images/%s", project, name)
	klog.V(4).Infof("Mapped image %q to URL %q", nameSpec, u)
	return u
}

func ShortenImageURL(defaultProject string, imageURL string) (string, error) {
	u, err := gce.ParseGoogleCloudURL(imageURL)
	if err != nil {
		return "", err
	}
	if u.Project == defaultProject {
		klog.V(4).Infof("Resolved image %q -> %q", imageURL, u.Name)
		return u.Name, nil
	} else {
		klog.V(4).Infof("Resolved image %q -> %q", imageURL, u.Project+"/"+u.Name)
		return u.Project + "/" + u.Name, nil
	}
}

type terraformInstance struct {
	terraformInstanceCommon

	Name string `json:"name"`
}

func (_ *Instance) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Instance) error {
	project := t.Project

	// This is a "little" hacky...
	ipAddressResolver := func(ip *Address) (*string, error) {
		tf := "${google_compute_address." + *ip.Name + ".address}"
		return &tf, nil
	}

	i, err := e.mapToGCE(project, ipAddressResolver)
	if err != nil {
		return err
	}

	tf := &terraformInstance{
		Name: i.Name,
	}
	tf.CanIPForward = i.CanIpForward
	tf.MachineType = lastComponent(i.MachineType)
	tf.Zone = i.Zone
	tf.Tags = i.Tags.Items

	// TF requires zone
	if tf.Zone == "" && e.Zone != nil {
		tf.Zone = *e.Zone
	}

	tf.AddServiceAccounts(i.ServiceAccounts)

	for _, d := range i.Disks {
		tfd := &terraformAttachedDisk{
			AutoDelete: d.AutoDelete,
			Scratch:    d.Type == "SCRATCH",
			DeviceName: d.DeviceName,

			// TODO: Does this need to be a TF link?
			Disk: lastComponent(d.Source),
		}
		if d.InitializeParams != nil {
			tfd.Disk = d.InitializeParams.DiskName
			tfd.Image = d.InitializeParams.SourceImage
			tfd.Type = d.InitializeParams.DiskType
			tfd.Size = d.InitializeParams.DiskSizeGb
		}
		tf.Disks = append(tf.Disks, tfd)
	}

	tf.AddNetworks(e.Network, e.Subnet, i.NetworkInterfaces)

	tf.AddMetadata(t, i.Name, i.Metadata)

	// Using metadata_startup_script is now mandatory (?)
	{
		startupScript, found := tf.Metadata["startup-script"]
		if found {
			delete(tf.Metadata, "startup-script")
		}
		tf.MetadataStartupScript = startupScript
	}

	if i.Scheduling != nil {
		tf.Scheduling = &terraformScheduling{
			AutomaticRestart:  fi.BoolValue(i.Scheduling.AutomaticRestart),
			OnHostMaintenance: i.Scheduling.OnHostMaintenance,
			Preemptible:       i.Scheduling.Preemptible,
		}
	}

	return t.RenderResource("google_compute_instance", i.Name, tf)
}
