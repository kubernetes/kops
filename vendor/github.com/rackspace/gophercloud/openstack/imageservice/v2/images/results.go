package images

import (
	"fmt"
	"io"
	"reflect"

	"github.com/mitchellh/mapstructure"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/pagination"
)

// Image model
// Does not include the literal image data; just metadata.
// returned by listing images, and by fetching a specific image.
type Image struct {
	// ID is the image UUID
	ID string

	// Name is the human-readable display name for the image.
	Name string

	// Status is the image status. It can be "queued" or "active"
	// See imageservice/v2/images/type.go
	Status ImageStatus

	// Tags is a list of image tags. Tags are arbitrarily defined strings
	// attached to an image.
	Tags []string

	// ContainerFormat is the format of the container.
	// Valid values are ami, ari, aki, bare, and ovf.
	ContainerFormat string `mapstructure:"container_format"`

	// DiskFormat is the format of the disk.
	// If set, valid values are ami, ari, aki, vhd, vmdk, raw, qcow2, vdi, and iso.
	DiskFormat string `mapstructure:"disk_format"`

	// MinDiskGigabytes is the amount of disk space in GB that is required to boot the image.
	MinDiskGigabytes int `mapstructure:"min_disk"`

	// MinRAMMegabytes [optional] is the amount of RAM in MB that is required to boot the image.
	MinRAMMegabytes int `mapstructure:"min_ram"`

	// Owner is the tenant the image belongs to.
	Owner string

	// Protected is whether the image is deletable or not.
	Protected bool

	// Visibility defines who can see/use the image.
	Visibility ImageVisibility

	// Checksum is the checksum of the data that's associated with the image
	Checksum string

	// SizeBytes is the size of the data that's associated with the image.
	SizeBytes int `mapstructure:"size"`

	// Metadata is a set of metadata associated with the image.
	// Image metadata allow for meaningfully define the image properties
	// and tags. See http://docs.openstack.org/developer/glance/metadefs-concepts.html.
	Metadata map[string]string

	// Properties is a set of key-value pairs, if any, that are associated with the image.
	Properties map[string]string

	// CreatedDate is the date when the image has been created.
	CreatedDate string `mapstructure:"created_at"`

	// LastUpdate is the date when the last change has been made to the image or it's properties.
	LastUpdate string `mapstructure:"updated_at"`

	// File is the trailing path after the glance endpoint that represent the location
	// of the image or the path to retrieve it.
	File string `mapstructure:"file"`

	// Schema is the path to the JSON-schema that represent the image or image entity.
	Schema string `mapstructure:"schema"`
}

type commonResult struct {
	gophercloud.Result
}

// Extract interprets any commonResult as an Image.
func (c commonResult) Extract() (*Image, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	var image *Image

	err := mapstructure.Decode(c.Result.Body, &image)
	return image, err
}

// CreateResult represents the result of a Create operation
type CreateResult struct {
	commonResult
}

// UpdateResult represents the result of an Update operation
type UpdateResult struct {
	commonResult
}

// GetResult represents the result of a Get operation
type GetResult struct {
	commonResult
}

//DeleteResult model
type DeleteResult struct {
	gophercloud.Result
}

// PutImageDataResult is model put image respose
type PutImageDataResult struct {
	gophercloud.Result
}

// GetImageDataResult model for image response
type GetImageDataResult struct {
	gophercloud.Result
}

// Extract builds images model from io.Reader
func (g GetImageDataResult) Extract() (io.Reader, error) {
	if r, ok := g.Body.(io.Reader); ok {
		return r, nil
	}
	return nil, fmt.Errorf("Expected io.Reader but got: %T(%#v)", g.Body, g.Body)
}

// ImagePage represents page
type ImagePage struct {
	pagination.LinkedPageBase
}

// IsEmpty returns true if a page contains no Images results.
func (page ImagePage) IsEmpty() (bool, error) {
	images, err := ExtractImages(page)
	if err != nil {
		return true, err
	}
	return len(images) == 0, nil
}

// NextPageURL uses the response's embedded link reference to navigate to the next page of results.
func (page ImagePage) NextPageURL() (string, error) {
	type resp struct {
		Next string `mapstructure:"next"`
	}

	var r resp
	err := mapstructure.Decode(page.Body, &r)
	if err != nil {
		return "", err
	}

	return nextPageURL(page.URL.String(), r.Next), nil
}

func toMapFromString(from reflect.Kind, to reflect.Kind, data interface{}) (interface{}, error) {
	if (from == reflect.String) && (to == reflect.Map) {
		return map[string]interface{}{}, nil
	}
	return data, nil
}

// ExtractImages interprets the results of a single page from a List() call, producing a slice of Image entities.
func ExtractImages(page pagination.Page) ([]Image, error) {
	casted := page.(ImagePage).Body

	var response struct {
		Images []Image `mapstructure:"images"`
	}

	config := &mapstructure.DecoderConfig{
		DecodeHook: toMapFromString,
		Result:     &response,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return nil, err
	}

	err = decoder.Decode(casted)
	if err != nil {
		fmt.Printf("Error happened %v \n", err)
	}

	return response.Images, err
}
