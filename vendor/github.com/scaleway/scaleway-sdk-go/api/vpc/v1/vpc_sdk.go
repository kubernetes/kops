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

// API: vPC API
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
	// ListPrivateNetworksRequestOrderByCreatedAtAsc is [insert doc].
	ListPrivateNetworksRequestOrderByCreatedAtAsc = ListPrivateNetworksRequestOrderBy("created_at_asc")
	// ListPrivateNetworksRequestOrderByCreatedAtDesc is [insert doc].
	ListPrivateNetworksRequestOrderByCreatedAtDesc = ListPrivateNetworksRequestOrderBy("created_at_desc")
	// ListPrivateNetworksRequestOrderByNameAsc is [insert doc].
	ListPrivateNetworksRequestOrderByNameAsc = ListPrivateNetworksRequestOrderBy("name_asc")
	// ListPrivateNetworksRequestOrderByNameDesc is [insert doc].
	ListPrivateNetworksRequestOrderByNameDesc = ListPrivateNetworksRequestOrderBy("name_desc")
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

// PrivateNetwork: private network
type PrivateNetwork struct {
	// ID: the private network ID
	ID string `json:"id"`
	// Name: the private network name
	Name string `json:"name"`
	// OrganizationID: the private network organization
	OrganizationID string `json:"organization_id"`
	// ProjectID: the private network project ID
	ProjectID string `json:"project_id"`
	// Zone: the zone in which the private network is available
	Zone scw.Zone `json:"zone"`
	// Tags: the private network tags
	Tags []string `json:"tags"`
	// CreatedAt: the private network creation date
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: the last private network modification date
	UpdatedAt *time.Time `json:"updated_at"`
	// Subnets: private network subnets CIDR
	Subnets []scw.IPNet `json:"subnets"`
}

// Service API

type ListPrivateNetworksRequest struct {
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// OrderBy: the sort order of the returned private networks
	//
	// Default value: created_at_asc
	OrderBy ListPrivateNetworksRequestOrderBy `json:"-"`
	// Page: the page number for the returned private networks
	Page *int32 `json:"-"`
	// PageSize: the maximum number of private networks per page
	PageSize *uint32 `json:"-"`
	// Name: filter private networks with names containing this string
	Name *string `json:"-"`
	// Tags: filter private networks with one or more matching tags
	Tags []string `json:"-"`
	// OrganizationID: the organization ID on which to filter the returned private networks
	OrganizationID *string `json:"-"`
	// ProjectID: the project ID on which to filter the returned private networks
	ProjectID *string `json:"-"`
}

// ListPrivateNetworks: list private networks
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
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// Name: the name of the private network
	Name string `json:"name"`
	// ProjectID: the project ID of the private network
	ProjectID string `json:"project_id"`
	// Tags: the private networks tags
	Tags []string `json:"tags"`
	// Subnets: private network subnets CIDR
	Subnets []scw.IPNet `json:"subnets"`
}

// CreatePrivateNetwork: create a private network
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
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// PrivateNetworkID: the private network id
	PrivateNetworkID string `json:"-"`
}

// GetPrivateNetwork: get a private network
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
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// PrivateNetworkID: the private network ID
	PrivateNetworkID string `json:"-"`
	// Name: the name of the private network
	Name *string `json:"name"`
	// Tags: the private networks tags
	Tags *[]string `json:"tags"`
	// Subnets: private network subnets CIDR
	Subnets *[]string `json:"subnets"`
}

// UpdatePrivateNetwork: update private network
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
	// Zone:
	//
	// Zone to target. If none is passed will use default zone from the config
	Zone scw.Zone `json:"-"`
	// PrivateNetworkID: the private network ID
	PrivateNetworkID string `json:"-"`
}

// DeletePrivateNetwork: delete a private network
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
