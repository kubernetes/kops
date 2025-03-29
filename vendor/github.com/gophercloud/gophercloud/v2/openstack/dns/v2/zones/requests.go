package zones

import (
	"context"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/pagination"
)

// ListOptsBuilder allows extensions to add parameters to the List request.
type ListOptsBuilder interface {
	ToZoneListQuery() (string, error)
}

// ListOpts allows the filtering and sorting of paginated collections through
// the API. Filtering is achieved by passing in struct field values that map to
// the server attributes you want to see returned. Marker and Limit are used
// for pagination.
// https://developer.openstack.org/api-ref/dns/
type ListOpts struct {
	// Integer value for the limit of values to return.
	Limit int `q:"limit"`

	// UUID of the zone at which you want to set a marker.
	Marker string `q:"marker"`

	Description string `q:"description"`
	Email       string `q:"email"`
	Name        string `q:"name"`
	SortDir     string `q:"sort_dir"`
	SortKey     string `q:"sort_key"`
	Status      string `q:"status"`
	TTL         int    `q:"ttl"`
	Type        string `q:"type"`
}

// ToZoneListQuery formats a ListOpts into a query string.
func (opts ListOpts) ToZoneListQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	return q.String(), err
}

// List implements a zone List request.
func List(client *gophercloud.ServiceClient, opts ListOptsBuilder) pagination.Pager {
	url := baseURL(client)
	if opts != nil {
		query, err := opts.ToZoneListQuery()
		if err != nil {
			return pagination.Pager{Err: err}
		}
		url += query
	}
	return pagination.NewPager(client, url, func(r pagination.PageResult) pagination.Page {
		return ZonePage{pagination.LinkedPageBase{PageResult: r}}
	})
}

// Get returns information about a zone, given its ID.
func Get(ctx context.Context, client *gophercloud.ServiceClient, zoneID string) (r GetResult) {
	resp, err := client.Get(ctx, zoneURL(client, zoneID), &r.Body, nil)
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// CreateOptsBuilder allows extensions to add additional attributes to the
// Create request.
type CreateOptsBuilder interface {
	ToZoneCreateMap() (map[string]any, error)
}

// CreateOpts specifies the attributes used to create a zone.
type CreateOpts struct {
	// Attributes are settings that supply hints and filters for the zone.
	Attributes map[string]string `json:"attributes,omitempty"`

	// Email contact of the zone.
	Email string `json:"email,omitempty"`

	// Description of the zone.
	Description string `json:"description,omitempty"`

	// Name of the zone.
	Name string `json:"name" required:"true"`

	// Masters specifies zone masters if this is a secondary zone.
	Masters []string `json:"masters,omitempty"`

	// TTL is the time to live of the zone.
	TTL int `json:"-"`

	// Type specifies if this is a primary or secondary zone.
	Type string `json:"type,omitempty"`
}

// ToZoneCreateMap formats an CreateOpts structure into a request body.
func (opts CreateOpts) ToZoneCreateMap() (map[string]any, error) {
	b, err := gophercloud.BuildRequestBody(opts, "")
	if err != nil {
		return nil, err
	}

	if opts.TTL > 0 {
		b["ttl"] = opts.TTL
	}

	return b, nil
}

// Create implements a zone create request.
func Create(ctx context.Context, client *gophercloud.ServiceClient, opts CreateOptsBuilder) (r CreateResult) {
	b, err := opts.ToZoneCreateMap()
	if err != nil {
		r.Err = err
		return
	}
	resp, err := client.Post(ctx, baseURL(client), &b, &r.Body, &gophercloud.RequestOpts{
		OkCodes: []int{201, 202},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// UpdateOptsBuilder allows extensions to add additional attributes to the
// Update request.
type UpdateOptsBuilder interface {
	ToZoneUpdateMap() (map[string]any, error)
}

// UpdateOpts specifies the attributes to update a zone.
type UpdateOpts struct {
	// Email contact of the zone.
	Email string `json:"email,omitempty"`

	// TTL is the time to live of the zone.
	TTL int `json:"-"`

	// Masters specifies zone masters if this is a secondary zone.
	Masters []string `json:"masters,omitempty"`

	// Description of the zone.
	Description *string `json:"description,omitempty"`
}

// ToZoneUpdateMap formats an UpdateOpts structure into a request body.
func (opts UpdateOpts) ToZoneUpdateMap() (map[string]any, error) {
	b, err := gophercloud.BuildRequestBody(opts, "")
	if err != nil {
		return nil, err
	}

	if opts.TTL > 0 {
		b["ttl"] = opts.TTL
	}

	return b, nil
}

// Update implements a zone update request.
func Update(ctx context.Context, client *gophercloud.ServiceClient, zoneID string, opts UpdateOptsBuilder) (r UpdateResult) {
	b, err := opts.ToZoneUpdateMap()
	if err != nil {
		r.Err = err
		return
	}
	resp, err := client.Patch(ctx, zoneURL(client, zoneID), &b, &r.Body, &gophercloud.RequestOpts{
		OkCodes: []int{200, 202},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// Delete implements a zone delete request.
func Delete(ctx context.Context, client *gophercloud.ServiceClient, zoneID string) (r DeleteResult) {
	resp, err := client.Delete(ctx, zoneURL(client, zoneID), &gophercloud.RequestOpts{
		OkCodes:      []int{202},
		JSONResponse: &r.Body,
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// request body for sharing a zone.
type ShareOptsBuilder interface {
	ToShareMap() (map[string]interface{}, error)
}

// ShareZoneOpts specifies the target project for sharing a zone.
type ShareZoneOpts struct {
	// TargetProjectID is the ID of the project to share the zone with.
	TargetProjectID string `json:"target_project_id" required:"true"`
}

// ToShareMap constructs a request body from a ShareZoneOpts.
func (opts ShareZoneOpts) ToShareMap() (map[string]interface{}, error) {
	return map[string]interface{}{
		"target_project_id": opts.TargetProjectID,
	}, nil
}

// Share shares a zone with another project.
func Share(ctx context.Context, client *gophercloud.ServiceClient, zoneID string, opts ShareOptsBuilder) (r gophercloud.ErrResult) {
	body, err := gophercloud.BuildRequestBody(opts, "")
	if err != nil {
		r.Err = err
		return
	}

	resp, err := client.Post(ctx, zoneShareURL(client, zoneID), body, nil, &gophercloud.RequestOpts{
		OkCodes: []int{201},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// Unshare removes a share for a zone.
func Unshare(ctx context.Context, client *gophercloud.ServiceClient, zoneID, shareID string) (r gophercloud.ErrResult) {
	resp, err := client.Delete(ctx, zoneUnshareURL(client, zoneID, shareID), &gophercloud.RequestOpts{
		OkCodes: []int{204},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}
