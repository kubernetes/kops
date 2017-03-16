package portsbinding

import (
	"github.com/mitchellh/mapstructure"
	"github.com/rackspace/gophercloud"

	"github.com/rackspace/gophercloud/openstack/networking/v2/ports"
	"github.com/rackspace/gophercloud/pagination"
)

type commonResult struct {
	gophercloud.Result
}

// Extract is a function that accepts a result and extracts a port resource.
func (r commonResult) Extract() (*Port, error) {
	if r.Err != nil {
		return nil, r.Err
	}

	var res struct {
		Port *Port `json:"port"`
	}

	err := mapstructure.Decode(r.Body, &res)

	return res.Port, err
}

// CreateResult represents the result of a create operation.
type CreateResult struct {
	commonResult
}

// GetResult represents the result of a get operation.
type GetResult struct {
	commonResult
}

// UpdateResult represents the result of an update operation.
type UpdateResult struct {
	commonResult
}

// IP is a sub-struct that represents an individual IP.
type IP struct {
	SubnetID  string `mapstructure:"subnet_id" json:"subnet_id"`
	IPAddress string `mapstructure:"ip_address" json:"ip_address,omitempty"`
}

// Port represents a Neutron port. See package documentation for a top-level
// description of what this is.
type Port struct {
	ports.Port `mapstructure:",squash"`
	// The ID of the host where the port is allocated
	HostID string `mapstructure:"binding:host_id" json:"binding:host_id"`
	// A dictionary that enables the application to pass information about
	// functions that the Networking API provides.
	VIFDetails map[string]interface{} `mapstructure:"binding:vif_details" json:"binding:vif_details"`
	// The VIF type for the port.
	VIFType string `mapstructure:"binding:vif_type" json:"binding:vif_type"`
	// The virtual network interface card (vNIC) type that is bound to the
	// neutron port
	VNICType string `mapstructure:"binding:vnic_type" json:"binding:vnic_type"`
	// A dictionary that enables the application running on the specified
	// host to pass and receive virtual network interface (VIF) port-specific
	// information to the plug-in
	Profile map[string]string `mapstructure:"binding:profile" json:"binding:profile"`
}

// ExtractPorts accepts a Page struct, specifically a PortPage struct,
// and extracts the elements into a slice of Port structs. In other words,
// a generic collection is mapped into a relevant slice.
func ExtractPorts(page pagination.Page) ([]Port, error) {
	var resp struct {
		Ports []Port `mapstructure:"ports" json:"ports"`
	}

	err := mapstructure.Decode(page.(ports.PortPage).Body, &resp)
	return resp.Ports, err
}
