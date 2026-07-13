package linodego

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/linode/linodego/v2/internal/parseabletime"
)

// ObjectStorageBucket represents a ObjectStorage object
type ObjectStorageBucket struct {
	Label string `json:"label"`

	Region string `json:"region"`

	S3Endpoint   string                    `json:"s3_endpoint"`
	EndpointType ObjectStorageEndpointType `json:"endpoint_type"`
	Created      *time.Time                `json:"-"`
	Hostname     string                    `json:"hostname"`
	Objects      int                       `json:"objects"`
	Size         int                       `json:"size"`
}

type ObjectStorageBucketAccess struct {
	ACL         ObjectStorageACL `json:"acl"`
	ACLXML      string           `json:"acl_xml"`
	CorsEnabled *bool            `json:"cors_enabled"`
	CorsXML     *string          `json:"cors_xml"`
}

// ObjectStorageBucketContent holds the content of an ObjectStorageBucket
type ObjectStorageBucketContent struct {
	Data        []ObjectStorageBucketContentData `json:"data"`
	IsTruncated bool                             `json:"is_truncated"`
	NextMarker  *string                          `json:"next_marker"`
}

// ObjectStorageBucketContentData holds the data of the content of an ObjectStorageBucket
type ObjectStorageBucketContentData struct {
	Etag         string     `json:"etag"`
	LastModified *time.Time `json:"last_modified"`
	Name         string     `json:"name"`
	Owner        string     `json:"owner"`
	Size         int        `json:"size"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (i *ObjectStorageBucket) UnmarshalJSON(b []byte) error {
	type Mask ObjectStorageBucket

	p := struct {
		*Mask

		Created *parseabletime.ParseableTime `json:"created"`
	}{
		Mask: (*Mask)(i),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	i.Created = (*time.Time)(p.Created)

	return nil
}

// ObjectStorageBucketCreateOptions fields are those accepted by CreateObjectStorageBucket
type ObjectStorageBucketCreateOptions struct {
	Region string `json:"region,omitzero"`

	Label        string                    `json:"label"`
	S3Endpoint   string                    `json:"s3_endpoint,omitzero"`
	EndpointType ObjectStorageEndpointType `json:"endpoint_type,omitzero"`

	ACL         ObjectStorageACL `json:"acl,omitzero"`
	CorsEnabled *bool            `json:"cors_enabled,omitzero"`
}

// ObjectStorageBucketModifyAccessOptions fields are those accepted by ModifyObjectStorageBucketAccess
type ObjectStorageBucketModifyAccessOptions struct {
	ACL         ObjectStorageACL `json:"acl,omitzero"`
	CorsEnabled *bool            `json:"cors_enabled,omitzero"`
}

// ObjectStorageBucketUpdateAccessOptions fields are those accepted by UpdateObjectStorageBucketAccess
type ObjectStorageBucketUpdateAccessOptions struct {
	ACL         ObjectStorageACL `json:"acl,omitzero"`
	CorsEnabled *bool            `json:"cors_enabled,omitzero"`
}

// ObjectStorageBucketListContentsParams fields are the query parameters for ListObjectStorageBucketContents
type ObjectStorageBucketListContentsParams struct {
	Marker    *string
	Delimiter *string
	Prefix    *string
	PageSize  *int
}

// ObjectStorageACL options start with ACL and include all known ACL types
type ObjectStorageACL string

// ObjectStorageACL options represent the access control level of a bucket.
const (
	ACLPrivate           ObjectStorageACL = "private"
	ACLPublicRead        ObjectStorageACL = "public-read"
	ACLAuthenticatedRead ObjectStorageACL = "authenticated-read"
	ACLPublicReadWrite   ObjectStorageACL = "public-read-write"
)

// ListObjectStorageBuckets lists ObjectStorageBuckets
func (c *Client) ListObjectStorageBuckets(ctx context.Context, opts *ListOptions) ([]ObjectStorageBucket, error) {
	return getPaginatedResults[ObjectStorageBucket](ctx, c, "object-storage/buckets", opts)
}

// ListObjectStorageBucketsInRegion lists all ObjectStorageBuckets in the specified region
func (c *Client) ListObjectStorageBucketsInRegion(ctx context.Context, opts *ListOptions, regionID string) ([]ObjectStorageBucket, error) {
	return getPaginatedResults[ObjectStorageBucket](ctx, c, formatAPIPath("object-storage/buckets/%s", regionID), opts)
}

// GetObjectStorageBucket gets the ObjectStorageBucket with the provided label
func (c *Client) GetObjectStorageBucket(ctx context.Context, regionID, label string) (*ObjectStorageBucket, error) {
	e := formatAPIPath("object-storage/buckets/%s/%s", regionID, label)
	return doGETRequest[ObjectStorageBucket](ctx, c, e)
}

// CreateObjectStorageBucket creates an ObjectStorageBucket
func (c *Client) CreateObjectStorageBucket(ctx context.Context, opts ObjectStorageBucketCreateOptions) (*ObjectStorageBucket, error) {
	return doPOSTRequest[ObjectStorageBucket](ctx, c, "object-storage/buckets", opts)
}

// ModifyObjectStorageBucketAccess modifies the access configuration for an ObjectStorageBucket
func (c *Client) ModifyObjectStorageBucketAccess(ctx context.Context, regionID, label string, opts ObjectStorageBucketModifyAccessOptions) error {
	e := formatAPIPath("object-storage/buckets/%s/%s/access", regionID, label)
	return doPOSTRequestNoResponseBody(ctx, c, e, opts)
}

// UpdateObjectStorageBucketAccess updates the access configuration for an ObjectStorageBucket
func (c *Client) UpdateObjectStorageBucketAccess(ctx context.Context, regionID, label string, opts ObjectStorageBucketUpdateAccessOptions) error {
	e := formatAPIPath("object-storage/buckets/%s/%s/access", regionID, label)
	return doPUTRequestNoResponseBody(ctx, c, e, opts)
}

// GetObjectStorageBucketAccess gets the current access config for a bucket
func (c *Client) GetObjectStorageBucketAccess(ctx context.Context, regionID, label string) (*ObjectStorageBucketAccess, error) {
	e := formatAPIPath("object-storage/buckets/%s/%s/access", regionID, label)
	return doGETRequest[ObjectStorageBucketAccess](ctx, c, e)
}

// DeleteObjectStorageBucket deletes the ObjectStorageBucket with the specified label
func (c *Client) DeleteObjectStorageBucket(ctx context.Context, regionID, label string) error {
	e := formatAPIPath("object-storage/buckets/%s/%s", regionID, label)
	return doDELETERequest(ctx, c, e)
}

// ListObjectStorageBucketContents lists the contents of the specified ObjectStorageBucket
func (c *Client) ListObjectStorageBucketContents(
	ctx context.Context,
	regionID, label string,
	params *ObjectStorageBucketListContentsParams,
) (*ObjectStorageBucketContent, error) {
	basePath := formatAPIPath("object-storage/buckets/%s/%s/object-list", regionID, label)

	queryString := ""

	if params != nil {
		values, err := query.Values(params)
		if err != nil {
			return nil, fmt.Errorf("failed to encode query params: %w", err)
		}

		queryString = "?" + values.Encode()
	}

	e := basePath + queryString

	return doGETRequest[ObjectStorageBucketContent](ctx, c, e)
}
