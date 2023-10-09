// This file was automatically generated. DO NOT EDIT.
// If you have any remark or suggestion do not hesitate to open an issue.

// Package ipam provides methods and message types of the ipam v1alpha1 API.
package ipam

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

// API: iPAM API.
type API struct {
	client *scw.Client
}

// NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

type ListIPsRequestOrderBy string

const (
	ListIPsRequestOrderByCreatedAtDesc  = ListIPsRequestOrderBy("created_at_desc")
	ListIPsRequestOrderByCreatedAtAsc   = ListIPsRequestOrderBy("created_at_asc")
	ListIPsRequestOrderByUpdatedAtDesc  = ListIPsRequestOrderBy("updated_at_desc")
	ListIPsRequestOrderByUpdatedAtAsc   = ListIPsRequestOrderBy("updated_at_asc")
	ListIPsRequestOrderByAttachedAtDesc = ListIPsRequestOrderBy("attached_at_desc")
	ListIPsRequestOrderByAttachedAtAsc  = ListIPsRequestOrderBy("attached_at_asc")
)

func (enum ListIPsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_desc"
	}
	return string(enum)
}

func (enum ListIPsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListIPsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListIPsRequestOrderBy(ListIPsRequestOrderBy(tmp).String())
	return nil
}

type ResourceType string

const (
	ResourceTypeUnknownType         = ResourceType("unknown_type")
	ResourceTypeInstanceServer      = ResourceType("instance_server")
	ResourceTypeInstanceIP          = ResourceType("instance_ip")
	ResourceTypeInstancePrivateNic  = ResourceType("instance_private_nic")
	ResourceTypeLBServer            = ResourceType("lb_server")
	ResourceTypeFipIP               = ResourceType("fip_ip")
	ResourceTypeVpcGateway          = ResourceType("vpc_gateway")
	ResourceTypeVpcGatewayNetwork   = ResourceType("vpc_gateway_network")
	ResourceTypeK8sNode             = ResourceType("k8s_node")
	ResourceTypeRdbInstance         = ResourceType("rdb_instance")
	ResourceTypeRedisCluster        = ResourceType("redis_cluster")
	ResourceTypeBaremetalServer     = ResourceType("baremetal_server")
	ResourceTypeBaremetalPrivateNic = ResourceType("baremetal_private_nic")
)

func (enum ResourceType) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown_type"
	}
	return string(enum)
}

func (enum ResourceType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ResourceType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ResourceType(ResourceType(tmp).String())
	return nil
}

type IP struct {
	ID string `json:"id"`

	Address scw.IPNet `json:"address"`

	ProjectID string `json:"project_id"`

	IsIPv6 bool `json:"is_ipv6"`

	CreatedAt *time.Time `json:"created_at"`

	UpdatedAt *time.Time `json:"updated_at"`

	// Precisely one of Regional, SubnetID, Zonal, ZonalNat must be set.
	Regional *bool `json:"regional,omitempty"`

	// Precisely one of Regional, SubnetID, Zonal, ZonalNat must be set.
	Zonal *string `json:"zonal,omitempty"`

	// Precisely one of Regional, SubnetID, Zonal, ZonalNat must be set.
	ZonalNat *string `json:"zonal_nat,omitempty"`

	// Precisely one of Regional, SubnetID, Zonal, ZonalNat must be set.
	SubnetID *string `json:"subnet_id,omitempty"`

	Resource *Resource `json:"resource"`

	Tags []string `json:"tags"`

	Region scw.Region `json:"region"`

	Zone *scw.Zone `json:"zone"`
}

type ListIPsResponse struct {
	TotalCount uint64 `json:"total_count"`

	IPs []*IP `json:"ips"`
}

type Resource struct {
	// Type: default value: unknown_type
	Type ResourceType `json:"type"`

	ID string `json:"id"`

	MacAddress *string `json:"mac_address"`

	Name *string `json:"name"`
}

type Source struct {

	// Precisely one of PrivateNetworkID, Regional, SubnetID, Zonal, ZonalNat must be set.
	Zonal *string `json:"zonal,omitempty"`

	// Precisely one of PrivateNetworkID, Regional, SubnetID, Zonal, ZonalNat must be set.
	ZonalNat *string `json:"zonal_nat,omitempty"`

	// Precisely one of PrivateNetworkID, Regional, SubnetID, Zonal, ZonalNat must be set.
	Regional *bool `json:"regional,omitempty"`

	// Precisely one of PrivateNetworkID, Regional, SubnetID, Zonal, ZonalNat must be set.
	PrivateNetworkID *string `json:"private_network_id,omitempty"`

	// Precisely one of PrivateNetworkID, Regional, SubnetID, Zonal, ZonalNat must be set.
	SubnetID *string `json:"subnet_id,omitempty"`
}

// Service API

// Regions list localities the api is available in
func (s *API) Regions() []scw.Region {
	return []scw.Region{scw.RegionFrPar, scw.RegionNlAms, scw.RegionPlWaw}
}

type ListIPsRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	Page *int32 `json:"-"`

	PageSize *uint32 `json:"-"`
	// OrderBy: default value: created_at_desc
	OrderBy ListIPsRequestOrderBy `json:"-"`

	ProjectID *string `json:"-"`

	OrganizationID *string `json:"-"`

	Zonal *string `json:"-"`

	ZonalNat *string `json:"-"`

	Regional *bool `json:"-"`

	PrivateNetworkID *string `json:"-"`

	SubnetID *string `json:"-"`

	Attached *bool `json:"-"`

	ResourceID *string `json:"-"`
	// ResourceType: default value: unknown_type
	ResourceType ResourceType `json:"-"`

	MacAddress *string `json:"-"`

	Tags *[]string `json:"-"`

	IsIPv6 *bool `json:"-"`

	ResourceName *string `json:"-"`

	ResourceIDs []string `json:"-"`
}

// ListIPs: find IP addresses.
func (s *API) ListIPs(req *ListIPsRequest, opts ...scw.RequestOption) (*ListIPsResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "project_id", req.ProjectID)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "zonal", req.Zonal)
	parameter.AddToQuery(query, "zonal_nat", req.ZonalNat)
	parameter.AddToQuery(query, "regional", req.Regional)
	parameter.AddToQuery(query, "private_network_id", req.PrivateNetworkID)
	parameter.AddToQuery(query, "subnet_id", req.SubnetID)
	parameter.AddToQuery(query, "attached", req.Attached)
	parameter.AddToQuery(query, "resource_id", req.ResourceID)
	parameter.AddToQuery(query, "resource_type", req.ResourceType)
	parameter.AddToQuery(query, "mac_address", req.MacAddress)
	parameter.AddToQuery(query, "tags", req.Tags)
	parameter.AddToQuery(query, "is_ipv6", req.IsIPv6)
	parameter.AddToQuery(query, "resource_name", req.ResourceName)
	parameter.AddToQuery(query, "resource_ids", req.ResourceIDs)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/ipam/v1alpha1/regions/" + fmt.Sprint(req.Region) + "/ips",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListIPsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListIPsResponse) UnsafeGetTotalCount() uint64 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListIPsResponse) UnsafeAppend(res interface{}) (uint64, error) {
	results, ok := res.(*ListIPsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.IPs = append(r.IPs, results.IPs...)
	r.TotalCount += uint64(len(results.IPs))
	return uint64(len(results.IPs)), nil
}
