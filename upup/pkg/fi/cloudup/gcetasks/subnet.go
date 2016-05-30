package gcetasks

import (
	"fmt"

	"github.com/golang/glog"
	"google.golang.org/api/compute/v1"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/gce"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=Subnet
type Subnet struct {
	Name    *string
	Network *Network
	Region  *string
	CIDR    *string
}

var _ fi.CompareWithID = &Subnet{}

func (e *Subnet) CompareWithID() *string {
	return e.Name
}

func (e *Subnet) Find(c *fi.Context) (*Subnet, error) {
	cloud := c.Cloud.(*gce.GCECloud)

	s, err := cloud.Compute.Subnetworks.Get(cloud.Project, cloud.Region, *e.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing Subnets: %v", err)
	}

	actual := &Subnet{}
	actual.Name = &s.Name
	actual.Network = &Network{Name: &s.Network}
	actual.Region = &s.Region
	actual.CIDR = &s.IpCidrRange

	return actual, nil
}

func (e *Subnet) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *Subnet) CheckChanges(a, e, changes *Subnet) error {
	return nil
}

func (_ *Subnet) RenderGCE(t *gce.GCEAPITarget, a, e, changes *Subnet) error {
	if a == nil {
		glog.V(2).Infof("Creating Subnet with CIDR: %q", *e.CIDR)

		subnet := &compute.Subnetwork{
			IpCidrRange: *e.CIDR,
			Name:        *e.Name,
			Network:     *e.Network.Name,
		}
		_, err := t.Cloud.Compute.Subnetworks.Insert(t.Cloud.Project, t.Cloud.Region, subnet).Do()
		if err != nil {
			return fmt.Errorf("error creating Subnet: %v", err)
		}
	}

	return nil
}

type terraformSubnet struct {
	Name    *string            `json:"name"`
	Network *terraform.Literal `json:"network"`
	Region  *string            `json:"region"`
	CIDR    *string            `json:"ip_cidr_range"`
}

func (_ *Subnet) RenderSubnet(t *terraform.TerraformTarget, a, e, changes *Subnet) error {
	tf := &terraformSubnet{
		Name:    e.Name,
		Network: e.Network.TerraformName(),
		Region:  e.Region,
		CIDR:    e.CIDR,
	}
	return t.RenderResource("google_compute_subnetwork", *e.Name, tf)
}

func (i *Subnet) TerraformName() *terraform.Literal {
	return terraform.LiteralProperty("google_compute_subnetwork", *i.Name, "name")
}
