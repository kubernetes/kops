package gcetasks

import (
	"fmt"

	"github.com/golang/glog"
	"google.golang.org/api/compute/v1"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/gce"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kube-deploy/upup/pkg/fi/utils"
	"strings"
)

var scopeAliases map[string]string

type Instance struct {
	Name        *string
	Network     *Network
	Tags        []string
	Preemptible *bool
	Image       *string
	Disks       map[string]*PersistentDisk

	CanIPForward *bool
	IPAddress    *IPAddress
	Subnet       *Subnet

	Scopes []string

	Metadata    map[string]fi.Resource
	Zone        *string
	MachineType *string
}

func (d *Instance) String() string {
	return utils.JsonString(d)
}

func (d *Instance) CompareWithID() *string {
	return d.Name
}

func (e *Instance) Find(c *fi.Context) (*Instance, error) {
	cloud := c.Cloud.(*gce.GCECloud)

	r, err := cloud.Compute.Instances.Get(cloud.Project, *e.Zone, *e.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing Instances: %v", err)
	}

	actual := &Instance{}
	actual.Name = &r.Name
	for _, tag := range r.Tags.Items {
		actual.Tags = append(actual.Tags, tag)
	}
	actual.Zone = fi.String(lastComponent(r.Zone))
	actual.MachineType = fi.String(lastComponent(r.MachineType))
	actual.CanIPForward = &r.CanIpForward
	actual.Image = &r.Disks[0].Source

	if r.Scheduling != nil {
		actual.Preemptible = &r.Scheduling.Preemptible
	}
	if len(r.NetworkInterfaces) != 0 {
		ni := r.NetworkInterfaces[0]
		actual.Network = &Network{Name: fi.String(lastComponent(ni.Network))}
		if len(ni.AccessConfigs) != 0 {
			ac := ni.AccessConfigs[0]
			if ac.NatIP != "" {
				addr, err := cloud.Compute.Addresses.List(cloud.Project, cloud.Region).Filter("address eq " + ac.NatIP).Do()
				if err != nil {
					return nil, fmt.Errorf("error querying for address %q: %v", ac.NatIP, err)
				} else if len(addr.Items) != 0 {
					actual.IPAddress = &IPAddress{Name: &addr.Items[0].Name}
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

	actual.Disks = make(map[string]*PersistentDisk)
	for i, disk := range r.Disks {
		if i == 0 {
			source := disk.Source

			// TODO: Parse source URL instead of assuming same project/zone?
			name := lastComponent(source)
			d, err := cloud.Compute.Disks.Get(cloud.Project, *e.Zone, name).Do()
			if err != nil {
				if gce.IsNotFound(err) {
					return nil, fmt.Errorf("disk not found %q: %v", source, err)
				}
				return nil, fmt.Errorf("error querying for disk %q: %v", source, err)
			} else {
				imageURL, err := gce.ParseGoogleCloudURL(d.SourceImage)
				if err != nil {
					return nil, fmt.Errorf("unable to parse image URL: %q", d.SourceImage)
				}
				actual.Image = fi.String(imageURL.Project + "/" + imageURL.Name)
			}
		} else {
			url, err := gce.ParseGoogleCloudURL(disk.Source)
			if err != nil {
				return nil, fmt.Errorf("unable to parse disk source URL: %q", disk.Source)
			}

			actual.Disks[disk.DeviceName] = &PersistentDisk{Name: &url.Name}
		}
	}

	return actual, nil
}

func (e *Instance) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *Instance) CheckChanges(a, e, changes *Instance) error {
	return nil
}

func expandScopeAlias(s string) string {
	switch s {
	case "storage-ro":
		s = "https://www.googleapis.com/auth/devstorage.read_only"
	case "storage-rw":
		s = "https://www.googleapis.com/auth/devstorage.read_write"
	case "compute-ro":
		s = "https://www.googleapis.com/auth/compute.read_only"
	case "compute-rw":
		s = "https://www.googleapis.com/auth/compute"
	case "monitoring":
		s = "https://www.googleapis.com/auth/monitoring"
	case "monitoring-write":
		s = "https://www.googleapis.com/auth/monitoring.write"
	case "logging-write":
		s = "https://www.googleapis.com/auth/logging.write"
	}
	return s
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

func (e *Instance) mapToGCE(project string, ipAddressResolver func(*IPAddress) (*string, error)) (*compute.Instance, error) {
	zone := *e.Zone

	var scheduling *compute.Scheduling
	if fi.BoolValue(e.Preemptible) {
		scheduling = &compute.Scheduling{
			OnHostMaintenance: "TERMINATE",
			Preemptible:       true,
		}
	} else {
		scheduling = &compute.Scheduling{
			AutomaticRestart: true,
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
	if e.Scopes != nil {
		var scopes []string
		for _, s := range e.Scopes {
			s = expandScopeAlias(s)

			scopes = append(scopes, s)
		}
		serviceAccounts = append(serviceAccounts, &compute.ServiceAccount{
			Email:  "default",
			Scopes: scopes,
		})
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

func (_ *Instance) RenderGCE(t *gce.GCEAPITarget, a, e, changes *Instance) error {
	project := t.Cloud.Project

	ipAddressResolver := func(ip *IPAddress) (*string, error) {
		return ip.FindAddress(t.Cloud)
	}

	i, err := e.mapToGCE(project, ipAddressResolver)
	if err != nil {
		return err
	}

	if a == nil {
		_, err := t.Cloud.Compute.Instances.Insert(t.Cloud.Project, *e.Zone, i).Do()
		if err != nil {
			return fmt.Errorf("error creating Instance: %v", err)
		}
	} else {
		// TODO: Make error again
		glog.Errorf("Cannot apply changes to Instance: %v", changes)
		//		return fmt.Errorf("Cannot apply changes to Instance: %v", changes)
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
		glog.Exitf("Cannot parse image spec: %q", nameSpec)
	}

	return fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/images/%s", project, name)
}

func ShortenImageURL(imageURL string) (string, error) {
	u, err := gce.ParseGoogleCloudURL(imageURL)
	if err != nil {
		return "", err
	}
	return u.Project + "/" + u.Name, nil
}

func (_ *Instance) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Instance) error {
	project := t.Project

	// This is a "little" hacky...
	ipAddressResolver := func(ip *IPAddress) (*string, error) {
		tf := "${google_compute_address." + *ip.Name + ".address}"
		return &tf, nil
	}

	i, err := e.mapToGCE(project, ipAddressResolver)
	if err != nil {
		return err
	}

	tf := &terraformInstanceTemplate{
		Name:         i.Name,
		CanIPForward: i.CanIpForward,
		MachineType:  lastComponent(i.MachineType),
		Zone:         i.Zone,
		Tags:         i.Tags.Items,
	}

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

	tf.AddMetadata(i.Metadata)

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
			AutomaticRestart:  i.Scheduling.AutomaticRestart,
			OnHostMaintenance: i.Scheduling.OnHostMaintenance,
			Preemptible:       i.Scheduling.Preemptible,
		}
	}

	return t.RenderResource("google_compute_instance", i.Name, tf)
}
