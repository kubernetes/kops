package schema

// Pricing defines the schema for pricing information.
type Pricing struct {
	Currency          string                    `json:"currency"`
	VATRate           string                    `json:"vat_rate"`
	Image             PricingImage              `json:"image"`
	FloatingIP        PricingFloatingIP         `json:"floating_ip"`
	FloatingIPs       []PricingFloatingIPType   `json:"floating_ips"`
	PrimaryIPs        []PricingPrimaryIP        `json:"primary_ips"`
	Traffic           PricingTraffic            `json:"traffic"`
	ServerBackup      PricingServerBackup       `json:"server_backup"`
	ServerTypes       []PricingServerType       `json:"server_types"`
	LoadBalancerTypes []PricingLoadBalancerType `json:"load_balancer_types"`
	Volume            PricingVolume             `json:"volume"`
}

// Price defines the schema of a single price with net and gross amount.
type Price struct {
	Net   string `json:"net"`
	Gross string `json:"gross"`
}

// PricingImage defines the schema of pricing information for an image.
type PricingImage struct {
	PricePerGBMonth Price `json:"price_per_gb_month"`
}

// PricingFloatingIP defines the schema of pricing information for a Floating IP.
type PricingFloatingIP struct {
	PriceMonthly Price `json:"price_monthly"`
}

// PricingFloatingIPType defines the schema of pricing information for a Floating IP per type.
type PricingFloatingIPType struct {
	Type   string                       `json:"type"`
	Prices []PricingFloatingIPTypePrice `json:"prices"`
}

// PricingFloatingIPTypePrice defines the schema of pricing information for a Floating IP
// type at a location.
type PricingFloatingIPTypePrice struct {
	Location     string `json:"location"`
	PriceMonthly Price  `json:"price_monthly"`
}

// PricingTraffic defines the schema of pricing information for traffic.
type PricingTraffic struct {
	PricePerTB Price `json:"price_per_tb"`
}

// PricingVolume defines the schema of pricing information for a Volume.
type PricingVolume struct {
	PricePerGBPerMonth Price `json:"price_per_gb_month"`
}

// PricingServerBackup defines the schema of pricing information for server backups.
type PricingServerBackup struct {
	Percentage string `json:"percentage"`
}

// PricingServerType defines the schema of pricing information for a server type.
type PricingServerType struct {
	ID     int                      `json:"id"`
	Name   string                   `json:"name"`
	Prices []PricingServerTypePrice `json:"prices"`
}

// PricingServerTypePrice defines the schema of pricing information for a server
// type at a location.
type PricingServerTypePrice struct {
	Location     string `json:"location"`
	PriceHourly  Price  `json:"price_hourly"`
	PriceMonthly Price  `json:"price_monthly"`
}

// PricingLoadBalancerType defines the schema of pricing information for a Load Balancer type.
type PricingLoadBalancerType struct {
	ID     int                            `json:"id"`
	Name   string                         `json:"name"`
	Prices []PricingLoadBalancerTypePrice `json:"prices"`
}

// PricingLoadBalancerTypePrice defines the schema of pricing information for a Load Balancer
// type at a location.
type PricingLoadBalancerTypePrice struct {
	Location     string `json:"location"`
	PriceHourly  Price  `json:"price_hourly"`
	PriceMonthly Price  `json:"price_monthly"`
}

// PricingGetResponse defines the schema of the response when retrieving pricing information.
type PricingGetResponse struct {
	Pricing Pricing `json:"pricing"`
}

// PricingPrimaryIPTypePrice defines the schema of pricing information for a primary IP
// type at a datacenter.
type PricingPrimaryIPTypePrice struct {
	Datacenter   string `json:"datacenter"`
	PriceHourly  Price  `json:"price_hourly"`
	PriceMonthly Price  `json:"price_monthly"`
}

// PricingPrimaryIP define the schema of pricing information for a primary IP at a datacenter
type PricingPrimaryIP struct {
	Type   string                      `json:"type"`
	Prices []PricingPrimaryIPTypePrice `json:"prices"`
}
