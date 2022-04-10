package hcloud

import (
	"context"

	"github.com/hetznercloud/hcloud-go/hcloud/schema"
)

// Pricing specifies pricing information for various resources.
type Pricing struct {
	Image             ImagePricing
	FloatingIP        FloatingIPPricing
	FloatingIPs       []FloatingIPTypePricing
	Traffic           TrafficPricing
	ServerBackup      ServerBackupPricing
	ServerTypes       []ServerTypePricing
	LoadBalancerTypes []LoadBalancerTypePricing
	Volume            VolumePricing
}

// Price represents a price. Net amount, gross amount, as well as VAT rate are
// specified as strings and it is the user's responsibility to convert them to
// appropriate types for calculations.
type Price struct {
	Currency string
	VATRate  string
	Net      string
	Gross    string
}

// ImagePricing provides pricing information for imaegs.
type ImagePricing struct {
	PerGBMonth Price
}

// FloatingIPPricing provides pricing information for Floating IPs.
type FloatingIPPricing struct {
	Monthly Price
}

// FloatingIPTypePricing provides pricing information for Floating IPs per Type.
type FloatingIPTypePricing struct {
	Type     FloatingIPType
	Pricings []FloatingIPTypeLocationPricing
}

// FloatingIPTypeLocationPricing provides pricing information for a Floating IP type
// at a location.
type FloatingIPTypeLocationPricing struct {
	Location *Location
	Monthly  Price
}

// TrafficPricing provides pricing information for traffic.
type TrafficPricing struct {
	PerTB Price
}

// VolumePricing provides pricing information for a Volume.
type VolumePricing struct {
	PerGBMonthly Price
}

// ServerBackupPricing provides pricing information for server backups.
type ServerBackupPricing struct {
	Percentage string
}

// ServerTypePricing provides pricing information for a server type.
type ServerTypePricing struct {
	ServerType *ServerType
	Pricings   []ServerTypeLocationPricing
}

// ServerTypeLocationPricing provides pricing information for a server type
// at a location.
type ServerTypeLocationPricing struct {
	Location *Location
	Hourly   Price
	Monthly  Price
}

// LoadBalancerTypePricing provides pricing information for a Load Balancer type.
type LoadBalancerTypePricing struct {
	LoadBalancerType *LoadBalancerType
	Pricings         []LoadBalancerTypeLocationPricing
}

// LoadBalancerTypeLocationPricing provides pricing information for a Load Balancer type
// at a location.
type LoadBalancerTypeLocationPricing struct {
	Location *Location
	Hourly   Price
	Monthly  Price
}

// PricingClient is a client for the pricing API.
type PricingClient struct {
	client *Client
}

// Get retrieves pricing information.
func (c *PricingClient) Get(ctx context.Context) (Pricing, *Response, error) {
	req, err := c.client.NewRequest(ctx, "GET", "/pricing", nil)
	if err != nil {
		return Pricing{}, nil, err
	}

	var body schema.PricingGetResponse
	resp, err := c.client.Do(req, &body)
	if err != nil {
		return Pricing{}, nil, err
	}
	return PricingFromSchema(body.Pricing), resp, nil
}
