// This file was automatically generated. DO NOT EDIT.
// If you have any remark or suggestion do not hesitate to open an issue.

// Package vpc provides methods and message types of the vpc v1 API.
package vpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/internal/marshaler"
	"github.com/scaleway/scaleway-sdk-go/internal/parameter"
	"github.com/scaleway/scaleway-sdk-go/namegenerator"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// always import dependencies
var (
	_ fmt.Stringer
	_ json.Unmarshaler
	_ url.URL
	_ net.IP
	_ http.Header
	_ bytes.Reader
	_ time.Time
	_ = strings.Join

	_ scw.ScalewayRequest
	_ marshaler.Duration
	_ scw.File
	_ = parameter.AddToQuery
	_ = namegenerator.GetRandomName
)

// API: vPC API.
type API struct {
	client *scw.Client
}

// NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

type ListPrivateNetworksRequestOrderBy string

const (
	ListPrivateNetworksRequestOrderByCreatedAtAsc  = ListPrivateNetworksRequestOrderBy("created_at_asc")
	ListPrivateNetworksRequestOrderByCreatedAtDesc = ListPrivateNetworksRequestOrderBy("created_at_desc")
	ListPrivateNetworksRequestOrderByNameAsc       = ListPrivateNetworksRequestOrderBy("name_asc")
	ListPrivateNetworksRequestOrderByNameDesc      = ListPrivateNetworksRequestOrderBy("name_desc")
)

func (enum ListPrivateNetworksRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListPrivateNetworksRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListPrivateNetworksRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListPrivateNetworksRequestOrderBy(ListPrivateNetworksRequestOrderBy(tmp).String())
	return nil
}

type ListPrivateNetworksResponse struct {
	PrivateNetworks []*PrivateNetwork `json:"private_networks"`

	TotalCount uint32 `json:"total_count"`
}

// PrivateNetwork: private network.
type PrivateNetwork struct {
	// ID: private Network ID.
	ID string `json:"id"`
	// Name: private Network name.
	Name string `json:"name"`
	// OrganizationID: scaleway Organization the Private Network belongs to.
	OrganizationID string `json:"organization_id"`
	// ProjectID: scaleway Project the Private Network belongs to.
	ProjectID string `json:"project_id"`
	// Zone: availability Zone in which the Private Network is available.
	Zone scw.Zone `json:"zone"`
	// Tags: tags of the Private Network.
	Tags []string `json:"tags"`
	// CreatedAt: date the Private Network was created.
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: date the Private Network was last modified.
	UpdatedAt *time.Time `json:"updated_at"`
	// Subnets: private Network subnets CIDR.
	Subnets []scw.IPNet `json:"subnets"`
}

// Service API

// Zones list localities the api is available in
func (s *API) Zones() []scw.Zone {
	return []scw.Zone{scw.ZoneFrPar1, scw.ZoneFrPar2, scw.ZoneFrPar3, scw.ZoneNlAms1, scw.ZoneNlAms2, scw.ZoneNlAms3, scw.ZonePlWaw1, scw.ZonePlWaw2}
}

type ListPrivateNetworksRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// OrderBy: sort order of the returned Private Networks.
	// Default value: created_at_asc
	OrderBy ListPrivateNetworksRequestOrderBy `json:"-"`
	// Page: page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// PageSize: maximum number of Private Networks to return per page.
	PageSize *uint32 `json:"-"`
	// Name: name to filter for. Only Private Networks with names containing this string will be returned.
	Name *string `json:"-"`
	// Tags: tags to filter for. Only Private Networks with one or more matching tags will be returned.
	Tags []string `json:"-"`
	// OrganizationID: organization ID to filter for. Only Private Networks belonging to this Organization will be returned.
	OrganizationID *string `json:"-"`
	// ProjectID: project ID to filter for. Only Private Networks belonging to this Project will be returned.
	ProjectID *string `json:"-"`
	// PrivateNetworkIDs: private Network IDs to filter for. Only Private Networks with one of these IDs will be returned.
	PrivateNetworkIDs []string `json:"-"`
	// IncludeRegional: defines whether to include regional Private Networks in the response.
	IncludeRegional *bool `json:"-"`
}

// ListPrivateNetworks: list Private Networks.
// List existing Private Networks in a specified Availability Zone. By default, the Private Networks returned in the list are ordered by creation date in ascending order, though this can be modified via the order_by field.
func (s *API) ListPrivateNetworks(req *ListPrivateNetworksRequest, opts ...scw.RequestOption) (*ListPrivateNetworksResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "tags", req.Tags)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "project_id", req.ProjectID)
	parameter.AddToQuery(query, "private_network_ids", req.PrivateNetworkIDs)
	parameter.AddToQuery(query, "include_regional", req.IncludeRegional)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/vpc/v1/zones/" + fmt.Sprint(req.Zone) + "/private-networks",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListPrivateNetworksResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreatePrivateNetworkRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// Name: name for the Private Network.
	Name string `json:"name"`
	// ProjectID: scaleway Project in which to create the Private Network.
	ProjectID string `json:"project_id"`
	// Tags: tags for the Private Network.
	Tags []string `json:"tags"`
	// Subnets: private Network subnets CIDR.
	Subnets []scw.IPNet `json:"subnets"`
}

// CreatePrivateNetwork: create a Private Network.
// Create a new Private Network. Once created, you can attach Scaleway resources in the same Availability Zone.
func (s *API) CreatePrivateNetwork(req *CreatePrivateNetworkRequest, opts ...scw.RequestOption) (*PrivateNetwork, error) {
	var err error

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("pn")
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/vpc/v1/zones/" + fmt.Sprint(req.Zone) + "/private-networks",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp PrivateNetwork

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetPrivateNetworkRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// PrivateNetworkID: private Network ID.
	PrivateNetworkID string `json:"-"`
}

// GetPrivateNetwork: get a Private Network.
// Retrieve information about an existing Private Network, specified by its Private Network ID. Its full details are returned in the response object.
func (s *API) GetPrivateNetwork(req *GetPrivateNetworkRequest, opts ...scw.RequestOption) (*PrivateNetwork, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.PrivateNetworkID) == "" {
		return nil, errors.New("field PrivateNetworkID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/vpc/v1/zones/" + fmt.Sprint(req.Zone) + "/private-networks/" + fmt.Sprint(req.PrivateNetworkID) + "",
		Headers: http.Header{},
	}

	var resp PrivateNetwork

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdatePrivateNetworkRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// PrivateNetworkID: private Network ID.
	PrivateNetworkID string `json:"-"`
	// Name: name of the private network.
	Name *string `json:"name"`
	// Tags: tags for the Private Network.
	Tags *[]string `json:"tags"`
	// Deprecated: Subnets: private Network subnets CIDR (deprecated).
	Subnets *[]string `json:"subnets,omitempty"`
}

// UpdatePrivateNetwork: update Private Network.
// Update parameters (such as name or tags) of an existing Private Network, specified by its Private Network ID.
func (s *API) UpdatePrivateNetwork(req *UpdatePrivateNetworkRequest, opts ...scw.RequestOption) (*PrivateNetwork, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.PrivateNetworkID) == "" {
		return nil, errors.New("field PrivateNetworkID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/vpc/v1/zones/" + fmt.Sprint(req.Zone) + "/private-networks/" + fmt.Sprint(req.PrivateNetworkID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp PrivateNetwork

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeletePrivateNetworkRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// PrivateNetworkID: private Network ID.
	PrivateNetworkID string `json:"-"`
}

// DeletePrivateNetwork: delete a Private Network.
// Delete an existing Private Network. Note that you must first detach all resources from the network, in order to delete it.
func (s *API) DeletePrivateNetwork(req *DeletePrivateNetworkRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.PrivateNetworkID) == "" {
		return errors.New("field PrivateNetworkID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/vpc/v1/zones/" + fmt.Sprint(req.Zone) + "/private-networks/" + fmt.Sprint(req.PrivateNetworkID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListPrivateNetworksResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListPrivateNetworksResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListPrivateNetworksResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.PrivateNetworks = append(r.PrivateNetworks, results.PrivateNetworks...)
	r.TotalCount += uint32(len(results.PrivateNetworks))
	return uint32(len(results.PrivateNetworks)), nil
}
