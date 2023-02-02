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
	"strings"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
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

	UserData     *fi.Resource
	LoadBalancer *LoadBalancer
	//Network        *Network
	NeedsUpdate []string
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

	// Check if servers have been added to the instance group, therefore an update is needed
	if len(servers) > s.Count {
		for _, server := range servers {
			alreadyTagged := false
			for _, tag := range server.Tags {
				if tag == scaleway.TagNeedsUpdate {
					alreadyTagged = true
				}
			}
			if alreadyTagged == true {
				continue
			}
			s.NeedsUpdate = append(s.NeedsUpdate, server.ID)
		}
	}
	//TODO(Mia-Cross): handle other changes like image, commercial type, userdata

	server := servers[0]

	igName := ""
	for _, tag := range server.Tags {
		if strings.HasPrefix(tag, scaleway.TagInstanceGroup) {
			igName = strings.TrimPrefix(tag, scaleway.TagInstanceGroup+"=")
		}
	}

	role := scaleway.TagRoleNode
	for _, tag := range server.Tags {
		if tag == scaleway.TagNameRolePrefix+"="+scaleway.TagRoleControlPlane {
			role = scaleway.TagRoleControlPlane
		}
	}

	return &Instance{
		Name:           fi.PtrTo(igName),
		Count:          len(servers),
		Zone:           fi.PtrTo(server.Zone.String()),
		Role:           fi.PtrTo(role),
		CommercialType: fi.PtrTo(server.CommercialType),
		Image:          s.Image,
		Tags:           server.Tags,
		UserData:       s.UserData,
		Lifecycle:      s.Lifecycle,
		//Network:        s.Network,
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
		if expected.Count == actual.Count {
			return nil
		}
		newInstanceCount = expected.Count - actual.Count

		// Add "kops.k8s.io/needs-update" label to servers needing update
		for _, serverID := range actual.NeedsUpdate {
			server, err := instanceService.GetServer(&instance.GetServerRequest{
				Zone:     zone,
				ServerID: serverID,
			})
			if err != nil {
				return fmt.Errorf("error rendering server group: error listing existing servers: %w", err)
			}
			_, err = instanceService.UpdateServer(&instance.UpdateServerRequest{
				Zone:     zone,
				ServerID: serverID,
				Tags:     scw.StringsPtr(append(server.Server.Tags, scaleway.TagNeedsUpdate)),
			})
			if err != nil {
				return fmt.Errorf("error rendering server group: error adding update tag to server %q (%s): %w", server.Server.Name, serverID, err)
			}
		}
	}

	// We get the private network to associate it with new instances
	//pn, err := cloud.GetClusterVPCs(c.Cluster.Name)
	//if err != nil {
	//	return fmt.Errorf("error listing private networks: %v", err)
	//}
	//if len(pn) != 1 {
	//	return fmt.Errorf("more than 1 private network named %s found", c.Cluster.Name)
	//}

	// If newInstanceCount > 0, we need to create new instances for this group
	for i := 0; i < newInstanceCount; i++ {

		// We create a unique name for each server
		actualCount := 0
		if actual != nil {
			actualCount = actual.Count
		}
		// TODO(Mia-Cross): check that this works even when instances were deleted before adding some again
		uniqueName := fmt.Sprintf("%s-%d", fi.ValueOf(expected.Name), i+actualCount)

		// We create the instance
		srv, err := instanceService.CreateServer(&instance.CreateServerRequest{
			Zone:           zone,
			Name:           uniqueName,
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

		// We put the instance inside the private network
		//pNIC, err := instanceService.CreatePrivateNIC(&instance.CreatePrivateNICRequest{
		//	Zone:             zone,
		//	ServerID:         srv.Server.ID,
		//	PrivateNetworkID: pn[0].ID,
		//})
		//if err != nil {
		//	return fmt.Errorf("error linking instance to private network: %v", err)
		//}
		//
		//// We wait for the private nic to be ready before proceeding
		//_, err = instanceService.WaitForPrivateNIC(&instance.WaitForPrivateNICRequest{
		//	ServerID:     srv.Server.ID,
		//	PrivateNicID: pNIC.PrivateNic.ID,
		//	Zone:         zone,
		//})
		//if err != nil {
		//	return fmt.Errorf("error waiting for private nic: %v", err)
		//}
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
			if backEnds.TotalCount > 1 {
				return fmt.Errorf("cannot have multiple back-ends for load-balancer %s", loadBalancer.Name)
			} else if backEnds.TotalCount < 1 {
				return fmt.Errorf("load-balancer %s should have 1 back-end, got 0", loadBalancer.Name)
			}
			backEnd := backEnds.Backends[0]

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

	// We create NAT rules linking the gateway to our instances in order to be able to connect via SSH
	// TODO(Mia-Cross): This part is for dev purposes only, remove when done
	//gwService := cloud.GatewayService()
	//rules := []*vpcgw.SetPATRulesRequestRule(nil)
	//port := uint32(2022)
	//gwNetwork, err := cloud.GetClusterGatewayNetworks(pn[0].ID)
	//if err != nil {
	//	return err
	//}
	//if len(gwNetwork) < 1 {
	//	klog.V(4).Infof("Could not find any gateway connexion, skipping NAT rules creation")
	//} else {
	//	entries, err := gwService.ListDHCPEntries(&vpcgw.ListDHCPEntriesRequest{
	//		Zone:             zone,
	//		GatewayNetworkID: scw.StringPtr(gwNetwork[0].ID),
	//	}, scw.WithAllPages())
	//	if err != nil {
	//		return fmt.Errorf("error listing DHCP entries")
	//	}
	//	klog.V(4).Infof("=== DHCP entries are %v", entries.DHCPEntries)
	//	for _, entry := range entries.DHCPEntries {
	//		rules = append(rules, &vpcgw.SetPATRulesRequestRule{
	//			PublicPort:  port,
	//			PrivateIP:   entry.IPAddress,
	//			PrivatePort: 22,
	//			Protocol:    "both",
	//		})
	//		port += 1
	//	}
	//
	//	_, err = gwService.SetPATRules(&vpcgw.SetPATRulesRequest{
	//		Zone:      zone,
	//		GatewayID: gwNetwork[0].GatewayID,
	//		PatRules:  rules,
	//	})
	//	if err != nil {
	//		return fmt.Errorf("error setting PAT rules for gateway")
	//	}
	//	klog.V(4).Infof("=== rules set")
	//}

	return nil
}

type terraformInstanceIP struct{}

type terraformUserData struct {
	CloudInit *terraformWriter.Literal `cty:"cloud-init"`
}

type terraformInstance struct {
	Name     *string                             `cty:"name"`
	IPID     *terraformWriter.Literal            `cty:"ip_id"`
	Type     *string                             `cty:"type"`
	Tags     []string                            `cty:"tags"`
	Image    *string                             `cty:"image"`
	UserData map[string]*terraformWriter.Literal `cty:"user_data"`
}

func (_ *Instance) RenderTerraform(t *terraform.TerraformTarget, actual, expected, changes *Instance) error {
	tfName := strings.Replace(fi.ValueOf(expected.Name), ".", "-", -1)
	{
		tf := terraformInstanceIP{}
		err := t.RenderResource("scaleway_instance_ip", tfName, tf)
		if err != nil {
			return err
		}
	}
	{
		tf := terraformInstance{
			Name:  expected.Name,
			IPID:  expected.TerraformLinkIPID(tfName),
			Type:  expected.CommercialType,
			Tags:  expected.Tags,
			Image: expected.Image,
			//UserData: expected.UserData,
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
				tf.UserData = map[string]*terraformWriter.Literal{
					"cloud-init": tfUserData,
				}
				//tf.UserData, err =
			}
		}

		return t.RenderResource("scaleway_instance_server", tfName, tf)
	}
}

func (i *Instance) TerraformLinkIPID(tfName string) *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("scaleway_instance_ip", tfName, "id")
}
