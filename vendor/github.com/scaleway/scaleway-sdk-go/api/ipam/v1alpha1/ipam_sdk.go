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

func (enum ListIPsRequestOrderBy) Values() []ListIPsRequestOrderBy {
	return []ListIPsRequestOrderBy{
		"created_at_desc",
		"created_at_asc",
		"updated_at_desc",
		"updated_at_asc",
		"attached_at_desc",
		"attached_at_asc",
	}
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
	ResourceTypeCustom              = ResourceType("custom")
	ResourceTypeInstanceServer      = ResourceType("instance_server")
	ResourceTypeInstanceIP          = ResourceType("instance_ip")
	ResourceTypeInstancePrivateNic  = ResourceType("instance_private_nic")
	ResourceTypeLBServer            = ResourceType("lb_server")
	ResourceTypeFipIP               = ResourceType("fip_ip")
	ResourceTypeVpcGateway          = ResourceType("vpc_gateway")
	ResourceTypeVpcGatewayNetwork   = ResourceType("vpc_gateway_network")
	ResourceTypeK8sNode             = ResourceType("k8s_node")
	ResourceTypeK8sCluster          = ResourceType("k8s_cluster")
	ResourceTypeRdbInstance         = ResourceType("rdb_instance")
	ResourceTypeRedisCluster        = ResourceType("redis_cluster")
	ResourceTypeBaremetalServer     = ResourceType("baremetal_server")
	ResourceTypeBaremetalPrivateNic = ResourceType("baremetal_private_nic")
	ResourceTypeLlmDeployment       = ResourceType("llm_deployment")
)

func (enum ResourceType) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown_type"
	}
	return string(enum)
}

func (enum ResourceType) Values() []ResourceType {
	return []ResourceType{
		"unknown_type",
		"custom",
		"instance_server",
		"instance_ip",
		"instance_private_nic",
		"lb_server",
		"fip_ip",
		"vpc_gateway",
		"vpc_gateway_network",
		"k8s_node",
		"k8s_cluster",
		"rdb_instance",
		"redis_cluster",
		"baremetal_server",
		"baremetal_private_nic",
		"llm_deployment",
	}
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

// Resource: resource.
type Resource struct {
	// Type: default value: unknown_type
	Type ResourceType `json:"type"`

	ID string `json:"id"`

	MacAddress *string `json:"mac_address"`

	Name *string `json:"name"`
}

// Reverse: reverse.
type Reverse struct {
	Hostname string `json:"hostname"`

	Address *net.IP `json:"address"`
}

// IP: ip.
type IP struct {
	ID string `json:"id"`

	Address scw.IPNet `json:"address"`

	ProjectID string `json:"project_id"`

	IsIPv6 bool `json:"is_ipv6"`

	CreatedAt *time.Time `json:"created_at"`

	UpdatedAt *time.Time `json:"updated_at"`

	// Precisely one of Regional, Zonal, ZonalNat, SubnetID must be set.
	Regional *bool `json:"regional,omitempty"`

	// Precisely one of Regional, Zonal, ZonalNat, SubnetID must be set.
	Zonal *string `json:"zonal,omitempty"`

	// Precisely one of Regional, Zonal, ZonalNat, SubnetID must be set.
	ZonalNat *string `json:"zonal_nat,omitempty"`

	// Precisely one of Regional, Zonal, ZonalNat, SubnetID must be set.
	SubnetID *string `json:"subnet_id,omitempty"`

	Resource *Resource `json:"resource"`

	Tags []string `json:"tags"`

	Reverses []*Reverse `json:"reverses"`

	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"region"`

	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone *scw.Zone `json:"zone"`
}

// ListIPsRequest: list i ps request.
type ListIPsRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	Page *int32 `json:"-"`

	PageSize *uint32 `json:"-"`

	// OrderBy: default value: created_at_desc
	OrderBy ListIPsRequestOrderBy `json:"-"`

	ProjectID *string `json:"-"`

	OrganizationID *string `json:"-"`

	// Precisely one of Zonal, ZonalNat, Regional, PrivateNetworkID, SubnetID must be set.
	Zonal *string `json:"zonal,omitempty"`

	// Precisely one of Zonal, ZonalNat, Regional, PrivateNetworkID, SubnetID must be set.
	ZonalNat *string `json:"zonal_nat,omitempty"`

	// Precisely one of Zonal, ZonalNat, Regional, PrivateNetworkID, SubnetID must be set.
	Regional *bool `json:"regional,omitempty"`

	// Precisely one of Zonal, ZonalNat, Regional, PrivateNetworkID, SubnetID must be set.
	PrivateNetworkID *string `json:"private_network_id,omitempty"`

	// Precisely one of Zonal, ZonalNat, Regional, PrivateNetworkID, SubnetID must be set.
	SubnetID *string `json:"subnet_id,omitempty"`

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

// ListIPsResponse: list i ps response.
type ListIPsResponse struct {
	TotalCount uint64 `json:"total_count"`

	IPs []*IP `json:"ips"`
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

// This API allows you to manage your Scaleway IP addresses with our IP Address Management tool.
type API struct {
	client *scw.Client
}

// NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}
func (s *API) Regions() []scw.Region {
	return []scw.Region{scw.RegionFrPar, scw.RegionNlAms, scw.RegionPlWaw}
}

// ListIPs: Find IP addresses.
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
	parameter.AddToQuery(query, "attached", req.Attached)
	parameter.AddToQuery(query, "resource_id", req.ResourceID)
	parameter.AddToQuery(query, "resource_type", req.ResourceType)
	parameter.AddToQuery(query, "mac_address", req.MacAddress)
	parameter.AddToQuery(query, "tags", req.Tags)
	parameter.AddToQuery(query, "is_ipv6", req.IsIPv6)
	parameter.AddToQuery(query, "resource_name", req.ResourceName)
	parameter.AddToQuery(query, "resource_ids", req.ResourceIDs)
	parameter.AddToQuery(query, "zonal", req.Zonal)
	parameter.AddToQuery(query, "zonal_nat", req.ZonalNat)
	parameter.AddToQuery(query, "regional", req.Regional)
	parameter.AddToQuery(query, "private_network_id", req.PrivateNetworkID)
	parameter.AddToQuery(query, "subnet_id", req.SubnetID)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/ipam/v1alpha1/regions/" + fmt.Sprint(req.Region) + "/ips",
		Query:  query,
	}

	var resp ListIPsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
