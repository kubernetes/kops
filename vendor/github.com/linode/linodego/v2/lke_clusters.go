package linodego

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// LKEClusterStatus represents the status of an LKECluster
type LKEClusterStatus string

// LKEClusterStatus enums start with LKECluster
const (
	LKEClusterReady    LKEClusterStatus = "ready"
	LKEClusterNotReady LKEClusterStatus = "not_ready"
)

type LKEClusterStackType string

const (
	LKEClusterStackIPv4 LKEClusterStackType = "ipv4"
	LKEClusterDualStack LKEClusterStackType = "ipv4-ipv6"
)

// LKECluster represents a LKECluster object
type LKECluster struct {
	ID           int                    `json:"id"`
	Created      *time.Time             `json:"-"`
	Updated      *time.Time             `json:"-"`
	Label        string                 `json:"label"`
	Region       string                 `json:"region"`
	Status       LKEClusterStatus       `json:"status"`
	K8sVersion   string                 `json:"k8s_version"`
	Tags         []string               `json:"tags"`
	ControlPlane LKEClusterControlPlane `json:"control_plane"`

	// NOTE: Tier may not currently be available to all users and can only be used with v4beta.
	Tier string `json:"tier"`

	// NOTE: APLEnabled is currently in beta and may only function with API version v4beta.
	APLEnabled bool `json:"apl_enabled"`

	// NOTE: SubnetID, VpcID, and StackType may not currently be available to all users and can only be used with v4beta.
	SubnetID  int                 `json:"subnet_id"`
	VpcID     int                 `json:"vpc_id"`
	StackType LKEClusterStackType `json:"stack_type"`

	// NOTE: Locks can only be used with v4beta.
	Locks []LockType `json:"locks"`
}

// LKEClusterCreateOptions fields are those accepted by CreateLKECluster
type LKEClusterCreateOptions struct {
	NodePools    []LKENodePoolCreateOptions     `json:"node_pools"`
	Label        string                         `json:"label"`
	Region       string                         `json:"region"`
	K8sVersion   string                         `json:"k8s_version"`
	Tags         []string                       `json:"tags,omitzero"`
	ControlPlane *LKEClusterControlPlaneOptions `json:"control_plane,omitzero"`

	// NOTE: Tier may not currently be available to all users and can only be used with v4beta.
	Tier string `json:"tier,omitzero"`

	// NOTE: APLEnabled is currently in beta and may only function with API version v4beta.
	APLEnabled bool `json:"apl_enabled,omitzero"`

	// NOTE: SubnetID, VpcID, and StackType may not currently be available to all users and can only be used with v4beta.
	SubnetID  *int                 `json:"subnet_id,omitzero"`
	VpcID     *int                 `json:"vpc_id,omitzero"`
	StackType *LKEClusterStackType `json:"stack_type,omitzero"`
}

// LKEClusterUpdateOptions fields are those accepted by UpdateLKECluster
type LKEClusterUpdateOptions struct {
	K8sVersion   string                         `json:"k8s_version,omitzero"`
	Label        string                         `json:"label,omitzero"`
	Tags         []string                       `json:"tags,omitzero"`
	ControlPlane *LKEClusterControlPlaneOptions `json:"control_plane,omitzero"`
}

// LKEClusterAPIEndpoint fields are those returned by ListLKEClusterAPIEndpoints
type LKEClusterAPIEndpoint struct {
	Endpoint string `json:"endpoint"`
}

// LKEClusterKubeconfig fields are those returned by GetLKEClusterKubeconfig
type LKEClusterKubeconfig struct {
	KubeConfig string `json:"kubeconfig"` // Base64-encoded Kubeconfig file for this Cluster.
}

// LKEVersion fields are those returned by GetLKEVersion
type LKEVersion struct {
	ID string `json:"id"`
}

// LKETierVersion fields are those returned by GetLKETierVersion
// NOTE: It may not currently be available to all users and can only be used with v4beta.
type LKETierVersion struct {
	ID   string         `json:"id"`
	Tier LKEVersionTier `json:"tier"`
}

// LKEVersionTier enums represents different LKE tiers
type LKEVersionTier string

// LKEVersionTier enums start with LKEVersion
const (
	LKEVersionStandard   LKEVersionTier = "standard"
	LKEVersionEnterprise LKEVersionTier = "enterprise"
)

// LKEClusterRegenerateOptions fields are those accepted by RegenerateLKECluster
type LKEClusterRegenerateOptions struct {
	KubeConfig   bool `json:"kubeconfig"`
	ServiceToken bool `json:"servicetoken"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (i *LKECluster) UnmarshalJSON(b []byte) error {
	type Mask LKECluster

	p := struct {
		*Mask

		Created *parseabletime.ParseableTime `json:"created"`
		Updated *parseabletime.ParseableTime `json:"updated"`
	}{
		Mask: (*Mask)(i),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	i.Created = (*time.Time)(p.Created)
	i.Updated = (*time.Time)(p.Updated)

	return nil
}

// GetCreateOptions converts a LKECluster to LKEClusterCreateOptions for use in CreateLKECluster
func (i LKECluster) GetCreateOptions() (o LKEClusterCreateOptions) {
	o.Label = i.Label
	o.Region = i.Region
	o.K8sVersion = i.K8sVersion
	o.Tags = i.Tags

	isHA := i.ControlPlane.HighAvailability

	o.ControlPlane = &LKEClusterControlPlaneOptions{
		HighAvailability: &isHA,
		// ACL will not be populated in the control plane response
	}

	// @TODO copy NodePools?
	return o
}

// GetUpdateOptions converts a LKECluster to LKEClusterUpdateOptions for use in UpdateLKECluster
func (i LKECluster) GetUpdateOptions() (o LKEClusterUpdateOptions) {
	o.K8sVersion = i.K8sVersion
	o.Label = i.Label
	o.Tags = i.Tags

	isHA := i.ControlPlane.HighAvailability

	o.ControlPlane = &LKEClusterControlPlaneOptions{
		HighAvailability: &isHA,
		// ACL will not be populated in the control plane response
	}

	return o
}

// ListLKEVersions lists the Kubernetes versions available through LKE. This endpoint is cached by default.
func (c *Client) ListLKEVersions(ctx context.Context, opts *ListOptions) ([]LKEVersion, error) {
	e := "lke/versions"

	endpoint, err := generateListCacheURL(e, opts)
	if err != nil {
		return nil, err
	}

	if result := c.getCachedResponse(endpoint); result != nil {
		return result.([]LKEVersion), nil
	}

	response, err := getPaginatedResults[LKEVersion](ctx, c, e, opts)
	if err != nil {
		return nil, err
	}

	c.addCachedResponse(endpoint, response, &cacheExpiryTime)

	return response, nil
}

// GetLKEVersion gets details about a specific LKE Version. This endpoint is cached by default.
func (c *Client) GetLKEVersion(ctx context.Context, version string) (*LKEVersion, error) {
	e := formatAPIPath("lke/versions/%s", version)

	if result := c.getCachedResponse(e); result != nil {
		result := result.(LKEVersion)
		return &result, nil
	}

	response, err := doGETRequest[LKEVersion](ctx, c, e)
	if err != nil {
		return nil, err
	}

	c.addCachedResponse(e, response, &cacheExpiryTime)

	return response, nil
}

// ListLKETierVersions lists all Kubernetes versions available given tier through LKE.
// NOTE: This endpoint may not currently be available to all users and can only be used with v4beta.
func (c *Client) ListLKETierVersions(ctx context.Context, tier string, opts *ListOptions) ([]LKETierVersion, error) {
	return getPaginatedResults[LKETierVersion](ctx, c, formatAPIPath("lke/tiers/%s/versions", tier), opts)
}

// GetLKETierVersion gets the details of a specific LKE tier version.
// NOTE: This endpoint may not currently be available to all users and can only be used with v4beta.
func (c *Client) GetLKETierVersion(ctx context.Context, tier string, versionID string) (*LKETierVersion, error) {
	return doGETRequest[LKETierVersion](ctx, c, formatAPIPath("lke/tiers/%s/versions/%s", tier, versionID))
}

// ListLKEClusterAPIEndpoints gets the API Endpoint for the LKE Cluster specified
func (c *Client) ListLKEClusterAPIEndpoints(ctx context.Context, clusterID int, opts *ListOptions) ([]LKEClusterAPIEndpoint, error) {
	return getPaginatedResults[LKEClusterAPIEndpoint](ctx, c, formatAPIPath("lke/clusters/%d/api-endpoints", clusterID), opts)
}

// ListLKEClusters lists LKEClusters
func (c *Client) ListLKEClusters(ctx context.Context, opts *ListOptions) ([]LKECluster, error) {
	return getPaginatedResults[LKECluster](ctx, c, "lke/clusters", opts)
}

// GetLKECluster gets the lkeCluster with the provided ID
func (c *Client) GetLKECluster(ctx context.Context, clusterID int) (*LKECluster, error) {
	e := formatAPIPath("lke/clusters/%d", clusterID)
	return doGETRequest[LKECluster](ctx, c, e)
}

// CreateLKECluster creates a LKECluster
func (c *Client) CreateLKECluster(ctx context.Context, opts LKEClusterCreateOptions) (*LKECluster, error) {
	return doPOSTRequest[LKECluster](ctx, c, "lke/clusters", opts)
}

// UpdateLKECluster updates the LKECluster with the specified id
func (c *Client) UpdateLKECluster(ctx context.Context, clusterID int, opts LKEClusterUpdateOptions) (*LKECluster, error) {
	e := formatAPIPath("lke/clusters/%d", clusterID)
	return doPUTRequest[LKECluster](ctx, c, e, opts)
}

// DeleteLKECluster deletes the LKECluster with the specified id
func (c *Client) DeleteLKECluster(ctx context.Context, clusterID int) error {
	e := formatAPIPath("lke/clusters/%d", clusterID)
	return doDELETERequest(ctx, c, e)
}

// GetLKEClusterKubeconfig gets the Kubeconfig for the LKE Cluster specified
func (c *Client) GetLKEClusterKubeconfig(ctx context.Context, clusterID int) (*LKEClusterKubeconfig, error) {
	e := formatAPIPath("lke/clusters/%d/kubeconfig", clusterID)
	return doGETRequest[LKEClusterKubeconfig](ctx, c, e)
}

// DeleteLKEClusterKubeconfig deletes the Kubeconfig for the LKE Cluster specified
func (c *Client) DeleteLKEClusterKubeconfig(ctx context.Context, clusterID int) error {
	e := formatAPIPath("lke/clusters/%d/kubeconfig", clusterID)
	return doDELETERequest(ctx, c, e)
}

// RecycleLKEClusterNodes recycles all nodes in all pools of the specified LKE Cluster.
func (c *Client) RecycleLKEClusterNodes(ctx context.Context, clusterID int) error {
	e := formatAPIPath("lke/clusters/%d/recycle", clusterID)
	return doPOSTRequestNoRequestResponseBody(ctx, c, e)
}

// RegenerateLKECluster regenerates the Kubeconfig file and/or the service account token for the specified LKE Cluster.
func (c *Client) RegenerateLKECluster(ctx context.Context, clusterID int, opts LKEClusterRegenerateOptions) (*LKECluster, error) {
	e := formatAPIPath("lke/clusters/%d/regenerate", clusterID)
	return doPOSTRequest[LKECluster](ctx, c, e, opts)
}

// DeleteLKEClusterServiceToken deletes and regenerate the service account token for a Cluster.
func (c *Client) DeleteLKEClusterServiceToken(ctx context.Context, clusterID int) error {
	e := formatAPIPath("lke/clusters/%d/servicetoken", clusterID)
	return doDELETERequest(ctx, c, e)
}

// GetLKEClusterAPLConsoleURL gets the URL of this cluster's APL installation if this cluster is APL-enabled.
func (c *Client) GetLKEClusterAPLConsoleURL(ctx context.Context, clusterID int) (string, error) {
	cluster, err := c.GetLKECluster(ctx, clusterID)
	if err != nil {
		return "", err
	}

	if cluster.APLEnabled {
		return fmt.Sprintf("https://console.lke%d.akamai-apl.net", cluster.ID), nil
	}

	return "", nil
}

// GetLKEClusterAPLHealthCheckURL gets the URL of this cluster's APL health check endpoint if this cluster is APL-enabled.
func (c *Client) GetLKEClusterAPLHealthCheckURL(ctx context.Context, clusterID int) (string, error) {
	cluster, err := c.GetLKECluster(ctx, clusterID)
	if err != nil {
		return "", err
	}

	if cluster.APLEnabled {
		return fmt.Sprintf("https://auth.lke%d.akamai-apl.net/ready", cluster.ID), nil
	}

	return "", nil
}
