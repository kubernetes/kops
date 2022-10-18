package scalewaytasks

import (
	"bytes"
	"fmt"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
)

// +kops:fitask
type Instance struct {
	Name      *string
	Lifecycle fi.Lifecycle

	Zone           *string
	CommercialType *string
	Image          *string
	Tags           []string
	Count          int
	UserData       *fi.Resource
}

var _ fi.Task = &Instance{}
var _ fi.CompareWithID = &Instance{}

func (s *Instance) CompareWithID() *string {
	return s.Name
}

func (s *Instance) Find(c *fi.Context) (*Instance, error) {
	cloud := c.Cloud.(scaleway.ScwCloud)

	servers, err := cloud.GetClusterServers(cloud.ClusterName(s.Tags), s.Name)
	if err != nil {
		return nil, fmt.Errorf("error finding instances: %w", err)
	}
	if len(servers) == 0 {
		return nil, nil
	}

	server := servers[0]

	return &Instance{
		Name:           fi.String(server.Name),
		Count:          len(servers),
		Zone:           fi.String(server.Zone.String()),
		CommercialType: fi.String(server.CommercialType),
		Image:          s.Image,
		Tags:           server.Tags,
		UserData:       s.UserData,
		Lifecycle:      s.Lifecycle,
	}, nil
}

func (s *Instance) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(s, c)
}

func (_ *Instance) RenderScw(c *fi.Context, a, e, changes *Instance) error {
	cloud := c.Cloud.(scaleway.ScwCloud)
	instanceService := cloud.InstanceService()
	zone := scw.Zone(fi.StringValue(e.Zone))

	userData, err := fi.ResourceAsBytes(*e.UserData)
	if err != nil {
		return fmt.Errorf("error rendering instances: %w", err)
	}

	var newInstanceCount int
	if a == nil {
		newInstanceCount = e.Count
	} else {
		expectedCount := e.Count
		actualCount := a.Count

		if expectedCount == actualCount {
			return nil
		}

		if actualCount > expectedCount {
			igInstances, err := cloud.GetClusterServers(cloud.ClusterName(a.Tags), a.Name)
			if err != nil {
				return fmt.Errorf("error deleting instance: %w", err)
			}
			for _, igInstance := range igInstances {
				err = cloud.DeleteServer(igInstance)
				if err != nil {
					return fmt.Errorf("error deleting instance of group %s: %w", igInstance.Name, err)
				}
				actualCount--
				if expectedCount == actualCount {
					break
				}
			}
		}

		newInstanceCount = expectedCount - actualCount
	}

	for i := 0; i < newInstanceCount; i++ {

		// We create the instance
		srv, err := instanceService.CreateServer(&instance.CreateServerRequest{
			Zone:           zone,
			Name:           fi.StringValue(e.Name),
			CommercialType: fi.StringValue(e.CommercialType),
			Image:          fi.StringValue(e.Image),
			Tags:           e.Tags,
		})
		if err != nil {
			return fmt.Errorf("error creating instance of group %q: %w", fi.StringValue(e.Name), err)
		}

		// We wait for the instance to be ready
		_, err = instanceService.WaitForServer(&instance.WaitForServerRequest{
			ServerID: srv.Server.ID,
			Zone:     zone,
		})
		if err != nil {
			return fmt.Errorf("error waiting for instance %s of group %q: %w", srv.Server.ID, fi.StringValue(e.Name), err)
		}

		// We load the cloud-init script in the instance user data
		err = instanceService.SetServerUserData(&instance.SetServerUserDataRequest{
			ServerID: srv.Server.ID,
			Zone:     srv.Server.Zone,
			Key:      "cloud-init",
			Content:  bytes.NewBuffer(userData),
		})
		if err != nil {
			return fmt.Errorf("error setting 'cloud-init' in user-data for instance %s of group %q: %w", srv.Server.ID, fi.StringValue(e.Name), err)
		}

		// We start the instance
		_, err = instanceService.ServerAction(&instance.ServerActionRequest{
			Zone:     zone,
			ServerID: srv.Server.ID,
			Action:   instance.ServerActionPoweron,
		})
		if err != nil {
			return fmt.Errorf("error powering on instance %s of group %q: %w", srv.Server.ID, fi.StringValue(e.Name), err)
		}

		// We wait for the instance to be ready
		_, err = instanceService.WaitForServer(&instance.WaitForServerRequest{
			ServerID: srv.Server.ID,
			Zone:     zone,
		})
		if err != nil {
			return fmt.Errorf("error waiting for instance %s of group %q: %w", srv.Server.ID, fi.StringValue(e.Name), err)
		}
	}

	return nil
}

func (_ *Instance) CheckChanges(a, e, changes *Instance) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.Zone != nil {
			return fi.CannotChangeField("Zone")
		}
		if changes.CommercialType != nil {
			return fi.CannotChangeField("CommercialType")
		}
		if changes.Image != nil {
			return fi.CannotChangeField("Image")
		}
	} else {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		if e.Zone == nil {
			return fi.RequiredField("Zone")
		}
		if e.CommercialType == nil {
			return fi.RequiredField("CommercialType")
		}
		if e.Image == nil {
			return fi.RequiredField("Image")
		}
	}
	return nil
}
