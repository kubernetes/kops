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
	"crypto/sha256"
	"fmt"
	"io"
	"strings"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/api/marketplace/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

var commercialTypesWithBlockStorageOnly = []string{"PRO", "PLAY", "ENT"}

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
	NeedsUpdate    []string

	UserData     *fi.Resource
	LoadBalancer *LoadBalancer
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

	// Check if servers updates are needed
	var needsUpdate []string
	for _, server := range servers {

		// Check if server is already marked as needing update
		alreadyTagged := false
		for _, tag := range server.Tags {
			if tag == scaleway.TagNeedsUpdate {
				alreadyTagged = true
			}
		}
		if alreadyTagged == true {
			continue
		}

		// Check commercial type differences
		if server.CommercialType != *s.CommercialType {
			needsUpdate = append(needsUpdate, server.ID)
			continue
		}

		// Check image differences
		diff, err := checkImageDifferences(c, cloud, server, fi.ValueOf(s.Image))
		if err != nil {
			return nil, fmt.Errorf("checking image differences in server %s (%s): %w", server.Name, server.ID, err)
		}
		if diff == true {
			needsUpdate = append(needsUpdate, server.ID)
			continue
		}

		// Check user-data differences
		diff, err = checkUserDataDifferences(c, cloud, server, s.UserData)
		if err != nil {
			return nil, fmt.Errorf("checking user-data differences in server %s (%s): %w", server.Name, server.ID, err)
		}
		if diff == true {
			needsUpdate = append(needsUpdate, server.ID)
		}
	}

	server := servers[0]

	igName := ""
	for _, tag := range server.Tags {
		if strings.HasPrefix(tag, scaleway.TagInstanceGroup) {
			igName = strings.TrimPrefix(tag, scaleway.TagInstanceGroup+"=")
		}
	}

	role := scaleway.TagRoleWorker
	for _, tag := range server.Tags {
		if tag == scaleway.TagNameRolePrefix+"="+scaleway.TagRoleControlPlane {
			role = scaleway.TagRoleControlPlane
		}
	}

	imageLabel, err := imageLabelFromID(c, cloud, server.Image.ID)
	if err != nil {
		return nil, err
	}

	return &Instance{
		Name:           fi.PtrTo(igName),
		Lifecycle:      s.Lifecycle,
		Zone:           fi.PtrTo(server.Zone.String()),
		Role:           fi.PtrTo(role),
		CommercialType: fi.PtrTo(server.CommercialType),
		Image:          fi.PtrTo(imageLabel),
		Tags:           server.Tags,
		Count:          len(servers),
		NeedsUpdate:    needsUpdate,
		UserData:       s.UserData,
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

func (_ *Instance) RenderScw(t *scaleway.ScwAPITarget, actual, expected, changes *Instance) error {
	cloud := t.Cloud.(scaleway.ScwCloud)
	instanceService := cloud.InstanceService()
	zone := scw.Zone(fi.ValueOf(expected.Zone))
	controlPlanePrivateIPs := []string(nil)

	userData, err := fi.ResourceAsBytes(*expected.UserData)
	if err != nil {
		return fmt.Errorf("error rendering instances: %w", err)
	}

	newInstanceCount := expected.Count
	if actual != nil {

		// Add "kops.k8s.io/needs-update" label to servers needing update
		for _, serverID := range actual.NeedsUpdate {
			server, err := instanceService.GetServer(&instance.GetServerRequest{
				Zone:     zone,
				ServerID: serverID,
			})
			if err != nil {
				return fmt.Errorf("rendering server group: listing existing servers: %w", err)
			}
			_, err = instanceService.UpdateServer(&instance.UpdateServerRequest{
				Zone:     zone,
				ServerID: serverID,
				Tags:     scw.StringsPtr(append(server.Server.Tags, scaleway.TagNeedsUpdate)),
			})
			if err != nil {
				return fmt.Errorf("rendering server group: adding update tag to server %q (%s): %w", server.Server.Name, serverID, err)
			}
		}

		if expected.Count == actual.Count {
			return nil
		}
		newInstanceCount = expected.Count - actual.Count

	}

	// If newInstanceCount > 0, we need to create new instances for this group
	for i := 0; i < newInstanceCount; i++ {
		// We create a unique name for each server
		actualCount := 0
		if actual != nil {
			actualCount = actual.Count
		}
		uniqueName := fmt.Sprintf("%s-%d", fi.ValueOf(expected.Name), i+actualCount)

		// If the instance's commercial type is one that has no local storage, we have to specify for the
		// block storage volume a big enough size (default size is 10GB)
		commercialType := fi.ValueOf(expected.CommercialType)
		createServerRequest := instance.CreateServerRequest{
			Zone:           zone,
			Name:           uniqueName,
			CommercialType: commercialType,
			Image:          fi.ValueOf(expected.Image),
			Tags:           expected.Tags,
		}
		for _, ct := range commercialTypesWithBlockStorageOnly {
			if strings.HasPrefix(commercialType, ct) {
				continue
			}
			createServerRequest.Volumes = map[string]*instance.VolumeServerTemplate{
				"0": {
					Boot:       fi.PtrTo(true),
					Size:       fi.PtrTo(scw.GB * 50),
					VolumeType: instance.VolumeVolumeTypeBSSD,
				},
			}
		}

		// We create the instance and wait for it to be ready
		srv, err := instanceService.CreateServer(&createServerRequest)
		if err != nil {
			return fmt.Errorf("error creating instance of group %q: %w", fi.ValueOf(expected.Name), err)
		}
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

		// If instance has control-plane role, we add its private IP to the list to add it to the lb's backend
		if fi.ValueOf(expected.Role) == scaleway.TagRoleControlPlane {

			// We update the server's infos (to get its IP)
			server, err := instanceService.GetServer(&instance.GetServerRequest{
				Zone:     zone,
				ServerID: srv.Server.ID,
			})
			if err != nil {
				return fmt.Errorf("getting server %s: %s", srv.Server.ID, err)
			}
			controlPlanePrivateIPs = append(controlPlanePrivateIPs, *server.Server.PrivateIP)
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

			if fi.ValueOf(actual.Role) == scaleway.TagRoleControlPlane {
				controlPlanePrivateIPs = append(controlPlanePrivateIPs, *toDelete.PrivateIP)
			}

			err = cloud.DeleteServer(toDelete)
			if err != nil {
				return fmt.Errorf("error deleting instance of group %s: %w", toDelete.Name, err)
			}
		}
	}

	// If IG is control-plane, we need to update the load-balancer's back-end
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

			for _, backEnd := range backEnds.Backends {
				// If we are adding instances, we also need to add them to the load-balancer's backend
				if newInstanceCount > 0 {
					_, err = lbService.AddBackendServers(&lb.ZonedAPIAddBackendServersRequest{
						Zone:      zone,
						BackendID: backEnd.ID,
						ServerIP:  controlPlanePrivateIPs,
					})
					if err != nil {
						return fmt.Errorf("adding servers' IPs to load-balancer's back-end: %w", err)
					}

				} else {
					// If we are deleting instances, we also need to delete them from the load-balancer's backend
					_, err = lbService.RemoveBackendServers(&lb.ZonedAPIRemoveBackendServersRequest{
						Zone:      zone,
						BackendID: backEnd.ID,
						ServerIP:  controlPlanePrivateIPs,
					})
					if err != nil {
						return fmt.Errorf("removing servers' IPs from load-balancer's back-end: %w", err)
					}
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
	}

	return nil
}

type terraformInstanceIP struct{}

type terraformInstance struct {
	Name     *string                             `cty:"name"`
	IPID     *terraformWriter.Literal            `cty:"ip_id"`
	Type     *string                             `cty:"type"`
	Tags     []string                            `cty:"tags"`
	Image    *string                             `cty:"image"`
	UserData map[string]*terraformWriter.Literal `cty:"user_data"`
}

func (_ *Instance) RenderTerraform(t *terraform.TerraformTarget, actual, expected, changes *Instance) error {
	tfName := strings.ReplaceAll(fi.ValueOf(expected.Name), ".", "-")

	tfInstanceIP := terraformInstanceIP{}
	err := t.RenderResource("scaleway_instance_ip", tfName, tfInstanceIP)
	if err != nil {
		return err
	}

	tfInstance := terraformInstance{
		Name:  expected.Name,
		IPID:  terraformWriter.LiteralProperty("scaleway_instance_ip", tfName, "id"),
		Type:  expected.CommercialType,
		Tags:  expected.Tags,
		Image: expected.Image,
	}
	if expected.UserData != nil {
		userDataBytes, err := fi.ResourceAsBytes(fi.ValueOf(expected.UserData))
		if err != nil {
			return err
		}
		if userDataBytes != nil {
			tfUserData, err := t.AddFileBytes("scaleway_instance_server", tfName, "user_data", userDataBytes, true)
			if err != nil {
				return err
			}
			tfInstance.UserData = map[string]*terraformWriter.Literal{
				"cloud-init": tfUserData}
		}
	}
	return t.RenderResource("scaleway_instance_server", tfName, tfInstance)
}

func checkImageDifferences(c *fi.CloudupContext, cloud scaleway.ScwCloud, actualServer *instance.Server, expectedImage string) (bool, error) {
	localImage, err := cloud.MarketplaceService().GetLocalImageByLabel(&marketplace.GetLocalImageByLabelRequest{
		ImageLabel:     expectedImage,
		Zone:           actualServer.Zone,
		CommercialType: actualServer.CommercialType,
	}, scw.WithContext(c.Context()))
	if err != nil {
		return false, fmt.Errorf("getting image from the marketplace: %w", err)
	}

	if actualServer.Image.ID != localImage.ID {
		return true, nil
	}
	return false, nil
}

func checkUserDataDifferences(c *fi.CloudupContext, cloud scaleway.ScwCloud, actualServer *instance.Server, expectedUserData *fi.Resource) (bool, error) {
	actualUserData, err := cloud.InstanceService().GetServerUserData(&instance.GetServerUserDataRequest{
		Zone:     actualServer.Zone,
		ServerID: actualServer.ID,
		Key:      "cloud-init",
	}, scw.WithContext(c.Context()))
	if err != nil {
		return false, fmt.Errorf("getting actual user-data: %w", err)
	}

	actualUserDataBytes, err := io.ReadAll(actualUserData)
	if err != nil {
		return false, fmt.Errorf("reading actual user-data: %w", err)
	}
	expectedUserDataBytes, err := fi.ResourceAsBytes(*expectedUserData)
	if err != nil {
		return false, fmt.Errorf("reading expected user-data: %w", err)
	}

	if sha256.Sum256(actualUserDataBytes) != sha256.Sum256(expectedUserDataBytes) {
		return true, nil
	}
	return false, nil
}

func imageLabelFromID(c *fi.CloudupContext, cloud scaleway.ScwCloud, id string) (string, error) {
	localImage, err := cloud.MarketplaceService().GetLocalImage(&marketplace.GetLocalImageRequest{
		LocalImageID: id,
	}, scw.WithContext(c.Context()))
	if err != nil {
		return "", fmt.Errorf("getting image from the marketplace: %w", err)
	}
	return localImage.Label, nil
}
