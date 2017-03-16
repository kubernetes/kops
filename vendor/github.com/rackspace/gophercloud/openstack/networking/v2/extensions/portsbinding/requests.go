package portsbinding

import (
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/networking/v2/ports"
)

// Get retrieves a specific port based on its unique ID.
func Get(c *gophercloud.ServiceClient, id string) GetResult {
	var res GetResult
	_, res.Err = c.Get(getURL(c, id), &res.Body, nil)
	return res
}

// CreateOpts represents the attributes used when creating a new
// port with extended attributes.
type CreateOpts struct {
	// CreateOptsBuilder is the interface options structs have to satisfy in order
	// to be used in the main Create operation in this package.
	ports.CreateOptsBuilder
	// The ID of the host where the port is allocated
	HostID string
	// The virtual network interface card (vNIC) type that is bound to the
	// neutron port
	VNICType string
	// A dictionary that enables the application running on the specified
	// host to pass and receive virtual network interface (VIF) port-specific
	// information to the plug-in
	Profile map[string]string
}

// ToPortCreateMap casts a CreateOpts struct to a map.
func (opts CreateOpts) ToPortCreateMap() (map[string]interface{}, error) {
	p, err := opts.CreateOptsBuilder.ToPortCreateMap()
	if err != nil {
		return nil, err
	}

	port := p["port"].(map[string]interface{})

	if opts.HostID != "" {
		port["binding:host_id"] = opts.HostID
	}
	if opts.VNICType != "" {
		port["binding:vnic_type"] = opts.VNICType
	}
	if opts.Profile != nil {
		port["binding:profile"] = opts.Profile
	}

	return map[string]interface{}{"port": port}, nil
}

// Create accepts a CreateOpts struct and creates a new port with extended attributes.
// You must remember to provide a NetworkID value.
func Create(c *gophercloud.ServiceClient, opts ports.CreateOptsBuilder) CreateResult {
	var res CreateResult

	reqBody, err := opts.ToPortCreateMap()
	if err != nil {
		res.Err = err
		return res
	}

	_, res.Err = c.Post(createURL(c), reqBody, &res.Body, nil)
	return res
}

// UpdateOpts represents the attributes used when updating an existing port.
type UpdateOpts struct {
	// UpdateOptsBuilder is the interface options structs have to satisfy in order
	// to be used in the main Update operation in this package.
	ports.UpdateOptsBuilder
	// The ID of the host where the port is allocated
	HostID string
	// The virtual network interface card (vNIC) type that is bound to the
	// neutron port
	VNICType string
	// A dictionary that enables the application running on the specified
	// host to pass and receive virtual network interface (VIF) port-specific
	// information to the plug-in
	Profile map[string]string
}

// ToPortUpdateMap casts an UpdateOpts struct to a map.
func (opts UpdateOpts) ToPortUpdateMap() (map[string]interface{}, error) {
	var port map[string]interface{}
	if opts.UpdateOptsBuilder != nil {
		p, err := opts.UpdateOptsBuilder.ToPortUpdateMap()
		if err != nil {
			return nil, err
		}

		port = p["port"].(map[string]interface{})
	}

	if port == nil {
		port = make(map[string]interface{})
	}

	if opts.HostID != "" {
		port["binding:host_id"] = opts.HostID
	}
	if opts.VNICType != "" {
		port["binding:vnic_type"] = opts.VNICType
	}
	if opts.Profile != nil {
		port["binding:profile"] = opts.Profile
	}

	return map[string]interface{}{"port": port}, nil
}

// Update accepts a UpdateOpts struct and updates an existing port using the
// values provided.
func Update(c *gophercloud.ServiceClient, id string, opts ports.UpdateOptsBuilder) UpdateResult {
	var res UpdateResult

	reqBody, err := opts.ToPortUpdateMap()
	if err != nil {
		res.Err = err
		return res
	}

	_, res.Err = c.Put(updateURL(c, id), reqBody, &res.Body, &gophercloud.RequestOpts{
		OkCodes: []int{200, 201},
	})
	return res
}
