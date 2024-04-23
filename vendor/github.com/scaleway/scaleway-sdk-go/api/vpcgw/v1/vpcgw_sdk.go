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

type DHCPEntryType string

const (
	DHCPEntryTypeUnknown     = DHCPEntryType("unknown")
	DHCPEntryTypeReservation = DHCPEntryType("reservation")
	DHCPEntryTypeLease       = DHCPEntryType("lease")
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
	GatewayNetworkStatusUnknown     = GatewayNetworkStatus("unknown")
	GatewayNetworkStatusCreated     = GatewayNetworkStatus("created")
	GatewayNetworkStatusAttaching   = GatewayNetworkStatus("attaching")
	GatewayNetworkStatusConfiguring = GatewayNetworkStatus("configuring")
	GatewayNetworkStatusReady       = GatewayNetworkStatus("ready")
	GatewayNetworkStatusDetaching   = GatewayNetworkStatus("detaching")
	GatewayNetworkStatusDeleted     = GatewayNetworkStatus("deleted")
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
	GatewayStatusUnknown     = GatewayStatus("unknown")
	GatewayStatusStopped     = GatewayStatus("stopped")
	GatewayStatusAllocating  = GatewayStatus("allocating")
	GatewayStatusConfiguring = GatewayStatus("configuring")
	GatewayStatusRunning     = GatewayStatus("running")
	GatewayStatusStopping    = GatewayStatus("stopping")
	GatewayStatusFailed      = GatewayStatus("failed")
	GatewayStatusDeleting    = GatewayStatus("deleting")
	GatewayStatusDeleted     = GatewayStatus("deleted")
	GatewayStatusLocked      = GatewayStatus("locked")
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
	ListDHCPEntriesRequestOrderByCreatedAtAsc  = ListDHCPEntriesRequestOrderBy("created_at_asc")
	ListDHCPEntriesRequestOrderByCreatedAtDesc = ListDHCPEntriesRequestOrderBy("created_at_desc")
	ListDHCPEntriesRequestOrderByIPAddressAsc  = ListDHCPEntriesRequestOrderBy("ip_address_asc")
	ListDHCPEntriesRequestOrderByIPAddressDesc = ListDHCPEntriesRequestOrderBy("ip_address_desc")
	ListDHCPEntriesRequestOrderByHostnameAsc   = ListDHCPEntriesRequestOrderBy("hostname_asc")
	ListDHCPEntriesRequestOrderByHostnameDesc  = ListDHCPEntriesRequestOrderBy("hostname_desc")
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
	ListDHCPsRequestOrderByCreatedAtAsc  = ListDHCPsRequestOrderBy("created_at_asc")
	ListDHCPsRequestOrderByCreatedAtDesc = ListDHCPsRequestOrderBy("created_at_desc")
	ListDHCPsRequestOrderBySubnetAsc     = ListDHCPsRequestOrderBy("subnet_asc")
	ListDHCPsRequestOrderBySubnetDesc    = ListDHCPsRequestOrderBy("subnet_desc")
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
	ListGatewayNetworksRequestOrderByCreatedAtAsc  = ListGatewayNetworksRequestOrderBy("created_at_asc")
	ListGatewayNetworksRequestOrderByCreatedAtDesc = ListGatewayNetworksRequestOrderBy("created_at_desc")
	ListGatewayNetworksRequestOrderByStatusAsc     = ListGatewayNetworksRequestOrderBy("status_asc")
	ListGatewayNetworksRequestOrderByStatusDesc    = ListGatewayNetworksRequestOrderBy("status_desc")
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
	ListGatewaysRequestOrderByCreatedAtAsc  = ListGatewaysRequestOrderBy("created_at_asc")
	ListGatewaysRequestOrderByCreatedAtDesc = ListGatewaysRequestOrderBy("created_at_desc")
	ListGatewaysRequestOrderByNameAsc       = ListGatewaysRequestOrderBy("name_asc")
	ListGatewaysRequestOrderByNameDesc      = ListGatewaysRequestOrderBy("name_desc")
	ListGatewaysRequestOrderByTypeAsc       = ListGatewaysRequestOrderBy("type_asc")
	ListGatewaysRequestOrderByTypeDesc      = ListGatewaysRequestOrderBy("type_desc")
	ListGatewaysRequestOrderByStatusAsc     = ListGatewaysRequestOrderBy("status_asc")
	ListGatewaysRequestOrderByStatusDesc    = ListGatewaysRequestOrderBy("status_desc")
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
	ListIPsRequestOrderByCreatedAtAsc  = ListIPsRequestOrderBy("created_at_asc")
	ListIPsRequestOrderByCreatedAtDesc = ListIPsRequestOrderBy("created_at_desc")
	ListIPsRequestOrderByIPAsc         = ListIPsRequestOrderBy("ip_asc")
	ListIPsRequestOrderByIPDesc        = ListIPsRequestOrderBy("ip_desc")
	ListIPsRequestOrderByReverseAsc    = ListIPsRequestOrderBy("reverse_asc")
	ListIPsRequestOrderByReverseDesc   = ListIPsRequestOrderBy("reverse_desc")
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
	ListPATRulesRequestOrderByCreatedAtAsc   = ListPATRulesRequestOrderBy("created_at_asc")
	ListPATRulesRequestOrderByCreatedAtDesc  = ListPATRulesRequestOrderBy("created_at_desc")
	ListPATRulesRequestOrderByPublicPortAsc  = ListPATRulesRequestOrderBy("public_port_asc")
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
	PATRuleProtocolUnknown = PATRuleProtocol("unknown")
	PATRuleProtocolBoth    = PATRuleProtocol("both")
	PATRuleProtocolTCP     = PATRuleProtocol("tcp")
	PATRuleProtocolUDP     = PATRuleProtocol("udp")
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

// DHCP: dhcp.
type DHCP struct {
	// ID: ID of the DHCP config.
	ID string `json:"id"`

	// OrganizationID: owning Organization.
	OrganizationID string `json:"organization_id"`

	// ProjectID: owning Project.
	ProjectID string `json:"project_id"`

	// CreatedAt: date the DHCP configuration was created.
	CreatedAt *time.Time `json:"created_at"`

	// UpdatedAt: configuration last modification date.
	UpdatedAt *time.Time `json:"updated_at"`

	// Subnet: subnet for the DHCP server.
	Subnet scw.IPNet `json:"subnet"`

	// Address: IP address of the DHCP server. This will be the Public Gateway's address in the Private Network. It must be part of config's subnet.
	Address net.IP `json:"address"`

	// PoolLow: low IP (inclusive) of the dynamic address pool. Must be in the config's subnet.
	PoolLow net.IP `json:"pool_low"`

	// PoolHigh: high IP (inclusive) of the dynamic address pool. Must be in the config's subnet.
	PoolHigh net.IP `json:"pool_high"`

	// EnableDynamic: defines whether to enable dynamic pooling of IPs. When false, only pre-existing DHCP reservations will be handed out.
	EnableDynamic bool `json:"enable_dynamic"`

	// ValidLifetime: how long DHCP entries will be valid for.
	ValidLifetime *scw.Duration `json:"valid_lifetime"`

	// RenewTimer: after how long a renew will be attempted. Must be 30s lower than `rebind_timer`.
	RenewTimer *scw.Duration `json:"renew_timer"`

	// RebindTimer: after how long a DHCP client will query for a new lease if previous renews fail. Must be 30s lower than `valid_lifetime`.
	RebindTimer *scw.Duration `json:"rebind_timer"`

	// PushDefaultRoute: defines whether the gateway should push a default route to DHCP clients, or only hand out IPs.
	PushDefaultRoute bool `json:"push_default_route"`

	// PushDNSServer: defines whether the gateway should push custom DNS servers to clients. This allows for instance hostname -> IP resolution.
	PushDNSServer bool `json:"push_dns_server"`

	// DNSServersOverride: array of DNS server IP addresses used to override the DNS server list pushed to DHCP clients, instead of the gateway itself.
	DNSServersOverride []string `json:"dns_servers_override"`

	// DNSSearch: array of search paths in addition to the pushed DNS configuration.
	DNSSearch []string `json:"dns_search"`

	// DNSLocalName: tLD given to hostnames in the Private Networks. If an Instance with hostname `foo` gets a lease, and this is set to `bar`, `foo.bar` will resolve.
	DNSLocalName string `json:"dns_local_name"`

	// Zone: zone of this DHCP configuration.
	Zone scw.Zone `json:"zone"`
}

// IpamConfig: ipam config.
type IpamConfig struct {
	// PushDefaultRoute: defines whether the default route is enabled on that Gateway Network.
	PushDefaultRoute bool `json:"push_default_route"`

	// IpamIPID: iPAM-booked IP ID as the Gateway's IP in this Private Network.
	IpamIPID string `json:"ipam_ip_id"`
}

// GatewayNetwork: gateway network.
type GatewayNetwork struct {
	// ID: ID of the Public Gateway-Private Network connection.
	ID string `json:"id"`

	// CreatedAt: connection creation date.
	CreatedAt *time.Time `json:"created_at"`

	// UpdatedAt: connection last modification date.
	UpdatedAt *time.Time `json:"updated_at"`

	// GatewayID: ID of the connected Public Gateway.
	GatewayID string `json:"gateway_id"`

	// PrivateNetworkID: ID of the connected Private Network.
	PrivateNetworkID string `json:"private_network_id"`

	// MacAddress: mAC address of the gateway in the Private Network (if the gateway is up and running).
	MacAddress *string `json:"mac_address"`

	// EnableMasquerade: defines whether the gateway masquerades traffic for this Private Network (Dynamic NAT).
	EnableMasquerade bool `json:"enable_masquerade"`

	// Status: current status of the Public Gateway's connection to the Private Network.
	// Default value: unknown
	Status GatewayNetworkStatus `json:"status"`

	// DHCP: DHCP configuration for the connected Private Network.
	DHCP *DHCP `json:"dhcp"`

	// EnableDHCP: defines whether DHCP is enabled on the connected Private Network.
	EnableDHCP bool `json:"enable_dhcp"`

	// Address: address of the Gateway (in CIDR form) to use when DHCP is not used.
	Address *scw.IPNet `json:"address"`

	// IpamConfig: iPAM IP configuration used.
	IpamConfig *IpamConfig `json:"ipam_config"`

	// Zone: zone of the GatewayNetwork connection.
	Zone scw.Zone `json:"zone"`
}

// GatewayType: gateway type.
type GatewayType struct {
	// Name: public Gateway type name.
	Name string `json:"name"`

	// Bandwidth: bandwidth, in bps, of the Public Gateway. This is the public bandwidth to the outer Internet, and the internal bandwidth to each connected Private Networks.
	Bandwidth uint64 `json:"bandwidth"`

	// Zone: zone the Public Gateway type is available in.
	Zone scw.Zone `json:"zone"`
}

// IP: ip.
type IP struct {
	// ID: IP address ID.
	ID string `json:"id"`

	// OrganizationID: owning Organization.
	OrganizationID string `json:"organization_id"`

	// ProjectID: owning Project.
	ProjectID string `json:"project_id"`

	// CreatedAt: IP address creation date.
	CreatedAt *time.Time `json:"created_at"`

	// UpdatedAt: IP address last modification date.
	UpdatedAt *time.Time `json:"updated_at"`

	// Tags: tags associated with the IP address.
	Tags []string `json:"tags"`

	// Address: the IP address itself.
	Address net.IP `json:"address"`

	// Reverse: reverse domain name for the IP address.
	Reverse *string `json:"reverse"`

	// GatewayID: public Gateway associated with the IP address.
	GatewayID *string `json:"gateway_id"`

	// Zone: zone of the IP address.
	Zone scw.Zone `json:"zone"`
}

// CreateDHCPRequest: create dhcp request.
type CreateDHCPRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// ProjectID: project to create the DHCP configuration in.
	ProjectID string `json:"project_id"`

	// Subnet: subnet for the DHCP server.
	Subnet scw.IPNet `json:"subnet"`

	// Address: IP address of the DHCP server. This will be the gateway's address in the Private Network. Defaults to the first address of the subnet.
	Address *net.IP `json:"address,omitempty"`

	// PoolLow: low IP (inclusive) of the dynamic address pool. Must be in the config's subnet. Defaults to the second address of the subnet.
	PoolLow *net.IP `json:"pool_low,omitempty"`

	// PoolHigh: high IP (inclusive) of the dynamic address pool. Must be in the config's subnet. Defaults to the last address of the subnet.
	PoolHigh *net.IP `json:"pool_high,omitempty"`

	// EnableDynamic: defines whether to enable dynamic pooling of IPs. When false, only pre-existing DHCP reservations will be handed out. Defaults to true.
	EnableDynamic *bool `json:"enable_dynamic,omitempty"`

	// ValidLifetime: how long DHCP entries will be valid for. Defaults to 1h (3600s).
	ValidLifetime *scw.Duration `json:"valid_lifetime,omitempty"`

	// RenewTimer: after how long a renew will be attempted. Must be 30s lower than `rebind_timer`. Defaults to 50m (3000s).
	RenewTimer *scw.Duration `json:"renew_timer,omitempty"`

	// RebindTimer: after how long a DHCP client will query for a new lease if previous renews fail. Must be 30s lower than `valid_lifetime`. Defaults to 51m (3060s).
	RebindTimer *scw.Duration `json:"rebind_timer,omitempty"`

	// PushDefaultRoute: defines whether the gateway should push a default route to DHCP clients or only hand out IPs. Defaults to true.
	PushDefaultRoute *bool `json:"push_default_route,omitempty"`

	// PushDNSServer: defines whether the gateway should push custom DNS servers to clients. This allows for Instance hostname -> IP resolution. Defaults to true.
	PushDNSServer *bool `json:"push_dns_server,omitempty"`

	// DNSServersOverride: array of DNS server IP addresses used to override the DNS server list pushed to DHCP clients, instead of the gateway itself.
	DNSServersOverride *[]string `json:"dns_servers_override,omitempty"`

	// DNSSearch: array of search paths in addition to the pushed DNS configuration.
	DNSSearch *[]string `json:"dns_search,omitempty"`

	// DNSLocalName: tLD given to hostnames in the Private Network. Allowed characters are `a-z0-9-.`. Defaults to the slugified Private Network name if created along a GatewayNetwork, or else to `priv`.
	DNSLocalName *string `json:"dns_local_name,omitempty"`
}

// CreateGatewayNetworkRequestIpamConfig: create gateway network request ipam config.
type CreateGatewayNetworkRequestIpamConfig struct {
	// PushDefaultRoute: enabling the default route also enables masquerading.
	PushDefaultRoute bool `json:"push_default_route"`

	// IpamIPID: use this IPAM-booked IP ID as the Gateway's IP in this Private Network.
	IpamIPID *string `json:"ipam_ip_id"`
}

// DHCPEntry: dhcp entry.
type DHCPEntry struct {
	// ID: DHCP entry ID.
	ID string `json:"id"`

	// CreatedAt: DHCP entry creation date.
	CreatedAt *time.Time `json:"created_at"`

	// UpdatedAt: DHCP entry last modification date.
	UpdatedAt *time.Time `json:"updated_at"`

	// GatewayNetworkID: owning GatewayNetwork.
	GatewayNetworkID string `json:"gateway_network_id"`

	// MacAddress: mAC address of the client device.
	MacAddress string `json:"mac_address"`

	// IPAddress: assigned IP address.
	IPAddress net.IP `json:"ip_address"`

	// Hostname: hostname of the client device.
	Hostname string `json:"hostname"`

	// Type: entry type, either static (DHCP reservation) or dynamic (DHCP lease).
	// Default value: unknown
	Type DHCPEntryType `json:"type"`

	// Zone: zone of this DHCP entry.
	Zone scw.Zone `json:"zone"`
}

// Gateway: gateway.
type Gateway struct {
	// ID: ID of the gateway.
	ID string `json:"id"`

	// OrganizationID: owning Organization.
	OrganizationID string `json:"organization_id"`

	// ProjectID: owning Project.
	ProjectID string `json:"project_id"`

	// CreatedAt: gateway creation date.
	CreatedAt *time.Time `json:"created_at"`

	// UpdatedAt: gateway last modification date.
	UpdatedAt *time.Time `json:"updated_at"`

	// Type: gateway type (commercial offer).
	Type *GatewayType `json:"type"`

	// Status: current status of the gateway.
	// Default value: unknown
	Status GatewayStatus `json:"status"`

	// Name: name of the gateway.
	Name string `json:"name"`

	// Tags: tags associated with the gateway.
	Tags []string `json:"tags"`

	// IP: public IP address of the gateway.
	IP *IP `json:"ip"`

	// GatewayNetworks: gatewayNetwork objects attached to the gateway (each one represents a connection to a Private Network).
	GatewayNetworks []*GatewayNetwork `json:"gateway_networks"`

	// UpstreamDNSServers: array of DNS server IP addresses to override the gateway's default recursive DNS servers.
	UpstreamDNSServers []string `json:"upstream_dns_servers"`

	// Version: version of the running gateway software.
	Version *string `json:"version"`

	// CanUpgradeTo: newly available gateway software version that can be updated to.
	CanUpgradeTo *string `json:"can_upgrade_to"`

	// BastionEnabled: defines whether SSH bastion is enabled on the gateway.
	BastionEnabled bool `json:"bastion_enabled"`

	// BastionPort: port of the SSH bastion.
	BastionPort uint32 `json:"bastion_port"`

	// SMTPEnabled: defines whether SMTP traffic is allowed to pass through the gateway.
	SMTPEnabled bool `json:"smtp_enabled"`

	// IsLegacy: defines whether the gateway uses non-IPAM IP configurations.
	IsLegacy bool `json:"is_legacy"`

	// IPMobilityEnabled: defines whether the gateway uses routed IPs (IP mobility) instead of NAT IPs.
	IPMobilityEnabled bool `json:"ip_mobility_enabled"`

	// Zone: zone of the gateway.
	Zone scw.Zone `json:"zone"`
}

// PATRule: pat rule.
type PATRule struct {
	// ID: pAT rule ID.
	ID string `json:"id"`

	// GatewayID: gateway the PAT rule applies to.
	GatewayID string `json:"gateway_id"`

	// CreatedAt: pAT rule creation date.
	CreatedAt *time.Time `json:"created_at"`

	// UpdatedAt: pAT rule last modification date.
	UpdatedAt *time.Time `json:"updated_at"`

	// PublicPort: public port to listen on.
	PublicPort uint32 `json:"public_port"`

	// PrivateIP: private IP address to forward data to.
	PrivateIP net.IP `json:"private_ip"`

	// PrivatePort: private port to translate to.
	PrivatePort uint32 `json:"private_port"`

	// Protocol: protocol the rule applies to.
	// Default value: unknown
	Protocol PATRuleProtocol `json:"protocol"`

	// Zone: zone of the PAT rule.
	Zone scw.Zone `json:"zone"`
}

// SetDHCPEntriesRequestEntry: set dhcp entries request entry.
type SetDHCPEntriesRequestEntry struct {
	// MacAddress: mAC address to give a static entry to. A matching entry will be upgraded to a reservation, and a matching reservation will be updated.
	MacAddress string `json:"mac_address"`

	// IPAddress: IP address to give to the device.
	IPAddress net.IP `json:"ip_address"`
}

// SetPATRulesRequestRule: set pat rules request rule.
type SetPATRulesRequestRule struct {
	// PublicPort: public port to listen on. Uniquely identifies the rule, and a matching rule will be updated with the new parameters.
	PublicPort uint32 `json:"public_port"`

	// PrivateIP: private IP to forward data to.
	PrivateIP net.IP `json:"private_ip"`

	// PrivatePort: private port to translate to.
	PrivatePort uint32 `json:"private_port"`

	// Protocol: protocol the rule should apply to.
	// Default value: unknown
	Protocol PATRuleProtocol `json:"protocol"`
}

// UpdateGatewayNetworkRequestIpamConfig: update gateway network request ipam config.
type UpdateGatewayNetworkRequestIpamConfig struct {
	// PushDefaultRoute: enabling the default route also enables masquerading.
	PushDefaultRoute *bool `json:"push_default_route"`

	// IpamIPID: use this IPAM-booked IP ID as the Gateway's IP in this Private Network.
	IpamIPID *string `json:"ipam_ip_id"`
}

// CreateDHCPEntryRequest: create dhcp entry request.
type CreateDHCPEntryRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// GatewayNetworkID: gatewayNetwork on which to create a DHCP reservation.
	GatewayNetworkID string `json:"gateway_network_id"`

	// MacAddress: mAC address to give a static entry to.
	MacAddress string `json:"mac_address"`

	// IPAddress: IP address to give to the device.
	IPAddress net.IP `json:"ip_address"`
}

// CreateGatewayNetworkRequest: create gateway network request.
type CreateGatewayNetworkRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// GatewayID: public Gateway to connect.
	GatewayID string `json:"gateway_id"`

	// PrivateNetworkID: private Network to connect.
	PrivateNetworkID string `json:"private_network_id"`

	// EnableMasquerade: note: this setting is ignored when passing `ipam_config`.
	EnableMasquerade bool `json:"enable_masquerade"`

	// EnableDHCP: defaults to `true` if either `dhcp_id` or `dhcp` are present. If set to `true`, either `dhcp_id` or `dhcp` must be present.
	// Note: this setting is ignored when passing `ipam_config`.
	EnableDHCP *bool `json:"enable_dhcp,omitempty"`

	// DHCPID: ID of an existing DHCP configuration object to use for this GatewayNetwork.
	// Precisely one of DHCPID, DHCP, Address, IpamConfig must be set.
	DHCPID *string `json:"dhcp_id,omitempty"`

	// DHCP: new DHCP configuration object to use for this GatewayNetwork.
	// Precisely one of DHCPID, DHCP, Address, IpamConfig must be set.
	DHCP *CreateDHCPRequest `json:"dhcp,omitempty"`

	// Address: static IP address in CIDR format to to use without DHCP.
	// Precisely one of DHCPID, DHCP, Address, IpamConfig must be set.
	Address *scw.IPNet `json:"address,omitempty"`

	// IpamConfig: note: all or none of the GatewayNetworks for a single gateway can use the IPAM. DHCP and IPAM configurations cannot be mixed. Some products may require that the Public Gateway uses the IPAM, to ensure correct functionality.
	// Precisely one of DHCPID, DHCP, Address, IpamConfig must be set.
	IpamConfig *CreateGatewayNetworkRequestIpamConfig `json:"ipam_config,omitempty"`
}

// CreateGatewayRequest: create gateway request.
type CreateGatewayRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// ProjectID: scaleway Project to create the gateway in.
	ProjectID string `json:"project_id"`

	// Name: name for the gateway.
	Name string `json:"name"`

	// Tags: tags for the gateway.
	Tags []string `json:"tags"`

	// Type: gateway type (commercial offer type).
	Type string `json:"type"`

	// UpstreamDNSServers: array of DNS server IP addresses to override the gateway's default recursive DNS servers.
	UpstreamDNSServers []string `json:"upstream_dns_servers"`

	// IPID: existing IP address to attach to the gateway.
	IPID *string `json:"ip_id,omitempty"`

	// EnableSMTP: defines whether SMTP traffic should be allowed pass through the gateway.
	EnableSMTP bool `json:"enable_smtp"`

	// EnableBastion: defines whether SSH bastion should be enabled the gateway.
	EnableBastion bool `json:"enable_bastion"`

	// BastionPort: port of the SSH bastion.
	BastionPort *uint32 `json:"bastion_port,omitempty"`
}

// CreateIPRequest: create ip request.
type CreateIPRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// ProjectID: project to create the IP address in.
	ProjectID string `json:"project_id"`

	// Tags: tags to give to the IP address.
	Tags []string `json:"tags"`
}

// CreatePATRuleRequest: create pat rule request.
type CreatePATRuleRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// GatewayID: ID of the Gateway on which to create the rule.
	GatewayID string `json:"gateway_id"`

	// PublicPort: public port to listen on.
	PublicPort uint32 `json:"public_port"`

	// PrivateIP: private IP to forward data to.
	PrivateIP net.IP `json:"private_ip"`

	// PrivatePort: private port to translate to.
	PrivatePort uint32 `json:"private_port"`

	// Protocol: protocol the rule should apply to.
	// Default value: unknown
	Protocol PATRuleProtocol `json:"protocol"`
}

// DeleteDHCPEntryRequest: delete dhcp entry request.
type DeleteDHCPEntryRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// DHCPEntryID: ID of the DHCP entry to delete.
	DHCPEntryID string `json:"-"`
}

// DeleteDHCPRequest: delete dhcp request.
type DeleteDHCPRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// DHCPID: DHCP configuration ID to delete.
	DHCPID string `json:"-"`
}

// DeleteGatewayNetworkRequest: delete gateway network request.
type DeleteGatewayNetworkRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// GatewayNetworkID: ID of the GatewayNetwork to delete.
	GatewayNetworkID string `json:"-"`

	// CleanupDHCP: defines whether to clean up attached DHCP configurations (if any, and if not attached to another Gateway Network).
	CleanupDHCP bool `json:"cleanup_dhcp"`
}

// DeleteGatewayRequest: delete gateway request.
type DeleteGatewayRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// GatewayID: ID of the gateway to delete.
	GatewayID string `json:"-"`

	// CleanupDHCP: defines whether to clean up attached DHCP configurations (if any, and if not attached to another Gateway Network).
	CleanupDHCP bool `json:"cleanup_dhcp"`
}

// DeleteIPRequest: delete ip request.
type DeleteIPRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// IPID: ID of the IP address to delete.
	IPID string `json:"-"`
}

// DeletePATRuleRequest: delete pat rule request.
type DeletePATRuleRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// PatRuleID: ID of the PAT rule to delete.
	PatRuleID string `json:"-"`
}

// EnableIPMobilityRequest: enable ip mobility request.
type EnableIPMobilityRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// GatewayID: ID of the gateway to upgrade to IP mobility.
	GatewayID string `json:"-"`
}

// GetDHCPEntryRequest: get dhcp entry request.
type GetDHCPEntryRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// DHCPEntryID: ID of the DHCP entry to fetch.
	DHCPEntryID string `json:"-"`
}

// GetDHCPRequest: get dhcp request.
type GetDHCPRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// DHCPID: ID of the DHCP configuration to fetch.
	DHCPID string `json:"-"`
}

// GetGatewayNetworkRequest: get gateway network request.
type GetGatewayNetworkRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// GatewayNetworkID: ID of the GatewayNetwork to fetch.
	GatewayNetworkID string `json:"-"`
}

// GetGatewayRequest: get gateway request.
type GetGatewayRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// GatewayID: ID of the gateway to fetch.
	GatewayID string `json:"-"`
}

// GetIPRequest: get ip request.
type GetIPRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// IPID: ID of the IP address to get.
	IPID string `json:"-"`
}

// GetPATRuleRequest: get pat rule request.
type GetPATRuleRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// PatRuleID: ID of the PAT rule to get.
	PatRuleID string `json:"-"`
}

// ListDHCPEntriesRequest: list dhcp entries request.
type ListDHCPEntriesRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// OrderBy: order in which to return results.
	// Default value: created_at_asc
	OrderBy ListDHCPEntriesRequestOrderBy `json:"-"`

	// Page: page number.
	Page *int32 `json:"-"`

	// PageSize: DHCP entries per page.
	PageSize *uint32 `json:"-"`

	// GatewayNetworkID: filter for entries on this GatewayNetwork.
	GatewayNetworkID *string `json:"-"`

	// MacAddress: filter for entries with this MAC address.
	MacAddress *string `json:"-"`

	// IPAddress: filter for entries with this IP address.
	IPAddress *net.IP `json:"-"`

	// Hostname: filter for entries with this hostname substring.
	Hostname *string `json:"-"`

	// Type: filter for entries of this type.
	// Default value: unknown
	Type DHCPEntryType `json:"-"`
}

// ListDHCPEntriesResponse: list dhcp entries response.
type ListDHCPEntriesResponse struct {
	// DHCPEntries: DHCP entries in this page.
	DHCPEntries []*DHCPEntry `json:"dhcp_entries"`

	// TotalCount: total count of DHCP entries matching the filter.
	TotalCount uint32 `json:"total_count"`
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

// ListDHCPsRequest: list dhc ps request.
type ListDHCPsRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// OrderBy: order in which to return results.
	// Default value: created_at_asc
	OrderBy ListDHCPsRequestOrderBy `json:"-"`

	// Page: page number.
	Page *int32 `json:"-"`

	// PageSize: DHCP configurations per page.
	PageSize *uint32 `json:"-"`

	// OrganizationID: include only DHCP configuration objects in this Organization.
	OrganizationID *string `json:"-"`

	// ProjectID: include only DHCP configuration objects in this Project.
	ProjectID *string `json:"-"`

	// Address: filter for DHCP configuration objects with this DHCP server IP address (the gateway's address in the Private Network).
	Address *net.IP `json:"-"`

	// HasAddress: filter for DHCP configuration objects with subnets containing this IP address.
	HasAddress *net.IP `json:"-"`
}

// ListDHCPsResponse: list dhc ps response.
type ListDHCPsResponse struct {
	// Dhcps: first page of DHCP configuration objects.
	Dhcps []*DHCP `json:"dhcps"`

	// TotalCount: total count of DHCP configuration objects matching the filter.
	TotalCount uint32 `json:"total_count"`
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

// ListGatewayNetworksRequest: list gateway networks request.
type ListGatewayNetworksRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// OrderBy: order in which to return results.
	// Default value: created_at_asc
	OrderBy ListGatewayNetworksRequestOrderBy `json:"-"`

	// Page: page number.
	Page *int32 `json:"-"`

	// PageSize: gatewayNetworks per page.
	PageSize *uint32 `json:"-"`

	// GatewayID: filter for GatewayNetworks connected to this gateway.
	GatewayID *string `json:"-"`

	// PrivateNetworkID: filter for GatewayNetworks connected to this Private Network.
	PrivateNetworkID *string `json:"-"`

	// EnableMasquerade: filter for GatewayNetworks with this `enable_masquerade` setting.
	EnableMasquerade *bool `json:"-"`

	// DHCPID: filter for GatewayNetworks using this DHCP configuration.
	DHCPID *string `json:"-"`

	// Status: filter for GatewayNetworks with this current status this status. Use `unknown` to include all statuses.
	// Default value: unknown
	Status GatewayNetworkStatus `json:"-"`
}

// ListGatewayNetworksResponse: list gateway networks response.
type ListGatewayNetworksResponse struct {
	// GatewayNetworks: gatewayNetworks on this page.
	GatewayNetworks []*GatewayNetwork `json:"gateway_networks"`

	// TotalCount: total GatewayNetworks count matching the filter.
	TotalCount uint32 `json:"total_count"`
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

// ListGatewayTypesRequest: list gateway types request.
type ListGatewayTypesRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
}

// ListGatewayTypesResponse: list gateway types response.
type ListGatewayTypesResponse struct {
	// Types: available types of Public Gateway.
	Types []*GatewayType `json:"types"`
}

// ListGatewaysRequest: list gateways request.
type ListGatewaysRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// OrderBy: order in which to return results.
	// Default value: created_at_asc
	OrderBy ListGatewaysRequestOrderBy `json:"-"`

	// Page: page number to return.
	Page *int32 `json:"-"`

	// PageSize: gateways per page.
	PageSize *uint32 `json:"-"`

	// OrganizationID: include only gateways in this Organization.
	OrganizationID *string `json:"-"`

	// ProjectID: include only gateways in this Project.
	ProjectID *string `json:"-"`

	// Name: filter for gateways which have this search term in their name.
	Name *string `json:"-"`

	// Tags: filter for gateways with these tags.
	Tags []string `json:"-"`

	// Type: filter for gateways of this type.
	Type *string `json:"-"`

	// Status: filter for gateways with this current status. Use `unknown` to include all statuses.
	// Default value: unknown
	Status GatewayStatus `json:"-"`

	// PrivateNetworkID: filter for gateways attached to this Private nNetwork.
	PrivateNetworkID *string `json:"-"`
}

// ListGatewaysResponse: list gateways response.
type ListGatewaysResponse struct {
	// Gateways: gateways on this page.
	Gateways []*Gateway `json:"gateways"`

	// TotalCount: total count of gateways matching the filter.
	TotalCount uint32 `json:"total_count"`
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

// ListIPsRequest: list i ps request.
type ListIPsRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// OrderBy: order in which to return results.
	// Default value: created_at_asc
	OrderBy ListIPsRequestOrderBy `json:"-"`

	// Page: page number.
	Page *int32 `json:"-"`

	// PageSize: IP addresses per page.
	PageSize *uint32 `json:"-"`

	// OrganizationID: filter for IP addresses in this Organization.
	OrganizationID *string `json:"-"`

	// ProjectID: filter for IP addresses in this Project.
	ProjectID *string `json:"-"`

	// Tags: filter for IP addresses with these tags.
	Tags []string `json:"-"`

	// Reverse: filter for IP addresses that have a reverse containing this string.
	Reverse *string `json:"-"`

	// IsFree: filter based on whether the IP is attached to a gateway or not.
	IsFree *bool `json:"-"`
}

// ListIPsResponse: list i ps response.
type ListIPsResponse struct {
	// IPs: IP addresses on this page.
	IPs []*IP `json:"ips"`

	// TotalCount: total count of IP addresses matching the filter.
	TotalCount uint32 `json:"total_count"`
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

// ListPATRulesRequest: list pat rules request.
type ListPATRulesRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// OrderBy: order in which to return results.
	// Default value: created_at_asc
	OrderBy ListPATRulesRequestOrderBy `json:"-"`

	// Page: page number.
	Page *int32 `json:"-"`

	// PageSize: pAT rules per page.
	PageSize *uint32 `json:"-"`

	// GatewayID: filter for PAT rules on this Gateway.
	GatewayID *string `json:"-"`

	// PrivateIP: filter for PAT rules targeting this private ip.
	PrivateIP *net.IP `json:"-"`

	// Protocol: filter for PAT rules with this protocol.
	// Default value: unknown
	Protocol PATRuleProtocol `json:"-"`
}

// ListPATRulesResponse: list pat rules response.
type ListPATRulesResponse struct {
	// PatRules: array of PAT rules matching the filter.
	PatRules []*PATRule `json:"pat_rules"`

	// TotalCount: total count of PAT rules matching the filter.
	TotalCount uint32 `json:"total_count"`
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

// RefreshSSHKeysRequest: refresh ssh keys request.
type RefreshSSHKeysRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// GatewayID: ID of the gateway to refresh SSH keys on.
	GatewayID string `json:"-"`
}

// SetDHCPEntriesRequest: set dhcp entries request.
type SetDHCPEntriesRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// GatewayNetworkID: ID of the Gateway Network on which to set DHCP reservation list.
	GatewayNetworkID string `json:"gateway_network_id"`

	// DHCPEntries: new list of DHCP reservations.
	DHCPEntries []*SetDHCPEntriesRequestEntry `json:"dhcp_entries"`
}

// SetDHCPEntriesResponse: set dhcp entries response.
type SetDHCPEntriesResponse struct {
	// DHCPEntries: list of DHCP entries.
	DHCPEntries []*DHCPEntry `json:"dhcp_entries"`
}

// SetPATRulesRequest: set pat rules request.
type SetPATRulesRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// GatewayID: ID of the gateway on which to set the PAT rules.
	GatewayID string `json:"gateway_id"`

	// PatRules: new list of PAT rules.
	PatRules []*SetPATRulesRequestRule `json:"pat_rules"`
}

// SetPATRulesResponse: set pat rules response.
type SetPATRulesResponse struct {
	// PatRules: list of PAT rules.
	PatRules []*PATRule `json:"pat_rules"`
}

// UpdateDHCPEntryRequest: update dhcp entry request.
type UpdateDHCPEntryRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// DHCPEntryID: ID of the DHCP entry to update.
	DHCPEntryID string `json:"-"`

	// IPAddress: new IP address to give to the device.
	IPAddress *net.IP `json:"ip_address,omitempty"`
}

// UpdateDHCPRequest: update dhcp request.
type UpdateDHCPRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// DHCPID: DHCP configuration to update.
	DHCPID string `json:"-"`

	// Subnet: subnet for the DHCP server.
	Subnet *scw.IPNet `json:"subnet,omitempty"`

	// Address: IP address of the DHCP server. This will be the Public Gateway's address in the Private Network. It must be part of config's subnet.
	Address *net.IP `json:"address,omitempty"`

	// PoolLow: low IP (inclusive) of the dynamic address pool. Must be in the config's subnet.
	PoolLow *net.IP `json:"pool_low,omitempty"`

	// PoolHigh: high IP (inclusive) of the dynamic address pool. Must be in the config's subnet.
	PoolHigh *net.IP `json:"pool_high,omitempty"`

	// EnableDynamic: defines whether to enable dynamic pooling of IPs. When false, only pre-existing DHCP reservations will be handed out. Defaults to true.
	EnableDynamic *bool `json:"enable_dynamic,omitempty"`

	// ValidLifetime: how long DHCP entries will be valid for.
	ValidLifetime *scw.Duration `json:"valid_lifetime,omitempty"`

	// RenewTimer: after how long a renew will be attempted. Must be 30s lower than `rebind_timer`.
	RenewTimer *scw.Duration `json:"renew_timer,omitempty"`

	// RebindTimer: after how long a DHCP client will query for a new lease if previous renews fail. Must be 30s lower than `valid_lifetime`.
	RebindTimer *scw.Duration `json:"rebind_timer,omitempty"`

	// PushDefaultRoute: defines whether the gateway should push a default route to DHCP clients, or only hand out IPs.
	PushDefaultRoute *bool `json:"push_default_route,omitempty"`

	// PushDNSServer: defines whether the gateway should push custom DNS servers to clients. This allows for instance hostname -> IP resolution.
	PushDNSServer *bool `json:"push_dns_server,omitempty"`

	// DNSServersOverride: array of DNS server IP addresses used to override the DNS server list pushed to DHCP clients, instead of the gateway itself.
	DNSServersOverride *[]string `json:"dns_servers_override,omitempty"`

	// DNSSearch: array of search paths in addition to the pushed DNS configuration.
	DNSSearch *[]string `json:"dns_search,omitempty"`

	// DNSLocalName: tLD given to hostnames in the Private Networks. If an instance with hostname `foo` gets a lease, and this is set to `bar`, `foo.bar` will resolve. Allowed characters are `a-z0-9-.`.
	DNSLocalName *string `json:"dns_local_name,omitempty"`
}

// UpdateGatewayNetworkRequest: update gateway network request.
type UpdateGatewayNetworkRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// GatewayNetworkID: ID of the GatewayNetwork to update.
	GatewayNetworkID string `json:"-"`

	// EnableMasquerade: note: this setting is ignored when passing `ipam_config`.
	EnableMasquerade *bool `json:"enable_masquerade,omitempty"`

	// EnableDHCP: defaults to `true` if `dhcp_id` is present. If set to `true`, `dhcp_id` must be present.
	// Note: this setting is ignored when passing `ipam_config`.
	EnableDHCP *bool `json:"enable_dhcp,omitempty"`

	// DHCPID: ID of the new DHCP configuration object to use with this GatewayNetwork.
	// Precisely one of DHCPID, Address, IpamConfig must be set.
	DHCPID *string `json:"dhcp_id,omitempty"`

	// Address: new static IP address.
	// Precisely one of DHCPID, Address, IpamConfig must be set.
	Address *scw.IPNet `json:"address,omitempty"`

	// IpamConfig: note: all or none of the GatewayNetworks for a single gateway can use the IPAM. DHCP and IPAM configurations cannot be mixed. Some products may require that the Public Gateway uses the IPAM, to ensure correct functionality.
	// Precisely one of DHCPID, Address, IpamConfig must be set.
	IpamConfig *UpdateGatewayNetworkRequestIpamConfig `json:"ipam_config,omitempty"`
}

// UpdateGatewayRequest: update gateway request.
type UpdateGatewayRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// GatewayID: ID of the gateway to update.
	GatewayID string `json:"-"`

	// Name: name for the gateway.
	Name *string `json:"name,omitempty"`

	// Tags: tags for the gateway.
	Tags *[]string `json:"tags,omitempty"`

	// UpstreamDNSServers: array of DNS server IP addresses to override the gateway's default recursive DNS servers.
	UpstreamDNSServers *[]string `json:"upstream_dns_servers,omitempty"`

	// EnableBastion: defines whether SSH bastion should be enabled the gateway.
	EnableBastion *bool `json:"enable_bastion,omitempty"`

	// BastionPort: port of the SSH bastion.
	BastionPort *uint32 `json:"bastion_port,omitempty"`

	// EnableSMTP: defines whether SMTP traffic should be allowed to pass through the gateway.
	EnableSMTP *bool `json:"enable_smtp,omitempty"`
}

// UpdateIPRequest: update ip request.
type UpdateIPRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// IPID: ID of the IP address to update.
	IPID string `json:"-"`

	// Tags: tags to give to the IP address.
	Tags *[]string `json:"tags,omitempty"`

	// Reverse: reverse to set on the address. Empty string to unset.
	Reverse *string `json:"reverse,omitempty"`

	// GatewayID: gateway to attach the IP address to. Empty string to detach.
	GatewayID *string `json:"gateway_id,omitempty"`
}

// UpdatePATRuleRequest: update pat rule request.
type UpdatePATRuleRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// PatRuleID: ID of the PAT rule to update.
	PatRuleID string `json:"-"`

	// PublicPort: public port to listen on.
	PublicPort *uint32 `json:"public_port,omitempty"`

	// PrivateIP: private IP to forward data to.
	PrivateIP *net.IP `json:"private_ip,omitempty"`

	// PrivatePort: private port to translate to.
	PrivatePort *uint32 `json:"private_port,omitempty"`

	// Protocol: protocol the rule should apply to.
	// Default value: unknown
	Protocol PATRuleProtocol `json:"protocol"`
}

// UpgradeGatewayRequest: upgrade gateway request.
type UpgradeGatewayRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// GatewayID: ID of the gateway to upgrade.
	GatewayID string `json:"-"`
}

// Public Gateways API.
type API struct {
	client *scw.Client
}

// NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}
func (s *API) Zones() []scw.Zone {
	return []scw.Zone{scw.ZoneFrPar1, scw.ZoneFrPar2, scw.ZoneNlAms1, scw.ZoneNlAms2, scw.ZoneNlAms3, scw.ZonePlWaw1, scw.ZonePlWaw2, scw.ZonePlWaw3}
}

// ListGateways: List Public Gateways in a given Scaleway Organization or Project. By default, results are displayed in ascending order of creation date.
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
		Method: "GET",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateways",
		Query:  query,
	}

	var resp ListGatewaysResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetGateway: Get details of a Public Gateway, specified by its gateway ID. The response object contains full details of the gateway, including its **name**, **type**, **status** and more.
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
		Method: "GET",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateways/" + fmt.Sprint(req.GatewayID) + "",
	}

	var resp Gateway

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateGateway: Create a new Public Gateway in the specified Scaleway Project, defining its **name**, **type** and other configuration details such as whether to enable SSH bastion.
func (s *API) CreateGateway(req *CreateGatewayRequest, opts ...scw.RequestOption) (*Gateway, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("gw")
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateways",
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

// UpdateGateway: Update the parameters of an existing Public Gateway, for example, its **name**, **tags**, **SSH bastion configuration**, and **DNS servers**.
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
		Method: "PATCH",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateways/" + fmt.Sprint(req.GatewayID) + "",
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

// DeleteGateway: Delete an existing Public Gateway, specified by its gateway ID. This action is irreversible.
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
		Method: "DELETE",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateways/" + fmt.Sprint(req.GatewayID) + "",
		Query:  query,
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// UpgradeGateway: Upgrade a given Public Gateway to the newest software version. This applies the latest bugfixes and features to your Public Gateway, but its service will be interrupted during the update.
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
		Method: "POST",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateways/" + fmt.Sprint(req.GatewayID) + "/upgrade",
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

// EnableIPMobility: Upgrade a Public Gateway to IP mobility (move from NAT IP to routed IP). This is idempotent: repeated calls after the first will return no error but have no effect.
func (s *API) EnableIPMobility(req *EnableIPMobilityRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.GatewayID) == "" {
		return errors.New("field GatewayID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateways/" + fmt.Sprint(req.GatewayID) + "/enable-ip-mobility",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return err
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ListGatewayNetworks: List the connections between Public Gateways and Private Networks (a connection = a GatewayNetwork). You can choose to filter by `gateway-id` to list all Private Networks attached to the specified Public Gateway, or by `private_network_id` to list all Public Gateways attached to the specified Private Network. Other query parameters are also available. The result is an array of GatewayNetwork objects, each giving details of the connection between a given Public Gateway and a given Private Network.
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
		Method: "GET",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateway-networks",
		Query:  query,
	}

	var resp ListGatewayNetworksResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetGatewayNetwork: Get details of a given connection between a Public Gateway and a Private Network (this connection = a GatewayNetwork), specified by its `gateway_network_id`. The response object contains details of the connection including the IDs of the Public Gateway and Private Network, the dates the connection was created/updated and its configuration settings.
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
		Method: "GET",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateway-networks/" + fmt.Sprint(req.GatewayNetworkID) + "",
	}

	var resp GatewayNetwork

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateGatewayNetwork: Attach a specific Public Gateway to a specific Private Network (create a GatewayNetwork). You can configure parameters for the connection including DHCP settings, whether to enable masquerade (dynamic NAT), and more.
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
		Method: "POST",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateway-networks",
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

// UpdateGatewayNetwork: Update the configuration parameters of a connection between a given Public Gateway and Private Network (the connection = a GatewayNetwork). Updatable parameters include DHCP settings and whether to enable traffic masquerade (dynamic NAT).
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
		Method: "PATCH",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateway-networks/" + fmt.Sprint(req.GatewayNetworkID) + "",
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

// DeleteGatewayNetwork: Detach a given Public Gateway from a given Private Network, i.e. delete a GatewayNetwork specified by a gateway_network_id.
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
		Method: "DELETE",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateway-networks/" + fmt.Sprint(req.GatewayNetworkID) + "",
		Query:  query,
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ListDHCPs: List DHCP configurations, optionally filtering by Organization, Project, Public Gateway IP address or more. The response is an array of DHCP configuration objects, each identified by a DHCP ID and containing configuration settings for the assignment of IP addresses to devices on a Private Network attached to a Public Gateway. Note that the response does not contain the IDs of any Private Network / Public Gateway the configuration is attached to. Use the `List Public Gateway connections to Private Networks` method for that purpose, filtering on DHCP ID.
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
		Method: "GET",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/dhcps",
		Query:  query,
	}

	var resp ListDHCPsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetDHCP: Get a DHCP configuration object, identified by its DHCP ID. The response object contains configuration settings for the assignment of IP addresses to devices on a Private Network attached to a Public Gateway. Note that the response does not contain the IDs of any Private Network / Public Gateway the configuration is attached to. Use the `List Public Gateway connections to Private Networks` method for that purpose, filtering on DHCP ID.
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
		Method: "GET",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/dhcps/" + fmt.Sprint(req.DHCPID) + "",
	}

	var resp DHCP

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateDHCP: Create a new DHCP configuration object, containing settings for the assignment of IP addresses to devices on a Private Network attached to a Public Gateway. The response object includes the ID of the DHCP configuration object. You can use this ID as part of a call to `Create a Public Gateway connection to a Private Network` or `Update a Public Gateway connection to a Private Network` to directly apply this DHCP configuration.
func (s *API) CreateDHCP(req *CreateDHCPRequest, opts ...scw.RequestOption) (*DHCP, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/dhcps",
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

// UpdateDHCP: Update a DHCP configuration object, identified by its DHCP ID.
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
		Method: "PATCH",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/dhcps/" + fmt.Sprint(req.DHCPID) + "",
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

// DeleteDHCP: Delete a DHCP configuration object, identified by its DHCP ID. Note that you cannot delete a DHCP configuration object that is currently being used by a Gateway Network.
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
		Method: "DELETE",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/dhcps/" + fmt.Sprint(req.DHCPID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ListDHCPEntries: List DHCP entries, whether dynamically assigned and/or statically reserved. DHCP entries can be filtered by the Gateway Network they are on, their MAC address, IP address, type or hostname.
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
		Method: "GET",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/dhcp-entries",
		Query:  query,
	}

	var resp ListDHCPEntriesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetDHCPEntry: Get a DHCP entry, specified by its DHCP entry ID.
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
		Method: "GET",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/dhcp-entries/" + fmt.Sprint(req.DHCPEntryID) + "",
	}

	var resp DHCPEntry

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateDHCPEntry: Create a static DHCP reservation, specifying the Gateway Network for the reservation, the MAC address of the target device and the IP address to assign this device. The response is a DHCP entry object, confirming the ID and configuration details of the static DHCP reservation.
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
		Method: "POST",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/dhcp-entries",
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

// UpdateDHCPEntry: Update the IP address for a DHCP entry, specified by its DHCP entry ID. You can update an existing DHCP entry of any type (`reservation` (static), `lease` (dynamic) or `unknown`), but in manually updating the IP address the entry will necessarily be of type `reservation` after the update.
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
		Method: "PATCH",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/dhcp-entries/" + fmt.Sprint(req.DHCPEntryID) + "",
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

// SetDHCPEntries: Set the list of DHCP reservations attached to a Gateway Network. Reservations are identified by their MAC address, and will sync the current DHCP entry list to the given list, creating, updating or deleting DHCP entries accordingly.
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
		Method: "PUT",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/dhcp-entries",
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

// DeleteDHCPEntry: Delete a static DHCP reservation, identified by its DHCP entry ID. Note that you cannot delete DHCP entries of type `lease`, these are deleted automatically when their time-to-live expires.
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
		Method: "DELETE",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/dhcp-entries/" + fmt.Sprint(req.DHCPEntryID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ListPATRules: List PAT rules. You can filter by gateway ID to list all PAT rules for a particular gateway, or filter for PAT rules targeting a specific IP address or using a specific protocol.
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
		Method: "GET",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/pat-rules",
		Query:  query,
	}

	var resp ListPATRulesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPATRule: Get a PAT rule, specified by its PAT rule ID. The response object gives full details of the PAT rule, including the Public Gateway it belongs to and the configuration settings in terms of public / private ports, private IP and protocol.
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
		Method: "GET",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/pat-rules/" + fmt.Sprint(req.PatRuleID) + "",
	}

	var resp PATRule

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreatePATRule: Create a new PAT rule on a specified Public Gateway, defining the protocol to use, public port to listen on, and private port / IP address to map to.
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
		Method: "POST",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/pat-rules",
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

// UpdatePATRule: Update a PAT rule, specified by its PAT rule ID. Configuration settings including private/public port, private IP address and protocol can all be updated.
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
		Method: "PATCH",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/pat-rules/" + fmt.Sprint(req.PatRuleID) + "",
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

// SetPATRules: Set a definitive list of PAT rules attached to a Public Gateway. Each rule is identified by its public port and protocol. This will sync the current PAT rule list on the gateway with the new list, creating, updating or deleting PAT rules accordingly.
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
		Method: "PUT",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/pat-rules",
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

// DeletePATRule: Delete a PAT rule, identified by its PAT rule ID. This action is irreversible.
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
		Method: "DELETE",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/pat-rules/" + fmt.Sprint(req.PatRuleID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ListGatewayTypes: List the different Public Gateway commercial offer types available at Scaleway. The response is an array of objects describing the name and technical details of each available gateway type.
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
		Method: "GET",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateway-types",
	}

	var resp ListGatewayTypesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListIPs: List Public Gateway flexible IP addresses. A number of filter options are available for limiting results in the response.
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
		Method: "GET",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/ips",
		Query:  query,
	}

	var resp ListIPsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetIP: Get details of a Public Gateway flexible IP address, identified by its IP ID. The response object contains information including which (if any) Public Gateway using this IP address, the reverse and various other metadata.
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
		Method: "GET",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/ips/" + fmt.Sprint(req.IPID) + "",
	}

	var resp IP

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateIP: Create (reserve) a new flexible IP address that can be used for a Public Gateway in a specified Scaleway Project.
func (s *API) CreateIP(req *CreateIPRequest, opts ...scw.RequestOption) (*IP, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/ips",
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

// UpdateIP: Update details of an existing flexible IP address, including its tags, reverse and the Public Gateway it is assigned to.
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
		Method: "PATCH",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/ips/" + fmt.Sprint(req.IPID) + "",
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

// DeleteIP: Delete a flexible IP address from your account. This action is irreversible.
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
		Method: "DELETE",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/ips/" + fmt.Sprint(req.IPID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// RefreshSSHKeys: Refresh the SSH keys of a given Public Gateway, specified by its gateway ID. This adds any new SSH keys in the gateway's Scaleway Project to the gateway itself.
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
		Method: "POST",
		Path:   "/vpc-gw/v1/zones/" + fmt.Sprint(req.Zone) + "/gateways/" + fmt.Sprint(req.GatewayID) + "/refresh-ssh-keys",
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
