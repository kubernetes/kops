package schema

import "time"

// PrimaryIP defines a Primary IP.
type PrimaryIP struct {
	ID           int                 `json:"id"`
	IP           string              `json:"ip"`
	Labels       map[string]string   `json:"labels"`
	Name         string              `json:"name"`
	Type         string              `json:"type"`
	Protection   PrimaryIPProtection `json:"protection"`
	DNSPtr       []PrimaryIPDNSPTR   `json:"dns_ptr"`
	AssigneeID   int                 `json:"assignee_id"`
	AssigneeType string              `json:"assignee_type"`
	AutoDelete   bool                `json:"auto_delete"`
	Blocked      bool                `json:"blocked"`
	Created      time.Time           `json:"created"`
	Datacenter   Datacenter          `json:"datacenter"`
}

// PrimaryIPProtection represents the protection level of a Primary IP.
type PrimaryIPProtection struct {
	Delete bool `json:"delete"`
}

// PrimaryIPDNSPTR contains reverse DNS information for a
// IPv4 or IPv6 Primary IP.
type PrimaryIPDNSPTR struct {
	DNSPtr string `json:"dns_ptr"`
	IP     string `json:"ip"`
}

// PrimaryIPCreateResponse defines the schema of the response
// when creating a Primary IP.
type PrimaryIPCreateResponse struct {
	PrimaryIP PrimaryIP `json:"primary_ip"`
	Action    *Action   `json:"action"`
}

// PrimaryIPGetResult defines the response when retrieving a single Primary IP.
type PrimaryIPGetResult struct {
	PrimaryIP PrimaryIP `json:"primary_ip"`
}

// PrimaryIPListResult defines the response when listing Primary IPs.
type PrimaryIPListResult struct {
	PrimaryIPs []PrimaryIP `json:"primary_ips"`
}

// PrimaryIPUpdateResult defines the response
// when updating a Primary IP.
type PrimaryIPUpdateResult struct {
	PrimaryIP PrimaryIP `json:"primary_ip"`
}
