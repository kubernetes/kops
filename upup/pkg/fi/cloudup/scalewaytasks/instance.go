/*
Copyright 2022 The Kubernetes Authors.

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

package scalewaytasks

import (
	"bytes"
	"fmt"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
)

// +kops:fitask
type Instance struct {
	Name      *string
	Lifecycle fi.Lifecycle

	Zone           *string
	Role           *string
	CommercialType *string
	Image          *string
	Tags           []string
	Count          int
	UserData       *fi.Resource
}

var _ fi.CloudupTask = &Instance{}
var _ fi.CompareWithID = &Instance{}

func (s *Instance) CompareWithID() *string {
	return s.Name
}

func (s *Instance) Find(c *fi.CloudupContext) (*Instance, error) {
	cloud := c.T.Cloud.(scaleway.ScwCloud)

	servers, err := cloud.GetClusterServers(cloud.ClusterName(s.Tags), s.Name)
	if err != nil {
		return nil, fmt.Errorf("error finding instances: %w", err)
	}
	if len(servers) == 0 {
		return nil, nil
	}
	server := servers[0]

	role := scaleway.TagRoleWorker
	for _, tag := range server.Tags {
		if tag == scaleway.TagNameRolePrefix+"="+scaleway.TagRoleControlPlane {
			role = scaleway.TagRoleControlPlane
		}
	}

	return &Instance{
		Name:           fi.PtrTo(server.Name),
		Count:          len(servers),
		Zone:           fi.PtrTo(server.Zone.String()),
		Role:           fi.PtrTo(role),
		CommercialType: fi.PtrTo(server.CommercialType),
		Image:          s.Image,
		Tags:           server.Tags,
		UserData:       s.UserData,
		Lifecycle:      s.Lifecycle,
	}, nil
}

func (s *Instance) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(s, c)
}

func (_ *Instance) CheckChanges(actual, expected, changes *Instance) error {
	if actual != nil {
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
		if expected.Name == nil {
			return fi.RequiredField("Name")
		}
		if expected.Zone == nil {
			return fi.RequiredField("Zone")
		}
		if expected.CommercialType == nil {
			return fi.RequiredField("CommercialType")
		}
		if expected.Image == nil {
			return fi.RequiredField("Image")
		}
	}
	return nil
}

func (_ *Instance) RenderScw(c *fi.CloudupContext, actual, expected, changes *Instance) error {
	cloud := c.T.Cloud.(scaleway.ScwCloud)
	instanceService := cloud.InstanceService()
	zone := scw.Zone(fi.ValueOf(expected.Zone))
	controlPlanePrivateIPs := []string(nil)

	userData, err := fi.ResourceAsBytes(*expected.UserData)
	if err != nil {
		return fmt.Errorf("error rendering instances: %w", err)
	}

	newInstanceCount := expected.Count
	if actual != nil {
		if expected.Count == actual.Count {
			return nil
		}
		newInstanceCount = expected.Count - actual.Count
	}

	// If newInstanceCount > 0, we need to create new instances for this group
	for i := 0; i < newInstanceCount; i++ {

		// We create the instance
		srv, err := instanceService.CreateServer(&instance.CreateServerRequest{
			Zone:           zone,
			Name:           fi.ValueOf(expected.Name),
			CommercialType: fi.ValueOf(expected.CommercialType),
			Image:          fi.ValueOf(expected.Image),
			Tags:           expected.Tags,
		})
		if err != nil {
			return fmt.Errorf("error creating instance of group %q: %w", fi.ValueOf(expected.Name), err)
		}

		// We wait for the instance to be ready
		_, err = instanceService.WaitForServer(&instance.WaitForServerRequest{
			ServerID: srv.Server.ID,
			Zone:     zone,
		})
		if err != nil {
			return fmt.Errorf("error waiting for instance %s of group %q: %w", srv.Server.ID, fi.ValueOf(expected.Name), err)
		}

		// We load the cloud-init script in the instance user data
		err = instanceService.SetServerUserData(&instance.SetServerUserDataRequest{
			ServerID: srv.Server.ID,
			Zone:     srv.Server.Zone,
			Key:      "cloud-init",
			Content:  bytes.NewBuffer(userData),
		})
		if err != nil {
			return fmt.Errorf("error setting 'cloud-init' in user-data for instance %s of group %q: %w", srv.Server.ID, fi.ValueOf(expected.Name), err)
		}

		// We start the instance
		_, err = instanceService.ServerAction(&instance.ServerActionRequest{
			Zone:     zone,
			ServerID: srv.Server.ID,
			Action:   instance.ServerActionPoweron,
		})
		if err != nil {
			return fmt.Errorf("error powering on instance %s of group %q: %w", srv.Server.ID, fi.ValueOf(expected.Name), err)
		}

		// We wait for the instance to be ready
		_, err = instanceService.WaitForServer(&instance.WaitForServerRequest{
			ServerID: srv.Server.ID,
			Zone:     zone,
		})
		if err != nil {
			return fmt.Errorf("error waiting for instance %s of group %q: %w", srv.Server.ID, fi.ValueOf(expected.Name), err)
		}

		// We update the server's infos (to get its IP)
		server, err := instanceService.GetServer(&instance.GetServerRequest{
			Zone:     zone,
			ServerID: srv.Server.ID,
		})
		if err != nil {
			return fmt.Errorf("getting server %s: %s", srv.Server.ID, err)
		}

		// If instance has control-plane role, we add its private IP to the list to add it to the lb's backend
		for _, tag := range expected.Tags {
			if tag == scaleway.TagNameRolePrefix+"="+scaleway.TagRoleControlPlane {
				controlPlanePrivateIPs = append(controlPlanePrivateIPs, *server.Server.PrivateIP)
			}
		}
	}

	// If IG is control-plane, we add the new servers' IPs to the load-balancer's back-end
	if len(controlPlanePrivateIPs) > 0 {
		lbService := cloud.LBService()
		zone := scw.Zone(cloud.Zone())

		lbs, err := cloud.GetClusterLoadBalancers(cloud.ClusterName(expected.Tags))
		if err != nil {
			return fmt.Errorf("listing load-balancers for instance creation: %w", err)
		}

		for _, loadBalancer := range lbs {
			backEnds, err := lbService.ListBackends(&lb.ZonedAPIListBackendsRequest{
				Zone: zone,
				LBID: loadBalancer.ID,
			})
			if err != nil {
				return fmt.Errorf("listing load-balancer's back-ends for instance creation: %w", err)
			}
			if backEnds.TotalCount > 1 {
				return fmt.Errorf("cannot have multiple back-ends for load-balancer %s", loadBalancer.ID)
			}
			backEnd := backEnds.Backends[0]

			_, err = lbService.AddBackendServers(&lb.ZonedAPIAddBackendServersRequest{
				Zone:      zone,
				BackendID: backEnd.ID,
				ServerIP:  controlPlanePrivateIPs,
			})
			if err != nil {
				return fmt.Errorf("adding servers' IPs to load-balancer's back-end: %w", err)
			}

			_, err = lbService.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
				LBID: loadBalancer.ID,
				Zone: zone,
			})
			if err != nil {
				return fmt.Errorf("waiting for load-balancer %s: %w", loadBalancer.ID, err)
			}
		}
	}

	// If newInstanceCount < 0, we need to delete instances of this group
	if newInstanceCount < 0 {

		igInstances, err := cloud.GetClusterServers(cloud.ClusterName(actual.Tags), actual.Name)
		if err != nil {
			return fmt.Errorf("error deleting instance: %w", err)
		}

		for i := 0; i > newInstanceCount; i-- {
			toDelete := igInstances[i*-1]
			err = cloud.DeleteServer(toDelete)
			if err != nil {
				return fmt.Errorf("error deleting instance of group %s: %w", toDelete.Name, err)
			}
		}
	}

	return nil
}
