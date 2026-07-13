package linodego

import "context"

// LKEClusterControlPlane fields contained within the `control_plane` attribute of an LKE cluster.
type LKEClusterControlPlane struct {
	HighAvailability bool `json:"high_availability"`

	// AuditLogsEnabled may not currently be available to all users and can only be used with v4beta.
	AuditLogsEnabled bool `json:"audit_logs_enabled,omitzero"`
}

// LKEClusterControlPlaneACLAddresses describes the
// allowed IP ranges for an LKE cluster's control plane.
type LKEClusterControlPlaneACLAddresses struct {
	IPv4 []string `json:"ipv4"`
	IPv6 []string `json:"ipv6"`
}

// LKEClusterControlPlaneACL describes the ACL configuration
// for an LKE cluster's control plane.
type LKEClusterControlPlaneACL struct {
	Enabled    bool                                `json:"enabled"`
	Addresses  *LKEClusterControlPlaneACLAddresses `json:"addresses"`
	RevisionID string                              `json:"revision-id"`
}

// LKEClusterControlPlaneACLAddressesOptions are the options used to
// specify the allowed IP ranges for an LKE cluster's control plane.
type LKEClusterControlPlaneACLAddressesOptions struct {
	IPv4 []string `json:"ipv4,omitzero"`
	IPv6 []string `json:"ipv6,omitzero"`
}

// LKEClusterControlPlaneACLOptions represents the options used when
// configuring an LKE cluster's control plane ACL policy.
type LKEClusterControlPlaneACLOptions struct {
	Enabled    *bool                                      `json:"enabled,omitzero"`
	Addresses  *LKEClusterControlPlaneACLAddressesOptions `json:"addresses,omitzero"`
	RevisionID string                                     `json:"revision-id,omitzero"`
}

// LKEClusterControlPlaneOptions represents the options used when
// configuring an LKE cluster's control plane.
type LKEClusterControlPlaneOptions struct {
	HighAvailability *bool                             `json:"high_availability,omitzero"`
	ACL              *LKEClusterControlPlaneACLOptions `json:"acl,omitzero"`

	// AuditLogsEnabled may not currently be available to all users and can only be used with v4beta.
	AuditLogsEnabled *bool `json:"audit_logs_enabled,omitzero"`
}

// LKEClusterControlPlaneACLUpdateOptions represents the options
// available when updating the ACL configuration of an LKE cluster's
// control plane.
type LKEClusterControlPlaneACLUpdateOptions struct {
	ACL LKEClusterControlPlaneACLOptions `json:"acl"`
}

// LKEClusterControlPlaneACLResponse represents the response structure
// for the Client.GetLKEClusterControlPlaneACL(...) method.
type LKEClusterControlPlaneACLResponse struct {
	ACL LKEClusterControlPlaneACL `json:"acl"`
}

// GetLKEClusterControlPlaneACL gets the ACL configuration for the
// given cluster's control plane.
func (c *Client) GetLKEClusterControlPlaneACL(ctx context.Context, clusterID int) (*LKEClusterControlPlaneACLResponse, error) {
	return doGETRequest[LKEClusterControlPlaneACLResponse](
		ctx,
		c,
		formatAPIPath("lke/clusters/%d/control_plane_acl", clusterID),
	)
}

// UpdateLKEClusterControlPlaneACL updates the ACL configuration for the
// given cluster's control plane.
func (c *Client) UpdateLKEClusterControlPlaneACL(
	ctx context.Context,
	clusterID int,
	opts LKEClusterControlPlaneACLUpdateOptions,
) (*LKEClusterControlPlaneACLResponse, error) {
	return doPUTRequest[LKEClusterControlPlaneACLResponse](
		ctx,
		c,
		formatAPIPath("lke/clusters/%d/control_plane_acl", clusterID),
		opts,
	)
}

// DeleteLKEClusterControlPlaneACL deletes the ACL configuration for the
// given cluster's control plane.
func (c *Client) DeleteLKEClusterControlPlaneACL(
	ctx context.Context,
	clusterID int,
) error {
	return doDELETERequest(
		ctx,
		c,
		formatAPIPath("lke/clusters/%d/control_plane_acl", clusterID),
	)
}
