package linodego

import (
	"context"
)

type ObjectStorageKeyRegion struct {
	ID           string                    `json:"id"`
	S3Endpoint   string                    `json:"s3_endpoint"`
	EndpointType ObjectStorageEndpointType `json:"endpoint_type"`
}

// ObjectStorageKey represents a linode object storage key object
type ObjectStorageKey struct {
	ID           int                             `json:"id"`
	Label        string                          `json:"label"`
	AccessKey    string                          `json:"access_key"`
	SecretKey    string                          `json:"secret_key"`
	Limited      bool                            `json:"limited"`
	BucketAccess *[]ObjectStorageKeyBucketAccess `json:"bucket_access"`
	Regions      []ObjectStorageKeyRegion        `json:"regions"`
}

// ObjectStorageKeyBucketAccess represents a linode limited object storage key's bucket access
type ObjectStorageKeyBucketAccess struct {
	Region string `json:"region,omitzero"`

	BucketName  string `json:"bucket_name"`
	Permissions string `json:"permissions"`
}

type ObjectStorageKeyBucketAccessCreateOptions struct {
	Region string `json:"region,omitzero"`

	BucketName  string `json:"bucket_name"`
	Permissions string `json:"permissions"`
}

// ObjectStorageKeyCreateOptions fields are those accepted by CreateObjectStorageKey
type ObjectStorageKeyCreateOptions struct {
	Label        string                                      `json:"label"`
	BucketAccess []ObjectStorageKeyBucketAccessCreateOptions `json:"bucket_access,omitzero"`
	Regions      []string                                    `json:"regions,omitzero"`
}

// ObjectStorageKeyUpdateOptions fields are those accepted by UpdateObjectStorageKey
type ObjectStorageKeyUpdateOptions struct {
	Label   string   `json:"label,omitzero"`
	Regions []string `json:"regions,omitzero"`
}

// ListObjectStorageKeys lists ObjectStorageKeys
func (c *Client) ListObjectStorageKeys(ctx context.Context, opts *ListOptions) ([]ObjectStorageKey, error) {
	return getPaginatedResults[ObjectStorageKey](ctx, c, "object-storage/keys", opts)
}

// CreateObjectStorageKey creates a ObjectStorageKey
func (c *Client) CreateObjectStorageKey(ctx context.Context, opts ObjectStorageKeyCreateOptions) (*ObjectStorageKey, error) {
	return doPOSTRequest[ObjectStorageKey](ctx, c, "object-storage/keys", opts)
}

// GetObjectStorageKey gets the object storage key with the provided ID
func (c *Client) GetObjectStorageKey(ctx context.Context, keyID int) (*ObjectStorageKey, error) {
	e := formatAPIPath("object-storage/keys/%d", keyID)
	return doGETRequest[ObjectStorageKey](ctx, c, e)
}

// UpdateObjectStorageKey updates the object storage key with the specified id
func (c *Client) UpdateObjectStorageKey(ctx context.Context, keyID int, opts ObjectStorageKeyUpdateOptions) (*ObjectStorageKey, error) {
	e := formatAPIPath("object-storage/keys/%d", keyID)
	return doPUTRequest[ObjectStorageKey](ctx, c, e, opts)
}

// DeleteObjectStorageKey deletes the ObjectStorageKey with the specified id
func (c *Client) DeleteObjectStorageKey(ctx context.Context, keyID int) error {
	e := formatAPIPath("object-storage/keys/%d", keyID)
	return doDELETERequest(ctx, c, e)
}
