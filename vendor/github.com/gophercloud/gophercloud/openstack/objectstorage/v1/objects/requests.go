package objects

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/objectstorage/v1/accounts"
	"github.com/gophercloud/gophercloud/openstack/objectstorage/v1/containers"
	"github.com/gophercloud/gophercloud/pagination"
)

// ErrTempURLKeyNotFound is an error indicating that the Temp URL key was
// neigther set nor resolved from a container or account metadata.
type ErrTempURLKeyNotFound struct{ gophercloud.ErrMissingInput }

func (e ErrTempURLKeyNotFound) Error() string {
	return "Unable to obtain the Temp URL key."
}

// ErrTempURLDigestNotValid is an error indicating that the requested
// cryptographic hash function is not supported.
type ErrTempURLDigestNotValid struct {
	gophercloud.ErrMissingInput
	Digest string
}

func (e ErrTempURLDigestNotValid) Error() string {
	return fmt.Sprintf("The requested %q digest is not supported.", e.Digest)
}

// ListOptsBuilder allows extensions to add additional parameters to the List
// request.
type ListOptsBuilder interface {
	ToObjectListParams() (bool, string, error)
}

// ListOpts is a structure that holds parameters for listing objects.
type ListOpts struct {
	// Full is a true/false value that represents the amount of object information
	// returned. If Full is set to true, then the content-type, number of bytes,
	// hash date last modified, and name are returned. If set to false or not set,
	// then only the object names are returned.
	Full      bool
	Limit     int    `q:"limit"`
	Marker    string `q:"marker"`
	EndMarker string `q:"end_marker"`
	Format    string `q:"format"`
	Prefix    string `q:"prefix"`
	Delimiter string `q:"delimiter"`
	Path      string `q:"path"`
	Versions  bool   `q:"versions"`
}

// ToObjectListParams formats a ListOpts into a query string and boolean
// representing whether to list complete information for each object.
func (opts ListOpts) ToObjectListParams() (bool, string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	return opts.Full, q.String(), err
}

// List is a function that retrieves all objects in a container. It also returns
// the details for the container. To extract only the object information or names,
// pass the ListResult response to the ExtractInfo or ExtractNames function,
// respectively.
func List(c *gophercloud.ServiceClient, containerName string, opts ListOptsBuilder) pagination.Pager {
	url, err := listURL(c, containerName)
	if err != nil {
		return pagination.Pager{Err: err}
	}

	headers := map[string]string{"Accept": "text/plain", "Content-Type": "text/plain"}
	if opts != nil {
		full, query, err := opts.ToObjectListParams()
		if err != nil {
			return pagination.Pager{Err: err}
		}
		url += query

		if full {
			headers = map[string]string{"Accept": "application/json", "Content-Type": "application/json"}
		}
	}

	pager := pagination.NewPager(c, url, func(r pagination.PageResult) pagination.Page {
		p := ObjectPage{pagination.MarkerPageBase{PageResult: r}}
		p.MarkerPageBase.Owner = p
		return p
	})
	pager.Headers = headers
	return pager
}

// DownloadOptsBuilder allows extensions to add additional parameters to the
// Download request.
type DownloadOptsBuilder interface {
	ToObjectDownloadParams() (map[string]string, string, error)
}

// DownloadOpts is a structure that holds parameters for downloading an object.
type DownloadOpts struct {
	IfMatch           string    `h:"If-Match"`
	IfModifiedSince   time.Time `h:"If-Modified-Since"`
	IfNoneMatch       string    `h:"If-None-Match"`
	IfUnmodifiedSince time.Time `h:"If-Unmodified-Since"`
	Newest            bool      `h:"X-Newest"`
	Range             string    `h:"Range"`
	Expires           string    `q:"expires"`
	MultipartManifest string    `q:"multipart-manifest"`
	Signature         string    `q:"signature"`
	ObjectVersionID   string    `q:"version-id"`
}

// ToObjectDownloadParams formats a DownloadOpts into a query string and map of
// headers.
func (opts DownloadOpts) ToObjectDownloadParams() (map[string]string, string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	if err != nil {
		return nil, "", err
	}
	h, err := gophercloud.BuildHeaders(opts)
	if err != nil {
		return nil, q.String(), err
	}
	if !opts.IfModifiedSince.IsZero() {
		h["If-Modified-Since"] = opts.IfModifiedSince.Format(time.RFC1123)
	}
	if !opts.IfUnmodifiedSince.IsZero() {
		h["If-Unmodified-Since"] = opts.IfUnmodifiedSince.Format(time.RFC1123)
	}
	return h, q.String(), nil
}

// Download is a function that retrieves the content and metadata for an object.
// To extract just the content, call the DownloadResult method ExtractContent,
// after checking DownloadResult's Err field.
func Download(c *gophercloud.ServiceClient, containerName, objectName string, opts DownloadOptsBuilder) (r DownloadResult) {
	url, err := downloadURL(c, containerName, objectName)
	if err != nil {
		r.Err = err
		return
	}

	h := make(map[string]string)
	if opts != nil {
		headers, query, err := opts.ToObjectDownloadParams()
		if err != nil {
			r.Err = err
			return
		}
		for k, v := range headers {
			h[k] = v
		}
		url += query
	}

	resp, err := c.Get(url, nil, &gophercloud.RequestOpts{
		MoreHeaders:      h,
		OkCodes:          []int{200, 206, 304},
		KeepResponseBody: true,
	})
	r.Body, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// CreateOptsBuilder allows extensions to add additional parameters to the
// Create request.
type CreateOptsBuilder interface {
	ToObjectCreateParams() (io.Reader, map[string]string, string, error)
}

// CreateOpts is a structure that holds parameters for creating an object.
type CreateOpts struct {
	Content            io.Reader
	Metadata           map[string]string
	NoETag             bool
	CacheControl       string `h:"Cache-Control"`
	ContentDisposition string `h:"Content-Disposition"`
	ContentEncoding    string `h:"Content-Encoding"`
	ContentLength      int64  `h:"Content-Length"`
	ContentType        string `h:"Content-Type"`
	CopyFrom           string `h:"X-Copy-From"`
	DeleteAfter        int64  `h:"X-Delete-After"`
	DeleteAt           int64  `h:"X-Delete-At"`
	DetectContentType  string `h:"X-Detect-Content-Type"`
	ETag               string `h:"ETag"`
	IfNoneMatch        string `h:"If-None-Match"`
	ObjectManifest     string `h:"X-Object-Manifest"`
	TransferEncoding   string `h:"Transfer-Encoding"`
	Expires            string `q:"expires"`
	MultipartManifest  string `q:"multipart-manifest"`
	Signature          string `q:"signature"`
}

// ToObjectCreateParams formats a CreateOpts into a query string and map of
// headers.
func (opts CreateOpts) ToObjectCreateParams() (io.Reader, map[string]string, string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	if err != nil {
		return nil, nil, "", err
	}
	h, err := gophercloud.BuildHeaders(opts)
	if err != nil {
		return nil, nil, "", err
	}

	for k, v := range opts.Metadata {
		h["X-Object-Meta-"+k] = v
	}

	if opts.NoETag {
		delete(h, "etag")
		return opts.Content, h, q.String(), nil
	}

	if h["ETag"] != "" {
		return opts.Content, h, q.String(), nil
	}

	// When we're dealing with big files an io.ReadSeeker allows us to efficiently calculate
	// the md5 sum. An io.Reader is only readable once which means we have to copy the entire
	// file content into memory first.
	readSeeker, isReadSeeker := opts.Content.(io.ReadSeeker)
	if !isReadSeeker {
		data, err := ioutil.ReadAll(opts.Content)
		if err != nil {
			return nil, nil, "", err
		}
		readSeeker = bytes.NewReader(data)
	}

	hash := md5.New()
	// io.Copy into md5 is very efficient as it's done in small chunks.
	if _, err := io.Copy(hash, readSeeker); err != nil {
		return nil, nil, "", err
	}
	readSeeker.Seek(0, io.SeekStart)

	h["ETag"] = fmt.Sprintf("%x", hash.Sum(nil))

	return readSeeker, h, q.String(), nil
}

// Create is a function that creates a new object or replaces an existing
// object. If the returned response's ETag header fails to match the local
// checksum, the failed request will automatically be retried up to a maximum
// of 3 times.
func Create(c *gophercloud.ServiceClient, containerName, objectName string, opts CreateOptsBuilder) (r CreateResult) {
	url, err := createURL(c, containerName, objectName)
	if err != nil {
		r.Err = err
		return
	}
	h := make(map[string]string)
	var b io.Reader
	if opts != nil {
		tmpB, headers, query, err := opts.ToObjectCreateParams()
		if err != nil {
			r.Err = err
			return
		}
		for k, v := range headers {
			h[k] = v
		}
		url += query
		b = tmpB
	}

	resp, err := c.Put(url, b, nil, &gophercloud.RequestOpts{
		MoreHeaders: h,
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// CopyOptsBuilder allows extensions to add additional parameters to the
// Copy request.
type CopyOptsBuilder interface {
	ToObjectCopyMap() (map[string]string, error)
}

// CopyOptsQueryBuilder allows extensions to add additional query parameters to
// the Copy request.
type CopyOptsQueryBuilder interface {
	ToObjectCopyQuery() (string, error)
}

// CopyOpts is a structure that holds parameters for copying one object to
// another.
type CopyOpts struct {
	Metadata           map[string]string
	ContentDisposition string `h:"Content-Disposition"`
	ContentEncoding    string `h:"Content-Encoding"`
	ContentType        string `h:"Content-Type"`
	Destination        string `h:"Destination" required:"true"`
	ObjectVersionID    string `q:"version-id"`
}

// ToObjectCopyMap formats a CopyOpts into a map of headers.
func (opts CopyOpts) ToObjectCopyMap() (map[string]string, error) {
	h, err := gophercloud.BuildHeaders(opts)
	if err != nil {
		return nil, err
	}
	for k, v := range opts.Metadata {
		h["X-Object-Meta-"+k] = v
	}
	return h, nil
}

// ToObjectCopyQuery formats a CopyOpts into a query.
func (opts CopyOpts) ToObjectCopyQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	if err != nil {
		return "", err
	}
	return q.String(), nil
}

// Copy is a function that copies one object to another.
func Copy(c *gophercloud.ServiceClient, containerName, objectName string, opts CopyOptsBuilder) (r CopyResult) {
	url, err := copyURL(c, containerName, objectName)
	if err != nil {
		r.Err = err
		return
	}

	h := make(map[string]string)
	headers, err := opts.ToObjectCopyMap()
	if err != nil {
		r.Err = err
		return
	}
	for k, v := range headers {
		h[k] = v
	}

	if opts, ok := opts.(CopyOptsQueryBuilder); ok {
		query, err := opts.ToObjectCopyQuery()
		if err != nil {
			r.Err = err
			return
		}
		url += query
	}

	resp, err := c.Request("COPY", url, &gophercloud.RequestOpts{
		MoreHeaders: h,
		OkCodes:     []int{201},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// DeleteOptsBuilder allows extensions to add additional parameters to the
// Delete request.
type DeleteOptsBuilder interface {
	ToObjectDeleteQuery() (string, error)
}

// DeleteOpts is a structure that holds parameters for deleting an object.
type DeleteOpts struct {
	MultipartManifest string `q:"multipart-manifest"`
	ObjectVersionID   string `q:"version-id"`
}

// ToObjectDeleteQuery formats a DeleteOpts into a query string.
func (opts DeleteOpts) ToObjectDeleteQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	return q.String(), err
}

// Delete is a function that deletes an object.
func Delete(c *gophercloud.ServiceClient, containerName, objectName string, opts DeleteOptsBuilder) (r DeleteResult) {
	url, err := deleteURL(c, containerName, objectName)
	if err != nil {
		r.Err = err
		return
	}
	if opts != nil {
		query, err := opts.ToObjectDeleteQuery()
		if err != nil {
			r.Err = err
			return
		}
		url += query
	}
	resp, err := c.Delete(url, nil)
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// GetOptsBuilder allows extensions to add additional parameters to the
// Get request.
type GetOptsBuilder interface {
	ToObjectGetParams() (map[string]string, string, error)
}

// GetOpts is a structure that holds parameters for getting an object's
// metadata.
type GetOpts struct {
	Newest          bool   `h:"X-Newest"`
	Expires         string `q:"expires"`
	Signature       string `q:"signature"`
	ObjectVersionID string `q:"version-id"`
}

// ToObjectGetParams formats a GetOpts into a query string and a map of headers.
func (opts GetOpts) ToObjectGetParams() (map[string]string, string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	if err != nil {
		return nil, "", err
	}
	h, err := gophercloud.BuildHeaders(opts)
	if err != nil {
		return nil, q.String(), err
	}
	return h, q.String(), nil
}

// Get is a function that retrieves the metadata of an object. To extract just
// the custom metadata, pass the GetResult response to the ExtractMetadata
// function.
func Get(c *gophercloud.ServiceClient, containerName, objectName string, opts GetOptsBuilder) (r GetResult) {
	url, err := getURL(c, containerName, objectName)
	if err != nil {
		r.Err = err
		return
	}
	h := make(map[string]string)
	if opts != nil {
		headers, query, err := opts.ToObjectGetParams()
		if err != nil {
			r.Err = err
			return
		}
		for k, v := range headers {
			h[k] = v
		}
		url += query
	}

	resp, err := c.Head(url, &gophercloud.RequestOpts{
		MoreHeaders: h,
		OkCodes:     []int{200, 204},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// UpdateOptsBuilder allows extensions to add additional parameters to the
// Update request.
type UpdateOptsBuilder interface {
	ToObjectUpdateMap() (map[string]string, error)
}

// UpdateOpts is a structure that holds parameters for updating, creating, or
// deleting an object's metadata.
type UpdateOpts struct {
	Metadata           map[string]string
	RemoveMetadata     []string
	ContentDisposition *string `h:"Content-Disposition"`
	ContentEncoding    *string `h:"Content-Encoding"`
	ContentType        *string `h:"Content-Type"`
	DeleteAfter        *int64  `h:"X-Delete-After"`
	DeleteAt           *int64  `h:"X-Delete-At"`
	DetectContentType  *bool   `h:"X-Detect-Content-Type"`
}

// ToObjectUpdateMap formats a UpdateOpts into a map of headers.
func (opts UpdateOpts) ToObjectUpdateMap() (map[string]string, error) {
	h, err := gophercloud.BuildHeaders(opts)
	if err != nil {
		return nil, err
	}

	for k, v := range opts.Metadata {
		h["X-Object-Meta-"+k] = v
	}

	for _, k := range opts.RemoveMetadata {
		h["X-Remove-Object-Meta-"+k] = "remove"
	}
	return h, nil
}

// Update is a function that creates, updates, or deletes an object's metadata.
func Update(c *gophercloud.ServiceClient, containerName, objectName string, opts UpdateOptsBuilder) (r UpdateResult) {
	url, err := updateURL(c, containerName, objectName)
	if err != nil {
		r.Err = err
		return
	}
	h := make(map[string]string)
	if opts != nil {
		headers, err := opts.ToObjectUpdateMap()
		if err != nil {
			r.Err = err
			return
		}

		for k, v := range headers {
			h[k] = v
		}
	}
	resp, err := c.Post(url, nil, nil, &gophercloud.RequestOpts{
		MoreHeaders: h,
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// HTTPMethod represents an HTTP method string (e.g. "GET").
type HTTPMethod string

var (
	// GET represents an HTTP "GET" method.
	GET HTTPMethod = "GET"
	// HEAD represents an HTTP "HEAD" method.
	HEAD HTTPMethod = "HEAD"
	// PUT represents an HTTP "PUT" method.
	PUT HTTPMethod = "PUT"
	// POST represents an HTTP "POST" method.
	POST HTTPMethod = "POST"
	// DELETE represents an HTTP "DELETE" method.
	DELETE HTTPMethod = "DELETE"
)

// CreateTempURLOpts are options for creating a temporary URL for an object.
type CreateTempURLOpts struct {
	// (REQUIRED) Method is the HTTP method to allow for users of the temp URL.
	// Valid values are "GET", "HEAD", "PUT", "POST" and "DELETE".
	Method HTTPMethod

	// (REQUIRED) TTL is the number of seconds the temp URL should be active.
	TTL int

	// (Optional) Split is the string on which to split the object URL. Since only
	// the object path is used in the hash, the object URL needs to be parsed. If
	// empty, the default OpenStack URL split point will be used ("/v1/").
	Split string

	// (Optional) Timestamp is the current timestamp used to calculate the Temp URL
	// signature. If not specified, the current UNIX timestamp is used as the base
	// timestamp.
	Timestamp time.Time

	// (Optional) TempURLKey overrides the Swift container or account Temp URL key.
	// TempURLKey must correspond to a target container/account key, otherwise the
	// generated link will be invalid. If not specified, the key is obtained from
	// a Swift container or account.
	TempURLKey string

	// (Optional) Digest specifies the cryptographic hash function used to
	// calculate the signature. Valid values include sha1, sha256, and
	// sha512. If not specified, the default hash function is sha1.
	Digest string
}

// CreateTempURL is a function for creating a temporary URL for an object. It
// allows users to have "GET" or "POST" access to a particular tenant's object
// for a limited amount of time.
func CreateTempURL(c *gophercloud.ServiceClient, containerName, objectName string, opts CreateTempURLOpts) (string, error) {
	url, err := getURL(c, containerName, objectName)
	if err != nil {
		return "", err
	}

	if opts.Split == "" {
		opts.Split = "/v1/"
	}

	// Initialize time if it was not passed as opts
	date := opts.Timestamp
	if date.IsZero() {
		date = time.Now()
	}
	duration := time.Duration(opts.TTL) * time.Second
	// UNIX time is always UTC
	expiry := date.Add(duration).Unix()

	// Initialize the tempURLKey to calculate a signature
	tempURLKey := opts.TempURLKey
	if tempURLKey == "" {
		// fallback to a container TempURL key
		getHeader, err := containers.Get(c, containerName, nil).Extract()
		if err != nil {
			return "", err
		}
		tempURLKey = getHeader.TempURLKey
		if tempURLKey == "" {
			// fallback to an account TempURL key
			getHeader, err := accounts.Get(c, nil).Extract()
			if err != nil {
				return "", err
			}
			tempURLKey = getHeader.TempURLKey
		}
		if tempURLKey == "" {
			return "", ErrTempURLKeyNotFound{}
		}
	}

	secretKey := []byte(tempURLKey)
	splitPath := strings.SplitN(url, opts.Split, 2)
	baseURL, objectPath := splitPath[0], splitPath[1]
	objectPath = opts.Split + objectPath
	body := fmt.Sprintf("%s\n%d\n%s", opts.Method, expiry, objectPath)
	var hash hash.Hash
	switch opts.Digest {
	case "", "sha1":
		hash = hmac.New(sha1.New, secretKey)
	case "sha256":
		hash = hmac.New(sha256.New, secretKey)
	case "sha512":
		hash = hmac.New(sha512.New, secretKey)
	default:
		return "", ErrTempURLDigestNotValid{Digest: opts.Digest}
	}
	hash.Write([]byte(body))
	hexsum := fmt.Sprintf("%x", hash.Sum(nil))
	return fmt.Sprintf("%s%s?temp_url_sig=%s&temp_url_expires=%d", baseURL, objectPath, hexsum, expiry), nil
}

// BulkDelete is a function that bulk deletes objects.
// In Swift, the maximum number of deletes per request is set by default to 10000.
//
// See:
// * https://github.com/openstack/swift/blob/6d3d4197151f44bf28b51257c1a4c5d33411dcae/etc/proxy-server.conf-sample#L1029-L1034
// * https://github.com/openstack/swift/blob/e8cecf7fcc1630ee83b08f9a73e1e59c07f8d372/swift/common/middleware/bulk.py#L309
func BulkDelete(c *gophercloud.ServiceClient, container string, objects []string) (r BulkDeleteResult) {
	err := containers.CheckContainerName(container)
	if err != nil {
		r.Err = err
		return
	}

	var body bytes.Buffer
	for i := range objects {
		if objects[i] == "" {
			r.Err = fmt.Errorf("object names must not be the empty string")
			return
		}
		body.WriteString(container)
		body.WriteRune('/')
		body.WriteString(objects[i])
		body.WriteRune('\n')
	}

	resp, err := c.Post(bulkDeleteURL(c), &body, &r.Body, &gophercloud.RequestOpts{
		MoreHeaders: map[string]string{
			"Accept":       "application/json",
			"Content-Type": "text/plain",
		},
		OkCodes: []int{200},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}
