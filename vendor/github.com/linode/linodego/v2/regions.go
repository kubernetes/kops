package linodego

import (
	"context"
	"time"
)

// RegionCapability constants start with Capability and include all known Linode region capabilities.
type RegionCapability string

// This is an enumeration of Capabilities Linode offers that can be referenced
// through the user-facing parts of the application.
const (
	CapabilityACLP                             RegionCapability = "Akamai Cloud Pulse"
	CapabilityACLPStreams                      RegionCapability = "Akamai Cloud Pulse Streams"
	CapabilityAkamaiRAMProtection              RegionCapability = "Akamai RAM Protection"
	CapabilityACLB                             RegionCapability = "Akamai Cloud Load Balancer"
	CapabilityBackups                          RegionCapability = "Backups"
	CapabilityBlockStorage                     RegionCapability = "Block Storage"
	CapabilityBlockStorageEncryption           RegionCapability = "Block Storage Encryption"
	CapabilityBlockStorageMigrations           RegionCapability = "Block Storage Migrations"
	CapabilityBlockStoragePerformanceB1        RegionCapability = "Block Storage Performance B1"
	CapabilityBlockStoragePerformanceB1Default RegionCapability = "Block Storage Performance B1 Default"
	CapabilityCloudFirewall                    RegionCapability = "Cloud Firewall"
	CapabilityCloudFirewallRuleSet             RegionCapability = "Cloud Firewall Rule Set"
	CapabilityCloudNAT                         RegionCapability = "Cloud NAT"
	CapabilityDBAAS                            RegionCapability = "Managed Databases"
	CapabilityDBAASBeta                        RegionCapability = "Managed Databases Beta"
	CapabilityDiskEncryption                   RegionCapability = "Disk Encryption"
	CapabilityDistributedPlans                 RegionCapability = "Distributed Plans"
	CapabilityEdgePlans                        RegionCapability = "Edge Plans"
	CapabilityGPU                              RegionCapability = "GPU Linodes"
	CapabilityKubernetesEnterprise             RegionCapability = "Kubernetes Enterprise"
	CapabilityKubernetesEnterpriseBYOVPC       RegionCapability = "Kubernetes Enterprise BYO VPC"
	CapabilityKubernetesEnterpriseDualStack    RegionCapability = "Kubernetes Enterprise Dual Stack"
	CapabilityLADiskEncryption                 RegionCapability = "LA Disk Encryption"
	CapabilityLinodeInterfaces                 RegionCapability = "Linode Interfaces"
	CapabilityLinodes                          RegionCapability = "Linodes"
	CapabilityLKE                              RegionCapability = "Kubernetes"
	CapabilityLKEControlPlaneACL               RegionCapability = "LKE Network Access Control List (IP ACL)"
	CapabilityLkeHaControlPlanes               RegionCapability = "LKE HA Control Planes"
	CapabilityMachineImages                    RegionCapability = "Machine Images"
	CapabilityMaintenancePolicy                RegionCapability = "Maintenance Policy"
	CapabilityMetadata                         RegionCapability = "Metadata"
	CapabilityNLB                              RegionCapability = "Network LoadBalancer"
	CapabilityNodeBalancers                    RegionCapability = "NodeBalancers"
	CapabilityObjectStorage                    RegionCapability = "Object Storage"
	CapabilityObjectStorageAccessKeyRegions    RegionCapability = "Object Storage Access Key Regions"
	CapabilityObjectStorageEndpointTypes       RegionCapability = "Object Storage Endpoint Types"
	CapabilityPlacementGroup                   RegionCapability = "Placement Group"
	CapabilityPremiumPlans                     RegionCapability = "Premium Plans"
	CapabilityQuadraT1UVPU                     RegionCapability = "NETINT Quadra T1U"
	CapabilitySMTPEnabled                      RegionCapability = "SMTP Enabled"
	CapabilityStackScripts                     RegionCapability = "StackScripts"
	CapabilitySupportTicketSeverity            RegionCapability = "Support Ticket Severity"
	CapabilityVlans                            RegionCapability = "Vlans"
	CapabilityVPCs                             RegionCapability = "VPCs"
	CapabilityVPCDualStack                     RegionCapability = "VPC Dual Stack"
	CapabilityVPCIPv6LargePrefixes             RegionCapability = "VPC IPv6 Large Prefixes"
	CapabilityVPCIPv6Stack                     RegionCapability = "VPC IPv6 Stack"
	CapabilityVPCsExtra                        RegionCapability = "VPCs Extra"
	CapabilityVPCCustomIPv4Ranges              RegionCapability = "Custom VPC IPv4 Ranges"
)

// Region-related endpoints have a custom expiry time as the
// `status` field may update for database outages.
var cacheExpiryTime = time.Minute

// Region represents a linode region object
type Region struct {
	ID      string `json:"id"`
	Country string `json:"country"`

	// A List of enums from the above constants
	Capabilities []string `json:"capabilities"`

	Monitors RegionMonitors `json:"monitors"`

	Status   string `json:"status"`
	Label    string `json:"label"`
	SiteType string `json:"site_type"`

	Resolvers            RegionResolvers             `json:"resolvers"`
	PlacementGroupLimits *RegionPlacementGroupLimits `json:"placement_group_limits"`
}

// RegionResolvers contains the DNS resolvers of a region
type RegionResolvers struct {
	IPv4 string `json:"ipv4"`
	IPv6 string `json:"ipv6"`
}

// RegionMonitors contains the monitoring configuration for a region
type RegionMonitors struct {
	Alerts  []string `json:"alerts"`
	Metrics []string `json:"metrics"`
}

// RegionPlacementGroupLimits contains information about the
// placement group limits for the current user in the current region.
type RegionPlacementGroupLimits struct {
	MaximumPGsPerCustomer int `json:"maximum_pgs_per_customer"`
	MaximumLinodesPerPG   int `json:"maximum_linodes_per_pg"`
}

// ListRegions lists Regions. This endpoint is cached by default.
func (c *Client) ListRegions(ctx context.Context, opts *ListOptions) ([]Region, error) {
	endpoint, err := generateListCacheURL("regions", opts)
	if err != nil {
		return nil, err
	}

	if result := c.getCachedResponse(endpoint); result != nil {
		return result.([]Region), nil
	}

	response, err := getPaginatedResults[Region](ctx, c, "regions", opts)
	if err != nil {
		return nil, err
	}

	c.addCachedResponse(endpoint, response, &cacheExpiryTime)

	return response, nil
}

// GetRegion gets the template with the provided ID. This endpoint is cached by default.
func (c *Client) GetRegion(ctx context.Context, regionID string) (*Region, error) {
	e := formatAPIPath("regions/%s", regionID)

	if result := c.getCachedResponse(e); result != nil {
		result := result.(Region)
		return &result, nil
	}

	response, err := doGETRequest[Region](ctx, c, e)
	if err != nil {
		return nil, err
	}

	c.addCachedResponse(e, response, &cacheExpiryTime)

	return response, nil
}
