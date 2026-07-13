package linodego

import (
	"context"
)

// ObjectStorageQuota represents a Object Storage related quota information on your account.
type ObjectStorageQuota struct {
	QuotaID        string `json:"quota_id"`
	QuotaName      string `json:"quota_name"`
	EndpointType   string `json:"endpoint_type"`
	S3Endpoint     string `json:"s3_endpoint"`
	Description    string `json:"description"`
	QuotaLimit     int    `json:"quota_limit"`
	ResourceMetric string `json:"resource_metric"`
	QuotaType      string `json:"quota_type"`
	HasUsage       bool   `json:"has_usage"`
}

// ObjectStorageQuotaUsage is the usage data for a specific Object Storage related quota on your account.
type ObjectStorageQuotaUsage struct {
	QuotaLimit int  `json:"quota_limit"`
	Usage      *int `json:"usage"`
}

// ObjectStorageGlobalQuota represents global/account-level Object Storage quota information.
type ObjectStorageGlobalQuota struct {
	QuotaID        string `json:"quota_id"`
	QuotaName      string `json:"quota_name"`
	QuotaType      string `json:"quota_type"`
	Description    string `json:"description"`
	QuotaLimit     int    `json:"quota_limit"`
	ResourceMetric string `json:"resource_metric"`
	HasUsage       bool   `json:"has_usage"`
}

// ObjectStorageGlobalQuotaUsage is the usage data for a specific global/account-level Object Storage quota.
type ObjectStorageGlobalQuotaUsage struct {
	QuotaLimit int  `json:"quota_limit"`
	Usage      *int `json:"usage"`
}

// ListObjectStorageQuotas lists the active ObjectStorage-related quotas applied to your account.
func (c *Client) ListObjectStorageQuotas(ctx context.Context, opts *ListOptions) ([]ObjectStorageQuota, error) {
	return getPaginatedResults[ObjectStorageQuota](ctx, c, formatAPIPath("object-storage/quotas"), opts)
}

// GetObjectStorageQuota gets information about a specific ObjectStorage-related quota on your account.
func (c *Client) GetObjectStorageQuota(ctx context.Context, quotaID string) (*ObjectStorageQuota, error) {
	e := formatAPIPath("object-storage/quotas/%s", quotaID)
	return doGETRequest[ObjectStorageQuota](ctx, c, e)
}

// GetObjectStorageQuotaUsage gets usage data for a specific ObjectStorage Quota resource you can have on your account and the current usage for that resource.
func (c *Client) GetObjectStorageQuotaUsage(ctx context.Context, quotaID string) (*ObjectStorageQuotaUsage, error) {
	e := formatAPIPath("object-storage/quotas/%s/usage", quotaID)
	return doGETRequest[ObjectStorageQuotaUsage](ctx, c, e)
}

// ListObjectStorageGlobalQuotas lists the global/account-level ObjectStorage-related quotas applied to your account.
func (c *Client) ListObjectStorageGlobalQuotas(ctx context.Context, opts *ListOptions) ([]ObjectStorageGlobalQuota, error) {
	return getPaginatedResults[ObjectStorageGlobalQuota](ctx, c, formatAPIPath("object-storage/global-quotas"), opts)
}

// GetObjectStorageGlobalQuota gets information about a specific global/account-level ObjectStorage-related quota on your account.
func (c *Client) GetObjectStorageGlobalQuota(ctx context.Context, quotaID string) (*ObjectStorageGlobalQuota, error) {
	e := formatAPIPath("object-storage/global-quotas/%s", quotaID)
	return doGETRequest[ObjectStorageGlobalQuota](ctx, c, e)
}

// GetObjectStorageGlobalQuotaUsage gets usage data for a specific global/account-level ObjectStorage quota resource.
func (c *Client) GetObjectStorageGlobalQuotaUsage(ctx context.Context, quotaID string) (*ObjectStorageGlobalQuotaUsage, error) {
	e := formatAPIPath("object-storage/global-quotas/%s/usage", quotaID)
	return doGETRequest[ObjectStorageGlobalQuotaUsage](ctx, c, e)
}
