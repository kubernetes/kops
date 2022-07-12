// This file was automatically generated. DO NOT EDIT.
// If you have any remark or suggestion do not hesitate to open an issue.

// Package vpcgw provides methods and message types of the vpcgw v1 API.
package vpcgw

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/internal/marshaler"
	"github.com/scaleway/scaleway-sdk-go/internal/parameter"
	"github.com/scaleway/scaleway-sdk-go/namegenerator"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// always import dependencies
var (
	_ fmt.Stringer
	_ json.Unmarshaler
	_ url.URL
	_ net.IP
	_ http.Header
	_ bytes.Reader
	_ time.Time
	_ = strings.Join

	_ scw.ScalewayRequest
	_ marshaler.Duration
	_ scw.File
	_ = parameter.AddToQuery
	_ = namegenerator.GetRandomName
)

// API: vPC Public Gateway API
type API struct {
	client *scw.Client
}

// NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

type DHCPEntryType string

const (
	// DHCPEntryTypeUnknown is [insert doc].
	DHCPEntryTypeUnknown = DHCPEntryType("unknown")
	// DHCPEntryTypeReservation is [insert doc].
	DHCPEntryTypeReservation = DHCPEntryType("reservation")
	// DHCPEntryTypeLease is [insert doc].
	DHCPEntryTypeLease = DHCPEntryType("lease")
)

func (enum DHCPEntryType) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum DHCPEntryType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *DHCPEntryType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = DHCPEntryType(DHCPEntryType(tmp).String())
	return nil
}

type GatewayNetworkStatus string

const (
	// GatewayNetworkStatusUnknown is [insert doc].
	GatewayNetworkStatusUnknown = GatewayNetworkStatus("unknown")
	// GatewayNetworkStatusCreated is [insert doc].
	GatewayNetworkStatusCreated = GatewayNetworkStatus("created")
	// GatewayNetworkStatusAttaching is [insert doc].
	GatewayNetworkStatusAttaching = GatewayNetworkStatus("attaching")
	// GatewayNetworkStatusConfiguring is [insert doc].
	GatewayNetworkStatusConfiguring = GatewayNetworkStatus("configuring")
	// GatewayNetworkStatusReady is [insert doc].
	GatewayNetworkStatusReady = GatewayNetworkStatus("ready")
	// GatewayNetworkStatusDetaching is [insert doc].
	GatewayNetworkStatusDetaching = GatewayNetworkStatus("detaching")
	// GatewayNetworkStatusDeleted is [insert doc].
	GatewayNetworkStatusDeleted = GatewayNetworkStatus("deleted")
)

func (enum GatewayNetworkStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum GatewayNetworkStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *GatewayNetworkStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = GatewayNetworkStatus(GatewayNetworkStatus(tmp).String())
	return nil
}

type GatewayStatus string

const (
	// GatewayStatusUnknown is [insert doc].
	GatewayStatusUnknown = GatewayStatus("unknown")
	// GatewayStatusStopped is [insert doc].
	GatewayStatusStopped = GatewayStatus("stopped")
	// GatewayStatusAllocating is [insert doc].
	GatewayStatusAllocating = GatewayStatus("allocating")
	// GatewayStatusConfiguring is [insert doc].
	GatewayStatusConfiguring = GatewayStatus("configuring")
	// GatewayStatusRunning is [insert doc].
	GatewayStatusRunning = GatewayStatus("running")
	// GatewayStatusStopping is [insert doc].
	GatewayStatusStopping = GatewayStatus("stopping")
	// GatewayStatusFailed is [insert doc].
	GatewayStatusFailed = GatewayStatus("failed")
	// GatewayStatusDeleting is [insert doc].
	GatewayStatusDeleting = GatewayStatus("deleting")
	// GatewayStatusDeleted is [insert doc].
	GatewayStatusDeleted = GatewayStatus("deleted")
	// GatewayStatusLocked is [insert doc].
	GatewayStatusLocked = GatewayStatus("locked")
)

func (enum GatewayStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum GatewayStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *GatewayStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = GatewayStatus(GatewayStatus(tmp).String())
	return nil
}

type ListDHCPEntriesRequestOrderBy string

const (
	// ListDHCPEntriesRequestOrderByCreatedAtAsc is [insert doc].
	ListDHCPEntriesRequestOrderByCreatedAtAsc = ListDHCPEntriesRequestOrderBy("created_at_asc")
	// ListDHCPEntriesRequestOrderByCreatedAtDesc is [insert doc].
	ListDHCPEntriesRequestOrderByCreatedAtDesc = ListDHCPEntriesRequestOrderBy("created_at_desc")
	// ListDHCPEntriesRequestOrderByIPAddressAsc is [insert doc].
	ListDHCPEntriesRequestOrderByIPAddressAsc = ListDHCPEntriesRequestOrderBy("ip_address_asc")
	// ListDHCPEntriesRequestOrderByIPAddressDesc is [insert doc].
	ListDHCPEntriesRequestOrderByIPAddressDesc = ListDHCPEntriesRequestOrderBy("ip_address_desc")
	// ListDHCPEntriesRequestOrderByHostnameAsc is [insert doc].
	ListDHCPEntriesRequestOrderByHostnameAsc = ListDHCPEntriesRequestOrderBy("hostname_asc")
	// ListDHCPEntriesRequestOrderByHostnameDesc is [insert doc].
	ListDHCPEntriesRequestOrderByHostnameDesc = ListDHCPEntriesRequestOrderBy("hostname_desc")
)

func (enum ListDHCPEntriesRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListDHCPEntriesRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListDHCPEntriesRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListDHCPEntriesRequestOrderBy(ListDHCPEntriesRequestOrderBy(tmp).String())
	return nil
}

type ListDHCPsRequestOrderBy string

const (
	// ListDHCPsRequestOrderByCreatedAtAsc is [insert doc].
	ListDHCPsRequestOrderByCreatedAtAsc = ListDHCPsRequestOrderBy("created_at_asc")
	// ListDHCPsRequestOrderByCreatedAtDesc is [insert doc].
	ListDHCPsRequestOrderByCreatedAtDesc = ListDHCPsRequestOrderBy("created_at_desc")
	// ListDHCPsRequestOrderBySubnetAsc is [insert doc].
	ListDHCPsRequestOrderBySubnetAsc = ListDHCPsRequestOrderBy("subnet_asc")
	// ListDHCPsRequestOrderBySubnetDesc is [insert doc].
	ListDHCPsRequestOrderBySubnetDesc = ListDHCPsRequestOrderBy("subnet_desc")
)

func (enum ListDHCPsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListDHCPsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListDHCPsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListDHCPsRequestOrderBy(ListDHCPsRequestOrderBy(tmp).String())
	return nil
}

type ListGatewayNetworksRequestOrderBy string

const (
	// ListGatewayNetworksRequestOrderByCreatedAtAsc is [insert doc].
	ListGatewayNetworksRequestOrderByCreatedAtAsc = ListGatewayNetworksRequestOrderBy("created_at_asc")
	// ListGatewayNetworksRequestOrderByCreatedAtDesc is [insert doc].
	ListGatewayNetworksRequestOrderByCreatedAtDesc = ListGatewayNetworksRequestOrderBy("created_at_desc")
	// ListGatewayNetworksRequestOrderByStatusAsc is [insert doc].
	ListGatewayNetworksRequestOrderByStatusAsc = ListGatewayNetworksRequestOrderBy("status_asc")
	// ListGatewayNetworksRequestOrderByStatusDesc is [insert doc].
	ListGatewayNetworksRequestOrderByStatusDesc = ListGatewayNetworksRequestOrderBy("status_desc")
)

func (enum ListGatewayNetworksRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListGatewayNetworksRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListGatewayNetworksRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListGatewayNetworksRequestOrderBy(ListGatewayNetworksRequestOrderBy(tmp).String())
	return nil
}

type ListGatewaysRequestOrderBy string

const (
	// ListGatewaysRequestOrderByCreatedAtAsc is [insert doc].
	ListGatewaysRequestOrderByCreatedAtAsc = ListGatewaysRequestOrderBy("created_at_asc")
	// ListGatewaysRequestOrderByCreatedAtDesc is [insert doc].
	ListGatewaysRequestOrderByCreatedAtDesc = ListGatewaysRequestOrderBy("created_at_desc")
	// ListGatewaysRequestOrderByNameAsc is [insert doc].
	ListGatewaysRequestOrderByNameAsc = ListGatewaysRequestOrderBy("name_asc")
	// ListGatewaysRequestOrderByNameDesc is [insert doc].
	ListGatewaysRequestOrderByNameDesc = ListGatewaysRequestOrderBy("name_desc")
	// ListGatewaysRequestOrderByTypeAsc is [insert doc].
	ListGatewaysRequestOrderByTypeAsc = ListGatewaysRequestOrderBy("type_asc")
	// ListGatewaysRequestOrderByTypeDesc is [insert doc].
	ListGatewaysRequestOrderByTypeDesc = ListGatewaysRequestOrderBy("type_desc")
	// ListGatewaysRequestOrderByStatusAsc is [insert doc].
	ListGatewaysRequestOrderByStatusAsc = ListGatewaysRequestOrderBy("status_asc")
	// ListGatewaysRequestOrderByStatusDesc is [insert doc].
	ListGatewaysRequestOrderByStatusDesc = ListGatewaysRequestOrderBy("status_desc")
)

func (enum ListGatewaysRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListGatewaysRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListGatewaysRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListGatewaysRequestOrderBy(ListGatewaysRequestOrderBy(tmp).String())
	return nil
}

type ListIPsRequestOrderBy string

const (
	// ListIPsRequestOrderByCreatedAtAsc is [insert doc].
	ListIPsRequestOrderByCreatedAtAsc = ListIPsRequestOrderBy("created_at_asc")
	// ListIPsRequestOrderByCreatedAtDesc is [insert doc].
	ListIPsRequestOrderByCreatedAtDesc = ListIPsRequestOrderBy("created_at_desc")
	// ListIPsRequestOrderByIPAsc is [insert doc].
	ListIPsRequestOrderByIPAsc = ListIPsRequestOrderBy("ip_asc")
	// ListIPsRequestOrderByIPDesc is [insert doc].
	ListIPsRequestOrderByIPDesc = ListIPsRequestOrderBy("ip_desc")
	// ListIPsRequestOrderByReverseAsc is [insert doc].
	ListIPsRequestOrderByReverseAsc = ListIPsRequestOrderBy("reverse_asc")
	// ListIPsRequestOrderByReverseDesc is [insert doc].
	ListIPsRequestOrderByReverseDesc = ListIPsRequestOrderBy("reverse_desc")
)

func (enum ListIPsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListIPsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListIPsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListIPsRequestOrderBy(ListIPsRequestOrderBy(tmp).String())
	return nil
}

type ListPATRulesRequestOrderBy string

const (
	// ListPATRulesRequestOrderByCreatedAtAsc is [insert doc].
	ListPATRulesRequestOrderByCreatedAtAsc = ListPATRulesRequestOrderBy("created_at_asc")
	// ListPATRulesRequestOrderByCreatedAtDesc is [insert doc].
	ListPATRulesRequestOrderByCreatedAtDesc = ListPATRulesRequestOrderBy("created_at_desc")
	// ListPATRulesRequestOrderByPublicPortAsc is [insert doc].
	ListPATRulesRequestOrderByPublicPortAsc = ListPATRulesRequestOrderBy("public_port_asc")
	// ListPATRulesRequestOrderByPublicPortDesc is [insert doc].
	ListPATRulesRequestOrderByPublicPortDesc = ListPATRulesRequestOrderBy("public_port_desc")
)

func (enum ListPATRulesRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListPATRulesRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListPATRulesRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListPATRulesRequestOrderBy(ListPATRulesRequestOrderBy(tmp).String())
	return nil
}

type PATRuleProtocol string

const (
	// PATRuleProtocolUnknown is [insert doc].
	PATRuleProtocolUnknown = PATRuleProtocol("unknown")
	// PATRuleProtocolBoth is [insert doc].
	PATRuleProtocolBoth = PATRuleProtocol("both")
	// PATRuleProtocolTCP is [insert doc].
	PATRuleProtocolTCP = PATRuleProtocol("tcp")
	// PATRuleProtocolUDP is [insert doc].
	PATRuleProtocolUDP = PATRuleProtocol("udp")
)

func (enum PATRuleProtocol) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum PATRuleProtocol) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *PATRuleProtocol) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = PATRuleProtocol(PATRuleProtocol(tmp).String())
	return nil
}

// DHCP: dhcp
type DHCP struct {
	// ID: ID of the DHCP config
	ID string `json:"id"`
	// OrganizationID: owning organization
	OrganizationID string `json:"organization_id"`
	// ProjectID: owning project
	ProjectID string `json:"project_id"`
	// CreatedAt: configuration creation date
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: configuration last modification date
	UpdatedAt *time.Time `json:"updated_at"`
	// Subnet: subnet for the DHCP server
	Subnet scw.IPNet `json:"subnet"`
	// Address: address of the DHCP server
	//
	// Address of the DHCP server. This will be the gateway's address in the private network. It must be part of config's subnet.
	//
	Address net.IP `json:"address"`
	// PoolLow: low IP (included) of the dynamic address pool. Must be in the config's subnet
	PoolLow net.IP `json:"pool_low"`
	// PoolHigh: high IP (included) of the dynamic address pool. Must be in the config's subnet
	PoolHigh net.IP `json:"pool_high"`
	// EnableDynamic: whether to enable dynamic pooling of IPs
	//
	// Whether to enable dynamic pooling of IPs. By turning the dynamic pool off, only pre-existing DHCP reservations will be handed out.
	//
	EnableDynamic bool `json:"enable_dynamic"`
	// ValidLifetime: how long, in seconds, DHCP entries will be valid for
	ValidLifetime *scw.Duration `json:"valid_lifetime"`
	// RenewTimer: after how long a renew will be attempted
	//
	// After how long, in seconds, a renew will be attempted. Must be 30s lower than `rebind_timer`.
	//
	RenewTimer *scw.Duration `json:"renew_timer"`
	// RebindTimer: after how long a DHCP client will query for a new lease if previous renews fail
	//
	// After how long, in seconds, a DHCP client will query for a new lease if previous renews fail. Must be 30s lower than `valid_lifetime`.
	//
	RebindTimer *scw.Duration `json:"rebind_timer"`
	// PushDefaultRoute: whether the gateway should push a default route to DHCP clients or only hand out IPs
	PushDefaultRoute bool `json:"push_default_route"`
	// PushDNSServer: whether the gateway should push custom DNS servers to clients
	//
	// Whether the gateway should push custom DNS servers to clients. This allows for instance hostname -> IP resolution.
	//
	PushDNSServer bool `json:"push_dns_server"`
	// DNSServersOverride: override the DNS server list pushed to DHCP clients, instead of the gateway itself
	DNSServersOverride []string `json:"dns_servers_override"`
	// DNSSearch: add search paths to the pushed DNS configuration
	DNSSearch []string `json:"dns_search"`
	// DNSLocalName: tLD given to hostnames in the Private Networks
	//
	// TLD given to hostnames in the Private Network. If an instance with hostname `foo` gets a lease, and this is set to `bar`, `foo.bar` will resolve.
	//
	DNSLocalName string `json:"dns_local_name"`
	// Zone: zone this configuration is available in
	Zone scw.Zone `json:"zone"`
}

// DHCPEntry: dhcp entry
type DHCPEntry struct {
	// ID: entry ID
	ID string `json:"id"`
	// CreatedAt: configuration creation date
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: configuration last modification date
	UpdatedAt *time.Time `json:"updated_at"`
	// GatewayNetworkID: owning GatewayNetwork
	GatewayNetworkID string `json:"gateway_network_id"`
	// MacAddress: mAC address of the client machine
	MacAddress string `json:"mac_address"`
	// IPAddress: assigned IP address
	IPAddress net.IP `json:"ip_address"`
	// Hostname: hostname of the client machine
	Hostname string `json:"hostname"`
	// Type: entry type, either static (DHCP reservation) or dynamic (DHCP lease)
	//
	// Default value: unknown
	Type DHCPEntryType `json:"type"`
	// Zone: zone this entry is available in
	Zone scw.Zone `json:"zone"`
}

// Gateway: gateway
type Gateway struct {
	// ID: ID of the gateway
	ID string `json:"id"`
	// OrganizationID: owning organization
	OrganizationID string `json:"organization_id"`
	// ProjectID: owning project
	ProjectID string `json:"project_id"`
	// CreatedAt: gateway creation date
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: gateway last modification date
	UpdatedAt *time.Time `json:"updated_at"`
	// Type: gateway type
	Type *GatewayType `json:"type"`
	// Status: gateway's current status
	//
	// Default value: unknown
	Status GatewayStatus `json:"status"`
	// Name: name of the gateway
	Name string `json:"name"`
	// Tags: tags of the gateway
	Tags []string `json:"tags"`
	// IP: public IP of the gateway
	IP *IP `json:"ip"`
	// GatewayNetworks: gatewayNetworks attached to the gateway
	GatewayNetworks []*GatewayNetwork `json:"gateway_networks"`
	// UpstreamDNSServers: override the gateway's default recursive DNS servers
	UpstreamDNSServers []string `json:"upstream_dns_servers"`
	// Version: version of the running gateway software
	Version *string `json:"version"`
	// CanUpgradeTo: newly available gateway software version that can be updated to
	CanUpgradeTo *string `json:"can_upgrade_to"`
	// BastionEnabled: whether SSH bastion is enabled on the gateway
	BastionEnabled bool `json:"bastion_enabled"`
	// BastionPort: port of the SSH bastion
	BastionPort uint32 `json:"bastion_port"`
	// SMTPEnabled: whether SMTP traffic is allowed to pass through the gateway
	SMTPEnabled bool `json:"smtp_enabled"`
	// Zone: zone the gateway is available in
	Zone scw.Zone `json:"zone"`
}

// GatewayNetwork: gateway network
type GatewayNetwork struct {
	// ID: ID of the connection
	ID string `json:"id"`
	// CreatedAt: connection creation date
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: connection last modification date
	UpdatedAt *time.Time `json:"updated_at"`
	// GatewayID: ID of the connected gateway
	GatewayID string `json:"gateway_id"`
	// PrivateNetworkID: ID of the connected private network
	PrivateNetworkID string `json:"private_network_id"`
	// MacAddress: mAC address of the gateway in the network (if the gateway is up and running)
	MacAddress *string `json:"mac_address"`
	// EnableMasquerade: whether the gateway masquerades traffic for this network
	EnableMasquerade bool `json:"enable_masquerade"`
	// Status: current status of the gateway network connection
	//
	// Default value: unknown
	Status GatewayNetworkStatus `json:"status"`
	// DHCP: DHCP configuration for the connected private network
	DHCP *DHCP `json:"dhcp"`
	// EnableDHCP: whether DHCP is enabled on the connected Private Network
	EnableDHCP bool `json:"enable_dhcp"`
	// Address: address of the Gateway in CIDR form to use when DHCP is not used
	Address *scw.IPNet `json:"address"`
	// Zone: zone the connection lives in
	Zone scw.Zone `json:"zone"`
}

// GatewayType: gateway type
type GatewayType struct {
	// Name: type name
	Name string `json:"name"`
	// Bandwidth: bandwidth, in bps, the gateway has
	//
	// Bandwidth, in bps, the gateway has. This is the public bandwidth to the outer internet, and the internal bandwidth to each connected Private Networks.
	//
	Bandwidth uint64 `json:"bandwidth"`
	// Zone: zone the type is available in
	Zone scw.Zone `json:"zone"`
}

// IP: ip
type IP struct {
	// ID: IP ID
	ID string `json:"id"`
	// OrganizationID: owning organization
	OrganizationID string `json:"organization_id"`
	// ProjectID: owning project
	ProjectID string `json:"project_id"`
	// CreatedAt: configuration creation date
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: configuration last modification date
	UpdatedAt *time.Time `json:"updated_at"`
	// Tags: tags associated with the IP
	Tags []string `json:"tags"`
	// Address: the IP itself
	Address net.IP `json:"address"`
	// Reverse: reverse domain name for the IP address
	Reverse *string `json:"reverse"`
	// GatewayID: gateway associated to the IP
	GatewayID *string `json:"gateway_id"`
	// Zone: zone this IP is available in
	Zone scw.Zone `json:"zone"`
}

// ListDHCPEntriesResponse: list dhcp entries response
type ListDHCPEntriesResponse struct {
	// DHCPEntries: DHCP entries in this page
	DHCPEntries []*DHCPEntry `json:"dhcp_entries"`
	// TotalCount: total DHCP entries matching the filter
	TotalCount uint32 `json:"total_count"`
}

// ListDHCPsResponse: list dhc ps response
type ListDHCPsResponse struct {
	// Dhcps: first page of DHCP configs
	Dhcps []*DHCP `json:"dhcps"`
	// TotalCount: total DHCP configs matching the filter
	TotalCount uint32 `json:"total_count"`
}

// ListGatewayNetworksResponse: list gateway networks response
type ListGatewayNetworksResponse struct {
	// GatewayNetworks: gatewayNetworks in this page
	GatewayNetworks []*GatewayNetwork `json:"gateway_networks"`
	// TotalCount: total GatewayNetworks count matching the filter
	TotalCount uint32 `json:"total_count"`
}

// ListGatewayTypesResponse: list gateway types response
type ListGatewayTypesResponse struct {
	// Types: available types of gateway
	Types []*GatewayType `json:"types"`
}

// ListGatewaysResponse: list gateways response
type ListGatewaysResponse struct {
	// Gateways: gateways in this page
	Gateways []*Gateway `json:"gateways"`
	// TotalCount: total count of gateways matching the filter
	TotalCount uint32 `json:"total_count"`
}

// ListIPsResponse: list i ps response
type ListIPsResponse struct {
	// IPs: iPs in this page
	IPs []*IP `json:"ips"`
	// TotalCount: total IP count matching the filter
	TotalCount uint32 `json:"total_count"`
}

// ListPATRulesResponse: list pat rules response
type ListPATRulesResponse struct {
	// PatRules: this page of PAT rules matching the filter
	PatRules []*PATRule `json:"pat_rules"`
	// TotalCount: total PAT rules matching the filter
	TotalCount uint32 `json:"total_count"`
}

// PATRule: pat rule
type PATRule struct {
	// ID: rule ID
	ID string `json:"id"`
	// GatewayID: gateway the PAT rule applies to
	GatewayID string `json:"gateway_id"`
	// CreatedAt: rule creation date
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: rule last modification date
	UpdatedAt *time.Time `json:"updated_at"`
	// PublicPort: public port to listen on
	PublicPort uint32 `json:"public_port"`
	// PrivateIP: private IP to forward data to
	PrivateIP net.IP `json:"private_ip"`
	// PrivatePort: private port to translate to
	PrivatePort uint32 `json:"private_port"`
	// Protocol: protocol the rule applies to
	//
	// Default value: unknown
	Protocol PATRuleProtocol `json:"protocol"`
	// Zone: zone this rule is available in
	Zone scw.Zone `json:"zone"`
}

// SetDHCPEntriesRequestEntry: set dhcp entries request. entry
type SetDHCPEntriesRequestEntry struct {
	// MacAddress: mAC address to give a static entry to
	//
	// MAC address to give a static entry to. A matching entry will be upgraded to a reservation, and a matching reservation will be updated.
	//
	MacAddress string `json:"mac_address"`
	// IPAddress: IP address to give to the machine
	IPAddress net.IP `json:"ip_address"`
}

// SetDHCPEntriesResponse: set dhcp entries response
type SetDHCPEntriesResponse struct {
	// DHCPEntries: list of DHCP entries
	DHCPEntries []*DHCPEntry `json:"dhcp_entries"`
}

// SetPATRulesRequestRule: set pat rules request. rule
type SetPATRulesRequestRule struct {
	// PublicPort: public port to listen on
	//
	// Public port to listen on. Uniquely identifies the rule, and a matching rule will be updated with the new parameters.
	//
	PublicPort uint32 `json:"public_port"`
	// PrivateIP: private IP to forward data to
	PrivateIP net.IP `json:"private_ip"`
	// PrivatePort: private port to translate to
	PrivatePort uint32 `json:"private_port"`
	// Protocol: protocol the rule should apply to
	//
	// Default value: unknown
	Protocol PATRuleProtocol `json:"protocol"`
}

// SetPATRulesResponse: set pat rules response
type SetPATRulesResponse struct {
	// PatRules: list of PAT rules
	PatRules []*PATRule `json:"pat_rules"`
}

// Service API

type ListGatewaysRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// OrderBy: order in which to return results
	//
	// Default value: created_at_asc
	OrderBy ListGatewaysRequestOrderBy `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: gateways per page
	PageSize *uint32 `json:"-"`
	// OrganizationID: include only gateways in this organization
	OrganizationID *string `json:"-"`
	// ProjectID: include only gateways in this project
	ProjectID *string `json:"-"`
	// Name: filter gateways including this name
	Name *string `json:"-"`
	// Tags: filter gateways with these tags
	Tags []string `json:"-"`
	// Type: filter gateways of this type
	Type *string `json:"-"`
	// Status: filter gateways in this status (unknown for any)
	//
	// Default value: unknown
	Status GatewayStatus `json:"-"`
	// PrivateNetworkID: filter gateways attached to this private network
	PrivateNetworkID *string `json:"-"`
}

// ListGateways: list VPC Public Gateways
func (s *API) ListGateways(req *ListGatewaysRequest, opts ...scw.RequestOption) (*ListGatewaysResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "project_id", req.ProjectID)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "tags", req.Tags)
	parameter.AddToQuery(query, "type", req.Type)
	parameter.AddToQuery(query, "status", req.Status)
	parameter.AddToQuery(query, "private_network_id", req.PrivateNetworkID)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateways",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListGatewaysResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetGatewayRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// GatewayID: ID of the gateway to fetch
	GatewayID string `json:"-"`
}

// GetGateway: get a VPC Public Gateway
func (s *API) GetGateway(req *GetGatewayRequest, opts ...scw.RequestOption) (*Gateway, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.GatewayID) == "" {
		return nil, errors.New("field GatewayID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateways/" + fmt.Sprint(req.GatewayID) + "",
		Headers: http.Header{},
	}

	var resp Gateway

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateGatewayRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// ProjectID: project to create the gateway into
	ProjectID string `json:"project_id"`
	// Name: name of the gateway
	Name string `json:"name"`
	// Tags: tags for the gateway
	Tags []string `json:"tags"`
	// Type: gateway type
	Type string `json:"type"`
	// UpstreamDNSServers: override the gateway's default recursive DNS servers, if DNS features are enabled
	UpstreamDNSServers []string `json:"upstream_dns_servers"`
	// IPID: attach an existing IP to the gateway
	IPID *string `json:"ip_id"`
	// EnableSMTP: allow SMTP traffic to pass through the gateway
	EnableSMTP bool `json:"enable_smtp"`
	// EnableBastion: enable SSH bastion on the gateway
	EnableBastion bool `json:"enable_bastion"`
	// BastionPort: port of the SSH bastion
	BastionPort *uint32 `json:"bastion_port"`
}

// CreateGateway: create a VPC Public Gateway
func (s *API) CreateGateway(req *CreateGatewayRequest, opts ...scw.RequestOption) (*Gateway, error) {
	var err error

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("gw")
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateways",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Gateway

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateGatewayRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// GatewayID: ID of the gateway to update
	GatewayID string `json:"-"`
	// Name: name fo the gateway
	Name *string `json:"name"`
	// Tags: tags for the gateway
	Tags *[]string `json:"tags"`
	// UpstreamDNSServers: override the gateway's default recursive DNS servers, if DNS features are enabled
	UpstreamDNSServers *[]string `json:"upstream_dns_servers"`
	// EnableBastion: enable SSH bastion on the gateway
	EnableBastion *bool `json:"enable_bastion"`
	// BastionPort: port of the SSH bastion
	BastionPort *uint32 `json:"bastion_port"`
	// EnableSMTP: allow SMTP traffic to pass through the gateway
	EnableSMTP *bool `json:"enable_smtp"`
}

// UpdateGateway: update a VPC Public Gateway
func (s *API) UpdateGateway(req *UpdateGatewayRequest, opts ...scw.RequestOption) (*Gateway, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.GatewayID) == "" {
		return nil, errors.New("field GatewayID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateways/" + fmt.Sprint(req.GatewayID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Gateway

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteGatewayRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// GatewayID: ID of the gateway to delete
	GatewayID string `json:"-"`
	// CleanupDHCP: whether to cleanup attached DHCP configurations
	//
	// Whether to cleanup attached DHCP configurations (if any, and if not attached to another Gateway Network).
	//
	CleanupDHCP bool `json:"-"`
}

// DeleteGateway: delete a VPC Public Gateway
func (s *API) DeleteGateway(req *DeleteGatewayRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	query := url.Values{}
	parameter.AddToQuery(query, "cleanup_dhcp", req.CleanupDHCP)

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.GatewayID) == "" {
		return errors.New("field GatewayID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateways/" + fmt.Sprint(req.GatewayID) + "",
		Query:   query,
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type UpgradeGatewayRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// GatewayID: ID of the gateway to upgrade
	GatewayID string `json:"-"`
}

// UpgradeGateway: upgrade a VPC Public Gateway to the latest version
func (s *API) UpgradeGateway(req *UpgradeGatewayRequest, opts ...scw.RequestOption) (*Gateway, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.GatewayID) == "" {
		return nil, errors.New("field GatewayID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateways/" + fmt.Sprint(req.GatewayID) + "/upgrade",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Gateway

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListGatewayNetworksRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// OrderBy: order in which to return results
	//
	// Default value: created_at_asc
	OrderBy ListGatewayNetworksRequestOrderBy `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: gatewayNetworks per page
	PageSize *uint32 `json:"-"`
	// GatewayID: filter by gateway
	GatewayID *string `json:"-"`
	// PrivateNetworkID: filter by private network
	PrivateNetworkID *string `json:"-"`
	// EnableMasquerade: filter by masquerade enablement
	EnableMasquerade *bool `json:"-"`
	// DHCPID: filter by DHCP configuration
	DHCPID *string `json:"-"`
	// Status: filter GatewayNetworks by this status (unknown for any)
	//
	// Default value: unknown
	Status GatewayNetworkStatus `json:"-"`
}

// ListGatewayNetworks: list gateway connections to Private Networks
func (s *API) ListGatewayNetworks(req *ListGatewayNetworksRequest, opts ...scw.RequestOption) (*ListGatewayNetworksResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "gateway_id", req.GatewayID)
	parameter.AddToQuery(query, "private_network_id", req.PrivateNetworkID)
	parameter.AddToQuery(query, "enable_masquerade", req.EnableMasquerade)
	parameter.AddToQuery(query, "dhcp_id", req.DHCPID)
	parameter.AddToQuery(query, "status", req.Status)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateway-networks",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListGatewayNetworksResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetGatewayNetworkRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// GatewayNetworkID: ID of the GatewayNetwork to fetch
	GatewayNetworkID string `json:"-"`
}

// GetGatewayNetwork: get a gateway connection to a Private Network
func (s *API) GetGatewayNetwork(req *GetGatewayNetworkRequest, opts ...scw.RequestOption) (*GatewayNetwork, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.GatewayNetworkID) == "" {
		return nil, errors.New("field GatewayNetworkID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateway-networks/" + fmt.Sprint(req.GatewayNetworkID) + "",
		Headers: http.Header{},
	}

	var resp GatewayNetwork

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateGatewayNetworkRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// GatewayID: gateway to connect
	GatewayID string `json:"gateway_id"`
	// PrivateNetworkID: private Network to connect
	PrivateNetworkID string `json:"private_network_id"`
	// EnableMasquerade: whether to enable masquerade on this network
	EnableMasquerade bool `json:"enable_masquerade"`
	// DHCPID: existing configuration
	// Precisely one of Address, DHCPID must be set.
	DHCPID *string `json:"dhcp_id,omitempty"`
	// Address: static IP address in CIDR format to to use without DHCP
	// Precisely one of Address, DHCPID must be set.
	Address *scw.IPNet `json:"address,omitempty"`
	// EnableDHCP: whether to enable DHCP on this Private Network
	//
	// Whether to enable DHCP on this Private Network. Defaults to `true` if either `dhcp_id` or `dhcp` short: are present. If set to `true`, requires that either `dhcp_id` or `dhcp` to be present.
	//
	EnableDHCP *bool `json:"enable_dhcp"`
}

// CreateGatewayNetwork: attach a gateway to a Private Network
func (s *API) CreateGatewayNetwork(req *CreateGatewayNetworkRequest, opts ...scw.RequestOption) (*GatewayNetwork, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateway-networks",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp GatewayNetwork

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateGatewayNetworkRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// GatewayNetworkID: ID of the GatewayNetwork to update
	GatewayNetworkID string `json:"-"`
	// EnableMasquerade: new masquerade enablement
	EnableMasquerade *bool `json:"enable_masquerade"`
	// DHCPID: new DHCP configuration
	// Precisely one of Address, DHCPID must be set.
	DHCPID *string `json:"dhcp_id,omitempty"`
	// EnableDHCP: whether to enable DHCP on the connected Private Network
	EnableDHCP *bool `json:"enable_dhcp"`
	// Address: new static IP address
	// Precisely one of Address, DHCPID must be set.
	Address *scw.IPNet `json:"address,omitempty"`
}

// UpdateGatewayNetwork: update a gateway connection to a Private Network
func (s *API) UpdateGatewayNetwork(req *UpdateGatewayNetworkRequest, opts ...scw.RequestOption) (*GatewayNetwork, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.GatewayNetworkID) == "" {
		return nil, errors.New("field GatewayNetworkID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateway-networks/" + fmt.Sprint(req.GatewayNetworkID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp GatewayNetwork

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteGatewayNetworkRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// GatewayNetworkID: gatewayNetwork to delete
	GatewayNetworkID string `json:"-"`
	// CleanupDHCP: whether to cleanup the attached DHCP configuration
	//
	// Whether to cleanup the attached DHCP configuration (if any, and if not attached to another gateway_network).
	//
	CleanupDHCP bool `json:"-"`
}

// DeleteGatewayNetwork: detach a gateway from a Private Network
func (s *API) DeleteGatewayNetwork(req *DeleteGatewayNetworkRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	query := url.Values{}
	parameter.AddToQuery(query, "cleanup_dhcp", req.CleanupDHCP)

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.GatewayNetworkID) == "" {
		return errors.New("field GatewayNetworkID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateway-networks/" + fmt.Sprint(req.GatewayNetworkID) + "",
		Query:   query,
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ListDHCPsRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// OrderBy: order in which to return results
	//
	// Default value: created_at_asc
	OrderBy ListDHCPsRequestOrderBy `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: DHCP configurations per page
	PageSize *uint32 `json:"-"`
	// OrganizationID: include only DHCPs in this organization
	OrganizationID *string `json:"-"`
	// ProjectID: include only DHCPs in this project
	ProjectID *string `json:"-"`
	// Address: filter on gateway address
	Address *net.IP `json:"-"`
	// HasAddress: filter on subnets containing address
	HasAddress *net.IP `json:"-"`
}

// ListDHCPs: list DHCP configurations
func (s *API) ListDHCPs(req *ListDHCPsRequest, opts ...scw.RequestOption) (*ListDHCPsResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "project_id", req.ProjectID)
	parameter.AddToQuery(query, "address", req.Address)
	parameter.AddToQuery(query, "has_address", req.HasAddress)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/dhcps",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListDHCPsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetDHCPRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// DHCPID: ID of the DHCP config to fetch
	DHCPID string `json:"-"`
}

// GetDHCP: get a DHCP configuration
func (s *API) GetDHCP(req *GetDHCPRequest, opts ...scw.RequestOption) (*DHCP, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.DHCPID) == "" {
		return nil, errors.New("field DHCPID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/dhcps/" + fmt.Sprint(req.DHCPID) + "",
		Headers: http.Header{},
	}

	var resp DHCP

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateDHCPRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// ProjectID: project to create the DHCP configuration in
	ProjectID string `json:"project_id"`
	// Subnet: subnet for the DHCP server
	Subnet scw.IPNet `json:"subnet"`
	// Address: address of the DHCP server. This will be the gateway's address in the private network. Defaults to the first address of the subnet
	Address *net.IP `json:"address"`
	// PoolLow: low IP (included) of the dynamic address pool
	//
	// Low IP (included) of the dynamic address pool. Defaults to the second address of the subnet.
	PoolLow *net.IP `json:"pool_low"`
	// PoolHigh: high IP (included) of the dynamic address pool
	//
	// High IP (included) of the dynamic address pool. Defaults to the last address of the subnet.
	PoolHigh *net.IP `json:"pool_high"`
	// EnableDynamic: whether to enable dynamic pooling of IPs
	//
	// Whether to enable dynamic pooling of IPs. By turning the dynamic pool off, only pre-existing DHCP reservations will be handed out. Defaults to true.
	//
	EnableDynamic *bool `json:"enable_dynamic"`
	// ValidLifetime: for how long will DHCP entries will be valid
	//
	// For how long, in seconds, will DHCP entries will be valid. Defaults to 1h (3600s).
	ValidLifetime *scw.Duration `json:"valid_lifetime"`
	// RenewTimer: after how long a renew will be attempted
	//
	// After how long, in seconds, a renew will be attempted. Must be 30s lower than `rebind_timer`. Defaults to 50m (3000s).
	//
	RenewTimer *scw.Duration `json:"renew_timer"`
	// RebindTimer: after how long a DHCP client will query for a new lease if previous renews fail
	//
	// After how long, in seconds, a DHCP client will query for a new lease if previous renews fail. Must be 30s lower than `valid_lifetime`. Defaults to 51m (3060s).
	//
	RebindTimer *scw.Duration `json:"rebind_timer"`
	// PushDefaultRoute: whether the gateway should push a default route to DHCP clients or only hand out IPs. Defaults to true
	PushDefaultRoute *bool `json:"push_default_route"`
	// PushDNSServer: whether the gateway should push custom DNS servers to clients
	//
	// Whether the gateway should push custom DNS servers to clients. This allows for instance hostname -> IP resolution. Defaults to true.
	//
	PushDNSServer *bool `json:"push_dns_server"`
	// DNSServersOverride: override the DNS server list pushed to DHCP clients, instead of the gateway itself
	DNSServersOverride *[]string `json:"dns_servers_override"`
	// DNSSearch: additional DNS search paths
	DNSSearch *[]string `json:"dns_search"`
	// DNSLocalName: tLD given to hosts in the Private Network
	//
	// TLD given to hostnames in the Private Network. Allowed characters are `a-z0-9-.`. Defaults to the slugified Private Network name if created along a GatewayNetwork, or else to `priv`.
	//
	DNSLocalName *string `json:"dns_local_name"`
}

// CreateDHCP: create a DHCP configuration
func (s *API) CreateDHCP(req *CreateDHCPRequest, opts ...scw.RequestOption) (*DHCP, error) {
	var err error

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/dhcps",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp DHCP

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateDHCPRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// DHCPID: DHCP config to update
	DHCPID string `json:"-"`
	// Subnet: subnet for the DHCP server
	Subnet *scw.IPNet `json:"subnet"`
	// Address: address of the DHCP server. This will be the gateway's address in the private network
	Address *net.IP `json:"address"`
	// PoolLow: low IP (included) of the dynamic address pool
	PoolLow *net.IP `json:"pool_low"`
	// PoolHigh: high IP (included) of the dynamic address pool
	PoolHigh *net.IP `json:"pool_high"`
	// EnableDynamic: whether to enable dynamic pooling of IPs
	//
	// Whether to enable dynamic pooling of IPs. By turning the dynamic pool off, only pre-existing DHCP reservations will be handed out. Defaults to true.
	//
	EnableDynamic *bool `json:"enable_dynamic"`
	// ValidLifetime: how long, in seconds, DHCP entries will be valid for
	ValidLifetime *scw.Duration `json:"valid_lifetime"`
	// RenewTimer: after how long a renew will be attempted
	//
	// After how long, in seconds, a renew will be attempted. Must be 30s lower than `rebind_timer`.
	RenewTimer *scw.Duration `json:"renew_timer"`
	// RebindTimer: after how long a DHCP client will query for a new lease if previous renews fail
	//
	// After how long, in seconds, a DHCP client will query for a new lease if previous renews fail. Must be 30s lower than `valid_lifetime`.
	//
	RebindTimer *scw.Duration `json:"rebind_timer"`
	// PushDefaultRoute: whether the gateway should push a default route to DHCP clients or only hand out IPs
	PushDefaultRoute *bool `json:"push_default_route"`
	// PushDNSServer: whether the gateway should push custom DNS servers to clients
	//
	// Whether the gateway should push custom DNS servers to clients. This allows for instance hostname -> IP resolution.
	//
	PushDNSServer *bool `json:"push_dns_server"`
	// DNSServersOverride: override the DNS server list pushed to DHCP clients, instead of the gateway itself
	DNSServersOverride *[]string `json:"dns_servers_override"`
	// DNSSearch: additional DNS search paths
	DNSSearch *[]string `json:"dns_search"`
	// DNSLocalName: tLD given to hosts in the Private Network
	//
	// TLD given to hostnames in the Private Network. Allowed characters are `a-z0-9-.`.
	DNSLocalName *string `json:"dns_local_name"`
}

// UpdateDHCP: update a DHCP configuration
func (s *API) UpdateDHCP(req *UpdateDHCPRequest, opts ...scw.RequestOption) (*DHCP, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.DHCPID) == "" {
		return nil, errors.New("field DHCPID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/dhcps/" + fmt.Sprint(req.DHCPID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp DHCP

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteDHCPRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// DHCPID: DHCP config id to delete
	DHCPID string `json:"-"`
}

// DeleteDHCP: delete a DHCP configuration
func (s *API) DeleteDHCP(req *DeleteDHCPRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.DHCPID) == "" {
		return errors.New("field DHCPID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/dhcps/" + fmt.Sprint(req.DHCPID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ListDHCPEntriesRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// OrderBy: order in which to return results
	//
	// Default value: created_at_asc
	OrderBy ListDHCPEntriesRequestOrderBy `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: DHCP entries per page
	PageSize *uint32 `json:"-"`
	// GatewayNetworkID: filter entries based on the gateway network they are on
	GatewayNetworkID *string `json:"-"`
	// MacAddress: filter entries on their MAC address
	MacAddress *string `json:"-"`
	// IPAddress: filter entries on their IP address
	IPAddress *net.IP `json:"-"`
	// Hostname: filter entries on their hostname substring
	Hostname *string `json:"-"`
	// Type: filter entries on their type
	//
	// Default value: unknown
	Type DHCPEntryType `json:"-"`
}

// ListDHCPEntries: list DHCP entries
func (s *API) ListDHCPEntries(req *ListDHCPEntriesRequest, opts ...scw.RequestOption) (*ListDHCPEntriesResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "gateway_network_id", req.GatewayNetworkID)
	parameter.AddToQuery(query, "mac_address", req.MacAddress)
	parameter.AddToQuery(query, "ip_address", req.IPAddress)
	parameter.AddToQuery(query, "hostname", req.Hostname)
	parameter.AddToQuery(query, "type", req.Type)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/dhcp-entries",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListDHCPEntriesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetDHCPEntryRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// DHCPEntryID: ID of the DHCP entry to fetch
	DHCPEntryID string `json:"-"`
}

// GetDHCPEntry: get DHCP entries
func (s *API) GetDHCPEntry(req *GetDHCPEntryRequest, opts ...scw.RequestOption) (*DHCPEntry, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.DHCPEntryID) == "" {
		return nil, errors.New("field DHCPEntryID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/dhcp-entries/" + fmt.Sprint(req.DHCPEntryID) + "",
		Headers: http.Header{},
	}

	var resp DHCPEntry

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateDHCPEntryRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// GatewayNetworkID: gatewayNetwork on which to create a DHCP reservation
	GatewayNetworkID string `json:"gateway_network_id"`
	// MacAddress: mAC address to give a static entry to
	MacAddress string `json:"mac_address"`
	// IPAddress: IP address to give to the machine
	IPAddress net.IP `json:"ip_address"`
}

// CreateDHCPEntry: create a static DHCP reservation
func (s *API) CreateDHCPEntry(req *CreateDHCPEntryRequest, opts ...scw.RequestOption) (*DHCPEntry, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/dhcp-entries",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp DHCPEntry

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateDHCPEntryRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// DHCPEntryID: DHCP entry ID to update
	DHCPEntryID string `json:"-"`
	// IPAddress: new IP address to give to the machine
	IPAddress *net.IP `json:"ip_address"`
}

// UpdateDHCPEntry: update a DHCP entry
func (s *API) UpdateDHCPEntry(req *UpdateDHCPEntryRequest, opts ...scw.RequestOption) (*DHCPEntry, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.DHCPEntryID) == "" {
		return nil, errors.New("field DHCPEntryID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/dhcp-entries/" + fmt.Sprint(req.DHCPEntryID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp DHCPEntry

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type SetDHCPEntriesRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// GatewayNetworkID: gateway Network on which to set DHCP reservation list
	GatewayNetworkID string `json:"gateway_network_id"`
	// DHCPEntries: new list of DHCP reservations
	DHCPEntries []*SetDHCPEntriesRequestEntry `json:"dhcp_entries"`
}

// SetDHCPEntries: set all DHCP reservations on a Gateway Network
//
// Set the list of DHCP reservations attached to a Gateway Network. Reservations are identified by their MAC address, and will sync the current DHCP entry list to the given list, creating, updating or deleting DHCP entries.
//
func (s *API) SetDHCPEntries(req *SetDHCPEntriesRequest, opts ...scw.RequestOption) (*SetDHCPEntriesResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/dhcp-entries",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp SetDHCPEntriesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteDHCPEntryRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// DHCPEntryID: DHCP entry ID to delete
	DHCPEntryID string `json:"-"`
}

// DeleteDHCPEntry: delete a DHCP reservation
func (s *API) DeleteDHCPEntry(req *DeleteDHCPEntryRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.DHCPEntryID) == "" {
		return errors.New("field DHCPEntryID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/dhcp-entries/" + fmt.Sprint(req.DHCPEntryID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ListPATRulesRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// OrderBy: order in which to return results
	//
	// Default value: created_at_asc
	OrderBy ListPATRulesRequestOrderBy `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: pAT rules per page
	PageSize *uint32 `json:"-"`
	// GatewayID: fetch rules for this gateway
	GatewayID *string `json:"-"`
	// PrivateIP: fetch rules targeting this private ip
	PrivateIP *net.IP `json:"-"`
	// Protocol: fetch rules for this protocol
	//
	// Default value: unknown
	Protocol PATRuleProtocol `json:"-"`
}

// ListPATRules: list PAT rules
func (s *API) ListPATRules(req *ListPATRulesRequest, opts ...scw.RequestOption) (*ListPATRulesResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "gateway_id", req.GatewayID)
	parameter.AddToQuery(query, "private_ip", req.PrivateIP)
	parameter.AddToQuery(query, "protocol", req.Protocol)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/pat-rules",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListPATRulesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetPATRuleRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// PatRuleID: pAT rule to get
	PatRuleID string `json:"-"`
}

// GetPATRule: get a PAT rule
func (s *API) GetPATRule(req *GetPATRuleRequest, opts ...scw.RequestOption) (*PATRule, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.PatRuleID) == "" {
		return nil, errors.New("field PatRuleID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/pat-rules/" + fmt.Sprint(req.PatRuleID) + "",
		Headers: http.Header{},
	}

	var resp PATRule

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreatePATRuleRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// GatewayID: gateway on which to attach the rule to
	GatewayID string `json:"gateway_id"`
	// PublicPort: public port to listen on
	PublicPort uint32 `json:"public_port"`
	// PrivateIP: private IP to forward data to
	PrivateIP net.IP `json:"private_ip"`
	// PrivatePort: private port to translate to
	PrivatePort uint32 `json:"private_port"`
	// Protocol: protocol the rule should apply to
	//
	// Default value: unknown
	Protocol PATRuleProtocol `json:"protocol"`
}

// CreatePATRule: create a PAT rule
func (s *API) CreatePATRule(req *CreatePATRuleRequest, opts ...scw.RequestOption) (*PATRule, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/pat-rules",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp PATRule

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdatePATRuleRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// PatRuleID: pAT rule to update
	PatRuleID string `json:"-"`
	// PublicPort: public port to listen on
	PublicPort *uint32 `json:"public_port"`
	// PrivateIP: private IP to forward data to
	PrivateIP *net.IP `json:"private_ip"`
	// PrivatePort: private port to translate to
	PrivatePort *uint32 `json:"private_port"`
	// Protocol: protocol the rule should apply to
	//
	// Default value: unknown
	Protocol PATRuleProtocol `json:"protocol"`
}

// UpdatePATRule: update a PAT rule
func (s *API) UpdatePATRule(req *UpdatePATRuleRequest, opts ...scw.RequestOption) (*PATRule, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.PatRuleID) == "" {
		return nil, errors.New("field PatRuleID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/pat-rules/" + fmt.Sprint(req.PatRuleID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp PATRule

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type SetPATRulesRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// GatewayID: gateway on which to set the PAT rules
	GatewayID string `json:"gateway_id"`
	// PatRules: new list of PAT rules
	PatRules []*SetPATRulesRequestRule `json:"pat_rules"`
}

// SetPATRules: set all PAT rules on a Gateway
//
// Set the list of PAT rules attached to a Gateway. Rules are identified by their public port and protocol. This will sync the current PAT rule list with the givent list, creating, updating or deleting PAT rules.
//
func (s *API) SetPATRules(req *SetPATRulesRequest, opts ...scw.RequestOption) (*SetPATRulesResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/pat-rules",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp SetPATRulesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeletePATRuleRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// PatRuleID: pAT rule to delete
	PatRuleID string `json:"-"`
}

// DeletePATRule: delete a PAT rule
func (s *API) DeletePATRule(req *DeletePATRuleRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.PatRuleID) == "" {
		return errors.New("field PatRuleID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/pat-rules/" + fmt.Sprint(req.PatRuleID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ListGatewayTypesRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
}

// ListGatewayTypes: list VPC Public Gateway types
func (s *API) ListGatewayTypes(req *ListGatewayTypesRequest, opts ...scw.RequestOption) (*ListGatewayTypesResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateway-types",
		Headers: http.Header{},
	}

	var resp ListGatewayTypesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListIPsRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// OrderBy: order in which to return results
	//
	// Default value: created_at_asc
	OrderBy ListIPsRequestOrderBy `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: iPs per page
	PageSize *uint32 `json:"-"`
	// OrganizationID: include only IPs in this organization
	OrganizationID *string `json:"-"`
	// ProjectID: include only IPs in this project
	ProjectID *string `json:"-"`
	// Tags: filter IPs with these tags
	Tags []string `json:"-"`
	// Reverse: filter by reverse containing this string
	Reverse *string `json:"-"`
	// IsFree: filter whether the IP is attached to a gateway or not
	IsFree *bool `json:"-"`
}

// ListIPs: list IPs
func (s *API) ListIPs(req *ListIPsRequest, opts ...scw.RequestOption) (*ListIPsResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "project_id", req.ProjectID)
	parameter.AddToQuery(query, "tags", req.Tags)
	parameter.AddToQuery(query, "reverse", req.Reverse)
	parameter.AddToQuery(query, "is_free", req.IsFree)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/ips",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListIPsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetIPRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// IPID: ID of the IP to get
	IPID string `json:"-"`
}

// GetIP: get an IP
func (s *API) GetIP(req *GetIPRequest, opts ...scw.RequestOption) (*IP, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.IPID) == "" {
		return nil, errors.New("field IPID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/ips/" + fmt.Sprint(req.IPID) + "",
		Headers: http.Header{},
	}

	var resp IP

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateIPRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// ProjectID: project to create the IP into
	ProjectID string `json:"project_id"`
	// Tags: tags to give to the IP
	Tags []string `json:"tags"`
}

// CreateIP: reserve an IP
func (s *API) CreateIP(req *CreateIPRequest, opts ...scw.RequestOption) (*IP, error) {
	var err error

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/ips",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp IP

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateIPRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// IPID: ID of the IP to update
	IPID string `json:"-"`
	// Tags: tags to give to the IP
	Tags *[]string `json:"tags"`
	// Reverse: reverse to set on the IP. Empty string to unset
	Reverse *string `json:"reverse"`
	// GatewayID: gateway to attach the IP to. Empty string to detach
	GatewayID *string `json:"gateway_id"`
}

// UpdateIP: update an IP
func (s *API) UpdateIP(req *UpdateIPRequest, opts ...scw.RequestOption) (*IP, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.IPID) == "" {
		return nil, errors.New("field IPID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/ips/" + fmt.Sprint(req.IPID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp IP

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteIPRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// IPID: ID of the IP to delete
	IPID string `json:"-"`
}

// DeleteIP: delete an IP
func (s *API) DeleteIP(req *DeleteIPRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.IPID) == "" {
		return errors.New("field IPID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/ips/" + fmt.Sprint(req.IPID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type RefreshSSHKeysRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`

	GatewayID string `json:"-"`
}

func (s *API) RefreshSSHKeys(req *RefreshSSHKeysRequest, opts ...scw.RequestOption) (*Gateway, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.GatewayID) == "" {
		return nil, errors.New("field GatewayID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateways/" + fmt.Sprint(req.GatewayID) + "/refresh-ssh-keys",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Gateway

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListGatewaysResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListGatewaysResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListGatewaysResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Gateways = append(r.Gateways, results.Gateways...)
	r.TotalCount += uint32(len(results.Gateways))
	return uint32(len(results.Gateways)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListGatewayNetworksResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListGatewayNetworksResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListGatewayNetworksResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.GatewayNetworks = append(r.GatewayNetworks, results.GatewayNetworks...)
	r.TotalCount += uint32(len(results.GatewayNetworks))
	return uint32(len(results.GatewayNetworks)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListDHCPsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListDHCPsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListDHCPsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Dhcps = append(r.Dhcps, results.Dhcps...)
	r.TotalCount += uint32(len(results.Dhcps))
	return uint32(len(results.Dhcps)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListDHCPEntriesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListDHCPEntriesResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListDHCPEntriesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.DHCPEntries = append(r.DHCPEntries, results.DHCPEntries...)
	r.TotalCount += uint32(len(results.DHCPEntries))
	return uint32(len(results.DHCPEntries)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListPATRulesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListPATRulesResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListPATRulesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.PatRules = append(r.PatRules, results.PatRules...)
	r.TotalCount += uint32(len(results.PatRules))
	return uint32(len(results.PatRules)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListIPsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListIPsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListIPsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.IPs = append(r.IPs, results.IPs...)
	r.TotalCount += uint32(len(results.IPs))
	return uint32(len(results.IPs)), nil
}
