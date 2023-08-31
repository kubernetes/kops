// This file was automatically generated. DO NOT EDIT.
// If you have any remark or suggestion do not hesitate to open an issue.

// Package marketplace provides methods and message types of the marketplace v2 API.
package marketplace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/internal/marshaler"
	"github.com/scaleway/scaleway-sdk-go/internal/parameter"
	"github.com/scaleway/scaleway-sdk-go/namegenerator"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// always import dependencies
var (
	_ fmt.Stringer
	_ json.Unmarshaler
	_ url.URL
	_ net.IP
	_ http.Header
	_ bytes.Reader
	_ time.Time
	_ = strings.Join

	_ scw.ScalewayRequest
	_ marshaler.Duration
	_ scw.File
	_ = parameter.AddToQuery
	_ = namegenerator.GetRandomName
)

// API: marketplace API.
type API struct {
	client *scw.Client
}

// NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

type ListImagesRequestOrderBy string

const (
	ListImagesRequestOrderByNameAsc       = ListImagesRequestOrderBy("name_asc")
	ListImagesRequestOrderByNameDesc      = ListImagesRequestOrderBy("name_desc")
	ListImagesRequestOrderByCreatedAtAsc  = ListImagesRequestOrderBy("created_at_asc")
	ListImagesRequestOrderByCreatedAtDesc = ListImagesRequestOrderBy("created_at_desc")
	ListImagesRequestOrderByUpdatedAtAsc  = ListImagesRequestOrderBy("updated_at_asc")
	ListImagesRequestOrderByUpdatedAtDesc = ListImagesRequestOrderBy("updated_at_desc")
)

func (enum ListImagesRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "name_asc"
	}
	return string(enum)
}

func (enum ListImagesRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListImagesRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListImagesRequestOrderBy(ListImagesRequestOrderBy(tmp).String())
	return nil
}

type ListLocalImagesRequestOrderBy string

const (
	ListLocalImagesRequestOrderByCreatedAtAsc  = ListLocalImagesRequestOrderBy("created_at_asc")
	ListLocalImagesRequestOrderByCreatedAtDesc = ListLocalImagesRequestOrderBy("created_at_desc")
)

func (enum ListLocalImagesRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListLocalImagesRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListLocalImagesRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListLocalImagesRequestOrderBy(ListLocalImagesRequestOrderBy(tmp).String())
	return nil
}

type ListVersionsRequestOrderBy string

const (
	ListVersionsRequestOrderByCreatedAtAsc  = ListVersionsRequestOrderBy("created_at_asc")
	ListVersionsRequestOrderByCreatedAtDesc = ListVersionsRequestOrderBy("created_at_desc")
)

func (enum ListVersionsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListVersionsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListVersionsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListVersionsRequestOrderBy(ListVersionsRequestOrderBy(tmp).String())
	return nil
}

type LocalImageType string

const (
	// Unspecified image type
	LocalImageTypeUnknownType = LocalImageType("unknown_type")
	// An image type that can be used to create volumes which are managed via the Instance API.
	LocalImageTypeInstanceLocal = LocalImageType("instance_local")
	// An image type that can be used to create volumes which are managed via the Scaleway Block Storage (SBS) API.
	LocalImageTypeInstanceSbs = LocalImageType("instance_sbs")
)

func (enum LocalImageType) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown_type"
	}
	return string(enum)
}

func (enum LocalImageType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *LocalImageType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = LocalImageType(LocalImageType(tmp).String())
	return nil
}

type Category struct {
	ID string `json:"id"`

	Name string `json:"name"`

	Description string `json:"description"`
}

// Image: image.
type Image struct {
	// ID: UUID of this image.
	ID string `json:"id"`
	// Name: name of the image.
	Name string `json:"name"`
	// Description: text description of this image.
	Description string `json:"description"`
	// Logo: URL of this image's logo.
	Logo string `json:"logo"`
	// Categories: list of categories this image belongs to.
	Categories []string `json:"categories"`
	// CreatedAt: creation date of this image.
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: date of the last modification of this image.
	UpdatedAt *time.Time `json:"updated_at"`
	// ValidUntil: expiration date of this image.
	ValidUntil *time.Time `json:"valid_until"`
	// Label: label of this image.
	// Typically an identifier for a distribution (ex. "ubuntu_focal").
	Label string `json:"label"`
}

type ListCategoriesResponse struct {
	Categories []*Category `json:"categories"`

	TotalCount uint32 `json:"total_count"`
}

type ListImagesResponse struct {
	Images []*Image `json:"images"`

	TotalCount uint32 `json:"total_count"`
}

type ListLocalImagesResponse struct {
	LocalImages []*LocalImage `json:"local_images"`

	TotalCount uint32 `json:"total_count"`
}

type ListVersionsResponse struct {
	Versions []*Version `json:"versions"`

	TotalCount uint32 `json:"total_count"`
}

// LocalImage: local image.
type LocalImage struct {
	// ID: UUID of this local image.
	// Version you will typically use to define an image in an API call.
	ID string `json:"id"`
	// CompatibleCommercialTypes: list of all commercial types that are compatible with this local image.
	CompatibleCommercialTypes []string `json:"compatible_commercial_types"`
	// Arch: supported architecture for this local image.
	Arch string `json:"arch"`
	// Zone: availability Zone where this local image is available.
	Zone scw.Zone `json:"zone"`
	// Label: image label this image belongs to.
	Label string `json:"label"`
	// Type: type of this local image.
	// Default value: unknown_type
	Type LocalImageType `json:"type"`
}

// Version: version.
type Version struct {
	// ID: UUID of this version.
	ID string `json:"id"`
	// Name: name of this version.
	Name string `json:"name"`
	// CreatedAt: creation date of this image version.
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: date of the last modification of this version.
	UpdatedAt *time.Time `json:"updated_at"`
	// PublishedAt: date this version was officially published.
	PublishedAt *time.Time `json:"published_at"`
}

// Service API

type ListImagesRequest struct {
	// PageSize: a positive integer lower or equal to 100 to select the number of items to display.
	PageSize *uint32 `json:"-"`
	// Page: a positive integer to choose the page to display.
	Page *int32 `json:"-"`
	// OrderBy: ordering to use.
	// Default value: name_asc
	OrderBy ListImagesRequestOrderBy `json:"-"`
	// Arch: choose for which machine architecture to return images.
	Arch *string `json:"-"`
	// Category: choose the category of images to get.
	Category *string `json:"-"`
	// IncludeEol: choose to include end-of-life images.
	IncludeEol bool `json:"-"`
}

// ListImages: list marketplace images.
// List all available images on the marketplace, their UUID, CPU architecture and description.
func (s *API) ListImages(req *ListImagesRequest, opts ...scw.RequestOption) (*ListImagesResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "arch", req.Arch)
	parameter.AddToQuery(query, "category", req.Category)
	parameter.AddToQuery(query, "include_eol", req.IncludeEol)

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/marketplace/v2/images",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListImagesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetImageRequest struct {
	// ImageID: display the image name.
	ImageID string `json:"-"`
}

// GetImage: get a specific marketplace image.
// Get detailed information about a marketplace image, specified by its `image_id` (UUID format).
func (s *API) GetImage(req *GetImageRequest, opts ...scw.RequestOption) (*Image, error) {
	var err error

	if fmt.Sprint(req.ImageID) == "" {
		return nil, errors.New("field ImageID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/marketplace/v2/images/" + fmt.Sprint(req.ImageID) + "",
		Headers: http.Header{},
	}

	var resp Image

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListVersionsRequest struct {
	ImageID string `json:"-"`

	PageSize *uint32 `json:"-"`

	Page *int32 `json:"-"`
	// OrderBy: default value: created_at_asc
	OrderBy ListVersionsRequestOrderBy `json:"-"`
}

// ListVersions: list versions of an Image.
// Get a list of all available version of an image, specified by its `image_id` (UUID format).
func (s *API) ListVersions(req *ListVersionsRequest, opts ...scw.RequestOption) (*ListVersionsResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "image_id", req.ImageID)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "order_by", req.OrderBy)

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/marketplace/v2/versions",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListVersionsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetVersionRequest struct {
	VersionID string `json:"-"`
}

// GetVersion: get a specific image version.
// Get information such as the name, creation date, last update and published date for an image version specified by its `version_id` (UUID format).
func (s *API) GetVersion(req *GetVersionRequest, opts ...scw.RequestOption) (*Version, error) {
	var err error

	if fmt.Sprint(req.VersionID) == "" {
		return nil, errors.New("field VersionID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/marketplace/v2/versions/" + fmt.Sprint(req.VersionID) + "",
		Headers: http.Header{},
	}

	var resp Version

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListLocalImagesRequest struct {
	ImageID *string `json:"-"`

	VersionID *string `json:"-"`

	PageSize *uint32 `json:"-"`

	Page *int32 `json:"-"`
	// OrderBy: default value: created_at_asc
	OrderBy ListLocalImagesRequestOrderBy `json:"-"`

	ImageLabel *string `json:"-"`

	Zone *scw.Zone `json:"-"`
	// Type: default value: unknown_type
	Type LocalImageType `json:"-"`
}

// ListLocalImages: list local images from a specific image or version.
// List information about local images in a specific Availability Zone, specified by its `image_id` (UUID format), `version_id` (UUID format) or `image_label`. Only one of these three parameters may be set.
func (s *API) ListLocalImages(req *ListLocalImagesRequest, opts ...scw.RequestOption) (*ListLocalImagesResponse, error) {
	var err error

	defaultZone, exist := s.client.GetDefaultZone()
	if (req.Zone == nil || *req.Zone == "") && exist {
		req.Zone = &defaultZone
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "image_id", req.ImageID)
	parameter.AddToQuery(query, "version_id", req.VersionID)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "image_label", req.ImageLabel)
	parameter.AddToQuery(query, "zone", req.Zone)
	parameter.AddToQuery(query, "type", req.Type)

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/marketplace/v2/local-images",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListLocalImagesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetLocalImageRequest struct {
	LocalImageID string `json:"-"`
}

// GetLocalImage: get a specific local image by ID.
// Get detailed information about a local image, including compatible commercial types, supported architecture, labels and the Availability Zone of the image, specified by its `local_image_id` (UUID format).
func (s *API) GetLocalImage(req *GetLocalImageRequest, opts ...scw.RequestOption) (*LocalImage, error) {
	var err error

	if fmt.Sprint(req.LocalImageID) == "" {
		return nil, errors.New("field LocalImageID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/marketplace/v2/local-images/" + fmt.Sprint(req.LocalImageID) + "",
		Headers: http.Header{},
	}

	var resp LocalImage

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListCategoriesRequest struct {
	PageSize *uint32 `json:"-"`

	Page *int32 `json:"-"`
}

// ListCategories: list existing image categories.
// Get a list of all existing categories. The output can be paginated.
func (s *API) ListCategories(req *ListCategoriesRequest, opts ...scw.RequestOption) (*ListCategoriesResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/marketplace/v2/categories",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListCategoriesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetCategoryRequest struct {
	CategoryID string `json:"-"`
}

// GetCategory: get a specific category.
// Get information about a specific category of the marketplace catalog, specified by its `category_id` (UUID format).
func (s *API) GetCategory(req *GetCategoryRequest, opts ...scw.RequestOption) (*Category, error) {
	var err error

	if fmt.Sprint(req.CategoryID) == "" {
		return nil, errors.New("field CategoryID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/marketplace/v2/categories/" + fmt.Sprint(req.CategoryID) + "",
		Headers: http.Header{},
	}

	var resp Category

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListImagesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListImagesResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListImagesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Images = append(r.Images, results.Images...)
	r.TotalCount += uint32(len(results.Images))
	return uint32(len(results.Images)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListVersionsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListVersionsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListVersionsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Versions = append(r.Versions, results.Versions...)
	r.TotalCount += uint32(len(results.Versions))
	return uint32(len(results.Versions)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListLocalImagesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListLocalImagesResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListLocalImagesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.LocalImages = append(r.LocalImages, results.LocalImages...)
	r.TotalCount += uint32(len(results.LocalImages))
	return uint32(len(results.LocalImages)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListCategoriesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListCategoriesResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListCategoriesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Categories = append(r.Categories, results.Categories...)
	r.TotalCount += uint32(len(results.Categories))
	return uint32(len(results.Categories)), nil
}
