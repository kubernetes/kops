package rules

import (
	"encoding/json"
	"time"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/pagination"
)

// SecGroupRule represents a rule to dictate the behaviour of incoming or
// outgoing traffic for a particular security group.
type SecGroupRule struct {
	// The UUID for this security group rule.
	ID string

	// The direction in which the security group rule is applied. The only values
	// allowed are "ingress" or "egress". For a compute instance, an ingress
	// security group rule is applied to incoming (ingress) traffic for that
	// instance. An egress rule is applied to traffic leaving the instance.
	Direction string

	// Description of the rule
	Description string `json:"description"`

	// Must be IPv4 or IPv6, and addresses represented in CIDR must match the
	// ingress or egress rules.
	EtherType string `json:"ethertype"`

	// The security group ID to associate with this security group rule.
	SecGroupID string `json:"security_group_id"`

	// The minimum port number in the range that is matched by the security group
	// rule. If the protocol is TCP or UDP, this value must be less than or equal
	// to the value of the PortRangeMax attribute. If the protocol is ICMP, this
	// value must be an ICMP type.
	PortRangeMin int `json:"port_range_min"`

	// The maximum port number in the range that is matched by the security group
	// rule. The PortRangeMin attribute constrains the PortRangeMax attribute. If
	// the protocol is ICMP, this value must be an ICMP type.
	PortRangeMax int `json:"port_range_max"`

	// The protocol that is matched by the security group rule. Valid values are
	// "tcp", "udp", "icmp" or an empty string.
	Protocol string

	// The remote address group ID to be associated with this security group rule.
	// You can specify either RemoteAddressGroupID, RemoteGroupID, or RemoteIPPrefix
	RemoteAddressGroupID string `json:"remote_address_group_id"`

	// The remote group ID to be associated with this security group rule. You
	// can specify either RemoteGroupID or RemoteIPPrefix.
	RemoteGroupID string `json:"remote_group_id"`

	// The remote IP prefix to be associated with this security group rule. You
	// can specify either RemoteGroupID or RemoteIPPrefix . This attribute
	// matches the specified IP prefix as the source IP address of the IP packet.
	RemoteIPPrefix string `json:"remote_ip_prefix"`

	// TenantID is the project owner of this security group rule.
	TenantID string `json:"tenant_id"`

	// ProjectID is the project owner of this security group rule.
	ProjectID string `json:"project_id"`

	// RevisionNumber optionally set via extensions/standard-attr-revisions
	RevisionNumber int `json:"revision_number"`

	// Timestamp when the rule was created
	CreatedAt time.Time `json:"-"`

	// Timestamp when the rule was last updated
	UpdatedAt time.Time `json:"-"`
}

func (r *SecGroupRule) UnmarshalJSON(b []byte) error {
	type tmp SecGroupRule

	// Support for older neutron time format
	var s1 struct {
		tmp
		CreatedAt gophercloud.JSONRFC3339NoZ `json:"created_at"`
		UpdatedAt gophercloud.JSONRFC3339NoZ `json:"updated_at"`
	}

	err := json.Unmarshal(b, &s1)
	if err == nil {
		*r = SecGroupRule(s1.tmp)
		r.CreatedAt = time.Time(s1.CreatedAt)
		r.UpdatedAt = time.Time(s1.UpdatedAt)

		return nil
	}

	// Support for newer neutron time format
	var s2 struct {
		tmp
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	err = json.Unmarshal(b, &s2)
	if err != nil {
		return err
	}

	*r = SecGroupRule(s2.tmp)
	r.CreatedAt = time.Time(s2.CreatedAt)
	r.UpdatedAt = time.Time(s2.UpdatedAt)

	return nil
}

// SecGroupRulePage is the page returned by a pager when traversing over a
// collection of security group rules.
type SecGroupRulePage struct {
	pagination.LinkedPageBase
}

// NextPageURL is invoked when a paginated collection of security group rules has
// reached the end of a page and the pager seeks to traverse over a new one. In
// order to do this, it needs to construct the next page's URL.
func (r SecGroupRulePage) NextPageURL() (string, error) {
	var s struct {
		Links []gophercloud.Link `json:"security_group_rules_links"`
	}
	err := r.ExtractInto(&s)
	if err != nil {
		return "", err
	}
	return gophercloud.ExtractNextURL(s.Links)
}

// IsEmpty checks whether a SecGroupRulePage struct is empty.
func (r SecGroupRulePage) IsEmpty() (bool, error) {
	if r.StatusCode == 204 {
		return true, nil
	}

	is, err := ExtractRules(r)
	return len(is) == 0, err
}

// ExtractRules accepts a Page struct, specifically a SecGroupRulePage struct,
// and extracts the elements into a slice of SecGroupRule structs. In other words,
// a generic collection is mapped into a relevant slice.
func ExtractRules(r pagination.Page) ([]SecGroupRule, error) {
	var s struct {
		SecGroupRules []SecGroupRule `json:"security_group_rules"`
	}
	err := (r.(SecGroupRulePage)).ExtractInto(&s)
	return s.SecGroupRules, err
}

type commonResult struct {
	gophercloud.Result
}

type bulkResult struct {
	gophercloud.Result
}

// Extract is a function that accepts a result and extracts a security rule.
func (r commonResult) Extract() (*SecGroupRule, error) {
	var s struct {
		SecGroupRule *SecGroupRule `json:"security_group_rule"`
	}
	err := r.ExtractInto(&s)
	return s.SecGroupRule, err
}

// Extract is a function that accepts a result and extracts security rules.
func (r bulkResult) Extract() ([]SecGroupRule, error) {
	var s struct {
		SecGroupRules []SecGroupRule `json:"security_group_rules"`
	}
	err := r.ExtractInto(&s)
	return s.SecGroupRules, err
}

// CreateResult represents the result of a create operation. Call its Extract
// method to interpret it as a SecGroupRule.
type CreateResult struct {
	commonResult
}

// CreateBulkResult represents the result of a bulk create operation. Call its
// Extract method to interpret it as a slice of SecGroupRules.
type CreateBulkResult struct {
	bulkResult
}

// GetResult represents the result of a get operation. Call its Extract
// method to interpret it as a SecGroupRule.
type GetResult struct {
	commonResult
}

// DeleteResult represents the result of a delete operation. Call its
// ExtractErr method to determine if the request succeeded or failed.
type DeleteResult struct {
	gophercloud.ErrResult
}
