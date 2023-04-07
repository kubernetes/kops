package routers

import (
	"encoding/json"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
)

// GatewayInfo represents the information of an external gateway for any
// particular network router.
type GatewayInfo struct {
	NetworkID        string            `json:"network_id,omitempty"`
	EnableSNAT       *bool             `json:"enable_snat,omitempty"`
	ExternalFixedIPs []ExternalFixedIP `json:"external_fixed_ips,omitempty"`
}

// ExternalFixedIP is the IP address and subnet ID of the external gateway of a
// router.
type ExternalFixedIP struct {
	IPAddress string `json:"ip_address,omitempty"`
	SubnetID  string `json:"subnet_id"`
}

// Route is a possible route in a router.
type Route struct {
	NextHop         string `json:"nexthop"`
	DestinationCIDR string `json:"destination"`
}

// Router represents a Neutron router. A router is a logical entity that
// forwards packets across internal subnets and NATs (network address
// translation) them on external networks through an appropriate gateway.
//
// A router has an interface for each subnet with which it is associated. By
// default, the IP address of such interface is the subnet's gateway IP. Also,
// whenever a router is associated with a subnet, a port for that router
// interface is added to the subnet's network.
type Router struct {
	// Status indicates whether or not a router is currently operational.
	Status string `json:"status"`

	// GateayInfo provides information on external gateway for the router.
	GatewayInfo GatewayInfo `json:"external_gateway_info"`

	// AdminStateUp is the administrative state of the router.
	AdminStateUp bool `json:"admin_state_up"`

	// Distributed is whether router is disitrubted or not.
	Distributed bool `json:"distributed"`

	// Name is the human readable name for the router. It does not have to be
	// unique.
	Name string `json:"name"`

	// Description for the router.
	Description string `json:"description"`

	// ID is the unique identifier for the router.
	ID string `json:"id"`

	// TenantID is the project owner of the router. Only admin users can
	// specify a project identifier other than its own.
	TenantID string `json:"tenant_id"`

	// ProjectID is the project owner of the router.
	ProjectID string `json:"project_id"`

	// Routes are a collection of static routes that the router will host.
	Routes []Route `json:"routes"`

	// Availability zone hints groups network nodes that run services like DHCP, L3, FW, and others.
	// Used to make network resources highly available.
	AvailabilityZoneHints []string `json:"availability_zone_hints"`

	// Tags optionally set via extensions/attributestags
	Tags []string `json:"tags"`
}

// RouterPage is the page returned by a pager when traversing over a
// collection of routers.
type RouterPage struct {
	pagination.LinkedPageBase
}

// NextPageURL is invoked when a paginated collection of routers has reached
// the end of a page and the pager seeks to traverse over a new one. In order
// to do this, it needs to construct the next page's URL.
func (r RouterPage) NextPageURL() (string, error) {
	var s struct {
		Links []gophercloud.Link `json:"routers_links"`
	}
	err := r.ExtractInto(&s)
	if err != nil {
		return "", err
	}
	return gophercloud.ExtractNextURL(s.Links)
}

// IsEmpty checks whether a RouterPage struct is empty.
func (r RouterPage) IsEmpty() (bool, error) {
	if r.StatusCode == 204 {
		return true, nil
	}

	is, err := ExtractRouters(r)
	return len(is) == 0, err
}

// ExtractRouters accepts a Page struct, specifically a RouterPage struct,
// and extracts the elements into a slice of Router structs. In other words,
// a generic collection is mapped into a relevant slice.
func ExtractRouters(r pagination.Page) ([]Router, error) {
	var s struct {
		Routers []Router `json:"routers"`
	}
	err := (r.(RouterPage)).ExtractInto(&s)
	return s.Routers, err
}

type commonResult struct {
	gophercloud.Result
}

// Extract is a function that accepts a result and extracts a router.
func (r commonResult) Extract() (*Router, error) {
	var s struct {
		Router *Router `json:"router"`
	}
	err := r.ExtractInto(&s)
	return s.Router, err
}

// CreateResult represents the result of a create operation. Call its Extract
// method to interpret it as a Router.
type CreateResult struct {
	commonResult
}

// GetResult represents the result of a get operation. Call its Extract
// method to interpret it as a Router.
type GetResult struct {
	commonResult
}

// UpdateResult represents the result of an update operation. Call its Extract
// method to interpret it as a Router.
type UpdateResult struct {
	commonResult
}

// DeleteResult represents the result of a delete operation. Call its ExtractErr
// method to determine if the request succeeded or failed.
type DeleteResult struct {
	gophercloud.ErrResult
}

// InterfaceInfo represents information about a particular router interface. As
// mentioned above, in order for a router to forward to a subnet, it needs an
// interface.
type InterfaceInfo struct {
	// SubnetID is the ID of the subnet which this interface is associated with.
	SubnetID string `json:"subnet_id"`

	// PortID is the ID of the port that is a part of the subnet.
	PortID string `json:"port_id"`

	// ID is the UUID of the interface.
	ID string `json:"id"`

	// TenantID is the owner of the interface.
	TenantID string `json:"tenant_id"`
}

// InterfaceResult represents the result of interface operations, such as
// AddInterface() and RemoveInterface(). Call its Extract method to interpret
// the result as a InterfaceInfo.
type InterfaceResult struct {
	gophercloud.Result
}

// Extract is a function that accepts a result and extracts an information struct.
func (r InterfaceResult) Extract() (*InterfaceInfo, error) {
	var s InterfaceInfo
	err := r.ExtractInto(&s)
	return &s, err
}

// L3Agent represents a Neutron agent for routers.
type L3Agent struct {
	// ID is the id of the agent.
	ID string `json:"id"`

	// AdminStateUp is an administrative state of the agent.
	AdminStateUp bool `json:"admin_state_up"`

	// AgentType is a type of the agent.
	AgentType string `json:"agent_type"`

	// Alive indicates whether agent is alive or not.
	Alive bool `json:"alive"`

	// ResourcesSynced indicates whether agent is synced or not.
	// Not all agent types track resources via Placement.
	ResourcesSynced bool `json:"resources_synced"`

	// AvailabilityZone is a zone of the agent.
	AvailabilityZone string `json:"availability_zone"`

	// Binary is an executable binary of the agent.
	Binary string `json:"binary"`

	// Configurations is a configuration specific key/value pairs that are
	// determined by the agent binary and type.
	Configurations map[string]interface{} `json:"configurations"`

	// CreatedAt is a creation timestamp.
	CreatedAt time.Time `json:"-"`

	// StartedAt is a starting timestamp.
	StartedAt time.Time `json:"-"`

	// HeartbeatTimestamp is a last heartbeat timestamp.
	HeartbeatTimestamp time.Time `json:"-"`

	// Description contains agent description.
	Description string `json:"description"`

	// Host is a hostname of the agent system.
	Host string `json:"host"`

	// Topic contains name of AMQP topic.
	Topic string `json:"topic"`

	// HAState is a ha state of agent(active/standby) for router
	HAState string `json:"ha_state"`

	// ResourceVersions is a list agent known objects and version numbers
	ResourceVersions map[string]interface{} `json:"resource_versions"`
}

// UnmarshalJSON helps to convert the timestamps into the time.Time type.
func (r *L3Agent) UnmarshalJSON(b []byte) error {
	type tmp L3Agent
	var s struct {
		tmp
		CreatedAt          gophercloud.JSONRFC3339ZNoTNoZ `json:"created_at"`
		StartedAt          gophercloud.JSONRFC3339ZNoTNoZ `json:"started_at"`
		HeartbeatTimestamp gophercloud.JSONRFC3339ZNoTNoZ `json:"heartbeat_timestamp"`
	}
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*r = L3Agent(s.tmp)

	r.CreatedAt = time.Time(s.CreatedAt)
	r.StartedAt = time.Time(s.StartedAt)
	r.HeartbeatTimestamp = time.Time(s.HeartbeatTimestamp)

	return nil
}

type ListL3AgentsPage struct {
	pagination.SinglePageBase
}

func (r ListL3AgentsPage) IsEmpty() (bool, error) {
	if r.StatusCode == 204 {
		return true, nil
	}

	v, err := ExtractL3Agents(r)
	return len(v) == 0, err
}

func ExtractL3Agents(r pagination.Page) ([]L3Agent, error) {
	var s struct {
		L3Agents []L3Agent `json:"agents"`
	}

	err := (r.(ListL3AgentsPage)).ExtractInto(&s)
	return s.L3Agents, err
}
