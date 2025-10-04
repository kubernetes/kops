package routers

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/pagination"
)

// ListOptsBuilder allows extensions to add additional parameters to the List
// request.
type ListOptsBuilder interface {
	ToRouterListQuery() (string, error)
}

// ListOpts allows the filtering and sorting of paginated collections through
// the API. Filtering is achieved by passing in struct field values that map to
// the floating IP attributes you want to see returned. SortKey allows you to
// sort by a particular network attribute. SortDir sets the direction, and is
// either `asc' or `desc'. Marker and Limit are used for pagination.
type ListOpts struct {
	ID             string `q:"id"`
	Name           string `q:"name"`
	Description    string `q:"description"`
	AdminStateUp   *bool  `q:"admin_state_up"`
	Distributed    *bool  `q:"distributed"`
	Status         string `q:"status"`
	TenantID       string `q:"tenant_id"`
	ProjectID      string `q:"project_id"`
	Limit          int    `q:"limit"`
	Marker         string `q:"marker"`
	SortKey        string `q:"sort_key"`
	SortDir        string `q:"sort_dir"`
	Tags           string `q:"tags"`
	TagsAny        string `q:"tags-any"`
	NotTags        string `q:"not-tags"`
	NotTagsAny     string `q:"not-tags-any"`
	RevisionNumber *int   `q:"revision_number"`
}

// ToRouterListQuery formats a ListOpts into a query string.
func (opts ListOpts) ToRouterListQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(&opts)
	if err != nil {
		return "", err
	}
	return q.String(), nil
}

// List returns a Pager which allows you to iterate over a collection of
// routers. It accepts a ListOpts struct, which allows you to filter and sort
// the returned collection for greater efficiency.
//
// Default policy settings return only those routers that are owned by the
// tenant who submits the request, unless an admin user submits the request.
func List(c *gophercloud.ServiceClient, opts ListOptsBuilder) pagination.Pager {
	url := rootURL(c)
	if opts != nil {
		query, err := opts.ToRouterListQuery()
		if err != nil {
			return pagination.Pager{Err: err}
		}
		url += query
	}
	return pagination.NewPager(c, url, func(r pagination.PageResult) pagination.Page {
		return RouterPage{pagination.LinkedPageBase{PageResult: r}}
	})
}

// CreateOptsBuilder allows extensions to add additional parameters to the
// Create request.
type CreateOptsBuilder interface {
	ToRouterCreateMap() (map[string]any, error)
}

// CreateOpts contains all the values needed to create a new router. There are
// no required values.
type CreateOpts struct {
	Name                  string       `json:"name,omitempty"`
	Description           string       `json:"description,omitempty"`
	AdminStateUp          *bool        `json:"admin_state_up,omitempty"`
	Distributed           *bool        `json:"distributed,omitempty"`
	TenantID              string       `json:"tenant_id,omitempty"`
	ProjectID             string       `json:"project_id,omitempty"`
	GatewayInfo           *GatewayInfo `json:"external_gateway_info,omitempty"`
	AvailabilityZoneHints []string     `json:"availability_zone_hints,omitempty"`
}

// ToRouterCreateMap builds a create request body from CreateOpts.
func (opts CreateOpts) ToRouterCreateMap() (map[string]any, error) {
	return gophercloud.BuildRequestBody(opts, "router")
}

// Create accepts a CreateOpts struct and uses the values to create a new
// logical router. When it is created, the router does not have an internal
// interface - it is not associated to any subnet.
//
// You can optionally specify an external gateway for a router using the
// GatewayInfo struct. The external gateway for the router must be plugged into
// an external network (it is external if its `router:external' field is set to
// true).
func Create(ctx context.Context, c *gophercloud.ServiceClient, opts CreateOptsBuilder) (r CreateResult) {
	b, err := opts.ToRouterCreateMap()
	if err != nil {
		r.Err = err
		return
	}
	resp, err := c.Post(ctx, rootURL(c), b, &r.Body, nil)
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// Get retrieves a particular router based on its unique ID.
func Get(ctx context.Context, c *gophercloud.ServiceClient, id string) (r GetResult) {
	resp, err := c.Get(ctx, resourceURL(c, id), &r.Body, nil)
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// UpdateOptsBuilder allows extensions to add additional parameters to the
// Update request.
type UpdateOptsBuilder interface {
	ToRouterUpdateMap() (map[string]any, error)
}

// UpdateOpts contains the values used when updating a router.
type UpdateOpts struct {
	Name         string       `json:"name,omitempty"`
	Description  *string      `json:"description,omitempty"`
	AdminStateUp *bool        `json:"admin_state_up,omitempty"`
	Distributed  *bool        `json:"distributed,omitempty"`
	GatewayInfo  *GatewayInfo `json:"external_gateway_info,omitempty"`
	Routes       *[]Route     `json:"routes,omitempty"`

	// RevisionNumber implements extension:standard-attr-revisions. If != "" it
	// will set revision_number=%s. If the revision number does not match, the
	// update will fail.
	RevisionNumber *int `json:"-" h:"If-Match"`
}

// ToRouterUpdateMap builds an update body based on UpdateOpts.
func (opts UpdateOpts) ToRouterUpdateMap() (map[string]any, error) {
	return gophercloud.BuildRequestBody(opts, "router")
}

// Update allows routers to be updated. You can update the name, administrative
// state, and the external gateway. For more information about how to set the
// external gateway for a router, see Create. This operation does not enable
// the update of router interfaces. To do this, use the AddInterface and
// RemoveInterface functions.
func Update(ctx context.Context, c *gophercloud.ServiceClient, id string, opts UpdateOptsBuilder) (r UpdateResult) {
	b, err := opts.ToRouterUpdateMap()
	if err != nil {
		r.Err = err
		return
	}
	h, err := gophercloud.BuildHeaders(opts)
	if err != nil {
		r.Err = err
		return
	}
	for k := range h {
		if k == "If-Match" {
			h[k] = fmt.Sprintf("revision_number=%s", h[k])
		}
	}
	resp, err := c.Put(ctx, resourceURL(c, id), b, &r.Body, &gophercloud.RequestOpts{
		MoreHeaders: h,
		OkCodes:     []int{200},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// Delete will permanently delete a particular router based on its unique ID.
func Delete(ctx context.Context, c *gophercloud.ServiceClient, id string) (r DeleteResult) {
	resp, err := c.Delete(ctx, resourceURL(c, id), nil)
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// AddInterfaceOptsBuilder allows extensions to add additional parameters to
// the AddInterface request.
type AddInterfaceOptsBuilder interface {
	ToRouterAddInterfaceMap() (map[string]any, error)
}

// AddInterfaceOpts represents the options for adding an interface to a router.
type AddInterfaceOpts struct {
	SubnetID string `json:"subnet_id,omitempty" xor:"PortID"`
	PortID   string `json:"port_id,omitempty" xor:"SubnetID"`
}

// ToRouterAddInterfaceMap builds a request body from AddInterfaceOpts.
func (opts AddInterfaceOpts) ToRouterAddInterfaceMap() (map[string]any, error) {
	return gophercloud.BuildRequestBody(opts, "")
}

// AddInterface attaches a subnet to an internal router interface. You must
// specify either a SubnetID or PortID in the request body. If you specify both,
// the operation will fail and an error will be returned.
//
// If you specify a SubnetID, the gateway IP address for that particular subnet
// is used to create the router interface. Alternatively, if you specify a
// PortID, the IP address associated with the port is used to create the router
// interface.
//
// If you reference a port that is associated with multiple IP addresses, or
// if the port is associated with zero IP addresses, the operation will fail and
// a 400 Bad Request error will be returned.
//
// If you reference a port already in use, the operation will fail and a 409
// Conflict error will be returned.
//
// The PortID that is returned after using Extract() on the result of this
// operation can either be the same PortID passed in or, on the other hand, the
// identifier of a new port created by this operation. After the operation
// completes, the device ID of the port is set to the router ID, and the
// device owner attribute is set to `network:router_interface'.
func AddInterface(ctx context.Context, c *gophercloud.ServiceClient, id string, opts AddInterfaceOptsBuilder) (r InterfaceResult) {
	b, err := opts.ToRouterAddInterfaceMap()
	if err != nil {
		r.Err = err
		return
	}
	resp, err := c.Put(ctx, addInterfaceURL(c, id), b, &r.Body, &gophercloud.RequestOpts{
		OkCodes: []int{200},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// RemoveInterfaceOptsBuilder allows extensions to add additional parameters to
// the RemoveInterface request.
type RemoveInterfaceOptsBuilder interface {
	ToRouterRemoveInterfaceMap() (map[string]any, error)
}

// RemoveInterfaceOpts represents options for removing an interface from
// a router.
type RemoveInterfaceOpts struct {
	SubnetID string `json:"subnet_id,omitempty" or:"PortID"`
	PortID   string `json:"port_id,omitempty" or:"SubnetID"`
}

// ToRouterRemoveInterfaceMap builds a request body based on
// RemoveInterfaceOpts.
func (opts RemoveInterfaceOpts) ToRouterRemoveInterfaceMap() (map[string]any, error) {
	return gophercloud.BuildRequestBody(opts, "")
}

// RemoveInterface removes an internal router interface, which detaches a
// subnet from the router. You must specify either a SubnetID or PortID, since
// these values are used to identify the router interface to remove.
//
// Unlike AddInterface, you can also specify both a SubnetID and PortID. If you
// choose to specify both, the subnet ID must correspond to the subnet ID of
// the first IP address on the port specified by the port ID. Otherwise, the
// operation will fail and return a 409 Conflict error.
//
// If the router, subnet or port which are referenced do not exist or are not
// visible to you, the operation will fail and a 404 Not Found error will be
// returned. After this operation completes, the port connecting the router
// with the subnet is removed from the subnet for the network.
func RemoveInterface(ctx context.Context, c *gophercloud.ServiceClient, id string, opts RemoveInterfaceOptsBuilder) (r InterfaceResult) {
	b, err := opts.ToRouterRemoveInterfaceMap()
	if err != nil {
		r.Err = err
		return
	}
	resp, err := c.Put(ctx, removeInterfaceURL(c, id), b, &r.Body, &gophercloud.RequestOpts{
		OkCodes: []int{200},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// ListL3Agents returns a list of l3-agents scheduled for a specific router.
func ListL3Agents(c *gophercloud.ServiceClient, id string) (result pagination.Pager) {
	return pagination.NewPager(c, listl3AgentsURL(c, id), func(r pagination.PageResult) pagination.Page {
		return ListL3AgentsPage{pagination.SinglePageBase(r)}
	})
}
