package hcloud

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter gen ./...

/*
This file generates conversions methods between the schema and the hcloud package.
Goverter (https://github.com/jmattheis/goverter) is used to generate these conversion
methods. Goverter is configured using comments in and on the [converter] interface.
A struct implementing the interface methods, [converterImpl], is generated in zz_schema.go.
The generated methods are then wrapped in schema.go to be exported.

You can find a documentation of goverter here: https://goverter.jmattheis.de/
*/

// goverter:converter
//
// Specify where and in which package to output the generated
// conversion methods.
// goverter:output:file zz_schema.go
// goverter:output:package github.com/hetznercloud/hcloud-go/v2/hcloud
//
// In case of *T -> T conversion, use zero value if *T is nil.
// goverter:useZeroValueOnPointerInconsistency yes
//
// Do not deep copy in case of *T -> *T conversion.
// goverter:skipCopySameType yes
//
// Explicit conversion methods are needed for non-trivial cases
// where the target, source or both are of primitive types. Struct
// to struct conversions can be handled by goverter directly.
// goverter:extend ipFromString
// goverter:extend stringFromIP
// goverter:extend ipNetFromString
// goverter:extend stringFromIPNet
// goverter:extend timeToTimePtr
// goverter:extend serverFromInt64
// goverter:extend int64FromServer
// goverter:extend networkFromInt64
// goverter:extend int64FromNetwork
// goverter:extend loadBalancerFromInt64
// goverter:extend int64FromLoadBalancer
// goverter:extend volumeFromInt64
// goverter:extend int64FromVolume
// goverter:extend certificateFromInt64
// goverter:extend int64FromCertificate
// goverter:extend locationFromString
// goverter:extend stringFromLocation
// goverter:extend serverTypeFromInt64
// goverter:extend int64FromServerType
// goverter:extend floatingIPFromInt64
// goverter:extend int64FromFloatingIP
// goverter:extend mapFromFloatingIPDNSPtrSchema
// goverter:extend floatingIPDNSPtrSchemaFromMap
// goverter:extend mapFromPrimaryIPDNSPtrSchema
// goverter:extend primaryIPDNSPtrSchemaFromMap
// goverter:extend mapFromServerPublicNetIPv6DNSPtrSchema
// goverter:extend serverPublicNetIPv6DNSPtrSchemaFromMap
// goverter:extend firewallStatusFromSchemaServerFirewall
// goverter:extend serverFirewallSchemaFromFirewallStatus
// goverter:extend durationFromIntSeconds
// goverter:extend intSecondsFromDuration
// goverter:extend serverFromImageCreatedFromSchema
// goverter:extend serverMetricsTimeSeriesFromSchema
// goverter:extend loadBalancerMetricsTimeSeriesFromSchema
// goverter:extend stringPtrFromLoadBalancerServiceProtocol
// goverter:extend stringPtrFromNetworkZone
// goverter:extend schemaFromLoadBalancerCreateOptsTargetLabelSelector
// goverter:extend schemaFromLoadBalancerCreateOptsTargetServer
// goverter:extend schemaFromLoadBalancerCreateOptsTargetIP
// goverter:extend stringMapToStringMapPtr
// goverter:extend int64SlicePtrFromCertificatePtrSlice
// goverter:extend stringSlicePtrFromStringSlice
// goverter:extend locationFromServerTypeLocationSchema
type converter interface {

	// goverter:map Error.Code ErrorCode
	// goverter:map Error.Message ErrorMessage
	ActionFromSchema(schema.Action) *Action

	// goverter:map . Error | schemaActionErrorFromAction
	SchemaFromAction(*Action) schema.Action

	ActionsFromSchema([]schema.Action) []*Action

	SchemaFromActions([]*Action) []schema.Action

	// goverter:map . IP | ipFromFloatingIPSchema
	// goverter:map . Network | networkFromFloatingIPSchema
	FloatingIPFromSchema(schema.FloatingIP) *FloatingIP

	// goverter:map . IP | floatingIPToIPString
	SchemaFromFloatingIP(*FloatingIP) schema.FloatingIP

	// goverter:map . IP | ipFromPrimaryIPSchema
	// goverter:map . Network | networkFromPrimaryIPSchema
	PrimaryIPFromSchema(schema.PrimaryIP) *PrimaryIP

	// goverter:map . IP | primaryIPToIPString
	// goverter:map AssigneeID | mapZeroInt64ToNil
	SchemaFromPrimaryIP(*PrimaryIP) schema.PrimaryIP

	SchemaFromPrimaryIPCreateOpts(PrimaryIPCreateOpts) schema.PrimaryIPCreateRequest
	SchemaFromPrimaryIPUpdateOpts(PrimaryIPUpdateOpts) schema.PrimaryIPUpdateRequest
	SchemaFromPrimaryIPChangeDNSPtrOpts(PrimaryIPChangeDNSPtrOpts) schema.PrimaryIPActionChangeDNSPtrRequest
	SchemaFromPrimaryIPChangeProtectionOpts(PrimaryIPChangeProtectionOpts) schema.PrimaryIPActionChangeProtectionRequest
	SchemaFromPrimaryIPAssignOpts(PrimaryIPAssignOpts) schema.PrimaryIPActionAssignRequest

	ISOFromSchema(schema.ISO) *ISO

	// We cannot use goverter settings when mapping a struct to a struct pointer
	// See [converter.ISOFromSchema]
	// See https://github.com/jmattheis/goverter/issues/114
	// goverter:map DeprecatableResource.Deprecation.UnavailableAfter Deprecated
	intISOFromSchema(schema.ISO) ISO

	SchemaFromISO(*ISO) schema.ISO

	LocationFromSchema(schema.Location) *Location

	SchemaFromLocation(*Location) schema.Location

	DatacenterFromSchema(schema.Datacenter) *Datacenter

	SchemaFromDatacenter(*Datacenter) schema.Datacenter

	ServerFromSchema(schema.Server) *Server

	// goverter:map OutgoingTraffic | mapZeroUint64ToNil
	// goverter:map IngoingTraffic | mapZeroUint64ToNil
	// goverter:map BackupWindow | mapEmptyStringToNil
	SchemaFromServer(*Server) schema.Server

	ServerPublicNetFromSchema(schema.ServerPublicNet) ServerPublicNet

	SchemaFromServerPublicNet(ServerPublicNet) schema.ServerPublicNet

	ServerPublicNetIPv4FromSchema(schema.ServerPublicNetIPv4) ServerPublicNetIPv4

	SchemaFromServerPublicNetIPv4(ServerPublicNetIPv4) schema.ServerPublicNetIPv4

	// goverter:map . IP | ipFromServerPublicNetIPv6Schema
	// goverter:map . Network | ipNetFromServerPublicNetIPv6Schema
	ServerPublicNetIPv6FromSchema(schema.ServerPublicNetIPv6) ServerPublicNetIPv6

	// goverter:map Network IP
	SchemaFromServerPublicNetIPv6(ServerPublicNetIPv6) schema.ServerPublicNetIPv6

	// goverter:map AliasIPs Aliases
	ServerPrivateNetFromSchema(schema.ServerPrivateNet) ServerPrivateNet

	// goverter:map Aliases AliasIPs
	SchemaFromServerPrivateNet(ServerPrivateNet) schema.ServerPrivateNet

	// goverter:map Prices Pricings
	ServerTypeFromSchema(schema.ServerType) *ServerType

	// goverter:map . Location
	serverTypeLocationFromSchema(schema.ServerTypeLocation) ServerTypeLocation

	// goverter:map Pricings Prices
	// goverter:map DeprecatableResource.Deprecation Deprecated | isDeprecationNotNil
	SchemaFromServerType(*ServerType) schema.ServerType

	// goverter:map Location.ID ID
	// goverter:map Location.Name Name
	schemaFromServerTypeLocation(location ServerTypeLocation) schema.ServerTypeLocation

	ImageFromSchema(schema.Image) *Image

	SchemaFromImage(*Image) schema.Image

	// Needed because of how goverter works internally, see https://github.com/jmattheis/goverter/issues/114
	// goverter:map ImageSize | mapZeroFloat32ToNil
	intSchemaFromImage(Image) schema.Image

	// goverter:ignore Currency
	// goverter:ignore VATRate
	PriceFromSchema(schema.Price) Price

	SSHKeyFromSchema(schema.SSHKey) *SSHKey

	SchemaFromSSHKey(*SSHKey) schema.SSHKey

	VolumeFromSchema(schema.Volume) *Volume

	SchemaFromVolume(*Volume) schema.Volume

	NetworkFromSchema(schema.Network) *Network

	SchemaFromNetwork(*Network) schema.Network

	NetworkSubnetFromSchema(schema.NetworkSubnet) NetworkSubnet

	SchemaFromNetworkSubnet(NetworkSubnet) schema.NetworkSubnet

	NetworkRouteFromSchema(schema.NetworkRoute) NetworkRoute

	SchemaFromNetworkRoute(NetworkRoute) schema.NetworkRoute

	LoadBalancerFromSchema(schema.LoadBalancer) *LoadBalancer

	// goverter:map OutgoingTraffic | mapZeroUint64ToNil
	// goverter:map IngoingTraffic | mapZeroUint64ToNil
	SchemaFromLoadBalancer(*LoadBalancer) schema.LoadBalancer

	// goverter:map Prices Pricings
	LoadBalancerTypeFromSchema(schema.LoadBalancerType) *LoadBalancerType

	// goverter:map Pricings Prices
	SchemaFromLoadBalancerType(*LoadBalancerType) schema.LoadBalancerType

	// goverter:map PriceHourly Hourly
	// goverter:map PriceMonthly Monthly
	// goverter:map PricePerTBTraffic PerTBTraffic
	LoadBalancerTypeLocationPricingFromSchema(schema.PricingLoadBalancerTypePrice) LoadBalancerTypeLocationPricing

	// goverter:map Hourly PriceHourly
	// goverter:map Monthly PriceMonthly
	// goverter:map PerTBTraffic PricePerTBTraffic
	SchemaFromLoadBalancerTypeLocationPricing(LoadBalancerTypeLocationPricing) schema.PricingLoadBalancerTypePrice

	LoadBalancerServiceFromSchema(schema.LoadBalancerService) LoadBalancerService

	SchemaFromLoadBalancerService(LoadBalancerService) schema.LoadBalancerService

	LoadBalancerServiceHealthCheckFromSchema(*schema.LoadBalancerServiceHealthCheck) LoadBalancerServiceHealthCheck

	SchemaFromLoadBalancerServiceHealthCheck(LoadBalancerServiceHealthCheck) *schema.LoadBalancerServiceHealthCheck

	LoadBalancerTargetFromSchema(schema.LoadBalancerTarget) LoadBalancerTarget

	SchemaFromLoadBalancerTarget(LoadBalancerTarget) schema.LoadBalancerTarget

	// goverter:map ID Server
	LoadBalancerTargetServerFromSchema(schema.LoadBalancerTargetServer) LoadBalancerTargetServer

	// goverter:map Server ID
	SchemaFromLoadBalancerServerTarget(LoadBalancerTargetServer) schema.LoadBalancerTargetServer

	LoadBalancerTargetHealthStatusFromSchema(schema.LoadBalancerTargetHealthStatus) LoadBalancerTargetHealthStatus

	SchemaFromLoadBalancerTargetHealthStatus(LoadBalancerTargetHealthStatus) schema.LoadBalancerTargetHealthStatus

	CertificateFromSchema(schema.Certificate) *Certificate

	SchemaFromCertificate(*Certificate) schema.Certificate

	PaginationFromSchema(schema.MetaPagination) Pagination

	SchemaFromPagination(Pagination) schema.MetaPagination

	// goverter:ignore response
	// goverter:map Details | errorDetailsFromSchema
	ErrorFromSchema(schema.Error) Error

	// goverter:map Details | schemaFromErrorDetails
	// goverter:map Details DetailsRaw | rawSchemaFromErrorDetails
	SchemaFromError(Error) schema.Error

	// goverter:map . Image | imagePricingFromSchema
	// goverter:map . FloatingIP | floatingIPPricingFromSchema
	// goverter:map . FloatingIPs | floatingIPTypePricingFromSchema
	// goverter:map . PrimaryIPs | primaryIPPricingFromSchema
	// goverter:map . Traffic | trafficPricingFromSchema
	// goverter:map . ServerTypes | serverTypePricingFromSchema
	// goverter:map . LoadBalancerTypes | loadBalancerTypePricingFromSchema
	// goverter:map . Volume | volumePricingFromSchema
	PricingFromSchema(schema.Pricing) Pricing

	// goverter:map PriceHourly Hourly
	// goverter:map PriceMonthly Monthly
	// goverter:map PricePerTBTraffic PerTBTraffic
	serverTypePricingFromSchema(schema.PricingServerTypePrice) ServerTypeLocationPricing

	// goverter:map Image.PerGBMonth.Currency Currency
	// goverter:map Image.PerGBMonth.VATRate VATRate
	SchemaFromPricing(Pricing) schema.Pricing

	// goverter:map PerGBMonth PricePerGBMonth
	schemaFromImagePricing(ImagePricing) schema.PricingImage

	// goverter:map Monthly PriceMonthly
	schemaFromFloatingIPPricing(FloatingIPPricing) schema.PricingFloatingIP

	// goverter:map Pricings Prices
	schemaFromFloatingIPTypePricing(FloatingIPTypePricing) schema.PricingFloatingIPType

	// goverter:map Monthly PriceMonthly
	schemaFromFloatingIPTypeLocationPricing(FloatingIPTypeLocationPricing) schema.PricingFloatingIPTypePrice

	// goverter:map Pricings Prices
	schemaFromPrimaryIPPricing(PrimaryIPPricing) schema.PricingPrimaryIP

	// goverter:map Monthly PriceMonthly
	// goverter:map Hourly PriceHourly
	schemaFromPrimaryIPTypePricing(PrimaryIPTypePricing) schema.PricingPrimaryIPTypePrice

	// goverter:map PerTB PricePerTB
	schemaFromTrafficPricing(TrafficPricing) schema.PricingTraffic

	// goverter:map Pricings Prices
	// goverter:map ServerType.ID ID
	// goverter:map ServerType.Name Name
	schemaFromServerTypePricing(ServerTypePricing) schema.PricingServerType

	// goverter:map Pricings Prices
	// goverter:map LoadBalancerType.ID ID
	// goverter:map LoadBalancerType.Name Name
	schemaFromLoadBalancerTypePricing(LoadBalancerTypePricing) schema.PricingLoadBalancerType

	// goverter:map PerGBMonthly PricePerGBPerMonth
	schemaFromVolumePricing(VolumePricing) schema.PricingVolume

	// goverter:map Monthly PriceMonthly
	// goverter:map Hourly PriceHourly
	// goverter:map PerTBTraffic PricePerTBTraffic
	schemaFromServerTypeLocationPricing(ServerTypeLocationPricing) schema.PricingServerTypePrice

	FirewallFromSchema(schema.Firewall) *Firewall

	SchemaFromFirewall(*Firewall) schema.Firewall

	PlacementGroupFromSchema(schema.PlacementGroup) *PlacementGroup

	SchemaFromPlacementGroup(*PlacementGroup) schema.PlacementGroup

	SchemaFromPlacementGroupCreateOpts(PlacementGroupCreateOpts) schema.PlacementGroupCreateRequest

	SchemaFromLoadBalancerCreateOpts(LoadBalancerCreateOpts) schema.LoadBalancerCreateRequest

	// goverter:map Server.ID ID
	SchemaFromLoadBalancerCreateOptsTargetServer(LoadBalancerCreateOptsTargetServer) schema.LoadBalancerCreateRequestTargetServer

	SchemaFromLoadBalancerAddServiceOpts(LoadBalancerAddServiceOpts) schema.LoadBalancerActionAddServiceRequest

	// goverter:ignore ListenPort
	SchemaFromLoadBalancerUpdateServiceOpts(LoadBalancerUpdateServiceOpts) schema.LoadBalancerActionUpdateServiceRequest

	SchemaFromFirewallCreateOpts(FirewallCreateOpts) schema.FirewallCreateRequest

	SchemaFromFirewallSetRulesOpts(FirewallSetRulesOpts) schema.FirewallActionSetRulesRequest

	SchemaFromFirewallResource(FirewallResource) schema.FirewallResource

	// goverter:autoMap Metrics
	ServerMetricsFromSchema(*schema.ServerGetMetricsResponse) (*ServerMetrics, error)

	// goverter:autoMap Metrics
	LoadBalancerMetricsFromSchema(*schema.LoadBalancerGetMetricsResponse) (*LoadBalancerMetrics, error)

	DeprecationFromSchema(*schema.DeprecationInfo) *DeprecationInfo

	SchemaFromDeprecation(*DeprecationInfo) *schema.DeprecationInfo
}

func schemaActionErrorFromAction(a Action) *schema.ActionError {
	if a.ErrorCode != "" && a.ErrorMessage != "" {
		return &schema.ActionError{
			Code:    a.ErrorCode,
			Message: a.ErrorMessage,
		}
	}
	return nil
}

func ipFromFloatingIPSchema(s schema.FloatingIP) net.IP {
	if s.Type == string(FloatingIPTypeIPv4) {
		return net.ParseIP(s.IP)
	}
	ip, _, _ := net.ParseCIDR(s.IP)
	return ip
}

func networkFromFloatingIPSchema(s schema.FloatingIP) *net.IPNet {
	if s.Type == string(FloatingIPTypeIPv4) {
		return nil
	}
	_, n, _ := net.ParseCIDR(s.IP)
	return n
}

func ipFromPrimaryIPSchema(s schema.PrimaryIP) net.IP {
	if s.Type == string(FloatingIPTypeIPv4) {
		return net.ParseIP(s.IP)
	}
	ip, _, _ := net.ParseCIDR(s.IP)
	return ip
}

func networkFromPrimaryIPSchema(s schema.PrimaryIP) *net.IPNet {
	if s.Type == string(FloatingIPTypeIPv4) {
		return nil
	}
	_, n, _ := net.ParseCIDR(s.IP)
	return n
}

func serverFromInt64(id int64) Server {
	return Server{ID: id}
}

func int64FromServer(s Server) int64 {
	return s.ID
}

func networkFromInt64(id int64) Network {
	return Network{ID: id}
}

func int64FromNetwork(network Network) int64 {
	return network.ID
}

func loadBalancerFromInt64(id int64) LoadBalancer {
	return LoadBalancer{ID: id}
}

func int64FromLoadBalancer(lb LoadBalancer) int64 {
	return lb.ID
}

func volumeFromInt64(id int64) *Volume {
	return &Volume{ID: id}
}

func int64FromVolume(volume *Volume) int64 {
	if volume == nil {
		return 0
	}
	return volume.ID
}

func serverTypeFromInt64(id int64) *ServerType {
	return &ServerType{ID: id}
}

func int64FromServerType(s *ServerType) int64 {
	if s == nil {
		return 0
	}
	return s.ID
}

func certificateFromInt64(id int64) *Certificate {
	return &Certificate{ID: id}
}

func int64FromCertificate(c *Certificate) int64 {
	if c == nil {
		return 0
	}
	return c.ID
}

func locationFromString(s string) Location {
	return Location{Name: s}
}

func stringFromLocation(l Location) string {
	return l.Name
}

func mapFromFloatingIPDNSPtrSchema(dnsPtr []schema.FloatingIPDNSPtr) map[string]string {
	m := make(map[string]string, len(dnsPtr))
	for _, entry := range dnsPtr {
		m[entry.IP] = entry.DNSPtr
	}
	return m
}

func floatingIPDNSPtrSchemaFromMap(m map[string]string) []schema.FloatingIPDNSPtr {
	dnsPtr := make([]schema.FloatingIPDNSPtr, 0, len(m))
	for ip, ptr := range m {
		dnsPtr = append(dnsPtr, schema.FloatingIPDNSPtr{
			IP:     ip,
			DNSPtr: ptr,
		})
	}
	return dnsPtr
}

func mapFromPrimaryIPDNSPtrSchema(dnsPtr []schema.PrimaryIPDNSPTR) map[string]string {
	m := make(map[string]string, len(dnsPtr))
	for _, entry := range dnsPtr {
		m[entry.IP] = entry.DNSPtr
	}
	return m
}

func primaryIPDNSPtrSchemaFromMap(m map[string]string) []schema.PrimaryIPDNSPTR {
	dnsPtr := make([]schema.PrimaryIPDNSPTR, 0, len(m))
	for ip, ptr := range m {
		dnsPtr = append(dnsPtr, schema.PrimaryIPDNSPTR{
			IP:     ip,
			DNSPtr: ptr,
		})
	}
	return dnsPtr
}

func mapFromServerPublicNetIPv6DNSPtrSchema(dnsPtr []schema.ServerPublicNetIPv6DNSPtr) map[string]string {
	m := make(map[string]string, len(dnsPtr))
	for _, entry := range dnsPtr {
		m[entry.IP] = entry.DNSPtr
	}
	return m
}

func serverPublicNetIPv6DNSPtrSchemaFromMap(m map[string]string) []schema.ServerPublicNetIPv6DNSPtr {
	dnsPtr := make([]schema.ServerPublicNetIPv6DNSPtr, 0, len(m))
	for ip, ptr := range m {
		dnsPtr = append(dnsPtr, schema.ServerPublicNetIPv6DNSPtr{
			IP:     ip,
			DNSPtr: ptr,
		})
	}
	return dnsPtr
}

func floatingIPToIPString(ip FloatingIP) string {
	if ip.Type == FloatingIPTypeIPv4 {
		return ip.IP.String()
	}
	return ip.Network.String()
}

func primaryIPToIPString(ip PrimaryIP) string {
	if ip.Type == PrimaryIPTypeIPv4 {
		return ip.IP.String()
	}
	return ip.Network.String()
}

func floatingIPFromInt64(id int64) *FloatingIP {
	return &FloatingIP{ID: id}
}

func int64FromFloatingIP(f *FloatingIP) int64 {
	if f == nil {
		return 0
	}
	return f.ID
}

func firewallStatusFromSchemaServerFirewall(fw schema.ServerFirewall) *ServerFirewallStatus {
	return &ServerFirewallStatus{
		Firewall: Firewall{ID: fw.ID},
		Status:   FirewallStatus(fw.Status),
	}
}

func serverFirewallSchemaFromFirewallStatus(s *ServerFirewallStatus) schema.ServerFirewall {
	return schema.ServerFirewall{
		ID:     s.Firewall.ID,
		Status: string(s.Status),
	}
}

func ipFromServerPublicNetIPv6Schema(s schema.ServerPublicNetIPv6) net.IP {
	ip, _, _ := net.ParseCIDR(s.IP)
	return ip
}

func ipNetFromServerPublicNetIPv6Schema(s schema.ServerPublicNetIPv6) *net.IPNet {
	_, n, _ := net.ParseCIDR(s.IP)
	return n
}

func serverFromImageCreatedFromSchema(s schema.ImageCreatedFrom) Server {
	return Server{
		ID:   s.ID,
		Name: s.Name,
	}
}

func ipFromString(s string) net.IP {
	return net.ParseIP(s)
}

func stringFromIP(ip net.IP) string {
	if ip == nil {
		return ""
	}
	return ip.String()
}

func ipNetFromString(s string) net.IPNet {
	_, n, _ := net.ParseCIDR(s)
	if n == nil {
		return net.IPNet{}
	}
	return *n
}

func stringFromIPNet(ip net.IPNet) string {
	return ip.String()
}

func timeToTimePtr(t time.Time) *time.Time {
	// Some hcloud structs don't use pointers for nullable times, so the zero value
	// should be treated as nil.
	if t == (time.Time{}) {
		return nil
	}
	return &t
}

func durationFromIntSeconds(s int) time.Duration {
	return time.Duration(s) * time.Second
}

func intSecondsFromDuration(d time.Duration) int {
	return int(d.Seconds())
}

func errorDetailsFromSchema(d interface{}) interface{} {
	switch typed := d.(type) {
	case schema.ErrorDetailsInvalidInput:
		details := ErrorDetailsInvalidInput{
			Fields: make([]ErrorDetailsInvalidInputField, len(typed.Fields)),
		}
		for i, field := range typed.Fields {
			details.Fields[i] = ErrorDetailsInvalidInputField{
				Name:     field.Name,
				Messages: field.Messages,
			}
		}
		return details

	case schema.ErrorDetailsDeprecatedAPIEndpoint:
		return ErrorDetailsDeprecatedAPIEndpoint{
			Announcement: typed.Announcement,
		}
	}
	return nil
}

func schemaFromErrorDetails(d interface{}) interface{} {
	switch typed := d.(type) {
	case ErrorDetailsInvalidInput:
		details := schema.ErrorDetailsInvalidInput{
			Fields: make([]struct {
				Name     string   `json:"name"`
				Messages []string `json:"messages"`
			}, len(typed.Fields)),
		}
		for i, field := range typed.Fields {
			details.Fields[i] = struct {
				Name     string   `json:"name"`
				Messages []string `json:"messages"`
			}{Name: field.Name, Messages: field.Messages}
		}
		return details

	case ErrorDetailsDeprecatedAPIEndpoint:
		return schema.ErrorDetailsDeprecatedAPIEndpoint{Announcement: typed.Announcement}
	}

	return nil
}

func imagePricingFromSchema(s schema.Pricing) ImagePricing {
	return ImagePricing{
		PerGBMonth: Price{
			Net:      s.Image.PricePerGBMonth.Net,
			Gross:    s.Image.PricePerGBMonth.Gross,
			Currency: s.Currency,
			VATRate:  s.VATRate,
		},
	}
}

func floatingIPPricingFromSchema(s schema.Pricing) FloatingIPPricing {
	return FloatingIPPricing{
		Monthly: Price{
			Net:      s.FloatingIP.PriceMonthly.Net,   // nolint:staticcheck // Field is deprecated, but removal is not planned
			Gross:    s.FloatingIP.PriceMonthly.Gross, // nolint:staticcheck // Field is deprecated, but removal is not planned
			Currency: s.Currency,
			VATRate:  s.VATRate,
		},
	}
}

func floatingIPTypePricingFromSchema(s schema.Pricing) []FloatingIPTypePricing {
	p := make([]FloatingIPTypePricing, len(s.FloatingIPs))
	for i, floatingIPType := range s.FloatingIPs {
		var pricings = make([]FloatingIPTypeLocationPricing, len(floatingIPType.Prices))
		for i, price := range floatingIPType.Prices {
			pricings[i] = FloatingIPTypeLocationPricing{
				Location: &Location{Name: price.Location},
				Monthly: Price{
					Currency: s.Currency,
					VATRate:  s.VATRate,
					Net:      price.PriceMonthly.Net,
					Gross:    price.PriceMonthly.Gross,
				},
			}
		}
		p[i] = FloatingIPTypePricing{Type: FloatingIPType(floatingIPType.Type), Pricings: pricings}
	}
	return p
}

func primaryIPPricingFromSchema(s schema.Pricing) []PrimaryIPPricing {
	p := make([]PrimaryIPPricing, len(s.FloatingIPs))
	for i, primaryIPType := range s.PrimaryIPs {
		var pricings = make([]PrimaryIPTypePricing, len(primaryIPType.Prices))
		for i, price := range primaryIPType.Prices {
			pricings[i] = PrimaryIPTypePricing{
				Location: price.Location,
				Monthly: PrimaryIPPrice{
					Net:   price.PriceMonthly.Net,
					Gross: price.PriceMonthly.Gross,
				},
				Hourly: PrimaryIPPrice{
					Net:   price.PriceHourly.Net,
					Gross: price.PriceHourly.Gross,
				},
			}
		}
		p[i] = PrimaryIPPricing{Type: primaryIPType.Type, Pricings: pricings}
	}
	return p
}

func trafficPricingFromSchema(s schema.Pricing) TrafficPricing {
	return TrafficPricing{
		PerTB: Price{
			Net:      s.Traffic.PricePerTB.Net,   // nolint:staticcheck // Field is deprecated, but we still need to map it as long as it is available
			Gross:    s.Traffic.PricePerTB.Gross, // nolint:staticcheck // Field is deprecated, but we still need to map it as long as it is available
			Currency: s.Currency,
			VATRate:  s.VATRate,
		},
	}
}

func serverTypePricingFromSchema(s schema.Pricing) []ServerTypePricing {
	p := make([]ServerTypePricing, len(s.ServerTypes))
	for i, serverType := range s.ServerTypes {
		var pricings = make([]ServerTypeLocationPricing, len(serverType.Prices))
		for i, price := range serverType.Prices {
			pricings[i] = ServerTypeLocationPricing{
				Location: &Location{Name: price.Location},
				Hourly: Price{
					Currency: s.Currency,
					VATRate:  s.VATRate,
					Net:      price.PriceHourly.Net,
					Gross:    price.PriceHourly.Gross,
				},
				Monthly: Price{
					Currency: s.Currency,
					VATRate:  s.VATRate,
					Net:      price.PriceMonthly.Net,
					Gross:    price.PriceMonthly.Gross,
				},
				IncludedTraffic: price.IncludedTraffic,
				PerTBTraffic: Price{
					Currency: s.Currency,
					VATRate:  s.VATRate,
					Net:      price.PricePerTBTraffic.Net,
					Gross:    price.PricePerTBTraffic.Gross,
				},
			}
		}
		p[i] = ServerTypePricing{
			ServerType: &ServerType{
				ID:   serverType.ID,
				Name: serverType.Name,
			},
			Pricings: pricings,
		}
	}
	return p
}

func loadBalancerTypePricingFromSchema(s schema.Pricing) []LoadBalancerTypePricing {
	p := make([]LoadBalancerTypePricing, len(s.LoadBalancerTypes))
	for i, loadBalancerType := range s.LoadBalancerTypes {
		var pricings = make([]LoadBalancerTypeLocationPricing, len(loadBalancerType.Prices))
		for i, price := range loadBalancerType.Prices {
			pricings[i] = LoadBalancerTypeLocationPricing{
				Location: &Location{Name: price.Location},
				Hourly: Price{
					Currency: s.Currency,
					VATRate:  s.VATRate,
					Net:      price.PriceHourly.Net,
					Gross:    price.PriceHourly.Gross,
				},
				Monthly: Price{
					Currency: s.Currency,
					VATRate:  s.VATRate,
					Net:      price.PriceMonthly.Net,
					Gross:    price.PriceMonthly.Gross,
				},
				IncludedTraffic: price.IncludedTraffic,
				PerTBTraffic: Price{
					Currency: s.Currency,
					VATRate:  s.VATRate,
					Net:      price.PricePerTBTraffic.Net,
					Gross:    price.PricePerTBTraffic.Gross,
				},
			}
		}
		p[i] = LoadBalancerTypePricing{
			LoadBalancerType: &LoadBalancerType{
				ID:   loadBalancerType.ID,
				Name: loadBalancerType.Name,
			},
			Pricings: pricings,
		}
	}
	return p
}

func volumePricingFromSchema(s schema.Pricing) VolumePricing {
	return VolumePricing{
		PerGBMonthly: Price{
			Net:      s.Volume.PricePerGBPerMonth.Net,
			Gross:    s.Volume.PricePerGBPerMonth.Gross,
			Currency: s.Currency,
			VATRate:  s.VATRate,
		},
	}
}

func serverMetricsTimeSeriesFromSchema(s schema.ServerTimeSeriesVals) ([]ServerMetricsValue, error) {
	vals := make([]ServerMetricsValue, len(s.Values))

	for i, rawVal := range s.Values {
		var val ServerMetricsValue

		tup, ok := rawVal.([]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to convert value to tuple: %v", rawVal)
		}
		if len(tup) != 2 {
			return nil, fmt.Errorf("invalid tuple size: %d: %v", len(tup), rawVal)
		}
		ts, ok := tup[0].(float64)
		if !ok {
			return nil, fmt.Errorf("convert to float64: %v", tup[0])
		}
		val.Timestamp = ts

		v, ok := tup[1].(string)
		if !ok {
			return nil, fmt.Errorf("not a string: %v", tup[1])
		}
		val.Value = v
		vals[i] = val
	}

	return vals, nil
}

func loadBalancerMetricsTimeSeriesFromSchema(s schema.LoadBalancerTimeSeriesVals) ([]LoadBalancerMetricsValue, error) {
	vals := make([]LoadBalancerMetricsValue, len(s.Values))

	for i, rawVal := range s.Values {
		var val LoadBalancerMetricsValue

		tup, ok := rawVal.([]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to convert value to tuple: %v", rawVal)
		}
		if len(tup) != 2 {
			return nil, fmt.Errorf("invalid tuple size: %d: %v", len(tup), rawVal)
		}
		ts, ok := tup[0].(float64)
		if !ok {
			return nil, fmt.Errorf("convert to float64: %v", tup[0])
		}
		val.Timestamp = ts

		v, ok := tup[1].(string)
		if !ok {
			return nil, fmt.Errorf("not a string: %v", tup[1])
		}
		val.Value = v
		vals[i] = val
	}

	return vals, nil
}

func mapEmptyStringToNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func stringPtrFromLoadBalancerServiceProtocol(p LoadBalancerServiceProtocol) *string {
	return mapEmptyStringToNil(string(p))
}

func stringPtrFromNetworkZone(z NetworkZone) *string {
	return mapEmptyStringToNil(string(z))
}

func mapZeroInt64ToNil(i int64) *int64 {
	if i == 0 {
		return nil
	}
	return &i
}

func mapZeroUint64ToNil(i uint64) *uint64 {
	if i == 0 {
		return nil
	}
	return &i
}

func schemaFromLoadBalancerCreateOptsTargetLabelSelector(l LoadBalancerCreateOptsTargetLabelSelector) *schema.LoadBalancerCreateRequestTargetLabelSelector {
	if l.Selector == "" {
		return nil
	}
	return &schema.LoadBalancerCreateRequestTargetLabelSelector{Selector: l.Selector}
}

func schemaFromLoadBalancerCreateOptsTargetIP(l LoadBalancerCreateOptsTargetIP) *schema.LoadBalancerCreateRequestTargetIP {
	if l.IP == "" {
		return nil
	}
	return &schema.LoadBalancerCreateRequestTargetIP{IP: l.IP}
}

func schemaFromLoadBalancerCreateOptsTargetServer(l LoadBalancerCreateOptsTargetServer) *schema.LoadBalancerCreateRequestTargetServer {
	if l.Server == nil {
		return nil
	}
	return &schema.LoadBalancerCreateRequestTargetServer{ID: l.Server.ID}
}

func stringMapToStringMapPtr(m map[string]string) *map[string]string {
	if m == nil {
		return nil
	}
	return &m
}

func rawSchemaFromErrorDetails(v interface{}) json.RawMessage {
	d := schemaFromErrorDetails(v)
	if v == nil {
		return nil
	}
	msg, err := json.Marshal(d)
	if err != nil {
		return nil
	}
	return msg
}

func mapZeroFloat32ToNil(f float32) *float32 {
	if f == 0 {
		return nil
	}
	return &f
}

func isDeprecationNotNil(d *DeprecationInfo) bool {
	return d != nil
}

// int64SlicePtrFromCertificatePtrSlice is needed so that a nil slice is mapped to nil instead of &nil.
func int64SlicePtrFromCertificatePtrSlice(s []*Certificate) *[]int64 {
	if s == nil {
		return nil
	}
	var ids = make([]int64, len(s))
	for i, cert := range s {
		ids[i] = cert.ID
	}
	return &ids
}

func stringSlicePtrFromStringSlice(s []string) *[]string {
	if s == nil {
		return nil
	}
	return &s
}

func locationFromServerTypeLocationSchema(serverTypeLocation schema.ServerTypeLocation) *Location {
	return &Location{
		ID:   serverTypeLocation.ID,
		Name: serverTypeLocation.Name,
	}
}
