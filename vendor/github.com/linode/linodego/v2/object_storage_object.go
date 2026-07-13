package linodego

import (
	"context"
)

type ObjectStorageObjectURLCreateOptions struct {
	Name               string `json:"name"`
	Method             string `json:"method"`
	ContentType        string `json:"content_type,omitzero"`
	ContentDisposition string `json:"content_disposition,omitzero"`
	ExpiresIn          *int   `json:"expires_in,omitzero"`
}

type ObjectStorageObjectURL struct {
	URL    string `json:"url"`
	Exists bool   `json:"exists"`
}

type ObjectStorageObjectACLConfig struct {
	ACL    *string `json:"acl"`
	ACLXML *string `json:"acl_xml"`
}

type ObjectStorageObjectACLConfigUpdateOptions struct {
	Name string `json:"name"`
	ACL  string `json:"acl"`
}

func (c *Client) CreateObjectStorageObjectURL(
	ctx context.Context,
	regionID, label string,
	opts ObjectStorageObjectURLCreateOptions,
) (*ObjectStorageObjectURL, error) {
	e := formatAPIPath("object-storage/buckets/%s/%s/object-url", regionID, label)
	return doPOSTRequest[ObjectStorageObjectURL](ctx, c, e, opts)
}

func (c *Client) GetObjectStorageObjectACLConfig(ctx context.Context, regionID, label, object string) (*ObjectStorageObjectACLConfig, error) {
	e := formatAPIPath("object-storage/buckets/%s/%s/object-acl?name=%s", regionID, label, object)
	return doGETRequest[ObjectStorageObjectACLConfig](ctx, c, e)
}

func (c *Client) UpdateObjectStorageObjectACLConfig(
	ctx context.Context,
	regionID, label string,
	opts ObjectStorageObjectACLConfigUpdateOptions,
) (*ObjectStorageObjectACLConfig, error) {
	e := formatAPIPath("object-storage/buckets/%s/%s/object-acl", regionID, label)
	return doPUTRequest[ObjectStorageObjectACLConfig](ctx, c, e, opts)
}
