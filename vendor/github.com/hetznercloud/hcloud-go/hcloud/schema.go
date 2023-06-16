package hcloud

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud/schema"
)

// This file provides converter functions to convert models in the
// schema package to models in the hcloud package and vice versa.

// ActionFromSchema converts a schema.Action to an Action.
func ActionFromSchema(s schema.Action) *Action {
	action := &Action{
		ID:        s.ID,
		Status:    ActionStatus(s.Status),
		Command:   s.Command,
		Progress:  s.Progress,
		Started:   s.Started,
		Resources: []*ActionResource{},
	}
	if s.Finished != nil {
		action.Finished = *s.Finished
	}
	if s.Error != nil {
		action.ErrorCode = s.Error.Code
		action.ErrorMessage = s.Error.Message
	}
	for _, r := range s.Resources {
		action.Resources = append(action.Resources, &ActionResource{
			ID:   r.ID,
			Type: ActionResourceType(r.Type),
		})
	}
	return action
}

// ActionsFromSchema converts a slice of schema.Action to a slice of Action.
func ActionsFromSchema(s []schema.Action) []*Action {
	actions := make([]*Action, len(s))
	for i, a := range s {
		actions[i] = ActionFromSchema(a)
	}
	return actions
}

// FloatingIPFromSchema converts a schema.FloatingIP to a FloatingIP.
func FloatingIPFromSchema(s schema.FloatingIP) *FloatingIP {
	f := &FloatingIP{
		ID:           s.ID,
		Type:         FloatingIPType(s.Type),
		HomeLocation: LocationFromSchema(s.HomeLocation),
		Created:      s.Created,
		Blocked:      s.Blocked,
		Protection: FloatingIPProtection{
			Delete: s.Protection.Delete,
		},
		Name: s.Name,
	}
	if s.Description != nil {
		f.Description = *s.Description
	}
	if s.Server != nil {
		f.Server = &Server{ID: *s.Server}
	}
	if f.Type == FloatingIPTypeIPv4 {
		f.IP = net.ParseIP(s.IP)
	} else {
		f.IP, f.Network, _ = net.ParseCIDR(s.IP)
	}
	f.DNSPtr = map[string]string{}
	for _, entry := range s.DNSPtr {
		f.DNSPtr[entry.IP] = entry.DNSPtr
	}
	f.Labels = map[string]string{}
	for key, value := range s.Labels {
		f.Labels[key] = value
	}
	return f
}

// PrimaryIPFromSchema converts a schema.PrimaryIP to a PrimaryIP.
func PrimaryIPFromSchema(s schema.PrimaryIP) *PrimaryIP {
	f := &PrimaryIP{
		ID:         s.ID,
		Type:       PrimaryIPType(s.Type),
		AutoDelete: s.AutoDelete,

		Created: s.Created,
		Blocked: s.Blocked,
		Protection: PrimaryIPProtection{
			Delete: s.Protection.Delete,
		},
		Name:         s.Name,
		AssigneeType: s.AssigneeType,
		AssigneeID:   s.AssigneeID,
		Datacenter:   DatacenterFromSchema(s.Datacenter),
	}

	if f.Type == PrimaryIPTypeIPv4 {
		f.IP = net.ParseIP(s.IP)
	} else {
		f.IP, f.Network, _ = net.ParseCIDR(s.IP)
	}
	f.DNSPtr = map[string]string{}
	for _, entry := range s.DNSPtr {
		f.DNSPtr[entry.IP] = entry.DNSPtr
	}
	f.Labels = map[string]string{}
	for key, value := range s.Labels {
		f.Labels[key] = value
	}
	return f
}

// ISOFromSchema converts a schema.ISO to an ISO.
func ISOFromSchema(s schema.ISO) *ISO {
	iso := &ISO{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Type:        ISOType(s.Type),
		Deprecated:  s.Deprecated,
	}
	if s.Architecture != nil {
		iso.Architecture = Ptr(Architecture(*s.Architecture))
	}
	return iso
}

// LocationFromSchema converts a schema.Location to a Location.
func LocationFromSchema(s schema.Location) *Location {
	return &Location{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Country:     s.Country,
		City:        s.City,
		Latitude:    s.Latitude,
		Longitude:   s.Longitude,
		NetworkZone: NetworkZone(s.NetworkZone),
	}
}

// DatacenterFromSchema converts a schema.Datacenter to a Datacenter.
func DatacenterFromSchema(s schema.Datacenter) *Datacenter {
	d := &Datacenter{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Location:    LocationFromSchema(s.Location),
		ServerTypes: DatacenterServerTypes{
			Available: []*ServerType{},
			Supported: []*ServerType{},
		},
	}
	for _, t := range s.ServerTypes.Available {
		d.ServerTypes.Available = append(d.ServerTypes.Available, &ServerType{ID: t})
	}
	for _, t := range s.ServerTypes.Supported {
		d.ServerTypes.Supported = append(d.ServerTypes.Supported, &ServerType{ID: t})
	}
	return d
}

// ServerFromSchema converts a schema.Server to a Server.
func ServerFromSchema(s schema.Server) *Server {
	server := &Server{
		ID:              s.ID,
		Name:            s.Name,
		Status:          ServerStatus(s.Status),
		Created:         s.Created,
		PublicNet:       ServerPublicNetFromSchema(s.PublicNet),
		ServerType:      ServerTypeFromSchema(s.ServerType),
		IncludedTraffic: s.IncludedTraffic,
		RescueEnabled:   s.RescueEnabled,
		Datacenter:      DatacenterFromSchema(s.Datacenter),
		Locked:          s.Locked,
		PrimaryDiskSize: s.PrimaryDiskSize,
		Protection: ServerProtection{
			Delete:  s.Protection.Delete,
			Rebuild: s.Protection.Rebuild,
		},
	}
	if s.Image != nil {
		server.Image = ImageFromSchema(*s.Image)
	}
	if s.BackupWindow != nil {
		server.BackupWindow = *s.BackupWindow
	}
	if s.OutgoingTraffic != nil {
		server.OutgoingTraffic = *s.OutgoingTraffic
	}
	if s.IngoingTraffic != nil {
		server.IngoingTraffic = *s.IngoingTraffic
	}
	if s.ISO != nil {
		server.ISO = ISOFromSchema(*s.ISO)
	}
	server.Labels = map[string]string{}
	for key, value := range s.Labels {
		server.Labels[key] = value
	}
	for _, id := range s.Volumes {
		server.Volumes = append(server.Volumes, &Volume{ID: id})
	}
	for _, privNet := range s.PrivateNet {
		server.PrivateNet = append(server.PrivateNet, ServerPrivateNetFromSchema(privNet))
	}
	if s.PlacementGroup != nil {
		server.PlacementGroup = PlacementGroupFromSchema(*s.PlacementGroup)
	}
	return server
}

// ServerPublicNetFromSchema converts a schema.ServerPublicNet to a ServerPublicNet.
func ServerPublicNetFromSchema(s schema.ServerPublicNet) ServerPublicNet {
	publicNet := ServerPublicNet{
		IPv4: ServerPublicNetIPv4FromSchema(s.IPv4),
		IPv6: ServerPublicNetIPv6FromSchema(s.IPv6),
	}
	for _, id := range s.FloatingIPs {
		publicNet.FloatingIPs = append(publicNet.FloatingIPs, &FloatingIP{ID: id})
	}
	for _, fw := range s.Firewalls {
		publicNet.Firewalls = append(publicNet.Firewalls,
			&ServerFirewallStatus{
				Firewall: Firewall{ID: fw.ID},
				Status:   FirewallStatus(fw.Status)},
		)
	}
	return publicNet
}

// ServerPublicNetIPv4FromSchema converts a schema.ServerPublicNetIPv4 to
// a ServerPublicNetIPv4.
func ServerPublicNetIPv4FromSchema(s schema.ServerPublicNetIPv4) ServerPublicNetIPv4 {
	return ServerPublicNetIPv4{
		ID:      s.ID,
		IP:      net.ParseIP(s.IP),
		Blocked: s.Blocked,
		DNSPtr:  s.DNSPtr,
	}
}

// ServerPublicNetIPv6FromSchema converts a schema.ServerPublicNetIPv6 to
// a ServerPublicNetIPv6.
func ServerPublicNetIPv6FromSchema(s schema.ServerPublicNetIPv6) ServerPublicNetIPv6 {
	ipv6 := ServerPublicNetIPv6{
		ID:      s.ID,
		Blocked: s.Blocked,
		DNSPtr:  map[string]string{},
	}
	ipv6.IP, ipv6.Network, _ = net.ParseCIDR(s.IP)

	for _, dnsPtr := range s.DNSPtr {
		ipv6.DNSPtr[dnsPtr.IP] = dnsPtr.DNSPtr
	}
	return ipv6
}

// ServerPrivateNetFromSchema converts a schema.ServerPrivateNet to a ServerPrivateNet.
func ServerPrivateNetFromSchema(s schema.ServerPrivateNet) ServerPrivateNet {
	n := ServerPrivateNet{
		Network:    &Network{ID: s.Network},
		IP:         net.ParseIP(s.IP),
		MACAddress: s.MACAddress,
	}
	for _, ip := range s.AliasIPs {
		n.Aliases = append(n.Aliases, net.ParseIP(ip))
	}
	return n
}

// ServerTypeFromSchema converts a schema.ServerType to a ServerType.
func ServerTypeFromSchema(s schema.ServerType) *ServerType {
	st := &ServerType{
		ID:              s.ID,
		Name:            s.Name,
		Description:     s.Description,
		Cores:           s.Cores,
		Memory:          s.Memory,
		Disk:            s.Disk,
		StorageType:     StorageType(s.StorageType),
		CPUType:         CPUType(s.CPUType),
		Architecture:    Architecture(s.Architecture),
		IncludedTraffic: s.IncludedTraffic,
		DeprecatableResource: DeprecatableResource{
			DeprecationFromSchema(s.Deprecation),
		},
	}
	for _, price := range s.Prices {
		st.Pricings = append(st.Pricings, ServerTypeLocationPricing{
			Location: &Location{Name: price.Location},
			Hourly: Price{
				Net:   price.PriceHourly.Net,
				Gross: price.PriceHourly.Gross,
			},
			Monthly: Price{
				Net:   price.PriceMonthly.Net,
				Gross: price.PriceMonthly.Gross,
			},
		})
	}

	return st
}

// SSHKeyFromSchema converts a schema.SSHKey to a SSHKey.
func SSHKeyFromSchema(s schema.SSHKey) *SSHKey {
	sshKey := &SSHKey{
		ID:          s.ID,
		Name:        s.Name,
		Fingerprint: s.Fingerprint,
		PublicKey:   s.PublicKey,
		Created:     s.Created,
	}
	sshKey.Labels = map[string]string{}
	for key, value := range s.Labels {
		sshKey.Labels[key] = value
	}
	return sshKey
}

// ImageFromSchema converts a schema.Image to an Image.
func ImageFromSchema(s schema.Image) *Image {
	i := &Image{
		ID:           s.ID,
		Type:         ImageType(s.Type),
		Status:       ImageStatus(s.Status),
		Description:  s.Description,
		DiskSize:     s.DiskSize,
		Created:      s.Created,
		RapidDeploy:  s.RapidDeploy,
		OSFlavor:     s.OSFlavor,
		Architecture: Architecture(s.Architecture),
		Protection: ImageProtection{
			Delete: s.Protection.Delete,
		},
		Deprecated: s.Deprecated,
		Deleted:    s.Deleted,
	}
	if s.Name != nil {
		i.Name = *s.Name
	}
	if s.ImageSize != nil {
		i.ImageSize = *s.ImageSize
	}
	if s.OSVersion != nil {
		i.OSVersion = *s.OSVersion
	}
	if s.CreatedFrom != nil {
		i.CreatedFrom = &Server{
			ID:   s.CreatedFrom.ID,
			Name: s.CreatedFrom.Name,
		}
	}
	if s.BoundTo != nil {
		i.BoundTo = &Server{
			ID: *s.BoundTo,
		}
	}
	i.Labels = map[string]string{}
	for key, value := range s.Labels {
		i.Labels[key] = value
	}
	return i
}

// VolumeFromSchema converts a schema.Volume to a Volume.
func VolumeFromSchema(s schema.Volume) *Volume {
	v := &Volume{
		ID:          s.ID,
		Name:        s.Name,
		Location:    LocationFromSchema(s.Location),
		Size:        s.Size,
		Status:      VolumeStatus(s.Status),
		LinuxDevice: s.LinuxDevice,
		Protection: VolumeProtection{
			Delete: s.Protection.Delete,
		},
		Created: s.Created,
	}
	if s.Server != nil {
		v.Server = &Server{ID: *s.Server}
	}
	v.Labels = map[string]string{}
	for key, value := range s.Labels {
		v.Labels[key] = value
	}
	return v
}

// NetworkFromSchema converts a schema.Network to a Network.
func NetworkFromSchema(s schema.Network) *Network {
	n := &Network{
		ID:      s.ID,
		Name:    s.Name,
		Created: s.Created,
		Protection: NetworkProtection{
			Delete: s.Protection.Delete,
		},
		Labels: map[string]string{},
	}

	_, n.IPRange, _ = net.ParseCIDR(s.IPRange)

	for _, subnet := range s.Subnets {
		n.Subnets = append(n.Subnets, NetworkSubnetFromSchema(subnet))
	}
	for _, route := range s.Routes {
		n.Routes = append(n.Routes, NetworkRouteFromSchema(route))
	}
	for _, serverID := range s.Servers {
		n.Servers = append(n.Servers, &Server{ID: serverID})
	}
	for key, value := range s.Labels {
		n.Labels[key] = value
	}

	return n
}

// NetworkSubnetFromSchema converts a schema.NetworkSubnet to a NetworkSubnet.
func NetworkSubnetFromSchema(s schema.NetworkSubnet) NetworkSubnet {
	sn := NetworkSubnet{
		Type:        NetworkSubnetType(s.Type),
		NetworkZone: NetworkZone(s.NetworkZone),
		Gateway:     net.ParseIP(s.Gateway),
		VSwitchID:   s.VSwitchID,
	}
	_, sn.IPRange, _ = net.ParseCIDR(s.IPRange)
	return sn
}

// NetworkRouteFromSchema converts a schema.NetworkRoute to a NetworkRoute.
func NetworkRouteFromSchema(s schema.NetworkRoute) NetworkRoute {
	r := NetworkRoute{
		Gateway: net.ParseIP(s.Gateway),
	}
	_, r.Destination, _ = net.ParseCIDR(s.Destination)
	return r
}

// LoadBalancerTypeFromSchema converts a schema.LoadBalancerType to a LoadBalancerType.
func LoadBalancerTypeFromSchema(s schema.LoadBalancerType) *LoadBalancerType {
	lt := &LoadBalancerType{
		ID:                      s.ID,
		Name:                    s.Name,
		Description:             s.Description,
		MaxConnections:          s.MaxConnections,
		MaxServices:             s.MaxServices,
		MaxTargets:              s.MaxTargets,
		MaxAssignedCertificates: s.MaxAssignedCertificates,
	}
	for _, price := range s.Prices {
		lt.Pricings = append(lt.Pricings, LoadBalancerTypeLocationPricing{
			Location: &Location{Name: price.Location},
			Hourly: Price{
				Net:   price.PriceHourly.Net,
				Gross: price.PriceHourly.Gross,
			},
			Monthly: Price{
				Net:   price.PriceMonthly.Net,
				Gross: price.PriceMonthly.Gross,
			},
		})
	}
	return lt
}

// LoadBalancerFromSchema converts a schema.LoadBalancer to a LoadBalancer.
func LoadBalancerFromSchema(s schema.LoadBalancer) *LoadBalancer {
	l := &LoadBalancer{
		ID:   s.ID,
		Name: s.Name,
		PublicNet: LoadBalancerPublicNet{
			Enabled: s.PublicNet.Enabled,
			IPv4: LoadBalancerPublicNetIPv4{
				IP:     net.ParseIP(s.PublicNet.IPv4.IP),
				DNSPtr: s.PublicNet.IPv4.DNSPtr,
			},
			IPv6: LoadBalancerPublicNetIPv6{
				IP:     net.ParseIP(s.PublicNet.IPv6.IP),
				DNSPtr: s.PublicNet.IPv6.DNSPtr,
			},
		},
		Location:         LocationFromSchema(s.Location),
		LoadBalancerType: LoadBalancerTypeFromSchema(s.LoadBalancerType),
		Algorithm:        LoadBalancerAlgorithm{Type: LoadBalancerAlgorithmType(s.Algorithm.Type)},
		Protection: LoadBalancerProtection{
			Delete: s.Protection.Delete,
		},
		Labels:          map[string]string{},
		Created:         s.Created,
		IncludedTraffic: s.IncludedTraffic,
	}
	for _, privateNet := range s.PrivateNet {
		l.PrivateNet = append(l.PrivateNet, LoadBalancerPrivateNet{
			Network: &Network{ID: privateNet.Network},
			IP:      net.ParseIP(privateNet.IP),
		})
	}
	if s.OutgoingTraffic != nil {
		l.OutgoingTraffic = *s.OutgoingTraffic
	}
	if s.IngoingTraffic != nil {
		l.IngoingTraffic = *s.IngoingTraffic
	}
	for _, service := range s.Services {
		l.Services = append(l.Services, LoadBalancerServiceFromSchema(service))
	}
	for _, target := range s.Targets {
		l.Targets = append(l.Targets, LoadBalancerTargetFromSchema(target))
	}
	for key, value := range s.Labels {
		l.Labels[key] = value
	}
	return l
}

// LoadBalancerServiceFromSchema converts a schema.LoadBalancerService to a LoadBalancerService.
func LoadBalancerServiceFromSchema(s schema.LoadBalancerService) LoadBalancerService {
	ls := LoadBalancerService{
		Protocol:        LoadBalancerServiceProtocol(s.Protocol),
		ListenPort:      s.ListenPort,
		DestinationPort: s.DestinationPort,
		Proxyprotocol:   s.Proxyprotocol,
		HealthCheck:     LoadBalancerServiceHealthCheckFromSchema(s.HealthCheck),
	}
	if s.HTTP != nil {
		ls.HTTP = LoadBalancerServiceHTTP{
			CookieName:     s.HTTP.CookieName,
			CookieLifetime: time.Duration(s.HTTP.CookieLifetime) * time.Second,
			RedirectHTTP:   s.HTTP.RedirectHTTP,
			StickySessions: s.HTTP.StickySessions,
		}
		for _, certificateID := range s.HTTP.Certificates {
			ls.HTTP.Certificates = append(ls.HTTP.Certificates, &Certificate{ID: certificateID})
		}
	}
	return ls
}

// LoadBalancerServiceHealthCheckFromSchema converts a schema.LoadBalancerServiceHealthCheck to a LoadBalancerServiceHealthCheck.
func LoadBalancerServiceHealthCheckFromSchema(s *schema.LoadBalancerServiceHealthCheck) LoadBalancerServiceHealthCheck {
	lsh := LoadBalancerServiceHealthCheck{
		Protocol: LoadBalancerServiceProtocol(s.Protocol),
		Port:     s.Port,
		Interval: time.Duration(s.Interval) * time.Second,
		Retries:  s.Retries,
		Timeout:  time.Duration(s.Timeout) * time.Second,
	}
	if s.HTTP != nil {
		lsh.HTTP = &LoadBalancerServiceHealthCheckHTTP{
			Domain:      s.HTTP.Domain,
			Path:        s.HTTP.Path,
			Response:    s.HTTP.Response,
			StatusCodes: s.HTTP.StatusCodes,
			TLS:         s.HTTP.TLS,
		}
	}
	return lsh
}

// LoadBalancerTargetFromSchema converts a schema.LoadBalancerTarget to a LoadBalancerTarget.
func LoadBalancerTargetFromSchema(s schema.LoadBalancerTarget) LoadBalancerTarget {
	lt := LoadBalancerTarget{
		Type:         LoadBalancerTargetType(s.Type),
		UsePrivateIP: s.UsePrivateIP,
	}
	if s.Server != nil {
		lt.Server = &LoadBalancerTargetServer{
			Server: &Server{ID: s.Server.ID},
		}
	}
	if s.LabelSelector != nil {
		lt.LabelSelector = &LoadBalancerTargetLabelSelector{
			Selector: s.LabelSelector.Selector,
		}
	}
	if s.IP != nil {
		lt.IP = &LoadBalancerTargetIP{IP: s.IP.IP}
	}

	for _, healthStatus := range s.HealthStatus {
		lt.HealthStatus = append(lt.HealthStatus, LoadBalancerTargetHealthStatusFromSchema(healthStatus))
	}
	for _, target := range s.Targets {
		lt.Targets = append(lt.Targets, LoadBalancerTargetFromSchema(target))
	}
	return lt
}

// LoadBalancerTargetHealthStatusFromSchema converts a schema.LoadBalancerTarget to a LoadBalancerTarget.
func LoadBalancerTargetHealthStatusFromSchema(s schema.LoadBalancerTargetHealthStatus) LoadBalancerTargetHealthStatus {
	return LoadBalancerTargetHealthStatus{
		ListenPort: s.ListenPort,
		Status:     LoadBalancerTargetHealthStatusStatus(s.Status),
	}
}

// CertificateFromSchema converts a schema.Certificate to a Certificate.
func CertificateFromSchema(s schema.Certificate) *Certificate {
	c := &Certificate{
		ID:             s.ID,
		Name:           s.Name,
		Type:           CertificateType(s.Type),
		Certificate:    s.Certificate,
		Created:        s.Created,
		NotValidBefore: s.NotValidBefore,
		NotValidAfter:  s.NotValidAfter,
		DomainNames:    s.DomainNames,
		Fingerprint:    s.Fingerprint,
	}
	if s.Status != nil {
		c.Status = &CertificateStatus{
			Issuance: CertificateStatusType(s.Status.Issuance),
			Renewal:  CertificateStatusType(s.Status.Renewal),
		}
		if s.Status.Error != nil {
			certErr := ErrorFromSchema(*s.Status.Error)
			c.Status.Error = &certErr
		}
	}
	if len(s.Labels) > 0 {
		c.Labels = s.Labels
	}
	if len(s.UsedBy) > 0 {
		c.UsedBy = make([]CertificateUsedByRef, len(s.UsedBy))
		for i, ref := range s.UsedBy {
			c.UsedBy[i] = CertificateUsedByRef{ID: ref.ID, Type: CertificateUsedByRefType(ref.Type)}
		}
	}

	return c
}

// PaginationFromSchema converts a schema.MetaPagination to a Pagination.
func PaginationFromSchema(s schema.MetaPagination) Pagination {
	return Pagination{
		Page:         s.Page,
		PerPage:      s.PerPage,
		PreviousPage: s.PreviousPage,
		NextPage:     s.NextPage,
		LastPage:     s.LastPage,
		TotalEntries: s.TotalEntries,
	}
}

// ErrorFromSchema converts a schema.Error to an Error.
func ErrorFromSchema(s schema.Error) Error {
	e := Error{
		Code:    ErrorCode(s.Code),
		Message: s.Message,
	}

	if d, ok := s.Details.(schema.ErrorDetailsInvalidInput); ok {
		details := ErrorDetailsInvalidInput{
			Fields: []ErrorDetailsInvalidInputField{},
		}
		for _, field := range d.Fields {
			details.Fields = append(details.Fields, ErrorDetailsInvalidInputField{
				Name:     field.Name,
				Messages: field.Messages,
			})
		}
		e.Details = details
	}
	return e
}

// PricingFromSchema converts a schema.Pricing to a Pricing.
func PricingFromSchema(s schema.Pricing) Pricing {
	p := Pricing{
		Image: ImagePricing{
			PerGBMonth: Price{
				Currency: s.Currency,
				VATRate:  s.VATRate,
				Net:      s.Image.PricePerGBMonth.Net,
				Gross:    s.Image.PricePerGBMonth.Gross,
			},
		},
		FloatingIP: FloatingIPPricing{
			Monthly: Price{
				Currency: s.Currency,
				VATRate:  s.VATRate,
				Net:      s.FloatingIP.PriceMonthly.Net,
				Gross:    s.FloatingIP.PriceMonthly.Gross,
			},
		},
		Traffic: TrafficPricing{
			PerTB: Price{
				Currency: s.Currency,
				VATRate:  s.VATRate,
				Net:      s.Traffic.PricePerTB.Net,
				Gross:    s.Traffic.PricePerTB.Gross,
			},
		},
		ServerBackup: ServerBackupPricing{
			Percentage: s.ServerBackup.Percentage,
		},
		Volume: VolumePricing{
			PerGBMonthly: Price{
				Currency: s.Currency,
				VATRate:  s.VATRate,
				Net:      s.Volume.PricePerGBPerMonth.Net,
				Gross:    s.Volume.PricePerGBPerMonth.Gross,
			},
		},
	}
	for _, floatingIPType := range s.FloatingIPs {
		var pricings []FloatingIPTypeLocationPricing
		for _, price := range floatingIPType.Prices {
			p := FloatingIPTypeLocationPricing{
				Location: &Location{Name: price.Location},
				Monthly: Price{
					Currency: s.Currency,
					VATRate:  s.VATRate,
					Net:      price.PriceMonthly.Net,
					Gross:    price.PriceMonthly.Gross,
				},
			}
			pricings = append(pricings, p)
		}
		p.FloatingIPs = append(p.FloatingIPs, FloatingIPTypePricing{Type: FloatingIPType(floatingIPType.Type), Pricings: pricings})
	}
	for _, primaryIPType := range s.PrimaryIPs {
		var pricings []PrimaryIPTypePricing
		for _, price := range primaryIPType.Prices {
			p := PrimaryIPTypePricing{
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
			pricings = append(pricings, p)
		}
		p.PrimaryIPs = append(p.PrimaryIPs, PrimaryIPPricing{Type: primaryIPType.Type, Pricings: pricings})
	}
	for _, serverType := range s.ServerTypes {
		var pricings []ServerTypeLocationPricing
		for _, price := range serverType.Prices {
			pricings = append(pricings, ServerTypeLocationPricing{
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
			})
		}
		p.ServerTypes = append(p.ServerTypes, ServerTypePricing{
			ServerType: &ServerType{
				ID:   serverType.ID,
				Name: serverType.Name,
			},
			Pricings: pricings,
		})
	}
	for _, loadBalancerType := range s.LoadBalancerTypes {
		var pricings []LoadBalancerTypeLocationPricing
		for _, price := range loadBalancerType.Prices {
			pricings = append(pricings, LoadBalancerTypeLocationPricing{
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
			})
		}
		p.LoadBalancerTypes = append(p.LoadBalancerTypes, LoadBalancerTypePricing{
			LoadBalancerType: &LoadBalancerType{
				ID:   loadBalancerType.ID,
				Name: loadBalancerType.Name,
			},
			Pricings: pricings,
		})
	}
	return p
}

// FirewallFromSchema converts a schema.Firewall to a Firewall.
func FirewallFromSchema(s schema.Firewall) *Firewall {
	f := &Firewall{
		ID:      s.ID,
		Name:    s.Name,
		Labels:  map[string]string{},
		Created: s.Created,
	}
	for key, value := range s.Labels {
		f.Labels[key] = value
	}
	for _, res := range s.AppliedTo {
		r := FirewallResource{Type: FirewallResourceType(res.Type)}
		switch r.Type {
		case FirewallResourceTypeLabelSelector:
			r.LabelSelector = &FirewallResourceLabelSelector{Selector: res.LabelSelector.Selector}
		case FirewallResourceTypeServer:
			r.Server = &FirewallResourceServer{ID: res.Server.ID}
		}
		f.AppliedTo = append(f.AppliedTo, r)
	}
	for _, rule := range s.Rules {
		sourceIPs := []net.IPNet{}
		for _, sourceIP := range rule.SourceIPs {
			_, mask, err := net.ParseCIDR(sourceIP)
			if err == nil && mask != nil {
				sourceIPs = append(sourceIPs, *mask)
			}
		}
		destinationIPs := []net.IPNet{}
		for _, destinationIP := range rule.DestinationIPs {
			_, mask, err := net.ParseCIDR(destinationIP)
			if err == nil && mask != nil {
				destinationIPs = append(destinationIPs, *mask)
			}
		}
		f.Rules = append(f.Rules, FirewallRule{
			Direction:      FirewallRuleDirection(rule.Direction),
			SourceIPs:      sourceIPs,
			DestinationIPs: destinationIPs,
			Protocol:       FirewallRuleProtocol(rule.Protocol),
			Port:           rule.Port,
			Description:    rule.Description,
		})
	}
	return f
}

// PlacementGroupFromSchema converts a schema.PlacementGroup to a PlacementGroup.
func PlacementGroupFromSchema(s schema.PlacementGroup) *PlacementGroup {
	g := &PlacementGroup{
		ID:      s.ID,
		Name:    s.Name,
		Labels:  s.Labels,
		Created: s.Created,
		Servers: s.Servers,
		Type:    PlacementGroupType(s.Type),
	}
	return g
}

func placementGroupCreateOptsToSchema(opts PlacementGroupCreateOpts) schema.PlacementGroupCreateRequest {
	req := schema.PlacementGroupCreateRequest{
		Name: opts.Name,
		Type: string(opts.Type),
	}
	if opts.Labels != nil {
		req.Labels = &opts.Labels
	}
	return req
}

func loadBalancerCreateOptsToSchema(opts LoadBalancerCreateOpts) schema.LoadBalancerCreateRequest {
	req := schema.LoadBalancerCreateRequest{
		Name:            opts.Name,
		PublicInterface: opts.PublicInterface,
	}
	if opts.Algorithm != nil {
		req.Algorithm = &schema.LoadBalancerCreateRequestAlgorithm{
			Type: string(opts.Algorithm.Type),
		}
	}
	if opts.LoadBalancerType.ID != 0 {
		req.LoadBalancerType = opts.LoadBalancerType.ID
	} else if opts.LoadBalancerType.Name != "" {
		req.LoadBalancerType = opts.LoadBalancerType.Name
	}
	if opts.Location != nil {
		if opts.Location.ID != 0 {
			req.Location = Ptr(strconv.Itoa(opts.Location.ID))
		} else {
			req.Location = Ptr(opts.Location.Name)
		}
	}
	if opts.NetworkZone != "" {
		req.NetworkZone = Ptr(string(opts.NetworkZone))
	}
	if opts.Labels != nil {
		req.Labels = &opts.Labels
	}
	if opts.Network != nil {
		req.Network = Ptr(opts.Network.ID)
	}
	for _, target := range opts.Targets {
		schemaTarget := schema.LoadBalancerCreateRequestTarget{
			UsePrivateIP: target.UsePrivateIP,
		}
		switch target.Type {
		case LoadBalancerTargetTypeServer:
			schemaTarget.Type = string(LoadBalancerTargetTypeServer)
			schemaTarget.Server = &schema.LoadBalancerCreateRequestTargetServer{ID: target.Server.Server.ID}
		case LoadBalancerTargetTypeLabelSelector:
			schemaTarget.Type = string(LoadBalancerTargetTypeLabelSelector)
			schemaTarget.LabelSelector = &schema.LoadBalancerCreateRequestTargetLabelSelector{Selector: target.LabelSelector.Selector}
		case LoadBalancerTargetTypeIP:
			schemaTarget.Type = string(LoadBalancerTargetTypeIP)
			schemaTarget.IP = &schema.LoadBalancerCreateRequestTargetIP{IP: target.IP.IP}
		}
		req.Targets = append(req.Targets, schemaTarget)
	}
	for _, service := range opts.Services {
		schemaService := schema.LoadBalancerCreateRequestService{
			Protocol:        string(service.Protocol),
			ListenPort:      service.ListenPort,
			DestinationPort: service.DestinationPort,
			Proxyprotocol:   service.Proxyprotocol,
		}
		if service.HTTP != nil {
			schemaService.HTTP = &schema.LoadBalancerCreateRequestServiceHTTP{
				RedirectHTTP:   service.HTTP.RedirectHTTP,
				StickySessions: service.HTTP.StickySessions,
				CookieName:     service.HTTP.CookieName,
			}
			if service.HTTP.CookieLifetime != nil {
				if sec := service.HTTP.CookieLifetime.Seconds(); sec != 0 {
					schemaService.HTTP.CookieLifetime = Ptr(int(sec))
				}
			}
			if service.HTTP.Certificates != nil {
				certificates := []int{}
				for _, certificate := range service.HTTP.Certificates {
					certificates = append(certificates, certificate.ID)
				}
				schemaService.HTTP.Certificates = &certificates
			}
		}
		if service.HealthCheck != nil {
			schemaHealthCheck := &schema.LoadBalancerCreateRequestServiceHealthCheck{
				Protocol: string(service.HealthCheck.Protocol),
				Port:     service.HealthCheck.Port,
				Retries:  service.HealthCheck.Retries,
			}
			if service.HealthCheck.Interval != nil {
				schemaHealthCheck.Interval = Ptr(int(service.HealthCheck.Interval.Seconds()))
			}
			if service.HealthCheck.Timeout != nil {
				schemaHealthCheck.Timeout = Ptr(int(service.HealthCheck.Timeout.Seconds()))
			}
			if service.HealthCheck.HTTP != nil {
				schemaHealthCheckHTTP := &schema.LoadBalancerCreateRequestServiceHealthCheckHTTP{
					Domain:   service.HealthCheck.HTTP.Domain,
					Path:     service.HealthCheck.HTTP.Path,
					Response: service.HealthCheck.HTTP.Response,
					TLS:      service.HealthCheck.HTTP.TLS,
				}
				if service.HealthCheck.HTTP.StatusCodes != nil {
					schemaHealthCheckHTTP.StatusCodes = &service.HealthCheck.HTTP.StatusCodes
				}
				schemaHealthCheck.HTTP = schemaHealthCheckHTTP
			}
			schemaService.HealthCheck = schemaHealthCheck
		}
		req.Services = append(req.Services, schemaService)
	}
	return req
}

func loadBalancerAddServiceOptsToSchema(opts LoadBalancerAddServiceOpts) schema.LoadBalancerActionAddServiceRequest {
	req := schema.LoadBalancerActionAddServiceRequest{
		Protocol:        string(opts.Protocol),
		ListenPort:      opts.ListenPort,
		DestinationPort: opts.DestinationPort,
		Proxyprotocol:   opts.Proxyprotocol,
	}
	if opts.HTTP != nil {
		req.HTTP = &schema.LoadBalancerActionAddServiceRequestHTTP{
			CookieName:     opts.HTTP.CookieName,
			RedirectHTTP:   opts.HTTP.RedirectHTTP,
			StickySessions: opts.HTTP.StickySessions,
		}
		if opts.HTTP.CookieLifetime != nil {
			req.HTTP.CookieLifetime = Ptr(int(opts.HTTP.CookieLifetime.Seconds()))
		}
		if opts.HTTP.Certificates != nil {
			certificates := []int{}
			for _, certificate := range opts.HTTP.Certificates {
				certificates = append(certificates, certificate.ID)
			}
			req.HTTP.Certificates = &certificates
		}
	}
	if opts.HealthCheck != nil {
		req.HealthCheck = &schema.LoadBalancerActionAddServiceRequestHealthCheck{
			Protocol: string(opts.HealthCheck.Protocol),
			Port:     opts.HealthCheck.Port,
			Retries:  opts.HealthCheck.Retries,
		}
		if opts.HealthCheck.Interval != nil {
			req.HealthCheck.Interval = Ptr(int(opts.HealthCheck.Interval.Seconds()))
		}
		if opts.HealthCheck.Timeout != nil {
			req.HealthCheck.Timeout = Ptr(int(opts.HealthCheck.Timeout.Seconds()))
		}
		if opts.HealthCheck.HTTP != nil {
			req.HealthCheck.HTTP = &schema.LoadBalancerActionAddServiceRequestHealthCheckHTTP{
				Domain:   opts.HealthCheck.HTTP.Domain,
				Path:     opts.HealthCheck.HTTP.Path,
				Response: opts.HealthCheck.HTTP.Response,
				TLS:      opts.HealthCheck.HTTP.TLS,
			}
			if opts.HealthCheck.HTTP.StatusCodes != nil {
				req.HealthCheck.HTTP.StatusCodes = &opts.HealthCheck.HTTP.StatusCodes
			}
		}
	}
	return req
}

func loadBalancerUpdateServiceOptsToSchema(opts LoadBalancerUpdateServiceOpts) schema.LoadBalancerActionUpdateServiceRequest {
	req := schema.LoadBalancerActionUpdateServiceRequest{
		DestinationPort: opts.DestinationPort,
		Proxyprotocol:   opts.Proxyprotocol,
	}
	if opts.Protocol != "" {
		req.Protocol = Ptr(string(opts.Protocol))
	}
	if opts.HTTP != nil {
		req.HTTP = &schema.LoadBalancerActionUpdateServiceRequestHTTP{
			CookieName:     opts.HTTP.CookieName,
			RedirectHTTP:   opts.HTTP.RedirectHTTP,
			StickySessions: opts.HTTP.StickySessions,
		}
		if opts.HTTP.CookieLifetime != nil {
			req.HTTP.CookieLifetime = Ptr(int(opts.HTTP.CookieLifetime.Seconds()))
		}
		if opts.HTTP.Certificates != nil {
			certificates := []int{}
			for _, certificate := range opts.HTTP.Certificates {
				certificates = append(certificates, certificate.ID)
			}
			req.HTTP.Certificates = &certificates
		}
	}
	if opts.HealthCheck != nil {
		req.HealthCheck = &schema.LoadBalancerActionUpdateServiceRequestHealthCheck{
			Port:    opts.HealthCheck.Port,
			Retries: opts.HealthCheck.Retries,
		}
		if opts.HealthCheck.Interval != nil {
			req.HealthCheck.Interval = Ptr(int(opts.HealthCheck.Interval.Seconds()))
		}
		if opts.HealthCheck.Timeout != nil {
			req.HealthCheck.Timeout = Ptr(int(opts.HealthCheck.Timeout.Seconds()))
		}
		if opts.HealthCheck.Protocol != "" {
			req.HealthCheck.Protocol = Ptr(string(opts.HealthCheck.Protocol))
		}
		if opts.HealthCheck.HTTP != nil {
			req.HealthCheck.HTTP = &schema.LoadBalancerActionUpdateServiceRequestHealthCheckHTTP{
				Domain:   opts.HealthCheck.HTTP.Domain,
				Path:     opts.HealthCheck.HTTP.Path,
				Response: opts.HealthCheck.HTTP.Response,
				TLS:      opts.HealthCheck.HTTP.TLS,
			}
			if opts.HealthCheck.HTTP.StatusCodes != nil {
				req.HealthCheck.HTTP.StatusCodes = &opts.HealthCheck.HTTP.StatusCodes
			}
		}
	}
	return req
}

func firewallCreateOptsToSchema(opts FirewallCreateOpts) schema.FirewallCreateRequest {
	req := schema.FirewallCreateRequest{
		Name: opts.Name,
	}
	if opts.Labels != nil {
		req.Labels = &opts.Labels
	}
	for _, rule := range opts.Rules {
		schemaRule := schema.FirewallRule{
			Direction:   string(rule.Direction),
			Protocol:    string(rule.Protocol),
			Port:        rule.Port,
			Description: rule.Description,
		}
		switch rule.Direction {
		case FirewallRuleDirectionOut:
			schemaRule.DestinationIPs = make([]string, len(rule.DestinationIPs))
			for i, destinationIP := range rule.DestinationIPs {
				schemaRule.DestinationIPs[i] = destinationIP.String()
			}
		case FirewallRuleDirectionIn:
			schemaRule.SourceIPs = make([]string, len(rule.SourceIPs))
			for i, sourceIP := range rule.SourceIPs {
				schemaRule.SourceIPs[i] = sourceIP.String()
			}
		}
		req.Rules = append(req.Rules, schemaRule)
	}
	for _, res := range opts.ApplyTo {
		schemaFirewallResource := schema.FirewallResource{
			Type: string(res.Type),
		}
		switch res.Type {
		case FirewallResourceTypeServer:
			schemaFirewallResource.Server = &schema.FirewallResourceServer{
				ID: res.Server.ID,
			}
		case FirewallResourceTypeLabelSelector:
			schemaFirewallResource.LabelSelector = &schema.FirewallResourceLabelSelector{Selector: res.LabelSelector.Selector}
		}

		req.ApplyTo = append(req.ApplyTo, schemaFirewallResource)
	}
	return req
}

func firewallSetRulesOptsToSchema(opts FirewallSetRulesOpts) schema.FirewallActionSetRulesRequest {
	req := schema.FirewallActionSetRulesRequest{Rules: []schema.FirewallRule{}}
	for _, rule := range opts.Rules {
		schemaRule := schema.FirewallRule{
			Direction:   string(rule.Direction),
			Protocol:    string(rule.Protocol),
			Port:        rule.Port,
			Description: rule.Description,
		}
		switch rule.Direction {
		case FirewallRuleDirectionOut:
			schemaRule.DestinationIPs = make([]string, len(rule.DestinationIPs))
			for i, destinationIP := range rule.DestinationIPs {
				schemaRule.DestinationIPs[i] = destinationIP.String()
			}
		case FirewallRuleDirectionIn:
			schemaRule.SourceIPs = make([]string, len(rule.SourceIPs))
			for i, sourceIP := range rule.SourceIPs {
				schemaRule.SourceIPs[i] = sourceIP.String()
			}
		}
		req.Rules = append(req.Rules, schemaRule)
	}
	return req
}

func firewallResourceToSchema(resource FirewallResource) schema.FirewallResource {
	s := schema.FirewallResource{
		Type: string(resource.Type),
	}
	switch resource.Type {
	case FirewallResourceTypeLabelSelector:
		s.LabelSelector = &schema.FirewallResourceLabelSelector{Selector: resource.LabelSelector.Selector}
	case FirewallResourceTypeServer:
		s.Server = &schema.FirewallResourceServer{ID: resource.Server.ID}
	}
	return s
}

func serverMetricsFromSchema(s *schema.ServerGetMetricsResponse) (*ServerMetrics, error) {
	ms := ServerMetrics{
		Start: s.Metrics.Start,
		End:   s.Metrics.End,
		Step:  s.Metrics.Step,
	}

	timeSeries := make(map[string][]ServerMetricsValue)
	for tsName, v := range s.Metrics.TimeSeries {
		vals := make([]ServerMetricsValue, len(v.Values))

		for i, rawVal := range v.Values {
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

		timeSeries[tsName] = vals
	}
	ms.TimeSeries = timeSeries

	return &ms, nil
}

func loadBalancerMetricsFromSchema(s *schema.LoadBalancerGetMetricsResponse) (*LoadBalancerMetrics, error) {
	ms := LoadBalancerMetrics{
		Start: s.Metrics.Start,
		End:   s.Metrics.End,
		Step:  s.Metrics.Step,
	}

	timeSeries := make(map[string][]LoadBalancerMetricsValue)
	for tsName, v := range s.Metrics.TimeSeries {
		vals := make([]LoadBalancerMetricsValue, len(v.Values))

		for i, rawVal := range v.Values {
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

		timeSeries[tsName] = vals
	}
	ms.TimeSeries = timeSeries

	return &ms, nil
}

// DeprecationFromSchema converts a [schema.DeprecationInfo] to a [DeprecationInfo].
func DeprecationFromSchema(s *schema.DeprecationInfo) *DeprecationInfo {
	if s == nil {
		return nil
	}

	return &DeprecationInfo{
		Announced:        s.Announced,
		UnavailableAfter: s.UnavailableAfter,
	}
}
