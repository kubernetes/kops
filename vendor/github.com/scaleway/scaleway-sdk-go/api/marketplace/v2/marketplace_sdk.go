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

	"github.com/scaleway/scaleway-sdk-go/errors"
	"github.com/scaleway/scaleway-sdk-go/marshaler"
	"github.com/scaleway/scaleway-sdk-go/namegenerator"
	"github.com/scaleway/scaleway-sdk-go/parameter"
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
		return string(ListImagesRequestOrderByNameAsc)
	}
	return string(enum)
}

func (enum ListImagesRequestOrderBy) Values() []ListImagesRequestOrderBy {
	return []ListImagesRequestOrderBy{
		"name_asc",
		"name_desc",
		"created_at_asc",
		"created_at_desc",
		"updated_at_asc",
		"updated_at_desc",
	}
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
	ListLocalImagesRequestOrderByTypeAsc       = ListLocalImagesRequestOrderBy("type_asc")
	ListLocalImagesRequestOrderByTypeDesc      = ListLocalImagesRequestOrderBy("type_desc")
	ListLocalImagesRequestOrderByCreatedAtAsc  = ListLocalImagesRequestOrderBy("created_at_asc")
	ListLocalImagesRequestOrderByCreatedAtDesc = ListLocalImagesRequestOrderBy("created_at_desc")
)

func (enum ListLocalImagesRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return string(ListLocalImagesRequestOrderByTypeAsc)
	}
	return string(enum)
}

func (enum ListLocalImagesRequestOrderBy) Values() []ListLocalImagesRequestOrderBy {
	return []ListLocalImagesRequestOrderBy{
		"type_asc",
		"type_desc",
		"created_at_asc",
		"created_at_desc",
	}
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
		return string(ListVersionsRequestOrderByCreatedAtAsc)
	}
	return string(enum)
}

func (enum ListVersionsRequestOrderBy) Values() []ListVersionsRequestOrderBy {
	return []ListVersionsRequestOrderBy{
		"created_at_asc",
		"created_at_desc",
	}
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
	// Unspecified image type.
	LocalImageTypeUnknownType = LocalImageType("unknown_type")
	// An image type that can be used to create volumes which are managed via the Instance API.
	LocalImageTypeInstanceLocal = LocalImageType("instance_local")
	// An image type that can be used to create volumes which are managed via the Scaleway Block Storage (SBS) API.
	LocalImageTypeInstanceSbs = LocalImageType("instance_sbs")
)

func (enum LocalImageType) String() string {
	if enum == "" {
		// return default value if empty
		return string(LocalImageTypeUnknownType)
	}
	return string(enum)
}

func (enum LocalImageType) Values() []LocalImageType {
	return []LocalImageType{
		"unknown_type",
		"instance_local",
		"instance_sbs",
	}
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

// Category: category.
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

	// Label: typically an identifier for a distribution (ex. "ubuntu_focal").
	// This label can be used in the image field of the server creation request.
	Label string `json:"label"`
}

// LocalImage: local image.
type LocalImage struct {
	// ID: version you will typically use to define an image in an API call.
	ID string `json:"id"`

	// CompatibleCommercialTypes: list of all commercial types that are compatible with this local image.
	CompatibleCommercialTypes []string `json:"compatible_commercial_types"`

	// Arch: supported architecture for this local image.
	Arch string `json:"arch"`

	// Zone: availability Zone where this local image is available.
	Zone scw.Zone `json:"zone"`

	// Label: this label can be used in the image field of the server creation request.
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

// GetCategoryRequest: get category request.
type GetCategoryRequest struct {
	CategoryID string `json:"-"`
}

// GetImageRequest: get image request.
type GetImageRequest struct {
	// ImageID: display the image name.
	ImageID string `json:"-"`
}

// GetLocalImageRequest: get local image request.
type GetLocalImageRequest struct {
	LocalImageID string `json:"-"`
}

// GetVersionRequest: get version request.
type GetVersionRequest struct {
	VersionID string `json:"-"`
}

// ListCategoriesRequest: list categories request.
type ListCategoriesRequest struct {
	PageSize *uint32 `json:"-"`

	Page *int32 `json:"-"`
}

// ListCategoriesResponse: list categories response.
type ListCategoriesResponse struct {
	Categories []*Category `json:"categories"`

	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListCategoriesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListCategoriesResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListCategoriesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Categories = append(r.Categories, results.Categories...)
	r.TotalCount += uint32(len(results.Categories))
	return uint32(len(results.Categories)), nil
}

// ListImagesRequest: list images request.
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

// ListImagesResponse: list images response.
type ListImagesResponse struct {
	Images []*Image `json:"images"`

	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListImagesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListImagesResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListImagesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Images = append(r.Images, results.Images...)
	r.TotalCount += uint32(len(results.Images))
	return uint32(len(results.Images)), nil
}

// ListLocalImagesRequest: list local images request.
type ListLocalImagesRequest struct {
	// PageSize: a positive integer lower or equal to 100 to select the number of items to display.
	PageSize *uint32 `json:"-"`

	// Page: a positive integer to choose the page to display.
	Page *int32 `json:"-"`

	// OrderBy: ordering to use.
	// Default value: type_asc
	OrderBy ListLocalImagesRequestOrderBy `json:"-"`

	// Zone: filter local images available on this Availability Zone.
	Zone *scw.Zone `json:"-"`

	// Arch: filter local images available for this machine architecture.
	Arch *string `json:"-"`

	// ImageID: filter by image id.
	// Precisely one of ImageID, VersionID, ImageLabel must be set.
	ImageID *string `json:"image_id,omitempty"`

	// VersionID: filter by version id.
	// Precisely one of ImageID, VersionID, ImageLabel must be set.
	VersionID *string `json:"version_id,omitempty"`

	// ImageLabel: filter by image label.
	// Precisely one of ImageID, VersionID, ImageLabel must be set.
	ImageLabel *string `json:"image_label,omitempty"`

	// Type: filter by type.
	// Default value: unknown_type
	Type LocalImageType `json:"-"`
}

// ListLocalImagesResponse: list local images response.
type ListLocalImagesResponse struct {
	LocalImages []*LocalImage `json:"local_images"`

	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListLocalImagesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListLocalImagesResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListLocalImagesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.LocalImages = append(r.LocalImages, results.LocalImages...)
	r.TotalCount += uint32(len(results.LocalImages))
	return uint32(len(results.LocalImages)), nil
}

// ListVersionsRequest: list versions request.
type ListVersionsRequest struct {
	ImageID string `json:"-"`

	PageSize *uint32 `json:"-"`

	Page *int32 `json:"-"`

	// OrderBy: default value: created_at_asc
	OrderBy ListVersionsRequestOrderBy `json:"-"`
}

// ListVersionsResponse: list versions response.
type ListVersionsResponse struct {
	Versions []*Version `json:"versions"`

	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListVersionsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListVersionsResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListVersionsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Versions = append(r.Versions, results.Versions...)
	r.TotalCount += uint32(len(results.Versions))
	return uint32(len(results.Versions)), nil
}

// This API allows you to find available images for use when launching a Scaleway Instance.
type API struct {
	client *scw.Client
}

// NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

// ListImages: List all available images on the marketplace, their UUID, CPU architecture and description.
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
		Method: "GET",
		Path:   "/marketplace/v2/images",
		Query:  query,
	}

	var resp ListImagesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetImage: Get detailed information about a marketplace image, specified by its `image_id` (UUID format).
func (s *API) GetImage(req *GetImageRequest, opts ...scw.RequestOption) (*Image, error) {
	var err error

	if fmt.Sprint(req.ImageID) == "" {
		return nil, errors.New("field ImageID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/marketplace/v2/images/" + fmt.Sprint(req.ImageID) + "",
	}

	var resp Image

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListVersions: Get a list of all available version of an image, specified by its `image_id` (UUID format).
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
		Method: "GET",
		Path:   "/marketplace/v2/versions",
		Query:  query,
	}

	var resp ListVersionsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetVersion: Get information such as the name, creation date, last update and published date for an image version specified by its `version_id` (UUID format).
func (s *API) GetVersion(req *GetVersionRequest, opts ...scw.RequestOption) (*Version, error) {
	var err error

	if fmt.Sprint(req.VersionID) == "" {
		return nil, errors.New("field VersionID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/marketplace/v2/versions/" + fmt.Sprint(req.VersionID) + "",
	}

	var resp Version

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListLocalImages: List information about local images in a specific Availability Zone, specified by its `image_id` (UUID format), `version_id` (UUID format) or `image_label`. Only one of these three parameters may be set.
func (s *API) ListLocalImages(req *ListLocalImagesRequest, opts ...scw.RequestOption) (*ListLocalImagesResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	defaultZone, exist := s.client.GetDefaultZone()
	if (req.Zone == nil || *req.Zone == "") && exist {
		req.Zone = &defaultZone
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "zone", req.Zone)
	parameter.AddToQuery(query, "arch", req.Arch)
	parameter.AddToQuery(query, "type", req.Type)
	parameter.AddToQuery(query, "image_id", req.ImageID)
	parameter.AddToQuery(query, "version_id", req.VersionID)
	parameter.AddToQuery(query, "image_label", req.ImageLabel)

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/marketplace/v2/local-images",
		Query:  query,
	}

	var resp ListLocalImagesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetLocalImage: Get detailed information about a local image, including compatible commercial types, supported architecture, labels and the Availability Zone of the image, specified by its `local_image_id` (UUID format).
func (s *API) GetLocalImage(req *GetLocalImageRequest, opts ...scw.RequestOption) (*LocalImage, error) {
	var err error

	if fmt.Sprint(req.LocalImageID) == "" {
		return nil, errors.New("field LocalImageID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/marketplace/v2/local-images/" + fmt.Sprint(req.LocalImageID) + "",
	}

	var resp LocalImage

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListCategories: Get a list of all existing categories. The output can be paginated.
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
		Method: "GET",
		Path:   "/marketplace/v2/categories",
		Query:  query,
	}

	var resp ListCategoriesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCategory: Get information about a specific category of the marketplace catalog, specified by its `category_id` (UUID format).
func (s *API) GetCategory(req *GetCategoryRequest, opts ...scw.RequestOption) (*Category, error) {
	var err error

	if fmt.Sprint(req.CategoryID) == "" {
		return nil, errors.New("field CategoryID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/marketplace/v2/categories/" + fmt.Sprint(req.CategoryID) + "",
	}

	var resp Category

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
