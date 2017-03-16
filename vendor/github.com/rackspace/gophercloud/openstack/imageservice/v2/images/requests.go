package images

import (
	"fmt"
	"io"
	"net/http"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/pagination"
)

// ListOptsBuilder allows extensions to add additional parameters to the
// List request.
type ListOptsBuilder interface {
	ToImageListQuery() (string, error)
}

// ListOpts allows the filtering and sorting of paginated collections through
// the API. Filtering is achieved by passing in struct field values that map to
// the server attributes you want to see returned. Marker and Limit are used
// for pagination.
//http://developer.openstack.org/api-ref-image-v2.html
type ListOpts struct {
	// Integer value for the limit of values to return.
	Limit int `q:"limit"`

	// UUID of the server at which you want to set a marker.
	Marker string `q:"marker"`

	Name         string            `q:"name"`
	Visibility   ImageVisibility   `q:"visibility"`
	MemberStatus ImageMemberStatus `q:"member_status"`
	Owner        string            `q:"owner"`
	Status       ImageStatus       `q:"status"`
	SizeMin      int64             `q:"size_min"`
	SizeMax      int64             `q:"size_max"`
	SortKey      string            `q:"sort_key"`
	SortDir      string            `q:"sort_dir"`
	Tag          string            `q:"tag"`
}

// ToImageListQuery formats a ListOpts into a query string.
func (opts ListOpts) ToImageListQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	if err != nil {
		return "", err
	}
	return q.String(), nil
}

// List implements image list request
func List(c *gophercloud.ServiceClient, opts ListOptsBuilder) pagination.Pager {
	url := listURL(c)

	if opts != nil {
		query, err := opts.ToImageListQuery()
		if err != nil {
			return pagination.Pager{Err: err}
		}
		url += query
	}

	createPageFn := func(r pagination.PageResult) pagination.Page {
		return ImagePage{pagination.LinkedPageBase{PageResult: r}}
	}
	return pagination.NewPager(c, url, createPageFn)
}

// Create implements create image request
func Create(client *gophercloud.ServiceClient, opts CreateOptsBuilder) CreateResult {
	var res CreateResult
	body, err := opts.ToImageCreateMap()
	if err != nil {
		res.Err = err
		return res
	}
	_, res.Err = client.Post(createURL(client), body, &res.Body, &gophercloud.RequestOpts{OkCodes: []int{201}})
	return res
}

// CreateOptsBuilder describes struct types that can be accepted by the Create call.
// The CreateOpts struct in this package does.
type CreateOptsBuilder interface {
	// Returns value that can be passed to json.Marshal
	ToImageCreateMap() (map[string]interface{}, error)
}

// CreateOpts implements CreateOptsBuilder
type CreateOpts struct {
	// Name [required] is the name of the new image.
	Name string

	// Id [optional] is the the image ID.
	ID string

	// Visibility [optional] defines who can see/use the image.
	Visibility *ImageVisibility

	// Tags [optional] is a set of image tags.
	Tags []string

	// ContainerFormat [optional] is the format of the
	// container. Valid values are ami, ari, aki, bare, and ovf.
	ContainerFormat string

	// DiskFormat [optional] is the format of the disk. If set,
	// valid values are ami, ari, aki, vhd, vmdk, raw, qcow2, vdi,
	// and iso.
	DiskFormat string

	// MinDiskGigabytes [optional] is the amount of disk space in
	// GB that is required to boot the image.
	MinDiskGigabytes int

	// MinRAMMegabytes [optional] is the amount of RAM in MB that
	// is required to boot the image.
	MinRAMMegabytes int

	// protected [optional] is whether the image is not deletable.
	Protected bool

	// properties [optional] is a set of properties, if any, that
	// are associated with the image.
	Properties map[string]string
}

// ToImageCreateMap assembles a request body based on the contents of
// a CreateOpts.
func (opts CreateOpts) ToImageCreateMap() (map[string]interface{}, error) {
	body := map[string]interface{}{}
	if opts.Name == "" {
		return body, fmt.Errorf("'Name' field is requered, but is not set (was: %v)'", opts.Name)
	}

	body["name"] = opts.Name
	if opts.ID != "" {
		body["id"] = opts.ID
	}
	if opts.Visibility != nil {
		body["visibility"] = opts.Visibility
	}
	if opts.Tags != nil {
		body["tags"] = opts.Tags
	}
	if opts.ContainerFormat != "" {
		body["container_format"] = opts.ContainerFormat
	}
	if opts.DiskFormat != "" {
		body["disk_format"] = opts.DiskFormat
	}
	if opts.MinDiskGigabytes != 0 {
		body["min_disk"] = opts.MinDiskGigabytes
	}
	if opts.MinRAMMegabytes != 0 {
		body["min_ram"] = opts.MinRAMMegabytes

	}

	body["protected"] = opts.Protected

	if opts.Properties != nil {
		for k, v := range opts.Properties {
			body[k] = v
		}
	}
	return body, nil
}

// Delete implements image delete request
func Delete(client *gophercloud.ServiceClient, id string) DeleteResult {
	var res DeleteResult
	_, res.Err = client.Delete(deleteURL(client, id), &gophercloud.RequestOpts{
		OkCodes: []int{204},
	})
	return res
}

// Get implements image get request
func Get(client *gophercloud.ServiceClient, id string) GetResult {
	var res GetResult
	_, res.Err = client.Get(getURL(client, id), &res.Body, nil)
	return res
}

// Update implements image updated request
func Update(client *gophercloud.ServiceClient, id string, opts UpdateOptsBuilder) UpdateResult {
	var res UpdateResult
	reqBody := opts.ToImageUpdateMap()

	_, res.Err = client.Patch(updateURL(client, id), reqBody, &res.Body, &gophercloud.RequestOpts{
		OkCodes:     []int{200},
		MoreHeaders: map[string]string{"Content-Type": "application/openstack-images-v2.1-json-patch"},
	})
	return res
}

// UpdateOptsBuilder implements UpdateOptsBuilder
type UpdateOptsBuilder interface {
	// returns value implementing json.Marshaler which when marshaled matches the patch schema:
	// http://specs.openstack.org/openstack/glance-specs/specs/api/v2/http-patch-image-api-v2.html
	ToImageUpdateMap() []interface{}
}

// UpdateOpts implements UpdateOpts
type UpdateOpts []Patch

// ToImageUpdateMap builder
func (opts UpdateOpts) ToImageUpdateMap() []interface{} {
	m := make([]interface{}, len(opts))
	for i, patch := range opts {
		patchJSON := patch.ToImagePatchMap()
		m[i] = patchJSON
	}
	return m
}

// Patch represents a single update to an existing image. Multiple updates to an image can be
// submitted at the same time.
type Patch interface {
	ToImagePatchMap() map[string]interface{}
}

// UpdateVisibility updated visibility
type UpdateVisibility struct {
	Visibility ImageVisibility
}

// ToImagePatchMap builder
func (u UpdateVisibility) ToImagePatchMap() map[string]interface{} {
	m := map[string]interface{}{}
	m["op"] = "relace"
	m["path"] = "/visibility"
	m["value"] = u.Visibility
	return m
}

// ReplaceImageName implements Patch
type ReplaceImageName struct {
	NewName string
}

// ToImagePatchMap builder
func (r ReplaceImageName) ToImagePatchMap() map[string]interface{} {
	m := map[string]interface{}{}
	m["op"] = "replace"
	m["path"] = "/name"
	m["value"] = r.NewName
	return m
}

// ReplaceImageChecksum implements Patch
type ReplaceImageChecksum struct {
	Checksum string
}

// ReplaceImageChecksum builder
func (rc ReplaceImageChecksum) ReplaceImageChecksum() map[string]interface{} {
	m := map[string]interface{}{}
	m["op"] = "replace"
	m["path"] = "/checksum"
	m["value"] = rc.Checksum
	return m
}

// ReplaceImageTags implements Patch
type ReplaceImageTags struct {
	NewTags []string
}

// ToImagePatchMap builder
func (r ReplaceImageTags) ToImagePatchMap() map[string]interface{} {
	m := map[string]interface{}{}
	m["op"] = "replace"
	m["path"] = "/tags"
	m["value"] = r.NewTags
	return m
}

// Upload uploads image file
func Upload(client *gophercloud.ServiceClient, id string, data io.ReadSeeker) PutImageDataResult {
	var res PutImageDataResult

	_, res.Err = client.Put(imageDataURL(client, id), data, nil, &gophercloud.RequestOpts{
		MoreHeaders: map[string]string{"Content-Type": "application/octet-stream"},
		OkCodes:     []int{204},
	})

	return res
}

// Download retrieves file
func Download(client *gophercloud.ServiceClient, id string) GetImageDataResult {
	var res GetImageDataResult

	var resp *http.Response
	resp, res.Err = client.Get(imageDataURL(client, id), nil, &gophercloud.RequestOpts{
		OkCodes: []int{200},
	})

	res.Body = resp.Body

	return res
}
