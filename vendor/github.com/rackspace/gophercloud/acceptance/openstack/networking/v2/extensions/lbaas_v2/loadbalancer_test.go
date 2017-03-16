// +build acceptance networking lbaas_v2 lbaasloadbalancer

package lbaas_v2

import (
	"testing"
	"time"

	"github.com/rackspace/gophercloud"
	base "github.com/rackspace/gophercloud/acceptance/openstack/networking/v2"
	"github.com/rackspace/gophercloud/openstack/networking/v2/extensions/lbaas_v2/listeners"
	"github.com/rackspace/gophercloud/openstack/networking/v2/extensions/lbaas_v2/loadbalancers"
	"github.com/rackspace/gophercloud/openstack/networking/v2/extensions/lbaas_v2/monitors"
	"github.com/rackspace/gophercloud/openstack/networking/v2/extensions/lbaas_v2/pools"
	"github.com/rackspace/gophercloud/openstack/networking/v2/networks"
	"github.com/rackspace/gophercloud/openstack/networking/v2/subnets"
	"github.com/rackspace/gophercloud/pagination"
	th "github.com/rackspace/gophercloud/testhelper"
)

// Note: when creating a new Loadbalancer (VM), it can take some time before it is ready for use,
// this timeout is used for waiting until the Loadbalancer provisioning status goes to ACTIVE state.
const loadbalancerActiveTimeoutSeconds = 120
const loadbalancerDeleteTimeoutSeconds = 10

func setupTopology(t *testing.T) (string, string) {
	// create network
	n, err := networks.Create(base.Client, networks.CreateOpts{Name: "tmp_network"}).Extract()
	th.AssertNoErr(t, err)

	t.Logf("Created network, ID %s", n.ID)

	// create subnet
	s, err := subnets.Create(base.Client, subnets.CreateOpts{
		NetworkID: n.ID,
		CIDR:      "192.168.199.0/24",
		IPVersion: subnets.IPv4,
		Name:      "tmp_subnet",
	}).Extract()
	th.AssertNoErr(t, err)

	t.Logf("Created subnet, ID %s", s.ID)

	return n.ID, s.ID
}

func deleteTopology(t *testing.T, networkID string) {
	res := networks.Delete(base.Client, networkID)
	th.AssertNoErr(t, res.Err)
	t.Logf("deleted network, ID %s", networkID)
}

func TestLoadbalancers(t *testing.T) {
	base.Setup(t)
	defer base.Teardown()

	// setup network topology
	networkID, subnetID := setupTopology(t)

	// create Loadbalancer
	loadbalancerID := createLoadbalancer(t, subnetID)

	// list Loadbalancers
	listLoadbalancers(t)

	// get Loadbalancer and wait until ACTIVE
	getLoadbalancerWaitActive(t, loadbalancerID)

	// update Loadbalancer
	updateLoadbalancer(t, loadbalancerID)

	// get Loadbalancer and wait until ACTIVE
	getLoadbalancerWaitActive(t, loadbalancerID)

	// create listener
	listenerID := createListener(t, listeners.ProtocolHTTP, 80, loadbalancerID)

	// list listeners
	listListeners(t)

	// get Loadbalancer and wait until ACTIVE
	getLoadbalancerWaitActive(t, loadbalancerID)

	// update listener
	updateListener(t, listenerID)

	// get listener
	getListener(t, listenerID)

	// get Loadbalancer and wait until ACTIVE
	getLoadbalancerWaitActive(t, loadbalancerID)

	// create pool
	poolID := createPool(t, pools.ProtocolHTTP, listenerID, pools.LBMethodRoundRobin)

	// list pools
	listPools(t)

	// get Loadbalancer and wait until ACTIVE
	getLoadbalancerWaitActive(t, loadbalancerID)

	// update pool
	updatePool(t, poolID)

	// get pool
	getPool(t, poolID)

	// get Loadbalancer and wait until ACTIVE
	getLoadbalancerWaitActive(t, loadbalancerID)

	// create member
	memberID := createMember(t, subnetID, poolID, "1.2.3.4", 80, 5)

	// list members
	listMembers(t, poolID)

	// get Loadbalancer and wait until ACTIVE
	getLoadbalancerWaitActive(t, loadbalancerID)

	// update member
	updateMember(t, poolID, memberID)

	// get member
	getMember(t, poolID, memberID)

	// get Loadbalancer and wait until ACTIVE
	getLoadbalancerWaitActive(t, loadbalancerID)

	// create monitor
	monitorID := createMonitor(t, poolID, monitors.TypePING, 10, 10, 3)

	// list monitors
	listMonitors(t)

	// get Loadbalancer and wait until ACTIVE
	getLoadbalancerWaitActive(t, loadbalancerID)

	// update monitor
	updateMonitor(t, monitorID)

	// get monitor
	getMonitor(t, monitorID)

	// get loadbalancer statuses tree
	rawStatusTree, err := loadbalancers.GetStatuses(base.Client, loadbalancerID).ExtractStatuses()
	if err == nil {
		// verify statuses tree ID's of relevant objects
		if rawStatusTree.Loadbalancer.ID != loadbalancerID {
			t.Errorf("Loadbalancer ID did not match")
		}
		if rawStatusTree.Loadbalancer.Listeners[0].ID != listenerID {
			t.Errorf("Listner ID did not match")
		}
		if rawStatusTree.Loadbalancer.Listeners[0].Pools[0].ID != poolID {
			t.Errorf("Pool ID did not match")
		}
		if rawStatusTree.Loadbalancer.Listeners[0].Pools[0].Members[0].ID != memberID {
			t.Errorf("Member ID did not match")
		}
		if rawStatusTree.Loadbalancer.Listeners[0].Pools[0].Monitor.ID != monitorID {
			t.Errorf("Monitor ID did not match")
		}
	} else {
		t.Errorf("Failed to extract Loadbalancer statuses tree: %v", err)
	}

	getLoadbalancerWaitActive(t, loadbalancerID)
	deleteMonitor(t, monitorID)
	getLoadbalancerWaitActive(t, loadbalancerID)
	deleteMember(t, poolID, memberID)
	getLoadbalancerWaitActive(t, loadbalancerID)
	deletePool(t, poolID)
	getLoadbalancerWaitActive(t, loadbalancerID)
	deleteListener(t, listenerID)
	getLoadbalancerWaitActive(t, loadbalancerID)
	deleteLoadbalancer(t, loadbalancerID)
	getLoadbalancerWaitDeleted(t, loadbalancerID)
	deleteTopology(t, networkID)
}

func createLoadbalancer(t *testing.T, subnetID string) string {
	lb, err := loadbalancers.Create(base.Client, loadbalancers.CreateOpts{
		VipSubnetID:  subnetID,
		Name:         "tmp_loadbalancer",
		AdminStateUp: loadbalancers.Up,
	}).Extract()

	th.AssertNoErr(t, err)
	t.Logf("Created Loadbalancer, ID %s", lb.ID)

	return lb.ID
}

func deleteLoadbalancer(t *testing.T, loadbalancerID string) {
	res := loadbalancers.Delete(base.Client, loadbalancerID)
	th.AssertNoErr(t, res.Err)
	t.Logf("deleted Loadbalancer, ID %s", loadbalancerID)
}

func listLoadbalancers(t *testing.T) {
	err := loadbalancers.List(base.Client, loadbalancers.ListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		loadbalancerList, err := loadbalancers.ExtractLoadbalancers(page)
		if err != nil {
			t.Errorf("Failed to extract Loadbalancers: %v", err)
			return false, err
		}

		for _, loadbalancer := range loadbalancerList {
			t.Logf("Listing Loadbalancer: ID [%s] Name [%s] Address [%s]",
				loadbalancer.ID, loadbalancer.Name, loadbalancer.VipAddress)
		}

		return true, nil
	})

	th.AssertNoErr(t, err)
}

func getLoadbalancerWaitDeleted(t *testing.T, loadbalancerID string) {
	start := time.Now().Second()
	for {
		time.Sleep(1 * time.Second)

		if time.Now().Second()-start >= loadbalancerDeleteTimeoutSeconds {
			t.Errorf("Loadbalancer failed to delete")
			return
		}

		_, err := loadbalancers.Get(base.Client, loadbalancerID).Extract()
		if err != nil {
			if errData, ok := err.(*(gophercloud.UnexpectedResponseCodeError)); ok {
				if errData.Actual == 404 {
					return
				}
			} else {
				th.AssertNoErr(t, err)
			}
		}
	}
}

func getLoadbalancerWaitActive(t *testing.T, loadbalancerID string) {
	start := time.Now().Second()
	for {
		time.Sleep(1 * time.Second)

		if time.Now().Second()-start >= loadbalancerActiveTimeoutSeconds {
			t.Errorf("Loadbalancer failed to go into ACTIVE provisioning status")
			return
		}

		loadbalancer, err := loadbalancers.Get(base.Client, loadbalancerID).Extract()
		th.AssertNoErr(t, err)
		if loadbalancer.ProvisioningStatus == "ACTIVE" {
			t.Logf("Retrieved Loadbalancer, ID [%s]: OperatingStatus [%s]", loadbalancer.ID, loadbalancer.ProvisioningStatus)
			return
		}
	}
}

func updateLoadbalancer(t *testing.T, loadbalancerID string) {
	_, err := loadbalancers.Update(base.Client, loadbalancerID, loadbalancers.UpdateOpts{Name: "tmp_newName"}).Extract()

	th.AssertNoErr(t, err)

	t.Logf("Updated Loadbalancer ID [%s]", loadbalancerID)
}

func listListeners(t *testing.T) {
	err := listeners.List(base.Client, listeners.ListOpts{Name: "tmp_listener"}).EachPage(func(page pagination.Page) (bool, error) {
		listenerList, err := listeners.ExtractListeners(page)
		if err != nil {
			t.Errorf("Failed to extract Listeners: %v", err)
			return false, err
		}

		for _, listener := range listenerList {
			t.Logf("Listing Listener: ID [%s] Name [%s] Loadbalancers [%s]",
				listener.ID, listener.Name, listener.Loadbalancers)
		}

		return true, nil
	})

	th.AssertNoErr(t, err)
}

func createListener(t *testing.T, protocol listeners.Protocol, protocolPort int, loadbalancerID string) string {
	l, err := listeners.Create(base.Client, listeners.CreateOpts{
		Protocol:       protocol,
		ProtocolPort:   protocolPort,
		LoadbalancerID: loadbalancerID,
		Name:           "tmp_listener",
	}).Extract()

	th.AssertNoErr(t, err)
	t.Logf("Created Listener, ID %s", l.ID)

	return l.ID
}

func deleteListener(t *testing.T, listenerID string) {
	res := listeners.Delete(base.Client, listenerID)
	th.AssertNoErr(t, res.Err)
	t.Logf("Deleted Loadbalancer, ID %s", listenerID)
}

func getListener(t *testing.T, listenerID string) {
	listener, err := listeners.Get(base.Client, listenerID).Extract()

	th.AssertNoErr(t, err)

	t.Logf("Getting Listener, ID [%s]: ", listener.ID)
}

func updateListener(t *testing.T, listenerID string) {
	_, err := listeners.Update(base.Client, listenerID, listeners.UpdateOpts{Name: "tmp_newName"}).Extract()

	th.AssertNoErr(t, err)

	t.Logf("Updated Listener, ID [%s]", listenerID)
}

func listPools(t *testing.T) {
	err := pools.List(base.Client, pools.ListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		poolsList, err := pools.ExtractPools(page)
		if err != nil {
			t.Errorf("Failed to extract Pools: %v", err)
			return false, err
		}

		for _, pool := range poolsList {
			t.Logf("Listing Pool: ID [%s] Name [%s] Listeners [%s] LBMethod [%s]",
				pool.ID, pool.Name, pool.Listeners, pool.LBMethod)
		}

		return true, nil
	})

	th.AssertNoErr(t, err)
}

func createPool(t *testing.T, protocol pools.Protocol, listenerID string, lbMethod pools.LBMethod) string {
	p, err := pools.Create(base.Client, pools.CreateOpts{
		LBMethod:   lbMethod,
		Protocol:   protocol,
		Name:       "tmp_pool",
		ListenerID: listenerID,
	}).Extract()

	th.AssertNoErr(t, err)

	t.Logf("Created Pool, ID %s", p.ID)

	return p.ID
}

func deletePool(t *testing.T, poolID string) {
	res := pools.Delete(base.Client, poolID)
	th.AssertNoErr(t, res.Err)
	t.Logf("Deleted Pool, ID %s", poolID)
}

func getPool(t *testing.T, poolID string) {
	pool, err := pools.Get(base.Client, poolID).Extract()

	th.AssertNoErr(t, err)

	t.Logf("Getting Pool, ID [%s]: ", pool.ID)
}

func updatePool(t *testing.T, poolID string) {
	_, err := pools.Update(base.Client, poolID, pools.UpdateOpts{Name: "tmp_newName"}).Extract()

	th.AssertNoErr(t, err)

	t.Logf("Updated Pool, ID [%s]", poolID)
}

func createMember(t *testing.T, subnetID string, poolID string, address string, protocolPort int, weight int) string {
	m, err := pools.CreateAssociateMember(base.Client, poolID, pools.MemberCreateOpts{
		SubnetID:     subnetID,
		Address:      address,
		ProtocolPort: protocolPort,
		Weight:       weight,
		Name:         "tmp_member",
	}).ExtractMember()

	th.AssertNoErr(t, err)

	t.Logf("Created Member, ID %s", m.ID)

	return m.ID
}

func deleteMember(t *testing.T, poolID string, memberID string) {
	res := pools.DeleteMember(base.Client, poolID, memberID)
	th.AssertNoErr(t, res.Err)
	t.Logf("Deleted Member, ID %s", memberID)
}

func listMembers(t *testing.T, poolID string) {
	err := pools.ListAssociateMembers(base.Client, poolID, pools.MemberListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		membersList, err := pools.ExtractMembers(page)
		if err != nil {
			t.Errorf("Failed to extract Members: %v", err)
			return false, err
		}

		for _, member := range membersList {
			t.Logf("Listing Member: ID [%s] Name [%s] Pool ID [%s] Weight [%s]",
				member.ID, member.Name, member.PoolID, member.Weight)
		}

		return true, nil
	})

	th.AssertNoErr(t, err)
}

func getMember(t *testing.T, poolID string, memberID string) {
	member, err := pools.GetAssociateMember(base.Client, poolID, memberID).ExtractMember()

	th.AssertNoErr(t, err)

	t.Logf("Getting Member, ID [%s]: ", member.ID)
}

func updateMember(t *testing.T, poolID string, memberID string) {
	_, err := pools.UpdateAssociateMember(base.Client, poolID, memberID, pools.MemberUpdateOpts{Name: "tmp_newName"}).Extract()

	th.AssertNoErr(t, err)

	t.Logf("Updated Member, ID [%s], in Pool, ID [%s]", memberID, poolID)
}

func createMonitor(t *testing.T, poolID string, checkType string, delay int, timeout int, maxRetries int) string {
	m, err := monitors.Create(base.Client, monitors.CreateOpts{
		PoolID:     poolID,
		Name:       "tmp_monitor",
		Delay:      delay,
		Timeout:    timeout,
		MaxRetries: maxRetries,
		Type:       checkType,
	}).Extract()

	th.AssertNoErr(t, err)

	t.Logf("Created Monitor, ID [%s]", m.ID)

	return m.ID
}

func deleteMonitor(t *testing.T, monitorID string) {
	res := monitors.Delete(base.Client, monitorID)
	th.AssertNoErr(t, res.Err)
	t.Logf("Deleted Monitor, ID %s", monitorID)
}

func listMonitors(t *testing.T) {
	err := monitors.List(base.Client, monitors.ListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		monitorsList, err := monitors.ExtractMonitors(page)
		if err != nil {
			t.Errorf("Failed to extract Monitors: %v", err)
			return false, err
		}

		for _, monitor := range monitorsList {
			t.Logf("Listing Monitors: ID [%s] Type [%s] HTTPMethod [%s] URLPath [%s]",
				monitor.ID, monitor.Type, monitor.HTTPMethod, monitor.URLPath)
		}

		return true, nil
	})

	th.AssertNoErr(t, err)
}

func getMonitor(t *testing.T, monitorID string) {
	monitor, err := monitors.Get(base.Client, monitorID).Extract()

	th.AssertNoErr(t, err)

	t.Logf("Getting Monitor, ID [%s]: ", monitor.ID)
}

func updateMonitor(t *testing.T, monitorID string) {
	_, err := monitors.Update(base.Client, monitorID, monitors.UpdateOpts{MaxRetries: 10}).Extract()

	th.AssertNoErr(t, err)

	t.Logf("Updated Monitor, ID [%s]", monitorID)
}
