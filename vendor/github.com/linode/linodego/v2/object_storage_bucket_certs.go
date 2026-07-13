package linodego

import (
	"context"
)

type ObjectStorageBucketCert struct {
	SSL *bool `json:"ssl"`
}

type ObjectStorageBucketCertUploadOptions struct {
	Certificate string `json:"certificate"`
	PrivateKey  string `json:"private_key"`
}

// UploadObjectStorageBucketCert uploads a TLS/SSL Cert to be used with an Object Storage Bucket.
func (c *Client) UploadObjectStorageBucketCert(
	ctx context.Context,
	regionID, bucket string,
	opts ObjectStorageBucketCertUploadOptions,
) (*ObjectStorageBucketCert, error) {
	e := formatAPIPath("object-storage/buckets/%s/%s/ssl", regionID, bucket)
	return doPOSTRequest[ObjectStorageBucketCert](ctx, c, e, opts)
}

// GetObjectStorageBucketCert gets an ObjectStorageBucketCert
func (c *Client) GetObjectStorageBucketCert(ctx context.Context, regionID, bucket string) (*ObjectStorageBucketCert, error) {
	e := formatAPIPath("object-storage/buckets/%s/%s/ssl", regionID, bucket)
	return doGETRequest[ObjectStorageBucketCert](ctx, c, e)
}

// DeleteObjectStorageBucketCert deletes an ObjectStorageBucketCert
func (c *Client) DeleteObjectStorageBucketCert(ctx context.Context, regionID, bucket string) error {
	e := formatAPIPath("object-storage/buckets/%s/%s/ssl", regionID, bucket)
	return doDELETERequest(ctx, c, e)
}
